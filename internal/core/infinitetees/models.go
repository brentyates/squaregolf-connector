package infinitetees

type Message struct {
	Message string `json:"Message"`
}

type PlayerInfo struct {
	Message string `json:"Message"`
	Player  Player `json:"Player"`
}

type Player struct {
	Club   string `json:"Club"`
	Handed string `json:"Handed"`
}

type ShotData struct {
	DeviceID        string      `json:"DeviceID"`
	Units           string      `json:"Units"`
	APIversion      string      `json:"APIversion"`
	ShotNumber      int         `json:"ShotNumber"`
	ShotDataOptions ShotOptions `json:"ShotDataOptions"`
	BallData        *BallData   `json:"BallData,omitempty"`
	ClubData        *ClubData   `json:"ClubData,omitempty"`
}

type ShotOptions struct {
	ContainsBallData          bool `json:"ContainsBallData"`
	ContainsClubData          bool `json:"ContainsClubData"`
	LaunchMonitorIsReady      bool `json:"LaunchMonitorIsReady,omitempty"`
	LaunchMonitorBallDetected bool `json:"LaunchMonitorBallDetected,omitempty"`
}

type BallData struct {
	Speed     float64 `json:"Speed"`
	SpinAxis  float64 `json:"SpinAxis"`
	TotalSpin int16   `json:"TotalSpin"`
	BackSpin  int16   `json:"BackSpin"`
	SideSpin  int16   `json:"SideSpin"`
	HLA       float64 `json:"HLA"`
	VLA       float64 `json:"VLA"`
}

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
