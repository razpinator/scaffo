#!/bin/bash
# Local Installer Script
# This script builds the binary locally and installs it to /usr/local/bin
# Useful for development or quick local installation without creating a .pkg

set -e

BINARY_NAME="scaffo"
INSTALL_DIR="/usr/local/bin"

echo "Building ${BINARY_NAME}..."
go build -o ${BINARY_NAME} ./cmd

echo "Installing to ${INSTALL_DIR}..."
# Check if we have write permission, otherwise use sudo
if [ -w "${INSTALL_DIR}" ]; then
    mv ${BINARY_NAME} "${INSTALL_DIR}/"
else
    echo "Sudo permission required to move binary to ${INSTALL_DIR}"
    sudo mv ${BINARY_NAME} "${INSTALL_DIR}/"
fi

echo "---------------------------------------------------"
echo "Installation Complete!"
echo "You can now run '${BINARY_NAME}' from anywhere."
echo "Version: $(${BINARY_NAME} --version)"
echo "---------------------------------------------------"
