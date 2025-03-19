package core

import (
	"fmt"
	"log"
)

// BluetoothClient interface defines the methods required for BLE communication
type BluetoothClient interface {
	Connect(deviceName, deviceAddress string) error
	Disconnect() error
	WriteCharacteristic(uuid string, data []byte) error
	ReadCharacteristic(uuid string) ([]byte, error)
	StartNotifications(uuid string, handler func([]byte)) error
	StopNotifications(uuid string) error
	IsConnected() bool
	StartScan(prefix string) error
	StopScan() error
	GetDiscoveredDevices() []string
	GetConnectedDeviceName() string
}

// WriteHistory represents a single write operation with its metadata
type WriteHistory struct {
	UUID string
	Data []byte
}

// MockBluetoothClient implements BluetoothClient interface for testing
type MockBluetoothClient struct {
	connected      bool
	writeCalled    bool
	readCalled     bool
	lastWriteData  []byte
	lastWriteUUID  string
	lastReadUUID   string
	readReturnData []byte
	readError      error
	writeError     error
	writeCount     int            // Track number of writes
	writeHistory   []WriteHistory // Track history of all writes
	deviceName     string         // Store the connected device name
}

// NewMockBluetoothClient creates a new mock Bluetooth client
func NewMockBluetoothClient() *MockBluetoothClient {
	return &MockBluetoothClient{
		connected:    false,
		writeCount:   0,
		writeHistory: make([]WriteHistory, 0),
	}
}

// Connect simulates connecting to a device
func (m *MockBluetoothClient) Connect(deviceName, deviceAddress string) error {
	m.connected = true
	m.deviceName = deviceName
	return nil
}

// Disconnect simulates disconnecting from a device
func (m *MockBluetoothClient) Disconnect() error {
	m.connected = false
	return nil
}

// WriteCharacteristic simulates writing to a characteristic
func (m *MockBluetoothClient) WriteCharacteristic(uuid string, data []byte) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	m.writeCalled = true
	m.lastWriteData = data
	m.lastWriteUUID = uuid
	m.writeCount++

	// Add to write history
	m.writeHistory = append(m.writeHistory, WriteHistory{
		UUID: uuid,
		Data: data,
	})

	log.Printf("Mock write to %s: %x", uuid, data)
	return m.writeError
}

// ReadCharacteristic simulates reading from a characteristic
func (m *MockBluetoothClient) ReadCharacteristic(uuid string) ([]byte, error) {
	if !m.connected {
		return nil, fmt.Errorf("not connected")
	}
	m.readCalled = true
	m.lastReadUUID = uuid
	if m.readReturnData != nil {
		return m.readReturnData, m.readError
	}
	if uuid == BatteryLevelCharUUID {
		return []byte{80}, nil // Mock 80% battery
	}
	return []byte{}, nil
}

// StartNotifications simulates starting notifications for a characteristic
func (m *MockBluetoothClient) StartNotifications(uuid string, handler func([]byte)) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	log.Printf("Mock start notifications for %s", uuid)
	return nil
}

// StopNotifications simulates stopping notifications for a characteristic
func (m *MockBluetoothClient) StopNotifications(uuid string) error {
	if !m.connected {
		return fmt.Errorf("not connected")
	}
	log.Printf("Mock stop notifications for %s", uuid)
	return nil
}

// IsConnected returns the connection status
func (m *MockBluetoothClient) IsConnected() bool {
	return m.connected
}

// GetWriteHistory returns the complete history of write operations
func (m *MockBluetoothClient) GetWriteHistory() []WriteHistory {
	return m.writeHistory
}

// ClearWriteHistory clears the write history
func (m *MockBluetoothClient) ClearWriteHistory() {
	m.writeHistory = make([]WriteHistory, 0)
}

// StartScan simulates starting a device scan
func (m *MockBluetoothClient) StartScan(prefix string) error {
	return nil
}

// StopScan simulates stopping a device scan
func (m *MockBluetoothClient) StopScan() error {
	return nil
}

// GetDiscoveredDevices returns a list of discovered devices
func (m *MockBluetoothClient) GetDiscoveredDevices() []string {
	return []string{"SquareGolf(****)"}
}

// GetConnectedDeviceName returns the name of the currently connected device
func (m *MockBluetoothClient) GetConnectedDeviceName() string {
	return m.deviceName
}
