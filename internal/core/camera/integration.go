package camera

import (
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
	stateManager *core.StateManager
	baseURL      string
	enabled      bool
	httpClient   *http.Client
	mu           sync.Mutex
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

// ShotDetected sends the shot-detected command to the camera (fire and forget)
func (m *Manager) ShotDetected() error {
	m.mu.Lock()
	baseURL := m.baseURL
	enabled := m.enabled
	m.mu.Unlock()

	if !enabled {
		log.Println("Camera integration disabled, skipping shot-detected command")
		return nil // Silent failure
	}

	url := fmt.Sprintf("%s/api/lm/shot-detected", baseURL)
	resp, err := m.httpClient.Post(url, "application/json", nil)
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

	log.Println("Camera shot-detected command sent successfully")
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
