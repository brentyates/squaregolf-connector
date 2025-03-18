package core

// ConnectionStatus represents the current state of the Bluetooth connection
type ConnectionStatus string

const (
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusError        ConnectionStatus = "error"
)

// GSProConnectionStatus represents the current state of the GSPro connection
type GSProConnectionStatus string

const (
	GSProStatusDisconnected GSProConnectionStatus = "disconnected"
	GSProStatusConnecting   GSProConnectionStatus = "connecting"
	GSProStatusConnected    GSProConnectionStatus = "connected"
	GSProStatusError        GSProConnectionStatus = "error"
)

// MockMode represents the type of mock implementation to use
type MockMode string

const (
	MockModeNone     MockMode = ""         // No mock, use real implementation
	MockModeStub     MockMode = "stub"     // Basic mock implementation
	MockModeSimulate MockMode = "simulate" // Simulated device with realistic behavior
)

// DeviceState represents the current state of the simulated device
type DeviceState string

const (
	DeviceStateIdle          DeviceState = "idle"
	DeviceStateBallDetection DeviceState = "ball_detection"
	DeviceStateReady         DeviceState = "ball_ready"
)

// BallState represents the current state of the ball in the simulator
type BallState int

const (
	BallStateNone BallState = iota
	BallStateDetected
	BallStateReady
)

// HandednessType represents player handedness
type HandednessType int

const (
	RightHanded HandednessType = iota
	LeftHanded
)

// DetectBallMode represents ball detection mode
type DetectBallMode int

const (
	Deactivate DetectBallMode = iota
	Activate
)

// SpinMode represents spin measurement mode
type SpinMode int

const (
	Standard SpinMode = iota
	Advanced
)

// ClubType represents the different types of golf clubs
type ClubType struct {
	RegularCode    string
	SwingStickCode string
}

// Club types as constants
var (
	// Putter
	ClubPutter = ClubType{RegularCode: "0107", SwingStickCode: "0103"}

	// Drivers and woods
	ClubDriver = ClubType{RegularCode: "0204", SwingStickCode: "0202"}
	ClubWood3  = ClubType{RegularCode: "0305", SwingStickCode: "0301"}
	ClubWood5  = ClubType{RegularCode: "0505", SwingStickCode: "0501"}
	ClubWood7  = ClubType{RegularCode: "0705", SwingStickCode: "0701"}

	// Irons
	ClubIron4 = ClubType{RegularCode: "0406", SwingStickCode: "0400"}
	ClubIron5 = ClubType{RegularCode: "0506", SwingStickCode: "0500"}
	ClubIron6 = ClubType{RegularCode: "0606", SwingStickCode: "0600"}
	ClubIron7 = ClubType{RegularCode: "0706", SwingStickCode: "0700"}
	ClubIron8 = ClubType{RegularCode: "0806", SwingStickCode: "0900"}
	ClubIron9 = ClubType{RegularCode: "0906", SwingStickCode: "0900"}

	// Wedges
	ClubPitchingWedge = ClubType{RegularCode: "0a06", SwingStickCode: "0a00"}
	ClubApproachWedge = ClubType{RegularCode: "0b06", SwingStickCode: "0b00"}
	ClubSandWedge     = ClubType{RegularCode: "0c06", SwingStickCode: "0c00"}
)

// ShotType represents the type of shot
type ShotType string

const (
	ShotTypeFull ShotType = "full"
	ShotTypePutt ShotType = "putt"
)

// BLE Characteristic UUIDs
const (
	CommandCharUUID      = "86602101-6b7e-439a-bdd1-489a3213e9bb"
	NotificationCharUUID = "86602102-6b7e-439a-bdd1-489a3213e9bb"
	BatteryLevelCharUUID = "00002a19-0000-1000-8000-00805f9b34fb"
)

const (
	// AppName is the consistent name used for directories and files
	AppName = "SquareGolf Connector"
	// AppDirName is the sanitized version of AppName for use in paths
	AppDirName = "squaregolf-connector"

	// WindowTitle is the title shown in the window title bar
	WindowTitle = AppName + " - Unofficial Launch Monitor Connector"

	// Navigation screen names
	ScreenDashboard = "Dashboard"
	ScreenAlignment = "Alignment"
	ScreenGSPro     = "GSPro"
	ScreenRange     = "Range"
	ScreenSettings  = "Settings"
)
