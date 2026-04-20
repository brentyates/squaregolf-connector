#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>" >&2
    exit 1
fi

"$ROOT_DIR/scripts/build-windows-app.sh" "$VERSION"

mkdir -p "$DIST_DIR"

ARCHIVE_NAME="SquareGolf-Connector-${VERSION}-Windows.zip"
ARCHIVE_PATH="$DIST_DIR/$ARCHIVE_NAME"
APP_DIR="$ROOT_DIR/build/SquareGolf Connector"

rm -f "$ARCHIVE_PATH"

if command -v zip >/dev/null 2>&1; then
    (cd "$ROOT_DIR/build" && zip -r "$ARCHIVE_PATH" "SquareGolf Connector" >/dev/null)
elif command -v powershell.exe >/dev/null 2>&1; then
    powershell.exe -NoProfile -Command \
        "Compress-Archive -Path \"$APP_DIR\" -DestinationPath \"$ARCHIVE_PATH\" -Force" >/dev/null
else
    echo "Neither zip nor powershell.exe is available to create the archive" >&2
    exit 1
fi

echo "$ARCHIVE_PATH"
