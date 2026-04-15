# JujuDB Makefile

.PHONY: build build-cli run dev dev-stop clean migrate logs lint fmt deps help


# Build the server
build:
	go build -o bin/jujudb-server ./cmd/server

# Build the CLI
build-cli:
	go build -o bin/jujudb ./cmd/cli

# Run the application locally
run:
	go run .

# Start development environment
dev:
	docker compose up -d --build

# Stop development environment
dev-stop:
	docker compose down

# Clean up
clean:
	docker compose down -v
	docker system prune -f

# Database migration (apply schema changes to running container)
migrate:
	docker exec jujudb-postgres psql -U jujudb -d jujudb -f /docker-entrypoint-initdb.d/init.sql

# View logs
logs:
	docker compose logs jujudb

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Help
help:
	@echo "Available commands:"

	@echo "  build          - Build the server"
	@echo "  build-cli      - Build the CLI"
	@echo "  run            - Run the application locally"
	@echo "  dev            - Start development environment"
	@echo "  dev-stop       - Stop development environment"
	@echo "  clean          - Clean up Docker containers and volumes"
	@echo "  migrate        - Apply database migrations"
	@echo "  logs           - View application logs"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  deps           - Install dependencies"
