package core

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	launchMonitorInstance *LaunchMonitor
	launchMonitorOnce     sync.Once
)

// GetLaunchMonitorInstance returns the singleton instance of LaunchMonitor
func GetLaunchMonitorInstance(sm *StateManager, btManager *BluetoothManager) *LaunchMonitor {
	launchMonitorOnce.Do(func() {
		launchMonitorInstance = &LaunchMonitor{
			stateManager:    sm,
			sequence:        0,
			bluetoothClient: btManager.GetClient(),
		}
	})
	return launchMonitorInstance
}

// NewLaunchMonitor is deprecated, use GetLaunchMonitorInstance instead
func NewLaunchMonitor(sm *StateManager, btManager *BluetoothManager) *LaunchMonitor {
	return GetLaunchMonitorInstance(sm, btManager)
}

// LaunchMonitor encapsulates the launch monitor functionality
type LaunchMonitor struct {
	stateManager      *StateManager
	sequence          int
	sequenceMutex     sync.Mutex
	heartbeatCancel   context.CancelFunc
	heartbeatCancelMu sync.Mutex
	bluetoothClient   BluetoothClient
}

// UpdateBluetoothClient updates the bluetooth client reference
func (lm *LaunchMonitor) UpdateBluetoothClient(client BluetoothClient) {
	lm.bluetoothClient = client
}

// NotificationHandler handles BLE notifications
func (lm *LaunchMonitor) NotificationHandler(uuid string, data []byte) {
	if len(data) == 0 {
		log.Println("Received empty notification data")
		return
	}

	hexData := hex.EncodeToString(data)

	// Handle battery level notification
	if uuid == BatteryLevelCharUUID {
		batteryLevel := int(data[0])
		lm.stateManager.SetBatteryLevel(&batteryLevel)
		return
	}

	// Split hex string into byte pairs
	var bytesList []string
	for i := 0; i < len(hexData); i += 2 {
		if i+2 <= len(hexData) {
			bytesList = append(bytesList, hexData[i:i+2])
		}
	}

	// Process by byte patterns
	if len(bytesList) >= 2 {
		// Handle alignment notifications (format 11 82)
		// TODO: Verify this is the correct format from Bluetooth traffic capture
		if bytesList[0] == "11" && bytesList[1] == "82" {
			lm.HandleAlignmentNotification(bytesList)
			return
		}

		// Sensor notifications (format 11 01)
		if bytesList[0] == "11" && bytesList[1] == "01" {
			lm.HandleSensorNotification(bytesList)
		} else if len(bytesList) >= 3 {
			// Shot Ball Metrics (format 11 02)
			if bytesList[0] == "11" && bytesList[1] == "02" {
				lm.HandleShotBallMetrics(bytesList)
				return
			}
			if bytesList[0] == "11" && bytesList[1] == "03" {
				// Heartbeat from the device
				return
			}
			// OS Version response (format 11 10)
			if bytesList[0] == "11" && bytesList[1] == "10" {
				lm.HandleOSVersionNotification(bytesList)
				return
			}
			// Shot Club Metrics (format 11 07 0f)
			if bytesList[0] == "11" && bytesList[1] == "07" && bytesList[2] == "0f" {
				lm.HandleShotClubMetrics(bytesList)
				return
			}
			// Check for specific "no club data available" response
			if bytesList[0] == "11" && bytesList[1] == "07" && bytesList[2] == "00" {
				// Clear club metrics in state manager to indicate no data is available
				lm.stateManager.SetLastClubMetrics(nil)
				return
			}
		}
	}
}

// HandleSensorNotification handles sensor notifications (format 11 01)
func (lm *LaunchMonitor) HandleSensorNotification(bytesList []string) {
	sensorData, err := ParseSensorData(bytesList)
	if err != nil {
		log.Printf("Error parsing sensor data: %v", err)
		return
	}

	lm.stateManager.SetBallDetected(sensorData.BallDetected)
	lm.stateManager.SetBallReady(sensorData.BallReady)

	ballPosition := &BallPosition{
		X: sensorData.PositionX,
		Y: sensorData.PositionY,
		Z: sensorData.PositionZ,
	}
	lm.stateManager.SetBallPosition(ballPosition)
}

// HandleAlignmentNotification handles alignment/aim notifications (format 11 82)
// TODO: The exact format and byte positions need to be verified by capturing
// Bluetooth traffic from the official Square Golf app
func (lm *LaunchMonitor) HandleAlignmentNotification(bytesList []string) {
	alignmentData, err := ParseAlignmentData(bytesList)
	if err != nil {
		log.Printf("Error parsing alignment data: %v", err)
		return
	}

	// Log for debugging until we confirm the format is correct
	log.Printf("Alignment data received - Aim: %.2f°, Aligned: %v",
		alignmentData.AimAngle, alignmentData.IsAligned)

	// Update alignment state - IsAligning is controlled by the UI
	lm.stateManager.SetAlignmentAngle(alignmentData.AimAngle)
	lm.stateManager.SetIsAligned(alignmentData.IsAligned)
}

// HandleShotBallMetrics handles shot ball metrics notifications (format 11 02 37)
func (lm *LaunchMonitor) HandleShotBallMetrics(bytesList []string) {
	shotMetrics, err := ParseShotBallMetrics(bytesList)
	if err != nil {
		log.Printf("Failed to parse shot metrics data: %v", err)
		return
	}

	// Update state manager with ball metrics
	lastBallMetrics := lm.stateManager.GetLastBallMetrics()

	// Convert RawData to string for comparison and storage
	rawDataStr := ""
	for i, b := range shotMetrics.RawData {
		if i > 0 {
			rawDataStr += " "
		}
		rawDataStr += b
	}

	// Check if this is a new shot by comparing raw data
	var lastRawData string
	if lastBallMetrics != nil {
		lastRawData = strings.Join(lastBallMetrics.RawData, " ")
	}

	if lastBallMetrics == nil || lastRawData != rawDataStr {
		lm.stateManager.SetLastBallMetrics(shotMetrics)

		// Automatically request club metrics after receiving shot metrics
		if lm.bluetoothClient != nil && lm.bluetoothClient.IsConnected() {
			seq := lm.getNextSequence()
			clubMetricsCommand := RequestClubMetricsCommand(seq)

			err := lm.SendCommand(clubMetricsCommand)
			if err != nil {
				log.Printf("Failed to request club metrics: %v", err)
			}
		}
	}
}

// HandleShotClubMetrics handles shot club metrics notifications (format 11 07 0f)
func (lm *LaunchMonitor) HandleShotClubMetrics(bytesList []string) {
	clubMetrics, err := ParseShotClubMetrics(bytesList)
	if err != nil {
		log.Printf("Failed to parse club metrics data: %v", err)
		return
	}

	// Update state manager with club metrics
	lm.stateManager.SetLastClubMetrics(clubMetrics)
}

// SendCommand sends a command to the BLE device
func (lm *LaunchMonitor) SendCommand(commandHex string) error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	commandBytes, err := hex.DecodeString(commandHex)
	if err != nil {
		return fmt.Errorf("invalid hex command: %w", err)
	}

	err = lm.bluetoothClient.WriteCharacteristic(CommandCharUUID, commandBytes)
	if err != nil {
		return fmt.Errorf("error sending command: %w", err)
	}

	return nil
}

// ReadBatteryLevel reads the battery level from the device
func (lm *LaunchMonitor) ReadBatteryLevel() (int, error) {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return 0, fmt.Errorf("not connected to device")
	}

	batteryLevelBytes, err := lm.bluetoothClient.ReadCharacteristic(BatteryLevelCharUUID)
	if err != nil {
		return 0, fmt.Errorf("could not read battery level: %w", err)
	}

	if len(batteryLevelBytes) == 0 {
		return 0, fmt.Errorf("received empty battery level data")
	}

	batteryLevel := int(batteryLevelBytes[0])

	// Update state manager with battery level
	lm.stateManager.SetBatteryLevel(&batteryLevel)

	return batteryLevel, nil
}

// ActivateBallDetection activates ball detection mode
func (lm *LaunchMonitor) ActivateBallDetection() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Get current club, handedness, and spin mode from state
	club := lm.stateManager.GetClub()
	handedness := lm.stateManager.GetHandedness()
	spinMode := lm.stateManager.GetSpinMode()

	// Default to right-handed driver if not set
	if club == nil {
		defaultClub := ClubDriver
		club = &defaultClub
	}
	if handedness == nil {
		defaultHandedness := RightHanded
		handedness = &defaultHandedness
	}
	if spinMode == nil {
		defaultSpinMode := Advanced
		spinMode = &defaultSpinMode
	}

	// Send club command
	seq := lm.getNextSequence()
	clubCommand := ClubCommand(seq, *club, *handedness)

	err := lm.SendCommand(clubCommand)
	if err != nil {
		return fmt.Errorf("failed to send club command: %w", err)
	}

	// Send detect ball command
	seq = lm.getNextSequence()
	detectCommand := DetectBallCommand(seq, Activate, *spinMode)

	err = lm.SendCommand(detectCommand)
	if err != nil {
		return fmt.Errorf("failed to send detect ball command: %w", err)
	}

	return nil
}

// DeactivateBallDetection deactivates ball detection mode
func (lm *LaunchMonitor) DeactivateBallDetection() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Get current spin mode from state
	spinMode := lm.stateManager.GetSpinMode()
	if spinMode == nil {
		defaultSpinMode := Advanced
		spinMode = &defaultSpinMode
	}

	seq := lm.getNextSequence()
	detectCommand := DetectBallCommand(seq, Deactivate, *spinMode)

	err := lm.SendCommand(detectCommand)
	if err != nil {
		return fmt.Errorf("failed to send detect ball command: %w", err)
	}

	return nil
}

// Helper functions

// getNextSequence gets the next sequence number with thread safety
func (lm *LaunchMonitor) getNextSequence() int {
	lm.sequenceMutex.Lock()
	defer lm.sequenceMutex.Unlock()

	seq := lm.sequence
	lm.sequence++
	if lm.sequence > 255 {
		lm.sequence = 0
	}
	return seq
}

// startHeartbeatTask starts the heartbeat task to maintain device connection
func (lm *LaunchMonitor) startHeartbeatTask() {
	lm.heartbeatCancelMu.Lock()
	defer lm.heartbeatCancelMu.Unlock()

	// Cancel any existing heartbeat task
	if lm.heartbeatCancel != nil {
		lm.heartbeatCancel()
		lm.heartbeatCancel = nil
	}

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	lm.heartbeatCancel = cancel

	// Start the heartbeat task in a goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if lm.bluetoothClient != nil && lm.bluetoothClient.IsConnected() {
					seq := lm.getNextSequence()
					command := HeartbeatCommand(seq)
					err := lm.SendCommand(command)
					if err != nil {
						log.Printf("Error sending heartbeat: %v", err)
					}
				}
			}
		}
	}()
}

// ManageHeartbeat initializes and manages the heartbeat communication with the device
func (lm *LaunchMonitor) ManageHeartbeat() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Start the heartbeat task
	lm.startHeartbeatTask()

	// Send initial heartbeat
	seq := lm.getNextSequence()
	heartbeatCommand := HeartbeatCommand(seq)
	err := lm.SendCommand(heartbeatCommand)
	if err != nil {
		return fmt.Errorf("failed to send initial heartbeat: %w", err)
	}

	return nil
}

// SetupNotifications registers the launch monitor's notification handler with the Bluetooth manager
func (lm *LaunchMonitor) SetupNotifications(btManager *BluetoothManager) {
	// Create a closure that adapts the LaunchMonitor's NotificationHandler to match
	// what BluetoothManager expects, while still providing the BluetoothClient
	btManager.SetNotificationHandler(func(uuid string, data []byte) {
		// Call the LaunchMonitor's NotificationHandler with the client
		lm.NotificationHandler(uuid, data)
	})

	// Register pre-disconnect hook to try to deactivate ball detection before disconnection
	btManager.SetPreDisconnectHook(func() {
		if lm.bluetoothClient != nil && lm.bluetoothClient.IsConnected() {
			log.Println("LaunchMonitor: Attempting to deactivate ball detection before disconnection")
			err := lm.DeactivateBallDetection()
			if err != nil {
				log.Printf("LaunchMonitor: Failed to deactivate ball detection: %v", err)
			} else {
				log.Println("LaunchMonitor: Successfully deactivated ball detection")
			}
		}
	})

	// Register for connection status changes to handle disconnects and connection setup
	lm.stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue ConnectionStatus) {
		if newValue == ConnectionStatusConnected && oldValue != ConnectionStatusConnected {
			// When Bluetooth connects, request firmware version
			go func() {
				// Add a small delay to ensure notifications are ready
				time.Sleep(100 * time.Millisecond)
				if err := lm.RequestFirmwareVersion(); err != nil {
					log.Printf("LaunchMonitor: Failed to request firmware version: %v", err)
				}
			}()
		} else if newValue == ConnectionStatusDisconnected {
			// When Bluetooth disconnects, reset ball detection state
			lm.HandleBluetoothDisconnect()
		}
	})

	// Start the heartbeat task to maintain connection
	lm.startHeartbeatTask()
}

// HandleBluetoothDisconnect handles cleanup when Bluetooth disconnects
func (lm *LaunchMonitor) HandleBluetoothDisconnect() {
	log.Println("LaunchMonitor: Bluetooth disconnected - resetting ball detection state")

	// Reset ball detection state in the state manager
	lm.stateManager.SetBallDetected(false)
	lm.stateManager.SetBallReady(false)
	lm.stateManager.SetBallPosition(nil)

	// Stop any heartbeat task
	lm.heartbeatCancelMu.Lock()
	if lm.heartbeatCancel != nil {
		lm.heartbeatCancel()
		lm.heartbeatCancel = nil
	}
	lm.heartbeatCancelMu.Unlock()
}

// StartAlignment starts alignment mode (command 1185 with confirm=0, angle=0)
func (lm *LaunchMonitor) StartAlignment() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Send start alignment command (confirm=0, angle=0)
	seq := lm.getNextSequence()
	command := StartAlignmentCommand(seq)
	err := lm.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to start alignment: %w", err)
	}

	log.Printf("Alignment started with command: %s", command)

	// Update state
	lm.stateManager.SetIsAligning(true)
	return nil
}

// StopAlignment stops alignment mode (command 1185 with confirm=1)
func (lm *LaunchMonitor) StopAlignment() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Get current alignment angle to send as target
	currentAngle := lm.stateManager.GetAlignmentAngle()

	// Send stop alignment command (confirm=1, current angle)
	seq := lm.getNextSequence()
	command := StopAlignmentCommand(seq, currentAngle)
	err := lm.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to stop alignment: %w", err)
	}

	log.Printf("Alignment stopped with command: %s (angle: %.2f°)", command, currentAngle)

	// Update state
	lm.stateManager.SetIsAligning(false)
	lm.stateManager.SetAlignmentAngle(0)
	lm.stateManager.SetIsAligned(false)
	return nil
}

// RequestFirmwareVersion requests the device firmware version
func (lm *LaunchMonitor) RequestFirmwareVersion() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	seq := lm.getNextSequence()
	command := GetOSVersionCommand(seq)
	err := lm.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to request firmware version: %w", err)
	}

	log.Printf("Firmware version requested with command: %s", command)
	return nil
}

// HandleOSVersionNotification handles OS version response notifications (format 11 10)
func (lm *LaunchMonitor) HandleOSVersionNotification(bytesList []string) {
	// Format: 11 10 {major} {minor}
	// Example: 11 10 01 06 = version 1.6
	if len(bytesList) < 4 {
		log.Printf("Invalid OS version notification format, expected at least 4 bytes, got %d", len(bytesList))
		return
	}

	major := bytesList[2]
	minor := bytesList[3]
	version := fmt.Sprintf("%s.%s", major, minor)

	log.Printf("Device firmware version: %s", version)

	// Update state manager
	lm.stateManager.SetFirmwareVersion(&version)
}
