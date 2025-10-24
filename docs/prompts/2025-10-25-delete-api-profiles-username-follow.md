# DELETE /api/profiles/:username/follow - Unfollow User

**Status:** [ ] Not Started [x] In Progress [ ] Completed
**PR:** #23

## Context

Implementing the unfollow user endpoint as specified in the RealWorld API specification. This endpoint allows authenticated users to unfollow another user.

**Original Request:** Implement DELETE /api/profiles/:username/follow

**API Specification Reference:** `@docs/spec.md` - "Unfollow user" section (lines 90-96)

## Methodology

This implementation follows Test-Driven Development (TDD) principles as defined in `@docs/prompts/TDD.md`.

**Testing Approach:**
- Integration tests with real HTTP server
- Real database operations (no mocks)
- Tests run in parallel where possible

## Feature Requirements

**Endpoint:** DELETE /api/profiles/:username/follow

**Authentication:** Required - user must be authenticated with valid JWT token

**Request Format:**
- No request body required
- Username provided as path parameter
- Authorization header: `Token <jwt>`

**Response Format (200 OK):**
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

**Error Cases:**
- 401 Unauthorized: Missing or invalid authentication token
- 404 Not Found: User to unfollow does not exist
- 422 Unprocessable Entity: Cannot unfollow yourself

**Database Operations:**
1. Verify target user exists by username
2. Delete follow relationship from `follows` table
3. Return profile with `following: false`

**Note:** Unfollowing a user that is not currently followed should be idempotent (not return an error)

## Implementation Steps

### Phase 1: Database Layer

- [x] Review existing SQL queries in `internal/sqlite/queries.sql`
- [x] Verify `DeleteFollow` query exists (added in follow endpoint implementation)
- [x] No new queries needed - can reuse existing queries

### Phase 2: Test First (RED)

- [ ] Add test to `profile_test.go`: `TestHandleDeleteProfilesUsernameFollow`
- [ ] Write failing integration test for happy path (successful unfollow)
- [ ] Run test to confirm it fails: `go test -v -run TestHandleDeleteProfilesUsernameFollow`
- [ ] Commit: "test: add failing test for DELETE /api/profiles/:username/follow"
- [ ] Push immediately: `git push -u origin feat/delete-api-profiles-username-follow`

### Phase 3: Minimal Implementation (GREEN)

- [ ] Create `handleDeleteProfilesUsernameFollow()` function in `profile.go`
- [ ] Extract username from path parameter
- [ ] Get authenticated user ID from context
- [ ] Delete follow relationship using existing `DeleteFollow` query
- [ ] Return profile with `following: false`
- [ ] Register route in `route()` function in `main.go`
- [ ] Run test to confirm it passes: `go test -v -run TestHandleDeleteProfilesUsernameFollow`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement DELETE /api/profiles/:username/follow"
- [ ] Push immediately: `git push`

### Phase 4: Edge Cases & Validation (RED → GREEN)

- [ ] Add test for unauthorized access (missing token)
- [ ] Verify authentication middleware handles this (should already work)
- [ ] Add test for user not found scenario
- [ ] Implement not found handling if needed
- [ ] Add test for attempting to unfollow yourself
- [ ] Implement self-unfollow validation
- [ ] Add test for idempotent unfollow (unfollowing already unfollowed user)
- [ ] Verify idempotent behavior works correctly
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add edge case tests for DELETE /api/profiles/:username/follow"
- [ ] Push immediately: `git push`

### Phase 5: Refactor (if needed)

- [ ] Review code for duplication between follow/unfollow handlers
- [ ] Consider extracting common profile response building logic if found in 3+ places
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit: "refactor: {description}" (if changes made)
- [ ] Push immediately: `git push`

### Phase 6: Verification

- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl
- [ ] Update this plan's status to "Completed"

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestHandleDeleteProfilesUsernameFollow

# Run all profile tests
go test -v -run TestHandle.*Profiles

# Run all tests
make test

# Lint check
make lint

# Manual test example
# 1. Register a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"follower","email":"follower@test.com","password":"password"}}'

# 2. Register another user to unfollow
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"followed","email":"followed@test.com","password":"password"}}'

# 3. Follow the user first
curl -X POST http://localhost:8080/api/profiles/followed/follow \
  -H "Authorization: Token <token_from_step_1>"

# 4. Unfollow the user
curl -X DELETE http://localhost:8080/api/profiles/followed/follow \
  -H "Authorization: Token <token_from_step_1>"
```

## Expected Behavior

1. Authenticated user can unfollow another user successfully
2. Response returns profile with `following: false`
3. Operation is idempotent - unfollowing an already unfollowed user succeeds
4. Cannot unfollow yourself
5. Returns 404 if target user does not exist
6. Returns 401 if authentication is missing or invalid
