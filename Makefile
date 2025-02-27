.PHONY: build run clean

# Binary name
BINARY_NAME=lexin-downloader

# Build directory
BUILD_DIR=./build

# Output directory for dictionaries
OUTPUT_DIR=./lexin_downloads

# Default concurrency
CONCURRENCY=3

# Main build target
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/lexin

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME) -out $(OUTPUT_DIR) -concurrency $(CONCURRENCY)

# Clean build artifacts
clean:
	@echo "Cleaning build directory..."
	@rm -rf $(BUILD_DIR)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go get -u github.com/charmbracelet/bubbles
	@go get -u github.com/charmbracelet/bubbletea
	@go get -u github.com/charmbracelet/lipgloss
	@go get -u github.com/dustin/go-humanize
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 ./cmd/lexin
	
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 ./cmd/lexin
	
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe ./cmd/lexin

# Help message
help:
	@echo "Available targets:"
	@echo "  build       - Build the application"
	@echo "  run         - Build and run the application"
	@echo "  clean       - Remove build artifacts"
	@echo "  deps        - Install dependencies"
	@echo "  fmt         - Format code"
	@echo "  test        - Run tests"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Configuration:"
	@echo "  BINARY_NAME  = $(BINARY_NAME)"
	@echo "  BUILD_DIR    = $(BUILD_DIR)"
	@echo "  OUTPUT_DIR   = $(OUTPUT_DIR)"
	@echo "  CONCURRENCY  = $(CONCURRENCY)"