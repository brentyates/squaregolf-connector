package camera

import "github.com/brentyates/squaregolf-connector/internal/core"

// CameraStatus represents the current status of the camera system
// Maps to the response from GET /api/lm/status
type CameraStatus struct {
	State                string  `json:"state"`                  // Current LM state: "idle", "armed", or "processing"
	RecordingDuration    float64 `json:"recording_duration"`     // How long the current recording has been running (seconds)
	MaxRecordingDuration int     `json:"max_recording_duration"` // Maximum recording duration before timeout (seconds)
}

// ArmResponse represents the response from POST /api/lm/arm
type ArmResponse struct {
	Success bool   `json:"success"` // Whether the arm command succeeded
	Message string `json:"message"` // Human-readable message
	State   string `json:"state"`   // New state after arming
}

// BallData represents ball metrics sent to SwingCam (flat structure, camelCase)
// Sent directly as request body to POST /api/lm/shot-detected
type BallData struct {
	BallSpeed       float64 `json:"ballSpeed,omitempty"`       // Ball speed in mph
	LaunchAngle     float64 `json:"launchAngle,omitempty"`     // Vertical launch angle in degrees
	LaunchDirection float64 `json:"launchDirection,omitempty"` // Horizontal direction in degrees
	SpinRate        int     `json:"spinRate,omitempty"`        // Total spin rate in rpm
	SpinAxis        float64 `json:"spinAxis,omitempty"`        // Spin axis tilt in degrees
	BackSpin        int     `json:"backSpin,omitempty"`        // Back spin component in rpm
	SideSpin        int     `json:"sideSpin,omitempty"`        // Side spin component in rpm
	CarryDistance   float64 `json:"carryDistance,omitempty"`   // Carry distance (yards or meters)
	TotalDistance   float64 `json:"totalDistance,omitempty"`   // Total distance with roll (yards or meters)
	MaxHeight       float64 `json:"maxHeight,omitempty"`       // Apex height (yards or meters)
	LandingAngle    float64 `json:"landingAngle,omitempty"`    // Descent angle at landing in degrees
	HangTime        float64 `json:"hangTime,omitempty"`        // Time in air in seconds
}

// ClubData represents club metrics sent to SwingCam (flat structure, camelCase)
// Sent directly as request body to PATCH /api/recordings/{filename}/metadata
type ClubData struct {
	ClubSpeed    float64 `json:"clubSpeed,omitempty"`    // Club head speed in mph
	ClubPath     float64 `json:"clubPath,omitempty"`     // Club path in degrees (+ = in-to-out, - = out-to-in)
	FaceAngle    float64 `json:"faceAngle,omitempty"`    // Face angle at impact in degrees (+ = open, - = closed)
	FaceToPath   float64 `json:"faceToPath,omitempty"`   // Face to path relationship in degrees
	AttackAngle  float64 `json:"attackAngle,omitempty"`  // Attack angle in degrees (+ = up, - = down)
	DynamicLoft  float64 `json:"dynamicLoft,omitempty"`  // Dynamic loft at impact in degrees
	SmashFactor  float64 `json:"smashFactor,omitempty"`  // Smash factor (ball speed / club speed)
	LowPoint     float64 `json:"lowPoint,omitempty"`     // Low point position (inches before/after ball)
	ClubType     string  `json:"clubType,omitempty"`     // Club name (e.g., "Driver", "7-iron")
}

// ShotResponse represents the response from POST /api/lm/shot-detected
type ShotResponse struct {
	Status   string `json:"status"`             // Status: "success" or "error"
	Filename string `json:"filename,omitempty"` // Filename of the saved video clip (if successful)
	Message  string `json:"message,omitempty"`  // Human-readable message
}

// CancelResponse represents the response from POST /api/lm/cancel
type CancelResponse struct {
	Success bool   `json:"success"` // Whether the cancel command succeeded
	Message string `json:"message"` // Human-readable message
	State   string `json:"state"`   // New state after cancellation
}

// convertBallMetrics converts core.BallMetrics to SwingCam BallData format
func convertBallMetrics(metrics *core.BallMetrics) *BallData {
	if metrics == nil {
		return nil
	}
	return &BallData{
		BallSpeed:       metrics.BallSpeedMPS * 2.23694, // Convert m/s to mph
		LaunchAngle:     metrics.VerticalAngle,
		LaunchDirection: metrics.HorizontalAngle,
		SpinRate:        int(metrics.TotalspinRPM),
		SpinAxis:        metrics.SpinAxis,
		BackSpin:        int(metrics.BackspinRPM),
		SideSpin:        int(metrics.SidespinRPM),
		// Note: CarryDistance, TotalDistance, MaxHeight, LandingAngle, HangTime not available from SquareGolf
	}
}

// convertClubMetrics converts core.ClubMetrics to SwingCam ClubData format
func convertClubMetrics(metrics *core.ClubMetrics) *ClubData {
	if metrics == nil {
		return nil
	}
	return &ClubData{
		ClubPath:    metrics.PathAngle,
		FaceAngle:   metrics.FaceAngle,
		AttackAngle: metrics.AttackAngle,
		DynamicLoft: metrics.DynamicLoftAngle,
		// Note: ClubSpeed, FaceToPath, SmashFactor, LowPoint, ClubType not available from SquareGolf
	}
}
