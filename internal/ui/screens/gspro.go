package screens

import (
	"fmt"
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
	config           AppConfig
	gsproIntegration *core.GSProIntegration
	launchMonitor    *core.LaunchMonitor
	preferences      fyne.Preferences
	ipBinding        binding.String
	portBinding      binding.String
}

func NewGSProScreen(window fyne.Window, stateManager *core.StateManager, launchMonitor *core.LaunchMonitor, config AppConfig) *GSProScreen {
	return &GSProScreen{
		window:        window,
		stateManager:  stateManager,
		launchMonitor: launchMonitor,
		config:        config,
		preferences:   fyne.CurrentApp().Preferences(),
		ipBinding:     binding.NewString(),
		portBinding:   binding.NewString(),
	}
}

func (s *GSProScreen) Initialize() {
	// Create data bindings
	statusText := binding.NewString()
	connectEnabled := binding.NewBool()
	disconnectEnabled := binding.NewBool()

	// Set initial values
	statusText.Set("Disconnected")
	connectEnabled.Set(true)
	disconnectEnabled.Set(false)

	// Create connection status
	status := widget.NewLabelWithData(statusText)
	status.Alignment = fyne.TextAlignCenter
	status.TextStyle = fyne.TextStyle{Bold: true}

	// Initialize IP and port bindings with saved values
	savedIP := s.preferences.String("gspro_ip")
	if savedIP == "" {
		savedIP = s.config.GSProIP
	}
	s.ipBinding.Set(savedIP)

	savedPort := fmt.Sprintf("%d", s.preferences.Int("gspro_port"))
	if savedPort == "0" {
		savedPort = fmt.Sprintf("%d", s.config.GSProPort)
	}
	s.portBinding.Set(savedPort)

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

	// Create connection controls with bound enabled state
	connectBtn := widget.NewButton("Connect to GSPro", func() {
		// Parse port number
		port, err := strconv.Atoi(portEntry.Text)
		if err != nil {
			// Create integration if it doesn't exist
			if s.gsproIntegration == nil {
				s.gsproIntegration = core.NewGSProIntegration(s.stateManager, s.launchMonitor, "", 0)
			}
			// Start the integration and let it handle the error state
			go func() {
				s.gsproIntegration.Start()
				s.gsproIntegration.Connect(ipEntry.Text, port)
			}()
			return
		}

		// Create integration if it doesn't exist
		if s.gsproIntegration == nil {
			s.gsproIntegration = core.NewGSProIntegration(s.stateManager, s.launchMonitor, "", 0)
		}

		// Start the integration and connect in a goroutine
		go func() {
			s.gsproIntegration.Start()
			s.gsproIntegration.Connect(ipEntry.Text, port)
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
			// Get a reference to the integration before clearing it
			integration := s.gsproIntegration
			// Clear the integration reference first so status updates see it as nil
			s.gsproIntegration = nil
			// Stop the integration in a goroutine to prevent UI blocking
			go func() {
				integration.Stop()
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
			// Only enable connect if there's no active integration
			connectEnabled.Set(s.gsproIntegration == nil)
			disconnectEnabled.Set(false)
			ipEntry.Enable()
			portEntry.Enable()
		case core.GSProStatusError:
			if err := s.stateManager.GetGSProError(); err != nil {
				statusText.Set(fmt.Sprintf("Error: %v", err))
			} else {
				statusText.Set("Connection Error")
			}
			// Only enable connect if there's no active integration
			connectEnabled.Set(s.gsproIntegration == nil)
			disconnectEnabled.Set(s.gsproIntegration != nil)
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
		status,
		widget.NewSeparator(),
		container.NewHBox(connectBtn, disconnectBtn),
	)
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
