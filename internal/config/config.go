package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/brentyates/squaregolf-connector/internal/core"
)

// Settings represents all persisted application settings
type Settings struct {
	DeviceName       string `json:"deviceName"`
	AutoConnect      bool   `json:"autoConnect"`
	SpinMode         string `json:"spinMode"`
	GSProIP          string `json:"gsproIP"`
	GSProPort        int    `json:"gsproPort"`
	GSProAutoConnect bool   `json:"gsproAutoConnect"`
	CameraURL        string `json:"cameraURL"`
	CameraEnabled    bool   `json:"cameraEnabled"`
}

// Manager handles loading and saving configuration
type Manager struct {
	settings     Settings
	configPath   string
	mu           sync.RWMutex
	saveCallback func() // Called when settings are saved
}

var (
	instance *Manager
	once     sync.Once
)

// GetInstance returns the singleton config manager instance
func GetInstance() *Manager {
	once.Do(func() {
		instance = &Manager{}
		instance.initialize()
	})
	return instance
}

// initialize sets up the config manager with default values
func (m *Manager) initialize() {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Create config directory in user's home
	configDir := filepath.Join(homeDir, ".squaregolf-connector")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		configDir = "."
	}

	m.configPath = filepath.Join(configDir, "config.json")

	// Set default settings
	m.settings = Settings{
		DeviceName:       "",
		AutoConnect:      false,
		SpinMode:         "advanced",
		GSProIP:          "127.0.0.1",
		GSProPort:        921,
		GSProAutoConnect: false,
		CameraURL:        "http://localhost:5000",
		CameraEnabled:    false,
	}

	// Try to load existing settings
	m.Load()
}

// Load reads settings from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, use defaults
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m.settings)
}

// Save writes settings to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	data, err := json.MarshalIndent(m.settings, "", "  ")
	m.mu.RUnlock()

	if err != nil {
		return err
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return err
	}

	// Call save callback if set
	if m.saveCallback != nil {
		m.saveCallback()
	}

	return nil
}

// GetSettings returns a copy of the current settings
func (m *Manager) GetSettings() Settings {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.settings
}

// UpdateSettings updates the settings and saves to disk
func (m *Manager) UpdateSettings(settings Settings) error {
	m.mu.Lock()
	m.settings = settings
	m.mu.Unlock()

	return m.Save()
}

// Update specific settings fields

func (m *Manager) SetDeviceName(name string) error {
	m.mu.Lock()
	m.settings.DeviceName = name
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetAutoConnect(autoConnect bool) error {
	m.mu.Lock()
	m.settings.AutoConnect = autoConnect
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetSpinMode(spinMode string) error {
	m.mu.Lock()
	m.settings.SpinMode = spinMode
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetGSProIP(ip string) error {
	m.mu.Lock()
	m.settings.GSProIP = ip
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetGSProPort(port int) error {
	m.mu.Lock()
	m.settings.GSProPort = port
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetGSProAutoConnect(autoConnect bool) error {
	m.mu.Lock()
	m.settings.GSProAutoConnect = autoConnect
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetCameraURL(url string) error {
	m.mu.Lock()
	m.settings.CameraURL = url
	m.mu.Unlock()
	return m.Save()
}

func (m *Manager) SetCameraEnabled(enabled bool) error {
	m.mu.Lock()
	m.settings.CameraEnabled = enabled
	m.mu.Unlock()
	return m.Save()
}

// ApplyToStateManager applies the configuration to the state manager
func (m *Manager) ApplyToStateManager(stateManager *core.StateManager) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Apply spin mode
	var spinMode core.SpinMode
	if m.settings.SpinMode == "standard" {
		spinMode = core.Standard
	} else {
		spinMode = core.Advanced
	}
	stateManager.SetSpinMode(&spinMode)

	// Apply camera settings
	stateManager.SetCameraURL(&m.settings.CameraURL)
	stateManager.SetCameraEnabled(m.settings.CameraEnabled)
}
