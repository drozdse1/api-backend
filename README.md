# API Backend

Go-based REST API backend for iOS app, designed for Google Cloud Platform with PostgreSQL database support.

## Features

- RESTful API with Gin framework
- PostgreSQL database with PostGIS extension
- Geospatial location tracking and proximity search
- Docker support for local development
- CORS middleware configured for mobile apps
- Health check endpoint
- User management with CRUD operations
- Real-time user location updates
- Efficient spatial queries for nearby user search
- Hot reload support with Air
- Google Cloud Platform ready

## Project Structure

```
api-backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   ├── database.go          # Database connection
│   │   └── migrations.go        # Schema migrations (PostGIS + tables)
│   ├── handlers/
│   │   ├── health.go            # Health check handler
│   │   ├── user.go              # User CRUD handlers
│   │   ├── radar.go             # Location tracking handlers
│   │   └── radar_test.go        # Radar integration tests
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware
│   │   └── logger.go            # Request logging
│   └── models/
│       └── user.go              # Data models (User, UserRadar, etc.)
├── pkg/
│   └── config/
│       └── config.go            # Configuration management
├── scripts/                     # Utility scripts
├── .air.toml                    # Hot reload configuration
├── .env.example                 # Environment variables template
├── docker-compose.yml           # Local development setup
├── Dockerfile                   # Container image
├── go.mod                       # Go dependencies
└── Makefile                     # Common commands
```

## Prerequisites

- Docker and Docker Compose

**Note:** This project uses Docker for all Go operations. No local Go installation is required.

## Getting Started

### Development with Docker

1. Clone and navigate to the project:
   ```bash
   cd api-backend
   ```

2. Copy environment variables (optional, docker-compose has defaults):
   ```bash
   cp .env.example .env
   ```

3. Start all services (database + API):
   ```bash
   make docker-up
   ```

   The API will be available at `http://localhost:8080`

4. View logs:
   ```bash
   make docker-logs
   ```

5. Stop services:
   ```bash
   make docker-down
   ```

### Available Make Commands

All commands use Docker internally:

```bash
make help          # Show available commands
make build         # Build the application binary via Docker
make run           # Run the application via Docker Compose
make test          # Run tests via Docker (starts database automatically)
make tidy          # Run go mod tidy via Docker
make docker-build  # Build Docker image
make docker-up     # Start Docker containers in background
make docker-down   # Stop and remove Docker containers
make docker-logs   # View container logs
make clean         # Remove build artifacts
```

## API Endpoints

### Health Check
```
GET /api/v1/health
```

### Users
```
POST   /api/v1/users       # Create user
GET    /api/v1/users       # List all users
GET    /api/v1/users/:id   # Get user by ID
DELETE /api/v1/users/:id   # Delete user
```

### Radar (Geospatial Location Tracking)
```
POST   /api/v1/radar/location   # Update user location
GET    /api/v1/radar/nearby     # Find nearby active users
```

### Example Requests

Create a user:
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com"}'
```

Get all users:
```bash
curl http://localhost:8080/api/v1/users
```

Update user location:
```bash
curl -X POST http://localhost:8080/api/v1/radar/location \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "latitude": 52.5200,
    "longitude": 13.4050,
    "is_active": true
  }'
```

Find nearby users (within 10km radius):
```bash
curl "http://localhost:8080/api/v1/radar/nearby?latitude=52.5200&longitude=13.4050&radius=10"
```

Response example:
```json
{
  "count": 2,
  "users": [
    {
      "user_id": 1,
      "email": "user1@example.com",
      "latitude": 52.5200,
      "longitude": 13.4050,
      "distance_km": 0.5,
      "last_update_at": "2024-01-15T10:30:00Z"
    },
    {
      "user_id": 2,
      "email": "user2@example.com",
      "latitude": 52.5250,
      "longitude": 13.4100,
      "distance_km": 0.8,
      "last_update_at": "2024-01-15T10:25:00Z"
    }
  ]
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Server port | 8080 |
| ENVIRONMENT | Environment (development/production) | development |
| DATABASE_URL | PostgreSQL connection string | - |
| ALLOWED_ORIGIN | CORS allowed origin | http://localhost:3000 |

## Deployment to Google Cloud Platform

### Using Cloud Run

1. Build and push container:
   ```bash
   gcloud builds submit --tag gcr.io/PROJECT_ID/api-backend
   ```

2. Deploy to Cloud Run:
   ```bash
   gcloud run deploy api-backend \
     --image gcr.io/PROJECT_ID/api-backend \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars DATABASE_URL="postgresql://..." \
     --set-env-vars ENVIRONMENT=production
   ```

### Using Cloud SQL

1. Create a Cloud SQL instance:
   ```bash
   gcloud sql instances create api-db \
     --database-version=POSTGRES_15 \
     --tier=db-f1-micro \
     --region=us-central1
   ```

2. Create database:
   ```bash
   gcloud sql databases create apidb --instance=api-db
   ```

3. Connect your Cloud Run service to Cloud SQL using the connection name

## Database Migrations

Migrations run automatically on application startup. To add new migrations, edit `internal/database/migrations.go`.

## Testing

Run all tests via Docker (database starts automatically):
```bash
make test
```

All tests run in a containerized Go environment with a PostGIS database.

## Building

Build the application binary via Docker:
```bash
make build
```

The binary will be in `./bin/api`

Build the Docker image:
```bash
make docker-build
```

## Adding New Endpoints

1. Create a new handler in `internal/handlers/`
2. Define routes in `cmd/api/main.go`
3. Add models in `internal/models/` if needed
4. Update database schema in `internal/database/migrations.go`

## Security Notes

- Never commit `.env` file
- Use environment variables for sensitive data
- Enable SSL/TLS in production
- Use Cloud SQL Proxy for secure database connections
- Implement authentication/authorization for production use

## Next Steps

- [ ] Add authentication (JWT/OAuth)
- [ ] Implement rate limiting
- [ ] Add comprehensive tests
- [ ] Set up CI/CD pipeline
- [ ] Add API documentation (Swagger)
- [ ] Implement request validation
- [ ] Add database connection pooling tuning
- [ ] Set up monitoring and logging (Cloud Logging)

## License

MIT
