# AGENTS.md

## Project Overview
Blog API is a RESTful backend service built with Go, Gin, PostgreSQL, and MinIO. It serves as the backend for a blog client, handling posts, image uploads, and URL metadata extraction for Editor.js integration. The project follows Clean Architecture (Hexagonal Architecture) principles.

## Tech Stack
- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Object Storage**: MinIO
- **Containerization**: Docker & Docker Compose

## Directory Structure
- `cmd/api`: Application entry point.
- `internal/api`: Interface adapters layer. Contains HTTP handlers, router setup, and middleware.
- `internal/application`: Application business rules. Contains services, DTOs, and mappers.
- `internal/domain`: Enterprise business rules. Contains core models and repository interfaces.
- `internal/infrastructure`: Frameworks and drivers. Contains database implementations, external API clients, and logging.
- `config`: Configuration management.
- `Dockerfile` & `docker-compose.yml`: Deployment configurations.

## Setup Commands
- **Start all services (Docker)**: `make run`
- **Start services in background**: `make run-detached`
- **Stop all services**: `make down`
- **Run API locally**: `make dev` (Requires running PostgreSQL and MinIO instances)
- **Build binary**: `make build`
- **Run tests**: `make test`
- **View logs**: `make logs` or `make logs-api`

## Architectural Rules
1. **Dependency Rule**: Dependencies must point inward. 
   - `domain` depends on *nothing*.
   - `application` depends on `domain`.
   - `api` and `infrastructure` depend on `application` and `domain`.
2. **Interfaces**: Define repository interfaces in `internal/domain/repositories`. Implement them in `internal/infrastructure`.
3. **Dependency Injection**: Inject dependencies (repositories, services) via constructor functions.
4. **DTOs**: Use DTOs in `internal/application/dtos` for data transfer between `api` and `application` layers. Do not expose domain models directly in API responses.
5. **Mappers**: Use mappers in `internal/application/mappers` to transform between DTOs and Domain Models.

## Code Style & Standards
- **Formatting**: Standard Go formatting (`gofmt`).
- **Error Handling**: 
  - Return errors explicitly; do not panic.
  - Wrap errors with context (e.g., `fmt.Errorf("failed to create post: %w", err)`).
- **Logging**: Use the structured logger provided in `internal/infrastructure/logging`. Log "why" something happened, not just "what".
- **Comments**: Focus on *why* a decision was made, avoiding obvious implementation details.
- **Naming**: 
  - Use `CamelCase` for struct and function names.
  - Use `snake_case` for database columns (GORM default).

## Contribution Rules
- **Commit Messages**: Follow Conventional Commits standard (e.g., `feat(api): add new endpoint`, `fix(db): resolve connection issue`).
- **Tests**: Write unit tests for services and handlers.
- **Documentation**: Update `API_SPECIFICATION.md` if API contracts change.

## API Specification
- **Base URL**: `/api`
- **Response Format**: JSON
- **Error Format**: Standardized JSON error response.
- **Content**: Supports Editor.js block format.

## Database
- **Migration**: Auto-migration is enabled via GORM in `internal/infrastructure/database/postgres.go`.
- **Indexing**: Indexes are explicitly defined for performance (e.g., proper indexes on `date`, `author`, and full-text search).
