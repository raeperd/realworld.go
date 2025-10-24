# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a **RealWorld Example App** backend implementation in Go, demonstrating a fully-featured Medium.com clone API. Built using the [kickstart.go](https://github.com/raeperd/kickstart.go) template, it implements the complete [RealWorld API specification](https://github.com/gothinkster/realworld) while maintaining the template's principles of simplicity and minimal dependencies.

### RealWorld Project Context
- **Purpose**: "The mother of all demo apps" - showcases real-world patterns vs simple todo apps  
- **API Specification**: Standardized across 100+ implementations in different languages
- **Features**: User authentication, articles, comments, favorites, following, tags, feeds
- **Demo**: [demo.realworld.show](https://demo.realworld.show) | **Docs**: [docs.realworld.show](https://docs.realworld.show)

## Common Development Commands

### Build and Run
```bash
# Build and run server on port 8080
make run

# Hot reload development (requires air)
make watch

# Docker development environment with Swagger UI
docker-compose up
```

### Testing
```bash
# Run all tests with race detection and coverage
make test

# Run specific test
go test -v -run TestHealth

# Run tests with verbose output
go test -v ./...
```

### Code Quality
```bash
# Run golangci-lint
make lint

# Clean build artifacts
make clean
```

### Docker Operations
```bash
# Build Docker image
make docker

# Multi-platform build (requires buildx)
docker buildx build --platform linux/amd64,linux/arm64 -t kickstart.go .
```

## Architecture

### RealWorld API Implementation
This implementation follows the [RealWorld API specification](https://github.com/gothinkster/realworld) exactly, enabling frontend interoperability with React, Angular, Vue, and other RealWorld frontends.

**API Specification Reference**: See `docs/spec.md` for the complete endpoint specifications, request/response formats, and error handling requirements that must be implemented.

### Single-File Design (from kickstart.go template)
The entire server implementation resides in `main.go` with a clear structure:
1. **main()** - Entry point that calls run() with dependencies
2. **run()** - Core server logic, returns error for testability
3. **Middleware** - accessLogMiddleware and recoveryMiddleware for cross-cutting concerns
4. **Handlers** - healthHandler and openapiHandler for core endpoints
5. **Embedded Assets** - OpenAPI spec embedded via go:embed

### Key Patterns
- **Dependency Injection**: run() accepts context, writer, args, and version for testability
- **Graceful Shutdown**: Proper signal handling (SIGINT/SIGTERM) with cleanup
- **Middleware Chain**: Composable middleware (logging → recovery → routes)
- **Integration Testing**: Tests use real HTTP server with dynamic port allocation

### Endpoints
- `GET /health` - Returns service health with version, uptime, git commit
- `GET /openapi.yaml` - Serves embedded OpenAPI specification (CORS enabled)
- `GET /debug/pprof/*` - Go profiling endpoints (CPU, heap, goroutines)
- `GET /debug/vars` - Runtime metrics via expvar

## Testing Approach

Tests follow integration testing principles with real server instances:
- **TestMain** sets up actual HTTP server on dynamic port
- Tests make real HTTP requests using http.Client
- All tests run in parallel with `t.Parallel()`
- Middleware components have focused unit tests

When adding new features:
1. Write integration test first that exercises the full HTTP path
2. Implement minimal code to make test pass
3. Add unit tests only for complex logic or middleware

## Version Management

Version information is embedded at build time:
- Set via `VERSION` environment variable or defaults to "dev"
- Git commit hash automatically extracted via debug.ReadBuildInfo()
- Exposed in /health endpoint response

Build with version:
```bash
VERSION=v1.0.0 make build
```

## Adding New Features

When extending the server:
1. **New Endpoints**: Add handler function and register in run() mux
2. **New Middleware**: Create middleware function following the pattern of existing ones
3. **Configuration**: Use environment variables, add to run() parameters if needed
4. **External Dependencies**: Consider if truly necessary - this template values simplicity

## Implementation Plans & History

**IMPORTANT**: Before implementing any new feature, check `docs/prompts/` directory for:
- Existing implementation plans and their approach
- Historical context and decisions made
- Examples of successful TDD workflows used in this project

When planning a new feature implementation, **always create a new plan document** in `docs/prompts/YYYY-MM-DD-feature-name.md` before starting work.

### Plan Document Guidelines

**Location**: `docs/prompts/YYYY-MM-DD-feature-name.md`

**Required Sections**:
1. **Status & Links**: Implementation status and PR link (when created)
2. **Context**: Original user request and clarifications
3. **Methodology**: Reference `@docs/prompts/TDD.md` (avoid repeating TDD content)
4. **Feature-Specific Requirements**: Only the unique requirements for this feature
5. **Implementation Steps**: Detailed, reproducible steps with checkboxes
6. **Verification Commands**: Commands to test and validate the implementation

**Key Principles**:
- **Conciseness**: Reference `@docs/prompts/TDD.md` instead of repeating TDD methodology
- **Reproducibility**: Another developer should be able to follow the plan exactly
- **Reusability**: Write as a guide that can be adapted to other repositories
- **Minimal Details**: Focus on feature-specific requirements, not general methodology

**Example Plans** (see `docs/prompts/` for full content):
- `2025-10-12-login-endpoint-implementation.md` - POST /api/users/login endpoint
- `2025-10-12-cors-middleware-implementation.md` - CORS middleware with TDD
- `2025-10-11-jwt-authentication-implementation.md` - JWT authentication setup

## Docker Development

The docker-compose setup includes:
- **app** service: Air hot reload container watching file changes
- **swagger-ui** service: Interactive API documentation at http://localhost:8081
- Health checks configured for production readiness

## CI/CD Notes

GitHub Actions workflows handle:
- **build.yaml**: Test, lint, coverage reporting, Docker build
- **deploy.yaml**: Multi-platform Docker images pushed to GitHub Container Registry
- Version injection happens automatically via git tags

When making changes that affect CI:
- Ensure tests pass locally first: `make test`
- Check lint issues: `make lint`  
- Verify Docker build: `make docker`
- do not merge pr before explicitly requested