.PHONY: help build run test clean docker-up docker-down docker-logs docker-build tidy

help:
	@echo "Available commands:"
	@echo "  make build         - Build the application via Docker"
	@echo "  make run           - Run the application via Docker Compose"
	@echo "  make test          - Run tests via Docker"
	@echo "  make tidy          - Run go mod tidy via Docker"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start Docker containers"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make docker-logs   - View Docker logs"

build:
	docker run --rm -v $$(pwd):/app -w /app golang:1.21-alpine go build -o bin/api ./cmd/api

run:
	docker compose up

test:
	@echo "Starting database for tests..."
	@docker compose up -d db
	@sleep 3
	@echo "Running tests..."
	docker run --rm -v $$(pwd):/app -w /app --network api-backend_default \
		-e DATABASE_URL=postgresql://apiuser:apipassword@db:5432/apidb?sslmode=disable \
		golang:1.21-alpine go test -v ./...

tidy:
	docker run --rm -v $$(pwd):/app -w /app golang:1.21-alpine go mod tidy

clean:
	rm -rf bin/

docker-build:
	docker build -t api-backend .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
