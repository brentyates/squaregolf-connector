package audio

import (
	"fmt"
	"os/exec"
	"runtime"
)

// AudioPlayer handles platform-specific audio playback
type AudioPlayer struct {
	volume float64
}

// NewAudioPlayer creates a new audio player instance
func NewAudioPlayer() (*AudioPlayer, error) {
	return &AudioPlayer{
		volume: 0.5,
	}, nil
}

// SetVolume sets the volume level (0.0 to 1.0)
func (ap *AudioPlayer) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}
	ap.volume = volume
}

// PlayChime plays the ball detection chime
func (ap *AudioPlayer) PlayChime() error {
	switch runtime.GOOS {
	case "darwin":
		return ap.playChimeMac()
	case "windows":
		return ap.playChimeWindows()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// playChimeMac plays the chime on macOS using the afplay command
func (ap *AudioPlayer) playChimeMac() error {
	// Use the system's built-in chime sound
	cmd := fmt.Sprintf("afplay /System/Library/Sounds/Glass.aiff -v %f", ap.volume)
	return exec.Command("sh", "-c", cmd).Run()
}

// playChimeWindows plays the chime on Windows using PowerShell
func (ap *AudioPlayer) playChimeWindows() error {
	// Use PowerShell to play the system's default notification sound
	cmd := fmt.Sprintf(`[System.Media.SystemSounds]::Asterisk.Play()`)
	return exec.Command("powershell", "-Command", cmd).Run()
}
