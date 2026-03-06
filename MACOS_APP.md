# Building The macOS App

This project is intended to ship as a regular macOS app bundle.

## Build

```bash
./scripts/build-macos-app.sh
```

That produces:

```text
build/SquareGolf Connector.app
```

To build the GitHub Release artifact:

```bash
./scripts/package-macos-release.sh v1.0.0
```

That produces:

```text
dist/SquareGolf-Connector-v1.0.0-macOS.zip
```

The build script:

- Compiles the Go executable into the app bundle
- Generates an `.icns` file from `icon.png`
- Writes the bundled `Info.plist`
- Applies ad-hoc signing to the finished app bundle

## Run

Open the generated app from Finder or run:

```bash
open "build/SquareGolf Connector.app"
```

For a quick smoke test without hardware:

```bash
open -n "build/SquareGolf Connector.app" --args --mock=stub
```

## Why use the app bundle

- macOS gives the app a proper bundle identifier: `com.squaregolf.connector`
- Bluetooth permission prompts have app metadata to attach to
- WebKit runs with the normal app identity instead of a raw terminal binary
- The Dock icon, app name, and windowing behavior are cleaner

## Notes

- The app still serves the frontend from a local embedded web server
- Closing the app window shuts down the connector
- `--headless` is still available if you want CLI-only behavior
- The raw `build/squaregolf-connector` binary is still useful for development and debugging
- GitHub Releases can be built automatically with [.github/workflows/release-macos.yml](/Users/byates/projects/squaregolf-connector/.github/workflows/release-macos.yml)
