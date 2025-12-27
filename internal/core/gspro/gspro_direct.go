package gspro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

// DirectIntegration handles direct communication with GSPro (bypassing GSPro Connect)
type DirectIntegration struct {
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

// NewDirectIntegration creates a new direct GSPro integration
func NewDirectIntegration(stateManager *core.StateManager, launchMonitor *core.LaunchMonitor, host string, port int) *DirectIntegration {
	if host == "" {
		host = "127.0.0.1"
	}
	// TODO: Determine the actual direct GSPro port via reverse engineering
	// For now, using a placeholder port (likely different from 921)
	if port == 0 {
		port = 921 // PLACEHOLDER - will be updated after reverse engineering
	}

	di := &DirectIntegration{
		stateManager:    stateManager,
		launchMonitor:   launchMonitor,
		host:            host,
		port:            port,
		shotListeners:   make([]func(ShotData), 0),
		autoReconnect:   true,
		backoffDuration: initialBackoff,
	}

	// Register state listeners
	di.registerStateListeners()

	return di
}

// Start starts the direct GSPro integration in a separate goroutine
func (d *DirectIntegration) Start() {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()

	if d.running {
		log.Println("GSPro Direct integration already running")
		return
	}

	log.Println("Starting GSPro Direct integration (experimental)")
	d.running = true
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.connectionThread()
	}()
}

// Stop stops the direct GSPro integration
func (d *DirectIntegration) Stop() {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()

	if !d.running {
		return
	}

	log.Println("Stopping GSPro Direct integration")
	d.running = false
	d.Disconnect()
	d.wg.Wait()
}

// Connect connects to GSPro server directly
func (d *DirectIntegration) Connect(host string, port int) {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()

	if d.connected {
		return
	}

	// Update host and port
	d.host = host
	d.port = port

	// Force cleanup of any existing socket
	if d.socket != nil {
		log.Println("[Direct] Forcing cleanup of stale socket before reconnection")
		d.socket.Close()
		d.socket = nil
	}

	// Set connecting state
	d.stateManager.SetGSProStatus(core.GSProStatusConnecting)
	d.lastConnectAttempt = time.Now()

	addr := net.JoinHostPort(d.host, fmt.Sprintf("%d", d.port))
	log.Printf("[Direct] Connecting to GSPro server at %s (attempt %d, backoff: %v)", addr, d.reconnectAttempts+1, d.backoffDuration)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		d.reconnectAttempts++
		log.Printf("[Direct] Error connecting to GSPro server: %v (attempt %d/%d)", err, d.reconnectAttempts, maxFailedAttempts)
		d.stateManager.SetGSProError(fmt.Errorf("failed to connect: %v", err))
		d.stateManager.SetGSProStatus(core.GSProStatusError)

		// Exponential backoff
		d.backoffDuration *= 2
		if d.backoffDuration > maxBackoff {
			d.backoffDuration = maxBackoff
		}

		return
	}

	// Enable TCP keepalive
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		log.Println("[Direct] TCP keepalive enabled on GSPro connection")
	}

	d.socket = conn
	d.connected = true

	// Reset reconnection state on successful connection
	d.reconnectAttempts = 0
	d.backoffDuration = initialBackoff

	log.Printf("[Direct] Successfully connected to GSPro server at %s", addr)

	// TODO: Send initial handshake to GSPro (protocol to be determined via reverse engineering)
	d.sendInitialHandshake()

	// Start receiving messages in a goroutine
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.receiveMessages()
	}()

	// Add a small delay before setting connected state
	time.Sleep(500 * time.Millisecond)
	d.stateManager.SetGSProStatus(core.GSProStatusConnected)
}

// Disconnect disconnects from GSPro server
func (d *DirectIntegration) Disconnect() {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()

	if !d.connected || d.socket == nil {
		d.connected = false
		d.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
		return
	}

	log.Println("[Direct] Disconnecting from GSPro server...")

	// Set a deadline for graceful disconnect
	_ = d.socket.SetDeadline(time.Now().Add(2 * time.Second))

	// Close the socket
	if d.socket != nil {
		err := d.socket.Close()
		if err != nil {
			log.Printf("[Direct] Error closing GSPro connection: %v", err)
			d.stateManager.SetGSProError(fmt.Errorf("error closing connection: %v", err))
			d.stateManager.SetGSProStatus(core.GSProStatusError)
		}
		d.socket = nil
	}

	d.connected = false
	d.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
	log.Println("[Direct] Disconnected from GSPro server")
}

// sendInitialHandshake sends the initial handshake to GSPro
// TODO: Protocol to be determined via reverse engineering of GSPro Connect
func (d *DirectIntegration) sendInitialHandshake() {
	log.Println("[Direct] TODO: Send initial handshake (protocol TBD)")
	// PLACEHOLDER: This will be implemented after reverse engineering reveals the protocol
	// Possible handshake formats:
	// - JSON registration message
	// - Binary protocol header
	// - Authentication token
	// - Device identification
}

// receiveMessages receives and processes messages from GSPro
// TODO: Message format to be determined via reverse engineering
func (d *DirectIntegration) receiveMessages() {
	if d.socket == nil {
		return
	}

	buffer := make([]byte, 4096)
	var messageBuffer []byte

	for d.running && d.connected {
		d.socket.SetReadDeadline(time.Now().Add(10 * time.Second))

		n, err := d.socket.Read(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("[Direct] Error reading from GSPro: %v", err)
			d.stateManager.SetGSProError(fmt.Errorf("error reading from GSPro: %v", err))
			d.stateManager.SetGSProStatus(core.GSProStatusError)
			break
		}

		if n == 0 {
			log.Println("[Direct] GSPro server closed connection")
			d.stateManager.SetGSProError(fmt.Errorf("server closed connection"))
			d.stateManager.SetGSProStatus(core.GSProStatusError)
			break
		}

		// Append the received data to the message buffer
		messageBuffer = append(messageBuffer, buffer[:n]...)

		// TODO: Determine message format (JSON, binary, protobuf, etc.)
		// For now, assume JSON similar to GSPro Connect until proven otherwise
		objects, remaining := findJSONObjects(messageBuffer)
		for _, obj := range objects {
			log.Printf("[Direct] Received message from GSPro: %s", obj)
			d.processMessage(obj)
		}

		// Keep the remaining unprocessed data
		messageBuffer = remaining
	}

	d.Disconnect()
}

// processMessage processes a message from GSPro
// TODO: Message types to be determined via reverse engineering
func (d *DirectIntegration) processMessage(message string) {
	log.Printf("[Direct] Processing message: %s", message)

	// PLACEHOLDER: Message processing logic
	// This will be implemented based on reverse engineering findings

	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("[Direct] Error parsing message: %v", err)
		return
	}

	// TODO: Handle different message types from GSPro
	// Examples (speculative until reverse engineering reveals actual protocol):
	// - Ready message
	// - Player information
	// - Shot acknowledgment
	// - Error messages
	// - Keep-alive pings

	messageType, ok := msg["Message"].(string)
	if !ok {
		log.Printf("[Direct] Unknown message format: %v", msg)
		return
	}

	switch messageType {
	case "GSPro ready":
		log.Println("[Direct] GSPro is ready")
		d.handleGSProReady()
	// TODO: Add other message types after reverse engineering
	default:
		log.Printf("[Direct] Unhandled message type: %s", messageType)
	}
}

// handleGSProReady handles the GSPro ready message
func (d *DirectIntegration) handleGSProReady() {
	log.Println("[Direct] GSPro ready - activating launch monitor")
	d.launchMonitor.Activate()
}

// SendShotData sends shot data to GSPro
// TODO: Message format to be determined via reverse engineering
func (d *DirectIntegration) SendShotData(shotData ShotData) error {
	d.connectMutex.Lock()
	connected := d.connected
	socket := d.socket
	d.connectMutex.Unlock()

	if !connected || socket == nil {
		return fmt.Errorf("not connected to GSPro")
	}

	log.Printf("[Direct] Sending shot data: %+v", shotData)

	// TODO: Format message according to GSPro's direct protocol
	// This might be different from GSPro Connect's format
	// PLACEHOLDER: Using GSPro Connect format until reverse engineering reveals differences

	message := formatShotDataMessage(shotData)
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("error marshaling shot data: %v", err)
	}

	// Add newline terminator (might need adjustment based on actual protocol)
	messageBytes = append(messageBytes, '\n')

	_, err = socket.Write(messageBytes)
	if err != nil {
		log.Printf("[Direct] Error sending shot data: %v", err)
		d.stateManager.SetGSProError(fmt.Errorf("error sending shot data: %v", err))
		return err
	}

	log.Printf("[Direct] Shot data sent successfully")
	return nil
}

// connectionThread manages connection to GSPro
func (d *DirectIntegration) connectionThread() {
	firstAttemptTime := time.Now()

	for d.running {
		d.connectMutex.Lock()
		connected := d.connected
		autoReconnect := d.autoReconnect
		reconnectAttempts := d.reconnectAttempts
		backoff := d.backoffDuration
		lastAttempt := d.lastConnectAttempt
		d.connectMutex.Unlock()

		if !connected && autoReconnect {
			// Check if we've exceeded the maximum reconnection time
			if time.Since(firstAttemptTime) > maxReconnectTime {
				log.Printf("[Direct] GSPro reconnection timeout: exceeded %v of reconnection attempts", maxReconnectTime)
				log.Println("[Direct] Auto-reconnect disabled. Please reconnect manually via the web UI.")
				d.DisableAutoReconnect()
				d.stateManager.SetGSProError(fmt.Errorf("reconnection timeout: please reconnect manually"))
				d.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
				continue
			}

			// Check if we've exceeded the maximum failed attempts
			if reconnectAttempts >= maxFailedAttempts {
				log.Printf("[Direct] GSPro reconnection failed: exceeded %d connection attempts", maxFailedAttempts)
				log.Println("[Direct] Auto-reconnect disabled. Please reconnect manually via the web UI.")
				d.DisableAutoReconnect()
				d.stateManager.SetGSProError(fmt.Errorf("too many failed attempts: please reconnect manually"))
				d.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
				continue
			}

			// Apply exponential backoff
			if !lastAttempt.IsZero() && time.Since(lastAttempt) < backoff {
				time.Sleep(1 * time.Second)
				continue
			}

			// Attempt to connect
			d.Connect(d.host, d.port)

			// Reset timer if we successfully connected
			d.connectMutex.Lock()
			if d.connected {
				firstAttemptTime = time.Now()
			}
			d.connectMutex.Unlock()
		} else if connected {
			// Reset timer when connected
			firstAttemptTime = time.Now()
		}

		// Check connection status regularly
		time.Sleep(1 * time.Second)
	}
}

// EnableAutoReconnect enables automatic reconnection
func (d *DirectIntegration) EnableAutoReconnect() {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()
	d.autoReconnect = true
	d.reconnectAttempts = 0
	d.backoffDuration = initialBackoff
	log.Println("[Direct] GSPro auto-reconnect enabled")
}

// DisableAutoReconnect disables automatic reconnection
func (d *DirectIntegration) DisableAutoReconnect() {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()
	d.autoReconnect = false
	log.Println("[Direct] GSPro auto-reconnect disabled")
}

// ResetReconnectionState resets the reconnection attempt counter and backoff
func (d *DirectIntegration) ResetReconnectionState() {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()
	d.reconnectAttempts = 0
	d.backoffDuration = initialBackoff
	d.lastConnectAttempt = time.Time{}
	log.Println("[Direct] GSPro reconnection state reset")
}

// GetConnectionInfo returns the current host and port configuration
func (d *DirectIntegration) GetConnectionInfo() (string, int) {
	return d.host, d.port
}

// IsConnected returns whether the integration is connected
func (d *DirectIntegration) IsConnected() bool {
	d.connectMutex.Lock()
	defer d.connectMutex.Unlock()
	return d.connected
}

// registerStateListeners registers listeners for state changes
func (d *DirectIntegration) registerStateListeners() {
	log.Println("[Direct] Registering state listeners")
	// TODO: Implement state listener registration similar to GSPro Connect
	// This will handle ball detection, shot events, etc.
}
