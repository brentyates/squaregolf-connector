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
type Integration struct {
	stateManager       *core.StateManager
	launchMonitor      *core.LaunchMonitor
	host               string
	port               int
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
}

const (
	initialBackoff    = 5 * time.Second
	maxBackoff        = 30 * time.Minute
	maxReconnectTime  = 20 * time.Minute
	maxFailedAttempts = 20
)

// GetInstance returns the singleton instance of GSProIntegration
func GetInstance(stateManager *core.StateManager, launchMonitor *core.LaunchMonitor, host string, port int) *Integration {
	gsproOnce.Do(func() {
		if host == "" {
			host = "127.0.0.1"
		}
		if port == 0 {
			port = 921
		}

		gsproInstance = &Integration{
			stateManager:    stateManager,
			launchMonitor:   launchMonitor,
			host:            host,
			port:            port,
			shotListeners:   make([]func(ShotData), 0),
			autoReconnect:   true,
			backoffDuration: initialBackoff,
		}

		// Register state listeners
		gsproInstance.registerStateListeners()
	})
	return gsproInstance
}

// Start starts the GSPro integration in a separate goroutine
func (g *Integration) Start() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if g.running {
		log.Println("GSPro integration already running")
		return
	}

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
