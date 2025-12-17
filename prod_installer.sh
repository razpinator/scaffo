#!/bin/bash
# Production Installer Script
# This script installs the appropriate .pkg file from the dist/ folder based on your system architecture.
# It assumes you have already run ./builder_prod.sh to generate the packages.

set -e

ARCH=$(uname -m)
PKG_FILE=""

echo "Detecting system architecture..."
if [ "$ARCH" == "arm64" ]; then
    echo "Detected Apple Silicon (arm64)"
    PKG_FILE="dist/scaffo_arm64.pkg"
elif [ "$ARCH" == "x86_64" ]; then
    echo "Detected Intel Mac (amd64)"
    PKG_FILE="dist/scaffo_amd64.pkg"
else
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
fi

if [ ! -f "$PKG_FILE" ]; then
    echo "Error: Package file not found at $PKG_FILE"
    echo "Please run ./builder_prod.sh first to generate the installers."
    exit 1
fi

echo "Installing $PKG_FILE..."
echo "This requires sudo privileges."

sudo installer -pkg "$PKG_FILE" -target /

echo "---------------------------------------------------"
echo "Installation Complete!"
echo "You can now run 'scaffo' from your terminal."
echo "Version: $(scaffo --version)"
echo "---------------------------------------------------"
