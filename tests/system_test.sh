#!/bin/bash

set -eu

if [ -z "$1" ]; then
    echo "No binary path provided. Usage: $0 /path/to/binary"
    exit 1
fi

BINARY_PATH="$1"
echo "Using binary at: ${BINARY_PATH}"

# TODO: Improve test 
echo "Running test mode..."
exec "${BINARY_PATH}" -test-mode 2>&1

echo "System test completed successfully"
