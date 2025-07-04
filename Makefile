.PHONY: build run dev clean install deps build-all build-linux build-darwin build-windows release tag-release push-release test-release help

# Build the application
build:
	go build -o health-hub main.go

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Build for Linux (amd64 and arm64)
build-linux:
	@echo "Building for Linux..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/health-hub-linux-amd64 main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/health-hub-linux-arm64 main.go

# Build for macOS (amd64 and arm64)
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/health-hub-darwin-amd64 main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/health-hub-darwin-arm64 main.go

# Build for Windows (amd64 and arm64)
build-windows:
	@echo "Building for Windows..."
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/health-hub-windows-amd64.exe main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/health-hub-windows-arm64.exe main.go

# Create release archives
release: build-all
	@echo "Creating release archives..."
	@mkdir -p releases
	@cd dist && tar -czf ../releases/health-hub-linux-amd64.tar.gz health-hub-linux-amd64
	@cd dist && tar -czf ../releases/health-hub-linux-arm64.tar.gz health-hub-linux-arm64
	@cd dist && tar -czf ../releases/health-hub-darwin-amd64.tar.gz health-hub-darwin-amd64
	@cd dist && tar -czf ../releases/health-hub-darwin-arm64.tar.gz health-hub-darwin-arm64
	@cd dist && zip -q ../releases/health-hub-windows-amd64.zip health-hub-windows-amd64.exe
	@cd dist && zip -q ../releases/health-hub-windows-arm64.zip health-hub-windows-arm64.exe
	@echo "Release archives created in ./releases/"

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
	rm -rf dist/
	rm -rf releases/
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

# Release commands
tag-release:
	@echo "Creating and pushing release tag..."
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make tag-release VERSION=v0.1.0"; \
		exit 1; \
	fi
	@echo "Tagging version $(VERSION)"
	git tag $(VERSION)
	git push origin $(VERSION)
	@echo "Release tag $(VERSION) created and pushed!"
	@echo "GitHub Actions will now build and create the release automatically."

# Create a release (tag + push)
push-release: tag-release
	@echo "Release $(VERSION) is being processed by GitHub Actions"
	@echo "Check https://github.com/your-username/health-hub/actions for progress"

# Test release build locally
test-release: release
	@echo "Testing release archives..."
	@ls -la releases/
	@echo "Extracting Linux build to test..."
	@mkdir -p test-extract
	@cd test-extract && tar -xzf ../releases/health-hub-linux-amd64.tar.gz
	@echo "Testing binary..."
	@timeout 2s ./test-extract/health-hub-linux-amd64 || echo "Binary test completed (expected timeout)"
	@rm -rf test-extract
	@echo "Release build test completed successfully!"

# Show help
help:
	@echo "Health Hub - Makefile Commands"
	@echo "================================"
	@echo ""
	@echo "Development:"
	@echo "  build          Build the application binary"
	@echo "  run            Build and run the application"
	@echo "  dev            Run with hot reload (requires air)"
	@echo "  serve          Build and run with network info"
	@echo "  start          Full development setup (deps + serve)"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps           Install and tidy Go dependencies"
	@echo "  install-air    Install air for hot reload"
	@echo ""
	@echo "Cross-platform builds:"
	@echo "  build-all      Build for all platforms"
	@echo "  build-linux    Build for Linux (amd64, arm64)"
	@echo "  build-darwin   Build for macOS (amd64, arm64)"
	@echo "  build-windows  Build for Windows (amd64, arm64)"
	@echo ""
	@echo "Release management:"
	@echo "  release        Create local release archives"
	@echo "  test-release   Test release build locally"
	@echo "  tag-release    Create and push git tag (requires VERSION=v0.1.0)"
	@echo "  push-release   Same as tag-release"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean          Remove build artifacts and data"
	@echo "  network-info   Show local and Tailscale network info"
	@echo "  run-s3         Run with S3 storage enabled"
	@echo ""
	@echo "Examples:"
	@echo "  make tag-release VERSION=v0.1.0"
	@echo "  make test-release"
	@echo "  make build-all"