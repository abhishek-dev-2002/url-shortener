# URL Shortener

A URL shortening service built in Go with PostgreSQL, clean architecture, and collision-free short code generation.

## Features

- `POST /api/v1/shorten` — shorten a URL (with optional custom alias)
- `GET /:code` — 301 redirect to the original URL
- `GET /health` — health check with DB status
- Custom aliases (3-20 alphanumeric chars)
- Duplicate URL detection (idempotent — same URL returns same short code)
- Click count tracking (async, non-blocking)
- Collision-free short codes (DB sequence + block allocation)
- Standardized API response format with request IDs

## Project Structure

```
url-shortner/
├── main.go                          # Entry point, DI wiring, graceful shutdown
├── config.json                      # Configuration
├── models/
│   └── url.go                       # Request, response, DB models
├── repo/
│   ├── repositorymanager.go         # Repo abstraction layer (services depend on this)
│   ├── repointerfaces/
│   │   └── urlrepo.go               # Repository contract (interface)
│   └── store/
│       ├── manager.go               # DB connection, pooling, migrations
│       └── postgres.go              # SQL implementation + ID block allocation
├── services/
│   ├── health.go                    # Health check handler
│   ├── middleware.go                # Request ID, request logger
│   └── urlshortener/
│       ├── router.go                # Route registration
│       ├── handler.go               # HTTP layer (read → call service → respond)
│       ├── service.go               # Business logic
│       └── shortcode.go             # Short code generator (block allocation)
├── validator/
│   └── validator.go                 # URL + alias validation
├── utils/                           # Logger, config, error, response helpers
├── .devops/Dockerfile
└── docker-compose.yml
```

## Quick Start

### Prerequisites
- Go 1.22+
- Docker & Docker Compose

### Run

```bash
# Start PostgreSQL
docker compose up -d postgres

# Run the service
go run .
```

Or run everything together:
```bash
docker compose up --build
```

Service starts at `http://localhost:8080`.

### Try it

```bash
# Shorten a URL
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/path"}'

# Shorten with custom alias
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "alias": "mysite"}'

# Redirect (follow with -L or open in browser)
curl -I http://localhost:8080/<short_code>

# Unknown code → 404
curl -I http://localhost:8080/doesnotexist
```

## API Response Format

**Success:**
```json
{
  "requestId": "550e8400-e29b-41d4-a716-446655440000",
  "payload": {
    "short_code": "0000XyZ",
    "short_url": "http://localhost:8080/0000XyZ",
    "original_url": "https://example.com/long/path"
  },
  "message": "SUCCESS"
}
```

**Error:**
```json
{
  "requestId": "550e8400-e29b-41d4-a716-446655440000",
  "error": {
    "code": "NOT_FOUND",
    "message": "short code not found"
  }
}
```

## Configuration

Loaded from `config.json`, overridden by environment variables:

| Env Variable   | Default                                                                   |
|----------------|---------------------------------------------------------------------------|
| `PORT`         | `8080`                                                                    |
| `BASE_URL`     | `http://localhost:8080`                                                   |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable` |

## Design Decisions

### Short Code Generation — DB Sequence + Block Allocation

Each instance pre-allocates a block of 10,000 sequential IDs from a PostgreSQL sequence, then serves codes from memory. When the block runs out, it fetches the next block.

**Why it won't collide:**
- The DB sequence is atomic — concurrent instances always get non-overlapping blocks
- Within a block, a mutex ensures sequential assignment
- No two instances ever share the same ID range

**Why not random?** Random codes need a DB existence check on every insert. Block allocation needs one DB call per 10,000 inserts.

**Why not Snowflake?** Snowflake produces 10-11 char codes and requires machine ID configuration. Block allocation gives fixed 7-char codes with no machine configuration — the DB sequence handles all coordination.

**Code length:** 7 characters (Base62 padded). 62^7 = 3.5 trillion possible codes.

### Duplicate URL Handling

- **Without alias:** idempotent — returns the existing short code. Saves storage, gives consistent links.
- **With alias:** always creates a new entry. One URL can have multiple custom aliases.

This is a deliberate choice documented in the code.

### Why 301 (Moved Permanently)?

- Standard for URL shorteners where mappings are permanent
- Browsers and CDNs cache the redirect — reduces server load
- Better SEO pass-through

### Architecture

```
Request → Gin Router → Middleware → Handler → Service → Repository → PostgreSQL
```

- Handlers only read request, call service, return response
- Business logic lives exclusively in the service layer
- DB operations only happen through the repository interface
- Handler depends on a local interface (not the concrete service) for testability
