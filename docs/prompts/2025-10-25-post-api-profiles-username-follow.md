# POST /api/profiles/:username/follow Implementation Plan

## Status & Links

Status: [x] Completed
PR: https://github.com/raeperd/realworld.go/pull/21

## Context

Implementing the `POST /api/profiles/:username/follow` endpoint from the RealWorld API specification.

**API Spec Reference:** See `docs/spec.md` - "Follow user" section (lines 82-88)

**Endpoint Details:**
- **Method:** POST
- **Path:** `/api/profiles/:username/follow`
- **Authentication:** Required
- **Purpose:** Follow a user and return their profile with `following: true`

## Methodology

This implementation follows the Test-Driven Development approach detailed in `@docs/prompts/TDD.md`.

**Testing Approach:**
- Integration tests using real HTTP requests
- Test authenticated requests (auth required)
- Use real database (not mocks)
- Tests run in parallel with `t.Parallel()`

## Feature Requirements

### Database Schema Changes

Need to add a `follows` table to track follower relationships:

```sql
CREATE TABLE follows (
    follower_id INTEGER NOT NULL,
    followed_id INTEGER NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (follower_id, followed_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (followed_id) REFERENCES users(id) ON DELETE CASCADE
);
```

### Request Format
- **Method:** POST
- **Path Parameter:** `:username` - the username of the user to follow
- **Headers (required):** `Authorization: Token <jwt>` - authenticated user becomes the follower
- **Body:** No request body required

### Response Format (Success - 200 OK)

```json
{
  "profile": {
    "username": "jake",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": true
  }
}
```

### Error Responses
- **401 Unauthorized:** Missing or invalid authentication token
- **404 Not Found:** User with the specified username doesn't exist
- **422 Unprocessable Entity:** Trying to follow yourself

### Database Operations
- Query: `GetUserByUsername` - fetch user to follow (already exists)
- Query: `CreateFollow` - insert follow relationship
- Query: `IsFollowing` - check if already following (for idempotency)
- Note: Update `handleGetProfilesUsername` to check `following` status using the new table

### Validation Rules
- Username path parameter must be extracted from URL
- Username must exist in database (404 if not found)
- User cannot follow themselves (422 if attempted)
- Authentication required (401 if missing/invalid)
- Idempotent: following an already-followed user should succeed (200 OK)

## Implementation Steps

### Phase 0: Database Schema Update
- [ ] Add `follows` table to `internal/sqlite/schema.sql`
- [ ] Add index on `followed_id` for performance: `CREATE INDEX idx_follows_followed_id ON follows(followed_id);`
- [ ] Verify schema compiles: `go build`

### Phase 1: Database Queries
- [ ] Add `CreateFollow` query to `internal/sqlite/query.sql`
- [ ] Add `IsFollowing` query to `internal/sqlite/query.sql`
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`

### Phase 2: Test First (RED) - Happy Path
- [ ] Write failing test `TestPostProfilesUsernameFollow_Success` in `profile_test.go`
  - Create two users (follower and followed) via registration
  - Extract follower's JWT token
  - Make POST request to `/api/profiles/{followed-username}/follow` with auth
  - Expect 200 OK with profile containing `following: true`
- [ ] Run test to confirm it fails: `go test -v -run TestPostProfilesUsernameFollow_Success`
- [ ] Commit: "test: add failing test for POST /api/profiles/:username/follow"

### Phase 3: Minimal Implementation (GREEN)
- [ ] Create `handlePostProfilesUsernameFollow` function in `profile.go`
- [ ] Extract username from URL path using `r.PathValue("username")`
- [ ] Extract follower user ID from context (set by authenticate middleware)
- [ ] Query database for user to follow by username
- [ ] Create follow relationship in database
- [ ] Return profile response with `following: true`
- [ ] Register route in `route()` function in `main.go` with authenticate middleware
- [ ] Run test to confirm it passes: `go test -v -run TestPostProfilesUsernameFollow_Success`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement POST /api/profiles/:username/follow"

### Phase 4: Edge Cases & Validation (RED → GREEN)

#### Test 4a: User Not Found
- [ ] Add test `TestPostProfilesUsernameFollow_NotFound`
  - Attempt to follow non-existent username
  - Expect 404 Not Found
- [ ] Implement 404 handling for `sql.ErrNoRows`
- [ ] Run test to confirm it passes: `go test -v -run TestPostProfilesUsernameFollow_NotFound`
- [ ] Commit: "test: add not found test for POST /api/profiles/:username/follow"

#### Test 4b: Follow Self
- [ ] Add test `TestPostProfilesUsernameFollow_FollowSelf`
  - User attempts to follow their own username
  - Expect 422 Unprocessable Entity with error message
- [ ] Implement validation to prevent self-follow
- [ ] Run test to confirm it passes: `go test -v -run TestPostProfilesUsernameFollow_FollowSelf`
- [ ] Commit: "test: add self-follow validation test"

#### Test 4c: Idempotent Follow
- [ ] Add test `TestPostProfilesUsernameFollow_AlreadyFollowing`
  - Create follow relationship first
  - Attempt to follow the same user again
  - Expect 200 OK with `following: true` (no error)
- [ ] Implement idempotent handling (INSERT OR IGNORE or check before insert)
- [ ] Run test to confirm it passes: `go test -v -run TestPostProfilesUsernameFollow_AlreadyFollowing`
- [ ] Commit: "test: add idempotent follow test"

#### Test 4d: Unauthorized Request
- [ ] Add test `TestPostProfilesUsernameFollow_Unauthorized`
  - Make request without Authorization header
  - Expect 401 Unauthorized
- [ ] Verify authenticate middleware handles this (should already work)
- [ ] Run test to confirm it passes: `go test -v -run TestPostProfilesUsernameFollow_Unauthorized`
- [ ] Commit: "test: add unauthorized test for POST /api/profiles/:username/follow"

### Phase 5: Update GET Profile Endpoint
- [ ] Update `handleGetProfilesUsername` to check actual following status
- [ ] Add query to check if authenticated user is following the profile user
- [ ] Update tests for `GET /api/profiles/:username` to verify following status
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: update GET /api/profiles/:username to show actual following status"

### Phase 6: Refactor (if needed)
- [ ] Review `profile.go` for duplication
- [ ] Extract common profile building logic if found in multiple places
- [ ] Ensure all tests still pass after each refactoring: `make test`
- [ ] Commit: "refactor: {description}" (if changes made)

### Phase 7: Verification
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl:
  ```bash
  # Create two users
  curl -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -d '{"user":{"username":"alice","email":"alice@test.com","password":"password"}}'

  curl -X POST http://localhost:8080/api/users \
    -H "Content-Type: application/json" \
    -d '{"user":{"username":"bob","email":"bob@test.com","password":"password"}}'

  # Login as alice to get token
  TOKEN=$(curl -X POST http://localhost:8080/api/users/login \
    -H "Content-Type: application/json" \
    -d '{"user":{"email":"alice@test.com","password":"password"}}' | jq -r '.user.token')

  # Follow bob as alice
  curl -X POST http://localhost:8080/api/profiles/bob/follow \
    -H "Authorization: Token $TOKEN"

  # Verify following status
  curl -X GET http://localhost:8080/api/profiles/bob \
    -H "Authorization: Token $TOKEN"
  ```
- [ ] Update this plan's status to "Completed"

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestPostProfilesUsernameFollow

# Run all tests
make test

# Lint check
make lint

# Build check
go build

# Generate sqlc code
make generate
```

## Notes

- This endpoint enables the social following feature
- The `follows` table uses a composite primary key (follower_id, followed_id)
- Foreign keys ensure referential integrity with CASCADE delete
- Idempotent design: following an already-followed user returns success
- Self-follow prevention improves user experience
- After this, implement `DELETE /api/profiles/:username/follow` for unfollowing
- The `following` field in GET /api/profiles/:username will now show actual status
