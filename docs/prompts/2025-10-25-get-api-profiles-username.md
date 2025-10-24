# GET /api/profiles/:username Implementation Plan

## Status & Links

Status: [x] Completed
PR: #19

## Context

Implementing the `GET /api/profiles/:username` endpoint from the RealWorld API specification.

**API Spec Reference:** See `docs/spec.md` - "Get Profile" section (lines 76-80)

**Endpoint Details:**
- **Method:** GET
- **Path:** `/api/profiles/:username`
- **Authentication:** Optional (affects `following` field)
- **Purpose:** Return a user's public profile

## Methodology

This implementation follows the Test-Driven Development approach detailed in `@docs/prompts/TDD.md`.

**Testing Approach:**
- Integration tests using real HTTP requests
- Test both authenticated and unauthenticated access
- Use real database (not mocks)
- Tests run in parallel with `t.Parallel()`

## Feature Requirements

### Request Format
- **Method:** GET
- **Path Parameter:** `:username` - the username of the profile to retrieve
- **Headers (optional):** `Authorization: Token <jwt>` - for authenticated requests

### Response Format (Success - 200 OK)

```json
{
  "profile": {
    "username": "jake",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": false
  }
}
```

### Error Responses
- **404 Not Found:** When user with the specified username doesn't exist
- **401 Unauthorized:** Only if auth token is provided but invalid (optional auth doesn't fail if missing)

### Database Operations
- Query: `GetUserByUsername` - fetch user by username from users table
- Note: `following` field will always be `false` until follow/unfollow endpoints are implemented

### Validation Rules
- Username path parameter must be extracted from URL
- Username must exist in database (404 if not found)
- Optional authentication - don't fail if Authorization header is missing

## Implementation Steps

### Phase 0: Create Profile File Structure
- [ ] Create `profile.go` for profile-related handlers
- [ ] Create `profile_test.go` for profile integration tests

### Phase 1: Database Layer
- [ ] Add `GetUserByUsername` query to `internal/sqlite/query.sql`
- [ ] Run `make sqlc` to generate Go code
- [ ] Verify generated code compiles: `go build`

### Phase 2: Test First (RED) - Happy Path
- [ ] Write failing test `TestGetProfilesUsername_Success` in `profile_test.go`
  - Create a test user via registration
  - Make GET request to `/api/profiles/{username}` without auth
  - Expect 200 OK with correct profile data
  - Expect `following: false`
- [ ] Run test to confirm it fails: `go test -v -run TestGetProfilesUsername_Success`
- [ ] Commit: "test: add failing test for GET /api/profiles/:username"

### Phase 3: Minimal Implementation (GREEN)
- [ ] Create `handleGetProfilesUsername` function in `profile.go`
- [ ] Add `profileResponseBody` type with proper JSON structure
- [ ] Extract username from URL path using `r.PathValue("username")`
- [ ] Query database for user by username
- [ ] Return profile response with `following: false`
- [ ] Register route in `route()` function in `main.go` (no auth middleware)
- [ ] Run test to confirm it passes: `go test -v -run TestGetProfilesUsername_Success`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement GET /api/profiles/:username"

### Phase 4: Edge Cases & Validation (RED → GREEN)

#### Test 4a: User Not Found
- [ ] Add test `TestGetProfilesUsername_NotFound`
  - Request profile for non-existent username
  - Expect 404 Not Found
- [ ] Implement 404 handling for `sql.ErrNoRows`
- [ ] Run test to confirm it passes: `go test -v -run TestGetProfilesUsername_NotFound`
- [ ] Commit: "test: add not found test for GET /api/profiles/:username"

#### Test 4b: Authenticated Request
- [ ] Add test `TestGetProfilesUsername_WithAuth`
  - Create two users (viewer and target)
  - Make authenticated request from viewer to get target's profile
  - Expect 200 OK with correct profile data
  - Expect `following: false` (will be true once follow feature is implemented)
- [ ] Update handler to optionally extract auth from context (if present)
- [ ] Run test to confirm it passes: `go test -v -run TestGetProfilesUsername_WithAuth`
- [ ] Commit: "test: add authenticated request test for GET /api/profiles/:username"

### Phase 5: Refactor (if needed)
- [ ] Review `profile.go` and `user.go` for duplication
- [ ] Check if `profileResponseBody` can be unified with user response types
- [ ] Ensure all tests still pass after each refactoring: `make test`
- [ ] Commit: "refactor: {description}" (if changes made)

### Phase 6: Verification
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (without auth):
  ```bash
  # Create a user first
  curl -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -d '{"user":{"username":"testuser","email":"test@test.com","password":"password"}}'

  # Get profile
  curl -X GET http://localhost:8080/api/profiles/testuser
  ```
- [ ] Manual test with auth:
  ```bash
  # Get token from login/registration, then:
  curl -X GET http://localhost:8080/api/profiles/testuser \
    -H "Authorization: Token <jwt-token>"
  ```
- [ ] Update this plan's status to "Completed"

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestGetProfilesUsername

# Run all tests
make test

# Lint check
make lint

# Build check
go build
```

## Notes

- The `following` field will always be `false` until we implement:
  - A `follows` table in the schema
  - `POST /api/profiles/:username/follow` endpoint
  - `DELETE /api/profiles/:username/follow` endpoint
- This endpoint does NOT require authentication but should handle it gracefully if provided
- Path parameter extraction uses `r.PathValue("username")` (Go 1.22+ pattern)
- Profile response uses `profile` wrapper, not `user` wrapper
