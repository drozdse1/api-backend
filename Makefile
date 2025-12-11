.PHONY: help build run test clean docker-up docker-down docker-logs dev

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application locally"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-up   - Start Docker containers"
	@echo "  make docker-down - Stop Docker containers"
	@echo "  make docker-logs - View Docker logs"
	@echo "  make dev         - Start local development with hot reload"

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

dev:
	@echo "Install 'air' for hot reload: go install github.com/air-verse/air@latest"
	air
