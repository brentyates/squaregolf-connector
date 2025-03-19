package core

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

// SimulatorBluetoothClient implements a more sophisticated mock
type SimulatorBluetoothClient struct {
	connected               bool
	batteryLevel            int
	deviceState             DeviceState
	ballState               BallState
	notifyHandlers          map[string]func([]byte)
	characteristics         map[string][]byte
	config                  SimulatorConfig
	lock                    sync.RWMutex
	errorRate               float64 // Probability of simulating errors (0.0-1.0)
	lastActivity            time.Time
	inactivityMonitorActive bool       // Flag to track if the inactivity monitor is running
	rand                    *rand.Rand // For deterministic random generation
	ballDetectionCancel     func()     // Cancel function for the ball detection simulation
	deviceName              string     // Added to match the new BluetoothClient interface
}

// SimulatorConfig holds configuration for the simulator
type SimulatorConfig struct {
	InitialBatteryLevel int
	BatteryDrainRate    float64
	ErrorRate           float64
	ResponseDelay       time.Duration
}

// NewSimulatorBluetoothClient creates a new simulator Bluetooth client
func NewSimulatorBluetoothClient(config SimulatorConfig) *SimulatorBluetoothClient {
	if config.InitialBatteryLevel <= 0 {
		config.InitialBatteryLevel = 80 // Default to 80% battery if not specified
	}

	sim := &SimulatorBluetoothClient{
		connected:               false,
		batteryLevel:            config.InitialBatteryLevel,
		deviceState:             DeviceStateIdle,
		ballState:               BallStateNone,
		notifyHandlers:          make(map[string]func([]byte)),
		characteristics:         make(map[string][]byte),
		config:                  config,
		errorRate:               config.ErrorRate,
		lastActivity:            time.Now(),
		inactivityMonitorActive: false,
		rand:                    rand.New(rand.NewSource(time.Now().UnixNano())),
		ballDetectionCancel:     nil,
	}

	// Initialize default characteristic values
	sim.characteristics[BatteryLevelCharUUID] = []byte{byte(config.InitialBatteryLevel)}

	return sim
}

// Connect simulates connecting to a device with realistic behavior
func (s *SimulatorBluetoothClient) Connect(deviceName, deviceAddress string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	log.Printf("Simulator: Attempting to connect to device: %s", deviceName)

	// Randomly fail connections based on error rate
	if s.simulateError() {
		log.Printf("Simulator: Connection failed due to simulated error")
		return fmt.Errorf("connection failed: device not in range")
	}

	// Store the device name
	s.deviceName = deviceName
	log.Printf("Simulator: Stored device name: %s", s.deviceName)

	// Simulate connection delay
	time.Sleep(s.config.ResponseDelay)
	s.connected = true
	s.deviceState = DeviceStateIdle
	s.lastActivity = time.Now()
	s.inactivityMonitorActive = false // Reset the flag before starting the monitor

	// Start battery drain simulation in background
	go s.simulateBatteryDrain()

	// Start inactivity monitor
	s.startInactivityMonitor()

	log.Printf("Simulator: Successfully connected to device: %s", s.deviceName)
	return nil
}

// Disconnect simulates disconnecting from a device
func (s *SimulatorBluetoothClient) Disconnect() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.connected {
		return fmt.Errorf("not connected to any device")
	}

	s.performDisconnection()
	return nil
}

// Internal method to handle disconnection without locking
func (s *SimulatorBluetoothClient) performDisconnection() {
	// Simulate disconnection delay
	time.Sleep(s.config.ResponseDelay / 2)

	// Cancel any active ball detection simulation
	if s.ballDetectionCancel != nil {
		log.Println("Simulator: Cancelling ball detection simulation due to disconnection")
		s.ballDetectionCancel()
		s.ballDetectionCancel = nil
	}

	s.connected = false
	s.deviceState = DeviceStateIdle
	s.ballState = BallStateNone
	s.inactivityMonitorActive = false

}

// WriteCharacteristic simulates writing to a characteristic with realistic behavior
func (s *SimulatorBluetoothClient) WriteCharacteristic(uuid string, data []byte) error {
	s.lock.Lock()

	if !s.connected {
		s.lock.Unlock()
		return fmt.Errorf("not connected to device")
	}

	// Update activity timestamp
	s.lastActivity = time.Now()

	// Store the written data for later use
	s.characteristics[uuid] = data

	// Make a copy of the data to use after releasing the lock
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	// Check if this is a command characteristic
	isCommandChar := uuid == CommandCharUUID

	// Unlock before time-consuming operations and external calls
	s.lock.Unlock()

	// Simulate write delay (after unlocking)
	time.Sleep(s.config.ResponseDelay)

	// Randomly fail writes based on error rate
	if s.simulateError() {
		return fmt.Errorf("write failed: connection interrupted")
	}

	// Launch monitor uses command UUID for sending commands
	// Call this after releasing the lock to prevent deadlock
	if isCommandChar {
		s.handleCommandData(dataCopy)
	}

	return nil
}

// ReadCharacteristic simulates reading from a characteristic with realistic behavior
func (s *SimulatorBluetoothClient) ReadCharacteristic(uuid string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.connected {
		return nil, fmt.Errorf("not connected to device")
	}

	// Simulate read delay
	time.Sleep(s.config.ResponseDelay)

	// Randomly fail reads based on error rate
	if s.simulateError() {
		return nil, fmt.Errorf("read failed: timeout")
	}

	// If we have a value for this characteristic, return it
	if value, exists := s.characteristics[uuid]; exists {
		return value, nil
	}

	// Special case for common characteristics
	if uuid == BatteryLevelCharUUID {
		return []byte{byte(s.batteryLevel)}, nil
	}

	// Default empty response
	return []byte{}, nil
}

// StartNotifications simulates starting notifications for a characteristic
func (s *SimulatorBluetoothClient) StartNotifications(uuid string, handler func([]byte)) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.connected {
		return fmt.Errorf("not connected to device")
	}

	// Simulate setup delay
	time.Sleep(s.config.ResponseDelay)

	// Store the notification handler
	s.notifyHandlers[uuid] = handler

	// Update activity timestamp
	s.lastActivity = time.Now()

	return nil
}

// StopNotifications simulates stopping notifications for a characteristic
func (s *SimulatorBluetoothClient) StopNotifications(uuid string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.connected {
		return fmt.Errorf("not connected to device")
	}

	// Remove the notification handler
	delete(s.notifyHandlers, uuid)

	// Update activity timestamp
	s.lastActivity = time.Now()

	return nil
}

// IsConnected returns the connection status
func (s *SimulatorBluetoothClient) IsConnected() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.connected
}

// GetDeviceState returns the current device state
func (s *SimulatorBluetoothClient) GetDeviceState() DeviceState {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.deviceState
}

// SetDeviceState changes the device state
func (s *SimulatorBluetoothClient) SetDeviceState(state DeviceState) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.deviceState = state
}

// GetDiscoveredDevices returns a list of discovered devices
func (s *SimulatorBluetoothClient) GetDiscoveredDevices() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	// Return a list of simulated devices
	devices := []string{"SquareGolf(****)"}
	log.Printf("Simulator: Returning discovered devices: %v", devices)
	return devices
}

// GetConnectedDeviceName returns the name of the currently connected device
func (s *SimulatorBluetoothClient) GetConnectedDeviceName() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.deviceName
}

// Private helper methods

// simulateBatteryDrain simulates battery drain over time
func (s *SimulatorBluetoothClient) simulateBatteryDrain() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.lock.Lock()
		if !s.connected {
			s.lock.Unlock()
			return
		}

		// Drain battery based on current state
		var drainAmount float64
		switch s.deviceState {
		default:
			drainAmount = s.config.BatteryDrainRate
		}

		newLevel := s.batteryLevel - int(drainAmount)
		if newLevel < 0 {
			newLevel = 0
		}
		s.batteryLevel = newLevel

		// Update the battery characteristic
		s.characteristics[BatteryLevelCharUUID] = []byte{byte(s.batteryLevel)}

		// Notify about battery level if handler exists
		if handler, exists := s.notifyHandlers[BatteryLevelCharUUID]; exists {
			handler([]byte{byte(s.batteryLevel)})
		}

		s.lock.Unlock()
	}
}

// simulateLaunchDataNotification generates realistic launch data notifications
func (s *SimulatorBluetoothClient) simulateLaunchDataNotification(uuid string, handler func([]byte)) {
	// This would contain logic to generate realistic launch data packets
	// For now we'll just send some dummy data occasionally

	// Check if we should send a notification now
	if time.Now().Second()%5 == 0 { // Every 5 seconds
		// In a real implementation, you would format data according to your device's protocol
		data := []byte{0x01, 0x02, 0x03, 0x04} // Dummy data
		handler(data)
	}
}

// simulateError returns true if an error should be simulated based on error rate
func (s *SimulatorBluetoothClient) simulateError() bool {
	return (s.errorRate > 0) && (s.errorRate > (float64(time.Now().UnixNano()%100) / 100.0))
}

// SetErrorRate allows changing the simulated error rate
func (s *SimulatorBluetoothClient) SetErrorRate(rate float64) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	s.errorRate = rate
}

// startInactivityMonitor starts a continuous monitor that checks for inactivity
func (s *SimulatorBluetoothClient) startInactivityMonitor() {
	const inactivityTimeout = 10 * time.Second
	const checkInterval = 1 * time.Second

	// If already monitoring, don't start another monitor
	if s.inactivityMonitorActive {
		return
	}

	s.inactivityMonitorActive = true

	// Create a ticker that checks for inactivity every second
	ticker := time.NewTicker(checkInterval)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			s.lock.Lock()

			if !s.connected || !s.inactivityMonitorActive {
				s.lock.Unlock()
				return
			}

			elapsedSinceActivity := time.Since(s.lastActivity)
			if elapsedSinceActivity >= inactivityTimeout {
				log.Printf("Disconnecting due to inactivity (no communication for %v)", elapsedSinceActivity)
				s.performDisconnection()
				s.lock.Unlock()
				return
			}

			s.lock.Unlock()
		}
	}()
}

// handleCommandData processes command data sent to the simulator
func (s *SimulatorBluetoothClient) handleCommandData(data []byte) {
	if len(data) < 1 {
		return
	}

	// Check if this is a heartbeat command (first byte is 0x11, second byte is 0x83)
	if len(data) > 1 && data[0] == 0x11 && data[1] == 0x83 {
		// Explicitly update the lastActivity timestamp to prevent disconnection
		s.lock.Lock()
		s.lastActivity = time.Now()
		s.lock.Unlock()
		return
	}

	if len(data) > 1 && data[0] == 0x11 && data[1] == 0x81 {
		log.Println("Simulator: Received command to activate ball detection")

		// Update device state under lock
		s.lock.Lock()
		// Cancel any existing ball detection simulation
		if s.ballDetectionCancel != nil {
			s.ballDetectionCancel()
			s.ballDetectionCancel = nil
		}

		// Set device state to ball detection
		s.deviceState = DeviceStateBallDetection
		// Reset ball state to none when entering ball detection mode
		s.ballState = BallStateNone
		s.lock.Unlock()

		// Create a new context with cancellation for this simulation
		ctx, cancel := context.WithCancel(context.Background())

		s.lock.Lock()
		s.ballDetectionCancel = cancel
		s.lock.Unlock()

		// Start ball detection in a separate goroutine to avoid blocking
		go s.simulateBallDetection(ctx)
	}
}

// simulateBallDetection simulates the ball detection process
func (s *SimulatorBluetoothClient) simulateBallDetection(ctx context.Context) {
	// Ensure we clean up the cancel function when we exit
	defer func() {
		s.lock.Lock()
		if s.ballDetectionCancel != nil {
			s.ballDetectionCancel = nil
		}
		s.lock.Unlock()
	}()

	for {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Println("Simulator: Ball detection simulation cancelled")
			return
		default:
			// Continue with simulation
		}

		if s.deviceState != DeviceStateBallDetection {
			// If not in ball detection mode, exit the simulation
			log.Println("Simulator: Ball detection deactivated, stopping simulation")
			return
		}

		// If we're already in a ball state other than none, wait for it to be reset
		if s.ballState != BallStateNone {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
				continue
			}
		}

		// Simulate the time it takes for a golfer to place a ball (3 seconds)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			// Continue with simulation
		}

		if s.deviceState != DeviceStateBallDetection {
			log.Println("Simulator: Ball detection deactivated while waiting for ball placement")
			return // Ball detection was deactivated while waiting
		}

		// Ball is detected
		s.lock.Lock()
		s.ballState = BallStateDetected
		s.lock.Unlock()

		sensorData := s.generateSensorData(s.ballState)
		handler := s.notifyHandlers[NotificationCharUUID]
		handler(sensorData)

		// Simulate the time it takes for the ball to become ready (1.5 seconds)
		select {
		case <-ctx.Done():
			return
		case <-time.After(1500 * time.Millisecond):
			// Continue with simulation
		}

		if s.deviceState != DeviceStateBallDetection {
			log.Println("Simulator: Ball detection deactivated while waiting for ball to be ready")
			return // Ball detection was deactivated while waiting
		}

		// Ball is ready
		s.lock.Lock()
		s.ballState = BallStateReady
		s.lock.Unlock()

		sensorData = s.generateSensorData(s.ballState)
		handler = s.notifyHandlers[NotificationCharUUID]
		handler(sensorData)

		// Simulate the time it takes for a golfer to take a shot (4 seconds)
		select {
		case <-ctx.Done():
			return
		case <-time.After(4 * time.Second):
			// Continue with simulation
		}

		if s.deviceState != DeviceStateBallDetection || s.ballState != BallStateReady {
			return // Ball detection was deactivated or ball state changed while waiting
		}

		// Ball is hit
		handler = s.notifyHandlers[NotificationCharUUID]
		s.sendBallMetrics(handler)

		s.lock.Lock()
		s.deviceState = DeviceStateIdle
		s.ballState = BallStateNone
		s.lock.Unlock()

		// Wait for the ball metrics to be sent and state to be reset to none
		// This happens in the notification handler
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			// Continue with simulation
		}
	}
}

// generateSensorData creates realistic sensor data packets based on ball state
func (s *SimulatorBluetoothClient) generateSensorData(ballState BallState) []byte {
	// Format: 11 01 [ball ready] [ball detected] [position X] [position Y] [position Z] ...

	ballDetected := byte(0x00)
	ballReady := byte(0x00)

	if ballState == BallStateDetected || ballState == BallStateReady {
		ballDetected = 0x01
	}

	if ballState == BallStateReady {
		ballReady = 0x01
	}

	// Generate realistic ball position
	var posX, posY, posZ int32

	if ballState == BallStateDetected {
		// Ball is detected but not perfectly positioned
		posX = int32(-50 + s.rand.Intn(100))
		posY = int32(-50 + s.rand.Intn(100))
		posZ = int32(s.rand.Intn(20))
	} else if ballState == BallStateReady {
		// Ball is properly positioned
		posX = int32(-10 + s.rand.Intn(20))
		posY = int32(-10 + s.rand.Intn(20))
		posZ = int32(s.rand.Intn(10))
	}

	// Convert to bytes (simplified format for the mock)
	data := []byte{
		0x11, 0x01, // Header (indexes 0-1)
		0x00,         // Placeholder at index 2
		ballReady,    // Ball ready flag at index 3
		ballDetected, // Ball detection flag at index 4
		byte(posX & 0xFF), byte((posX >> 8) & 0xFF), byte((posX >> 16) & 0xFF), byte((posX >> 24) & 0xFF),
		byte(posY & 0xFF), byte((posY >> 8) & 0xFF), byte((posY >> 16) & 0xFF), byte((posY >> 24) & 0xFF),
		byte(posZ & 0xFF), byte((posZ >> 8) & 0xFF), byte((posZ >> 16) & 0xFF), byte((posZ >> 24) & 0xFF),
	}

	return data
}

// sendBallMetrics sends simulated ball metrics data with realistic values
func (s *SimulatorBluetoothClient) sendBallMetrics(handler func([]byte)) {
	s.lock.RLock()
	if !s.connected {
		s.lock.RUnlock()
		return
	}
	s.lock.RUnlock()

	// Generate realistic ball metrics with the correct header format (0x11, 0x02, 0x37)
	// Generate realistic ball metrics with slight randomization
	// Ball speed between 250-350 mph
	ballSpeed := uint16(2500 + s.rand.Intn(1000))
	// Launch angle between 8-14 degrees
	launchAngle := uint16(80 + s.rand.Intn(60))
	// Side angle between -5 to 5 degrees
	sideAngle := uint16(s.rand.Intn(100) - 50)
	// Total spin between 100-200 rpm
	totalSpin := uint16(100 + s.rand.Intn(100))
	// Back spin between 80-120 rpm
	backSpin := uint16(80 + s.rand.Intn(40))
	// Side spin between 30-70 rpm
	sideSpin := uint16(30 + s.rand.Intn(40))
	// Rifle spin between 0-10 rpm
	rifleSpin := uint16(s.rand.Intn(10))

	ballData := []byte{
		0x11, 0x02, 0x37, // Header for ball data that matches what the launch monitor expects
		byte(ballSpeed & 0xFF), byte((ballSpeed >> 8) & 0xFF),
		byte(launchAngle & 0xFF), byte((launchAngle >> 8) & 0xFF),
		byte(sideAngle & 0xFF), byte((sideAngle >> 8) & 0xFF),
		byte(totalSpin & 0xFF), byte((totalSpin >> 8) & 0xFF),
		byte(backSpin & 0xFF), byte((backSpin >> 8) & 0xFF),
		byte(sideSpin & 0xFF), byte((sideSpin >> 8) & 0xFF),
		byte(rifleSpin & 0xFF), byte((rifleSpin >> 8) & 0xFF),
	}

	// Send the ball metrics
	handler(ballData)
}

// StartScan simulates starting a scan for BLE devices
func (s *SimulatorBluetoothClient) StartScan(prefix string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Simulate scan delay
	time.Sleep(s.config.ResponseDelay)

	// Randomly fail scans based on error rate
	if s.simulateError() {
		return fmt.Errorf("scan failed: device not responding")
	}

	return nil
}

// StopScan simulates stopping a BLE scan
func (s *SimulatorBluetoothClient) StopScan() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Simulate stop delay
	time.Sleep(s.config.ResponseDelay / 2)

	// Randomly fail stop based on error rate
	if s.simulateError() {
		return fmt.Errorf("failed to stop scan: device not responding")
	}

	return nil
}
