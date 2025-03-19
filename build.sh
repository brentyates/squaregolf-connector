#!/bin/bash

# Exit on error, undefined variables, and propagate pipeline errors
set -euo pipefail

# Build script for squaregolf-connector using fyne-cross
# This script builds the application for macOS (arm64, amd64) and Windows (amd64)

PACKAGE_NAME="github.com/brentyates/squaregolf-connector"
APP_ID="com.brentyates.squaregolf-connector"
VERSION="0.1.0-alpha.4"
# Create a simplified version without the alpha suffix for fyne-cross compatibility
SIMPLE_VERSION="${VERSION%%-*}"  # Removes everything after the first '-'
# Print colored output
function print_status() {
    echo -e "\e[1;34m>> $1\e[0m"
}

# Print error message and exit
function print_error() {
    echo -e "\e[1;31mERROR: $1\e[0m" >&2
    exit 1
}

# Check if fyne-cross is installed
if ! command -v fyne-cross &> /dev/null; then
    print_error "fyne-cross is not installed. Install it with: go install github.com/fyne-io/fyne-cross@latest"
fi

# Check if docker is running
if ! docker info &> /dev/null; then
    print_error "Docker is not running. Please start Docker and try again."
fi

# Create build timestamp
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

# Get git commit hash
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Prepare ldflags to inject build information
# macOS build uses single quotes within the ldflags
MACOS_LDFLAGS="-X '${PACKAGE_NAME}/internal/version.BuildTime=${BUILD_TIME}' -X '${PACKAGE_NAME}/internal/version.GitCommit=${GIT_COMMIT}'"
# Windows build uses simplified flags - just the GUI flag without version info to avoid quoting issues
WINDOWS_LDFLAGS="-H=windowsgui"

print_status "Building squaregolf-connector v${VERSION} (Git: ${GIT_COMMIT})"

# Build for macOS (arm64)
print_status "Building for macOS (arm64)..."
fyne-cross darwin -arch=arm64 \
    -app-id="${APP_ID}" \
    -app-version="${SIMPLE_VERSION}" \
    -ldflags="${MACOS_LDFLAGS}" \
    -icon="icon.png" \
    -output="squaregolf-connector" || print_error "macOS arm64 build failed" \

# Build for macOS (amd64)
print_status "Building for macOS (amd64)..."
fyne-cross darwin -arch=amd64 \
    -app-id="${APP_ID}" \
    -app-version="${SIMPLE_VERSION}" \
    -ldflags="${MACOS_LDFLAGS}" \
    -icon="icon.png" \
    -output="squaregolf-connector" || print_error "macOS amd64 build failed" \

# Build for Windows (amd64)
print_status "Building for Windows (amd64)..."
fyne-cross windows -arch=amd64 \
    -app-id="${APP_ID}" \
    -app-version="${SIMPLE_VERSION}" \
    -ldflags="${WINDOWS_LDFLAGS}" \
    -icon="icon.png" \
    -output="squaregolf-connector" || print_error "Windows amd64 build failed" \

# Create zip archives for macOS builds
print_status "Creating zip archives for macOS builds..."

# Create zip for macOS (arm64)
print_status "Zipping macOS (arm64) app bundle..."
cd fyne-cross/dist/darwin-arm64 && \
zip -r squaregolf-connector-macos-arm64.zip squaregolf-connector.app && \
cd ../../../ || print_error "Failed to create zip for macOS arm64"

# Create zip for macOS (amd64)
print_status "Zipping macOS (amd64) app bundle..."
cd fyne-cross/dist/darwin-amd64 && \
zip -r squaregolf-connector-macos-amd64.zip squaregolf-connector.app && \
cd ../../../ || print_error "Failed to create zip for macOS amd64"

print_status "Build completed successfully!"
print_status "Build artifacts can be found in the fyne-cross/dist directory"

exit 0

