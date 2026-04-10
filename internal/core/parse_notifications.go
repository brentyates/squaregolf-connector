package core

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
)

// SensorData represents data from the sensor
type SensorData struct {
	RawData      []string `json:"rawData,omitempty"`
	BallReady    bool     `json:"ballReady"`
	BallDetected bool     `json:"ballDetected"`
	PositionX    int32    `json:"positionX"`
	PositionY    int32    `json:"positionY"`
	PositionZ    int32    `json:"positionZ"`
}

// BallMetrics represents ball metrics from a shot
type BallMetrics struct {
	RawData          []string `json:"rawData,omitempty"`
	BallSpeedMPS     float64  `json:"speed"`
	VerticalAngle    float64  `json:"launchAngle"`
	HorizontalAngle  float64  `json:"horizontalAngle"`
	TotalspinRPM     int16    `json:"totalSpin"`
	SpinAxis         float64  `json:"spinAxis"`
	BackspinRPM      int16    `json:"backSpin"`
	SidespinRPM      int16    `json:"sideSpin"`
	IsBallSpeedValid bool     `json:"isBallSpeedValid"`
	IsTotalSpinValid bool     `json:"isTotalSpinValid"`
	IsSpinAxisValid  bool     `json:"isSpinAxisValid"`
	IsBackspinValid  bool     `json:"isBackSpinValid"`
	IsSidespinValid  bool     `json:"isSideSpinValid"`
	ShotType         ShotType `json:"shotType"`
}

// ClubMetrics represents club metrics from a shot
type ClubMetrics struct {
	RawData                 []string `json:"rawData,omitempty"`
	PathAngle               float64  `json:"path"`
	FaceAngle               float64  `json:"angle"`
	AttackAngle             float64  `json:"attackAngle"`
	DynamicLoftAngle        float64  `json:"dynamicLoft"`
	ImpactHorizontal        float64  `json:"impactHorizontal"`
	ImpactVertical          float64  `json:"impactVertical"`
	ClubSpeed               float64  `json:"clubSpeed"`
	SmashFactor             float64  `json:"smashFactor"`
	IsPathAngleValid        bool     `json:"isPathValid"`
	IsFaceAngleValid        bool     `json:"isFaceAngleValid"`
	IsAttackAngleValid      bool     `json:"isAttackAngleValid"`
	IsDynamicLoftValid      bool     `json:"isDynamicLoftValid"`
	IsImpactHorizontalValid bool     `json:"isImpactHorizontalValid"`
	IsImpactVerticalValid   bool     `json:"isImpactVerticalValid"`
	IsClubSpeedValid        bool     `json:"isClubSpeedValid"`
	IsSmashFactorValid      bool     `json:"isSmashFactorValid"`
}

// AlignmentData represents device alignment/aim information
type AlignmentData struct {
	RawData   []string `json:"rawData,omitempty"`
	AimAngle  float64  `json:"aimAngle"`  // Degrees left (negative) or right (positive) of center
	IsAligned bool     `json:"isAligned"` // Whether device is pointing at target (within ±2° threshold)
}

// ParseSensorData parses raw sensor data bytes
func ParseSensorData(bytesList []string) (*SensorData, error) {
	if len(bytesList) < 17 {
		return nil, fmt.Errorf("insufficient data for parsing sensor data")
	}

	sensorData := &SensorData{
		RawData:      bytesList,
		BallReady:    bytesList[3] == "01" || bytesList[3] == "02",
		BallDetected: bytesList[4] == "01",
	}

	// Parse position data
	posXBytes, err := hex.DecodeString(bytesList[5] + bytesList[6] + bytesList[7] + bytesList[8])
	if err == nil && len(posXBytes) == 4 {
		sensorData.PositionX = int32(binary.LittleEndian.Uint32(posXBytes))
	}

	posYBytes, err := hex.DecodeString(bytesList[9] + bytesList[10] + bytesList[11] + bytesList[12])
	if err == nil && len(posYBytes) == 4 {
		sensorData.PositionY = int32(binary.LittleEndian.Uint32(posYBytes))
	}

	posZBytes, err := hex.DecodeString(bytesList[13] + bytesList[14] + bytesList[15] + bytesList[16])
	if err == nil && len(posZBytes) == 4 {
		sensorData.PositionZ = int32(binary.LittleEndian.Uint32(posZBytes))
	}

	return sensorData, nil
}

// ParseShotBallMetrics parses ball metrics from shot data
func ParseShotBallMetrics(bytesList []string) (*BallMetrics, error) {
	if len(bytesList) < 17 {
		return nil, fmt.Errorf("insufficient data for parsing ball metrics")
	}

	metrics := &BallMetrics{
		RawData:          bytesList,
		IsBallSpeedValid: true,
		IsTotalSpinValid: true,
		IsSpinAxisValid:  true,
		IsBackspinValid:  true,
		IsSidespinValid:  true,
	}

	// Parse ball speed
	if ballSpeed, valid, ok := parseScaledInt16Metric(bytesList[3], bytesList[4], 100.0); ok {
		metrics.BallSpeedMPS = ballSpeed
		metrics.IsBallSpeedValid = valid
	} else {
		metrics.IsBallSpeedValid = false
	}

	// Parse vertical angle
	if verticalAngle, _, ok := parseScaledInt16Metric(bytesList[5], bytesList[6], 100.0); ok {
		metrics.VerticalAngle = verticalAngle
	}

	// Parse horizontal angle
	if horizontalAngle, _, ok := parseScaledInt16Metric(bytesList[7], bytesList[8], 100.0); ok {
		metrics.HorizontalAngle = horizontalAngle
	}

	// Parse total spin
	if totalSpin, valid, ok := parseInt16Metric(bytesList[9], bytesList[10]); ok {
		metrics.TotalspinRPM = totalSpin
		metrics.IsTotalSpinValid = valid
	} else {
		metrics.IsTotalSpinValid = false
	}

	// Parse spin axis
	if spinAxis, valid, ok := parseScaledInt16Metric(bytesList[11], bytesList[12], 100.0); ok {
		metrics.SpinAxis = spinAxis
		metrics.IsSpinAxisValid = valid
	} else {
		metrics.IsSpinAxisValid = false
	}

	// Parse backspin
	if backspin, valid, ok := parseInt16Metric(bytesList[13], bytesList[14]); ok {
		metrics.BackspinRPM = backspin
		metrics.IsBackspinValid = valid
	} else {
		metrics.IsBackspinValid = false
	}

	// Parse sidespin
	if sidespin, valid, ok := parseInt16Metric(bytesList[15], bytesList[16]); ok {
		metrics.SidespinRPM = sidespin
		metrics.IsSidespinValid = valid
	} else {
		metrics.IsSidespinValid = false
	}

	return metrics, nil
}

// ParseShotClubMetrics parses club metrics from shot data
func ParseShotClubMetrics(bytesList []string) (*ClubMetrics, error) {
	if len(bytesList) < 11 {
		return nil, fmt.Errorf("insufficient data for parsing club metrics")
	}

	metrics := &ClubMetrics{
		RawData:            bytesList,
		IsPathAngleValid:   true,
		IsFaceAngleValid:   true,
		IsAttackAngleValid: true,
		IsDynamicLoftValid: true,
	}

	// Parse path angle
	if pathAngle, valid, ok := parseScaledInt16Metric(bytesList[3], bytesList[4], 100.0); ok {
		metrics.PathAngle = pathAngle
		metrics.IsPathAngleValid = valid
	} else {
		metrics.IsPathAngleValid = false
	}

	// Parse face angle
	if faceAngle, valid, ok := parseScaledInt16Metric(bytesList[5], bytesList[6], 100.0); ok {
		metrics.FaceAngle = faceAngle
		metrics.IsFaceAngleValid = valid
	} else {
		metrics.IsFaceAngleValid = false
	}

	// Parse attack angle
	if attackAngle, valid, ok := parseScaledInt16Metric(bytesList[7], bytesList[8], 100.0); ok {
		metrics.AttackAngle = attackAngle
		metrics.IsAttackAngleValid = valid
	} else {
		metrics.IsAttackAngleValid = false
	}

	// Parse dynamic loft angle
	if loftAngle, valid, ok := parseScaledInt16Metric(bytesList[9], bytesList[10], 100.0); ok {
		metrics.DynamicLoftAngle = loftAngle
		metrics.IsDynamicLoftValid = valid
	} else {
		metrics.IsDynamicLoftValid = false
	}

	return metrics, nil
}

// ParseOmniShotClubMetrics parses club metrics from an Omni device (8 fields with validity bitmask)
func ParseOmniShotClubMetrics(bytesList []string) (*ClubMetrics, error) {
	if len(bytesList) < 19 {
		return nil, fmt.Errorf("insufficient data for parsing Omni club metrics (need 19, got %d)", len(bytesList))
	}

	validityByte, err := strconv.ParseUint(bytesList[2], 16, 8)
	if err != nil {
		validityByte = 0
	}
	validity := byte(validityByte)

	metrics := &ClubMetrics{
		RawData: bytesList,
	}

	type fieldDef struct {
		lowIdx, highIdx int
		target          *float64
		validTarget     *bool
		bit             uint
	}

	fields := []fieldDef{
		{3, 4, &metrics.PathAngle, &metrics.IsPathAngleValid, 0},
		{5, 6, &metrics.FaceAngle, &metrics.IsFaceAngleValid, 1},
		{7, 8, &metrics.AttackAngle, &metrics.IsAttackAngleValid, 2},
		{9, 10, &metrics.DynamicLoftAngle, &metrics.IsDynamicLoftValid, 3},
		{11, 12, &metrics.ImpactHorizontal, &metrics.IsImpactHorizontalValid, 4},
		{13, 14, &metrics.ImpactVertical, &metrics.IsImpactVerticalValid, 5},
		{15, 16, &metrics.ClubSpeed, &metrics.IsClubSpeedValid, 6},
		{17, 18, &metrics.SmashFactor, &metrics.IsSmashFactorValid, 7},
	}

	for _, f := range fields {
		bitmaskValid := (validity & (1 << f.bit)) != 0
		if val, sentinelValid, ok := parseScaledInt16Metric(bytesList[f.lowIdx], bytesList[f.highIdx], 100.0); ok {
			*f.target = val
			*f.validTarget = bitmaskValid && sentinelValid
		}
	}

	return metrics, nil
}

func parseInt16Metric(lowByte, highByte string) (int16, bool, bool) {
	metricBytes, err := hex.DecodeString(lowByte + highByte)
	if err != nil || len(metricBytes) != 2 {
		return 0, false, false
	}

	value := int16(binary.LittleEndian.Uint16(metricBytes))
	if value == -32768 {
		return 0, false, true
	}

	return value, true, true
}

func parseScaledInt16Metric(lowByte, highByte string, scale float64) (float64, bool, bool) {
	value, valid, ok := parseInt16Metric(lowByte, highByte)
	if !ok {
		return 0, false, false
	}

	return float64(value) / scale, valid, true
}

// ParseAlignmentData parses alignment/aim data from device accelerometer
func ParseAlignmentData(bytesList []string) (*AlignmentData, error) {
	// Format: 11 04 {seq} {status} 00 {angle_int16} ...
	// Angle is signed 16-bit little-endian at bytes 5-6, divided by 100.0
	// Negative = left, positive = right
	if len(bytesList) < 7 {
		return nil, fmt.Errorf("insufficient data for parsing alignment data (need at least 7 bytes, got %d)", len(bytesList))
	}

	alignment := &AlignmentData{
		RawData: bytesList,
	}

	angleBytes, err := hex.DecodeString(bytesList[5] + bytesList[6])
	if err == nil && len(angleBytes) == 2 {
		angleRaw := int16(binary.LittleEndian.Uint16(angleBytes))
		alignment.AimAngle = float64(angleRaw) / 100.0
	}

	const alignmentThreshold = 2.0
	alignment.IsAligned = alignment.AimAngle >= -alignmentThreshold && alignment.AimAngle <= alignmentThreshold

	return alignment, nil
}
