# SquareGolf Connector - Web Interface

The SquareGolf Connector now supports a modern web-based interface as an alternative to the desktop Fyne UI.

## Features

The web interface includes all the same functionality as the desktop version:

### 🔗 Device Screen
- Bluetooth device connection and disconnection
- Real-time device status and battery level
- Ball detection and readiness status
- Ball position tracking
- Club and handedness information
- Live ball and club metrics from shots

### 🏌️ GSPro Screen
- GSPro server connection configuration
- Real-time connection status
- Auto-connect settings
- Connection troubleshooting guide

### ⚙️ Settings Screen
- Device name management
- Auto-connect preferences
- Spin detection mode (Standard/Advanced)
- Ball ready chime sound selection
- Volume control

### 🎯 Alignment Screen
- Device calibration interface (placeholder)

## Running in Web Mode

### Command Line Options
```bash
# Run web server on default port 8080
./squaregolf-connector -web

# Run web server on custom port
./squaregolf-connector -web -web-port=3000

# Run with GSPro integration enabled
./squaregolf-connector -web -enable-gspro

# Run with mock device for testing
./squaregolf-connector -web -mock=simulate

# Full example with all options
./squaregolf-connector -web -web-port=8080 -device="My Device" -enable-gspro -gspro-ip=192.168.1.100
```

### Available Modes
- **Desktop UI (default)**: Traditional Fyne-based desktop application
- **Web Mode (`-web`)**: Modern web-based interface
- **Headless (`-headless`)**: Command-line only mode

## Web Interface Features

### Real-time Updates
- WebSocket connection for live status updates
- Automatic reconnection if connection is lost
- Real-time device metrics and ball status

### Modern UI
- Responsive design that works on desktop and mobile
- Professional sidebar navigation
- Status indicators with color coding
- Toast notifications for actions and errors

### Audio Support
- Ball ready chime sounds play through the backend (not browser)
- Volume control and sound selection
- Preview sounds from settings

## API Endpoints

The web server exposes RESTful APIs:

### Device Control
- `GET /api/device/status` - Get current device status
- `POST /api/device/connect` - Connect to device
- `POST /api/device/disconnect` - Disconnect from device

### GSPro Control
- `GET /api/gspro/status` - Get GSPro connection status
- `POST /api/gspro/connect` - Connect to GSPro
- `POST /api/gspro/disconnect` - Disconnect from GSPro
- `GET/POST /api/gspro/config` - Get/Set GSPro configuration

### Settings
- `GET/POST /api/settings` - Get/Set application settings
- `GET /api/settings/chime/sounds` - Get available chime sounds
- `POST /api/settings/chime/play` - Play a chime sound

### WebSocket
- `ws://localhost:8080/ws` - Real-time status updates

## Architecture

### Backend (Go)
- Maintains all existing Bluetooth, GSPro, and launch monitor functionality
- HTTP/WebSocket server using Gorilla toolkit
- RESTful APIs for configuration and control
- Real-time status broadcasting via WebSocket

### Frontend (Web)
- Modern HTML5/CSS3/JavaScript interface
- No build process required - pure vanilla JS
- WebSocket client for real-time updates
- Responsive design with Material Icons

### Audio System
- Chime sounds are played by the Go backend, not the browser
- This ensures consistent audio playback across platforms
- Volume and sound selection controlled via web interface

## Browser Compatibility

The web interface works with modern browsers:
- Chrome/Edge 80+
- Firefox 75+
- Safari 13+

## Development and Testing

### Mock Mode
For development and testing without hardware:
```bash
./squaregolf-connector -web -mock=simulate
```

This provides realistic simulated device behavior including:
- Connection sequences
- Battery level changes
- Ball detection cycles
- Simulated shot metrics

### Debugging
- WebSocket connection status shown in sidebar
- Browser developer tools for debugging
- Go server logs for backend debugging

## Migration from Desktop UI

The web interface provides the same functionality as the desktop Fyne UI:

✅ **Complete Feature Parity**
- All device connection features
- All GSPro integration features  
- All settings and preferences
- All real-time status updates
- Audio chime support

✅ **Additional Benefits**
- Cross-platform compatibility
- Remote access capability
- Mobile-friendly responsive design
- No desktop environment required
- Easy integration with web-based golf simulators

The backend remains identical - only the user interface has changed from desktop to web-based.
