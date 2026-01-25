#!/bin/sh
set -e

# Configuration
REPO="nanocluster/nanoctl" # Replace with your actual GitHub repository user/repo
BINARY_NAME="nanoctl"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

# Check for root
if [ "$(id -u)" -ne 0 ]; then
    log_error "This script must be run as root (sudo)."
    exit 1
fi

# Detect Architecture
ARCH=$(uname -m)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

case $ARCH in
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        log_error "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

log_info "Detected system: $OS/$ARCH"

# Determine latest version
log_info "Fetching latest version..."
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_TAG" ]; then
    log_error "Failed to fetch latest version. Please check the repository URL or internet connection."
    exit 1
fi

log_info "Latest version is: $LATEST_TAG"

# Construct download URL
# Assuming naming convention: nanoctl-linux-arm64
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/${BINARY_NAME}-${OS}-${ARCH}"

log_info "Downloading from: $DOWNLOAD_URL"

# Download binary
TMP_FILE="/tmp/${BINARY_NAME}_install"
if curl -L -o "$TMP_FILE" "$DOWNLOAD_URL"; then
    log_success "Download complete."
else
    log_error "Download failed."
    exit 1
fi

# Install binary
log_info "Installing to $INSTALL_DIR/$BINARY_NAME..."
mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Verify installation
if ! command -v "$BINARY_NAME" >/dev/null; then
    log_error "Installation failed: $BINARY_NAME not found in path."
    exit 1
fi

INSTALLED_VERSION=$($BINARY_NAME --version 2>/dev/null || echo "unknown")
log_success "Installed $BINARY_NAME (version: $INSTALLED_VERSION)"

# Run service installation
log_info "Setting up systemd service..."
if $BINARY_NAME install-service; then
    log_success "Service installed and started successfully!"
else
    log_error "Failed to install service."
    exit 1
fi

echo ""
echo "Installation finished! You can check the service status with:"
echo "  systemctl status nanoctl-fan"
