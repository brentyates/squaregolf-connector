package gspro

import (
	"log"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

func (g *Integration) registerStateListeners() {
	g.stateManager.RegisterBallReadyCallback(g.onBallReadyChanged)
	g.stateManager.RegisterLastBallMetricsCallback(g.onLastBallMetricsChanged)
	g.stateManager.RegisterLastClubMetricsCallback(g.onLastClubMetricsChanged)
}

func (g *Integration) onBallReadyChanged(oldValue, newValue bool) {
	if oldValue == newValue {
		return
	}

	if !g.Base.Connected || g.Base.Socket == nil {
		return
	}

	emptyShotData := ShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: g.lastShotNumber,
		ShotDataOptions: ShotOptions{
			ContainsBallData:          false,
			ContainsClubData:          false,
			LaunchMonitorIsReady:      newValue,
			LaunchMonitorBallDetected: newValue,
		},
	}

	if err := g.sendData(emptyShotData); err != nil {
		log.Printf("Error sending empty shot data to GSPro: %v", err)
	}
}

func (g *Integration) onLastBallMetricsChanged(oldValue, newValue *core.BallMetrics) {
	if oldValue == newValue {
		return
	}

	if !g.Base.Connected || g.Base.Socket == nil {
		return
	}

	if newValue == nil {
		return
	}

	gsproShotData := g.convertToGSProShotFormat(*newValue, true)
	if err := g.sendData(gsproShotData); err != nil {
		log.Printf("Error sending shot data to GSPro: %v", err)
	}
}

func (g *Integration) onLastClubMetricsChanged(oldValue, newValue *core.ClubMetrics) {
	if oldValue == newValue {
		return
	}

	if !g.Base.Connected || g.Base.Socket == nil {
		return
	}

	if newValue == nil {
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

		gsproShotData := g.convertToGSProShotFormat(core.BallMetrics{}, false)
		gsproShotData.ShotDataOptions.ContainsBallData = false
		gsproShotData.ShotDataOptions.ContainsClubData = true
		gsproShotData.ClubData = zeroedClubData
		if err := g.sendData(gsproShotData); err != nil {
			log.Printf("Error sending zeroed club data to GSPro: %v", err)
		}
		return
	}

	gsproShotData := g.convertToGSProShotFormat(core.BallMetrics{}, false)
	gsproShotData.ShotDataOptions.ContainsBallData = false
	gsproShotData.ShotDataOptions.ContainsClubData = true
	gsproShotData.ClubData = g.convertClubDataToGSPro(*newValue)
	if err := g.sendData(gsproShotData); err != nil {
		log.Printf("Error sending club data to GSPro: %v", err)
	}
}
