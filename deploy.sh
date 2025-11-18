#!/bin/bash
# Quick deploy script - just the commands I always forget!

set -e  # Exit on error

echo "Building and installing ingredients..."

# Regenerate corpus
go generate

# Run tests
#go test ./...

# Build
go build -o ingredients cmd/ingredients/main.go

# Install
sudo cp ingredients /usr/local/bin/
sudo chmod +x /usr/local/bin/ingredients

# Cleanup
rm ingredients

echo "Done! Installed to $(which ingredients)"
