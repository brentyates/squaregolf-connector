# SquareGolf Connector

An unofficial launch monitor connector for SquareGolf devices with GSPro integration.

## Download And Run (macOS)

1. Open the [latest release](https://github.com/byates/squaregolf-connector/releases/latest).
2. Download the macOS zip file: `SquareGolf-Connector-<version>-macOS.zip`.
3. Unzip it.
4. Drag `SquareGolf Connector.app` into `Applications` if you want.
5. Open `SquareGolf Connector.app`.

When the app is running, it opens its own desktop window. Closing the window shuts the app down.

## First Launch On macOS

Because this app is not signed with a paid Apple Developer ID, macOS may warn that it is from an unidentified developer.

If that happens:

1. Right-click the app and choose `Open`.
2. Click `Open` again in the confirmation dialog.

If macOS still blocks it:

1. Open `System Settings`.
2. Go to `Privacy & Security`.
3. Find the message about `SquareGolf Connector`.
4. Click `Open Anyway`.

## What It Does

SquareGolf Connector connects to SquareGolf Bluetooth launch monitors and provides:

- **Bluetooth connectivity** to SquareGolf devices
- **GSPro integration** with automatic reconnection
- **Desktop window UI** powered by the existing web frontend
- **External camera integration** (experimental)
- **Persistent saved settings**

## Features

- Real-time ball and club metrics
- Battery monitoring
- Device alignment tracking
- Ball detection and position tracking
- Configurable club selection and handedness
- Persistent settings storage
- Auto-connect functionality

## Requirements

- **macOS** for the ready-to-download app release
- Bluetooth adapter
- A SquareGolf launch monitor for normal use

Windows is still a supported development target in the codebase, but the automated downloadable release currently targets macOS.

## Quick Start

1. Launch `SquareGolf Connector.app`.
2. Turn on your SquareGolf device.
3. Open GSPro if you use it.
4. In the app, connect to your device.
5. If needed, connect GSPro from the app settings.

## Troubleshooting

### macOS says the app cannot be opened

- Right-click the app and choose `Open`
- If needed, allow it in `System Settings` -> `Privacy & Security`

### Cannot connect to Bluetooth device

- Ensure your Bluetooth adapter is enabled
- Make sure your SquareGolf device is turned on
- Move the device closer to your Mac

### GSPro not receiving data

- Make sure GSPro is open
- Check the connection settings in the app
- Enable auto-reconnect in settings

### App window opens but looks blank or broken

- Quit the app and open it again
- Download the latest release from GitHub
- If the problem continues, open an issue on GitHub

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Disclaimer

This is an unofficial, community-developed connector and is not affiliated with or endorsed by SquareGolf.
