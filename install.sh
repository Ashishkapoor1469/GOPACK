#!/bin/bash

# GoPack Installer for macOS/Linux

set -e

echo -e "\033[36mInstalling GoPack CLI...\033[0m"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "\033[31mError: Go is not installed. Please install Go from https://golang.org/dl/\033[0m"
    exit 1
fi

# Compile the binary
echo -e "\033[33mBuilding GoPack executable...\033[0m"
go build -o gp main.go

# Ensure go bin exists
GOBIN="${HOME}/go/bin"
mkdir -p "$GOBIN"

# Move binary
mv gp "$GOBIN/gp"

echo -e "\033[32mInstalled gp to $GOBIN/gp\033[0m"

# Check if in PATH
if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
    echo -e "\033[33mWarning: $GOBIN is not in your PATH. You might want to add it to your shell config (.bashrc or .zshrc):\033[0m"
    echo "export PATH=\$PATH:\$HOME/go/bin"
else
    echo -e "\033[32mGoPack is ready! Run 'gp' to start.\033[0m"
fi
