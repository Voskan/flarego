# Makefile for FlareGo project

.PHONY: build clean test lint dev proto docker help

# Default target
all: build

# Build all binaries
build: build-cli build-agent build-gateway

build-cli:
	@echo "Building flarego CLI..."
	go build -tags=cli -o bin/flarego ./cmd/flarego

build-agent:
	@echo "Building flarego-agent..."
	go build -o bin/flarego-agent ./cmd/flarego-agent

build-gateway:
	@echo "Building flarego-gateway..."
	go build -o bin/flarego-gateway ./cmd/flarego-gateway

# Development mode
dev:
	@echo "Starting development environment..."
	docker compose -f deployments/docker-compose.yaml up -d
	cd web && npm run dev

# Stop and clean up development environment
stop-dev:
	@echo "Stopping development environment..."
	docker compose -f deployments/docker-compose.yaml down

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/node_modules/.cache/

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Lint code
lint:
	@echo "Running linters..."
	@bash build/scripts/lint.sh

# Format code
fmt:
	@echo "Formatting code..."
	@bash build/scripts/lint.sh --fix

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@bash build/scripts/generate-proto.sh

# Build Docker images
docker:
	@echo "Building Docker images..."
	docker build -f build/Dockerfile.gateway -t flarego/gateway:latest .
	docker build -f build/Dockerfile.agent -t flarego/agent:latest .

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod download
	@echo "Installing web dependencies..."
	cd web && npm ci

# Setup development environment
setup: deps proto
	@echo "Setting up development environment..."
	@mkdir -p bin/

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build all binaries"
	@echo "  build-cli   - Build flarego CLI"
	@echo "  build-agent - Build flarego-agent"
	@echo "  build-gateway - Build flarego-gateway"
	@echo "  dev         - Start development environment"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linters"
	@echo "  fmt         - Format code"
	@echo "  proto       - Generate protobuf files"
	@echo "  docker      - Build Docker images"
	@echo "  deps        - Install dependencies"
	@echo "  setup       - Setup development environment"
	@echo "  help        - Show this help message"
