# RealWorld Example App (Go)

> ### Go backend implementation of the [RealWorld](https://github.com/gothinkster/realworld) API specification

This codebase was created to demonstrate a fully fledged fullstack application built with **Go** including CRUD operations, authentication, routing, pagination, and more.

We've gone to great lengths to adhere to the **Go** community styleguides & best practices.

For more information on how to this works with other frontends/backends, head over to the [RealWorld](https://github.com/gothinkster/realworld) repo.

## About RealWorld

**"The mother of all demo apps"** — Exemplary fullstack [Medium.com](https://medium.com) clone powered by React, Angular, Node, Django, and many more 🏅

While most "todo" demos provide an excellent cursory glance at a framework's capabilities, they typically don't convey the knowledge & perspective required to actually build _real_ applications with it.

**RealWorld** solves this by allowing you to choose any frontend (React, Angular, & more) and any backend (Node, Django, & more) and see how the exact same Medium.com clone is built using different tech stacks.

## Built with kickstart.go template

This implementation is built using [raeperd/kickstart.go](https://github.com/raeperd/kickstart.go), a minimalistic HTTP server template in Go that provides:
- Small codebase (production-ready starting point)
- Single file server design
- Only standard library dependencies
- Graceful shutdown and health monitoring

## Features

### RealWorld API Implementation
- **User Management**: Registration, authentication, and profile management
- **Articles**: CRUD operations for articles with slug-based URLs
- **Comments**: Add, view, and delete comments on articles  
- **Favorites**: Like and unlike articles
- **Following**: Follow and unfollow other users
- **Tags**: Discover articles by tags
- **Feeds**: Global feed and personalized feed for followed users

### Technical Features
- **RESTful API**: Following RealWorld API specification
- **JWT Authentication**: Secure token-based authentication
- **Input Validation**: Request validation and error handling
- **CORS Support**: Cross-origin resource sharing enabled
- **OpenAPI Documentation**: Interactive API documentation
- **Graceful Shutdown**: Handles `SIGINT` and `SIGTERM` signals
- **Health Monitoring**: Service health status with version info
- **Access Logging**: Comprehensive HTTP request logging
- **Panic Recovery**: Graceful error handling and recovery
- **Debug Endpoints**: Built-in profiling and metrics

## Getting started

### Requirements
- Go 1.24 or later
- Make (for build automation)

### Quick Start

1. **Clone and setup**
   ```console
   git clone <your-repo-url>
   cd realworld.go
   ```

2. **Build and run**
   ```console
   make run
   ```
   Server will start on port 8080

3. **Test the API**
   ```console
   # Health check
   curl http://localhost:8080/health
   
   # View API documentation
   curl http://localhost:8080/openapi.yaml
   ```

### Development Setup

#### Suggested Dependencies
- [golangci-lint](https://golangci-lint.run/) - Code linting
- [air](https://github.com/air-verse/air) - Hot reload development

#### Development with Docker
```console
# Start development environment with hot reload
docker-compose up

# API server: http://localhost:8080  
# Swagger UI: http://localhost:8081
```

## API Endpoints

### RealWorld API
All endpoints follow the [RealWorld API specification](https://github.com/gothinkster/realworld):

- **Authentication**
  - `POST /api/users/login` - Login user
  - `POST /api/users` - Register user
  - `GET /api/user` - Get current user
  - `PUT /api/user` - Update user

- **Profiles**  
  - `GET /api/profiles/:username` - Get profile
  - `POST /api/profiles/:username/follow` - Follow user
  - `DELETE /api/profiles/:username/follow` - Unfollow user

- **Articles**
  - `GET /api/articles` - List articles (with filters)
  - `GET /api/articles/feed` - Get user feed
  - `POST /api/articles` - Create article
  - `GET /api/articles/:slug` - Get article
  - `PUT /api/articles/:slug` - Update article  
  - `DELETE /api/articles/:slug` - Delete article

- **Comments**
  - `GET /api/articles/:slug/comments` - Get comments
  - `POST /api/articles/:slug/comments` - Add comment
  - `DELETE /api/articles/:slug/comments/:id` - Delete comment

- **Favorites & Tags**
  - `POST /api/articles/:slug/favorite` - Favorite article
  - `DELETE /api/articles/:slug/favorite` - Unfavorite article
  - `GET /api/tags` - Get tags

### Service Endpoints
- `GET /health` - Service health with version info
- `GET /openapi.yaml` - OpenAPI specification  
- `GET /debug/pprof/*` - Profiling information
- `GET /debug/vars` - Runtime metrics

## Testing

```console
# Run all tests
make test

# Run specific test
go test -v -run TestHealth

# Run tests with coverage
go test -v -cover ./...
```

## Deployment

### Using Docker
```console
# Build Docker image
make docker

# Run containerized
docker run -p 8080:8080 realworld.go
```

### Environment Variables
- `VERSION` - Application version (default: "dev")
- `PORT` - Server port (default: "8080")

## Project Structure

```
.
├── main.go              # Main server implementation
├── main_test.go         # Integration tests  
├── api/
│   └── openapi.yaml     # OpenAPI specification
├── docker-compose.yaml  # Development environment
├── Dockerfile          # Container build
├── Makefile            # Build automation
└── README.md           # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes following Go best practices
4. Add tests for new functionality  
5. Ensure tests pass: `make test`
6. Run linter: `make lint`
7. Submit a pull request

## Links

- [RealWorld Project](https://github.com/gothinkster/realworld) - Main RealWorld repository
- [RealWorld Demo](https://demo.realworld.show) - Live demo application  
- [RealWorld Docs](https://docs.realworld.show) - API documentation
- [kickstart.go](https://github.com/raeperd/kickstart.go) - Original template used
