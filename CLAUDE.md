# API Backend Project Specifications

## Docker First Citizen Approach

This project follows a **Docker-first architecture** where Docker is the primary and required tool for all development operations.

### Core Principles

1. **No Local Go Installation Required**
   - All Go operations (build, test, run, tidy) execute inside Docker containers
   - Ensures consistent development environment across all platforms (Linux, macOS, Windows)
   - Eliminates "works on my machine" problems

2. **Platform Independence**
   - All commands must work identically on any platform with Docker installed
   - Build, test, and compilation processes are platform-agnostic
   - Developers can use any OS without additional setup

3. **Docker-Based Workflows**
   - Building: `make build` uses Docker container with Go Alpine image
   - Testing: `make test` spins up containerized database and runs tests in Go container
   - Development: `make run` or `make docker-up` uses docker-compose
   - Dependencies: `make tidy` runs go mod tidy in Docker container

4. **Prerequisites**
   - Docker and Docker Compose only
   - No Go installation needed
   - No PostgreSQL installation needed (uses PostGIS container)

5. **Benefits**
   - Consistent builds across all environments
   - Easy onboarding for new developers
   - Reproducible test environments
   - Simplified CI/CD pipeline integration