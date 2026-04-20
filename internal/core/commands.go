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

// OmniClubCommand generates club selection command for Omni devices.
// The Omni expects clubSel - 4 compared to the Home encoding.
func OmniClubCommand(sequence int, club ClubType, handedness HandednessType) string {
	var clubNumber, clubSel int
	fmt.Sscanf(club.RegularCode, "%02x%02x", &clubNumber, &clubSel)
	omniSel := clubSel - 4
	if omniSel < 0 {
		omniSel = 0
	}
	return fmt.Sprintf("1182%02x%02x%02x%02x000000", sequence, clubNumber, omniSel, handedness)
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

// GetChargeCommand queries capacitor charge status (command 0x86)
func GetChargeCommand(sequence int) string {
	return fmt.Sprintf("1186%02x0000000000", sequence)
}

// OmniSetUnitsCommand configures the Omni's speed and distance units (command 0x88).
// speedUnit: 0 = m/s, 1 = mph
// distanceUnit: 0 = meters, 1 = yards/feet, 2 = yards/yards
func OmniSetUnitsCommand(sequence int, speedUnit int, distanceUnit int) string {
	distMarker := 0
	distSub := 0
	if distanceUnit > 0 {
		distMarker = 1
		distSub = distanceUnit
	}
	return fmt.Sprintf("1188%02x%02x%02x%02x0000", sequence, speedUnit, distMarker, distSub)
}

// OmniSetGreenSpeedCommand configures the Omni's green speed (command 0x89).
// greenSpeed: 0=8, 1=9, 2=10, 3=11, 4=12, 5=13
func OmniSetGreenSpeedCommand(sequence int, greenSpeed int) string {
	return fmt.Sprintf("1189%02x%02x00000000", sequence, greenSpeed)
}

// OmniSetCarryDistanceAdjustmentCommand configures carry distance adjustment (command 0x8a).
// adjustment is offset by +100 for encoding (e.g. 0 → 100, -5 → 95, +5 → 105).
func OmniSetCarryDistanceAdjustmentCommand(sequence int, adjustment int) string {
	encoded := byte(adjustment + 100)
	return fmt.Sprintf("118a%02x%02x00000000", sequence, encoded)
}

// OmniSetHandedCommand sends handedness to the Omni using clubSel=99 (command 0x82).
func OmniSetHandedCommand(sequence int, handedness HandednessType) string {
	return fmt.Sprintf("1182%02x0063%02x000000", sequence, handedness)
}
