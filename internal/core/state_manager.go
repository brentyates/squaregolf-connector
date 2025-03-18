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
	Handedness        *HandednessType
	GSProStatus       GSProConnectionStatus
	GSProError        error
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
	sm.state = AppState{
		ConnectionStatus: ConnectionStatusDisconnected,
		BallDetected:     false,
		BallReady:        false,
		GSProStatus:      GSProStatusDisconnected,
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
