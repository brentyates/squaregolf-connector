package screens

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fyne.io/fyne/v2/dialog"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/ui/components"
)

type Dashboard struct {
	window           fyne.Window
	stateManager     *core.StateManager
	bluetoothManager *core.BluetoothManager
	launchMonitor    *core.LaunchMonitor
	connectionStatus *components.ConnectionStatus
	statusCard       *components.StatusCard
	metricsCard      *components.MetricsCard
	config           AppConfig
	content          fyne.CanvasObject
}

type AppConfig struct {
	DeviceName  string
	EnableGSPro bool
	GSProIP     string
	GSProPort   int
}

func NewDashboard(window fyne.Window, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor, config AppConfig) *Dashboard {
	return &Dashboard{
		window:           window,
		stateManager:     stateManager,
		bluetoothManager: bluetoothManager,
		launchMonitor:    launchMonitor,
		config:           config,
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

	// Create header with connection controls
	title := widget.NewLabel("Device")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Create device selection dropdown
	deviceSelect := widget.NewSelect([]string{}, func(deviceName string) {
		d.config.DeviceName = deviceName
		fyne.CurrentApp().Preferences().SetString("device_name", deviceName)
	})

	// Create scan button
	scanBtn := widget.NewButton("Scan", func() {
		go func() {
			if err := d.bluetoothManager.StartScan(); err != nil {
				dialog.ShowError(err, d.window)
				return
			}

			// Wait for scan to complete
			time.Sleep(5 * time.Second)
			d.bluetoothManager.StopScan()

			// Update device list
			devices := d.bluetoothManager.GetDiscoveredDevices()
			deviceSelect.Options = devices
			if len(devices) > 0 {
				deviceSelect.SetSelected(devices[0])
			}
		}()
	})
	scanBtn.SetIcon(theme.SearchIcon())

	// Create connection controls with better styling
	connectBtn := widget.NewButton("Connect", func() {
		if d.config.DeviceName == "" {
			dialog.ShowError(fmt.Errorf("no device selected"), d.window)
			return
		}
		go d.bluetoothManager.StartBluetoothConnection(d.config.DeviceName, "")
	})
	connectBtn.Importance = widget.HighImportance
	connectBtn.SetIcon(theme.ConfirmIcon())

	disconnectBtn := widget.NewButton("Disconnect", func() {
		go d.bluetoothManager.DisconnectBluetooth()
	})
	disconnectBtn.Importance = widget.MediumImportance
	disconnectBtn.SetIcon(theme.CancelIcon())

	// Create a container for device selection and connection controls with proper spacing
	connectionControls := container.NewHBox(
		deviceSelect,
		scanBtn,
		connectBtn,
		disconnectBtn,
		widget.NewSeparator(),
		d.connectionStatus.GetContainer(),
	)

	// Create header with title and connection controls
	header := container.NewBorder(nil, nil, title, connectionControls, layout.NewSpacer())

	// Create main content area with a grid layout
	content := container.NewGridWithColumns(2,
		d.statusCard.GetCard(),
		d.metricsCard.GetCard(),
	)

	// Create the main content
	d.content = container.NewVBox(
		header,
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
			deviceSelect.SetSelected(lastDevice)
			go d.bluetoothManager.StartBluetoothConnection(lastDevice, "")
		} else {
			go func() {
				if err := d.bluetoothManager.StartScan(); err != nil {
					dialog.ShowError(err, d.window)
					return
				}

				// Wait for scan to complete
				time.Sleep(5 * time.Second)
				d.bluetoothManager.StopScan()

				// Update device list
				devices := d.bluetoothManager.GetDiscoveredDevices()
				deviceSelect.Options = devices
				if len(devices) > 0 {
					deviceSelect.SetSelected(devices[0])
				}
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
