WRITE-UP — URL Shortener & Link Analytics
==========================================


1. What did I ask the AI to do, and what did I write or decide myself?
----------------------------------------------------------------------

I used AI (Kiro) as a coding partner throughout the project. Here's how the work split:

AI generated:
- Initial boilerplate: Gin router setup, middleware, config loader, Dockerfile, docker-compose
- PostgreSQL repository CRUD methods (the repetitive query/scan patterns)
- Validator functions and error struct scaffolding
- README and project file structure

I decided and directed:
- Architecture: I chose the layered pattern (handler > service > repository) based on a production codebase I work with. The AI defaulted to Go's internal/ convention; I overrode it.
- Short code generation strategy: I evaluated three approaches and picked DB sequence + block allocation. This was entirely my decision after thinking through the trade-offs.
- Duplicate URL policy: I decided that shortening the same URL twice (without alias) should be idempotent — return the existing short code. This is a product decision.
- API response format: I standardized on {requestId, payload, error} based on a team convention I follow.
- DB connection ownership: I moved all database lifecycle logic out of main.go into the store layer, so main.go is pure orchestration.


2. Where did I override, correct, or throw away the AI's output — and why?
---------------------------------------------------------------------------

Short code generator (replaced three times):
- v1: AI generated crypto/rand Base62. Simple, but requires a DB existence check on every single insert to avoid collisions. Unnecessary latency at scale.
- v2: I asked for Snowflake-style. Works well, but produces 10-11 character codes and requires a MACHINE_ID environment variable on every instance. Over-configured for this problem.
- v3 (final): DB sequence + block allocation. Shorter codes (7 chars), no machine config needed, one DB call per 10,000 inserts. I directed this approach and the AI implemented it.

Project structure (overridden):
- AI created a top-level interfaces/ package with a single interface. I removed it — in Go, interfaces belong where they're consumed. The handler defines its own URLService interface locally. Cleaner, less indirection.

Repository Manager:
- AI initially passed *sql.DB directly to services from main.go. I restructured so the store layer owns the DB connection, the RepositoryManager acts as the abstraction layer, and services never see *sql.DB.

Unnecessary abstractions (removed):
- repo/repoerr/ package: AI created it but nothing imported it. Dead code, deleted.
- utils/common.go: Only had one function (UUID generation). Inlined it into the middleware file.
- SendInternalError helper: Never used anywhere. Removed.


3. The two or three biggest trade-offs I made
----------------------------------------------

Trade-off 1: Block allocation vs Snowflake vs Random

I chose DB sequence + block allocation because:
- Produces 7-character codes (vs 10-11 for Snowflake). For a URL shortener, shorter is better.
- No machine/node configuration needed. The database sequence coordinates all instances automatically.
- One DB call per 10,000 inserts vs one per insert (random approach).

The downside: if the database is unreachable, we cannot allocate new blocks. But the service already depends on PostgreSQL for storage, so this doesn't introduce a new failure mode.

If I needed fully decentralized ID generation without any shared state, I'd use Snowflake with machine IDs assigned by the orchestrator.

Trade-off 2: Idempotent shortening vs always-generate-new

I chose: same URL without alias returns the existing short code.
Alternative: always generate a new code (simpler, no GetByOriginalURL query needed).

Why idempotent: Users expect the same URL to produce the same short link. It saves storage, reduces confusion, and makes the API cacheable. The extra DB lookup (indexed on original_url) is cheap.

Trade-off 3: 301 vs 302 redirect

I chose 301 (Moved Permanently). Browsers and CDNs cache this, so repeated visits to the same short URL don't hit the server. This is the standard for permanent URL shorteners.

The downside: cached redirects bypass the server, so click analytics under-count repeat visitors. In a production system with strict analytics requirements, I'd use 302 or 307 instead.


4. What's missing, or what I'd do with another day
----------------------------------------------------

- Rate limiting: No protection against abuse. Would add per-IP rate limiting middleware.
- Expiration/TTL: Short codes live forever. Would add an optional expires_at field and a background cleanup job.
- Analytics API: Click counts are tracked but not exposed via an endpoint. Would add GET /api/v1/stats/:code.
- Caching: Would add Redis in front of redirect lookups for hot codes.
- Integration tests: Would use testcontainers to spin up a real PostgreSQL and verify the block allocation round-trip end-to-end.
- Observability: Structured logging is in place, but no Prometheus metrics or distributed tracing.
- URL safety: Validation exists but there's no check against redirect to known phishing/malware domains.
