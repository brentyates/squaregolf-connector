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

type Dashboard struct {
	window            fyne.Window
	stateManager      *core.StateManager
	bluetoothManager  *core.BluetoothManager
	launchMonitor     *core.LaunchMonitor
	connectionStatus  *components.ConnectionStatus
	statusCard        *components.StatusCard
	metricsCard       *components.MetricsCard
	config            AppConfig
	content           fyne.CanvasObject
	connectEnabled    binding.Bool
	disconnectEnabled binding.Bool
	statusText        binding.String
}

type AppConfig struct {
	DeviceName  string
	EnableGSPro bool
	GSProIP     string
	GSProPort   int
}

func NewDashboard(window fyne.Window, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor, config AppConfig) *Dashboard {
	return &Dashboard{
		window:            window,
		stateManager:      stateManager,
		bluetoothManager:  bluetoothManager,
		launchMonitor:     launchMonitor,
		config:            config,
		connectEnabled:    binding.NewBool(),
		disconnectEnabled: binding.NewBool(),
		statusText:        binding.NewString(),
	}
}

func (d *Dashboard) Initialize() {
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

	// Create status label with data binding
	status := widget.NewLabelWithData(d.statusText)
	status.Alignment = fyne.TextAlignCenter
	status.TextStyle = fyne.TextStyle{Bold: true}

	// Create connection controls with bound enabled state
	connectBtn := widget.NewButton("Connect to Device", func() {
		if d.config.DeviceName == "" {
			dialog.ShowError(fmt.Errorf("no device selected"), d.window)
			return
		}
		go d.bluetoothManager.StartBluetoothConnection(d.config.DeviceName, "")
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
			d.connectEnabled.Set(false)
			d.disconnectEnabled.Set(true)
		case core.ConnectionStatusConnecting:
			d.statusText.Set("Connecting...")
			d.connectEnabled.Set(false)
			d.disconnectEnabled.Set(false)
		case core.ConnectionStatusDisconnected:
			d.statusText.Set("Disconnected")
			d.connectEnabled.Set(true)
			d.disconnectEnabled.Set(false)
		case core.ConnectionStatusError:
			if err := d.stateManager.GetLastError(); err != nil {
				d.statusText.Set(fmt.Sprintf("Error: %v", err))
			} else {
				d.statusText.Set("Connection Error")
			}
			d.connectEnabled.Set(true)
			d.disconnectEnabled.Set(false)
		}
	})

	// Create main content area with a grid layout
	content := container.NewGridWithColumns(2,
		d.statusCard.GetCard(),
		d.metricsCard.GetCard(),
	)

	// Create the main content with connection controls at the top
	d.content = container.NewVBox(
		widget.NewLabel("Device Connection"),
		widget.NewSeparator(),
		status,
		widget.NewSeparator(),
		container.NewHBox(connectBtn, disconnectBtn),
		widget.NewSeparator(),
		content,
	)

	gspro := NewGSProScreen(d.window, d.stateManager, d.launchMonitor, AppConfig{
		DeviceName:  d.config.DeviceName,
		EnableGSPro: d.config.EnableGSPro,
		GSProIP:     d.config.GSProIP,
		GSProPort:   d.config.GSProPort,
	})
	gspro.Initialize()

	// Start initial scan if no device is selected
	if d.config.DeviceName == "" {
		// Try to get the last used device from preferences
		lastDevice := fyne.CurrentApp().Preferences().String("device_name")
		if lastDevice != "" {
			d.config.DeviceName = lastDevice
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

func (d *Dashboard) GetContent() fyne.CanvasObject {
	return d.content
}

func (d *Dashboard) OnShow() {
	// Refresh any necessary components
	d.window.Canvas().Refresh(d.content)
}

func (d *Dashboard) OnHide() {
	// Clean up any resources if needed
}
