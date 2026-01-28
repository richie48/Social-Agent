BINARY_NAME=social-agent
MAIN_PATH=cmd/agent/main.go

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

run: build
	@echo "Starting Social Agent..."
	./$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "Clean up complete"

test: build
	@echo "Running system tests..."
	./tests/system_test.sh ./${BINARY_NAME}

deps:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies downloaded"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

.PHONY: build run clean test deps fmt 
