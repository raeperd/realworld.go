# GET /api/user Endpoint Implementation

## Status & Links

**Status**: In Progress
**PR Link**: TBD

## Context

Implementing the `GET /api/user` endpoint to retrieve the authenticated user's information. This is the natural next step after implementing user registration and login endpoints.

### Original Request
Implement `GET /api/user` endpoint from the RealWorld API specification.

### Key Requirements from spec.md
- **Endpoint**: `GET /api/user` (line 52)
- **Authentication**: Required (JWT token in `Authorization: Token <jwt>` header)
- **Response**: Returns User object with email, token, username, bio, image (lines 263-273)
- **Error Handling**: 401 for unauthorized requests (line 434)

## Methodology

Following TDD principles as defined in `@docs/prompts/TDD.md`:
- Red → Green → Refactor cycle
- Write failing test first
- Implement minimum code to pass
- Refactor only when tests pass

## Feature-Specific Requirements

### Authentication Middleware
- Extract JWT token from `Authorization: Token <jwt>` header
- Validate token using existing JWT verification (internal/auth package)
- Attach user ID to request context
- Return 401 for missing/invalid tokens

### Handler Implementation
- Retrieve user from database using ID from context
- Return user data in spec-compliant format
- Include the same token in response
- Handle database errors appropriately

### Database Query
- Query users table by ID
- Return email, username, bio, image fields
- Use existing sqlc queries if available, or raw SQL

## Implementation Steps

### Step 1: Write failing test for valid token
- [ ] Create test in `user_test.go` for `TestGetUser`
- [ ] Register a user to get valid JWT token
- [ ] Make GET request to `/api/user` with `Authorization: Token <jwt>` header
- [ ] Assert 200 status code
- [ ] Assert response contains correct user data (email, username, token, bio, image)
- [ ] Run test - should fail (handler not implemented)

### Step 2: Implement JWT authentication middleware
- [ ] Create `auth` middleware function in `main.go`
- [ ] Extract token from `Authorization` header (strip "Token " prefix)
- [ ] Parse and validate JWT using `internal/auth/Verify()`
- [ ] Extract user ID from claims
- [ ] Store user ID in request context
- [ ] Return 401 if token missing/invalid
- [ ] Run test - should still fail (handler not implemented)

### Step 3: Implement handleGetUser handler
- [ ] Create `handleGetUser(db *sql.DB, jwtSecret string)` function
- [ ] Query database for user by ID from context
- [ ] Build response JSON matching spec format
- [ ] Include original token in response
- [ ] Register route with auth middleware: `mux.Handle("GET /api/user", auth(handleGetUser(db, jwtSecret), jwtSecret))`
- [ ] Run test - should pass

### Step 4: Add test for invalid token
- [ ] Write test in `user_test.go` for invalid JWT token
- [ ] Make request with malformed token
- [ ] Assert 401 status code
- [ ] Run test - should pass (middleware handles this)

### Step 5: Add test for missing token
- [ ] Write test in `user_test.go` for missing Authorization header
- [ ] Make request without Authorization header
- [ ] Assert 401 status code
- [ ] Run test - should pass (middleware handles this)

### Step 6: Refactor if needed
- [ ] Check for code duplication
- [ ] Improve naming and structure
- [ ] Ensure all tests still pass

## Verification Commands

```bash
# Run specific test
go test -v -run TestGetUser

# Run all tests
make test

# Check test coverage
go test -cover ./...

# Verify server starts
make run

# Manual test with curl (after creating user)
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"test","email":"test@test.com","password":"password"}}'

# Use returned token
curl -X GET http://localhost:8080/api/user \
  -H "Authorization: Token <jwt-from-above>"
```

## Expected Outcome

A fully tested `GET /api/user` endpoint that:
1. Validates JWT authentication
2. Returns current user's data
3. Handles error cases (401 for missing/invalid tokens)
4. Follows RealWorld API specification
5. Includes reusable auth middleware for future endpoints
