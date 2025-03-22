package core

import (
	"log"
	"strings"
	"sync"

	"github.com/brentyates/squaregolf-connector/internal/core/audio"
)

var (
	chimeManagerInstance *ChimeManager
	chimeManagerOnce     sync.Once
)

// GetChimeManagerInstance returns the singleton instance of ChimeManager
func GetChimeManagerInstance(stateManager *StateManager) *ChimeManager {
	chimeManagerOnce.Do(func() {
		audioPlayer, err := audio.NewAudioPlayer()
		if err != nil {
			// Log the error but continue without audio
			log.Printf("Error initializing audio player: %v - application will run without sound capabilities", err)
			chimeManagerInstance = &ChimeManager{
				stateManager: stateManager,
				volume:       stateManager.GetChimeVolume(),
			}
		} else {
			chimeManagerInstance = &ChimeManager{
				audioPlayer:  audioPlayer,
				stateManager: stateManager,
				volume:       stateManager.GetChimeVolume(),
			}
		}
	})
	return chimeManagerInstance
}

// NewChimeManager is deprecated, use GetChimeManagerInstance instead
func NewChimeManager(stateManager *StateManager) *ChimeManager {
	return GetChimeManagerInstance(stateManager)
}

// ChimeManager handles the chime functionality
type ChimeManager struct {
	audioPlayer  *audio.AudioPlayer
	stateManager *StateManager
	volume       float64
}

// Initialize sets up the chime manager and registers callbacks
func (cm *ChimeManager) Initialize() {
	cm.stateManager.RegisterBallReadyCallback(func(oldValue, newValue bool) {
		if !oldValue && newValue {
			cm.PlayChime()
		}
	})

	// Register callback for volume changes
	cm.stateManager.RegisterChimeVolumeCallback(func(oldValue, newValue float64) {
		cm.volume = newValue
		if cm.audioPlayer != nil {
			cm.audioPlayer.SetVolume(newValue)
		}
	})
}

// GetAvailableSounds returns a list of user-friendly sound names
func (cm *ChimeManager) GetAvailableSounds() []string {
	if cm.audioPlayer == nil {
		// Return default list if audio player is not available
		return []string{"Ready 1", "Ready 2", "Ready 3", "Ready 4", "Ready 5"}
	}

	fileNames := cm.audioPlayer.GetAvailableSounds()
	displayNames := make([]string, 0, len(fileNames))

	for _, fileName := range fileNames {
		displayNames = append(displayNames, cm.GetSoundDisplayName(fileName))
	}

	return displayNames
}

// GetSoundFileName converts a display name to its corresponding file name
func (cm *ChimeManager) GetSoundFileName(displayName string) string {
	// Handle conversion from display names to file names
	// For example: "Ready 1" -> "ready1.mp3"
	switch displayName {
	case "Ready 1":
		return "ready1.mp3"
	case "Ready 2":
		return "ready2.mp3"
	case "Ready 3":
		return "ready3.mp3"
	case "Ready 4":
		return "ready4.mp3"
	case "Ready 5":
		return "ready5.mp3"
	default:
		// Default to the first sound if the display name is not recognized
		return "ready1.mp3"
	}
}

// GetSoundDisplayName converts a file name to its display name
func (cm *ChimeManager) GetSoundDisplayName(fileName string) string {
	// Handle conversion from file names to display names
	// For example: "ready1.mp3" -> "Ready 1"
	fileName = strings.TrimSuffix(fileName, ".mp3")

	switch fileName {
	case "ready1":
		return "Ready 1"
	case "ready2":
		return "Ready 2"
	case "ready3":
		return "Ready 3"
	case "ready4":
		return "Ready 4"
	case "ready5":
		return "Ready 5"
	default:
		// If filename doesn't match any known pattern, use a generic format
		return "Sound: " + fileName
	}
}

// PlayChime plays the chime sound if enabled
func (cm *ChimeManager) PlayChime() {
	if cm.volume <= 0 || cm.audioPlayer == nil {
		return
	}

	// Convert the display name from state manager to file name for audio player
	displayName := cm.stateManager.GetChimeSound()
	fileName := cm.GetSoundFileName(displayName)
	cm.audioPlayer.PlaySound(fileName)
}

// PlaySound plays a specific sound by its display name
func (cm *ChimeManager) PlaySound(displayName string) {
	if cm.volume <= 0 || cm.audioPlayer == nil {
		return
	}

	// Convert the display name to file name for audio player
	fileName := cm.GetSoundFileName(displayName)
	cm.audioPlayer.PlaySound(fileName)
}
