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

## API Endpoints

### Posts

- `GET /api/posts` - List posts with pagination and search
- `GET /api/posts/:id` - Get a single post
- `POST /api/posts` - Create a new post
- `PUT /api/posts/:id` - Update a post
- `DELETE /api/posts/:id` - Delete a post

### File Upload

- `POST /api/upload` - Upload an image (returns Editor.js compatible response)

### URL Metadata

- `GET /api/fetch-url?url=<url>` - Fetch metadata from a URL

## Example Usage

### 1. Upload an Image

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@image.jpg"
```

Response:
```json
{
  "success": 1,
  "file": {
    "url": "http://localhost:9000/blog/uploads/550e8400-e29b-41d4-a716-446655440000.jpg"
  }
}
```

### 2. Create a Post

```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "description": "This is a test post",
    "image": "http://localhost:9000/blog/uploads/550e8400-e29b-41d4-a716-446655440000.jpg",
    "author": "David",
    "content": {
      "blocks": [{
        "id": "1",
        "type": "paragraph",
        "data": {"text": "Hello World!"}
      }],
      "time": 1633046400000,
      "version": "2.28.0"
    }
  }'
```

### 3. List Posts

```bash
# List all posts
curl "http://localhost:8080/api/posts?page=1&limit=10"

# Search posts
curl "http://localhost:8080/api/posts?search=test"

# Filter by author
curl "http://localhost:8080/api/posts?author=David"
```

## Development

### Project Structure

```
blog-api/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── domain/                 # Core business logic
│   │   ├── models/            # Domain entities
│   │   └── repositories/      # Repository interfaces
│   ├── application/           # Application services
│   │   ├── dtos/             # Data transfer objects
│   │   ├── mappers/          # DTO to model mappers
│   │   └── services/         # Business logic services
│   ├── infrastructure/        # External integrations
│   │   ├── database/         # PostgreSQL setup
│   │   ├── repository/       # Repository implementations
│   │   ├── storage/          # MinIO client
│   │   └── logging/          # Structured logger
│   └── api/                   # HTTP layer
│       ├── handlers/         # HTTP handlers
│       ├── middleware/       # Middleware
│       └── router/           # Route configuration
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
