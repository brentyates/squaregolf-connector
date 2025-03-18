package core

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// Preferences represents application preferences
type Preferences struct {
	LastDeviceName string   `json:"last_device_name"`
	SpinMode       SpinMode `json:"spin_mode"`
}

var (
	preferencesInstance *Preferences
	preferencesOnce     sync.Once
)

// GetPreferences returns the singleton instance of Preferences
func GetPreferences() *Preferences {
	preferencesOnce.Do(func() {
		preferencesInstance = &Preferences{}
		preferencesInstance.load()
	})
	return preferencesInstance
}

// load loads preferences from disk
func (p *Preferences) load() {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get home directory: %v", err)
		return
	}

	// Create the preferences directory if it doesn't exist
	prefsDir := filepath.Join(homeDir, ".square")
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		log.Printf("Failed to create preferences directory: %v", err)
		return
	}

	// Read the preferences file
	prefsFile := filepath.Join(prefsDir, "preferences.json")
	data, err := os.ReadFile(prefsFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Failed to read preferences file: %v", err)
		}
		return
	}

	// Parse the preferences
	if err := json.Unmarshal(data, p); err != nil {
		log.Printf("Failed to parse preferences: %v", err)
	}
}

// save saves preferences to disk
func (p *Preferences) save() {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Failed to get home directory: %v", err)
		return
	}

	// Create the preferences directory if it doesn't exist
	prefsDir := filepath.Join(homeDir, ".square")
	if err := os.MkdirAll(prefsDir, 0755); err != nil {
		log.Printf("Failed to create preferences directory: %v", err)
		return
	}

	// Marshal the preferences
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal preferences: %v", err)
		return
	}

	// Write the preferences file
	prefsFile := filepath.Join(prefsDir, "preferences.json")
	if err := os.WriteFile(prefsFile, data, 0644); err != nil {
		log.Printf("Failed to write preferences file: %v", err)
	}
}

// SetLastDeviceName sets the last used device name
func (p *Preferences) SetLastDeviceName(deviceName string) {
	p.LastDeviceName = deviceName
	p.save()
}

// GetLastDeviceName gets the last used device name
func (p *Preferences) GetLastDeviceName() string {
	return p.LastDeviceName
}

// SetSpinMode sets the spin detection mode
func (p *Preferences) SetSpinMode(mode SpinMode) {
	p.SpinMode = mode
	p.save()
}

// GetSpinMode gets the spin detection mode
func (p *Preferences) GetSpinMode() SpinMode {
	return p.SpinMode
}
