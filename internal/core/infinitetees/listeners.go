package infinitetees

import (
	"log"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

func (it *Integration) registerStateListeners() {
	it.stateManager.RegisterBallReadyCallback(it.onBallReadyChanged)
	it.stateManager.RegisterLastBallMetricsCallback(it.onLastBallMetricsChanged)
	it.stateManager.RegisterLastClubMetricsCallback(it.onLastClubMetricsChanged)
}

func (it *Integration) onBallReadyChanged(oldValue, newValue bool) {
	if oldValue == newValue {
		return
	}

	if !it.Base.Connected || it.Base.Socket == nil {
		return
	}

	emptyShotData := ShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: it.lastShotNumber,
		ShotDataOptions: ShotOptions{
			ContainsBallData:          false,
			ContainsClubData:          false,
			LaunchMonitorIsReady:      newValue,
			LaunchMonitorBallDetected: newValue,
		},
	}

	if err := it.sendData(emptyShotData); err != nil {
		log.Printf("[%s] Error sending empty shot data: %v", it.Name(), err)
	}
}

func (it *Integration) onLastBallMetricsChanged(oldValue, newValue *core.BallMetrics) {
	if oldValue == newValue {
		return
	}

	if !it.Base.Connected || it.Base.Socket == nil {
		return
	}

	if newValue == nil {
		return
	}

	shotData := it.convertToShotFormat(*newValue, true)
	if err := it.sendData(shotData); err != nil {
		log.Printf("[%s] Error sending shot data: %v", it.Name(), err)
	}
}

func (it *Integration) onLastClubMetricsChanged(oldValue, newValue *core.ClubMetrics) {
	if oldValue == newValue {
		return
	}

	if !it.Base.Connected || it.Base.Socket == nil {
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

		shotData := it.convertToShotFormat(core.BallMetrics{}, false)
		shotData.ShotDataOptions.ContainsBallData = false
		shotData.ShotDataOptions.ContainsClubData = true
		shotData.ClubData = zeroedClubData
		if err := it.sendData(shotData); err != nil {
			log.Printf("[%s] Error sending zeroed club data: %v", it.Name(), err)
		}
		return
	}

	shotData := it.convertToShotFormat(core.BallMetrics{}, false)
	shotData.ShotDataOptions.ContainsBallData = false
	shotData.ShotDataOptions.ContainsClubData = true
	shotData.ClubData = it.convertClubData(*newValue)
	if err := it.sendData(shotData); err != nil {
		log.Printf("[%s] Error sending club data: %v", it.Name(), err)
	}
}
