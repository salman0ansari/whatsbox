.PHONY: build build-frontend build-server dev clean docker docker-up docker-down help

# Default target
all: build

# Build everything (frontend + server)
build: build-frontend build-server

# Build frontend
build-frontend:
	@echo "Building frontend..."
	cd web && npm ci && npm run build
	@echo "Copying frontend to embed location..."
	rm -rf internal/frontend/dist/*
	cp -r web/dist/* internal/frontend/dist/

# Build Go server (requires frontend to be built first)
build-server:
	@echo "Building server..."
	CGO_ENABLED=1 go build -o whatsbox ./cmd/server

# Development mode - run frontend and backend separately
dev:
	@echo "Starting development mode..."
	@echo "Run 'cd web && npm run dev' in another terminal for frontend"
	@echo "Starting backend..."
	go run ./cmd/server

# Run frontend dev server
dev-frontend:
	cd web && npm run dev

# Run backend only (uses embedded frontend if available)
dev-server:
	go run ./cmd/server

# Clean build artifacts
clean:
	rm -f whatsbox
	rm -rf web/dist
	rm -rf web/node_modules
	find internal/frontend/dist -type f ! -name '.gitkeep' -delete

# Build Docker image
docker:
	docker build -t whatsbox .

# Start with Docker Compose
docker-up:
	docker compose up -d

# Stop Docker Compose
docker-down:
	docker compose down

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	cd web && npm ci

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build frontend and server"
	@echo "  build-frontend - Build only the frontend"
	@echo "  build-server   - Build only the Go server"
	@echo "  dev            - Run backend in dev mode"
	@echo "  dev-frontend   - Run frontend dev server"
	@echo "  dev-server     - Run backend only"
	@echo "  clean          - Remove build artifacts"
	@echo "  docker         - Build Docker image"
	@echo "  docker-up      - Start with Docker Compose"
	@echo "  docker-down    - Stop Docker Compose"
	@echo "  test           - Run tests"
	@echo "  lint           - Run linter"
	@echo "  deps           - Install dependencies"
	@echo "  help           - Show this help"
