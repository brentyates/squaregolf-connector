#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="$ROOT_DIR/build"
APP_NAME="SquareGolf Connector"
APP_DIR="$BUILD_DIR/$APP_NAME"
EXE_NAME="SquareGolf Connector.exe"
SYSO_PATH="$ROOT_DIR/rsrc_windows_amd64.syso"

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
    VERSION="$(grep -E 'Version = "' "$ROOT_DIR/internal/version/version.go" | head -n1 | sed -E 's/.*"([^"]+)".*/\1/')"
fi

VERSION="${VERSION#v}"
NUMERIC_VERSION="$(echo "$VERSION" | sed -E 's/-.*$//')"
IFS='.' read -r MAJOR MINOR PATCH <<<"$NUMERIC_VERSION"
MAJOR="${MAJOR:-0}"
MINOR="${MINOR:-0}"
PATCH="${PATCH:-0}"
COMMA_VERSION="${MAJOR},${MINOR},${PATCH},0"

rm -rf "$APP_DIR" "$SYSO_PATH"
mkdir -p "$APP_DIR"

if ! command -v windres >/dev/null 2>&1; then
    echo "windres is required to build the Windows resource file" >&2
    exit 1
fi

RC_STAGE_DIR="$BUILD_DIR/rc-stage"
rm -rf "$RC_STAGE_DIR"
mkdir -p "$RC_STAGE_DIR"

sed \
    -e "s|@APP_VERSION_COMMA@|${COMMA_VERSION}|g" \
    -e "s|@APP_VERSION_STRING@|${VERSION}|g" \
    "$ROOT_DIR/windows/app.rc" > "$RC_STAGE_DIR/app.rc"

cp "$ROOT_DIR/windows/icon.ico" "$RC_STAGE_DIR/icon.ico"

(cd "$RC_STAGE_DIR" && windres app.rc -O coff -o "$SYSO_PATH")

rm -rf "$RC_STAGE_DIR"

export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1

go build \
    -trimpath \
    -ldflags "-H windowsgui -s -w" \
    -o "$APP_DIR/$EXE_NAME" \
    "$ROOT_DIR/main.go"

rm -f "$SYSO_PATH"

cp -R "$ROOT_DIR/web" "$APP_DIR/web"

echo "Built $APP_DIR/$EXE_NAME"
