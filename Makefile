BINARY_NAME=social-agent
MAIN_PATH=cmd/agent/main.go
OUTPUT_PATH=./$(BINARY_NAME)

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(OUTPUT_PATH) $(MAIN_PATH)
	@echo "Build complete: $(OUTPUT_PATH)"

run: build
	@echo "Starting Social Agent..."
	./$(OUTPUT_PATH)

clean:
	@echo "Cleaning up..."
	rm -f $(OUTPUT_PATH)
	go clean
	@echo "Clean complete"

test:
	@echo "Running system tests..."
	./tests/system_test.sh

deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

.PHONY: build run clean test deps fmt 
