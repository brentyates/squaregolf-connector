package core

import (
	"testing"
)

func TestHeartbeatCommand(t *testing.T) {
	tests := []struct {
		name     string
		sequence int
		expected string
	}{
		{"Sequence 0", 0, "1183000000000000"},
		{"Sequence 15", 15, "11830f0000000000"},
		{"Sequence 255", 255, "1183ff0000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HeartbeatCommand(tt.sequence)
			if result != tt.expected {
				t.Errorf("HeartbeatCommand(%d) = %s, want %s", tt.sequence, result, tt.expected)
			}
		})
	}
}

func TestDetectBallCommand(t *testing.T) {
	tests := []struct {
		name     string
		sequence int
		mode     DetectBallMode
		spinMode SpinMode
		expected string
	}{
		{"Deactivate Standard", 0, Deactivate, Standard, "118100001000000000"},
		{"Activate Standard", 5, Activate, Standard, "118105011000000000"},
		{"Deactivate Advanced", 10, Deactivate, Advanced, "11810a001100000000"},
		{"Activate Advanced", 255, Activate, Advanced, "1181ff011100000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectBallCommand(tt.sequence, tt.mode, tt.spinMode)
			if result != tt.expected {
				t.Errorf("DetectBallCommand(%d, %d, %d) = %s, want %s",
					tt.sequence, tt.mode, tt.spinMode, result, tt.expected)
			}
		})
	}
}

func TestClubCommand(t *testing.T) {
	tests := []struct {
		name       string
		sequence   int
		club       ClubType
		handedness HandednessType
		expected   string
	}{
		{"Putter RightHanded", 0, ClubPutter, RightHanded, "118200010700000000"},
		{"Driver LeftHanded", 5, ClubDriver, LeftHanded, "118205020401000000"},
		{"Iron7 RightHanded", 10, ClubIron7, RightHanded, "11820a070600000000"},
		{"SandWedge LeftHanded", 255, ClubSandWedge, LeftHanded, "1182ff0c0601000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClubCommand(tt.sequence, tt.club, tt.handedness)
			if result != tt.expected {
				t.Errorf("ClubCommand(%d, %v, %d) = %s, want %s",
					tt.sequence, tt.club, tt.handedness, result, tt.expected)
			}
		})
	}
}

func TestSwingStickCommand(t *testing.T) {
	tests := []struct {
		name       string
		sequence   int
		club       ClubType
		handedness HandednessType
		expected   string
	}{
		{"Putter RightHanded", 0, ClubPutter, RightHanded, "1182000103000000"},
		{"Driver LeftHanded", 5, ClubDriver, LeftHanded, "1182050202010000"},
		{"Iron7 RightHanded", 10, ClubIron7, RightHanded, "11820a0700000000"},
		{"SandWedge LeftHanded", 255, ClubSandWedge, LeftHanded, "1182ff0c00010000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SwingStickCommand(tt.sequence, tt.club, tt.handedness)
			if result != tt.expected {
				t.Errorf("SwingStickCommand(%d, %v, %d) = %s, want %s",
					tt.sequence, tt.club, tt.handedness, result, tt.expected)
			}
		})
	}
}

func TestAlignmentStickCommand(t *testing.T) {
	tests := []struct {
		name       string
		sequence   int
		handedness HandednessType
		expected   string
	}{
		{"RightHanded", 0, RightHanded, "118200080800000000"},
		{"LeftHanded", 5, LeftHanded, "118205080801000000"},
		{"RightHanded Max Seq", 255, RightHanded, "1182ff080800000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AlignmentStickCommand(tt.sequence, tt.handedness)
			if result != tt.expected {
				t.Errorf("AlignmentStickCommand(%d, %d) = %s, want %s",
					tt.sequence, tt.handedness, result, tt.expected)
			}
		})
	}
}

func TestRequestClubMetricsCommand(t *testing.T) {
	tests := []struct {
		name     string
		sequence int
		expected string
	}{
		{"Sequence 0", 0, "118700000000000000"},
		{"Sequence 15", 15, "11870f000000000000"},
		{"Sequence 255", 255, "1187ff000000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequestClubMetricsCommand(tt.sequence)
			if result != tt.expected {
				t.Errorf("RequestClubMetricsCommand(%d) = %s, want %s", tt.sequence, result, tt.expected)
			}
		})
	}
}

func TestHandednessTypeString(t *testing.T) {
	tests := []struct {
		handedness HandednessType
		expected   int
	}{
		{RightHanded, 0},
		{LeftHanded, 1},
	}

	for _, tt := range tests {
		t.Run("HandednessValue", func(t *testing.T) {
			if int(tt.handedness) != tt.expected {
				t.Errorf("HandednessType %v has value %d, want %d", tt.handedness, tt.handedness, tt.expected)
			}
		})
	}
}

func TestDetectBallModeString(t *testing.T) {
	tests := []struct {
		mode     DetectBallMode
		expected int
	}{
		{Deactivate, 0},
		{Activate, 1},
	}

	for _, tt := range tests {
		t.Run("DetectBallModeValue", func(t *testing.T) {
			if int(tt.mode) != tt.expected {
				t.Errorf("DetectBallMode %v has value %d, want %d", tt.mode, tt.mode, tt.expected)
			}
		})
	}
}

func TestSpinModeString(t *testing.T) {
	tests := []struct {
		mode     SpinMode
		expected int
	}{
		{Standard, 0},
		{Advanced, 1},
	}

	for _, tt := range tests {
		t.Run("SpinModeValue", func(t *testing.T) {
			if int(tt.mode) != tt.expected {
				t.Errorf("SpinMode %v has value %d, want %d", tt.mode, tt.mode, tt.expected)
			}
		})
	}
}

func TestClubTypeValues(t *testing.T) {
	// Test a subset of clubs to verify their code values
	if ClubPutter.RegularCode != "0107" || ClubPutter.SwingStickCode != "0103" {
		t.Errorf("ClubPutter has incorrect codes: %v", ClubPutter)
	}

	if ClubDriver.RegularCode != "0204" || ClubDriver.SwingStickCode != "0202" {
		t.Errorf("ClubDriver has incorrect codes: %v", ClubDriver)
	}

	if ClubSandWedge.RegularCode != "0c06" || ClubSandWedge.SwingStickCode != "0c00" {
		t.Errorf("ClubSandWedge has incorrect codes: %v", ClubSandWedge)
	}
}
