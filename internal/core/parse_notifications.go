package core

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

// SensorData represents data from the sensor
type SensorData struct {
	RawData      []string
	BallReady    bool
	BallDetected bool
	PositionX    int32
	PositionY    int32
	PositionZ    int32
}

// BallMetrics represents ball metrics from a shot
type BallMetrics struct {
	RawData         []string
	BallSpeedMPS    float64
	VerticalAngle   float64
	HorizontalAngle float64
	TotalspinRPM    int16
	SpinAxis        float64
	BackspinRPM     int16
	SidespinRPM     int16
	ShotType        ShotType
}

// ClubMetrics represents club metrics from a shot
type ClubMetrics struct {
	RawData          []string
	PathAngle        float64
	FaceAngle        float64
	AttackAngle      float64
	DynamicLoftAngle float64
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
		RawData: bytesList,
	}

	// Determine shot type from header
	if len(bytesList) >= 3 {
		if bytesList[2] == "37" {
			metrics.ShotType = ShotTypeFull
		} else if bytesList[2] == "13" {
			metrics.ShotType = ShotTypePutt
		}
	}

	// Parse ball speed
	ballSpeedBytes, err := hex.DecodeString(bytesList[3] + bytesList[4])
	if err == nil && len(ballSpeedBytes) == 2 {
		metrics.BallSpeedMPS = float64(int16(binary.LittleEndian.Uint16(ballSpeedBytes))) / 100.0
	}

	// Parse vertical angle
	verticalAngleBytes, err := hex.DecodeString(bytesList[5] + bytesList[6])
	if err == nil && len(verticalAngleBytes) == 2 {
		metrics.VerticalAngle = float64(int16(binary.LittleEndian.Uint16(verticalAngleBytes))) / 100.0
	}

	// Parse horizontal angle
	horizontalAngleBytes, err := hex.DecodeString(bytesList[7] + bytesList[8])
	if err == nil && len(horizontalAngleBytes) == 2 {
		metrics.HorizontalAngle = float64(int16(binary.LittleEndian.Uint16(horizontalAngleBytes))) / 100.0
	}

	// Parse total spin
	totalSpinBytes, err := hex.DecodeString(bytesList[9] + bytesList[10])
	if err == nil && len(totalSpinBytes) == 2 {
		metrics.TotalspinRPM = int16(binary.LittleEndian.Uint16(totalSpinBytes))
	}

	// Parse spin axis
	spinAxisBytes, err := hex.DecodeString(bytesList[11] + bytesList[12])
	if err == nil && len(spinAxisBytes) == 2 {
		metrics.SpinAxis = float64(int16(binary.LittleEndian.Uint16(spinAxisBytes))) / 100.0
	}

	// Parse backspin
	backspinBytes, err := hex.DecodeString(bytesList[13] + bytesList[14])
	if err == nil && len(backspinBytes) == 2 {
		metrics.BackspinRPM = int16(binary.LittleEndian.Uint16(backspinBytes))
	}

	// Parse sidespin
	sidespinBytes, err := hex.DecodeString(bytesList[15] + bytesList[16])
	if err == nil && len(sidespinBytes) == 2 {
		metrics.SidespinRPM = int16(binary.LittleEndian.Uint16(sidespinBytes))
	}

	return metrics, nil
}

// ParseShotClubMetrics parses club metrics from shot data
func ParseShotClubMetrics(bytesList []string) (*ClubMetrics, error) {
	if len(bytesList) < 11 {
		return nil, fmt.Errorf("insufficient data for parsing club metrics")
	}

	metrics := &ClubMetrics{
		RawData: bytesList,
	}

	// Parse path angle
	pathAngleBytes, err := hex.DecodeString(bytesList[3] + bytesList[4])
	if err == nil && len(pathAngleBytes) == 2 {
		metrics.PathAngle = float64(int16(binary.LittleEndian.Uint16(pathAngleBytes))) / 100.0
	}

	// Parse face angle
	faceAngleBytes, err := hex.DecodeString(bytesList[5] + bytesList[6])
	if err == nil && len(faceAngleBytes) == 2 {
		metrics.FaceAngle = float64(int16(binary.LittleEndian.Uint16(faceAngleBytes))) / 100.0
	}

	// Parse attack angle
	attackAngleBytes, err := hex.DecodeString(bytesList[7] + bytesList[8])
	if err == nil && len(attackAngleBytes) == 2 {
		metrics.AttackAngle = float64(int16(binary.LittleEndian.Uint16(attackAngleBytes))) / 100.0
	}

	// Parse dynamic loft angle
	loftAngleBytes, err := hex.DecodeString(bytesList[9] + bytesList[10])
	if err == nil && len(loftAngleBytes) == 2 {
		metrics.DynamicLoftAngle = float64(int16(binary.LittleEndian.Uint16(loftAngleBytes))) / 100.0
	}

	return metrics, nil
}
