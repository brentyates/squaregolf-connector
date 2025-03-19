package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"

	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/logging"
	"github.com/brentyates/squaregolf-connector/internal/ui/screens"
	"github.com/brentyates/squaregolf-connector/internal/ui/theme"
)

// Application configuration
type AppConfig struct {
	UseMock     core.MockMode
	DeviceName  string
	Headless    bool
	GSProIP     string
	GSProPort   int
	EnableGSPro bool
}

// Initialize the backend services (Bluetooth, state manager, etc.)
func initializeBackend(config AppConfig) (*core.StateManager, *core.BluetoothManager, *core.LaunchMonitor) {
	// Initialize logging
	logging.SetAppName(core.AppName)
	if err := logging.Init(); err != nil {
		os.Exit(1)
	}
	log.Println("Starting Square BT application...")

	// Get the state manager instance
	stateManager := core.GetInstance()

	// Create the appropriate Bluetooth client
	var bleClient core.BluetoothClient
	var err error

	if config.UseMock == core.MockModeStub {
		log.Println("Using mock Bluetooth implementation")
		bleClient = core.NewMockBluetoothClient()
	} else if config.UseMock == core.MockModeSimulate {
		log.Println("Using simulated device implementation")
		simulatorConfig := core.SimulatorConfig{
			BatteryDrainRate: 1,
			ResponseDelay:    100 * time.Millisecond,
		}
		bleClient = core.NewSimulatorBluetoothClient(simulatorConfig)
	} else {
		log.Println("Using real Bluetooth implementation with TinyGo")
		bleClient, err = core.NewTinyGoBluetoothClient()
		if err != nil {
			log.Printf("Failed to initialize Bluetooth: %v", err)
			// Exit the application if Bluetooth initialization fails
			os.Exit(1)
		}
	}

	// Create a bluetooth manager with just the state manager
	bluetoothManager := core.NewBluetoothManager(stateManager)

	// Set the bluetooth client on the bluetooth manager
	bluetoothManager.SetClient(bleClient)

	// Create a launch monitor instance
	launchMonitor := core.NewLaunchMonitor(stateManager, bleClient)

	// Set up launch monitor to handle notifications from the bluetooth manager
	launchMonitor.SetupNotifications(bluetoothManager)

	return stateManager, bluetoothManager, launchMonitor
}

// setupHeadlessCallbacks configures callbacks for headless mode
func setupHeadlessCallbacks(stateManager *core.StateManager) {
	stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue core.ConnectionStatus) {
		log.Printf("Connection status changed from %v to %v", oldValue, newValue)
	})

	stateManager.RegisterLastBallMetricsCallback(func(oldValue, newValue *core.BallMetrics) {
		if newValue != nil {
			log.Printf("New ball metrics received: %v", newValue)
		}
	})

	stateManager.RegisterLastClubMetricsCallback(func(oldValue, newValue *core.ClubMetrics) {
		if newValue != nil {
			log.Printf("New club metrics received: %v", newValue)
		}
	})

	stateManager.RegisterBatteryLevelCallback(func(oldValue, newValue *int) {
		if newValue != nil {
			log.Printf("Battery level: %d%%", *newValue)
		}
	})

	stateManager.RegisterDeviceDisplayNameCallback(func(oldValue, newValue *string) {
		if newValue != nil {
			log.Printf("Device name: %s", *newValue)
		}
	})

	stateManager.RegisterClubCallback(func(oldValue, newValue *core.ClubType) {
		if newValue != nil {
			log.Printf("Club changed to: %s", newValue.RegularCode)
		}
	})

	stateManager.RegisterHandednessCallback(func(oldValue, newValue *core.HandednessType) {
		if newValue != nil {
			handedness := "Right"
			if *newValue == core.LeftHanded {
				handedness = "Left"
			}
			log.Printf("Handedness: %s", handedness)
		}
	})

	stateManager.RegisterBallDetectedCallback(func(oldValue, newValue bool) {
		log.Printf("Ball detected: %v", newValue)
	})

	stateManager.RegisterBallReadyCallback(func(oldValue, newValue bool) {
		log.Printf("Ball ready: %v", newValue)
	})

	stateManager.RegisterBallPositionCallback(func(oldValue, newValue *core.BallPosition) {
		if newValue != nil {
			log.Printf("Ball position: X=%d, Y=%d, Z=%d", newValue.X, newValue.Y, newValue.Z)
		}
	})

	stateManager.RegisterLastErrorCallback(func(oldValue, newValue error) {
		if newValue != nil {
			log.Printf("Error: %v", newValue)
		}
	})
}

// startUI initializes and runs the graphical user interface
func startUI(config AppConfig, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor) {
	a := app.NewWithID("io.github.byates.squaregolf-connector")
	a.Settings().SetTheme(theme.NewSquareTheme())
	w := a.NewWindow(core.WindowTitle)
	w.Resize(fyne.NewSize(800, 600))

	// Create system menu
	systemMenu := screens.NewSystemMenu(w)
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Submit Bug Report", systemMenu.ShowBugReport),
			fyne.NewMenuItem("Open Log Directory", systemMenu.OpenLogDirectory),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("About", systemMenu.ShowAbout),
		),
	)
	w.SetMainMenu(mainMenu)

	// Create navigation manager
	navManager := screens.NewNavigationManager(w)

	// Create and initialize the chime manager
	chimeManager := core.NewChimeManager(stateManager)
	chimeManager.Initialize()

	// Create and initialize screens
	device := screens.NewDevice(w, stateManager, bluetoothManager, launchMonitor, screens.AppConfig{
		DeviceName:  config.DeviceName,
		EnableGSPro: config.EnableGSPro,
		GSProIP:     config.GSProIP,
		GSProPort:   config.GSProPort,
	})
	device.Initialize()

	alignment := screens.NewAlignmentScreen(w, stateManager)
	alignment.Initialize()

	gspro := screens.NewGSProScreen(w, stateManager, launchMonitor, screens.AppConfig{
		DeviceName:  config.DeviceName,
		EnableGSPro: config.EnableGSPro,
		GSProIP:     config.GSProIP,
		GSProPort:   config.GSProPort,
	})
	gspro.Initialize()

	rangeScreen := screens.NewRangeScreen(w, stateManager)
	rangeScreen.Initialize()

	settings := screens.NewSettingsScreen(w, stateManager)
	settings.Initialize()

	// Add screens to navigation manager
	navManager.AddScreen("device", device)
	navManager.AddScreen("alignment", alignment)
	navManager.AddScreen("gspro", gspro)
	navManager.AddScreen("range", rangeScreen)
	navManager.AddScreen("settings", settings)

	// Update navigation sidebar
	navManager.UpdateSidebar()

	// Create main layout
	mainContent := container.NewBorder(
		nil,
		nil, // Removed footer
		navManager.GetSidebar(),
		nil,
		navManager.GetContent(),
	)

	// Set window content
	w.SetContent(mainContent)

	// Handle window close
	w.SetOnClosed(func() {
		bluetoothManager.DisconnectBluetooth()
	})

	// Show device by default
	navManager.ShowScreen("device")

	w.ShowAndRun()
}

// startCLI initializes and runs the command-line interface
func startCLI(config AppConfig, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor) {
	// Setup callbacks for headless mode
	setupHeadlessCallbacks(stateManager)

	// Start bluetooth connection
	log.Println("Starting Bluetooth connection...")
	bluetoothManager.StartBluetoothConnection(config.DeviceName, "")

	// Wait for connection to be established
	log.Println("Waiting for Bluetooth connection...")
	connectionTimeout := time.After(10 * time.Second)
	connectionEstablished := make(chan struct{})

	// Register a one-time callback for successful connection
	stateManager.RegisterConnectionStatusCallback(func(oldValue, newValue core.ConnectionStatus) {
		if newValue == core.ConnectionStatusConnected {
			close(connectionEstablished)
		}
	})

	select {
	case <-connectionEstablished:
		log.Println("Bluetooth connection established")
	case <-connectionTimeout:
		log.Println("Timeout waiting for Bluetooth connection")
		bluetoothManager.DisconnectBluetooth()
		return
	}

	// Setup GSPro integration if enabled
	if config.EnableGSPro {
		log.Println("Starting GSPro integration")
		gsproIntegration := core.NewGSProIntegration(stateManager, launchMonitor, config.GSProIP, config.GSProPort)
		gsproIntegration.Start()
	}

	// Wait for interrupt signal to gracefully shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-sigChan
	log.Println("Shutting down...")

	// Clean up
	bluetoothManager.DisconnectBluetooth()

	// Give everything a moment to clean up
	time.Sleep(1 * time.Second)
	log.Println("Application stopped")
}

func main() {
	// Parse command line flags
	useMock := flag.String("mock", "", "Mock mode: 'stub' for basic mock, 'simulate' for simulated device with realistic behavior, or empty for real hardware")
	deviceName := flag.String("device", "", "Name of the Bluetooth device to connect to")
	headless := flag.Bool("headless", false, "Run in headless CLI mode without UI")
	gsproIP := flag.String("gspro-ip", "127.0.0.1", "IP address of GSPro server")
	gsproPort := flag.Int("gspro-port", 921, "Port of GSPro server")
	enableGSPro := flag.Bool("enable-gspro", false, "Enable GSPro integration")
	flag.Parse()

	// Create configuration
	config := AppConfig{
		UseMock:     core.MockMode(*useMock),
		DeviceName:  *deviceName,
		Headless:    *headless,
		GSProIP:     *gsproIP,
		GSProPort:   *gsproPort,
		EnableGSPro: *enableGSPro,
	}

	// Initialize common backend components
	stateManager, bluetoothManager, launchMonitor := initializeBackend(config)

	// Launch the appropriate interface based on mode
	if config.Headless {
		startCLI(config, stateManager, bluetoothManager, launchMonitor)
	} else {
		startUI(config, stateManager, bluetoothManager, launchMonitor)
	}
}
