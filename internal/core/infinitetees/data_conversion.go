package infinitetees

import (
	"github.com/brentyates/squaregolf-connector/internal/core"
)

func (it *Integration) convertToShotFormat(ballMetrics core.BallMetrics, incrementShot bool) ShotData {
	if incrementShot {
		it.shotNumber++
		it.lastShotNumber = it.shotNumber
	}

	return ShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: it.lastShotNumber,
		ShotDataOptions: ShotOptions{
			ContainsBallData: true,
			ContainsClubData: false,
		},
		BallData: &BallData{
			Speed:     ballMetrics.BallSpeedMPS * 2.23694,
			SpinAxis:  ballMetrics.SpinAxis * -1,
			TotalSpin: ballMetrics.TotalspinRPM,
			BackSpin:  ballMetrics.BackspinRPM,
			SideSpin:  ballMetrics.SidespinRPM * -1,
			HLA:       ballMetrics.HorizontalAngle,
			VLA:       ballMetrics.VerticalAngle,
		},
		ClubData: &ClubData{},
	}
}

func (it *Integration) convertClubData(clubMetrics core.ClubMetrics) *ClubData {
	return &ClubData{
		Speed:                clubMetrics.ClubSpeed * 2.23694,
		AngleOfAttack:        clubMetrics.AttackAngle,
		FaceToTarget:         clubMetrics.FaceAngle,
		Lie:                  0,
		Loft:                 clubMetrics.DynamicLoftAngle,
		Path:                 clubMetrics.PathAngle,
		SpeedAtImpact:        0,
		VerticalFaceImpact:   clubMetrics.ImpactVertical,
		HorizontalFaceImpact: clubMetrics.ImpactHorizontal,
		ClosureRate:          0,
	}
}

func (it *Integration) mapClubToInternal(clubName string) *core.ClubType {
	clubMap := map[string]core.ClubType{
		"DR": core.ClubDriver,
		"W2": core.ClubWood3,
		"W3": core.ClubWood3,
		"W4": core.ClubWood5,
		"W5": core.ClubWood5,
		"W6": core.ClubWood7,
		"W7": core.ClubWood7,
		"H2": core.ClubWood3,
		"H3": core.ClubWood3,
		"H4": core.ClubWood3,
		"H5": core.ClubWood3,
		"H6": core.ClubWood5,
		"H7": core.ClubIron4,
		"I1": core.ClubWood3,
		"I2": core.ClubWood3,
		"I3": core.ClubWood5,
		"I4": core.ClubIron4,
		"I5": core.ClubIron5,
		"I6": core.ClubIron6,
		"I7": core.ClubIron7,
		"I8": core.ClubIron8,
		"I9": core.ClubIron9,
		"PW": core.ClubPitchingWedge,
		"AW": core.ClubApproachWedge,
		"GW": core.ClubApproachWedge,
		"SW": core.ClubSandWedge,
		"LW": core.ClubSandWedge,
		"PT": core.ClubPutter,
	}

	if club, ok := clubMap[clubName]; ok {
		return &club
	}
	return nil
}

func mapClubToFriendlyName(clubCode string) string {
	nameMap := map[string]string{
		"DR": "DR",
		"W2": "2W",
		"W3": "3W",
		"W4": "4W",
		"W5": "5W",
		"W6": "6W",
		"W7": "7W",
		"H2": "2H",
		"H3": "3H",
		"H4": "4H",
		"H5": "5H",
		"H6": "6H",
		"H7": "7H",
		"I1": "1I",
		"I2": "2I",
		"I3": "3I",
		"I4": "4I",
		"I5": "5I",
		"I6": "6I",
		"I7": "7I",
		"I8": "8I",
		"I9": "9I",
		"PW": "PW",
		"AW": "AW",
		"GW": "GW",
		"SW": "SW",
		"LW": "LW",
		"PT": "PUTT",
	}

	if name, ok := nameMap[clubCode]; ok {
		return name
	}
	return clubCode
}
