package core

import (
	"bytes"
	"sync"
	"testing"
)

func resetSingletonsForTest(t *testing.T) {
	launchMonitorOnce = sync.Once{}
	launchMonitorInstance = nil
	bluetoothOnce = sync.Once{}
	bluetoothInstance = nil
	once = sync.Once{}
	instance = nil
}

func newTestLaunchMonitor(t *testing.T) (*StateManager, *LaunchMonitor, *MockBluetoothClient, *BluetoothManager) {
	resetSingletonsForTest(t)
	sm := GetInstance()
	btManager := NewBluetoothManager(sm)
	mockClient := NewMockBluetoothClient()
	btManager.SetClient(mockClient)
	lm := NewLaunchMonitor(sm, btManager)
	lm.UpdateBluetoothClient(mockClient)
	return sm, lm, mockClient, btManager
}

func TestNewLaunchMonitor(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)

	if lm == nil || lm.stateManager != sm || lm.bluetoothClient != mockClient || lm.sequence != 0 {
		t.Error("LaunchMonitor not properly initialized")
	}
}

func TestNotificationHandler_BatteryLevel(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test battery level notification
	batteryData := []byte{75} // 75% battery
	lm.NotificationHandler(BatteryLevelCharUUID, batteryData)

	batteryLevel := sm.GetBatteryLevel()
	if batteryLevel == nil || *batteryLevel != 75 {
		t.Errorf("Expected battery level to be 75, got %v", batteryLevel)
	}
}

func TestNotificationHandler_EmptyData(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Set initial battery level
	initialBatteryLevel := 75
	sm.SetBatteryLevel(&initialBatteryLevel)

	// Test empty notification data
	lm.NotificationHandler(BatteryLevelCharUUID, []byte{})

	// Verify battery level remains unchanged
	batteryLevel := sm.GetBatteryLevel()
	if batteryLevel == nil || *batteryLevel != initialBatteryLevel {
		t.Errorf("Expected battery level to remain %d, got %v", initialBatteryLevel, batteryLevel)
	}
}

func TestGetNextSequence(t *testing.T) {
	_, lm, _, _ := newTestLaunchMonitor(t)

	// Test sequence increment
	seq1 := lm.getNextSequence()
	seq2 := lm.getNextSequence()
	seq3 := lm.getNextSequence()

	if seq1 != 0 || seq2 != 1 || seq3 != 2 {
		t.Errorf("Expected sequence numbers 0,1,2, got %d,%d,%d", seq1, seq2, seq3)
	}

	// Test sequence wrap around
	lm.sequence = 255
	seq4 := lm.getNextSequence()
	if seq4 != 255 {
		t.Errorf("Expected sequence 255, got %d", seq4)
	}
	seq5 := lm.getNextSequence()
	if seq5 != 0 {
		t.Errorf("Expected sequence to wrap to 0, got %d", seq5)
	}
}

func TestSendCommand(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test successful command send
	command := "1101"
	err := lm.SendCommand(command)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockClient.writeCalled {
		t.Error("Expected WriteCharacteristic to be called")
	}
	if mockClient.lastWriteUUID != CommandCharUUID {
		t.Errorf("Expected UUID %s, got %s", CommandCharUUID, mockClient.lastWriteUUID)
	}

	// Test command send when not connected
	mockClient.connected = false
	err = lm.SendCommand(command)
	if err == nil {
		t.Error("Expected error when not connected")
	}
}

func TestReadBatteryLevel(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true
	mockClient.readReturnData = []byte{85} // 85% battery

	// Test successful battery read
	batteryLevel, err := lm.ReadBatteryLevel()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if batteryLevel != 85 {
		t.Errorf("Expected battery level 85, got %d", batteryLevel)
	}
	if !mockClient.readCalled {
		t.Error("Expected ReadCharacteristic to be called")
	}
	if mockClient.lastReadUUID != BatteryLevelCharUUID {
		t.Errorf("Expected UUID %s, got %s", BatteryLevelCharUUID, mockClient.lastReadUUID)
	}

	// Test battery read when not connected
	mockClient.connected = false
	batteryLevel, err = lm.ReadBatteryLevel()
	if err == nil {
		t.Error("Expected error when not connected")
	}
}

func TestActivateBallDetection(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Set default club and handedness
	club := ClubDriver
	handedness := RightHanded
	sm.SetClub(&club)
	sm.SetHandedness(&handedness)

	// Test successful activation
	err := lm.ActivateBallDetection()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify write history
	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 2 {
		t.Fatalf("Expected 2 writes, got %d", len(writeHistory))
	}

	// Verify club command (first write)
	clubWrite := writeHistory[0]
	if clubWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for club command, got %s", CommandCharUUID, clubWrite.UUID)
	}
	expectedClubCmd := []byte{0x11, 0x82, 0x00, 0x02, 0x04, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(clubWrite.Data, expectedClubCmd) {
		t.Errorf("Expected club command %x, got %x", expectedClubCmd, clubWrite.Data)
	}

	// Verify detect ball command (second write)
	detectWrite := writeHistory[1]
	if detectWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for detect command, got %s", CommandCharUUID, detectWrite.UUID)
	}
	expectedDetectCmd := []byte{0x11, 0x81, 0x01, 0x01, 0x11, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(detectWrite.Data, expectedDetectCmd) {
		t.Errorf("Expected detect command %x, got %x", expectedDetectCmd, detectWrite.Data)
	}

	// Test activation when not connected
	mockClient.connected = false
	err = lm.ActivateBallDetection()
	if err == nil {
		t.Error("Expected error when not connected")
	}
}

func TestDeactivateBallDetection(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test successful deactivation
	err := lm.DeactivateBallDetection()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockClient.writeCalled {
		t.Error("Expected WriteCharacteristic to be called")
	}

	// Test deactivation when not connected
	mockClient.connected = false
	err = lm.DeactivateBallDetection()
	if err == nil {
		t.Error("Expected error when not connected")
	}
}

func TestManageHeartbeat(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test successful heartbeat management
	err := lm.ManageHeartbeat()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mockClient.writeCalled {
		t.Error("Expected WriteCharacteristic to be called")
	}

	// Test heartbeat management when not connected
	mockClient.connected = false
	err = lm.ManageHeartbeat()
	if err == nil {
		t.Error("Expected error when not connected")
	}
}

func TestNotificationHandler_SensorData(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test sensor notification data
	// Format: 11 01 (sensor data) with ball detected, ready, and position
	sensorData := []byte{
		0x11, 0x01, // Header
		0x00,                   // Padding
		0x01,                   // Ball ready
		0x01,                   // Ball detected
		0x0A, 0x00, 0x00, 0x00, // Position X (10) - little-endian
		0x14, 0x00, 0x00, 0x00, // Position Y (20) - little-endian
		0x1E, 0x00, 0x00, 0x00, // Position Z (30) - little-endian
		0x00, 0x00, 0x00, 0x00, // Additional padding
	}

	lm.NotificationHandler(NotificationCharUUID, sensorData)

	// Verify state changes
	if !sm.GetBallDetected() {
		t.Error("Expected ball detected to be true")
	}
	if !sm.GetBallReady() {
		t.Error("Expected ball ready to be true")
	}

	pos := sm.GetBallPosition()
	if pos == nil {
		t.Fatal("Expected ball position to be set")
	}
	if pos.X != 10 || pos.Y != 20 || pos.Z != 30 {
		t.Errorf("Expected position (10,20,30), got (%d,%d,%d)", pos.X, pos.Y, pos.Z)
	}
}

func TestNotificationHandler_ShotBallMetrics(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test shot ball metrics notification
	// Format: 11 02 xx (shot metrics)
	shotData := []byte{
		0x11, 0x02, 0x37, // Header metadata byte
		0x32, 0x00, // Ball speed (50 = 0.5 m/s)
		0x14, 0x00, // Vertical angle (20 = 0.2 degrees)
		0x0A, 0x00, // Horizontal angle (10 = 0.1 degrees)
		0x28, 0x00, // Total spin (40 rpm)
		0x1E, 0x00, // Spin axis (30 = 0.3)
		0x32, 0x00, // Back spin (50 rpm)
		0x1E, 0x00, // Side spin (30 rpm)
	}

	lm.NotificationHandler("", shotData)

	// Verify ball metrics were updated
	metrics := sm.GetLastBallMetrics()
	if metrics == nil {
		t.Fatal("Expected ball metrics to be set")
	}

	if metrics.BallSpeedMPS != 0.5 {
		t.Errorf("Expected ball speed 0.5 m/s, got %v", metrics.BallSpeedMPS)
	}
	if metrics.VerticalAngle != 0.2 {
		t.Errorf("Expected vertical angle 0.2, got %v", metrics.VerticalAngle)
	}
	if metrics.HorizontalAngle != 0.1 {
		t.Errorf("Expected horizontal angle 0.1, got %v", metrics.HorizontalAngle)
	}
	if metrics.TotalspinRPM != 40 {
		t.Errorf("Expected total spin 40 rpm, got %v", metrics.TotalspinRPM)
	}
	if metrics.SpinAxis != 0.3 {
		t.Errorf("Expected spin axis 0.3, got %v", metrics.SpinAxis)
	}
	if metrics.BackspinRPM != 50 {
		t.Errorf("Expected back spin 50 rpm, got %v", metrics.BackspinRPM)
	}
	if metrics.SidespinRPM != 30 {
		t.Errorf("Expected side spin 30 rpm, got %v", metrics.SidespinRPM)
	}
	if !metrics.IsBallSpeedValid || !metrics.IsTotalSpinValid || !metrics.IsSpinAxisValid || !metrics.IsBackspinValid || !metrics.IsSidespinValid {
		t.Error("Expected all ball metric validity flags to be true")
	}

	// Verify club metrics request was made
	if !mockClient.writeCalled {
		t.Error("Expected club metrics request to be sent")
	}
}

func TestNotificationHandler_ShotClubMetrics(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)

	// Test shot club metrics notification with a non-0f metadata byte.
	clubData := []byte{
		0x11, 0x07, 0x0d, // Header
		0x32, 0x00, // Path angle (50 = 0.5 degrees)
		0x14, 0x00, // Face angle (20 = 0.2 degrees)
		0x0A, 0x00, // Attack angle (10 = 0.1 degrees)
		0x28, 0x00, // Dynamic loft angle (40 = 0.4 degrees)
	}

	lm.NotificationHandler("", clubData)

	// Verify club metrics were updated
	metrics := sm.GetLastClubMetrics()
	if metrics == nil {
		t.Fatal("Expected club metrics to be set")
	}

	if metrics.PathAngle != 0.5 {
		t.Errorf("Expected path angle 0.5, got %v", metrics.PathAngle)
	}
	if metrics.FaceAngle != 0.2 {
		t.Errorf("Expected face angle 0.2, got %v", metrics.FaceAngle)
	}
	if metrics.AttackAngle != 0.1 {
		t.Errorf("Expected attack angle 0.1, got %v", metrics.AttackAngle)
	}
	if metrics.DynamicLoftAngle != 0.4 {
		t.Errorf("Expected dynamic loft angle 0.4, got %v", metrics.DynamicLoftAngle)
	}
	if !metrics.IsPathAngleValid || !metrics.IsFaceAngleValid || !metrics.IsAttackAngleValid || !metrics.IsDynamicLoftValid {
		t.Error("Expected all club metric validity flags to be true")
	}
}

func TestSendHeartbeatTick(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	lm.sendHeartbeatTick()

	// Verify heartbeat was sent
	if !mockClient.writeCalled || mockClient.lastWriteUUID != CommandCharUUID {
		t.Error("Expected heartbeat to be sent")
	}
	if mockClient.writeCount != 1 {
		t.Errorf("Expected exactly one heartbeat write, got %d", mockClient.writeCount)
	}
}

func TestStopHeartbeatTask(t *testing.T) {
	_, lm, _, _ := newTestLaunchMonitor(t)

	lm.startHeartbeatTask()
	firstCancel := lm.heartbeatCancel
	if firstCancel == nil {
		t.Fatal("Expected heartbeat cancel function to be set")
	}

	lm.stopHeartbeatTask()
	if lm.heartbeatCancel != nil {
		t.Fatal("Expected heartbeat cancel function to be cleared")
	}
}

func TestStartHeartbeatTask_ReplacesCancelFunc(t *testing.T) {
	_, lm, _, _ := newTestLaunchMonitor(t)

	lm.startHeartbeatTask()
	if lm.heartbeatCancel == nil {
		t.Fatal("Expected first heartbeat cancel function to be set")
	}

	lm.startHeartbeatTask()
	if lm.heartbeatCancel == nil {
		t.Fatal("Expected second heartbeat cancel function to be set")
	}

	lm.stopHeartbeatTask()
	if lm.heartbeatCancel != nil {
		t.Fatal("Expected heartbeat cancel function to be cleared after stop")
	}
}

func TestNotificationHandler_Status(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)

	statusData := []byte{0x11, 0x03, 0x04}
	lm.NotificationHandler("", statusData)

	if sm.GetLaunchMonitorStatus() != LaunchMonitorStatusReady {
		t.Fatalf("Expected launch monitor status %q, got %q", LaunchMonitorStatusReady, sm.GetLaunchMonitorStatus())
	}
}

func TestNotificationHandler_Status_OmniUsesOmniStatusByte(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)
	sm.SetDeviceType(DeviceTypeOmni)

	statusData := []byte{0x11, 0x03, 0x01, 0x04, 0x07, 0x00, 0x01, 0x02}
	lm.NotificationHandler("", statusData)

	if sm.GetLaunchMonitorStatus() != LaunchMonitorStatusReady {
		t.Fatalf("Expected Omni launch monitor status %q, got %q", LaunchMonitorStatusReady, sm.GetLaunchMonitorStatus())
	}

	handedness := sm.GetHandedness()
	if handedness == nil || *handedness != LeftHanded {
		t.Fatalf("Expected handedness to update to %q, got %v", LeftHanded, handedness)
	}

	omniHomeGolfStatus := sm.GetOmniHomeGolfStatus()
	if omniHomeGolfStatus == nil || *omniHomeGolfStatus != 1 {
		t.Fatalf("Expected Omni HomeGolf status 1, got %v", omniHomeGolfStatus)
	}

	omniStatus := sm.GetOmniStatus()
	if omniStatus == nil || *omniStatus != 4 {
		t.Fatalf("Expected Omni status 4, got %v", omniStatus)
	}

	omniClubSelection := sm.GetOmniClubSelection()
	if omniClubSelection == nil || *omniClubSelection != 7 {
		t.Fatalf("Expected Omni club selection 7, got %v", omniClubSelection)
	}

	omniSensorStatus := sm.GetOmniSensorStatus()
	if omniSensorStatus == nil || *omniSensorStatus != 2 {
		t.Fatalf("Expected Omni sensor status 2, got %v", omniSensorStatus)
	}
}

func TestSetupNotifications(t *testing.T) {
	sm, lm, mockClient, btManager := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Connect the mock client
	err := mockClient.Connect("", "")
	if err != nil {
		t.Fatalf("Failed to connect mock client: %v", err)
	}

	// Setup notifications
	lm.SetupNotifications(btManager)

	// Create test data
	testData := []struct {
		name     string
		uuid     string
		data     []byte
		checkFn  func() bool
		expected interface{}
	}{
		{
			name: "Battery Level",
			uuid: BatteryLevelCharUUID,
			data: []byte{75}, // 75% battery
			checkFn: func() bool {
				level := sm.GetBatteryLevel()
				return level != nil && *level == 75
			},
			expected: 75,
		},
		{
			name: "Empty Battery Level",
			uuid: BatteryLevelCharUUID,
			data: []byte{}, // Empty data
			checkFn: func() bool {
				level := sm.GetBatteryLevel()
				return level != nil && *level == 75 // Should keep previous value
			},
			expected: 75,
		},
		{
			name: "Invalid UUID",
			uuid: "invalid-uuid",
			data: []byte{75}, // 75% battery
			checkFn: func() bool {
				level := sm.GetBatteryLevel()
				return level != nil && *level == 75 // Should keep previous value
			},
			expected: 75,
		},
		{
			name: "Ball Detection",
			uuid: "",
			data: []byte{
				0x11, 0x01, // Header
				0x00,                   // Padding
				0x01,                   // Ball ready
				0x01,                   // Ball detected
				0x0A, 0x00, 0x00, 0x00, // Position X (10) - little-endian
				0x14, 0x00, 0x00, 0x00, // Position Y (20) - little-endian
				0x1E, 0x00, 0x00, 0x00, // Position Z (30) - little-endian
				0x00, 0x00, 0x00, 0x00, // Additional padding
			},
			checkFn: func() bool {
				pos := sm.GetBallPosition()
				return sm.GetBallDetected() && sm.GetBallReady() &&
					pos != nil && pos.X == 10 && pos.Y == 20 && pos.Z == 30
			},
			expected: true,
		},
		{
			name: "Invalid Ball Detection Header",
			uuid: "",
			data: []byte{
				0x12, 0x01, // Wrong header
				0x01,                   // Ball detected
				0x01,                   // Ball ready
				0x00, 0x00, 0x00, 0x0A, // Position X
				0x00, 0x00, 0x00, 0x14, // Position Y
				0x00, 0x00, 0x00, 0x1E, // Position Z
				0x00, 0x00, 0x00, 0x00, // Additional padding
			},
			checkFn: func() bool {
				// State should not change
				return !sm.GetBallDetected() && !sm.GetBallReady()
			},
			expected: false,
		},
	}

	// Run tests
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state before each test
			sm.SetBatteryLevel(nil)
			sm.SetBallDetected(false)
			sm.SetBallReady(false)
			sm.SetBallPosition(nil)
			sm.SetLastBallMetrics(nil)
			sm.SetLastClubMetrics(nil)

			// Set initial battery level for tests that expect it
			if tt.name == "Empty Battery Level" || tt.name == "Invalid UUID" {
				initialBatteryLevel := 75
				sm.SetBatteryLevel(&initialBatteryLevel)
			}

			// Call the notification handler through the manager
			btManager.notificationHandler(tt.uuid, tt.data)

			// Verify the state was updated correctly
			if !tt.checkFn() {
				t.Errorf("Expected state update for %s with value %v", tt.name, tt.expected)
			}
		})
	}

	// Test disconnected state
	mockClient.Disconnect()
	initialBatteryLevel := 75
	sm.SetBatteryLevel(&initialBatteryLevel)

	// Send a battery level notification while disconnected
	btManager.notificationHandler(BatteryLevelCharUUID, []byte{50})
	level := sm.GetBatteryLevel()
	if level == nil || *level != 50 {
		t.Errorf("Battery level not updated when disconnected: expected 50, got %v", *level)
	}

	// Reconnect and verify notifications work
	err = mockClient.Connect("", "")
	if err != nil {
		t.Fatalf("Failed to reconnect mock client: %v", err)
	}

	// Send a new battery level notification
	btManager.notificationHandler(BatteryLevelCharUUID, []byte{60})
	level = sm.GetBatteryLevel()
	if level == nil || *level != 60 {
		t.Errorf("Battery level not updated after reconnect: expected 60, got %v", *level)
	}

	// Cleanup
	mockClient.Disconnect()
}

func TestNotificationHandler_InvalidSensorData(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)

	// Test invalid sensor data (too short)
	invalidData := []byte{
		0x11, 0x01, // Header
		0x01, // Ball detected
		0x01, // Ball ready
		// Missing position data
	}

	lm.NotificationHandler("", invalidData)

	// Verify state remains unchanged
	if sm.GetBallDetected() {
		t.Error("Expected ball detected to remain false")
	}
	if sm.GetBallReady() {
		t.Error("Expected ball ready to remain false")
	}
	if sm.GetBallPosition() != nil {
		t.Error("Expected ball position to remain nil")
	}
}

func TestNotificationHandler_InvalidBallMetrics(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test invalid ball metrics data (too short)
	invalidData := []byte{
		0x11, 0x02, 0x37, // Header for full shot
		0x32, 0x00, // Ball speed
		// Missing remaining metrics
	}

	lm.NotificationHandler("", invalidData)

	// Verify state remains unchanged
	metrics := sm.GetLastBallMetrics()
	if metrics != nil {
		t.Error("Expected ball metrics to remain nil")
	}

	// Verify no club metrics request was made
	if mockClient.writeCalled {
		t.Error("Expected no club metrics request to be sent")
	}
}

func TestNotificationHandler_InvalidClubMetrics(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)

	// Test invalid club metrics data (too short)
	invalidData := []byte{
		0x11, 0x07, 0x0f, // Header
		0x32, 0x00, // Path angle
		// Missing remaining metrics
	}

	lm.NotificationHandler("", invalidData)

	// Verify state remains unchanged
	metrics := sm.GetLastClubMetrics()
	if metrics != nil {
		t.Error("Expected club metrics to remain nil")
	}
}

func TestActivateBallDetection_Putter(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Set putter and right-handed
	club := ClubPutter
	handedness := RightHanded
	sm.SetClub(&club)
	sm.SetHandedness(&handedness)

	// Test successful activation
	err := lm.ActivateBallDetection()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify write history
	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 2 {
		t.Fatalf("Expected 2 writes, got %d", len(writeHistory))
	}

	// Verify club command (first write)
	clubWrite := writeHistory[0]
	if clubWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for club command, got %s", CommandCharUUID, clubWrite.UUID)
	}
	expectedClubCmd := []byte{0x11, 0x82, 0x00, 0x01, 0x07, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(clubWrite.Data, expectedClubCmd) {
		t.Errorf("Expected club command %x, got %x", expectedClubCmd, clubWrite.Data)
	}

	// Verify detect ball command (second write)
	detectWrite := writeHistory[1]
	if detectWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for detect command, got %s", CommandCharUUID, detectWrite.UUID)
	}
	expectedDetectCmd := []byte{0x11, 0x81, 0x01, 0x01, 0x11, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(detectWrite.Data, expectedDetectCmd) {
		t.Errorf("Expected detect command %x, got %x", expectedDetectCmd, detectWrite.Data)
	}
}

func TestActivateBallDetection_LeftHanded(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Set driver and left-handed
	club := ClubDriver
	handedness := LeftHanded
	sm.SetClub(&club)
	sm.SetHandedness(&handedness)

	// Test successful activation
	err := lm.ActivateBallDetection()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify write history
	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 2 {
		t.Fatalf("Expected 2 writes, got %d", len(writeHistory))
	}

	// Verify club command (first write)
	clubWrite := writeHistory[0]
	if clubWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for club command, got %s", CommandCharUUID, clubWrite.UUID)
	}
	expectedClubCmd := []byte{0x11, 0x82, 0x00, 0x02, 0x04, 0x01, 0x00, 0x00, 0x00}
	if !bytes.Equal(clubWrite.Data, expectedClubCmd) {
		t.Errorf("Expected club command %x, got %x", expectedClubCmd, clubWrite.Data)
	}

	// Verify detect ball command (second write)
	detectWrite := writeHistory[1]
	if detectWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for detect command, got %s", CommandCharUUID, detectWrite.UUID)
	}
	expectedDetectCmd := []byte{0x11, 0x81, 0x01, 0x01, 0x11, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(detectWrite.Data, expectedDetectCmd) {
		t.Errorf("Expected detect command %x, got %x", expectedDetectCmd, detectWrite.Data)
	}
}

func TestDeactivateBallDetection_ExactBytes(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	// Test successful deactivation
	err := lm.DeactivateBallDetection()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify write history
	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected 1 write, got %d", len(writeHistory))
	}

	// Verify deactivate command
	deactivateWrite := writeHistory[0]
	if deactivateWrite.UUID != CommandCharUUID {
		t.Errorf("Expected UUID %s for deactivate command, got %s", CommandCharUUID, deactivateWrite.UUID)
	}
	expectedCmd := []byte{0x11, 0x81, 0x00, 0x00, 0x11, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(deactivateWrite.Data, expectedCmd) {
		t.Errorf("Expected deactivate command %x, got %x", expectedCmd, deactivateWrite.Data)
	}
}

func TestSwingStickCommand_ExactBytes(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	testCases := []struct {
		name       string
		club       ClubType
		handedness HandednessType
		expected   []byte
	}{
		{
			name:       "Putter RightHanded",
			club:       ClubPutter,
			handedness: RightHanded,
			expected:   []byte{0x11, 0x82, 0x00, 0x01, 0x03, 0x00, 0x00, 0x00},
		},
		{
			name:       "Driver LeftHanded",
			club:       ClubDriver,
			handedness: LeftHanded,
			expected:   []byte{0x11, 0x82, 0x00, 0x02, 0x02, 0x01, 0x00, 0x00},
		},
		{
			name:       "Iron7 RightHanded",
			club:       ClubIron7,
			handedness: RightHanded,
			expected:   []byte{0x11, 0x82, 0x00, 0x07, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.ClearWriteHistory()
			command := SwingStickCommand(0, tc.club, tc.handedness)
			err := lm.SendCommand(command)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			writeHistory := mockClient.GetWriteHistory()
			if len(writeHistory) != 1 {
				t.Fatalf("Expected 1 write, got %d", len(writeHistory))
			}

			write := writeHistory[0]
			if write.UUID != CommandCharUUID {
				t.Errorf("Expected UUID %s, got %s", CommandCharUUID, write.UUID)
			}
			if !bytes.Equal(write.Data, tc.expected) {
				t.Errorf("Expected command %x, got %x", tc.expected, write.Data)
			}
		})
	}
}

func TestAlignmentStickCommand_ExactBytes(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	testCases := []struct {
		name       string
		handedness HandednessType
		expected   []byte
	}{
		{
			name:       "RightHanded",
			handedness: RightHanded,
			expected:   []byte{0x11, 0x82, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:       "LeftHanded",
			handedness: LeftHanded,
			expected:   []byte{0x11, 0x82, 0x00, 0x00, 0x08, 0x01, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.ClearWriteHistory()
			command := ClubCommand(0, ClubAlignmentStick, tc.handedness)
			err := lm.SendCommand(command)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			writeHistory := mockClient.GetWriteHistory()
			if len(writeHistory) != 1 {
				t.Fatalf("Expected 1 write, got %d", len(writeHistory))
			}

			write := writeHistory[0]
			if write.UUID != CommandCharUUID {
				t.Errorf("Expected UUID %s, got %s", CommandCharUUID, write.UUID)
			}
			if !bytes.Equal(write.Data, tc.expected) {
				t.Errorf("Expected command %x, got %x", tc.expected, write.Data)
			}
		})
	}
}

func TestRequestClubMetricsCommand_ExactBytes(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	testCases := []struct {
		name     string
		sequence int
		expected []byte
	}{
		{
			name:     "Sequence 0",
			sequence: 0,
			expected: []byte{0x11, 0x87, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "Sequence 15",
			sequence: 15,
			expected: []byte{0x11, 0x87, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "Sequence 255",
			sequence: 255,
			expected: []byte{0x11, 0x87, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.ClearWriteHistory()
			command := RequestClubMetricsCommand(tc.sequence)
			err := lm.SendCommand(command)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			writeHistory := mockClient.GetWriteHistory()
			if len(writeHistory) != 1 {
				t.Fatalf("Expected 1 write, got %d", len(writeHistory))
			}

			write := writeHistory[0]
			if write.UUID != CommandCharUUID {
				t.Errorf("Expected UUID %s, got %s", CommandCharUUID, write.UUID)
			}
			if !bytes.Equal(write.Data, tc.expected) {
				t.Errorf("Expected command %x, got %x", tc.expected, write.Data)
			}
		})
	}
}

func TestNotificationHandler_ShotClubMetrics_HomeDevice(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)
	sm.SetDeviceType(DeviceTypeHome)

	clubData := []byte{
		0x11, 0x07, 0x0d,
		0x32, 0x00, // Path 0.5
		0x14, 0x00, // Face 0.2
		0x0A, 0x00, // Attack 0.1
		0x28, 0x00, // DynLoft 0.4
	}

	lm.NotificationHandler("", clubData)

	metrics := sm.GetLastClubMetrics()
	if metrics == nil {
		t.Fatal("Expected club metrics to be set")
	}
	if metrics.PathAngle != 0.5 {
		t.Errorf("Expected path angle 0.5, got %v", metrics.PathAngle)
	}
	if metrics.ImpactHorizontal != 0 || metrics.ImpactVertical != 0 || metrics.ClubSpeed != 0 || metrics.SmashFactor != 0 {
		t.Error("Expected Omni-only fields to be zero for Home device")
	}
}

func TestNotificationHandler_ShotClubMetrics_OmniDevice(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)
	sm.SetDeviceType(DeviceTypeOmni)

	clubData := []byte{
		0x11, 0x07,
		0xFF,       // validity bitmask: all 8 valid
		0x32, 0x00, // Path 0.5
		0x14, 0x00, // Face 0.2
		0x0A, 0x00, // Attack 0.1
		0x28, 0x00, // DynLoft 0.4
		0xE8, 0x03, // ImpactH 10.0 (1000/100)
		0xD0, 0x07, // ImpactV 20.0 (2000/100)
		0x10, 0x27, // ClubSpeed 100.0 (10000/100)
		0xC8, 0x00, // SmashFactor 2.0 (200/100)
	}

	lm.NotificationHandler("", clubData)

	metrics := sm.GetLastClubMetrics()
	if metrics == nil {
		t.Fatal("Expected club metrics to be set")
	}
	if metrics.PathAngle != 0.5 {
		t.Errorf("Expected path angle 0.5, got %v", metrics.PathAngle)
	}
	if metrics.ImpactHorizontal != 10.0 {
		t.Errorf("Expected impact horizontal 10.0, got %v", metrics.ImpactHorizontal)
	}
	if metrics.ImpactVertical != 20.0 {
		t.Errorf("Expected impact vertical 20.0, got %v", metrics.ImpactVertical)
	}
	if metrics.ClubSpeed != 100.0 {
		t.Errorf("Expected club speed 100.0, got %v", metrics.ClubSpeed)
	}
	if metrics.SmashFactor != 2.0 {
		t.Errorf("Expected smash factor 2.0, got %v", metrics.SmashFactor)
	}
	if !metrics.IsImpactHorizontalValid || !metrics.IsImpactVerticalValid || !metrics.IsClubSpeedValid || !metrics.IsSmashFactorValid {
		t.Error("Expected all Omni validity flags to be true")
	}
}

func TestNotificationHandler_ShotClubMetrics_OmniPutterInvalidatesClubMetrics(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)
	sm.SetDeviceType(DeviceTypeOmni)
	club := ClubPutter
	sm.SetClub(&club)

	clubData := []byte{
		0x11, 0x07,
		0xFF,
		0x32, 0x00,
		0x14, 0x00,
		0x0A, 0x00,
		0x28, 0x00,
		0xE8, 0x03,
		0xD0, 0x07,
		0x10, 0x27,
		0xC8, 0x00,
	}

	lm.NotificationHandler("", clubData)

	metrics := sm.GetLastClubMetrics()
	if metrics == nil {
		t.Fatal("Expected club metrics to be set")
	}
	if metrics.IsPathAngleValid || metrics.IsFaceAngleValid || metrics.IsAttackAngleValid || metrics.IsDynamicLoftValid {
		t.Fatalf("Expected putter Omni face/path validity to be false, got %+v", metrics)
	}
	if metrics.IsImpactHorizontalValid || metrics.IsImpactVerticalValid || metrics.IsClubSpeedValid || metrics.IsSmashFactorValid {
		t.Fatalf("Expected putter Omni impact/speed validity to be false, got %+v", metrics)
	}
}

func TestNotificationHandler_ShotClubMetrics_OmniShortPacketRetriesThenInvalidates(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	sm.SetDeviceType(DeviceTypeOmni)
	mockClient.connected = true

	shotData := []byte{
		0x11, 0x02, 0x37,
		0x32, 0x00,
		0x14, 0x00,
		0x0A, 0x00,
		0x28, 0x00,
		0x1E, 0x00,
		0x32, 0x00,
		0x1E, 0x00,
	}

	lm.NotificationHandler("", shotData)
	if len(mockClient.GetWriteHistory()) != 1 {
		t.Fatalf("Expected initial club metrics request after Omni shot, got %d writes", len(mockClient.GetWriteHistory()))
	}

	gen := lm.omniClubRetryGen

	clubData := []byte{
		0x11, 0x07, 0x0d,
		0x32, 0x00,
		0x14, 0x00,
		0x0A, 0x00,
		0x28, 0x00,
	}

	lm.NotificationHandler("", clubData)

	if metrics := sm.GetLastClubMetrics(); metrics != nil {
		t.Fatalf("Expected short Omni club packet to be ignored, got %+v", metrics)
	}

	lm.handleOmniClubMetricsTimeout(gen)

	if len(mockClient.GetWriteHistory()) != 2 {
		t.Fatalf("Expected retry club metrics request after first Omni timeout, got %d writes", len(mockClient.GetWriteHistory()))
	}

	if metrics := sm.GetLastClubMetrics(); metrics != nil {
		t.Fatalf("Expected no fallback metrics after first Omni timeout, got %+v", metrics)
	}

	lm.handleOmniClubMetricsTimeout(gen)

	metrics := sm.GetLastClubMetrics()
	if metrics == nil {
		t.Fatal("Expected invalid fallback club metrics after second Omni timeout")
	}
	if metrics.PathAngle != 0 || metrics.FaceAngle != 0 || metrics.AttackAngle != 0 || metrics.DynamicLoftAngle != 0 {
		t.Fatalf("Expected zeroed fallback club metrics, got %+v", metrics)
	}
	if metrics.IsPathAngleValid || metrics.IsFaceAngleValid || metrics.IsAttackAngleValid || metrics.IsDynamicLoftValid {
		t.Fatalf("Expected fallback club validity flags to remain false, got %+v", metrics)
	}
}

func TestSetupNotifications_OmniHandednessChangeSendsCommand(t *testing.T) {
	sm, lm, mockClient, btManager := newTestLaunchMonitor(t)
	mockClient.connected = true
	sm.SetDeviceType(DeviceTypeOmni)

	lm.SetupNotifications(btManager)
	mockClient.ClearWriteHistory()

	handedness := LeftHanded
	sm.SetHandedness(&handedness)

	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected 1 Omni handedness write, got %d", len(writeHistory))
	}

	expected := []byte{0x11, 0x82, 0x00, 0x00, 0x63, 0x01, 0x00, 0x00, 0x00}
	if !bytes.Equal(writeHistory[0].Data, expected) {
		t.Fatalf("Expected Omni handedness command %x, got %x", expected, writeHistory[0].Data)
	}
}

func TestNotificationHandler_Status_OmniRepeatedIdleRearmsDetectMode(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true
	sm.SetDeviceType(DeviceTypeOmni)

	err := lm.ActivateBallDetection()
	if err != nil {
		t.Fatalf("Expected ball detection activation to succeed, got %v", err)
	}

	mockClient.ClearWriteHistory()

	statusData1 := []byte{0x11, 0x03, 0x01, 0x01, 0x07, 0x00, 0x00, 0x02}
	lm.NotificationHandler(NotificationCharUUID, statusData1)
	if len(mockClient.GetWriteHistory()) != 0 {
		t.Fatalf("Expected no Omni re-arm after first idle status, got %d writes", len(mockClient.GetWriteHistory()))
	}

	statusData2 := []byte{0x11, 0x03, 0x02, 0x01, 0x07, 0x00, 0x00, 0x02}
	lm.NotificationHandler(NotificationCharUUID, statusData2)
	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected Omni detect re-arm after second idle status, got %d writes", len(writeHistory))
	}

	data := writeHistory[0].Data
	if len(data) < 5 || data[0] != 0x11 || data[1] != 0x81 || data[3] != 0x01 || data[4] != 0x11 {
		t.Fatalf("Expected Omni detect re-arm command 1181xx011100000000, got %x", data)
	}
}

func TestNotificationHandler_Status_OmniRepeatedIdenticalIdleRearmsDetectMode(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true
	sm.SetDeviceType(DeviceTypeOmni)

	lm.detectStateMu.Lock()
	lm.detectModeActive = true
	lm.omniIdleCount = 0
	lm.detectStateMu.Unlock()

	mockClient.ClearWriteHistory()

	statusData := []byte{0x11, 0x03, 0x01, 0x01, 0x07, 0x00, 0x00, 0x02}
	lm.NotificationHandler(NotificationCharUUID, statusData)
	lm.NotificationHandler(NotificationCharUUID, statusData)

	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected Omni detect re-arm after identical idle status packets, got %d writes", len(writeHistory))
	}

	data := writeHistory[0].Data
	if len(data) < 5 || data[0] != 0x11 || data[1] != 0x81 || data[3] != 0x01 || data[4] != 0x11 {
		t.Fatalf("Expected Omni detect re-arm command 1181xx011100000000, got %x", data)
	}
}

func TestNotificationHandler_ShotBallMetrics_OmniPutterInvalidatesSpinValidity(t *testing.T) {
	sm, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true
	sm.SetDeviceType(DeviceTypeOmni)
	club := ClubPutter
	sm.SetClub(&club)

	shotData := []byte{
		0x11, 0x02, 0x37,
		0x32, 0x00,
		0x14, 0x00,
		0x0A, 0x00,
		0x28, 0x00,
		0x1E, 0x00,
		0x32, 0x00,
		0x1E, 0x00,
	}

	lm.NotificationHandler("", shotData)

	metrics := sm.GetLastBallMetrics()
	if metrics == nil {
		t.Fatal("Expected ball metrics to be set")
	}
	if !metrics.IsBallSpeedValid {
		t.Fatal("Expected ball speed to remain valid for Omni putter shot")
	}
	if metrics.IsTotalSpinValid || metrics.IsSpinAxisValid || metrics.IsBackspinValid || metrics.IsSidespinValid {
		t.Fatalf("Expected Omni putter shot spin validity to be false, got %+v", metrics)
	}
}

func TestSetupNotifications_OmniSettingChangesSendCommands(t *testing.T) {
	sm, lm, mockClient, btManager := newTestLaunchMonitor(t)
	mockClient.connected = true
	sm.SetDeviceType(DeviceTypeOmni)

	lm.SetupNotifications(btManager)

	speedUnit := "mph"
	sm.SetOmniSpeedUnit(&speedUnit)
	writeHistory := mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected 1 Omni units write after speed unit change, got %d", len(writeHistory))
	}
	if !bytes.Equal(writeHistory[0].Data, []byte{0x11, 0x88, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}) {
		t.Fatalf("Expected Omni speed unit command, got %x", writeHistory[0].Data)
	}

	mockClient.ClearWriteHistory()
	distanceUnit := "yards"
	sm.SetOmniDistanceUnit(&distanceUnit)
	writeHistory = mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected 1 Omni units write after distance unit change, got %d", len(writeHistory))
	}
	if !bytes.Equal(writeHistory[0].Data, []byte{0x11, 0x88, 0x01, 0x01, 0x01, 0x02, 0x00, 0x00}) {
		t.Fatalf("Expected Omni distance unit command, got %x", writeHistory[0].Data)
	}

	mockClient.ClearWriteHistory()
	greenSpeed := 12
	sm.SetOmniGreenSpeed(&greenSpeed)
	writeHistory = mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected 1 Omni green speed write, got %d", len(writeHistory))
	}
	if !bytes.Equal(writeHistory[0].Data, []byte{0x11, 0x89, 0x02, 0x04, 0x00, 0x00, 0x00, 0x00}) {
		t.Fatalf("Expected Omni green speed command, got %x", writeHistory[0].Data)
	}

	mockClient.ClearWriteHistory()
	carryAdjustment := -5
	sm.SetOmniCarryAdjustment(&carryAdjustment)
	writeHistory = mockClient.GetWriteHistory()
	if len(writeHistory) != 1 {
		t.Fatalf("Expected 1 Omni carry adjustment write, got %d", len(writeHistory))
	}
	if !bytes.Equal(writeHistory[0].Data, []byte{0x11, 0x8a, 0x03, 0x5f, 0x00, 0x00, 0x00, 0x00}) {
		t.Fatalf("Expected Omni carry adjustment command, got %x", writeHistory[0].Data)
	}
}

func TestBuildOmniInitCommands_UsesConfiguredState(t *testing.T) {
	sm, lm, _, _ := newTestLaunchMonitor(t)
	sm.SetDeviceType(DeviceTypeOmni)

	speedUnit := "mph"
	distanceUnit := "yards"
	greenSpeed := 12
	carryAdjustment := -5
	handedness := LeftHanded
	sm.SetOmniSpeedUnit(&speedUnit)
	sm.SetOmniDistanceUnit(&distanceUnit)
	sm.SetOmniGreenSpeed(&greenSpeed)
	sm.SetOmniCarryAdjustment(&carryAdjustment)
	sm.SetHandedness(&handedness)

	commands := lm.buildOmniInitCommands()
	if len(commands) != 4 {
		t.Fatalf("Expected 4 Omni init commands, got %d", len(commands))
	}

	expectedCommands := []string{
		"1188000101020000",
		"118a015f00000000",
		"1189020400000000",
		"118203006301000000",
	}

	for i, expected := range expectedCommands {
		if commands[i].cmd != expected {
			t.Fatalf("Expected Omni init command %d to be %s, got %s", i, expected, commands[i].cmd)
		}
	}
}

func TestHeartbeatCommand_ExactBytes(t *testing.T) {
	_, lm, mockClient, _ := newTestLaunchMonitor(t)
	mockClient.connected = true

	testCases := []struct {
		name     string
		sequence int
		expected []byte
	}{
		{
			name:     "Sequence 0",
			sequence: 0,
			expected: []byte{0x11, 0x83, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "Sequence 15",
			sequence: 15,
			expected: []byte{0x11, 0x83, 0x0f, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "Sequence 255",
			sequence: 255,
			expected: []byte{0x11, 0x83, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient.ClearWriteHistory()
			command := HeartbeatCommand(tc.sequence)
			err := lm.SendCommand(command)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			writeHistory := mockClient.GetWriteHistory()
			if len(writeHistory) != 1 {
				t.Fatalf("Expected 1 write, got %d", len(writeHistory))
			}

			write := writeHistory[0]
			if write.UUID != CommandCharUUID {
				t.Errorf("Expected UUID %s, got %s", CommandCharUUID, write.UUID)
			}
			if !bytes.Equal(write.Data, tc.expected) {
				t.Errorf("Expected command %x, got %x", tc.expected, write.Data)
			}
		})
	}
}
