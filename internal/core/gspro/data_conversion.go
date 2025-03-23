package gspro

import (
	"github.com/brentyates/squaregolf-connector/internal/core"
)

// convertToGSProShotFormat converts internal shot data format to GSPro format
func (g *Integration) convertToGSProShotFormat(ballMetrics core.BallMetrics) ShotData {
	// Increment shot number for each shot
	g.shotNumber++

	return ShotData{
		DeviceID:   "CustomLaunchMonitor",
		Units:      "Yards",
		APIversion: "1",
		ShotNumber: g.shotNumber,
		ShotDataOptions: ShotOptions{
			ContainsBallData: true,
			ContainsClubData: false,
		},
		BallData: &BallData{
			Speed:     ballMetrics.BallSpeedMPS * 2.23694, // Convert m/s to mph
			SpinAxis:  ballMetrics.SpinAxis * -1,
			TotalSpin: ballMetrics.TotalspinRPM,
			BackSpin:  ballMetrics.BackspinRPM,
			SideSpin:  ballMetrics.SidespinRPM * -1,
			HLA:       ballMetrics.HorizontalAngle,
			VLA:       ballMetrics.VerticalAngle,
		},
		ClubData: &ClubData{}, // Empty club data
	}
}

// convertClubDataToGSPro converts internal club data format to GSPro format
func (g *Integration) convertClubDataToGSPro(clubMetrics core.ClubMetrics) *ClubData {
	return &ClubData{
		Speed:                0, // Not provided by our sensor
		AngleOfAttack:        clubMetrics.AttackAngle,
		FaceToTarget:         clubMetrics.FaceAngle,
		Lie:                  0, // Not provided by our sensor
		Loft:                 clubMetrics.DynamicLoftAngle,
		Path:                 clubMetrics.PathAngle,
		SpeedAtImpact:        0, // Not provided by our sensor
		VerticalFaceImpact:   0, // Not provided by our sensor
		HorizontalFaceImpact: 0, // Not provided by our sensor
		ClosureRate:          0, // Not provided by our sensor
	}
}

// mapGSProClubToInternal maps GSPro club name to internal ClubType
func (g *Integration) mapGSProClubToInternal(clubName string) *core.ClubType {
	// Map GSPro club names to our internal ClubType
	clubMap := map[string]core.ClubType{
		// Drivers and woods
		"DR": core.ClubDriver,
		"W2": core.ClubWood3,
		"W3": core.ClubWood3,
		"W4": core.ClubWood5,
		"W5": core.ClubWood5,
		"W6": core.ClubWood7,
		"W7": core.ClubWood7,

		// Hybrids
		"H2": core.ClubWood3,
		"H3": core.ClubWood3,
		"H4": core.ClubWood3,
		"H5": core.ClubWood3,
		"H6": core.ClubWood5,
		"H7": core.ClubIron4,

		// Irons
		"I1": core.ClubWood3,
		"I2": core.ClubWood3,
		"I3": core.ClubWood5,
		"I4": core.ClubIron4,
		"I5": core.ClubIron5,
		"I6": core.ClubIron6,
		"I7": core.ClubIron7,
		"I8": core.ClubIron8,
		"I9": core.ClubIron9,

		// Wedges
		"PW": core.ClubPitchingWedge,
		"AW": core.ClubApproachWedge,
		"GW": core.ClubApproachWedge,
		"SW": core.ClubSandWedge,

		// Putter
		"PT": core.ClubPutter,
	}

	if club, ok := clubMap[clubName]; ok {
		return &club
	}
	return nil
}