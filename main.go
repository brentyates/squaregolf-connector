package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	appcfg "github.com/brentyates/squaregolf-connector/internal/config"
	"github.com/brentyates/squaregolf-connector/internal/core"
	"github.com/brentyates/squaregolf-connector/internal/core/camera"
	"github.com/brentyates/squaregolf-connector/internal/core/gspro"
	"github.com/brentyates/squaregolf-connector/internal/logging"
	"github.com/brentyates/squaregolf-connector/internal/web"
)

// Application configuration
type AppConfig struct {
	UseMock              core.MockMode
	DeviceName           string
	Headless             bool
	WebMode              bool
	WebPort              int
	GSProIP              string
	GSProPort            int
	EnableGSPro          bool
	EnableExternalCamera bool
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

	// Get the singleton bluetooth manager instance
	bluetoothManager := core.GetBluetoothInstance(stateManager)

	// Set the bluetooth client on the bluetooth manager
	bluetoothManager.SetClient(bleClient)

	// Get the singleton launch monitor instance
	launchMonitor := core.GetLaunchMonitorInstance(stateManager, bluetoothManager)

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
		gsproIntegration := gspro.GetInstance(stateManager, launchMonitor, config.GSProIP, config.GSProPort)
		gsproIntegration.EnableAutoReconnect()
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

// startWebServer initializes and runs the web server
func startWebServer(config AppConfig, stateManager *core.StateManager, bluetoothManager *core.BluetoothManager, launchMonitor *core.LaunchMonitor) {
	// Initialize config manager and load settings (happens behind the scenes like Fyne)
	settings := appcfg.GetInstance().GetSettings()

	// Apply loaded settings to state manager
	appcfg.GetInstance().ApplyToStateManager(stateManager)

	// Initialize camera manager only if external camera feature is enabled
	var cameraManager *camera.Manager
	if config.EnableExternalCamera {
		cameraManager = camera.GetInstance(stateManager, settings.CameraURL, settings.CameraEnabled)
	}

	// Create web server
	server := web.NewServer(stateManager, bluetoothManager, launchMonitor, cameraManager, config.GSProIP, config.GSProPort, config.EnableExternalCamera)

	// Setup GSPro integration if enabled via command line OR auto-connect is enabled in settings
	if config.EnableGSPro || settings.GSProAutoConnect {
		log.Println("GSPro integration enabled for web mode")
		// Use command line args if provided, otherwise use saved settings
		gsproIP := config.GSProIP
		gsproPort := config.GSProPort
		if !config.EnableGSPro && settings.GSProAutoConnect {
			gsproIP = settings.GSProIP
			gsproPort = settings.GSProPort
			log.Printf("Auto-connecting to GSPro at %s:%d", gsproIP, gsproPort)
		}

		gsproIntegration := gspro.GetInstance(stateManager, launchMonitor, gsproIP, gsproPort)
		go func() {
			gsproIntegration.EnableAutoReconnect()
			gsproIntegration.Start()
			gsproIntegration.Connect(gsproIP, gsproPort)
		}()
	}

	// Start auto-connect if device name is provided via command line OR if auto-connect is enabled in settings
	if config.DeviceName != "" {
		log.Printf("Auto-connecting to device: %s", config.DeviceName)
		go bluetoothManager.StartBluetoothConnection(config.DeviceName, "")
	} else if settings.AutoConnect && settings.DeviceName != "" {
		log.Printf("Auto-connecting to saved device: %s", settings.DeviceName)
		go bluetoothManager.StartBluetoothConnection(settings.DeviceName, "")
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down web server...")
		bluetoothManager.DisconnectBluetooth()
		os.Exit(0)
	}()

	// Start the web server
	log.Printf("Starting web server on http://localhost:%d", config.WebPort)
	if err := server.Start(config.WebPort); err != nil {
		log.Fatalf("Web server failed to start: %v", err)
	}
}

func main() {
	// Parse command line flags
	useMock := flag.String("mock", "", "Mock mode: 'stub' for basic mock, 'simulate' for simulated device with realistic behavior, or empty for real hardware")
	deviceName := flag.String("device", "", "Name of the Bluetooth device to connect to")
	headless := flag.Bool("headless", false, "Run in headless CLI mode without UI")
	webPort := flag.Int("web-port", 8080, "Port for web server")
	gsproIP := flag.String("gspro-ip", "127.0.0.1", "IP address of GSPro server")
	gsproPort := flag.Int("gspro-port", 921, "Port of GSPro server")
	enableGSPro := flag.Bool("enable-gspro", false, "Enable GSPro integration")
	enableExternalCamera := flag.Bool("enable-external-camera", false, "Enable external camera integration (experimental)")
	flag.Parse()

	// Create configuration
	config := AppConfig{
		UseMock:              core.MockMode(*useMock),
		DeviceName:           *deviceName,
		Headless:             *headless,
		WebMode:              !*headless,
		WebPort:              *webPort,
		GSProIP:              *gsproIP,
		GSProPort:            *gsproPort,
		EnableGSPro:          *enableGSPro,
		EnableExternalCamera: *enableExternalCamera,
	}

	// Initialize common backend components
	stateManager, bluetoothManager, launchMonitor := initializeBackend(config)

	// Launch the appropriate interface based on mode
	if config.Headless {
		startCLI(config, stateManager, bluetoothManager, launchMonitor)
	} else {
		startWebServer(config, stateManager, bluetoothManager, launchMonitor)
	}
}
