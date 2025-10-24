package camera

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
