# Running SquareGolf Connector as a Daemon

This guide explains how to run the SquareGolf Connector as a background daemon/service on macOS.

## Features

The daemon is designed to run reliably in the background with:
- **Smart Auto-Reconnection**: Automatically reconnects to GSPro if the connection drops
- **Exponential Backoff**: Starts with 5s retry delay, increasing to max 30 minutes
- **Connection Timeout**: Stops trying after 20 minutes or 20 failed attempts
- **Manual Override**: Use the web UI to manually reconnect after timeout
- **TCP Keepalive**: Detects stale connections automatically
- **Proper Logging**: All activity logged to `~/Library/Logs/SquareGolf Connector/`

## macOS Setup (launchd)

### Prerequisites

1. Build and install the connector binary:
```bash
go build -o squaregolf-connector
sudo cp squaregolf-connector /usr/local/bin/
sudo chmod +x /usr/local/bin/squaregolf-connector
```

### Installation Steps

1. **Edit the plist file** to match your setup:
```bash
# Open the plist file
nano com.squaregolf.connector.plist

# Update the following:
# - Replace YOUR_USERNAME with your actual username in log paths
# - Adjust GSPro IP/port if needed
# - Add device name if you want auto-connect: --device "YourDeviceName"
```

2. **Copy the plist to LaunchAgents**:
```bash
cp com.squaregolf.connector.plist ~/Library/LaunchAgents/
```

3. **Load the service**:
```bash
launchctl load ~/Library/LaunchAgents/com.squaregolf.connector.plist
```

### Managing the Service

**Start the service:**
```bash
launchctl start com.squaregolf.connector
```

**Stop the service:**
```bash
launchctl stop com.squaregolf.connector
```

**Restart the service:**
```bash
launchctl stop com.squaregolf.connector
launchctl start com.squaregolf.connector
```

**Unload the service (disable):**
```bash
launchctl unload ~/Library/LaunchAgents/com.squaregolf.connector.plist
```

**Check if service is running:**
```bash
launchctl list | grep squaregolf
```

### Viewing Logs

View daemon logs:
```bash
# Main application logs (JSON format)
tail -f ~/Library/Logs/SquareGolf\ Connector/connector.log

# Daemon stdout
tail -f ~/Library/Logs/SquareGolf\ Connector/daemon-stdout.log

# Daemon stderr (errors)
tail -f ~/Library/Logs/SquareGolf\ Connector/daemon-stderr.log
```

## Web UI Access

Once the daemon is running, access the web UI at:
```
http://localhost:8080
```

The web UI can be used to:
- Connect/disconnect from devices
- Connect/disconnect from GSPro manually
- Monitor connection status
- View shot data

## Configuration Options

You can customize the daemon by editing the `ProgramArguments` section in the plist file:

```xml
<key>ProgramArguments</key>
<array>
    <string>/usr/local/bin/squaregolf-connector</string>
    <string>--enable-gspro</string>
    <string>--gspro-ip</string>
    <string>127.0.0.1</string>
    <string>--gspro-port</string>
    <string>921</string>
    <string>--web-port</string>
    <string>8080</string>
    <string>--device</string>
    <string>Your Device Name</string>  <!-- Optional: auto-connect to device -->
</array>
```

Available flags:
- `--enable-gspro`: Enable GSPro integration
- `--gspro-ip`: GSPro server IP (default: 127.0.0.1)
- `--gspro-port`: GSPro server port (default: 921)
- `--web-port`: Web server port (default: 8080)
- `--device`: Device name for auto-connection (optional)
- `--mock`: Use mock device ("stub" or "simulate") for testing

## Reconnection Behavior

### Automatic Reconnection
The daemon will automatically try to reconnect when:
- GSPro connection drops
- Network connectivity is restored
- GSPro server restarts

### Reconnection Timeout
The daemon will **stop trying** and require manual reconnection when:
- More than 20 minutes of reconnection attempts have elapsed
- OR more than 20 consecutive connection failures occur

After timeout:
1. Open the web UI at http://localhost:8080
2. Click "Connect" in the GSPro section
3. This will reset the reconnection counter and retry

### Expected Behavior
- **GSPro temporarily closed**: Daemon will retry for up to 20 minutes
- **GSPro closed for extended time**: Daemon will stop retrying after 20 minutes, manual reconnect needed
- **Brief network issue**: Reconnects automatically within seconds
- **Device turned off**: Daemon continues running, ready for device reconnection

## Troubleshooting

### Service won't start
```bash
# Check for errors in stderr log
cat ~/Library/Logs/SquareGolf\ Connector/daemon-stderr.log

# Verify binary exists and is executable
ls -la /usr/local/bin/squaregolf-connector

# Check launchd service status
launchctl list | grep squaregolf
```

### Permission denied errors
```bash
# Ensure binary has correct permissions
sudo chmod +x /usr/local/bin/squaregolf-connector

# Ensure log directory exists
mkdir -p ~/Library/Logs/SquareGolf\ Connector
```

### Can't connect to GSPro
1. Verify GSPro is running and OpenAPI connector is enabled
2. Check the connection settings in the web UI
3. View logs for specific error messages:
   ```bash
   tail -50 ~/Library/Logs/SquareGolf\ Connector/connector.log | grep -i gspro
   ```

### Reset everything
```bash
# Stop and unload the service
launchctl unload ~/Library/LaunchAgents/com.squaregolf.connector.plist

# Wait a moment
sleep 2

# Reload and restart
launchctl load ~/Library/LaunchAgents/com.squaregolf.connector.plist
launchctl start com.squaregolf.connector
```

## Uninstalling

To completely remove the daemon:
```bash
# Stop and unload the service
launchctl unload ~/Library/LaunchAgents/com.squaregolf.connector.plist

# Remove the plist
rm ~/Library/LaunchAgents/com.squaregolf.connector.plist

# Optional: Remove the binary
sudo rm /usr/local/bin/squaregolf-connector

# Optional: Remove logs
rm -rf ~/Library/Logs/SquareGolf\ Connector
```

## Notes

- The daemon runs in **web mode** by default, not headless mode
- Access the web UI to monitor status and control connections
- Logs are rotated automatically (5MB per file, 5 backups, compressed)
- The service will restart automatically if it crashes (with 10 second throttle)
- Use `--mock stub` or `--mock simulate` for testing without hardware

