#!/bin/bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="$ROOT_DIR/dist"
VERSION="${1:-}"

if [[ -z "$VERSION" ]]; then
    echo "Usage: $0 <version>" >&2
    exit 1
fi

"$ROOT_DIR/scripts/build-macos-app.sh"

mkdir -p "$DIST_DIR"

ARCHIVE_NAME="SquareGolf-Connector-${VERSION}-macOS.zip"
ARCHIVE_PATH="$DIST_DIR/$ARCHIVE_NAME"

rm -f "$ARCHIVE_PATH"
ditto -c -k --sequesterRsrc --keepParent \
    "$ROOT_DIR/build/SquareGolf Connector.app" \
    "$ARCHIVE_PATH"

echo "$ARCHIVE_PATH"
