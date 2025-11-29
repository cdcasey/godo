# Project Instructions for Claude Code

## Project Overview
A todo API built with Go and Chi router. Authentication and authorization are critical - the API must enforce role-based access control strictly.

## Stack
- **Language**: Go (latest stable version)
- **Router**: Chi (github.com/go-chi/chi/v5)
- **Database**: SQLite (for now, should be swappable later)
- **Auth**: JWT tokens with role-based access control
- **Testing**: 100% test coverage required
- **Logging**: Structured logging with slog (Go standard library)
- **Deployment**: Docker with docker-compose

## Architecture
- RESTful API design
- Middleware for auth, logging, CORS, and error handling
- Repository pattern for database access
- Clear separation between handlers, services, and data layers

## Authorization Rules
- **Users**: Can create todos, mark their own todos as complete, view their own todos
- **Admins**: Can do everything users can, plus delete any todo
- All endpoints (except login/register) require authentication
- Role checks must happen at the middleware level where possible

## Database Schema
- `users` table: id, email, password_hash, role (user/admin), created_at
- `todos` table: id, user_id, title, description, completed, created_at, updated_at
- Proper foreign key constraints and indexes

## Migration Strategy
- Use `golang-migrate/migrate` for database migrations
- Migrations stored in `/migrations` directory
- Format: `YYYYMMDDHHMMSS_description.up.sql` and `YYYYMMDDHHMMSS_description.down.sql`
- Migrations run automatically on application startup
- Each migration should be reversible
- Never modify existing migrations - create new ones for changes

Example migration structure:
```
/migrations
  000001_create_users_table.up.sql
  000001_create_users_table.down.sql
  000002_create_todos_table.up.sql
  000002_create_todos_table.down.sql
```

## Logging Strategy
- Use Go's standard library `log/slog` for structured logging
- Log levels: DEBUG, INFO, WARN, ERROR
- Log format: JSON for production, text for development
- Configure via environment variable: `LOG_LEVEL` and `LOG_FORMAT`
- Include request ID in all logs for tracing
- Log all auth attempts (success and failure)
- Log all authorization failures
- Never log sensitive data (passwords, full tokens)

What to log:
- INFO: Request start/end, auth success, todo operations
- WARN: Invalid input, auth failures, rate limit hits
- ERROR: Database errors, unexpected failures
- DEBUG: Detailed request/response data (dev only)

## CORS Policy
- Use Chi's CORS middleware
- Allowed origins: Configure via environment variable `ALLOWED_ORIGINS`
- Default for development: `http://localhost:3000`
- Allowed methods: GET, POST, PATCH, DELETE, OPTIONS
- Allowed headers: Content-Type, Authorization
- Allow credentials: true (for cookies if needed later)
- Max age: 300 seconds

## Code Style Requirements
- Idiomatic Go: follow standard Go conventions
- Error handling: always handle errors explicitly, never ignore them
- Use standard library where possible
- Clear, descriptive variable and function names
- Comments for exported functions and non-obvious logic

## Testing Requirements
- 100% test coverage (use `go test -cover`)
- Table-driven tests for handlers
- Mock database for unit tests
- Integration tests for critical paths (auth flow, CRUD operations)
- Test both success and failure cases
- Test authorization rules thoroughly

## Important: Code Delivery Style
**CRITICAL**: Do not generate complete files or large blocks of code at once.

Instead:
1. Provide small, incremental snippets (5-20 lines max)
2. Explain what each snippet does before showing the code
3. Wait for confirmation before moving to the next piece
4. Break down file creation into logical steps:
   - Imports first
   - Type definitions next
   - Then individual functions one at a time
5. Ask "What would you like to add next?" or "Should I continue?" between steps

Example flow:
- Me: "Let's create the user model"
- You: "Here are the imports and User struct..."
- Me: "Looks good, what's next?"
- You: "Here's the password hashing function..."

## API Endpoints (to implement)
```
POST   /api/register          - Create new user account
POST   /api/login             - Get JWT token
GET    /api/health            - Health check endpoint
GET    /api/todos             - List user's todos (or all for admins)
POST   /api/todos             - Create new todo
GET    /api/todos/:id         - Get specific todo
PATCH  /api/todos/:id         - Update todo (complete/uncomplete)
DELETE /api/todos/:id         - Delete todo (admin only)
```

## Security Considerations
- Hash passwords with bcrypt
- Use environment variables for secrets (JWT secret, DB connection)
- Validate all input
- Rate limiting on auth endpoints (consider using Chi middleware)
- SQL injection prevention (use parameterized queries)
- Proper HTTP status codes

## Docker Deployment

### Dockerfile
- Multi-stage build: builder stage and minimal runtime stage
- Use `golang:1.21-alpine` for builder
- Use `alpine:latest` for runtime
- Copy only necessary files to final image
- Run as non-root user
- Expose port 8080

### docker-compose.yml
- API service with environment variables
- Volume for SQLite database persistence
- Health check configuration
- Restart policy: unless-stopped

### Environment Variables
```
PORT=8080
DATABASE_URL=/data/todos.db
JWT_SECRET=<generate-strong-secret>
LOG_LEVEL=info
LOG_FORMAT=json
ALLOWED_ORIGINS=http://localhost:3000
```

## Project Structure
```
/cmd/api            - Main application entry point
/internal/handlers  - HTTP handlers
/internal/models    - Data models
/internal/store     - Database layer
/internal/auth      - Auth middleware and JWT logic
/internal/testutil  - Testing helpers
/migrations         - Database migrations
/docker             - Dockerfile and related files
Makefile            - Common commands (build, test, migrate, etc.)
docker-compose.yml  - Docker composition
.env.example        - Example environment variables
```

## Makefile Commands
```
make build          - Build the application
make test           - Run tests with coverage
make migrate-up     - Run migrations
make migrate-down   - Rollback last migration
make docker-build   - Build Docker image
make docker-up      - Start with docker-compose
make docker-down    - Stop docker-compose
make lint           - Run golangci-lint
```
