package gspro

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

// Connect connects to GSPro server
func (g *Integration) Connect(host string, port int) {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if g.connected {
		return
	}

	// Update host and port
	g.host = host
	g.port = port

	// Set connecting state
	g.stateManager.SetGSProStatus(core.GSProStatusConnecting)

	addr := net.JoinHostPort(g.host, fmt.Sprintf("%d", g.port))
	log.Printf("Connecting to GSPro server at %s", addr)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Printf("Error connecting to GSPro server: %v", err)
		g.stateManager.SetGSProError(fmt.Errorf("failed to connect: %v", err))
		g.stateManager.SetGSProStatus(core.GSProStatusError)
		return
	}

	g.socket = conn
	g.connected = true
	log.Printf("Connected to GSPro server at %s", addr)

	// Start receiving messages in a goroutine
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		g.receiveMessages()
	}()

	// Add a small delay before setting connected state to ensure UI shows "Connecting..."
	time.Sleep(500 * time.Millisecond)
	g.stateManager.SetGSProStatus(core.GSProStatusConnected)
}

// Disconnect disconnects from GSPro server
func (g *Integration) Disconnect() {
	g.connectMutex.Lock()
	defer g.connectMutex.Unlock()

	if !g.connected || g.socket == nil {
		g.connected = false
		g.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
		return
	}

	log.Println("Disconnecting from GSPro server...")

	// Set a deadline for graceful disconnect
	_ = g.socket.SetDeadline(time.Now().Add(2 * time.Second))

	// Close the socket with proper error handling
	if g.socket != nil {
		err := g.socket.Close()
		if err != nil {
			log.Printf("Error closing GSPro connection: %v", err)
			g.stateManager.SetGSProError(fmt.Errorf("error closing connection: %v", err))
			g.stateManager.SetGSProStatus(core.GSProStatusError)
		}
		g.socket = nil
	}

	g.connected = false
	g.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
	log.Println("Disconnected from GSPro server")
}

// receiveMessages receives and processes messages from GSPro
func (g *Integration) receiveMessages() {
	if g.socket == nil {
		return
	}

	buffer := make([]byte, 4096)

	for g.running && g.connected {
		g.socket.SetReadDeadline(time.Now().Add(10 * time.Second))

		n, err := g.socket.Read(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("Error reading from GSPro: %v", err)
			g.stateManager.SetGSProError(fmt.Errorf("error reading from GSPro: %v", err))
			g.stateManager.SetGSProStatus(core.GSProStatusError)
			break
		}

		if n == 0 {
			log.Println("GSPro server closed connection")
			g.stateManager.SetGSProError(fmt.Errorf("server closed connection"))
			g.stateManager.SetGSProStatus(core.GSProStatusError)
			break
		}

		message := string(buffer[:n])
		log.Printf("Received message from GSPro: %s", message)
		g.processMessage(message)
	}

	g.Disconnect()
}

// connectionThread manages connection to GSPro
func (g *Integration) connectionThread() {
	for g.running {
		if !g.connected {
			g.Connect(g.host, g.port)
		}

		// Check connection every 5 seconds
		time.Sleep(5 * time.Second)
	}
}