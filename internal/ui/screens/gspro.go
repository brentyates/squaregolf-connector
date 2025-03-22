package screens

import (
	"fmt"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/brentyates/squaregolf-connector/internal/core"
)

type GSProScreen struct {
	window           fyne.Window
	stateManager     *core.StateManager
	content          fyne.CanvasObject
	gsproIP          string
	gsproPort        int
	gsproIntegration *core.GSProIntegration
	launchMonitor    *core.LaunchMonitor
	bluetoothManager *core.BluetoothManager
	preferences      fyne.Preferences
	ipBinding        binding.String
	portBinding      binding.String
	autoConnect      binding.Bool
}

func NewGSProScreen(window fyne.Window, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, gsproIP string, gsproPort int) *GSProScreen {
	// Get the singleton launch monitor instance
	launchMonitor := core.GetLaunchMonitorInstance(stateManager, bluetoothManager)

	return &GSProScreen{
		window:           window,
		stateManager:     stateManager,
		bluetoothManager: bluetoothManager,
		launchMonitor:    launchMonitor,
		gsproIP:          gsproIP,
		gsproPort:        gsproPort,
		preferences:      fyne.CurrentApp().Preferences(),
		ipBinding:        binding.NewString(),
		portBinding:      binding.NewString(),
		autoConnect:      binding.NewBool(),
		gsproIntegration: core.GetGSProInstance(stateManager, launchMonitor, gsproIP, gsproPort),
	}
}

func (s *GSProScreen) Initialize() {
	// Create data bindings
	statusText := binding.NewString()
	errorText := binding.NewString()
	connectEnabled := binding.NewBool()
	disconnectEnabled := binding.NewBool()

	// Set initial values
	statusText.Set("Disconnected")
	errorText.Set("")
	connectEnabled.Set(true)
	disconnectEnabled.Set(false)

	// Create connection status and error labels
	status := widget.NewLabelWithData(statusText)
	status.Alignment = fyne.TextAlignCenter
	status.TextStyle = fyne.TextStyle{Bold: true}

	errorLabel := widget.NewLabelWithData(errorText)
	errorLabel.Alignment = fyne.TextAlignCenter
	errorLabel.TextStyle = fyne.TextStyle{Bold: true}
	errorLabel.Hide()

	// Initialize IP and port bindings with saved values
	savedIP := s.preferences.String("gspro_ip")
	if savedIP == "" {
		savedIP = s.gsproIP
	}
	s.ipBinding.Set(savedIP)

	savedPort := fmt.Sprintf("%d", s.preferences.Int("gspro_port"))
	if savedPort == "0" {
		savedPort = fmt.Sprintf("%d", s.gsproPort)
	}
	s.portBinding.Set(savedPort)

	// Initialize auto-connect binding with saved value (default to false)
	s.autoConnect.Set(s.preferences.BoolWithFallback("gspro_auto_connect", false))

	// Create IP and port entries
	ipEntry := widget.NewEntryWithData(s.ipBinding)
	ipEntry.SetPlaceHolder("GSPro IP Address")

	portEntry := widget.NewEntryWithData(s.portBinding)
	portEntry.SetPlaceHolder("GSPro Port")

	// Save settings when values change
	s.ipBinding.AddListener(binding.NewDataListener(func() {
		value, _ := s.ipBinding.Get()
		s.preferences.SetString("gspro_ip", value)
	}))

	s.portBinding.AddListener(binding.NewDataListener(func() {
		value, _ := s.portBinding.Get()
		if port, err := strconv.Atoi(value); err == nil {
			s.preferences.SetInt("gspro_port", port)
		}
	}))

	// Create auto-connect checkbox with data binding
	autoConnectCheck := widget.NewCheckWithData("Connect automatically on start", s.autoConnect)

	// Save auto-connect when changed
	s.autoConnect.AddListener(binding.NewDataListener(func() {
		value, _ := s.autoConnect.Get()
		s.preferences.SetBool("gspro_auto_connect", value)
	}))

	// Create connection controls with bound enabled state
	connectBtn := widget.NewButton("Connect to GSPro", func() {
		// Get IP and port from bindings
		ip, _ := s.ipBinding.Get()
		portStr, _ := s.portBinding.Get()
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("Invalid port number: %v", err)
			return
		}

		// Start the integration and connect in a goroutine
		go func() {
			s.gsproIntegration.Start()
			s.gsproIntegration.Connect(ip, port)
		}()
	})
	connectBtn.Importance = widget.HighImportance
	connectBtn.Disable()
	go connectEnabled.AddListener(binding.NewDataListener(func() {
		enabled, _ := connectEnabled.Get()
		if enabled {
			connectBtn.Enable()
		} else {
			connectBtn.Disable()
		}
	}))

	disconnectBtn := widget.NewButton("Disconnect from GSPro", func() {
		if s.gsproIntegration != nil {
			go func() {
				s.gsproIntegration.Stop()
			}()
		}
	})
	disconnectBtn.Importance = widget.MediumImportance
	disconnectBtn.Disable()
	go disconnectEnabled.AddListener(binding.NewDataListener(func() {
		enabled, _ := disconnectEnabled.Get()
		if enabled {
			disconnectBtn.Enable()
		} else {
			disconnectBtn.Disable()
		}
	}))

	// Register callbacks for GSPro status and error changes
	s.stateManager.RegisterGSProStatusCallback(func(oldValue, newValue core.GSProConnectionStatus) {
		switch newValue {
		case core.GSProStatusConnected:
			statusText.Set("Connected")
			errorText.Set("")
			errorLabel.Hide()
			connectEnabled.Set(false)
			disconnectEnabled.Set(true)
			ipEntry.Disable()
			portEntry.Disable()
		case core.GSProStatusConnecting:
			statusText.Set("Connecting...")
			connectEnabled.Set(false)
			disconnectEnabled.Set(false)
			ipEntry.Disable()
			portEntry.Disable()
		case core.GSProStatusDisconnected:
			statusText.Set("Disconnected")
			errorText.Set("")
			errorLabel.Hide()
			// Only enable connect if there's no active integration
			connectEnabled.Set(true)
			disconnectEnabled.Set(false)
			ipEntry.Enable()
			portEntry.Enable()
		case core.GSProStatusError:
			statusText.Set("Connecting (retrying)...")
			if err := s.stateManager.GetGSProError(); err != nil {
				errorText.Set(fmt.Sprintf("Error: %v", err))
				errorLabel.Show()
			}
			connectEnabled.Set(true)
			disconnectEnabled.Set(false)
			ipEntry.Enable()
			portEntry.Enable()
		}
	})

	// Create the main content
	s.content = container.NewVBox(
		widget.NewLabel("GSPro Connection"),
		widget.NewSeparator(),
		ipEntry,
		widget.NewSeparator(),
		portEntry,
		widget.NewSeparator(),
		autoConnectCheck,
		widget.NewSeparator(),
		status,
		errorLabel,
		widget.NewSeparator(),
		container.NewHBox(connectBtn, disconnectBtn),
		widget.NewSeparator(),
		widget.NewAccordion(
			widget.NewAccordionItem("Troubleshooting", container.NewVBox(
				widget.NewLabel("If the launch monitor will not go into ball detection mode:"),
				widget.NewLabel("1. Try changing the club in GSPro"),
				widget.NewLabel("2. If still not working, restart GSPconnect:"),
				widget.NewLabel("   • Go to Settings -> System -> Reset GSPro Connect"),
				widget.NewLabel("   • This app will automatically reconnect"),
			)),
		),
	)

	// Check if auto-connect is enabled and attempt to connect
	if s.preferences.BoolWithFallback("gspro_auto_connect", false) {
		// Get IP and port from bindings
		ip, _ := s.ipBinding.Get()
		portStr, _ := s.portBinding.Get()
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("Invalid port number for auto-connect: %v", err)
			return
		}

		// Start the integration and connect in a goroutine
		go func() {
			s.gsproIntegration.Start()
			s.gsproIntegration.Connect(ip, port)
		}()
	}
}

func (s *GSProScreen) GetContent() fyne.CanvasObject {
	return s.content
}

func (s *GSProScreen) OnShow() {
	// No need to manually update UI elements - they will be updated through data binding
}

func (s *GSProScreen) OnHide() {
	// No cleanup needed - we want GSPro connection to persist while navigating
}
