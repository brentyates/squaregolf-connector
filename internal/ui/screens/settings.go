package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/brentyates/squaregolf-connector/internal/core"
)

type SettingsScreen struct {
	window       fyne.Window
	stateManager *core.StateManager
	chimeManager *core.ChimeManager
	content      fyne.CanvasObject
	preferences  fyne.Preferences
	deviceName   binding.String
	autoConnect  binding.Bool
	chimeSound   binding.String
}

func NewSettingsScreen(w fyne.Window, stateManager *core.StateManager, chimeManager *core.ChimeManager) *SettingsScreen {
	return &SettingsScreen{
		window:       w,
		stateManager: stateManager,
		chimeManager: chimeManager,
		preferences:  fyne.CurrentApp().Preferences(),
		deviceName:   binding.NewString(),
		autoConnect:  binding.NewBool(),
		chimeSound:   binding.NewString(),
	}
}

func (ss *SettingsScreen) Initialize() {
	// Initialize device name binding with saved value
	savedDeviceName := ss.preferences.String("device_name")
	ss.deviceName.Set(savedDeviceName)

	// Initialize auto-connect binding with saved value (default to true)
	ss.autoConnect.Set(ss.preferences.BoolWithFallback("auto_connect", true))

	// Create device name entry with data binding
	deviceNameEntry := widget.NewEntryWithData(ss.deviceName)
	deviceNameEntry.SetPlaceHolder("Device Name")

	// Create forget button
	forgetBtn := widget.NewButton("Forget", func() {
		ss.deviceName.Set("")
		ss.preferences.SetString("device_name", "")
	})

	// Save device name when changed
	ss.deviceName.AddListener(binding.NewDataListener(func() {
		value, _ := ss.deviceName.Get()
		ss.preferences.SetString("device_name", value)
	}))

	// Create auto-connect checkbox with data binding
	autoConnectCheck := widget.NewCheckWithData("Connect automatically on start", ss.autoConnect)

	// Save auto-connect when changed
	ss.autoConnect.AddListener(binding.NewDataListener(func() {
		value, _ := ss.autoConnect.Get()
		ss.preferences.SetBool("auto_connect", value)
	}))

	// Create spin mode radio group
	spinModeRadio := widget.NewRadioGroup([]string{"Standard", "Advanced"}, func(value string) {
		var spinMode core.SpinMode
		if value == "Standard" {
			spinMode = core.Standard
		} else {
			spinMode = core.Advanced
		}
		ss.stateManager.SetSpinMode(&spinMode)
	})

	// Set initial selection based on state
	if spinMode := ss.stateManager.GetSpinMode(); spinMode != nil {
		if *spinMode == core.Standard {
			spinModeRadio.SetSelected("Standard")
		} else {
			spinModeRadio.SetSelected("Advanced")
		}
	} else {
		// Default to Advanced if not set
		defaultSpinMode := core.Advanced
		ss.stateManager.SetSpinMode(&defaultSpinMode)
		spinModeRadio.SetSelected("Advanced")
	}

	// Register callback for device name updates when connecting
	ss.stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue core.ConnectionStatus) {
		if newValue == core.ConnectionStatusConnected {
			if deviceName := ss.stateManager.GetDeviceDisplayName(); deviceName != nil {
				ss.deviceName.Set(*deviceName)
			}
		}
	})

	// Register callback for device name changes
	ss.stateManager.RegisterDeviceDisplayNameCallback(func(oldValue, newValue *string) {
		if newValue != nil {
			ss.deviceName.Set(*newValue)
		}
	})

	chimeVolumePreference := ss.preferences.FloatWithFallback("chime_volume", ss.stateManager.GetChimeVolume())
	chimeVolume := binding.BindFloat(&chimeVolumePreference)
	ss.stateManager.SetChimeVolume(chimeVolumePreference)
	chimeVolumeSlider := widget.NewSliderWithData(0.0, 1.0, chimeVolume)
	chimeVolumeSlider.Step = 0.05
	chimeVolumeSlider.OnChanged = func(value float64) {
		ss.stateManager.SetChimeVolume(value)
		ss.preferences.SetFloat("chime_volume", value)
	}

	ss.stateManager.RegisterChimeVolumeCallback(func(oldValue, newValue float64) {
		chimeVolume.Set(newValue)
	})
	// Initialize chime sound selection
	// Get the saved chime sound from preferences or use the state manager value
	savedChimeSound := ss.preferences.StringWithFallback("chime_sound", ss.stateManager.GetChimeSound())

	// Check if the saved sound is a file name (old format) and convert to display name if needed
	if ss.chimeManager != nil && savedChimeSound != "" && savedChimeSound[len(savedChimeSound)-4:] == ".mp3" {
		savedChimeSound = ss.chimeManager.GetSoundDisplayName(savedChimeSound)
	}

	ss.chimeSound.Set(savedChimeSound)
	ss.stateManager.SetChimeSound(savedChimeSound)

	// Get available sounds from ChimeManager
	var availableSounds []string
	if ss.chimeManager != nil {
		availableSounds = ss.chimeManager.GetAvailableSounds()
	} else {
		// Fallback if ChimeManager is not available
		availableSounds = []string{"Ready 1", "Ready 2", "Ready 3", "Ready 4", "Ready 5"}
	}

	// Create the sound selection dropdown
	chimeSoundSelect := widget.NewSelect(availableSounds, func(value string) {
		ss.chimeSound.Set(value)
		ss.stateManager.SetChimeSound(value)
		ss.preferences.SetString("chime_sound", value)
	})

	// Create a play button to preview the selected sound
	playButton := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		// Play the currently selected sound
		if ss.chimeManager != nil {
			// Get the currently selected sound display name from the dropdown
			selectedSound := chimeSoundSelect.Selected
			// Play the sound using the ChimeManager
			ss.chimeManager.PlaySound(selectedSound)
		}
	})

	// Create a horizontal container for the dropdown and play button
	soundSelectionContainer := container.NewHBox(chimeSoundSelect, playButton)

	// Set the current selection
	chimeSoundSelect.SetSelected(savedChimeSound)

	// Register callback for chime sound changes from other sources
	ss.stateManager.RegisterChimeSoundCallback(func(oldValue, newValue string) {
		if chimeSoundSelect.Selected != newValue {
			chimeSoundSelect.SetSelected(newValue)
		}
	})

	// Listen for changes in the binding
	ss.chimeSound.AddListener(binding.NewDataListener(func() {
		value, _ := ss.chimeSound.Get()
		if chimeSoundSelect.Selected != value {
			chimeSoundSelect.SetSelected(value)
		}
		ss.preferences.SetString("chime_sound", value)
	}))

	// Create the main content
	ss.content = container.NewVBox(
		widget.NewLabel("Device Name:"),
		container.NewBorder(nil, nil, nil, forgetBtn, deviceNameEntry),
		autoConnectCheck,
		widget.NewSeparator(),
		widget.NewLabel("Spin Detection Mode:"),
		spinModeRadio,
		widget.NewSeparator(),
		widget.NewLabel("Ball Ready Sound:"),
		widget.NewLabel("Sound:"),
		soundSelectionContainer,
		widget.NewLabel("Volume:"),
		chimeVolumeSlider,
	)
}

func (ss *SettingsScreen) Show() {
	ss.window.SetContent(ss.GetContent())
}

func (ss *SettingsScreen) GetContent() fyne.CanvasObject {
	return ss.content
}

func (ss *SettingsScreen) OnShow() {
	// No need to manually update UI elements - they will be updated through data binding
}

func (ss *SettingsScreen) OnHide() {
	// No cleanup needed
}
