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

// AlignmentCommand generates alignment command (command ID 1185)
// confirm: 0 = cancel (exit without saving), 1 = confirm/OK (save calibration)
// targetAngle: target angle in degrees (will be multiplied by 100 and encoded as int32 little-endian)
func AlignmentCommand(sequence int, confirm int, targetAngle float64) string {
	// Convert angle to int32 (angle * 100)
	angleInt := int32(targetAngle * 100)

	// Convert to little-endian bytes
	angleByte0 := byte(angleInt & 0xFF)
	angleByte1 := byte((angleInt >> 8) & 0xFF)
	angleByte2 := byte((angleInt >> 16) & 0xFF)
	angleByte3 := byte((angleInt >> 24) & 0xFF)

	return fmt.Sprintf("1185%02x%02x%02x%02x%02x%02x",
		sequence,
		confirm,
		angleByte0,
		angleByte1,
		angleByte2,
		angleByte3)
}

// StartAlignmentCommand generates command to start alignment mode (confirm=0, angle=0)
func StartAlignmentCommand(sequence int) string {
	return AlignmentCommand(sequence, 0, 0.0)
}

// StopAlignmentCommand generates command to stop alignment and save calibration (confirm=1, OK button)
func StopAlignmentCommand(sequence int, targetAngle float64) string {
	return AlignmentCommand(sequence, 1, targetAngle)
}

// CancelAlignmentCommand generates command to cancel alignment without saving (confirm=0, Cancel button)
func CancelAlignmentCommand(sequence int, targetAngle float64) string {
	return AlignmentCommand(sequence, 0, targetAngle)
}

// RequestClubMetricsCommand generates club metrics request command
func RequestClubMetricsCommand(sequence int) string {
	return fmt.Sprintf("1187%02x000000000000", sequence)
}

// GetOSVersionCommand generates firmware version request command (command ID 1192/0x92)
func GetOSVersionCommand(sequence int) string {
	return fmt.Sprintf("1192%02x0000000000", sequence)
}
