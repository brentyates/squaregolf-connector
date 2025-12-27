package gspro

import (
	"log"
	"net"
	"sync"
	"time"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

var (
	gsproInstance *Integration
	gsproOnce     sync.Once
)

// Integration handles communication with GSPro
// Supports both GSPro Connect mode and Direct mode
type Integration struct {
	stateManager       *core.StateManager
	launchMonitor      *core.LaunchMonitor
	host               string
	port               int
	mode               string // "connect" or "direct"
	socket             net.Conn
	connected          bool
	running            bool
	autoReconnect      bool
	connectMutex       sync.Mutex
	shotNumber         int
	lastShotNumber     int
	shotListeners      []func(ShotData)
	lastPlayerInfo     *PlayerInfo
	wg                 sync.WaitGroup
	reconnectAttempts  int
	lastConnectAttempt time.Time
	backoffDuration    time.Duration
	// Direct mode implementation (nil if in Connect mode)
	directImpl *DirectIntegration
}

const (
	initialBackoff    = 5 * time.Second
	maxBackoff        = 30 * time.Minute
	maxReconnectTime  = 20 * time.Minute
	maxFailedAttempts = 20
)

// GetInstance returns the singleton instance of GSProIntegration
// mode should be "connect" or "direct"
func GetInstance(stateManager *core.StateManager, launchMonitor *core.LaunchMonitor, host string, port int) *Integration {
	gsproOnce.Do(func() {
		if host == "" {
			host = "127.0.0.1"
		}
		if port == 0 {
			port = 921
		}

		// Default to Connect mode for backward compatibility
		gsproInstance = &Integration{
			stateManager:    stateManager,
			launchMonitor:   launchMonitor,
			host:            host,
			port:            port,
			mode:            "connect", // Default mode
			shotListeners:   make([]func(ShotData), 0),
			autoReconnect:   true,
			backoffDuration: initialBackoff,
			directImpl:      nil,
		}

		// Register state listeners (for Connect mode)
		gsproInstance.registerStateListeners()
	})
	return gsproInstance
}

// SetMode sets the GSPro integration mode ("connect" or "direct")
// Must be called before Start()
func (g *Integration) SetMode(mode string) {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if g.running {
		log.Printf("Cannot change GSPro mode while running. Stop the integration first.")
		return
	}

	if mode != "connect" && mode != "direct" {
		log.Printf("Invalid GSPro mode: %s (must be 'connect' or 'direct')", mode)
		return
	}

	g.mode = mode
	log.Printf("GSPro mode set to: %s", mode)

	// Initialize Direct implementation if switching to direct mode
	if mode == "direct" && g.directImpl == nil {
		g.directImpl = NewDirectIntegration(g.stateManager, g.launchMonitor, g.host, g.port)
	}
}

// GetMode returns the current GSPro integration mode
func (g *Integration) GetMode() string {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()
	return g.mode
}

// Start starts the GSPro integration in a separate goroutine
func (g *Integration) Start() {
	g.connectMutex.Lock()
	mode := g.mode
	g.connectMutex.Unlock()

	// Delegate to Direct implementation if in direct mode
	if mode == "direct" {
		if g.directImpl != nil {
			log.Println("Starting GSPro in Direct mode")
			g.directImpl.Start()
		} else {
			log.Println("Error: Direct mode selected but directImpl is nil")
		}
		return
	}

	// Connect mode logic
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if g.running {
		log.Println("GSPro integration already running")
		return
	}

	log.Println("Starting GSPro in Connect mode")
	g.running = true
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.connectionThread()
	}()
}

// Stop stops the GSPro integration
func (g *Integration) Stop() {
	g.connectMutex.Lock()
	mode := g.mode
	g.connectMutex.Unlock()

	// Delegate to Direct implementation if in direct mode
	if mode == "direct" {
		if g.directImpl != nil {
			g.directImpl.Stop()
		}
		return
	}

	// Connect mode logic
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if !g.running {
		return
	}

	g.running = false
	g.Disconnect()
	g.wg.Wait()
}

// GetConnectionInfo returns the current host and port configuration
func (g *Integration) GetConnectionInfo() (string, int) {
	return g.host, g.port
}

// EnableAutoReconnect enables automatic reconnection
func (g *Integration) EnableAutoReconnect() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()
	g.autoReconnect = true
	g.reconnectAttempts = 0
	g.backoffDuration = initialBackoff
	log.Println("GSPro auto-reconnect enabled")
}

// DisableAutoReconnect disables automatic reconnection
func (g *Integration) DisableAutoReconnect() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()
	g.autoReconnect = false
	log.Println("GSPro auto-reconnect disabled")
}

// ResetReconnectionState resets the reconnection attempt counter and backoff
func (g *Integration) ResetReconnectionState() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()
	g.reconnectAttempts = 0
	g.backoffDuration = initialBackoff
	g.lastConnectAttempt = time.Time{}
	log.Println("GSPro reconnection state reset")
}
