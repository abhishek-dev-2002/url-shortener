# URL Shortener

A URL shortening service built in Go with PostgreSQL, clean layered architecture, and collision-free short code generation via DB sequence block allocation.

## Architecture

```
                    ┌─────────────────────────┐
                    │       HTTP Client        │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │    Gorilla Mux Router    │
                    │  (middleware: recovery,  │
                    │   requestID, logging)    │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │     Service Layer        │
                    │  (business logic only)   │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │    Repository Layer      │
                    │  (interface-based DI)    │
                    └────────────┬────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │       PostgreSQL         │
                    └─────────────────────────┘
```

## API Endpoints

| Method | Path              | Description                        |
|--------|-------------------|------------------------------------|
| POST   | /api/v1/shorten   | Shorten a URL (optional alias)     |
| GET    | /{code}           | 301 redirect to original URL       |
| GET    | /health           | Health check with DB status        |
| GET    | /health/check     | Simple liveness check              |

## Project Structure

```
url-shortner/
├── main.go                              # Entry point, DI wiring, graceful shutdown
├── router.go                            # Route registration
├── config.json                          # Configuration
├── models/                              # Shared data models
├── repo/
│   ├── repositorymanager.go             # DB lifecycle + repo access
│   ├── repointerfaces/                  # Repository contracts
│   └── store/                           # PostgreSQL implementation
├── services/
│   ├── health.go                        # Health check
│   ├── middleware.go                    # Recovery, request ID, logging, decoder
│   └── urlshortener/                    # URL shortening domain
│       ├── router.go                    # Feature route setup
│       ├── service.go                   # Business logic
│       ├── shortcode.go                 # Block allocation code generator
│       ├── model.go                     # Internal DTOs
│       ├── mapper.go                    # Model conversions
│       └── service_test.go             # Unit tests
├── validator/                           # Input validation
├── utils/                               # Config, logger, error, response
├── .devops/Dockerfile
└── docker-compose.yml
```

## Quick Start

```bash
# Start PostgreSQL
docker compose up -d postgres

# Run the service
go run .

# Or run everything
docker compose up --build
```

Service runs at `http://localhost:8080`.

## Usage

```bash
# Shorten a URL
curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/path"}' | jq

# Shorten with custom alias
curl -s -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "alias": "mysite"}' | jq

# Redirect (returns 301)
curl -I http://localhost:8080/0000001

# Unknown code (returns 404)
curl -s http://localhost:8080/doesnotexist | jq
```

## Running Tests

```bash
go test ./... -v
```

## Configuration

Loaded from `config.json`, overridden by env vars:

| Env Variable   | Default                                                                   |
|----------------|---------------------------------------------------------------------------|
| `PORT`         | `8080`                                                                    |
| `BASE_URL`     | `http://localhost:8080`                                                   |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable` |

## Design Decisions

**Short Code Generation — DB Sequence + Block Allocation**

Each instance allocates 10,000 IDs from a PostgreSQL sequence, serves from memory, then fetches next block. Zero collisions by design. Multiple instances get non-overlapping blocks. One DB call per 10K URLs.

**Duplicate URL Policy**

Same URL without alias → returns existing code (idempotent). With alias → creates new mapping.

**301 Redirect**

Permanent redirect. Browsers/CDNs cache it, reducing server load.
