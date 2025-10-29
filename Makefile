.PHONY: build install clean test fmt vet

# Binary name
BINARY=smart-keyvault
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build the project
build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY) ./cmd/$(BINARY)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY)"

# Install the binary to $GOPATH/bin or $HOME/.local/bin
install: build
	@echo "Installing $(BINARY)..."
	@if [ -n "$(GOPATH)" ]; then \
		cp $(BUILD_DIR)/$(BINARY) $(GOPATH)/bin/; \
		echo "Installed to $(GOPATH)/bin/$(BINARY)"; \
	else \
		mkdir -p $(HOME)/.local/bin; \
		cp $(BUILD_DIR)/$(BINARY) $(HOME)/.local/bin/; \
		echo "Installed to $(HOME)/.local/bin/$(BINARY)"; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run all checks
check: fmt vet test

# Display help
help:
	@echo "Makefile commands:"
	@echo "  make build    - Build the binary"
	@echo "  make install  - Build and install the binary"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make fmt      - Format code"
	@echo "  make vet      - Run go vet"
	@echo "  make check    - Run fmt, vet, and test"
	@echo "  make help     - Show this help message"
