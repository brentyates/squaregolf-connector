package camera

import (
	"log"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

// registerStateListeners registers callbacks for state changes
func (m *Manager) registerStateListeners() {
	m.stateManager.RegisterBallReadyCallback(m.onBallReadyChanged)
	m.stateManager.RegisterLastBallMetricsCallback(m.onLastBallMetricsChanged)
	m.stateManager.RegisterLastClubMetricsCallback(m.onLastClubMetricsChanged)
}

// onBallReadyChanged handles ball ready state changed event from state manager
// When the ball becomes ready (detected and positioned), arm the camera
func (m *Manager) onBallReadyChanged(oldValue, newValue bool) {
	// Only act when ball transitions from not ready to ready
	if oldValue == newValue {
		return
	}

	// Check if camera is enabled
	m.mu.Lock()
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled {
		return
	}

	// When ball becomes ready, arm the camera to start recording
	if newValue {
		log.Println("Ball ready detected, arming camera")
		go m.Arm() // Run in goroutine to avoid blocking
	} else {
		// When ball is no longer ready, cancel any armed recording
		log.Println("Ball no longer ready, canceling camera")
		go m.Cancel() // Run in goroutine to avoid blocking
	}
}

// onLastBallMetricsChanged handles last ball metrics changed event from state manager
// When shot metrics are received, trigger shot-detected to save the recording
func (m *Manager) onLastBallMetricsChanged(oldValue, newValue *core.BallMetrics) {
	// Only act when metrics actually change
	if oldValue == newValue {
		return
	}

	// Ignore nil metrics
	if newValue == nil {
		return
	}

	// Check if camera is enabled
	m.mu.Lock()
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled {
		return
	}

	// Get club metrics (may be nil, which is fine)
	clubMetrics := m.stateManager.GetLastClubMetrics()

	// New shot detected, tell camera to stop recording and save the clip with metrics
	log.Printf("Shot metrics received (ball speed: %.1f m/s), triggering camera shot-detected with metadata", newValue.BallSpeedMPS)
	go m.ShotDetected(newValue, clubMetrics) // Run in goroutine to avoid blocking
}

// onLastClubMetricsChanged handles club metrics changed event from state manager
// When club metrics are received, update the pending recording's metadata
func (m *Manager) onLastClubMetricsChanged(oldValue, newValue *core.ClubMetrics) {
	// Only act when metrics actually change
	if oldValue == newValue {
		return
	}

	// Ignore nil metrics
	if newValue == nil {
		return
	}

	// Check if camera is enabled
	m.mu.Lock()
	enabled := m.enabled
	pendingFilename := m.pendingFilename
	m.mu.Unlock()

	if !enabled {
		return
	}

	// Only update if we have a pending filename from a recent shot-detected
	if pendingFilename == "" {
		log.Println("Club metrics received but no pending filename for metadata update")
		return
	}

	// Send PATCH request to update metadata with club data
	log.Printf("Club metrics received, updating metadata for %s", pendingFilename)
	go m.UpdateMetadata(pendingFilename, newValue) // Run in goroutine to avoid blocking
}
