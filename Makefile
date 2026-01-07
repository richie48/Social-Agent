.PHONY: build run run-debug run-dry test clean help

BINARY_NAME=threads-agent
MAIN_PATH=cmd/agent/main.go
OUTPUT_PATH=./$(BINARY_NAME)

help:
	@echo "Threads Influencer Agent - Available Commands"
	@echo ""
	@echo "  make build      - Build the agent binary"
	@echo "  make run        - Build and run the agent"
	@echo "  make run-debug  - Build and run with debug logging"
	@echo "  make run-dry    - Build and run in dry-run mode"
	@echo "  make test       - Run all tests"
	@echo "  make clean      - Remove built binary"
	@echo ""

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(OUTPUT_PATH) $(MAIN_PATH)
	@echo "✓ Build complete: $(OUTPUT_PATH)"

run: build
	@echo "Starting Threads Influencer Agent..."
	./$(OUTPUT_PATH)

run-debug: build
	@echo "Starting Threads Influencer Agent (Debug Mode)..."
	./$(OUTPUT_PATH) -debug

run-dry: build
	@echo "Starting Threads Influencer Agent (Dry-Run Mode)..."
	./$(OUTPUT_PATH) -dry-run

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -f $(OUTPUT_PATH)
	go clean
	@echo "✓ Clean complete"

deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "✓ Dependencies downloaded"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Formatting complete"

lint:
	@echo "Running linter..."
	golangci-lint run ./...

vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "✓ Vet complete"
