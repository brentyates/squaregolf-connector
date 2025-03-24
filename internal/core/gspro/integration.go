package gspro

import (
	"fmt"
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
	stateManager   *core.StateManager
	launchMonitor  *core.LaunchMonitor
	host           string
	port           int
	socket         net.Conn
	connected      bool
	running        bool
	connectMutex   sync.Mutex
	shotNumber     int
	lastShotNumber int
	shotListeners  []func(ShotData)
	lastPlayerInfo *PlayerInfo
	wg             sync.WaitGroup
}

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
			stateManager:  stateManager,
			launchMonitor: launchMonitor,
			host:          host,
			port:          port,
			shotListeners: make([]func(ShotData), 0),
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
	log.Println("Stopping GSPro integration...")
	g.running = false
	g.Disconnect()

	// Wait for all goroutines to complete with a timeout
	done := make(chan struct{})
	go func() {
		g.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("GSPro integration stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Println("Timeout waiting for GSPro integration to stop")
		g.stateManager.SetGSProError(fmt.Errorf("timeout waiting for integration to stop"))
		g.stateManager.SetGSProStatus(core.GSProStatusError)
	}
}
