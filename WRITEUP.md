# WRITE-UP — URL Shortener & Link Analytics

## 1. What did I ask the AI to do, and what did I write or decide myself?

I used AI primarily as an implementation assistant to speed up repetitive coding tasks and boilerplate generation. This included generating the initial project scaffolding, router and middleware setup, Docker configuration, configuration loading, repository CRUD boilerplate, validator scaffolding, and the initial README structure. Every AI-generated change was reviewed, modified where necessary, and tested before being accepted.

The architecture and key design decisions were my own:

* **Architecture:** I chose a layered architecture (Handler → Service → Repository → Store) inspired by the production Go services I work with. The initial AI suggestion followed Go's `internal/` convention, but I reorganized the project to better match the architecture I wanted.
* **Short code generation:** I evaluated multiple approaches and selected **database sequence + block allocation** after considering scalability, operational simplicity, and URL length.
* **Duplicate URL behavior:** I decided that shortening the same URL (without a custom alias) should be idempotent and return the existing short URL rather than creating duplicates.
* **API response format:** I standardized all responses using a consistent `{requestId, payload, message}` / `{requestId, error}` format for easier debugging and tracing.
* **Database ownership:** I moved database lifecycle management into the store layer so `main.go` only performs application orchestration and business services never depend directly on `*sql.DB`.
* **Route design:** I separated API endpoints (`POST /api/v1/shorten`) from public redirect URLs (`GET /{code}`) to match how production URL shorteners are typically exposed.

Throughout the project, I treated AI as an implementation accelerator rather than an architectural decision maker.

---

## 2. Where did I override, correct, or throw away the AI's output — and why?

### Short code generation (replaced three times)

I evaluated three different approaches before selecting the final implementation:

* **Random Base62:** The initial implementation generated random codes, but every insert required checking the database for collisions. While simple, it introduces unnecessary database lookups as the system grows.
* **Snowflake-based IDs:** I then explored a Snowflake-style generator. It solves distributed ID generation well, but produces longer URLs (typically 10–11 Base62 characters) and requires machine/node configuration, which felt unnecessary for this problem.
* **Final approach — Database sequence + block allocation:** I chose this because it produces shorter URLs (around 7 characters), requires no machine configuration, guarantees uniqueness, and reduces database contention by allocating IDs in configurable blocks (10,000 IDs by default).

### Project structure

I refactored several AI-generated structural decisions to better match the architecture:

* The initial version contained a top-level `interfaces/` package with a single interface. I removed it and defined interfaces where they are consumed, following common Go practices.
* The AI also generated a `TenantRepository` that simply forwarded every call to the store without adding behavior. Since this project is not multi-tenant, I removed that layer and had the service depend directly on the repository interface.

### Repository ownership

Initially the AI passed `*sql.DB` directly from `main.go` into the application. I restructured the code so the store layer owns the database connection, while the `RepositoryManager` provides repositories to the service layer. This keeps infrastructure concerns separate from business logic.

### Cleanup

During the final review I removed several unused AI-generated artifacts, including unused helper functions, response models, context keys, imports, and an unused `repoerr` package.

---

## 3. The two or three biggest trade-offs I made

### Trade-off 1: Block allocation vs Snowflake vs Random IDs

I chose **database sequence + block allocation**.

Compared to the alternatives:

* **Random IDs:** Simpler implementation but requires collision checks on every insert.
* **Snowflake IDs:** Excellent for distributed systems but produces longer URLs and requires machine configuration.
* **Block allocation:** Produces shorter URLs, guarantees uniqueness, minimizes database round-trips (one allocation per 10,000 IDs by default), and keeps the implementation operationally simple.

The trade-off is that new ID blocks cannot be allocated if PostgreSQL is unavailable. Since the application already depends on PostgreSQL for persistence, this does not introduce an additional dependency or failure mode.

---

### Trade-off 2: Idempotent shortening vs always generating new short URLs

I chose to make shortening **idempotent** when no custom alias is provided.

If the same URL is shortened multiple times, the existing short URL is returned instead of creating duplicates.

This improves user experience, avoids unnecessary storage growth, and allows clients to safely retry requests after transient failures. The alternative—always generating a new short URL—would simplify the implementation slightly but create duplicate records for identical URLs.

---

### Trade-off 3: 301 vs 302 redirect

I chose **301 (Moved Permanently)** because it allows browsers and CDNs to cache redirects, reducing repeated requests to the service and improving redirect performance.

The downside is that cached redirects reduce the accuracy of repeat-visit analytics because subsequent requests may never reach the server. If analytics accuracy were the primary requirement, I would instead choose **302** or **307** redirects.

---

## 4. What's missing, or what I'd do with another day

With additional time, I would focus on production-oriented improvements rather than adding new features:

* **Rate limiting:** Add per-IP rate limiting to prevent abuse and automated URL creation.
* **Redis caching:** Cache frequently accessed short URLs to reduce database lookups during redirects.
* **URL expiration:** Support optional expiration dates and background cleanup of expired links.
* **Analytics API:** Expose click statistics through endpoints such as `GET /api/v1/stats/{code}`.
* **Integration tests:** Add end-to-end tests using Testcontainers with a real PostgreSQL instance to validate block allocation and repository behavior.
* **Observability:** Add Prometheus metrics and distributed tracing alongside the existing structured logging.
* **Security improvements:** Add protection against redirecting to known malicious or phishing domains.
* **Custom alias concurrency:** Improve handling of concurrent requests attempting to create the same custom alias with clearer conflict reporting.

Overall, my goal was to use AI to accelerate implementation while ensuring that the architecture, design decisions, and trade-offs remained my own. I intentionally kept the solution simple and appropriate for the assessment instead of introducing enterprise patterns that would add complexity without providing meaningful value.
