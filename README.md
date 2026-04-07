# Blog API

A RESTful blog API built with Go, following hexagonal architecture principles. This API provides full CRUD operations for blog posts with rich content support via Editor.js, image uploads via MinIO, and PostgreSQL for data persistence.

## Features

- **Hexagonal Architecture**: Clean separation of concerns across domain, application, infrastructure, and API layers
- **Full CRUD Operations**: Create, read, update, and delete blog posts
- **Rich Content Support**: Native Editor.js integration with multiple block types
- **Image Upload**: MinIO S3-compatible object storage with public URLs
- **Full-Text Search**: PostgreSQL full-text search across posts
- **Pagination & Filtering**: Query posts with pagination, search, and author filtering
- **URL Metadata Fetching**: Extract Open Graph metadata from URLs
- **Structured Logging**: Color-coded logs for better debugging
- **Dockerized**: Complete containerized setup

## Tech Stack

- **Go 1.22.2**: Programming language
- **Gin**: Web framework
- **GORM**: ORM for PostgreSQL
- **PostgreSQL**: Database
- **MinIO**: S3-compatible object storage
- **Docker**: Containerization

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Make (optional, for convenience commands)

### Running the Application

1. **Clone and navigate to the project**
   ```bash
   cd blog-api
   ```

2. **Start all services**
   ```bash
   make run
   # or
   docker-compose up --build
   ```

   This will start:
   - PostgreSQL on port 5432
   - MinIO on port 9000 (API) and 9001 (Console)
   - Blog API on port 8080

3. **Verify services are running**
   ```bash
   curl http://localhost:8080/health
   ```

### Environment Variables

All configuration is done via environment variables. See `.env.example` for available options.

## Development

### Project Structure

```
blog-api/
├── cmd/api/                   # Application entry point
├── internal/                  # Internal packages
│   ├── domain/                # Core business logic
│   │   ├── models/            # Domain entities
│   │   └── repositories/      # Repository interfaces
│   ├── application/           # Application services
│   │   ├── dtos/              # Data transfer objects
│   │   ├── mappers/           # DTO to model mappers
│   │   └── services/          # Business logic services
│   ├── infrastructure/        # External integrations
│   │   ├── ai/                # AI integrations
│   │   ├── database/          # PostgreSQL setup
│   │   ├── repository/        # Repository implementations
│   │   ├── storage/           # MinIO client
│   │   └── logging/           # Structured logger
│   └── api/                   # HTTP layer
│       ├── handlers/          # HTTP handlers
│       ├── middleware/        # Middleware
│       └── router/            # Route configuration
├── config/                    # Configuration management
├── docker-compose.yml         # Docker services
├── Dockerfile                 # Application container
└── Makefile                   # Helper commands
```

### Available Make Commands

```bash
make run          # Start all services
make down         # Stop all services
make logs         # View logs
make clean        # Remove containers and volumes
make build        # Build Go binary locally
make dev          # Run locally (requires postgres & minio running)
```

### Running Tests

```bash
make test
# or
go test ./... -v
```

## API Documentation (Swagger)

The API is documented with Swagger UI, available at `http://localhost:8080/swagger/index.html` when the server is running.

### Accessing the UI

Start the server (`make run` or `make dev`) and open:
```
http://localhost:8080/swagger/index.html
```

### Regenerating the docs

Run this after adding or changing any endpoint annotation:
```bash
swag init -g cmd/api/main.go
```

If `swag` is not in your `PATH`, use the full path:
```bash
~/go/bin/swag init -g cmd/api/main.go
```

The command regenerates `docs/docs.go`, `docs/swagger.json`, and `docs/swagger.yaml`.

## MinIO Console

Access the MinIO console at http://localhost:9001

- Username: `minioadmin`
- Password: `minioadmin`

## CORS Configuration

By default, the API allows requests from:
- `http://localhost:3000`
- `http://localhost:5173`

Update `CORS_ORIGINS` in `.env` to add more origins.

## Logging

The application uses structured, color-coded logging:

- **INFO** (Green): Normal operations
- **WARN** (Yellow): Warnings and validation failures
- **ERROR** (Red): Errors and failures
- **DEBUG** (Cyan): Debug information

Example output:
```
[2024-01-25 21:00:00] [INFO] Server listening | port=8080
[2024-01-25 21:00:05] [INFO] [POST /api/posts] | method=POST path=/api/posts status=201 duration=45ms
```

## Production Considerations

- Update MinIO credentials in production
- Use proper database credentials
- Configure HTTPS/TLS
- Set up proper CORS origins
- Add authentication/authorization
- Implement rate limiting
- Set up monitoring and alerting

## License

MIT
