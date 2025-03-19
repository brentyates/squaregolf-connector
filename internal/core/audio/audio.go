package audio

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

//go:embed ready*.mp3
var audioFiles embed.FS

// AudioPlayer handles platform-specific audio playback
type AudioPlayer struct {
	volume float64
	ctx    *oto.Context
}

// NewAudioPlayer creates a new audio player instance
func NewAudioPlayer() (*AudioPlayer, error) {
	ctx, ready, err := oto.NewContext(44100, 2, 2)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio context: %w", err)
	}
	<-ready

	return &AudioPlayer{
		volume: 0.5,
		ctx:    ctx,
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

// GetAvailableSounds returns a list of available sound files
func (ap *AudioPlayer) GetAvailableSounds() []string {
	entries, err := audioFiles.ReadDir(".")
	if err != nil {
		return nil
	}

	var sounds []string
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) >= 5 && entry.Name()[:5] == "ready" {
			sounds = append(sounds, entry.Name())
		}
	}
	return sounds
}

// PlaySound plays the selected sound file
func (ap *AudioPlayer) PlaySound(filename string) error {
	// Read the embedded MP3 file
	data, err := audioFiles.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read embedded audio file: %w", err)
	}

	// Create an MP3 decoder
	decoder, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create MP3 decoder: %w", err)
	}

	// Create a new player
	player := ap.ctx.NewPlayer(decoder)
	defer player.Close()

	// Set volume
	player.SetVolume(ap.volume)

	// Play the audio
	player.Play()

	// Wait for playback to complete
	for player.IsPlaying() {
		// Wait
	}

	return nil
}

// PlayChime plays the default ball detection chime (ready1.mp3)
func (ap *AudioPlayer) PlayChime() error {
	return ap.PlaySound("ready1.mp3")
}
