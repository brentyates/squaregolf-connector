# SquareGolf Connector

An unofficial launch monitor connector for SquareGolf devices with GSPro integration.

## Overview

SquareGolf Connector is a Go-based application that connects to SquareGolf Bluetooth launch monitors and provides:

- **Bluetooth connectivity** to SquareGolf devices
- **GSPro integration** with automatic reconnection
- **Web-based UI** for easy monitoring and control
- **External camera integration** (experimental)
- **Multiple operation modes** including headless CLI

## Features

- Real-time ball and club metrics
- Battery monitoring
- Device alignment tracking
- Ball detection and position tracking
- Configurable club selection and handedness
- Persistent settings storage
- Auto-connect functionality
- Mock and simulation modes for development

## Requirements

- Go 1.23 or later
- Bluetooth adapter (for real hardware)
- macOS, Linux, or Windows

## Installation

```bash
# Clone the repository
git clone https://github.com/brentyates/squaregolf-connector.git
cd squaregolf-connector

# Install dependencies
go mod download

# Build the application
go build -o squaregolf-connector main.go
```

## Usage

### Web Interface Mode (Default)

```bash
# Start with web UI
./squaregolf-connector

# Specify custom port
./squaregolf-connector --web-port=8080
```

The web interface will be available at `http://localhost:8080`

### Headless CLI Mode

```bash
# Run in headless mode with auto-connect
./squaregolf-connector --headless --device="SquareGolf-XXXX"
```

### GSPro Integration

```bash
# Enable GSPro integration
./squaregolf-connector --enable-gspro --gspro-ip=127.0.0.1 --gspro-port=921
```

### Mock Modes

```bash
# Basic stub mode (no real hardware required)
./squaregolf-connector --mock=stub

# Simulated device mode (realistic behavior)
./squaregolf-connector --mock=simulate
```

### External Camera Integration (Experimental)

```bash
# Enable external camera support
./squaregolf-connector --enable-external-camera
```

## Command-Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `--mock` | "" | Mock mode: 'stub' or 'simulate' |
| `--device` | "" | Bluetooth device name to auto-connect |
| `--headless` | false | Run in CLI mode without web UI |
| `--web-port` | 8080 | Port for web server |
| `--gspro-ip` | 127.0.0.1 | GSPro server IP address |
| `--gspro-port` | 921 | GSPro server port |
| `--enable-gspro` | false | Enable GSPro integration |
| `--enable-external-camera` | false | Enable external camera (experimental) |

## Configuration

Settings are automatically saved and loaded from:
- macOS: `~/Library/Application Support/squaregolf-connector/settings.json`
- Linux: `~/.config/squaregolf-connector/settings.json`
- Windows: `%APPDATA%\squaregolf-connector\settings.json`

## Development

### Running Tests

```bash
go test ./...
```

### Project Structure

```
.
├── main.go                      # Application entry point
├── internal/
│   ├── core/                    # Core business logic
│   │   ├── bluetooth_manager.go # Bluetooth connection management
│   │   ├── launch_monitor.go   # Launch monitor data handling
│   │   ├── state_manager.go    # Application state management
│   │   ├── gspro/               # GSPro integration
│   │   └── camera/              # Camera integration
│   ├── config/                  # Configuration management
│   ├── logging/                 # Logging utilities
│   ├── web/                     # Web server and API
│   └── version/                 # Version information
├── frontend/                    # Web UI assets
└── web/                         # Static web files
```

## How It Works

1. **Bluetooth Connection**: Connects to SquareGolf devices via Bluetooth Low Energy (BLE)
2. **Data Processing**: Parses ball and club metrics from device notifications
3. **State Management**: Maintains application state with reactive callbacks
4. **GSPro Integration**: Forwards shot data to GSPro in real-time
5. **Web Interface**: Provides a user-friendly dashboard for monitoring and control

## Troubleshooting

### Cannot connect to Bluetooth device

- Ensure your Bluetooth adapter is enabled
- Check that the device name is correct (starts with "SquareGolf")
- Try running with elevated privileges if on Linux

### GSPro not receiving data

- Verify GSPro is running and listening on the specified port
- Check firewall settings
- Enable auto-reconnect in settings

### Web UI not accessible

- Confirm the port is not already in use
- Check that no firewall is blocking the port
- Try a different port with `--web-port`

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Disclaimer

This is an unofficial, community-developed connector and is not affiliated with or endorsed by SquareGolf.
