package core

import (
	"fmt"
)

// HeartbeatCommand generates heartbeat command bytes
func HeartbeatCommand(sequence int) string {
	return fmt.Sprintf("1183%02x0000000000", sequence)
}

// DetectBallCommand generates spin mode configuration command
func DetectBallCommand(sequence int, mode DetectBallMode, spinMode SpinMode) string {
	return fmt.Sprintf("1181%02x0%d1%d00000000", sequence, mode, spinMode)
}

// ClubCommand generates club selection command
func ClubCommand(sequence int, club ClubType, handedness HandednessType) string {
	return fmt.Sprintf("1182%02x%s0%d000000", sequence, club.RegularCode, handedness)
}

// SwingStickCommand generates swing stick mode command
func SwingStickCommand(sequence int, club ClubType, handedness HandednessType) string {
	return fmt.Sprintf("1182%02x%s0%d0000", sequence, club.SwingStickCode, handedness)
}

// AlignmentStickCommand generates alignment stick command
func AlignmentStickCommand(sequence int, handedness HandednessType) string {
	return fmt.Sprintf("1182%02x08080%d000000", sequence, handedness)
}

// RequestClubMetricsCommand generates club metrics request command
func RequestClubMetricsCommand(sequence int) string {
	return fmt.Sprintf("1187%02x000000000000", sequence)
}
