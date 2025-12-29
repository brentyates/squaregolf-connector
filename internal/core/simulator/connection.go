package simulator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	InitialBackoff    = 5 * time.Second
	MaxBackoff        = 30 * time.Minute
	MaxReconnectTime  = 20 * time.Minute
	MaxFailedAttempts = 20
)

type Base struct {
	Protocol           Protocol
	Host               string
	Port               int
	Socket             net.Conn
	Connected          bool
	Running            bool
	AutoReconnect      bool
	ConnectMutex       sync.Mutex
	Wg                 sync.WaitGroup
	ReconnectAttempts  int
	LastConnectAttempt time.Time
	BackoffDuration    time.Duration
}

func NewBase(protocol Protocol, host string, port int) *Base {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = protocol.DefaultPort()
	}
	return &Base{
		Protocol:        protocol,
		Host:            host,
		Port:            port,
		AutoReconnect:   true,
		BackoffDuration: InitialBackoff,
	}
}

func (b *Base) Connect(host string, port int) {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()

	if b.Connected {
		return
	}

	b.Host = host
	b.Port = port

	if b.Socket != nil {
		log.Printf("[%s] Forcing cleanup of stale socket before reconnection", b.Protocol.Name())
		b.Socket.Close()
		b.Socket = nil
	}

	b.Protocol.SetStatus(StatusConnecting)
	b.LastConnectAttempt = time.Now()

	addr := net.JoinHostPort(b.Host, fmt.Sprintf("%d", b.Port))
	log.Printf("[%s] Connecting to server at %s (attempt %d, backoff: %v)", b.Protocol.Name(), addr, b.ReconnectAttempts+1, b.BackoffDuration)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		b.ReconnectAttempts++
		log.Printf("[%s] Error connecting to server: %v (attempt %d/%d)", b.Protocol.Name(), err, b.ReconnectAttempts, MaxFailedAttempts)
		b.Protocol.SetError(fmt.Errorf("failed to connect: %v", err))
		b.Protocol.SetStatus(StatusError)

		b.BackoffDuration *= 2
		if b.BackoffDuration > MaxBackoff {
			b.BackoffDuration = MaxBackoff
		}
		return
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		log.Printf("[%s] TCP keepalive enabled", b.Protocol.Name())
	}

	b.Socket = conn
	b.Connected = true
	b.ReconnectAttempts = 0
	b.BackoffDuration = InitialBackoff

	log.Printf("[%s] Successfully connected to server at %s", b.Protocol.Name(), addr)

	b.Wg.Add(1)
	go func() {
		defer b.Wg.Done()
		b.receiveMessages()
	}()

	time.Sleep(500 * time.Millisecond)
	b.Protocol.SetStatus(StatusConnected)
	b.Protocol.OnConnected()
}

func (b *Base) Disconnect() {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()

	if !b.Connected || b.Socket == nil {
		b.Connected = false
		b.Protocol.SetStatus(StatusDisconnected)
		b.Protocol.OnDisconnected()
		return
	}

	log.Printf("[%s] Disconnecting from server...", b.Protocol.Name())

	_ = b.Socket.SetDeadline(time.Now().Add(2 * time.Second))

	if b.Socket != nil {
		err := b.Socket.Close()
		if err != nil {
			log.Printf("[%s] Error closing connection: %v", b.Protocol.Name(), err)
			b.Protocol.SetError(fmt.Errorf("error closing connection: %v", err))
			b.Protocol.SetStatus(StatusError)
		}
		b.Socket = nil
	}

	b.Connected = false
	b.Protocol.SetStatus(StatusDisconnected)
	b.Protocol.OnDisconnected()
	log.Printf("[%s] Disconnected from server", b.Protocol.Name())
}

func (b *Base) IsConnected() bool {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()
	return b.Connected
}

func (b *Base) GetConnectionInfo() (string, int) {
	return b.Host, b.Port
}

func (b *Base) Start() {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()

	if b.Running {
		log.Printf("[%s] Integration already running", b.Protocol.Name())
		return
	}

	b.Running = true
	b.Wg.Add(1)
	go func() {
		defer b.Wg.Done()
		b.connectionThread()
	}()
}

func (b *Base) Stop() {
	b.ConnectMutex.Lock()
	if !b.Running {
		b.ConnectMutex.Unlock()
		return
	}
	b.Running = false
	b.ConnectMutex.Unlock()

	b.Disconnect()
	b.Wg.Wait()
}

func (b *Base) EnableAutoReconnect() {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()
	b.AutoReconnect = true
	b.ReconnectAttempts = 0
	b.BackoffDuration = InitialBackoff
	log.Printf("[%s] Auto-reconnect enabled", b.Protocol.Name())
}

func (b *Base) DisableAutoReconnect() {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()
	b.AutoReconnect = false
	log.Printf("[%s] Auto-reconnect disabled", b.Protocol.Name())
}

func (b *Base) ResetReconnectionState() {
	b.ConnectMutex.Lock()
	defer b.ConnectMutex.Unlock()
	b.ReconnectAttempts = 0
	b.BackoffDuration = InitialBackoff
	b.LastConnectAttempt = time.Time{}
	log.Printf("[%s] Reconnection state reset", b.Protocol.Name())
}

func (b *Base) SendMessage(data []byte) error {
	if !b.Connected || b.Socket == nil {
		return fmt.Errorf("not connected to %s", b.Protocol.Name())
	}

	message := append(data, '\n')
	_, err := b.Socket.Write(message)
	if err != nil {
		b.Disconnect()
		return fmt.Errorf("error sending data: %w", err)
	}
	return nil
}

func isValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func findJSONObjects(data []byte) ([]string, []byte) {
	var validObjects []string
	var remaining []byte

	startIdx := bytes.IndexByte(data, '{')
	if startIdx == -1 {
		return validObjects, data
	}

	data = data[startIdx:]
	remaining = data

	for i := 1; i <= len(data); i++ {
		candidateObj := string(data[:i])

		if balancedBraces(candidateObj) && isValidJSON(candidateObj) {
			validObjects = append(validObjects, candidateObj)

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

func (b *Base) receiveMessages() {
	if b.Socket == nil {
		return
	}

	buffer := make([]byte, 4096)
	var messageBuffer []byte

	for b.Running && b.Connected {
		b.Socket.SetReadDeadline(time.Now().Add(10 * time.Second))

		n, err := b.Socket.Read(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			log.Printf("[%s] Error reading from server: %v", b.Protocol.Name(), err)
			b.Protocol.SetError(fmt.Errorf("error reading from server: %v", err))
			b.Protocol.SetStatus(StatusError)
			break
		}

		if n == 0 {
			log.Printf("[%s] Server closed connection", b.Protocol.Name())
			b.Protocol.SetError(fmt.Errorf("server closed connection"))
			b.Protocol.SetStatus(StatusError)
			break
		}

		messageBuffer = append(messageBuffer, buffer[:n]...)

		objects, remaining := findJSONObjects(messageBuffer)
		for _, obj := range objects {
			log.Printf("[%s] Received message: %s", b.Protocol.Name(), obj)
			b.Protocol.ProcessMessage(obj)
		}

		messageBuffer = remaining
	}

	b.Disconnect()
}

func (b *Base) connectionThread() {
	firstAttemptTime := time.Now()

	for b.Running {
		b.ConnectMutex.Lock()
		connected := b.Connected
		autoReconnect := b.AutoReconnect
		reconnectAttempts := b.ReconnectAttempts
		backoff := b.BackoffDuration
		lastAttempt := b.LastConnectAttempt
		b.ConnectMutex.Unlock()

		if !connected && autoReconnect {
			if time.Since(firstAttemptTime) > MaxReconnectTime {
				log.Printf("[%s] Reconnection timeout: exceeded %v of reconnection attempts", b.Protocol.Name(), MaxReconnectTime)
				log.Printf("[%s] Auto-reconnect disabled. Please reconnect manually via the web UI.", b.Protocol.Name())
				b.DisableAutoReconnect()
				b.Protocol.SetError(fmt.Errorf("reconnection timeout: please reconnect manually"))
				b.Protocol.SetStatus(StatusDisconnected)
				continue
			}

			if reconnectAttempts >= MaxFailedAttempts {
				log.Printf("[%s] Reconnection failed: exceeded %d connection attempts", b.Protocol.Name(), MaxFailedAttempts)
				log.Printf("[%s] Auto-reconnect disabled. Please reconnect manually via the web UI.", b.Protocol.Name())
				b.DisableAutoReconnect()
				b.Protocol.SetError(fmt.Errorf("too many failed attempts: please reconnect manually"))
				b.Protocol.SetStatus(StatusDisconnected)
				continue
			}

			if !lastAttempt.IsZero() && time.Since(lastAttempt) < backoff {
				time.Sleep(1 * time.Second)
				continue
			}

			b.Connect(b.Host, b.Port)

			b.ConnectMutex.Lock()
			if b.Connected {
				firstAttemptTime = time.Now()
			}
			b.ConnectMutex.Unlock()
		} else if connected {
			firstAttemptTime = time.Now()
		}

		time.Sleep(1 * time.Second)
	}
}
