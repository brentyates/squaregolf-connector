package gspro

// Models for GSPro integration
// These data structures represent the messages exchanged with GSPro

// Message represents the base message structure from GSPro
type Message struct {
	Message string `json:"Message"`
}

// PlayerInfo represents player information from GSPro
type PlayerInfo struct {
	Message string `json:"Message"`
	Player  Player `json:"Player"`
}

// Player represents player details from GSPro
type Player struct {
	Club   string `json:"Club"`
	Handed string `json:"Handed"`
}

// ShotData represents the shot data sent to GSPro
type ShotData struct {
	DeviceID        string      `json:"DeviceID"`
	Units           string      `json:"Units"`
	APIversion      string      `json:"APIversion"`
	ShotNumber      int         `json:"ShotNumber"`
	ShotDataOptions ShotOptions `json:"ShotDataOptions"`
	BallData        *BallData   `json:"BallData,omitempty"`
	ClubData        *ClubData   `json:"ClubData,omitempty"`
}

// ShotOptions represents shot data options
type ShotOptions struct {
	ContainsBallData          bool `json:"ContainsBallData"`
	ContainsClubData          bool `json:"ContainsClubData"`
	LaunchMonitorIsReady      bool `json:"LaunchMonitorIsReady,omitempty"`
	LaunchMonitorBallDetected bool `json:"LaunchMonitorBallDetected,omitempty"`
}

// BallData represents ball data sent to GSPro
type BallData struct {
	Speed     float64 `json:"Speed"`
	SpinAxis  float64 `json:"SpinAxis"`
	TotalSpin int16   `json:"TotalSpin"`
	BackSpin  int16   `json:"BackSpin"`
	SideSpin  int16   `json:"SideSpin"`
	HLA       float64 `json:"HLA"`
	VLA       float64 `json:"VLA"`
}

// ClubData represents club data sent to GSPro
type ClubData struct {
	Speed                float64 `json:"Speed"`
	AngleOfAttack        float64 `json:"AngleOfAttack"`
	FaceToTarget         float64 `json:"FaceToTarget"`
	Lie                  float64 `json:"Lie"`
	Loft                 float64 `json:"Loft"`
	Path                 float64 `json:"Path"`
	SpeedAtImpact        float64 `json:"SpeedAtImpact"`
	VerticalFaceImpact   float64 `json:"VerticalFaceImpact"`
	HorizontalFaceImpact float64 `json:"HorizontalFaceImpact"`
	ClosureRate          float64 `json:"ClosureRate"`
}