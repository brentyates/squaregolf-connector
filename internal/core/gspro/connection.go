package gspro

import (
	"bytes"
	"encoding/json"
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

	// Force cleanup of any existing socket
	if g.socket != nil {
		log.Println("Forcing cleanup of stale socket before reconnection")
		g.socket.Close()
		g.socket = nil
	}

	// Set connecting state
	g.stateManager.SetGSProStatus(core.GSProStatusConnecting)
	g.lastConnectAttempt = time.Now()

	addr := net.JoinHostPort(g.host, fmt.Sprintf("%d", g.port))
	log.Printf("Connecting to GSPro server at %s (attempt %d, backoff: %v)", addr, g.reconnectAttempts+1, g.backoffDuration)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		g.reconnectAttempts++
		log.Printf("Error connecting to GSPro server: %v (attempt %d/%d)", err, g.reconnectAttempts, maxFailedAttempts)
		g.stateManager.SetGSProError(fmt.Errorf("failed to connect: %v", err))
		g.stateManager.SetGSProStatus(core.GSProStatusError)

		// Exponential backoff
		g.backoffDuration *= 2
		if g.backoffDuration > maxBackoff {
			g.backoffDuration = maxBackoff
		}

		return
	}

	// Enable TCP keepalive for better connection health detection
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		log.Println("TCP keepalive enabled on GSPro connection")
	}

	g.socket = conn
	g.connected = true

	// Reset reconnection state on successful connection
	g.reconnectAttempts = 0
	g.backoffDuration = initialBackoff

	log.Printf("Successfully connected to GSPro server at %s", addr)

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

// isValidJSON checks if a string is valid JSON
func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// findJSONObjects scans a buffer for valid JSON objects and returns them
// along with the remaining unconsumed buffer
func findJSONObjects(data []byte) ([]string, []byte) {
	var validObjects []string
	var remaining []byte

	// Skip any non-JSON leading characters
	startIdx := bytes.IndexByte(data, '{')
	if startIdx == -1 {
		// No JSON object start found
		return validObjects, data
	}

	data = data[startIdx:]
	remaining = data

	// Try to find JSON objects by testing increasing slices
	for i := 1; i <= len(data); i++ {
		candidateObj := string(data[:i])

		// Check for balanced braces as a quick heuristic before trying to parse JSON
		if balancedBraces(candidateObj) && isValidJSON(candidateObj) {
			validObjects = append(validObjects, candidateObj)

			// Process the rest of the buffer
			if i < len(data) {
				newObjects, newRemaining := findJSONObjects(data[i:])
				validObjects = append(validObjects, newObjects...)
				remaining = newRemaining
				break
			} else {
				remaining = nil
				break
			}
		}
	}

	return validObjects, remaining
}

// balancedBraces checks if braces are balanced in a string
func balancedBraces(s string) bool {
	var count int
	for _, c := range s {
		if c == '{' {
			count++
		} else if c == '}' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// receiveMessages receives and processes messages from GSPro
func (g *Integration) receiveMessages() {
	if g.socket == nil {
		return
	}

	buffer := make([]byte, 4096)
	var messageBuffer []byte

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

		// Append the received data to the message buffer
		messageBuffer = append(messageBuffer, buffer[:n]...)

		// Find and process complete JSON objects
		objects, remaining := findJSONObjects(messageBuffer)
		for _, obj := range objects {
			log.Printf("Received message from GSPro: %s", obj)
			g.processMessage(obj)
		}

		// Keep the remaining unprocessed data
		messageBuffer = remaining
	}

	g.Disconnect()
}

// connectionThread manages connection to GSPro
func (g *Integration) connectionThread() {
	firstAttemptTime := time.Now()

	for g.running {
		g.connectMutex.Lock()
		connected := g.connected
		autoReconnect := g.autoReconnect
		reconnectAttempts := g.reconnectAttempts
		backoff := g.backoffDuration
		lastAttempt := g.lastConnectAttempt
		g.connectMutex.Unlock()

		if !connected && autoReconnect {
			// Check if we've exceeded the maximum reconnection time
			if time.Since(firstAttemptTime) > maxReconnectTime {
				log.Printf("GSPro reconnection timeout: exceeded %v of reconnection attempts", maxReconnectTime)
				log.Println("GSPro auto-reconnect disabled. Please reconnect manually via the web UI.")
				g.DisableAutoReconnect()
				g.stateManager.SetGSProError(fmt.Errorf("reconnection timeout: please reconnect manually"))
				g.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
				continue
			}

			// Check if we've exceeded the maximum failed attempts
			if reconnectAttempts >= maxFailedAttempts {
				log.Printf("GSPro reconnection failed: exceeded %d connection attempts", maxFailedAttempts)
				log.Println("GSPro auto-reconnect disabled. Please reconnect manually via the web UI.")
				g.DisableAutoReconnect()
				g.stateManager.SetGSProError(fmt.Errorf("too many failed attempts: please reconnect manually"))
				g.stateManager.SetGSProStatus(core.GSProStatusDisconnected)
				continue
			}

			// Apply exponential backoff
			if !lastAttempt.IsZero() && time.Since(lastAttempt) < backoff {
				time.Sleep(1 * time.Second)
				continue
			}

			// Attempt to connect
			g.Connect(g.host, g.port)

			// Reset timer if we successfully connected
			g.connectMutex.Lock()
			if g.connected {
				firstAttemptTime = time.Now()
			}
			g.connectMutex.Unlock()
		} else if connected {
			// Reset timer when connected
			firstAttemptTime = time.Now()
		}

		// Check connection status regularly
		time.Sleep(1 * time.Second)
	}
}
