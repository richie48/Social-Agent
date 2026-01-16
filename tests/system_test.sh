#!/bin/bash

set -eu

echo "Running Social Agent System Test!"

# Check if Go is installed
if ! command -v go &1> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

GOPATH=$(dirname $(go env GOMOD))
BINARY="social-agent"

# Build application
echo "Building social agent..."
go build -o ${BINARY} "${GOPATH}/cmd/agent"
if [ ! -f "${BINARY}" ]; then
    echo "Error: Binary not found after build"
    exit 1
fi

# TODO: Improve this tests 
echo "Running in test mode..."
exec "${BINARY}" -test-mode 2>&1 || {
    echo "Error: Application failed to run in test mode"
    exit 1
}

echo "Cleaning up executable..."
rm -f "${BINARY}"

echo "System test completed successfully"
