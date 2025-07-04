.PHONY: build run dev clean install deps

# Build the application
build:
	go build -o health-hub main.go

# Run the application
run: build
	./health-hub

# Run in development mode with auto-reload (requires 'air' tool)
dev:
	@command -v air >/dev/null 2>&1 || (echo "Installing air for hot reload..." && go install github.com/air-verse/air@latest)
	air

# Install dependencies
deps:
	go mod tidy
	go mod download

# Clean build artifacts
clean:
	rm -f health-hub
	rm -rf data/

# Install air for development
install-air:
	go install github.com/air-verse/air@latest

# Run with S3 enabled
run-s3: build
	@echo "Make sure to set S3_BUCKET environment variable"
	USE_S3=true ./health-hub

# Show network info including Tailscale
network-info:
	@echo "=== Network Information ==="
	@echo "Local interfaces:"
	@ip addr show | grep -E "(inet|tailscale)" || echo "No Tailscale interface found"
	@echo ""
	@echo "Tailscale status:"
	@tailscale status 2>/dev/null || echo "Tailscale not running or not installed"

# Run and show connection info
serve: build
	@echo "=== Health Hub Server ==="
	@echo "Starting server..."
	@echo ""
	@echo "Local access: http://localhost:8080"
	@echo ""
	@if command -v tailscale >/dev/null 2>&1; then \
		echo "Tailscale addresses:"; \
		tailscale ip -4 2>/dev/null | while read ip; do \
			echo "  http://$$ip:8080"; \
		done; \
		echo ""; \
	fi
	@echo "=== Server Starting ==="
	./health-hub

# Quick start for development
start: deps serve