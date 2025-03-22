package core

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

var (
	bluetoothInstance *BluetoothManager
	bluetoothOnce     sync.Once
)

// GetBluetoothInstance returns the singleton instance of BluetoothManager
func GetBluetoothInstance(stateManager *StateManager) *BluetoothManager {
	bluetoothOnce.Do(func() {
		bluetoothInstance = &BluetoothManager{
			stateManager: stateManager,
		}
	})
	return bluetoothInstance
}

// NewBluetoothManager is deprecated, use GetBluetoothInstance instead
func NewBluetoothManager(stateManager *StateManager) *BluetoothManager {
	return GetBluetoothInstance(stateManager)
}

// BluetoothManager is responsible for handling Bluetooth connection logic
type BluetoothManager struct {
	bluetoothClient     BluetoothClient
	stateManager        *StateManager
	currentCtx          context.Context
	currentCancelFunc   context.CancelFunc
	connectionMutex     sync.Mutex
	notificationHandler func(uuid string, data []byte)
	connectMutex        sync.Mutex
	connecting          bool
	preDisconnectHook   func() // Hook to run before disconnecting
}

// GetClient returns the current Bluetooth client
func (bm *BluetoothManager) GetClient() BluetoothClient {
	return bm.bluetoothClient
}

// SetClient sets a new Bluetooth client
func (bm *BluetoothManager) SetClient(client BluetoothClient) {
	bm.bluetoothClient = client
}

// StartBluetoothConnection starts the Bluetooth connection in a background goroutine
func (bm *BluetoothManager) StartBluetoothConnection(deviceName, deviceAddress string) {
	bm.connectionMutex.Lock()
	defer bm.connectionMutex.Unlock()

	log.Printf("BluetoothManager: Starting connection to device: %s", deviceName)

	// Cancel any existing connection attempt
	if bm.currentCancelFunc != nil {
		log.Println("BluetoothManager: Cancelling existing connection attempt")
		bm.currentCancelFunc()
	}

	// Check if the Bluetooth client is nil and reinitialize if needed
	if bm.bluetoothClient == nil {
		log.Println("BluetoothManager: Bluetooth client is nil, reinitializing...")

		// Add a small delay before reinitialization to ensure resources are fully released
		time.Sleep(500 * time.Millisecond)

		// We'll create a real TinyGo client since that's what was used in the logs
		bleClient, err := NewTinyGoBluetoothClient()
		if err != nil {
			log.Printf("BluetoothManager: Failed to initialize Bluetooth client: %v", err)
			bm.stateManager.SetLastError(fmt.Errorf("Failed to initialize Bluetooth: %v", err))
			bm.stateManager.SetConnectionStatus(ConnectionStatusError)

			// Try to explicitly release adapter resources in case of failure
			if globalAdapter != nil {
				go func() {
					time.Sleep(1 * time.Second)
					releaseAdapter()
				}()
			}

			return
		}

		// Set the new client
		bm.bluetoothClient = bleClient
		log.Println("BluetoothManager: Successfully reinitialized Bluetooth client")
	}

	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	bm.currentCtx = ctx
	bm.currentCancelFunc = cancel

	// Update state
	bm.stateManager.SetConnectionStatus(ConnectionStatusConnecting)

	// Start connection in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Printf("BluetoothManager: Recovered from panic in Bluetooth connection: %v\nStack trace:\n%s", r, stack)
				bm.stateManager.SetLastError(fmt.Errorf("Panic: %v", r))
				bm.stateManager.SetConnectionStatus(ConnectionStatusError)
			}
		}()

		err := bm.connectDevice(ctx, deviceName, deviceAddress)
		if err != nil {
			if ctx.Err() == context.Canceled {
				log.Println("BluetoothManager: Bluetooth connection was cancelled")
			} else {
				log.Printf("BluetoothManager: Error in Bluetooth connection: %v", err)
				bm.stateManager.SetLastError(err)
				bm.stateManager.SetConnectionStatus(ConnectionStatusError)
			}
		}
	}()
}

// CancelBluetoothConnection cancels an in-progress connection attempt
func (bm *BluetoothManager) CancelBluetoothConnection() {
	bm.connectionMutex.Lock()
	defer bm.connectionMutex.Unlock()

	if bm.currentCancelFunc != nil {
		bm.currentCancelFunc()
		bm.currentCancelFunc = nil
	}

	bm.stateManager.SetConnectionStatus(ConnectionStatusDisconnected)
}

// DisconnectBluetooth disconnects from the current device
func (bm *BluetoothManager) DisconnectBluetooth() {
	// First, cancel any in-progress connection attempt
	bm.connectionMutex.Lock()
	if bm.currentCancelFunc != nil {
		log.Println("BluetoothManager: Cancelling in-progress connection attempt")
		bm.currentCancelFunc()
		bm.currentCancelFunc = nil
	}
	bm.connectionMutex.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Printf("Recovered from panic in Bluetooth disconnection: %v\nStack trace:\n%s", r, stack)
			}
		}()

		// Execute pre-disconnect hook if set
		if bm.preDisconnectHook != nil {
			log.Println("BluetoothManager: Executing pre-disconnect hook")
			bm.preDisconnectHook()
		}

		err := bm.disconnectDevice()
		if err != nil {
			log.Printf("Error disconnecting bluetooth: %v", err)
		}

		// Reset states
		bm.stateManager.SetConnectionStatus(ConnectionStatusDisconnected)
		bm.stateManager.SetBatteryLevel(nil)
		bm.stateManager.SetDeviceDisplayName(nil)
	}()
}

// SetNotificationHandler sets the handler for BLE notifications
func (bm *BluetoothManager) SetNotificationHandler(handler func(uuid string, data []byte)) {
	bm.notificationHandler = handler
}

// EnableNotifications enables BLE notifications
func (bm *BluetoothManager) EnableNotifications() error {
	if bm.bluetoothClient == nil {
		return fmt.Errorf("bluetooth client is nil")
	}

	if bm.notificationHandler == nil {
		return fmt.Errorf("notification handler is not set")
	}

	// Enable main notifications
	err := bm.bluetoothClient.StartNotifications(NotificationCharUUID, func(data []byte) {
		bm.notificationHandler(NotificationCharUUID, data)
	})
	if err != nil {
		return fmt.Errorf("error enabling main notifications: %w", err)
	}

	// Enable battery level notifications
	err = bm.bluetoothClient.StartNotifications(BatteryLevelCharUUID, func(data []byte) {
		bm.notificationHandler(BatteryLevelCharUUID, data)
	})
	if err != nil {
		log.Printf("Could not enable battery level notifications: %v", err)
		// Continue even if battery notifications fail
	} else {
	}

	return nil
}

// WriteCharacteristic writes a value to a specific characteristic UUID
func (bm *BluetoothManager) WriteCharacteristic(characteristicUUID string, data []byte) error {
	if bm.bluetoothClient == nil || !bm.bluetoothClient.IsConnected() {
		return fmt.Errorf("not connected to device")
	}

	err := bm.bluetoothClient.WriteCharacteristic(characteristicUUID, data)
	if err != nil {
		return fmt.Errorf("error writing to characteristic %s: %w", characteristicUUID, err)
	}

	return nil
}

// ReadBatteryLevel reads the battery level from the device
func (bm *BluetoothManager) ReadBatteryLevel() (int, error) {
	if bm.bluetoothClient == nil || !bm.bluetoothClient.IsConnected() {
		return 0, fmt.Errorf("not connected to device")
	}

	batteryLevelBytes, err := bm.bluetoothClient.ReadCharacteristic(BatteryLevelCharUUID)
	if err != nil {
		return 0, fmt.Errorf("could not read battery level: %w", err)
	}

	if len(batteryLevelBytes) == 0 {
		return 0, fmt.Errorf("received empty battery level data")
	}

	batteryLevel := int(batteryLevelBytes[0])

	// Update state manager with battery level
	bm.stateManager.SetBatteryLevel(&batteryLevel)

	return batteryLevel, nil
}

// connectDevice connects to the BLE device
func (bm *BluetoothManager) connectDevice(ctx context.Context, deviceName, deviceAddress string) error {
	log.Printf("BluetoothManager: Connecting to device: %s", deviceName)

	// Update state manager
	bm.stateManager.SetConnectionStatus(ConnectionStatusConnecting)

	// Connect to the device
	err := bm.bluetoothClient.Connect(deviceName, deviceAddress)
	if err != nil {
		log.Printf("BluetoothManager: Failed to connect to device: %v", err)
		return fmt.Errorf("failed to connect to device: %w", err)
	}

	// Get the actual connected device name from the client
	connectedDeviceName := bm.bluetoothClient.GetConnectedDeviceName()
	log.Printf("BluetoothManager: Successfully connected to device: %s", connectedDeviceName)

	// Update state with device name
	bm.stateManager.SetConnectionStatus(ConnectionStatusConnected)
	bm.stateManager.SetDeviceDisplayName(&connectedDeviceName)

	// Read battery level
	batteryLevel, err := bm.ReadBatteryLevel()
	if err != nil {
		log.Printf("BluetoothManager: Could not read battery level: %v", err)
	} else {
		log.Printf("BluetoothManager: Battery level: %d%%", batteryLevel)
		bm.stateManager.SetBatteryLevel(&batteryLevel)
	}

	// Enable notifications
	err = bm.EnableNotifications()
	if err != nil {
		log.Printf("BluetoothManager: Failed to enable notifications: %v", err)
		return fmt.Errorf("failed to enable notifications: %w", err)
	}

	log.Printf("BluetoothManager: Successfully enabled notifications")
	return nil
}

// disconnectDevice disconnects from the BLE device
func (bm *BluetoothManager) disconnectDevice() error {
	if bm.bluetoothClient == nil {
		return nil
	}

	// Make a local reference to client to ensure it's not nil during operations
	client := bm.bluetoothClient
	isConnected := client.IsConnected()

	// Stop notifications first if connected
	if isConnected {
		err := client.StopNotifications(NotificationCharUUID)
		if err != nil {
			log.Printf("Error stopping main notifications: %v", err)
		}

		err = client.StopNotifications(BatteryLevelCharUUID)
		if err != nil {
			log.Printf("Error stopping battery notifications: %v", err)
		}
	}

	// Attempt to disconnect regardless of IsConnected status
	// This handles cases where we're in the middle of connecting
	var disconnectErr error
	if isConnected {
		disconnectErr = client.Disconnect()
		if disconnectErr != nil {
			log.Printf("Error disconnecting: %v", disconnectErr)
		}
	} else {
		// Even if not fully connected, try to disconnect to clean up any partial connection state
		disconnectErr = client.Disconnect()
		if disconnectErr != nil {
			log.Printf("Error during cleanup: %v", disconnectErr)
		}
	}

	// Always update status after disconnect attempt, even if there was an error
	bm.stateManager.SetConnectionStatus(ConnectionStatusDisconnected)

	if disconnectErr != nil {
		return fmt.Errorf("error disconnecting: %w", disconnectErr)
	}
	return nil
}

// Initialize initializes the Bluetooth manager
func (bm *BluetoothManager) Initialize() error {
	// No initialization needed for the client itself
	return nil
}

// Connect initiates a connection to the device
func (bm *BluetoothManager) Connect() {
	bm.connectMutex.Lock()
	if bm.connecting {
		bm.connectMutex.Unlock()
		return
	}
	bm.connecting = true
	bm.connectMutex.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errMsg := fmt.Sprintf("Panic: %v", r)
				bm.stateManager.SetLastError(fmt.Errorf("%s", errMsg))
				bm.stateManager.SetConnectionStatus(ConnectionStatusError)
				bm.connecting = false
			}
		}()

		err := bm.connectToDevice()
		if err != nil {
			bm.stateManager.SetLastError(fmt.Errorf("%s", err.Error()))
			bm.stateManager.SetConnectionStatus(ConnectionStatusError)
			bm.connecting = false
			return
		}

		bm.connecting = false
	}()
}

// Disconnect disconnects from the device
func (bm *BluetoothManager) Disconnect() {
	bm.connectMutex.Lock()
	defer bm.connectMutex.Unlock()

	if !bm.bluetoothClient.IsConnected() {
		return
	}

	bm.bluetoothClient.Disconnect()
	bm.stateManager.SetConnectionStatus(ConnectionStatusDisconnected)
}

// Stop stops the Bluetooth manager
func (bm *BluetoothManager) Stop() {
	bm.connectMutex.Lock()
	defer bm.connectMutex.Unlock()

	if bm.bluetoothClient != nil {
		bm.bluetoothClient.Disconnect()
	}

	bm.stateManager.SetConnectionStatus(ConnectionStatusDisconnected)
	bm.stateManager.SetBatteryLevel(nil)
	bm.stateManager.SetDeviceDisplayName(nil)
}

// connectToDevice handles the actual connection process
func (bm *BluetoothManager) connectToDevice() error {
	// Get battery level before connecting
	batteryLevelBytes, err := bm.bluetoothClient.ReadCharacteristic(BatteryLevelCharUUID)
	if err != nil {
		log.Printf("Failed to get battery level: %v", err)
		// Non-fatal error, continue with connection
	} else {
		batteryLevel := int(batteryLevelBytes[0])
		bm.stateManager.SetBatteryLevel(&batteryLevel)
	}

	// Connect to the device
	err = bm.bluetoothClient.Connect("", "") // Empty strings for both parameters since we're using the client's default behavior
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	// Set connecting state
	bm.stateManager.SetConnectionStatus(ConnectionStatusConnecting)

	// Subscribe to notifications
	err = bm.bluetoothClient.StartNotifications(NotificationCharUUID, func(data []byte) {
		if bm.notificationHandler != nil {
			bm.notificationHandler(NotificationCharUUID, data)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to notifications: %v", err)
	}

	// Update connection state
	bm.stateManager.SetConnectionStatus(ConnectionStatusConnected)
	deviceName := "Device" // Default name, should be provided by the client implementation
	bm.stateManager.SetDeviceDisplayName(&deviceName)

	// Get battery level after connecting
	batteryLevelBytes, err = bm.bluetoothClient.ReadCharacteristic(BatteryLevelCharUUID)
	if err == nil {
		batteryLevel := int(batteryLevelBytes[0])
		bm.stateManager.SetBatteryLevel(&batteryLevel)
	}

	return nil
}

// StartScan starts scanning for SquareGolf devices
func (bm *BluetoothManager) StartScan() error {
	if bm.bluetoothClient == nil {
		return fmt.Errorf("bluetooth client is nil")
	}

	return bm.bluetoothClient.StartScan("SquareGolf")
}

// StopScan stops scanning for devices
func (bm *BluetoothManager) StopScan() error {
	if bm.bluetoothClient == nil {
		return fmt.Errorf("bluetooth client is nil")
	}

	bm.bluetoothClient.StopScan()
	return nil
}

// GetDiscoveredDevices returns the list of discovered SquareGolf devices
func (bm *BluetoothManager) GetDiscoveredDevices() []string {
	if bm.bluetoothClient == nil {
		return nil
	}

	return bm.bluetoothClient.GetDiscoveredDevices()
}

// SetPreDisconnectHook sets a function to be called before disconnecting
func (bm *BluetoothManager) SetPreDisconnectHook(hook func()) {
	bm.preDisconnectHook = hook
}
