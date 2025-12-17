#!/bin/bash
# Production builder using GoReleaser
# This script generates optimized binaries and macOS .pkg installers in the dist/ folder.

# Ensure the script stops on errors
set -e

echo "Cleaning previous builds..."
rm -rf dist/

echo "Running GoReleaser (Snapshot Mode)..."
# We use --snapshot so we don't need to tag a git release every time we test the build.
# Remove --snapshot if you want to publish a real release based on the current git tag.
goreleaser release --snapshot --clean

echo "---------------------------------------------------"
echo "Build Complete!"
echo "Artifacts are located in the 'dist/' folder:"
echo " - Apple Silicon Installer: dist/scaffo_arm64.pkg"
echo " - Intel Mac Installer:     dist/scaffo_amd64.pkg"
echo "---------------------------------------------------"