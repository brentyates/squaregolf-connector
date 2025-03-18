package screens

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/brentyates/squaregolf-connector/internal/core"
)

type SettingsScreen struct {
	window       fyne.Window
	stateManager *core.StateManager
	content      fyne.CanvasObject
	preferences  fyne.Preferences
	spinMode     binding.String
	deviceName   binding.String
	autoConnect  binding.Bool
}

func NewSettingsScreen(w fyne.Window, stateManager *core.StateManager) *SettingsScreen {
	return &SettingsScreen{
		window:       w,
		stateManager: stateManager,
		preferences:  fyne.CurrentApp().Preferences(),
		spinMode:     binding.NewString(),
		deviceName:   binding.NewString(),
		autoConnect:  binding.NewBool(),
	}
}

func (ss *SettingsScreen) Initialize() {
	// Initialize device name binding with saved value
	savedDeviceName := ss.preferences.String("device_name")
	ss.deviceName.Set(savedDeviceName)

	// Initialize auto-connect binding with saved value (default to true)
	ss.autoConnect.Set(ss.preferences.BoolWithFallback("auto_connect", true))

	// Initialize spin mode binding with saved value
	savedMode := ss.preferences.String("spin_mode")
	if savedMode == "" {
		savedMode = "Advanced"
	}
	ss.spinMode.Set(savedMode)

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

	// Create spin mode radio group with data binding
	spinModeRadio := widget.NewRadioGroup([]string{"Standard", "Advanced"}, func(value string) {
		ss.spinMode.Set(value)
		ss.preferences.SetString("spin_mode", value)
	})

	// Set initial selection
	if value, err := ss.spinMode.Get(); err == nil {
		spinModeRadio.SetSelected(value)
	} else {
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

	// Create the main content
	ss.content = container.NewVBox(
		widget.NewLabel("Settings"),
		widget.NewSeparator(),
		widget.NewLabel("Device Name:"),
		container.NewBorder(nil, nil, nil, forgetBtn, deviceNameEntry),
		widget.NewSeparator(),
		autoConnectCheck,
		widget.NewSeparator(),
		widget.NewLabel("Spin Detection Mode:"),
		spinModeRadio,
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
