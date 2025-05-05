.PHONY: build install clean run

# Default target
all: build

# Build the binary
build:
	@echo "Building coding-agent..."
	@go build -o coding-agent ./cmd/original

# Install the binary to /usr/local/bin
install: build
	@echo "Installing coding-agent to /usr/local/bin..."
	@mkdir -p /usr/local/bin
	@cp coding-agent /usr/local/bin/
	@echo "Installation complete!"

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f coding-agent
	@echo "Clean complete!"

# Run the agent
run: build
	@./coding-agent
