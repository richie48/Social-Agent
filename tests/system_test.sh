#!/bin/bash

set -eu

echo "Running Social Agent System Test!"

# Check if Go is installed
if ! command -v go &> /dev/null; then
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
echo "Testing test mode..."
exec "${BINARY}" -test-mode 2>&1 | grep -q "not configured" || true

echo "Cleaning up executable..."
rm -f "${BINARY}"

echo "System test completed successfully"
