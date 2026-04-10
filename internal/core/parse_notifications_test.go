package core

import (
	"reflect"
	"testing"
)

func TestParseSensorData(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []string
		want    *SensorData
		wantErr bool
	}{
		{
			name:    "Insufficient data",
			bytes:   []string{"00", "01", "02"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Ball ready and not detected",
			bytes: []string{
				"00", "01", "02", "01", "00",
				"01", "00", "00", "00",
				"02", "00", "00", "00",
				"03", "00", "00", "00",
			},
			want: &SensorData{
				RawData: []string{
					"00", "01", "02", "01", "00",
					"01", "00", "00", "00",
					"02", "00", "00", "00",
					"03", "00", "00", "00",
				},
				BallReady:    true,
				BallDetected: false,
				PositionX:    1,
				PositionY:    2,
				PositionZ:    3,
			},
			wantErr: false,
		},
		{
			name: "Ball ready (value 2) and detected",
			bytes: []string{
				"00", "01", "02", "02", "01",
				"0A", "00", "00", "00",
				"14", "00", "00", "00",
				"1E", "00", "00", "00",
			},
			want: &SensorData{
				RawData: []string{
					"00", "01", "02", "02", "01",
					"0A", "00", "00", "00",
					"14", "00", "00", "00",
					"1E", "00", "00", "00",
				},
				BallReady:    true,
				BallDetected: true,
				PositionX:    10,
				PositionY:    20,
				PositionZ:    30,
			},
			wantErr: false,
		},
		{
			name: "Ball not ready",
			bytes: []string{
				"00", "01", "02", "00", "00",
				"FF", "FF", "FF", "FF",
				"FF", "FF", "FF", "FF",
				"FF", "FF", "FF", "FF",
			},
			want: &SensorData{
				RawData: []string{
					"00", "01", "02", "00", "00",
					"FF", "FF", "FF", "FF",
					"FF", "FF", "FF", "FF",
					"FF", "FF", "FF", "FF",
				},
				BallReady:    false,
				BallDetected: false,
				PositionX:    -1,
				PositionY:    -1,
				PositionZ:    -1,
			},
			wantErr: false,
		},
		{
			name: "Invalid position X hex data",
			bytes: []string{
				"00", "01", "02", "01", "00",
				"ZZ", "00", "00", "00",
				"02", "00", "00", "00",
				"03", "00", "00", "00",
			},
			want: &SensorData{
				RawData: []string{
					"00", "01", "02", "01", "00",
					"ZZ", "00", "00", "00",
					"02", "00", "00", "00",
					"03", "00", "00", "00",
				},
				BallReady:    true,
				BallDetected: false,
				PositionX:    0,
				PositionY:    2,
				PositionZ:    3,
			},
			wantErr: false,
		},
		{
			name: "Invalid position Y hex data",
			bytes: []string{
				"00", "01", "02", "01", "00",
				"01", "00", "00", "00",
				"ZZ", "00", "00", "00",
				"03", "00", "00", "00",
			},
			want: &SensorData{
				RawData: []string{
					"00", "01", "02", "01", "00",
					"01", "00", "00", "00",
					"ZZ", "00", "00", "00",
					"03", "00", "00", "00",
				},
				BallReady:    true,
				BallDetected: false,
				PositionX:    1,
				PositionY:    0,
				PositionZ:    3,
			},
			wantErr: false,
		},
		{
			name: "Invalid position Z hex data",
			bytes: []string{
				"00", "01", "02", "01", "00",
				"01", "00", "00", "00",
				"02", "00", "00", "00",
				"ZZ", "00", "00", "00",
			},
			want: &SensorData{
				RawData: []string{
					"00", "01", "02", "01", "00",
					"01", "00", "00", "00",
					"02", "00", "00", "00",
					"ZZ", "00", "00", "00",
				},
				BallReady:    true,
				BallDetected: false,
				PositionX:    1,
				PositionY:    2,
				PositionZ:    0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSensorData(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSensorData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseSensorData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseShotBallMetrics(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []string
		want    *BallMetrics
		wantErr bool
	}{
		{
			name:    "Insufficient data",
			bytes:   []string{"00", "01", "02"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Valid ball metrics data",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"E8", "03", // Total spin: 1000 RPM
				"F4", "01", // Spin axis: 500 (5.00)
				"D0", "07", // Backspin: 2000 RPM
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"E8", "03", // Total spin: 1000 RPM
					"F4", "01", // Spin axis: 500 (5.00)
					"D0", "07", // Backspin: 2000 RPM
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Negative values",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"9C", "FF", // Ball speed: -100 (-1.00 m/s)
				"38", "FF", // Vertical angle: -200 (-2.00 degrees)
				"D4", "FE", // Horizontal angle: -300 (-3.00 degrees)
				"18", "FC", // Total spin: -1000 RPM
				"0C", "FE", // Spin axis: -500 (-5.00)
				"30", "F8", // Backspin: -2000 RPM
				"48", "F4", // Sidespin: -3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"9C", "FF", // Ball speed: -100 (-1.00 m/s)
					"38", "FF", // Vertical angle: -200 (-2.00 degrees)
					"D4", "FE", // Horizontal angle: -300 (-3.00 degrees)
					"18", "FC", // Total spin: -1000 RPM
					"0C", "FE", // Spin axis: -500 (-5.00)
					"30", "F8", // Backspin: -2000 RPM
					"48", "F4", // Sidespin: -3000 RPM
				},
				BallSpeedMPS:     -1.0,
				VerticalAngle:    -2.0,
				HorizontalAngle:  -3.0,
				TotalspinRPM:     -1000,
				SpinAxis:         -5.0,
				BackspinRPM:      -2000,
				SidespinRPM:      -3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for ball speed",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"ZZ", "00", // Invalid ball speed
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"E8", "03", // Total spin: 1000 RPM
				"F4", "01", // Spin axis: 500 (5.00)
				"D0", "07", // Backspin: 2000 RPM
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"ZZ", "00", // Invalid ball speed
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"E8", "03", // Total spin: 1000 RPM
					"F4", "01", // Spin axis: 500 (5.00)
					"D0", "07", // Backspin: 2000 RPM
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      3000,
				IsBallSpeedValid: false,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for vertical angle",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"ZZ", "00", // Invalid vertical angle
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"E8", "03", // Total spin: 1000 RPM
				"F4", "01", // Spin axis: 500 (5.00)
				"D0", "07", // Backspin: 2000 RPM
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"ZZ", "00", // Invalid vertical angle
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"E8", "03", // Total spin: 1000 RPM
					"F4", "01", // Spin axis: 500 (5.00)
					"D0", "07", // Backspin: 2000 RPM
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for horizontal angle",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"ZZ", "01", // Invalid horizontal angle
				"E8", "03", // Total spin: 1000 RPM
				"F4", "01", // Spin axis: 500 (5.00)
				"D0", "07", // Backspin: 2000 RPM
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"ZZ", "01", // Invalid horizontal angle
					"E8", "03", // Total spin: 1000 RPM
					"F4", "01", // Spin axis: 500 (5.00)
					"D0", "07", // Backspin: 2000 RPM
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for total spin",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"ZZ", "03", // Invalid total spin
				"F4", "01", // Spin axis: 500 (5.00)
				"D0", "07", // Backspin: 2000 RPM
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"ZZ", "03", // Invalid total spin
					"F4", "01", // Spin axis: 500 (5.00)
					"D0", "07", // Backspin: 2000 RPM
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     0,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: false,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for spin axis",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"E8", "03", // Total spin: 1000 RPM
				"ZZ", "01", // Invalid spin axis
				"D0", "07", // Backspin: 2000 RPM
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"E8", "03", // Total spin: 1000 RPM
					"ZZ", "01", // Invalid spin axis
					"D0", "07", // Backspin: 2000 RPM
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         0,
				BackspinRPM:      2000,
				SidespinRPM:      3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  false,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for backspin",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"E8", "03", // Total spin: 1000 RPM
				"F4", "01", // Spin axis: 500 (5.00)
				"ZZ", "07", // Invalid backspin
				"B8", "0B", // Sidespin: 3000 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"E8", "03", // Total spin: 1000 RPM
					"F4", "01", // Spin axis: 500 (5.00)
					"ZZ", "07", // Invalid backspin
					"B8", "0B", // Sidespin: 3000 RPM
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      0,
				SidespinRPM:      3000,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  false,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for sidespin",
			bytes: []string{
				"11", "02", "37", // Full shot header
				"64", "00", // Ball speed: 100 (1.00 m/s)
				"C8", "00", // Vertical angle: 200 (2.00 degrees)
				"2C", "01", // Horizontal angle: 300 (3.00 degrees)
				"E8", "03", // Total spin: 1000 RPM
				"F4", "01", // Spin axis: 500 (5.00)
				"D0", "07", // Backspin: 2000 RPM
				"ZZ", "0B", // Invalid sidespin
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37", // Full shot header
					"64", "00", // Ball speed: 100 (1.00 m/s)
					"C8", "00", // Vertical angle: 200 (2.00 degrees)
					"2C", "01", // Horizontal angle: 300 (3.00 degrees)
					"E8", "03", // Total spin: 1000 RPM
					"F4", "01", // Spin axis: 500 (5.00)
					"D0", "07", // Backspin: 2000 RPM
					"ZZ", "0B", // Invalid sidespin
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      0,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  false,
			},
			wantErr: false,
		},
		{
			name: "Short putt",
			bytes: []string{
				"11", "02", "13", // Putt header
				"6B", "00", // Ball speed: 107 (1.07 m/s)
				"00", "00", // Vertical angle: 0 (0.00 degrees)
				"42", "00", // Horizontal angle: 66 (0.66 degrees)
				"4B", "00", // Total spin: 75 RPM
				"00", "00", // Spin axis: 0 (0.00)
				"00", "00", // Backspin: 0 RPM
				"00", "00", // Sidespin: 0 RPM
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "13", // Putt header
					"6B", "00", // Ball speed: 107 (1.07 m/s)
					"00", "00", // Vertical angle: 0 (0.00 degrees)
					"42", "00", // Horizontal angle: 66 (0.66 degrees)
					"4B", "00", // Total spin: 75 RPM
					"00", "00", // Spin axis: 0 (0.00)
					"00", "00", // Backspin: 0 RPM
					"00", "00", // Sidespin: 0 RPM
				},
				BallSpeedMPS:     1.07,
				VerticalAngle:    0,
				HorizontalAngle:  0.66,
				TotalspinRPM:     75,
				SpinAxis:         0,
				BackspinRPM:      0,
				SidespinRPM:      0,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  true,
			},
			wantErr: false,
		},
		{
			name: "Sentinel sidespin normalizes to zero and invalid",
			bytes: []string{
				"11", "02", "37",
				"64", "00",
				"C8", "00",
				"2C", "01",
				"E8", "03",
				"F4", "01",
				"D0", "07",
				"00", "80",
			},
			want: &BallMetrics{
				RawData: []string{
					"11", "02", "37",
					"64", "00",
					"C8", "00",
					"2C", "01",
					"E8", "03",
					"F4", "01",
					"D0", "07",
					"00", "80",
				},
				BallSpeedMPS:     1.0,
				VerticalAngle:    2.0,
				HorizontalAngle:  3.0,
				TotalspinRPM:     1000,
				SpinAxis:         5.0,
				BackspinRPM:      2000,
				SidespinRPM:      0,
				IsBallSpeedValid: true,
				IsTotalSpinValid: true,
				IsSpinAxisValid:  true,
				IsBackspinValid:  true,
				IsSidespinValid:  false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseShotBallMetrics(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseShotBallMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseShotBallMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseShotClubMetrics(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []string
		want    *ClubMetrics
		wantErr bool
	}{
		{
			name:    "Insufficient data",
			bytes:   []string{"00", "01", "02"},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Valid club metrics data",
			bytes: []string{
				"00", "01", "02",
				"64", "00", // Path angle: 100 (1.00 degrees)
				"C8", "00", // Face angle: 200 (2.00 degrees)
				"2C", "01", // Attack angle: 300 (3.00 degrees)
				"90", "01", // Dynamic loft: 400 (4.00 degrees)
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"64", "00", // Path angle: 100 (1.00 degrees)
					"C8", "00", // Face angle: 200 (2.00 degrees)
					"2C", "01", // Attack angle: 300 (3.00 degrees)
					"90", "01", // Dynamic loft: 400 (4.00 degrees)
				},
				PathAngle:          1.0,
				FaceAngle:          2.0,
				AttackAngle:        3.0,
				DynamicLoftAngle:   4.0,
				IsPathAngleValid:   true,
				IsFaceAngleValid:   true,
				IsAttackAngleValid: true,
				IsDynamicLoftValid: true,
			},
			wantErr: false,
		},
		{
			name: "Negative values",
			bytes: []string{
				"00", "01", "02",
				"9C", "FF", // Path angle: -100 (-1.00 degrees)
				"38", "FF", // Face angle: -200 (-2.00 degrees)
				"D4", "FE", // Attack angle: -300 (-3.00 degrees)
				"70", "FE", // Dynamic loft: -400 (-4.00 degrees)
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"9C", "FF", // Path angle: -100 (-1.00 degrees)
					"38", "FF", // Face angle: -200 (-2.00 degrees)
					"D4", "FE", // Attack angle: -300 (-3.00 degrees)
					"70", "FE", // Dynamic loft: -400 (-4.00 degrees)
				},
				PathAngle:          -1.0,
				FaceAngle:          -2.0,
				AttackAngle:        -3.0,
				DynamicLoftAngle:   -4.0,
				IsPathAngleValid:   true,
				IsFaceAngleValid:   true,
				IsAttackAngleValid: true,
				IsDynamicLoftValid: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for path angle",
			bytes: []string{
				"00", "01", "02",
				"ZZ", "00", // Invalid path angle
				"C8", "00", // Face angle: 200 (2.00 degrees)
				"2C", "01", // Attack angle: 300 (3.00 degrees)
				"90", "01", // Dynamic loft: 400 (4.00 degrees)
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"ZZ", "00", // Invalid path angle
					"C8", "00", // Face angle: 200 (2.00 degrees)
					"2C", "01", // Attack angle: 300 (3.00 degrees)
					"90", "01", // Dynamic loft: 400 (4.00 degrees)
				},
				PathAngle:          0,
				FaceAngle:          2.0,
				AttackAngle:        3.0,
				DynamicLoftAngle:   4.0,
				IsPathAngleValid:   false,
				IsFaceAngleValid:   true,
				IsAttackAngleValid: true,
				IsDynamicLoftValid: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for face angle",
			bytes: []string{
				"00", "01", "02",
				"64", "00", // Path angle: 100 (1.00 degrees)
				"ZZ", "00", // Invalid face angle
				"2C", "01", // Attack angle: 300 (3.00 degrees)
				"90", "01", // Dynamic loft: 400 (4.00 degrees)
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"64", "00", // Path angle: 100 (1.00 degrees)
					"ZZ", "00", // Invalid face angle
					"2C", "01", // Attack angle: 300 (3.00 degrees)
					"90", "01", // Dynamic loft: 400 (4.00 degrees)
				},
				PathAngle:          1.0,
				FaceAngle:          0,
				AttackAngle:        3.0,
				DynamicLoftAngle:   4.0,
				IsPathAngleValid:   true,
				IsFaceAngleValid:   false,
				IsAttackAngleValid: true,
				IsDynamicLoftValid: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for attack angle",
			bytes: []string{
				"00", "01", "02",
				"64", "00", // Path angle: 100 (1.00 degrees)
				"C8", "00", // Face angle: 200 (2.00 degrees)
				"ZZ", "01", // Invalid attack angle
				"90", "01", // Dynamic loft: 400 (4.00 degrees)
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"64", "00", // Path angle: 100 (1.00 degrees)
					"C8", "00", // Face angle: 200 (2.00 degrees)
					"ZZ", "01", // Invalid attack angle
					"90", "01", // Dynamic loft: 400 (4.00 degrees)
				},
				PathAngle:          1.0,
				FaceAngle:          2.0,
				AttackAngle:        0,
				DynamicLoftAngle:   4.0,
				IsPathAngleValid:   true,
				IsFaceAngleValid:   true,
				IsAttackAngleValid: false,
				IsDynamicLoftValid: true,
			},
			wantErr: false,
		},
		{
			name: "Invalid hex data for dynamic loft",
			bytes: []string{
				"00", "01", "02",
				"64", "00", // Path angle: 100 (1.00 degrees)
				"C8", "00", // Face angle: 200 (2.00 degrees)
				"2C", "01", // Attack angle: 300 (3.00 degrees)
				"ZZ", "01", // Invalid dynamic loft
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"64", "00", // Path angle: 100 (1.00 degrees)
					"C8", "00", // Face angle: 200 (2.00 degrees)
					"2C", "01", // Attack angle: 300 (3.00 degrees)
					"ZZ", "01", // Invalid dynamic loft
				},
				PathAngle:          1.0,
				FaceAngle:          2.0,
				AttackAngle:        3.0,
				DynamicLoftAngle:   0,
				IsPathAngleValid:   true,
				IsFaceAngleValid:   true,
				IsAttackAngleValid: true,
				IsDynamicLoftValid: false,
			},
			wantErr: false,
		},
		{
			name: "Sentinel path normalizes to zero and invalid",
			bytes: []string{
				"00", "01", "02",
				"00", "80",
				"C8", "00",
				"2C", "01",
				"90", "01",
			},
			want: &ClubMetrics{
				RawData: []string{
					"00", "01", "02",
					"00", "80",
					"C8", "00",
					"2C", "01",
					"90", "01",
				},
				PathAngle:          0,
				FaceAngle:          2.0,
				AttackAngle:        3.0,
				DynamicLoftAngle:   4.0,
				IsPathAngleValid:   false,
				IsFaceAngleValid:   true,
				IsAttackAngleValid: true,
				IsDynamicLoftValid: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseShotClubMetrics(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseShotClubMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseShotClubMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOmniShotClubMetrics(t *testing.T) {
	tests := []struct {
		name    string
		bytes   []string
		want    *ClubMetrics
		wantErr bool
	}{
		{
			name:    "Insufficient data",
			bytes:   []string{"11", "07", "ff", "00", "01", "00", "02"},
			wantErr: true,
		},
		{
			name: "All fields valid with bitmask 0xFF",
			bytes: []string{
				"11", "07", "ff",
				"d8", "fe", // path = -296 -> -2.96
				"90", "01", // face = 400 -> 4.00
				"38", "ff", // attack = -200 -> -2.00
				"d0", "07", // loft = 2000 -> 20.00
				"64", "00", // impactH = 100 -> 1.00
				"c8", "ff", // impactV = -56 -> -0.56
				"b8", "0b", // clubSpeed = 3000 -> 30.00
				"82", "00", // smash = 130 -> 1.30
			},
			want: &ClubMetrics{
				RawData: []string{
					"11", "07", "ff",
					"d8", "fe", "90", "01", "38", "ff", "d0", "07",
					"64", "00", "c8", "ff", "b8", "0b", "82", "00",
				},
				PathAngle: -2.96, FaceAngle: 4.00, AttackAngle: -2.00, DynamicLoftAngle: 20.00,
				ImpactHorizontal: 1.00, ImpactVertical: -0.56, ClubSpeed: 30.00, SmashFactor: 1.30,
				IsPathAngleValid: true, IsFaceAngleValid: true, IsAttackAngleValid: true, IsDynamicLoftValid: true,
				IsImpactHorizontalValid: true, IsImpactVerticalValid: true, IsClubSpeedValid: true, IsSmashFactorValid: true,
			},
		},
		{
			name: "Partial bitmask 0x0F (first 4 valid, last 4 invalid)",
			bytes: []string{
				"11", "07", "0f",
				"d8", "fe", "90", "01", "38", "ff", "d0", "07",
				"64", "00", "c8", "ff", "b8", "0b", "82", "00",
			},
			want: &ClubMetrics{
				RawData: []string{
					"11", "07", "0f",
					"d8", "fe", "90", "01", "38", "ff", "d0", "07",
					"64", "00", "c8", "ff", "b8", "0b", "82", "00",
				},
				PathAngle: -2.96, FaceAngle: 4.00, AttackAngle: -2.00, DynamicLoftAngle: 20.00,
				ImpactHorizontal: 1.00, ImpactVertical: -0.56, ClubSpeed: 30.00, SmashFactor: 1.30,
				IsPathAngleValid: true, IsFaceAngleValid: true, IsAttackAngleValid: true, IsDynamicLoftValid: true,
				IsImpactHorizontalValid: false, IsImpactVerticalValid: false, IsClubSpeedValid: false, IsSmashFactorValid: false,
			},
		},
		{
			name: "Sentinel value overrides bitmask",
			bytes: []string{
				"11", "07", "ff",
				"00", "80", // path = -32768 sentinel
				"90", "01", "38", "ff", "d0", "07",
				"64", "00", "c8", "ff", "b8", "0b", "82", "00",
			},
			want: &ClubMetrics{
				RawData: []string{
					"11", "07", "ff",
					"00", "80", "90", "01", "38", "ff", "d0", "07",
					"64", "00", "c8", "ff", "b8", "0b", "82", "00",
				},
				PathAngle: 0, FaceAngle: 4.00, AttackAngle: -2.00, DynamicLoftAngle: 20.00,
				ImpactHorizontal: 1.00, ImpactVertical: -0.56, ClubSpeed: 30.00, SmashFactor: 1.30,
				IsPathAngleValid: false, IsFaceAngleValid: true, IsAttackAngleValid: true, IsDynamicLoftValid: true,
				IsImpactHorizontalValid: true, IsImpactVerticalValid: true, IsClubSpeedValid: true, IsSmashFactorValid: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOmniShotClubMetrics(tt.bytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOmniShotClubMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseOmniShotClubMetrics() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestDetectDeviceType(t *testing.T) {
	tests := []struct {
		name       string
		mfgDataHex string
		want       DeviceType
	}{
		{"Empty data returns Home", "", DeviceTypeHome},
		{"Non-matching data returns Home", "aabbccddee", DeviceTypeHome},
		{"Omni identifier returns Omni", "some3033303041data", DeviceTypeOmni},
		{"Exact Omni identifier", OmniManufacturerDataHex, DeviceTypeOmni},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectDeviceType(tt.mfgDataHex)
			if got != tt.want {
				t.Errorf("DetectDeviceType(%q) = %v, want %v", tt.mfgDataHex, got, tt.want)
			}
		})
	}
}
