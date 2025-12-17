#!/bin/bash
set -e

VERSION="$1"

# Function to build pkg for an architecture
build_pkg() {
    ARCH="$1"
    FOLDER_ARCH="$2" # GoReleaser uses different naming sometimes
    
    echo "Building package for ${ARCH}..."
    
    # Define paths
    # GoReleaser output path pattern: dist/scaffo_darwin_<arch>_<version>/scaffo
    # We need to find the directory because the version part (v1, v8.0) might change
    BINARY_DIR=$(find dist -type d -name "scaffo_darwin_${FOLDER_ARCH}_*" | head -n 1)
    
    if [ -z "$BINARY_DIR" ]; then
        echo "Error: Could not find binary directory for ${ARCH}"
        return 1
    fi
    
    BINARY_PATH="${BINARY_DIR}/scaffo"
    
    if [ ! -f "$BINARY_PATH" ]; then
        echo "Error: Binary not found at ${BINARY_PATH}"
        return 1
    fi

    PKG_ROOT="dist/pkg_root_${ARCH}"
    rm -rf "${PKG_ROOT}"
    mkdir -p "${PKG_ROOT}/usr/local/bin"
    
    # Copy binary to the root structure
    cp "${BINARY_PATH}" "${PKG_ROOT}/usr/local/bin/"
    chmod 755 "${PKG_ROOT}/usr/local/bin/scaffo"

    # Create the package
    pkgbuild --root "${PKG_ROOT}" \
             --identifier "com.scaffo.cli" \
             --version "${VERSION}" \
             --install-location "/" \
             "dist/scaffo_${ARCH}.pkg"
             
    echo "Created dist/scaffo_${ARCH}.pkg"
}

# Build for AMD64 (Intel)
build_pkg "amd64" "amd64"

# Build for ARM64 (Apple Silicon)
build_pkg "arm64" "arm64"
