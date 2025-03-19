package core

import (
	"github.com/brentyates/squaregolf-connector/internal/core/audio"
)

// ChimeManager handles the chime functionality
type ChimeManager struct {
	audioPlayer  *audio.AudioPlayer
	stateManager *StateManager
	volume       float64
}

// NewChimeManager creates a new chime manager
func NewChimeManager(stateManager *StateManager) *ChimeManager {
	audioPlayer, err := audio.NewAudioPlayer()
	if err != nil {
		// Log the error but continue without audio
		return &ChimeManager{
			stateManager: stateManager,
			volume:       stateManager.GetChimeVolume(),
		}
	}

	cm := &ChimeManager{
		audioPlayer:  audioPlayer,
		stateManager: stateManager,
		volume:       stateManager.GetChimeVolume(),
	}

	return cm
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

// PlayChime plays the chime sound if enabled
func (cm *ChimeManager) PlayChime() {
	if cm.volume <= 0 || cm.audioPlayer == nil {
		return
	}

	cm.audioPlayer.PlayChime()
}
