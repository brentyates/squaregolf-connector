#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
APP_NAME="SquareGolf Connector.app"
APP_DIR="$BUILD_DIR/$APP_NAME"
CONTENTS_DIR="$APP_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"
ICONSET_DIR="$BUILD_DIR/SquareGolfConnector.iconset"
ICON_FILE="$RESOURCES_DIR/SquareGolfConnector.icns"

if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "This script only builds the macOS app bundle." >&2
    exit 1
fi

rm -rf "$APP_DIR" "$ICONSET_DIR"
mkdir -p "$MACOS_DIR" "$RESOURCES_DIR" "$ICONSET_DIR"

go build -o "$MACOS_DIR/squaregolf-connector" "$ROOT_DIR/main.go"

cp "$ROOT_DIR/macos/Info.plist" "$CONTENTS_DIR/Info.plist"
cp -R "$ROOT_DIR/web" "$RESOURCES_DIR/web"

sips -z 16 16     "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_16x16.png" >/dev/null
sips -z 32 32     "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_16x16@2x.png" >/dev/null
sips -z 32 32     "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_32x32.png" >/dev/null
sips -z 64 64     "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_32x32@2x.png" >/dev/null
sips -z 128 128   "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_128x128.png" >/dev/null
sips -z 256 256   "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_128x128@2x.png" >/dev/null
sips -z 256 256   "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_256x256.png" >/dev/null
sips -z 512 512   "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_256x256@2x.png" >/dev/null
sips -z 512 512   "$ROOT_DIR/icon.png" --out "$ICONSET_DIR/icon_512x512.png" >/dev/null
cp "$ROOT_DIR/icon.png" "$ICONSET_DIR/icon_512x512@2x.png"

iconutil -c icns "$ICONSET_DIR" -o "$ICON_FILE"
rm -rf "$ICONSET_DIR"

chmod +x "$MACOS_DIR/squaregolf-connector"
codesign --force --deep --sign - "$APP_DIR" >/dev/null

echo "Built $APP_DIR"
