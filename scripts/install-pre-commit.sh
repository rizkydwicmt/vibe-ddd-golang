#!/bin/bash

# Install pre-commit hooks script
set -e

echo "Installing pre-commit hooks..."

# Check if pre-commit is available
if ! command -v pre-commit &> /dev/null; then
    echo "pre-commit not found. Installing..."
    
    # Try to install via pip
    if command -v pip3 &> /dev/null; then
        pip3 install pre-commit
    elif command -v pip &> /dev/null; then
        pip install pre-commit
    elif command -v brew &> /dev/null; then
        brew install pre-commit
    else
        echo "Error: Could not install pre-commit. Please install it manually:"
        echo "  pip install pre-commit"
        echo "  or visit: https://pre-commit.com/#installation"
        exit 1
    fi
fi

# Install the git hook scripts
pre-commit install

# Install required Go tools
echo "Installing Go development tools..."
# Install golangci-lint if not available
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
else
    echo "golangci-lint already installed"
fi

echo "Pre-commit hooks installed successfully!"
echo "To run all hooks manually: pre-commit run --all-files"
echo "To skip hooks during commit: git commit --no-verify"