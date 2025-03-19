package core

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

// Global adapter state variables to manage singleton pattern
var (
	globalAdapter     *bluetooth.Adapter
	globalAdapterMu   sync.Mutex
	adapterInUse      bool
	adapterCooldownCh = make(chan struct{}, 1)
)

// initializeAdapter initializes the Bluetooth adapter with proper synchronization
func initializeAdapter() (*bluetooth.Adapter, error) {
	globalAdapterMu.Lock()
	defer globalAdapterMu.Unlock()

	// If adapter is marked as in use, wait for it to become available
	if adapterInUse {
		log.Println("Adapter in use, waiting for release...")
		globalAdapterMu.Unlock()
		select {
		case <-adapterCooldownCh:
			log.Println("Adapter released, continuing initialization")
		case <-time.After(3 * time.Second):
			log.Println("Timed out waiting for adapter release, forcing re-initialization")
		}
		globalAdapterMu.Lock()
	}

	// Initialize adapter if it doesn't exist
	if globalAdapter == nil {
		adapter := bluetooth.DefaultAdapter
		err := adapter.Enable()
		if err != nil {
			return nil, fmt.Errorf("failed to enable Bluetooth adapter: %w", err)
		}
		globalAdapter = adapter
	}

	adapterInUse = true
	return globalAdapter, nil
}

// releaseAdapter marks the adapter as no longer in use
func releaseAdapter() {
	globalAdapterMu.Lock()
	defer globalAdapterMu.Unlock()

	// Reset adapter state
	if adapterInUse {
		adapterInUse = false

		// Signal that the adapter is released
		select {
		case adapterCooldownCh <- struct{}{}:
		default:
		}

		log.Println("Bluetooth adapter released")
	}
}

// TinyGoBluetoothClient implements BluetoothClient interface using tinygo-org/bluetooth
type TinyGoBluetoothClient struct {
	adapter              *bluetooth.Adapter
	device               *bluetooth.Device
	connected            bool
	mutex                sync.Mutex
	characteristics      map[string]*bluetooth.DeviceCharacteristic
	notificationHandlers map[string]func([]byte)
	connectedDeviceName  string // Store the name of the connected device

	// New fields for scan management
	scanning    bool
	scanResults map[string]bluetooth.ScanResult
	scanMutex   sync.Mutex
	scanDone    chan struct{}
}

// NewTinyGoBluetoothClient creates a new Bluetooth client using tinygo-org/bluetooth
func NewTinyGoBluetoothClient() (*TinyGoBluetoothClient, error) {
	// Use the singleton adapter with proper synchronization
	adapter, err := initializeAdapter()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
	}

	return &TinyGoBluetoothClient{
		adapter:              adapter,
		connected:            false,
		characteristics:      make(map[string]*bluetooth.DeviceCharacteristic),
		notificationHandlers: make(map[string]func([]byte)),
		scanResults:          make(map[string]bluetooth.ScanResult),
	}, nil
}

// StartScan starts scanning for BLE devices in the background
func (t *TinyGoBluetoothClient) StartScan(prefix string) error {
	t.scanMutex.Lock()
	defer t.scanMutex.Unlock()

	if t.scanning {
		log.Println("Scan already in progress")
		return nil
	}

	// Reset scan results
	t.scanResults = make(map[string]bluetooth.ScanResult)
	t.scanDone = make(chan struct{})
	t.scanning = true

	log.Printf("Starting Bluetooth scan for devices with prefix: %s", prefix)

	// Start scan in a goroutine
	go func() {
		err := t.adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			addrStr := device.Address.String()
			name := device.LocalName()

			// Only store devices that match the prefix
			if prefix == "" || (name != "" && len(name) >= len(prefix) && name[:len(prefix)] == prefix) {
				log.Printf("Scan found matching device: %s [%s]", name, addrStr)
				t.scanMutex.Lock()
				t.scanResults[addrStr] = device
				t.scanMutex.Unlock()
			}
		})

		t.scanMutex.Lock()
		t.scanning = false
		close(t.scanDone)
		t.scanMutex.Unlock()

		if err != nil {
			log.Printf("Error during scan: %v", err)
		}

		log.Println("Scan completed")
	}()

	return nil
}

// StopScan stops an ongoing BLE scan
func (t *TinyGoBluetoothClient) StopScan() error {
	t.scanMutex.Lock()
	defer t.scanMutex.Unlock()

	if t.scanning {
		log.Println("Stopping scan...")
		t.adapter.StopScan()
		t.scanning = false
	}

	return nil
}

// GetScanResults returns the current scan results
func (t *TinyGoBluetoothClient) GetScanResults() map[string]bluetooth.ScanResult {
	t.scanMutex.Lock()
	defer t.scanMutex.Unlock()

	// Return a copy to avoid concurrent access issues
	results := make(map[string]bluetooth.ScanResult, len(t.scanResults))
	for k, v := range t.scanResults {
		results[k] = v
	}

	return results
}

// GetConnectedDeviceName returns the name of the currently connected device
func (t *TinyGoBluetoothClient) GetConnectedDeviceName() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connectedDeviceName
}

// Connect connects to the BLE device
func (t *TinyGoBluetoothClient) Connect(targetName, targetAddress string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Println("Starting Bluetooth connection process...")

	if t.connected {
		log.Println("Already connected to device, skipping connection")
		return nil
	}

	// Try to find the device in already scanned results
	t.scanMutex.Lock()
	var deviceToConnect bluetooth.ScanResult
	var found bool

	// Check if we have the device in our scan results
	for _, result := range t.scanResults {
		if (targetName != "" && result.LocalName() == targetName) ||
			(targetAddress != "" && result.Address.String() == targetAddress) ||
			(targetName == "" && strings.HasPrefix(result.LocalName(), BluetoothDevicePrefix)) {
			deviceToConnect = result
			found = true
			break
		}
	}

	// Stop any existing scan since we're going to connect
	if t.scanning {
		log.Println("Stopping existing scan before connection attempt...")
		t.adapter.StopScan()
		t.scanning = false
	}

	t.scanMutex.Unlock()

	// If device not found in existing scan results, start a new scan
	if !found {
		log.Printf("Target device not in scan results, starting new scan for '%s' or '%s'...", targetName, targetAddress)

		// Use a local scan just for this connection attempt
		foundDevice := false

		err := t.adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
			if (targetName != "" && device.LocalName() == targetName) ||
				(targetAddress != "" && device.Address.String() == targetAddress) ||
				(targetName == "" && strings.HasPrefix(device.LocalName(), BluetoothDevicePrefix)) {
				log.Printf("Found target device: %s [%s]", device.LocalName(), device.Address.String())
				adapter.StopScan()
				deviceToConnect = device
				foundDevice = true
			}
		})

		if err != nil {
			log.Printf("Error starting scan: %v", err)
			return fmt.Errorf("failed to start scan: %w", err)
		}

		// Wait for device to be found
		log.Println("Waiting for device to be found (timeout: 10s)...")
		startTime := time.Now()
		for !foundDevice && time.Since(startTime) < 10*time.Second {
			time.Sleep(100 * time.Millisecond)
		}

		if !foundDevice {
			log.Println("Device not found after timeout")
			return errors.New("device not found")
		}
	}

	// Connect to the device
	log.Printf("Connecting to device %s [%s]...", deviceToConnect.LocalName(), deviceToConnect.Address.String())
	device, err := t.adapter.Connect(deviceToConnect.Address, bluetooth.ConnectionParams{})
	if err != nil {
		log.Printf("Error connecting to device: %v", err)
		return fmt.Errorf("failed to connect to device: %w", err)
	}
	log.Println("Successfully connected to device")
	t.device = &device
	t.connectedDeviceName = deviceToConnect.LocalName() // Store the connected device name

	// Discover services and characteristics
	log.Println("Discovering services...")
	services, err := device.DiscoverServices(nil)
	if err != nil {
		log.Printf("Error discovering services: %v", err)
		return fmt.Errorf("failed to discover services: %w", err)
	}
	log.Printf("Found %d services", len(services))

	for i, service := range services {
		log.Printf("Service %d: %s", i+1, service.UUID().String())
		log.Printf("Discovering characteristics for service %s...", service.UUID().String())
		characteristics, err := service.DiscoverCharacteristics(nil)
		if err != nil {
			log.Printf("Failed to discover characteristics for service %v: %v", service.UUID(), err)
			continue
		}
		log.Printf("Found %d characteristics for service %s", len(characteristics), service.UUID().String())

		for j, char := range characteristics {
			uuidStr := char.UUID().String()
			charCopy := char // Create a copy to avoid issues with loop variable capture
			t.characteristics[uuidStr] = &charCopy
			log.Printf("Characteristic %d.%d: %s", i+1, j+1, uuidStr)
		}
	}

	log.Println("Connection process completed successfully")
	t.connected = true
	return nil
}

// Disconnect disconnects from the BLE device
func (t *TinyGoBluetoothClient) Disconnect() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.device == nil {
		return nil
	}

	// Stop all notifications first
	for uuid := range t.notificationHandlers {
		t.StopNotifications(uuid)
	}

	// Stop any ongoing scan
	t.scanMutex.Lock()
	if t.scanning {
		log.Println("Stopping scan during disconnect...")
		t.adapter.StopScan()
		t.scanning = false
	}
	t.scanMutex.Unlock()

	err := t.device.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	t.connected = false
	t.device = nil
	t.characteristics = make(map[string]*bluetooth.DeviceCharacteristic)
	t.notificationHandlers = make(map[string]func([]byte))

	// Release the adapter after disconnection
	go func() {
		// Add a cooldown period before releasing the adapter
		// to ensure all resources are properly cleaned up
		time.Sleep(1 * time.Second)
		releaseAdapter()
	}()

	return nil
}

// WriteCharacteristic writes data to a characteristic
func (t *TinyGoBluetoothClient) WriteCharacteristic(uuid string, data []byte) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.device == nil {
		return errors.New("not connected")
	}

	char, ok := t.characteristics[uuid]
	if !ok {
		return fmt.Errorf("characteristic not found: %s", uuid)
	}

	_, err := char.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to characteristic: %w", err)
	}

	return nil
}

// ReadCharacteristic reads data from a characteristic
func (t *TinyGoBluetoothClient) ReadCharacteristic(uuid string) ([]byte, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.device == nil {
		return nil, errors.New("not connected")
	}

	char, ok := t.characteristics[uuid]
	if !ok {
		return nil, fmt.Errorf("characteristic not found: %s", uuid)
	}

	// Create a buffer to read the data into
	buf := make([]byte, 100) // Assuming maximum size of 100 bytes
	n, err := char.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read characteristic: %w", err)
	}

	return buf[:n], nil
}

// StartNotifications starts notifications for a characteristic
func (t *TinyGoBluetoothClient) StartNotifications(uuid string, handler func([]byte)) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.device == nil {
		return errors.New("not connected")
	}

	char, ok := t.characteristics[uuid]
	if !ok {
		return fmt.Errorf("characteristic not found: %s", uuid)
	}

	// Store the handler
	t.notificationHandlers[uuid] = handler

	// Create a safe notification handler that doesn't hold the mutex
	safeHandler := func(buf []byte) {
		if handler != nil {
			// Copying the data to avoid potential race conditions
			dataCopy := make([]byte, len(buf))
			copy(dataCopy, buf)

			// Use the handler directly instead of spawning a new goroutine
			// This avoids creating too many goroutines that might cause problems
			handler(dataCopy)
		}
	}

	// Enable notifications
	err := char.EnableNotifications(safeHandler)

	if err != nil {
		delete(t.notificationHandlers, uuid)
		return fmt.Errorf("failed to enable notifications: %w", err)
	}

	return nil
}

// StopNotifications stops notifications for a characteristic
func (t *TinyGoBluetoothClient) StopNotifications(uuid string) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.connected || t.device == nil {
		return errors.New("not connected")
	}

	//Just remove the handler since tinygo-org/bluetooth doesn't have a DisableNotifications method
	delete(t.notificationHandlers, uuid)
	log.Printf("Stopped notifications for %s", uuid)

	return nil
}

// IsConnected returns the connection status
func (t *TinyGoBluetoothClient) IsConnected() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.connected
}

// GetDiscoveredDevices returns the list of discovered devices
func (t *TinyGoBluetoothClient) GetDiscoveredDevices() []string {
	t.scanMutex.Lock()
	defer t.scanMutex.Unlock()

	// Convert scan results to a list of device names
	devices := make([]string, 0, len(t.scanResults))
	for _, result := range t.scanResults {
		if name := result.LocalName(); name != "" {
			devices = append(devices, name)
		}
	}

	return devices
}
