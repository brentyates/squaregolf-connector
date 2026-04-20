package core

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
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

type cmdEntry struct {
	hexCmd string
	errCh  chan error
}

// LaunchMonitor encapsulates the launch monitor functionality
type LaunchMonitor struct {
	stateManager      *StateManager
	sequence          int
	sequenceMutex     sync.Mutex
	heartbeatCancel   context.CancelFunc
	heartbeatCancelMu sync.Mutex
	bluetoothClient   BluetoothClient
	omniClubRetryMu   sync.Mutex
	omniClubRetryGen  int
	omniClubRetried   bool
	detectStateMu     sync.Mutex
	detectModeActive  bool
	omniIdleCount     int
	cmdQueue          chan cmdEntry
	cmdQueueCtx       context.Context
	cmdQueueStop      context.CancelFunc
	cmdQueueOnce      sync.Once
	chargeCancel      context.CancelFunc
	chargeCancelMu    sync.Mutex
	capacitorReady    bool
	capacitorReadyMu  sync.Mutex
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

	// Handle battery level notification (separate characteristic, not deduplicated)
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

	// Battery message on main characteristic (type 0x91 = 145)
	if len(bytesList) >= 2 && bytesList[0] == "91" {
		lm.HandleBatteryMessage(bytesList)
		return
	}

	// Process by byte patterns
	if len(bytesList) >= 2 {
		// Handle alignment notifications (format 11 04)
		if bytesList[0] == "11" && bytesList[1] == "04" {
			lm.HandleAlignmentNotification(bytesList)
			return
		}

		// Sensor notifications (format 11 01)
		if bytesList[0] == "11" && bytesList[1] == "01" {
			lm.HandleSensorNotification(bytesList)
			return
		} else if len(bytesList) >= 3 {
			// Shot Ball Metrics (format 11 02)
			if bytesList[0] == "11" && bytesList[1] == "02" {
				lm.HandleShotBallMetrics(bytesList)
				return
			}
			if bytesList[0] == "11" && bytesList[1] == "03" {
				lm.HandleStatusNotification(bytesList)
				return
			}
			// Capacitor charge status (format 11 06)
			if bytesList[0] == "11" && bytesList[1] == "06" {
				lm.HandleChargeNotification(bytesList)
				return
			}
			// OS Version response (format 11 10)
			if bytesList[0] == "11" && bytesList[1] == "10" {
				lm.HandleOSVersionNotification(bytesList)
				return
			}
			// Shot Club Metrics (format 11 07)
			if bytesList[0] == "11" && bytesList[1] == "07" && len(bytesList) >= 11 {
				lm.HandleShotClubMetrics(bytesList)
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

// HandleAlignmentNotification handles alignment/aim notifications (format 11 04)
func (lm *LaunchMonitor) HandleAlignmentNotification(bytesList []string) {
	alignmentData, err := ParseAlignmentData(bytesList)
	if err != nil {
		log.Printf("Error parsing alignment data: %v", err)
		return
	}

	// Update alignment state - IsAligning is controlled by the UI
	lm.stateManager.SetAlignmentAngle(alignmentData.AimAngle)
	lm.stateManager.SetIsAligned(alignmentData.IsAligned)
}

// HandleShotBallMetrics handles shot ball metrics notifications (format 11 02).
func (lm *LaunchMonitor) HandleShotBallMetrics(bytesList []string) {
	shotMetrics, err := ParseShotBallMetrics(bytesList)
	if err != nil {
		log.Printf("Failed to parse shot metrics data: %v", err)
		return
	}

	if lm.stateManager.GetDeviceType() == DeviceTypeOmni {
		ApplyOmniBallValidityBitmask(shotMetrics)
		lm.applyOmniPutterBallValidityFilter(shotMetrics)
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
			if lm.stateManager.GetDeviceType() == DeviceTypeOmni {
				lm.startOmniClubMetricsRequest()
			}

			seq := lm.getNextSequence()
			clubMetricsCommand := RequestClubMetricsCommand(seq)

			err := lm.SendCommand(clubMetricsCommand)
			if err != nil {
				log.Printf("Failed to request club metrics: %v", err)
			}
		}
	}
}

// HandleShotClubMetrics handles shot club metrics notifications (format 11 07).
func (lm *LaunchMonitor) HandleShotClubMetrics(bytesList []string) {
	var clubMetrics *ClubMetrics
	var err error

	if lm.stateManager.GetDeviceType() == DeviceTypeOmni {
		if len(bytesList) < 19 {
			log.Printf("Ignoring short Omni club metrics packet (got %d bytes, need 19)", len(bytesList))
			return
		}
		clubMetrics, err = ParseOmniShotClubMetrics(bytesList)
		lm.completeOmniClubMetricsRequest()
		lm.applyOmniPutterValidityFilter(clubMetrics)
	} else {
		clubMetrics, err = ParseShotClubMetrics(bytesList)
		lm.applyPutterClubFilter(clubMetrics)
	}

	if err != nil {
		log.Printf("Failed to parse club metrics data: %v", err)
		return
	}

	lm.stateManager.SetLastClubMetrics(clubMetrics)
}

// HandleStatusNotification handles device status notifications (format 11 03 {status}).
func (lm *LaunchMonitor) HandleStatusNotification(bytesList []string) {
	if len(bytesList) < 3 {
		return
	}

	statusIndex := 2
	if lm.stateManager.GetDeviceType() == DeviceTypeOmni {
		if len(bytesList) < 4 {
			return
		}
		if omniHomeGolfStatus, ok := parseHexByteToInt(bytesList[2]); ok {
			lm.stateManager.SetOmniHomeGolfStatus(&omniHomeGolfStatus)
		}
		if omniStatus, ok := parseHexByteToInt(bytesList[3]); ok {
			lm.stateManager.SetOmniStatus(&omniStatus)
		}
		if len(bytesList) >= 5 {
			if omniClubSelection, ok := parseHexByteToInt(bytesList[4]); ok {
				lm.stateManager.SetOmniClubSelection(&omniClubSelection)
			}
		}
		if len(bytesList) >= 8 {
			if omniSensorStatus, ok := parseHexByteToInt(bytesList[7]); ok {
				lm.stateManager.SetOmniSensorStatus(&omniSensorStatus)
			}
		}
		statusIndex = 3
	}

	var status LaunchMonitorStatus
	switch bytesList[statusIndex] {
	case "00":
		status = LaunchMonitorStatusNone
	case "01":
		status = LaunchMonitorStatusIdle
	case "02":
		status = LaunchMonitorStatusInit
	case "03":
		status = LaunchMonitorStatusDetect
	case "04":
		status = LaunchMonitorStatusReady
	case "05":
		status = LaunchMonitorStatusShot
	case "06":
		status = LaunchMonitorStatusDone
	default:
		return
	}

	lm.stateManager.SetLaunchMonitorStatus(status)
	lm.handleOmniStatusRecovery(status)

	if lm.stateManager.GetDeviceType() == DeviceTypeOmni && len(bytesList) >= 7 {
		switch bytesList[6] {
		case "00":
			handedness := RightHanded
			currentHandedness := lm.stateManager.GetHandedness()
			if currentHandedness == nil || *currentHandedness != handedness {
				lm.stateManager.SetHandedness(&handedness)
			}
		case "01":
			handedness := LeftHanded
			currentHandedness := lm.stateManager.GetHandedness()
			if currentHandedness == nil || *currentHandedness != handedness {
				lm.stateManager.SetHandedness(&handedness)
			}
		}
	}
}

func (lm *LaunchMonitor) handleOmniStatusRecovery(status LaunchMonitorStatus) {
	if lm.stateManager.GetDeviceType() != DeviceTypeOmni {
		return
	}

	lm.detectStateMu.Lock()
	defer lm.detectStateMu.Unlock()

	if !lm.detectModeActive {
		lm.omniIdleCount = 0
		return
	}

	if status == LaunchMonitorStatusNone || status == LaunchMonitorStatusIdle {
		lm.omniIdleCount++
		if lm.omniIdleCount <= 1 {
			return
		}
		lm.omniIdleCount = 0
	} else {
		lm.omniIdleCount = 0
		return
	}

	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	spinMode := lm.stateManager.GetSpinMode()
	if spinMode == nil {
		defaultSpinMode := Advanced
		spinMode = &defaultSpinMode
	}

	seq := lm.getNextSequence()
	detectCommand := DetectBallCommand(seq, Activate, *spinMode)
	if err := lm.SendCommand(detectCommand); err != nil {
		log.Printf("LaunchMonitor: Failed to re-arm Omni detect mode after idle status: %v", err)
		return
	}

	log.Println("LaunchMonitor: Re-armed Omni detect mode after repeated idle status")
}

func parseHexByteToInt(value string) (int, bool) {
	parsed, err := strconv.ParseUint(value, 16, 8)
	if err != nil {
		return 0, false
	}

	return int(parsed), true
}

func (lm *LaunchMonitor) startOmniClubMetricsRequest() {
	lm.omniClubRetryMu.Lock()
	lm.omniClubRetryGen++
	gen := lm.omniClubRetryGen
	lm.omniClubRetried = false
	lm.omniClubRetryMu.Unlock()

	time.AfterFunc(1*time.Second, func() {
		lm.handleOmniClubMetricsTimeout(gen)
	})
}

func (lm *LaunchMonitor) completeOmniClubMetricsRequest() {
	lm.omniClubRetryMu.Lock()
	lm.omniClubRetryGen++
	lm.omniClubRetried = false
	lm.omniClubRetryMu.Unlock()
}

func (lm *LaunchMonitor) handleOmniClubMetricsTimeout(gen int) {
	lm.omniClubRetryMu.Lock()
	if gen != lm.omniClubRetryGen {
		lm.omniClubRetryMu.Unlock()
		return
	}

	if !lm.omniClubRetried {
		lm.omniClubRetried = true
		lm.omniClubRetryMu.Unlock()

		if lm.bluetoothClient != nil && lm.bluetoothClient.IsConnected() {
			seq := lm.getNextSequence()
			clubMetricsCommand := RequestClubMetricsCommand(seq)
			if err := lm.SendCommand(clubMetricsCommand); err != nil {
				log.Printf("Failed to retry Omni club metrics request: %v", err)
			}
		}

		time.AfterFunc(1*time.Second, func() {
			lm.handleOmniClubMetricsTimeout(gen)
		})
		return
	}

	lm.omniClubRetryGen++
	lm.omniClubRetried = false
	lm.omniClubRetryMu.Unlock()

	log.Printf("Omni club metrics request timed out twice, applying invalid fallback")
	lm.stateManager.SetLastClubMetrics(&ClubMetrics{})
}

func (lm *LaunchMonitor) applyPutterClubFilter(clubMetrics *ClubMetrics) {
	if clubMetrics == nil {
		return
	}

	club := lm.stateManager.GetClub()
	if club == nil || *club != ClubPutter {
		return
	}

	clubMetrics.IsPathAngleValid = false
	clubMetrics.IsFaceAngleValid = false
	clubMetrics.IsAttackAngleValid = false
	clubMetrics.IsDynamicLoftValid = false
}

func (lm *LaunchMonitor) applyOmniPutterValidityFilter(clubMetrics *ClubMetrics) {
	if clubMetrics == nil {
		return
	}

	club := lm.stateManager.GetClub()
	if club == nil || *club != ClubPutter {
		return
	}

	clubMetrics.IsPathAngleValid = false
	clubMetrics.IsFaceAngleValid = false
	clubMetrics.IsAttackAngleValid = false
	clubMetrics.IsDynamicLoftValid = false
	clubMetrics.IsImpactHorizontalValid = false
	clubMetrics.IsImpactVerticalValid = false
	clubMetrics.IsClubSpeedValid = false
	clubMetrics.IsSmashFactorValid = false
}

func (lm *LaunchMonitor) applyOmniPutterBallValidityFilter(ballMetrics *BallMetrics) {
	if ballMetrics == nil {
		return
	}

	club := lm.stateManager.GetClub()
	if club == nil || *club != ClubPutter {
		return
	}

	ballMetrics.IsTotalSpinValid = false
	ballMetrics.IsSpinAxisValid = false
	ballMetrics.IsBackspinValid = false
	ballMetrics.IsSidespinValid = false
}

func (lm *LaunchMonitor) syncOmniHandedness(handedness *HandednessType) {
	if handedness == nil {
		return
	}
	if lm.stateManager.GetDeviceType() != DeviceTypeOmni {
		return
	}
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	command := OmniSetHandedCommand(lm.getNextSequence(), *handedness)
	if err := lm.SendCommand(command); err != nil {
		log.Printf("LaunchMonitor: Failed to send Omni handedness update: %v", err)
	}
}

func (lm *LaunchMonitor) omniUnitsFromState() (int, int) {
	speedUnit := 0
	if configuredSpeedUnit := lm.stateManager.GetOmniSpeedUnit(); configuredSpeedUnit != nil && *configuredSpeedUnit == "mph" {
		speedUnit = 1
	}

	distanceUnit := 0
	if configuredDistanceUnit := lm.stateManager.GetOmniDistanceUnit(); configuredDistanceUnit != nil {
		switch *configuredDistanceUnit {
		case "mixed":
			distanceUnit = 1
		case "yards":
			distanceUnit = 2
		}
	}

	return speedUnit, distanceUnit
}

func (lm *LaunchMonitor) omniGreenSpeedFromState() int {
	if configuredGreenSpeed := lm.stateManager.GetOmniGreenSpeed(); configuredGreenSpeed != nil {
		if *configuredGreenSpeed < 8 {
			return 0
		}
		if *configuredGreenSpeed > 13 {
			return 5
		}
		return *configuredGreenSpeed - 8
	}

	return 2
}

func (lm *LaunchMonitor) omniCarryAdjustmentFromState() int {
	if configuredCarryAdjustment := lm.stateManager.GetOmniCarryAdjustment(); configuredCarryAdjustment != nil {
		return *configuredCarryAdjustment
	}

	return 0
}

func (lm *LaunchMonitor) syncOmniUnits() {
	if lm.stateManager.GetDeviceType() != DeviceTypeOmni {
		return
	}
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	speedUnit, distanceUnit := lm.omniUnitsFromState()
	command := OmniSetUnitsCommand(lm.getNextSequence(), speedUnit, distanceUnit)
	if err := lm.SendCommand(command); err != nil {
		log.Printf("LaunchMonitor: Failed to send Omni units update: %v", err)
	}
}

func (lm *LaunchMonitor) syncOmniGreenSpeed() {
	if lm.stateManager.GetDeviceType() != DeviceTypeOmni {
		return
	}
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	command := OmniSetGreenSpeedCommand(lm.getNextSequence(), lm.omniGreenSpeedFromState())
	if err := lm.SendCommand(command); err != nil {
		log.Printf("LaunchMonitor: Failed to send Omni green speed update: %v", err)
	}
}

func (lm *LaunchMonitor) syncOmniCarryAdjustment() {
	if lm.stateManager.GetDeviceType() != DeviceTypeOmni {
		return
	}
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	command := OmniSetCarryDistanceAdjustmentCommand(lm.getNextSequence(), lm.omniCarryAdjustmentFromState())
	if err := lm.SendCommand(command); err != nil {
		log.Printf("LaunchMonitor: Failed to send Omni carry adjustment update: %v", err)
	}
}

func (lm *LaunchMonitor) ensureCommandQueue() {
	lm.cmdQueueOnce.Do(func() {
		lm.cmdQueue = make(chan cmdEntry, 32)
		lm.cmdQueueCtx, lm.cmdQueueStop = context.WithCancel(context.Background())
		go lm.drainCommandQueue()
	})
}

func (lm *LaunchMonitor) drainCommandQueue() {
	for {
		select {
		case <-lm.cmdQueueCtx.Done():
			return
		case entry := <-lm.cmdQueue:
			err := lm.writeCommand(entry.hexCmd)
			entry.errCh <- err
			select {
			case <-lm.cmdQueueCtx.Done():
				return
			case <-time.After(150 * time.Millisecond):
			}
		}
	}
}

func (lm *LaunchMonitor) writeCommand(commandHex string) error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	commandBytes, err := hex.DecodeString(commandHex)
	if err != nil {
		return fmt.Errorf("invalid hex command: %w", err)
	}

	return lm.bluetoothClient.WriteCharacteristic(CommandCharUUID, commandBytes)
}

func (lm *LaunchMonitor) stopCommandQueue() {
	if lm.cmdQueueStop != nil {
		lm.cmdQueueStop()
	}
}

// SendCommand sends a command to the BLE device via the rate-limited queue
func (lm *LaunchMonitor) SendCommand(commandHex string) error {
	lm.ensureCommandQueue()

	entry := cmdEntry{
		hexCmd: commandHex,
		errCh:  make(chan error, 1),
	}

	select {
	case lm.cmdQueue <- entry:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("command queue full")
	}

	select {
	case err := <-entry.errCh:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("command execution timed out")
	}
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

	// Send club command (Omni uses adjusted clubSel encoding)
	seq := lm.getNextSequence()
	var clubCommand string
	if lm.stateManager.GetDeviceType() == DeviceTypeOmni {
		clubCommand = OmniClubCommand(seq, *club, *handedness)
	} else {
		clubCommand = ClubCommand(seq, *club, *handedness)
	}

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

	lm.detectStateMu.Lock()
	lm.detectModeActive = true
	lm.omniIdleCount = 0
	lm.detectStateMu.Unlock()

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

	lm.detectStateMu.Lock()
	lm.detectModeActive = false
	lm.omniIdleCount = 0
	lm.detectStateMu.Unlock()

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

	lm.stopHeartbeatTaskLocked()

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
				lm.sendHeartbeatTick()
			}
		}
	}()
}

func (lm *LaunchMonitor) sendHeartbeatTick() {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	seq := lm.getNextSequence()
	command := HeartbeatCommand(seq)
	err := lm.SendCommand(command)
	if err != nil {
		log.Printf("Error sending heartbeat: %v", err)
	}
}

func (lm *LaunchMonitor) stopHeartbeatTaskLocked() {
	if lm.heartbeatCancel != nil {
		lm.heartbeatCancel()
		lm.heartbeatCancel = nil
	}
}

func (lm *LaunchMonitor) stopHeartbeatTask() {
	lm.heartbeatCancelMu.Lock()
	defer lm.heartbeatCancelMu.Unlock()
	lm.stopHeartbeatTaskLocked()
}

// ManageHeartbeat initializes and manages the heartbeat communication with the device
func (lm *LaunchMonitor) ManageHeartbeat() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Start the heartbeat task
	lm.startHeartbeatTask()

	// Send initial heartbeat
	if lm.bluetoothClient != nil && lm.bluetoothClient.IsConnected() {
		seq := lm.getNextSequence()
		heartbeatCommand := HeartbeatCommand(seq)
		err := lm.SendCommand(heartbeatCommand)
		if err != nil {
			return fmt.Errorf("failed to send initial heartbeat: %w", err)
		}
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
			log.Println("LaunchMonitor: Device connected")
			lm.setCapacitorReady(false)
			lm.startChargePolling()
			go lm.sendOmniInitSequence()
		} else if newValue == ConnectionStatusDisconnected {
			// When Bluetooth disconnects, reset ball detection state
			lm.HandleBluetoothDisconnect()
		}
	})

	lm.stateManager.RegisterHandednessCallback(func(oldValue, newValue *HandednessType) {
		if newValue == nil {
			return
		}
		if oldValue != nil && *oldValue == *newValue {
			return
		}
		lm.syncOmniHandedness(newValue)
	})

	lm.stateManager.RegisterOmniSpeedUnitCallback(func(oldValue, newValue *string) {
		if newValue == nil {
			return
		}
		if oldValue != nil && *oldValue == *newValue {
			return
		}
		lm.syncOmniUnits()
	})

	lm.stateManager.RegisterOmniDistanceUnitCallback(func(oldValue, newValue *string) {
		if newValue == nil {
			return
		}
		if oldValue != nil && *oldValue == *newValue {
			return
		}
		lm.syncOmniUnits()
	})

	lm.stateManager.RegisterOmniGreenSpeedCallback(func(oldValue, newValue *int) {
		if newValue == nil {
			return
		}
		if oldValue != nil && *oldValue == *newValue {
			return
		}
		lm.syncOmniGreenSpeed()
	})

	lm.stateManager.RegisterOmniCarryAdjustmentCallback(func(oldValue, newValue *int) {
		if newValue == nil {
			return
		}
		if oldValue != nil && *oldValue == *newValue {
			return
		}
		lm.syncOmniCarryAdjustment()
	})

	// Start the heartbeat task to maintain connection
	lm.startHeartbeatTask()
}

// sendOmniInitSequence sends the Omni-specific configuration commands after connection.
// The Omni requires SetUnits, SetCarryDistanceAdjustment, SetGreenSpeed, and SetHanded
// to be sent after connection with delays between each command.
func (lm *LaunchMonitor) sendOmniInitSequence() {
	if lm.stateManager.GetDeviceType() != DeviceTypeOmni {
		return
	}
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}

	log.Println("LaunchMonitor: Sending Omni init sequence")
	commands := lm.buildOmniInitCommands()

	for _, c := range commands {
		if !lm.bluetoothClient.IsConnected() {
			log.Println("LaunchMonitor: Device disconnected during Omni init, aborting")
			return
		}
		err := lm.SendCommand(c.cmd)
		if err != nil {
			log.Printf("LaunchMonitor: Failed to send Omni %s: %v", c.name, err)
		}
	}

	log.Println("LaunchMonitor: Omni init sequence complete")
}

func (lm *LaunchMonitor) buildOmniInitCommands() []struct {
	name string
	cmd  string
} {
	speedUnit, distanceUnit := lm.omniUnitsFromState()

	commands := []struct {
		name string
		cmd  string
	}{
		{"SetUnits", OmniSetUnitsCommand(lm.getNextSequence(), speedUnit, distanceUnit)},
		{"SetCarryDistanceAdjustment", OmniSetCarryDistanceAdjustmentCommand(lm.getNextSequence(), lm.omniCarryAdjustmentFromState())},
		{"SetGreenSpeed", OmniSetGreenSpeedCommand(lm.getNextSequence(), lm.omniGreenSpeedFromState())},
	}

	handedness := lm.stateManager.GetHandedness()
	if handedness != nil {
		commands = append(commands, struct {
			name string
			cmd  string
		}{"SetHanded", OmniSetHandedCommand(lm.getNextSequence(), *handedness)})
	}

	return commands
}

func (lm *LaunchMonitor) HandleBatteryMessage(bytesList []string) {
	if len(bytesList) < 2 {
		return
	}
	if level, ok := parseHexByteToInt(bytesList[1]); ok {
		lm.stateManager.SetBatteryLevel(&level)
	}
	if len(bytesList) >= 3 {
		if chargingStatus, ok := parseHexByteToInt(bytesList[2]); ok {
			lm.stateManager.SetBatteryCharging(&chargingStatus)
		}
	}
}

func (lm *LaunchMonitor) HandleChargeNotification(bytesList []string) {
	if len(bytesList) < 4 {
		return
	}
	lm.setCapacitorReady(bytesList[3] == "01")
}

func (lm *LaunchMonitor) setCapacitorReady(ready bool) {
	lm.capacitorReadyMu.Lock()
	lm.capacitorReady = ready
	lm.capacitorReadyMu.Unlock()

	if ready {
		lm.stopChargePolling()
	}

	lm.stateManager.SetCapacitorReady(ready)
}

func (lm *LaunchMonitor) GetCapacitorReady() bool {
	lm.capacitorReadyMu.Lock()
	defer lm.capacitorReadyMu.Unlock()
	return lm.capacitorReady
}

func (lm *LaunchMonitor) startChargePolling() {
	lm.chargeCancelMu.Lock()
	defer lm.chargeCancelMu.Unlock()

	lm.stopChargePollingLocked()

	ctx, cancel := context.WithCancel(context.Background())
	lm.chargeCancel = cancel

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		lm.sendChargeCommand()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if lm.GetCapacitorReady() {
					return
				}
				lm.sendChargeCommand()
			}
		}
	}()
}

func (lm *LaunchMonitor) sendChargeCommand() {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return
	}
	seq := lm.getNextSequence()
	if err := lm.SendCommand(GetChargeCommand(seq)); err != nil {
		log.Printf("LaunchMonitor: Failed to send GetCharge: %v", err)
	}
}

func (lm *LaunchMonitor) stopChargePolling() {
	lm.chargeCancelMu.Lock()
	defer lm.chargeCancelMu.Unlock()
	lm.stopChargePollingLocked()
}

func (lm *LaunchMonitor) stopChargePollingLocked() {
	if lm.chargeCancel != nil {
		lm.chargeCancel()
		lm.chargeCancel = nil
	}
}

// HandleBluetoothDisconnect handles cleanup when Bluetooth disconnects
func (lm *LaunchMonitor) HandleBluetoothDisconnect() {
	log.Println("LaunchMonitor: Bluetooth disconnected - resetting ball detection state")

	// Reset ball detection state in the state manager
	lm.stateManager.SetBallDetected(false)
	lm.stateManager.SetBallReady(false)
	lm.stateManager.SetBallPosition(nil)
	lm.stateManager.SetLaunchMonitorStatus(LaunchMonitorStatusNone)
	lm.stateManager.SetDeviceType(DeviceTypeUnknown)
	lm.stateManager.SetOmniHomeGolfStatus(nil)
	lm.stateManager.SetOmniStatus(nil)
	lm.stateManager.SetOmniClubSelection(nil)
	lm.stateManager.SetOmniSensorStatus(nil)
	lm.detectStateMu.Lock()
	lm.detectModeActive = false
	lm.omniIdleCount = 0
	lm.detectStateMu.Unlock()

	lm.setCapacitorReady(false)
	lm.stopChargePolling()

	// Stop any heartbeat task
	lm.stopHeartbeatTask()

	lm.stopCommandQueue()
	lm.cmdQueueOnce = sync.Once{}
}

// StartAlignment starts alignment mode
func (lm *LaunchMonitor) StartAlignment() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Get handedness, default to RightHanded if not set
	handednessPtr := lm.stateManager.GetHandedness()
	handedness := RightHanded
	if handednessPtr != nil {
		handedness = *handednessPtr
	}

	// Check if already in alignment mode
	// If so, only send commands if this is NOT a duplicate call from navigation
	// We can tell it's a handedness change request because the frontend always
	// updates handedness state before calling StartAlignment
	if lm.stateManager.GetIsAligning() {
		// Already aligning - this is likely just navigation returning to the screen
		// Don't send duplicate commands
		return nil
	}

	// First, send club command with alignment stick (clubSel=0x08)
	// This puts the device in alignment mode (Windows app Awake method)
	seq := lm.getNextSequence()

	command := ClubCommand(seq, ClubAlignmentStick, handedness)
	err := lm.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to start alignment: %w", err)
	}

	time.Sleep(1 * time.Second)

	// Activate ball detection mode 2 to turn on the red LED
	detectSeq := lm.getNextSequence()
	detectCmd := DetectBallCommand(detectSeq, ActivateAlignmentMode, Advanced)
	err = lm.SendCommand(detectCmd)
	if err != nil {
		return fmt.Errorf("failed to activate ball detection: %w", err)
	}

	time.Sleep(100 * time.Millisecond)

	lm.stateManager.SetIsAligning(true)
	return nil
}

// StopAlignment stops alignment mode and saves calibration (OK button)
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

	// Update state
	lm.stateManager.SetIsAligning(false)
	lm.stateManager.SetAlignmentAngle(0)
	lm.stateManager.SetIsAligned(false)
	return nil
}

// CancelAlignment cancels alignment mode without saving calibration (Cancel button)
func (lm *LaunchMonitor) CancelAlignment() error {
	if lm.bluetoothClient == nil || !lm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	// Get current alignment angle to send with cancel
	currentAngle := lm.stateManager.GetAlignmentAngle()

	// Send cancel alignment command (confirm=0, current angle)
	seq := lm.getNextSequence()
	command := CancelAlignmentCommand(seq, currentAngle)
	err := lm.SendCommand(command)
	if err != nil {
		return fmt.Errorf("failed to cancel alignment: %w", err)
	}

	// Update state
	lm.stateManager.SetIsAligning(false)
	lm.stateManager.SetAlignmentAngle(0)
	lm.stateManager.SetIsAligned(false)
	return nil
}

// RequestFirmwareVersion requests the device firmware version
func (lm *LaunchMonitor) RequestFirmwareVersion() error {
	log.Printf("LaunchMonitor: RequestFirmwareVersion called")

	if lm.bluetoothClient == nil {
		log.Printf("LaunchMonitor: bluetoothClient is nil")
		return fmt.Errorf("bluetoothClient is nil")
	}

	if !lm.bluetoothClient.IsConnected() {
		log.Printf("LaunchMonitor: device not connected")
		return fmt.Errorf("not connected to device")
	}

	seq := lm.getNextSequence()
	command := GetOSVersionCommand(seq)
	log.Printf("LaunchMonitor: Sending firmware version request command: %v", command)

	err := lm.SendCommand(command)
	if err != nil {
		log.Printf("LaunchMonitor: Failed to send firmware version command: %v", err)
		return fmt.Errorf("failed to request firmware version: %w", err)
	}

	log.Printf("LaunchMonitor: Firmware version request sent successfully")
	return nil
}

// HandleOSVersionNotification handles OS version response notifications (format 11 10)
func (lm *LaunchMonitor) HandleOSVersionNotification(bytesList []string) {
	// Format: 11 10 {major} {minor}
	// Example: 11 10 01 09 = version 1.9
	// The bytes are hex strings representing decimal values
	log.Printf("Raw OS version bytes: %v (len=%d)", bytesList, len(bytesList))

	if len(bytesList) < 4 {
		log.Printf("Invalid OS version notification format, expected at least 4 bytes, got %d", len(bytesList))
		return
	}

	// Parse hex strings as hex values to get decimal
	// bytesList[2] is major version (hex string like "01" = decimal 1)
	// bytesList[3] is minor version (hex string like "09" = decimal 9)
	major, err := strconv.ParseInt(bytesList[2], 16, 64)
	if err != nil {
		log.Printf("Error parsing major version from '%s': %v", bytesList[2], err)
		return
	}

	minor, err := strconv.ParseInt(bytesList[3], 16, 64)
	if err != nil {
		log.Printf("Error parsing minor version from '%s': %v", bytesList[3], err)
		return
	}

	version := fmt.Sprintf("%d.%d", major, minor)

	log.Printf("Device firmware version: %s (major=%d, minor=%d)", version, major, minor)

	// Update state manager
	lm.stateManager.SetFirmwareVersion(&version)
}
