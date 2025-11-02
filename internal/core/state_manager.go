package core

import (
	"sync"
)

// BallPosition represents the 3D position of the ball
type BallPosition struct {
	X int32
	Y int32
	Z int32
}

// AppState represents the complete application state
type AppState struct {
	DeviceDisplayName *string
	ConnectionStatus  ConnectionStatus
	BatteryLevel      *int
	BallDetected      bool
	BallReady         bool
	BallPosition      *BallPosition
	LastBallMetrics   *BallMetrics
	LastClubMetrics   *ClubMetrics
	LastError         error
	Club              *ClubType
	ClubName          *string // Human-readable club name from GSPro (e.g., "Driver", "7-iron")
	Handedness        *HandednessType
	GSProStatus       GSProConnectionStatus
	GSProError        error
	SpinMode          *SpinMode
	CameraURL         *string
	CameraEnabled     bool
	IsAligning        bool    // Whether alignment mode UI is active
	AlignmentAngle    float64 // Current aim angle in degrees (left negative, right positive)
	IsAligned         bool    // Whether device is currently aligned (within tolerance)
	FirmwareVersion   *string // Device firmware version (e.g., "1.6.18")
	LauncherVersion   *string // Launcher version
	MMIVersion        *string // MMI version
}

// StateCallback is a generic type for state change callbacks
type StateCallback[T any] func(oldValue, newValue T)

// StateManager manages the application state with type safety
type StateManager struct {
	state     AppState
	callbacks struct {
		DeviceDisplayName []StateCallback[*string]
		ConnectionStatus  []StateCallback[ConnectionStatus]
		BatteryLevel      []StateCallback[*int]
		BallDetected      []StateCallback[bool]
		BallReady         []StateCallback[bool]
		BallPosition      []StateCallback[*BallPosition]
		LastBallMetrics   []StateCallback[*BallMetrics]
		LastClubMetrics   []StateCallback[*ClubMetrics]
		LastError         []StateCallback[error]
		Club              []StateCallback[*ClubType]
		Handedness        []StateCallback[*HandednessType]
		GSProStatus       []StateCallback[GSProConnectionStatus]
		GSProError        []StateCallback[error]
		SpinMode          []StateCallback[*SpinMode]
		CameraURL         []StateCallback[*string]
		CameraEnabled     []StateCallback[bool]
		IsAligning        []StateCallback[bool]
		AlignmentAngle    []StateCallback[float64]
		IsAligned         []StateCallback[bool]
		FirmwareVersion   []StateCallback[*string]
		LauncherVersion   []StateCallback[*string]
		MMIVersion        []StateCallback[*string]
	}
	mu sync.RWMutex
}

var (
	instance *StateManager
	once     sync.Once
)

// GetInstance returns the singleton instance of StateManager
func GetInstance() *StateManager {
	once.Do(func() {
		instance = &StateManager{}
		instance.initialize()
	})
	return instance
}

// initialize sets up the default state values
func (sm *StateManager) initialize() {
	defaultCameraURL := "http://localhost:5000"
	sm.state = AppState{
		ConnectionStatus: ConnectionStatusDisconnected,
		BallDetected:     false,
		BallReady:        false,
		GSProStatus:      GSProStatusDisconnected,
		CameraURL:        &defaultCameraURL,
		CameraEnabled:    false,
		IsAligning:       false,
		AlignmentAngle:   0.0,
		IsAligned:        false,
	}
}

// GetDeviceDisplayName returns the device display name
func (sm *StateManager) GetDeviceDisplayName() *string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.DeviceDisplayName
}

// SetDeviceDisplayName sets the device display name
func (sm *StateManager) SetDeviceDisplayName(value *string) {
	sm.mu.Lock()
	oldValue := sm.state.DeviceDisplayName
	sm.state.DeviceDisplayName = value
	callbacks := sm.callbacks.DeviceDisplayName
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetConnectionStatus returns the connection status
func (sm *StateManager) GetConnectionStatus() ConnectionStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.ConnectionStatus
}

// SetConnectionStatus sets the connection status
func (sm *StateManager) SetConnectionStatus(value ConnectionStatus) {
	sm.mu.Lock()
	oldValue := sm.state.ConnectionStatus
	sm.state.ConnectionStatus = value
	callbacks := sm.callbacks.ConnectionStatus
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetBatteryLevel returns the battery level
func (sm *StateManager) GetBatteryLevel() *int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.BatteryLevel
}

// SetBatteryLevel sets the battery level
func (sm *StateManager) SetBatteryLevel(value *int) {
	sm.mu.Lock()
	oldValue := sm.state.BatteryLevel
	sm.state.BatteryLevel = value
	callbacks := sm.callbacks.BatteryLevel
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetBallDetected returns whether a ball is detected
func (sm *StateManager) GetBallDetected() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.BallDetected
}

// SetBallDetected sets whether a ball is detected
func (sm *StateManager) SetBallDetected(value bool) {
	sm.mu.Lock()
	oldValue := sm.state.BallDetected
	sm.state.BallDetected = value
	callbacks := sm.callbacks.BallDetected
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetBallReady returns whether a ball is ready
func (sm *StateManager) GetBallReady() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.BallReady
}

// SetBallReady sets whether a ball is ready
func (sm *StateManager) SetBallReady(value bool) {
	sm.mu.Lock()
	oldValue := sm.state.BallReady
	sm.state.BallReady = value
	callbacks := sm.callbacks.BallReady
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetBallPosition returns the ball position
func (sm *StateManager) GetBallPosition() *BallPosition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.BallPosition
}

// SetBallPosition sets the ball position
func (sm *StateManager) SetBallPosition(value *BallPosition) {
	sm.mu.Lock()
	oldValue := sm.state.BallPosition
	sm.state.BallPosition = value
	callbacks := sm.callbacks.BallPosition
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetLastBallMetrics returns the last ball metrics
func (sm *StateManager) GetLastBallMetrics() *BallMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.LastBallMetrics
}

// SetLastBallMetrics sets the last ball metrics
func (sm *StateManager) SetLastBallMetrics(value *BallMetrics) {
	sm.mu.Lock()
	oldValue := sm.state.LastBallMetrics
	sm.state.LastBallMetrics = value
	callbacks := sm.callbacks.LastBallMetrics
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetLastClubMetrics returns the last club metrics
func (sm *StateManager) GetLastClubMetrics() *ClubMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.LastClubMetrics
}

// SetLastClubMetrics sets the last club metrics
func (sm *StateManager) SetLastClubMetrics(value *ClubMetrics) {
	sm.mu.Lock()
	oldValue := sm.state.LastClubMetrics
	sm.state.LastClubMetrics = value
	callbacks := sm.callbacks.LastClubMetrics
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetLastError returns the last error
func (sm *StateManager) GetLastError() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.LastError
}

// SetLastError sets the last error
func (sm *StateManager) SetLastError(value error) {
	sm.mu.Lock()
	oldValue := sm.state.LastError
	sm.state.LastError = value
	callbacks := sm.callbacks.LastError
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetClub returns the current club
func (sm *StateManager) GetClub() *ClubType {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.Club
}

// SetClub sets the current club
func (sm *StateManager) SetClub(value *ClubType) {
	sm.mu.Lock()
	oldValue := sm.state.Club
	sm.state.Club = value
	callbacks := sm.callbacks.Club
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetClubName returns the human-readable club name
func (sm *StateManager) GetClubName() *string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.ClubName
}

// SetClubName sets the human-readable club name
func (sm *StateManager) SetClubName(value *string) {
	sm.mu.Lock()
	sm.state.ClubName = value
	sm.mu.Unlock()
}

// GetHandedness returns the current handedness
func (sm *StateManager) GetHandedness() *HandednessType {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.Handedness
}

// SetHandedness sets the current handedness
func (sm *StateManager) SetHandedness(value *HandednessType) {
	sm.mu.Lock()
	oldValue := sm.state.Handedness
	sm.state.Handedness = value
	callbacks := sm.callbacks.Handedness
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterDeviceDisplayNameCallback registers a callback for device display name changes
func (sm *StateManager) RegisterDeviceDisplayNameCallback(callback StateCallback[*string]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.DeviceDisplayName = append(sm.callbacks.DeviceDisplayName, callback)
}

// RegisterConnectionStatusCallback registers a callback for connection status changes
func (sm *StateManager) RegisterConnectionStatusCallback(callback StateCallback[ConnectionStatus]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.ConnectionStatus = append(sm.callbacks.ConnectionStatus, callback)
}

// RegisterBatteryLevelCallback registers a callback for battery level changes
func (sm *StateManager) RegisterBatteryLevelCallback(callback StateCallback[*int]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.BatteryLevel = append(sm.callbacks.BatteryLevel, callback)
}

// RegisterBallDetectedCallback registers a callback for ball detected changes
func (sm *StateManager) RegisterBallDetectedCallback(callback StateCallback[bool]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.BallDetected = append(sm.callbacks.BallDetected, callback)
}

// RegisterBallReadyCallback registers a callback for ball ready changes
func (sm *StateManager) RegisterBallReadyCallback(callback StateCallback[bool]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.BallReady = append(sm.callbacks.BallReady, callback)
}

// RegisterBallPositionCallback registers a callback for ball position changes
func (sm *StateManager) RegisterBallPositionCallback(callback StateCallback[*BallPosition]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.BallPosition = append(sm.callbacks.BallPosition, callback)
}

// RegisterLastBallMetricsCallback registers a callback for last ball metrics changes
func (sm *StateManager) RegisterLastBallMetricsCallback(callback StateCallback[*BallMetrics]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.LastBallMetrics = append(sm.callbacks.LastBallMetrics, callback)
}

// RegisterLastClubMetricsCallback registers a callback for last club metrics changes
func (sm *StateManager) RegisterLastClubMetricsCallback(callback StateCallback[*ClubMetrics]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.LastClubMetrics = append(sm.callbacks.LastClubMetrics, callback)
}

// RegisterLastErrorCallback registers a callback for last error changes
func (sm *StateManager) RegisterLastErrorCallback(callback StateCallback[error]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.LastError = append(sm.callbacks.LastError, callback)
}

// RegisterClubCallback registers a callback for club changes
func (sm *StateManager) RegisterClubCallback(callback StateCallback[*ClubType]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.Club = append(sm.callbacks.Club, callback)
}

// RegisterHandednessCallback registers a callback for handedness changes
func (sm *StateManager) RegisterHandednessCallback(callback StateCallback[*HandednessType]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.Handedness = append(sm.callbacks.Handedness, callback)
}

// GetGSProStatus returns the GSPro connection status
func (sm *StateManager) GetGSProStatus() GSProConnectionStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.GSProStatus
}

// SetGSProStatus sets the GSPro connection status
func (sm *StateManager) SetGSProStatus(value GSProConnectionStatus) {
	sm.mu.Lock()
	oldValue := sm.state.GSProStatus
	sm.state.GSProStatus = value
	callbacks := sm.callbacks.GSProStatus
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetGSProError returns the GSPro error
func (sm *StateManager) GetGSProError() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.GSProError
}

// SetGSProError sets the GSPro error
func (sm *StateManager) SetGSProError(value error) {
	sm.mu.Lock()
	oldValue := sm.state.GSProError
	sm.state.GSProError = value
	callbacks := sm.callbacks.GSProError
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterGSProStatusCallback registers a callback for GSPro status changes
func (sm *StateManager) RegisterGSProStatusCallback(callback StateCallback[GSProConnectionStatus]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.GSProStatus = append(sm.callbacks.GSProStatus, callback)
}

// RegisterGSProErrorCallback registers a callback for GSPro error changes
func (sm *StateManager) RegisterGSProErrorCallback(callback StateCallback[error]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.GSProError = append(sm.callbacks.GSProError, callback)
}

// GetSpinMode returns the current spin mode
func (sm *StateManager) GetSpinMode() *SpinMode {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.SpinMode
}

// SetSpinMode sets the current spin mode
func (sm *StateManager) SetSpinMode(value *SpinMode) {
	sm.mu.Lock()
	oldValue := sm.state.SpinMode
	sm.state.SpinMode = value
	callbacks := sm.callbacks.SpinMode
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterSpinModeCallback registers a callback for spin mode changes
func (sm *StateManager) RegisterSpinModeCallback(callback StateCallback[*SpinMode]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.SpinMode = append(sm.callbacks.SpinMode, callback)
}

// GetCameraURL returns the camera URL
func (sm *StateManager) GetCameraURL() *string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.CameraURL
}

// SetCameraURL sets the camera URL
func (sm *StateManager) SetCameraURL(value *string) {
	sm.mu.Lock()
	oldValue := sm.state.CameraURL
	sm.state.CameraURL = value
	callbacks := sm.callbacks.CameraURL
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetCameraEnabled returns whether camera integration is enabled
func (sm *StateManager) GetCameraEnabled() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.CameraEnabled
}

// SetCameraEnabled sets whether camera integration is enabled
func (sm *StateManager) SetCameraEnabled(value bool) {
	sm.mu.Lock()
	oldValue := sm.state.CameraEnabled
	sm.state.CameraEnabled = value
	callbacks := sm.callbacks.CameraEnabled
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterCameraURLCallback registers a callback for camera URL changes
func (sm *StateManager) RegisterCameraURLCallback(callback StateCallback[*string]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.CameraURL = append(sm.callbacks.CameraURL, callback)
}

// RegisterCameraEnabledCallback registers a callback for camera enabled changes
func (sm *StateManager) RegisterCameraEnabledCallback(callback StateCallback[bool]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.CameraEnabled = append(sm.callbacks.CameraEnabled, callback)
}

// GetIsAligning returns whether alignment mode is active
func (sm *StateManager) GetIsAligning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.IsAligning
}

// SetIsAligning sets whether alignment mode is active
func (sm *StateManager) SetIsAligning(value bool) {
	sm.mu.Lock()
	oldValue := sm.state.IsAligning
	sm.state.IsAligning = value
	callbacks := sm.callbacks.IsAligning
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetAlignmentAngle returns the current alignment angle in degrees
func (sm *StateManager) GetAlignmentAngle() float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.AlignmentAngle
}

// SetAlignmentAngle sets the current alignment angle in degrees
func (sm *StateManager) SetAlignmentAngle(value float64) {
	sm.mu.Lock()
	oldValue := sm.state.AlignmentAngle
	sm.state.AlignmentAngle = value
	callbacks := sm.callbacks.AlignmentAngle
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// GetIsAligned returns whether the device is currently aligned
func (sm *StateManager) GetIsAligned() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.IsAligned
}

// SetIsAligned sets whether the device is currently aligned
func (sm *StateManager) SetIsAligned(value bool) {
	sm.mu.Lock()
	oldValue := sm.state.IsAligned
	sm.state.IsAligned = value
	callbacks := sm.callbacks.IsAligned
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterIsAligningCallback registers a callback for alignment mode changes
func (sm *StateManager) RegisterIsAligningCallback(callback StateCallback[bool]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.IsAligning = append(sm.callbacks.IsAligning, callback)
}

// RegisterAlignmentAngleCallback registers a callback for alignment angle changes
func (sm *StateManager) RegisterAlignmentAngleCallback(callback StateCallback[float64]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.AlignmentAngle = append(sm.callbacks.AlignmentAngle, callback)
}

// RegisterIsAlignedCallback registers a callback for alignment status changes
func (sm *StateManager) RegisterIsAlignedCallback(callback StateCallback[bool]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.IsAligned = append(sm.callbacks.IsAligned, callback)
}

// GetFirmwareVersion returns the device firmware version
func (sm *StateManager) GetFirmwareVersion() *string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.FirmwareVersion
}

// SetFirmwareVersion sets the device firmware version
func (sm *StateManager) SetFirmwareVersion(value *string) {
	sm.mu.Lock()
	oldValue := sm.state.FirmwareVersion
	sm.state.FirmwareVersion = value
	callbacks := sm.callbacks.FirmwareVersion
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterFirmwareVersionCallback registers a callback for firmware version changes
func (sm *StateManager) RegisterFirmwareVersionCallback(callback StateCallback[*string]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.FirmwareVersion = append(sm.callbacks.FirmwareVersion, callback)
}

// GetLauncherVersion returns the launcher version
func (sm *StateManager) GetLauncherVersion() *string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.LauncherVersion
}

// SetLauncherVersion sets the launcher version
func (sm *StateManager) SetLauncherVersion(value *string) {
	sm.mu.Lock()
	oldValue := sm.state.LauncherVersion
	sm.state.LauncherVersion = value
	callbacks := sm.callbacks.LauncherVersion
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterLauncherVersionCallback registers a callback for launcher version changes
func (sm *StateManager) RegisterLauncherVersionCallback(callback StateCallback[*string]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.LauncherVersion = append(sm.callbacks.LauncherVersion, callback)
}

// GetMMIVersion returns the MMI version
func (sm *StateManager) GetMMIVersion() *string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state.MMIVersion
}

// SetMMIVersion sets the MMI version
func (sm *StateManager) SetMMIVersion(value *string) {
	sm.mu.Lock()
	oldValue := sm.state.MMIVersion
	sm.state.MMIVersion = value
	callbacks := sm.callbacks.MMIVersion
	sm.mu.Unlock()

	for _, callback := range callbacks {
		callback(oldValue, value)
	}
}

// RegisterMMIVersionCallback registers a callback for MMI version changes
func (sm *StateManager) RegisterMMIVersionCallback(callback StateCallback[*string]) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.callbacks.MMIVersion = append(sm.callbacks.MMIVersion, callback)
}
