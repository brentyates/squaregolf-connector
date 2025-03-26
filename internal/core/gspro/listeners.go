package gspro

import (
	"log"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

// Register state listeners
func (g *Integration) registerStateListeners() {
	g.stateManager.RegisterBallReadyCallback(g.onBallReadyChanged)
	g.stateManager.RegisterLastBallMetricsCallback(g.onLastBallMetricsChanged)
	g.stateManager.RegisterLastClubMetricsCallback(g.onLastClubMetricsChanged)
}

// onBallReadyChanged handles ball ready state changed event from state manager
func (g *Integration) onBallReadyChanged(oldValue, newValue bool) {
	if oldValue == newValue {
		return
	}

	if !g.connected || g.socket == nil {
		return
	}

	// Send empty shot data - don't increment shot number for ball ready changes
	emptyShotData := ShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: g.lastShotNumber, // Use the last shot number instead of incrementing
		ShotDataOptions: ShotOptions{
			ContainsBallData:          false,
			ContainsClubData:          false,
			LaunchMonitorIsReady:      newValue,
			LaunchMonitorBallDetected: newValue,
		},
	}

	try := func() error {
		return g.sendData(emptyShotData)
	}

	if err := try(); err != nil {
		log.Printf("Error sending empty shot data to GSPro: %v", err)
	}
}

// onLastBallMetricsChanged handles last ball metrics changed event from state manager
func (g *Integration) onLastBallMetricsChanged(oldValue, newValue *core.BallMetrics) {
	if oldValue == newValue {
		return
	}

	if !g.connected || g.socket == nil {
		return
	}

	if newValue == nil {
		return
	}

	try := func() error {
		gsproShotData := g.convertToGSProShotFormat(*newValue, true)
		return g.sendData(gsproShotData)
	}

	if err := try(); err != nil {
		log.Printf("Error sending shot data to GSPro: %v", err)
	}
}

// onLastClubMetricsChanged handles last club metrics changed event from state manager
func (g *Integration) onLastClubMetricsChanged(oldValue, newValue *core.ClubMetrics) {
	if oldValue == newValue {
		return
	}

	if !g.connected || g.socket == nil {
		return
	}

	// Handle the case where club metrics are nil, which means no club data was available
	if newValue == nil {
		// Send zeroed club data to GSPro
		zeroedClubData := &ClubData{
			Speed:                0,
			AngleOfAttack:        0,
			FaceToTarget:         0,
			Lie:                  0,
			Loft:                 0,
			Path:                 0,
			SpeedAtImpact:        0,
			VerticalFaceImpact:   0,
			HorizontalFaceImpact: 0,
			ClosureRate:          0,
		}

		try := func() error {
			gsproShotData := g.convertToGSProShotFormat(core.BallMetrics{}, false) // Don't increment shot number
			gsproShotData.ShotDataOptions.ContainsBallData = false
			gsproShotData.ShotDataOptions.ContainsClubData = true
			gsproShotData.ClubData = zeroedClubData
			return g.sendData(gsproShotData)
		}

		if err := try(); err != nil {
			log.Printf("Error sending zeroed club data to GSPro: %v", err)
		}
		return
	}

	// Update the last shot data with club metrics
	try := func() error {
		gsproShotData := g.convertToGSProShotFormat(core.BallMetrics{}, false) // Don't increment shot number for club data
		gsproShotData.ShotDataOptions.ContainsBallData = false
		gsproShotData.ShotDataOptions.ContainsClubData = true
		gsproShotData.ClubData = g.convertClubDataToGSPro(*newValue)
		return g.sendData(gsproShotData)
	}

	if err := try(); err != nil {
		log.Printf("Error sending club data to GSPro: %v", err)
	}
}

// AddShotListener adds a listener for shot events
func (g *Integration) AddShotListener(listener func(ShotData)) {
	g.shotListeners = append(g.shotListeners, listener)
}
