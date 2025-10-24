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

// ShotDetectedRequest represents the request body for POST /api/lm/shot-detected
// Includes ball and club metrics to be saved with the video clip
type ShotDetectedRequest struct {
	BallMetrics *ShotBallMetrics `json:"ball_metrics,omitempty"`
	ClubMetrics *ShotClubMetrics `json:"club_metrics,omitempty"`
}

// ShotBallMetrics represents ball metrics for a shot
type ShotBallMetrics struct {
	BallSpeedMPH    float64 `json:"ball_speed_mph,omitempty"`
	BallSpeedMPS    float64 `json:"ball_speed_mps,omitempty"`
	VerticalAngle   float64 `json:"vertical_angle,omitempty"`
	HorizontalAngle float64 `json:"horizontal_angle,omitempty"`
	TotalSpinRPM    int16   `json:"total_spin_rpm,omitempty"`
	SpinAxis        float64 `json:"spin_axis,omitempty"`
	BackspinRPM     int16   `json:"backspin_rpm,omitempty"`
	SidespinRPM     int16   `json:"sidespin_rpm,omitempty"`
}

// ShotClubMetrics represents club metrics for a shot
type ShotClubMetrics struct {
	PathAngle        float64 `json:"path_angle,omitempty"`
	FaceAngle        float64 `json:"face_angle,omitempty"`
	AttackAngle      float64 `json:"attack_angle,omitempty"`
	DynamicLoftAngle float64 `json:"dynamic_loft_angle,omitempty"`
}

// ShotResponse represents the response from POST /api/lm/shot-detected
type ShotResponse struct {
	Success  bool   `json:"success"`            // Whether the shot detection succeeded
	Message  string `json:"message"`            // Human-readable message
	State    string `json:"state"`              // New state after shot detection
	Filename string `json:"filename,omitempty"` // Filename of the saved video clip (if successful)
}

// CancelResponse represents the response from POST /api/lm/cancel
type CancelResponse struct {
	Success bool   `json:"success"` // Whether the cancel command succeeded
	Message string `json:"message"` // Human-readable message
	State   string `json:"state"`   // New state after cancellation
}

// convertBallMetrics converts core.BallMetrics to camera-specific ShotBallMetrics
func convertBallMetrics(metrics *core.BallMetrics) *ShotBallMetrics {
	if metrics == nil {
		return nil
	}
	return &ShotBallMetrics{
		BallSpeedMPH:    metrics.BallSpeedMPS * 2.23694, // Convert m/s to mph
		BallSpeedMPS:    metrics.BallSpeedMPS,
		VerticalAngle:   metrics.VerticalAngle,
		HorizontalAngle: metrics.HorizontalAngle,
		TotalSpinRPM:    metrics.TotalspinRPM,
		SpinAxis:        metrics.SpinAxis,
		BackspinRPM:     metrics.BackspinRPM,
		SidespinRPM:     metrics.SidespinRPM,
	}
}

// convertClubMetrics converts core.ClubMetrics to camera-specific ShotClubMetrics
func convertClubMetrics(metrics *core.ClubMetrics) *ShotClubMetrics {
	if metrics == nil {
		return nil
	}
	return &ShotClubMetrics{
		PathAngle:        metrics.PathAngle,
		FaceAngle:        metrics.FaceAngle,
		AttackAngle:      metrics.AttackAngle,
		DynamicLoftAngle: metrics.DynamicLoftAngle,
	}
}
