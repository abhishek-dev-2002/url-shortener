# Project Go Rules

## Architecture
- Always follow the existing project structure (interfaces, services, repo, utils, etc.)
- Business logic must reside in the services layer
- Database operations must be handled only via the repo layer
- Do NOT introduce new architectural patterns or layers

## Code Practices
- Always pass context.Context through all request flows
- Follow existing error handling patterns used in the codebase
- Avoid introducing new coding styles; maintain consistency with existing files
- Keep functions small, readable, and consistent with current patterns

## Logging
- Use the existing logger from the utils package
- Include relevant context (requestId, userId, etc.) in logs
- Log errors with sufficient detail for debugging

## Database
- Use only existing repository methods for DB access
- Do NOT access the database directly from services or handlers
- Follow existing MongoDB schema and query patterns

## Consistency
- Follow naming conventions already used in the project
- Reuse existing utilities and helpers instead of creating new ones
- Keep code aligned with similar modules in the repository

## Safety Rules
- Do NOT refactor existing code unless explicitly asked
- Do NOT modify unrelated files
- Preserve backward compatibility unless specified