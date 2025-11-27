#!/bin/bash

# Create dist directory if it doesn't exist
mkdir -p dist

# macOS (Apple Silicon)
echo "Building for macOS (Apple Silicon)..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o dist/scaffo-darwin-arm64 ./cmd

# macOS (Intel)
echo "Building for macOS (Intel)..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/scaffo-darwin-amd64 ./cmd

# Linux (x86_64)
echo "Building for Linux (x86_64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/scaffo-linux-amd64 ./cmd

# Windows (x86_64)
echo "Building for Windows (x86_64)..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/scaffo-windows-amd64.exe ./cmd

echo "Build complete. Artifacts are in dist/"
