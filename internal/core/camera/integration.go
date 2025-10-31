package camera

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

var (
	cameraInstance *Manager
	cameraOnce     sync.Once
)

// Manager handles communication with the swing camera via HTTP REST API
type Manager struct {
	stateManager        *core.StateManager
	baseURL             string
	enabled             bool
	httpClient          *http.Client
	pendingFilename     string              // Stores filename from shot-detected to update with club metrics later
	pendingClubMetrics  *core.ClubMetrics   // Buffers club metrics that arrive before shot-detected response
	mu                  sync.Mutex
}

// GetInstance returns the singleton instance of CameraManager
func GetInstance(stateManager *core.StateManager, baseURL string, enabled bool) *Manager {
	cameraOnce.Do(func() {
		if baseURL == "" {
			baseURL = "http://localhost:5000"
		}

		cameraInstance = &Manager{
			stateManager: stateManager,
			baseURL:      baseURL,
			enabled:      enabled,
			httpClient: &http.Client{
				Timeout: 10 * time.Second,
			},
		}

		// Register state listeners if enabled
		if enabled {
			cameraInstance.registerStateListeners()
			log.Printf("Camera integration initialized with URL: %s", baseURL)
		} else {
			log.Println("Camera integration initialized but disabled")
		}
	})
	return cameraInstance
}

// IsEnabled returns whether the camera integration is enabled
func (m *Manager) IsEnabled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enabled
}

// SetEnabled enables or disables the camera integration
func (m *Manager) SetEnabled(enabled bool) {
	m.mu.Lock()
	wasEnabled := m.enabled
	m.enabled = enabled
	m.mu.Unlock()

	if wasEnabled == enabled {
		return
	}

	// Update state manager
	m.stateManager.SetCameraEnabled(enabled)

	if enabled {
		// Register state listeners when enabling
		m.registerStateListeners()
		log.Println("Camera integration enabled")
	} else {
		log.Println("Camera integration disabled")
	}
}

// SetBaseURL updates the camera base URL
func (m *Manager) SetBaseURL(baseURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if baseURL == "" {
		baseURL = "http://localhost:5000"
	}

	m.baseURL = baseURL
	m.stateManager.SetCameraURL(&baseURL)
	log.Printf("Camera base URL updated to: %s", baseURL)
}

// GetBaseURL returns the current camera base URL
func (m *Manager) GetBaseURL() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.baseURL
}

// Arm sends the arm command to the camera (fire and forget)
func (m *Manager) Arm() error {
	m.mu.Lock()
	baseURL := m.baseURL
	enabled := m.enabled
	// Clear any pending state from previous shot
	m.pendingFilename = ""
	m.pendingClubMetrics = nil
	m.mu.Unlock()

	if !enabled {
		log.Println("Camera integration disabled, skipping arm command")
		return nil // Silent failure as requested
	}

	url := fmt.Sprintf("%s/api/lm/arm", baseURL)
	resp, err := m.httpClient.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Failed to arm camera: %v", err)
		return nil // Silent failure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Camera arm request failed: %d - %s", resp.StatusCode, string(body))
		return nil // Silent failure
	}

	log.Println("Camera arm command sent successfully")
	return nil
}

// ShotDetected sends the shot-detected command to the camera with ball metrics only (fire and forget)
// Club metrics are sent separately via UpdateMetadata() when they arrive
func (m *Manager) ShotDetected(ballMetrics *core.BallMetrics) error {
	m.mu.Lock()
	baseURL := m.baseURL
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled {
		log.Println("Camera integration disabled, skipping shot-detected command")
		return nil // Silent failure
	}

	// Convert ball metrics to SwingCam format (flat structure)
	ballData := convertBallMetrics(ballMetrics)

	// Marshal ball data directly (no wrapper object)
	payloadBytes, err := json.Marshal(ballData)
	if err != nil {
		log.Printf("Failed to marshal ball data for shot-detected: %v", err)
		return nil // Silent failure
	}

	url := fmt.Sprintf("%s/api/lm/shot-detected", baseURL)
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to send shot-detected to camera: %v", err)
		return nil // Silent failure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Camera shot-detected request failed: %d - %s", resp.StatusCode, string(body))
		return nil // Silent failure
	}

	// Parse response to get filename for potential club metrics update later
	var shotResponse ShotResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read shot-detected response: %v", err)
	} else if err := json.Unmarshal(body, &shotResponse); err != nil {
		log.Printf("Failed to parse shot-detected response: %v", err)
	} else if shotResponse.Filename != "" {
		// Store filename and check for buffered club metrics
		m.mu.Lock()
		m.pendingFilename = shotResponse.Filename
		bufferedClubMetrics := m.pendingClubMetrics
		m.pendingClubMetrics = nil // Clear buffer after retrieving
		m.mu.Unlock()

		log.Printf("Camera shot-detected successful, filename: %s", shotResponse.Filename)

		// If club metrics arrived before the filename (race condition), send them now
		if bufferedClubMetrics != nil {
			log.Printf("Applying buffered club metrics to %s", shotResponse.Filename)
			go m.UpdateMetadata(shotResponse.Filename, bufferedClubMetrics)
		}
	}

	// Log success with metrics info
	if ballData != nil {
		log.Printf("Camera shot-detected sent successfully with ball metrics (ball speed: %.1f mph)", ballData.BallSpeed)
	} else {
		log.Println("Camera shot-detected sent successfully (no ball metrics)")
	}
	return nil
}

// Cancel sends the cancel command to the camera (fire and forget)
func (m *Manager) Cancel() error {
	m.mu.Lock()
	baseURL := m.baseURL
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled {
		log.Println("Camera integration disabled, skipping cancel command")
		return nil // Silent failure
	}

	url := fmt.Sprintf("%s/api/lm/cancel", baseURL)
	resp, err := m.httpClient.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("Failed to cancel camera: %v", err)
		return nil // Silent failure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Camera cancel request failed: %d - %s", resp.StatusCode, string(body))
		return nil // Silent failure
	}

	log.Println("Camera cancel command sent successfully")
	return nil
}

// UpdateMetadata sends club metrics to update the metadata of a recorded video (fire and forget)
// Sends club data directly (flat structure) via PATCH /api/recordings/{filename}/metadata
func (m *Manager) UpdateMetadata(filename string, clubMetrics *core.ClubMetrics) error {
	m.mu.Lock()
	baseURL := m.baseURL
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled {
		log.Println("Camera integration disabled, skipping metadata update")
		return nil // Silent failure
	}

	if filename == "" {
		log.Println("No filename available for metadata update")
		return nil
	}

	if clubMetrics == nil {
		log.Println("No club metrics available for metadata update")
		return nil
	}

	// Convert club metrics to SwingCam format (flat structure)
	clubData := convertClubMetrics(clubMetrics)
	if clubData == nil {
		log.Println("Failed to convert club metrics")
		return nil
	}

	// Add club name from state manager (set by GSPro)
	if clubName := m.stateManager.GetClubName(); clubName != nil {
		clubData.ClubType = *clubName
	}

	// Marshal club data directly (no wrapper object)
	payloadBytes, err := json.Marshal(clubData)
	if err != nil {
		log.Printf("Failed to marshal club data for metadata update: %v", err)
		return nil // Silent failure
	}

	url := fmt.Sprintf("%s/api/recordings/%s/metadata", baseURL, filename)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Failed to create metadata update request: %v", err)
		return nil // Silent failure
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to send metadata update to camera: %v", err)
		return nil // Silent failure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Camera metadata update request failed: %d - %s", resp.StatusCode, string(body))
		return nil // Silent failure
	}

	log.Printf("Camera metadata updated successfully for %s with club data", filename)
	return nil
}
