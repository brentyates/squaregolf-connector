package screens

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2/dialog"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/ui/components"
)

type Device struct {
	window            fyne.Window
	stateManager      *core.StateManager
	bluetoothManager  *core.BluetoothManager
	launchMonitor     *core.LaunchMonitor
	connectionStatus  *components.ConnectionStatus
	statusCard        *components.StatusCard
	metricsCard       *components.MetricsCard
	deviceName        string
	content           fyne.CanvasObject
	connectEnabled    binding.Bool
	disconnectEnabled binding.Bool
	statusText        binding.String
	errorText         binding.String
}

func NewDevice(window fyne.Window, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor, deviceName string) *Device {
	return &Device{
		window:            window,
		stateManager:      stateManager,
		bluetoothManager:  bluetoothManager,
		launchMonitor:     launchMonitor,
		deviceName:        deviceName,
		connectEnabled:    binding.NewBool(),
		disconnectEnabled: binding.NewBool(),
		statusText:        binding.NewString(),
		errorText:         binding.NewString(),
	}
}

func (d *Device) Initialize() {
	// Create components
	d.connectionStatus = components.NewConnectionStatus()
	d.connectionStatus.SetWindow(d.window)
	d.statusCard = components.NewStatusCard()
	d.metricsCard = components.NewMetricsCard()

	// Register callbacks
	d.connectionStatus.RegisterCallback(d.stateManager)
	d.statusCard.RegisterCallbacks(d.stateManager, d.window)
	d.metricsCard.RegisterCallbacks(d.stateManager, d.window)

	// Set initial states
	d.connectEnabled.Set(true)
	d.disconnectEnabled.Set(false)
	d.statusText.Set("Disconnected")
	d.errorText.Set("")

	// Create status label with data binding
	status := widget.NewLabelWithData(d.statusText)
	status.Alignment = fyne.TextAlignCenter
	status.TextStyle = fyne.TextStyle{Bold: true}

	// Create error label with data binding
	errorLabel := widget.NewLabelWithData(d.errorText)
	errorLabel.Alignment = fyne.TextAlignCenter
	errorLabel.TextStyle = fyne.TextStyle{Bold: true}
	errorLabel.Hide()

	// Create connection controls with bound enabled state
	connectBtn := widget.NewButton("Connect to Device", func() {
		go d.bluetoothManager.StartBluetoothConnection(d.deviceName, "")
	})
	connectBtn.Importance = widget.HighImportance
	connectBtn.SetIcon(theme.ConfirmIcon())
	connectBtn.Disable()
	go d.connectEnabled.AddListener(binding.NewDataListener(func() {
		enabled, _ := d.connectEnabled.Get()
		if enabled {
			connectBtn.Enable()
		} else {
			connectBtn.Disable()
		}
	}))

	disconnectBtn := widget.NewButton("Disconnect from Device", func() {
		go d.bluetoothManager.DisconnectBluetooth()
	})
	disconnectBtn.Importance = widget.MediumImportance
	disconnectBtn.SetIcon(theme.CancelIcon())
	disconnectBtn.Disable()
	go d.disconnectEnabled.AddListener(binding.NewDataListener(func() {
		enabled, _ := d.disconnectEnabled.Get()
		if enabled {
			disconnectBtn.Enable()
		} else {
			disconnectBtn.Disable()
		}
	}))

	// Register callback for connection status changes
	d.stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue core.ConnectionStatus) {
		switch newValue {
		case core.ConnectionStatusConnected:
			d.statusText.Set("Connected")
			d.errorText.Set("")
			errorLabel.Hide()
			d.connectEnabled.Set(false)
			d.disconnectEnabled.Set(true)
		case core.ConnectionStatusConnecting:
			d.statusText.Set("Connecting...")
			d.connectEnabled.Set(false)
			d.disconnectEnabled.Set(true)
		case core.ConnectionStatusDisconnected:
			d.statusText.Set("Disconnected")
			d.errorText.Set("")
			errorLabel.Hide()
			d.connectEnabled.Set(true)
			d.disconnectEnabled.Set(false)
		case core.ConnectionStatusError:
			d.statusText.Set("Connecting (retrying)...")
			if err := d.stateManager.GetLastError(); err != nil {
				d.errorText.Set(fmt.Sprintf("Error: %v", err))
				errorLabel.Show()
			}
			d.connectEnabled.Set(true)
			d.disconnectEnabled.Set(true)
		}
	})

	// Register callback for device name changes to update preferences
	d.stateManager.RegisterDeviceDisplayNameCallback(func(oldValue, newValue *string) {
		if newValue != nil {
			fyne.CurrentApp().Preferences().SetString("device_name", *newValue)
		}
	})

	// Create main content area with a vertical layout
	content := container.NewVBox(
		widget.NewLabel("Device Information"),
		widget.NewSeparator(),
		container.NewVBox(
			d.statusCard.GetDeviceLabel(),
			d.statusCard.GetBatteryLabel(),
		),
		widget.NewSeparator(),
		widget.NewLabel("Ball Status"),
		widget.NewSeparator(),
		container.NewVBox(d.statusCard.GetCard().Content.(*fyne.Container).Objects[2].(*fyne.Container).Objects[2:]...),
		widget.NewSeparator(),
		widget.NewLabel("System Status"),
		widget.NewSeparator(),
		container.NewVBox(d.statusCard.GetCard().Content.(*fyne.Container).Objects[4].(*fyne.Container).Objects[2:]...),
	)

	// Create the main content with connection controls at the top
	d.content = container.NewVBox(
		widget.NewLabel("Device Connection"),
		widget.NewSeparator(),
		status,
		errorLabel,
		widget.NewSeparator(),
		container.NewHBox(connectBtn, disconnectBtn),
		widget.NewSeparator(),
		content,
	)

	// Start initial scan if no device is selected
	if d.deviceName == "" {
		// Try to get the last used device from preferences
		lastDevice := fyne.CurrentApp().Preferences().String("device_name")
		if lastDevice != "" {
			d.deviceName = lastDevice
			// Only auto-connect if the setting is enabled
			if fyne.CurrentApp().Preferences().BoolWithFallback("auto_connect", true) {
				go d.bluetoothManager.StartBluetoothConnection(lastDevice, "")
			}
		} else {
			go func() {
				if err := d.bluetoothManager.StartScan(); err != nil {
					dialog.ShowError(err, d.window)
					return
				}

				// Wait for scan to complete
				time.Sleep(5 * time.Second)
				d.bluetoothManager.StopScan()
			}()
		}
	}
}

func (d *Device) GetContent() fyne.CanvasObject {
	return d.content
}

func (d *Device) OnShow() {
	// Refresh any necessary components
	d.window.Canvas().Refresh(d.content)
}

func (d *Device) OnHide() {
	// Clean up any resources if needed
}
