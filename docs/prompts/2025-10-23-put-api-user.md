# PUT /api/user Endpoint Implementation

## Status & Links

**Status**: ✅ Completed
**PR Link**: https://github.com/raeperd/realworld.go/pull/16

## Context

Implementing the `PUT /api/user` endpoint to allow authenticated users to update their profile information.

### Original Request
Command: `/api PUT /api/user`

### Key Requirements from spec.md
- **Endpoint**: `PUT /api/user` (line 58)
- **Authentication**: Required (JWT token in `Authorization: Token <jwt>` header)
- **Request Body**: JSON with optional fields: email, username, password, image, bio (lines 63-69)
- **Response**: Returns updated User object with email, token, username, bio, image
- **Accepted Fields**: email, username, password, image, bio (line 74)
- **Error Handling**:
  - 401 for unauthorized requests
  - 422 for validation errors

## Methodology

Following TDD principles as defined in `@docs/prompts/TDD.md`:
- Red → Green → Refactor cycle
- Write failing test first
- Implement minimum code to pass
- Refactor only when tests pass

## Feature-Specific Requirements

### Authentication
- Must use existing `authenticate` middleware
- User ID extracted from JWT token in context

### Request Validation
- All fields are optional (can update any combination)
- Empty/missing fields should be ignored (not updated)
- Email format validation (if provided)
- Username uniqueness check (if changing username)
- Email uniqueness check (if changing email)

### Database Operation
- Need to add `UpdateUser` query to `internal/sqlite/query.sql`
- Only update fields that are provided in request
- Use transaction for consistency
- Return updated user data

### Response Format
- Must match User object format from spec
- Include fresh JWT token in response
- Wrap in "user" object as per spec

## Implementation Steps

### Phase 1: Test First (RED)

- Add test to `user_test.go` for successful update
- Test should update email, bio, and image fields
- Run test to confirm it fails: `go test -v -run TestHandlePutUser`
- Commit: "test: add failing test for PUT /api/user"

### Phase 2: Database Layer

- Add `UpdateUser` SQL query to `internal/sqlite/query.sql`
- Run `make sqlc` to generate Go code
- Verify generated code compiles: `go build`

### Phase 3: Minimal Implementation (GREEN)

- Create `handlePutUser` function in `user.go`
- Add `userPutRequestBody` type
- Add validation logic (optional fields)
- Register route in `route()` function with authenticate middleware
- Run test to confirm it passes: `go test -v -run TestHandlePutUser`
- Run all tests: `make test`
- Commit: "feat: implement PUT /api/user"

### Phase 4: Edge Cases & Validation (RED → GREEN)

- Add test for unauthorized access (no token)
- Verify 401 response (should already work via middleware)
- Add test for partial update (only bio)
- Implement partial update logic to make test pass
- Add test for empty request body (no updates)
- Handle empty request body gracefully
- Add test for invalid user ID in token
- Handle user not found scenario
- Run all tests: `make test`
- Commit: "test: add edge case tests for PUT /api/user"

### Phase 5: Refactor (if needed)

- Review for duplication with other user handlers
- Check if response encoding can be shared
- Ensure all tests still pass after each refactoring
- Commit: "refactor: {description}" (if changes made)

### Phase 6: Verification

- Run full test suite: `make test`
- Run linter: `make lint`
- Manual test with curl
- Update this plan's status to "Completed"

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestHandlePutUser

# Run all tests
make test

# Lint check
make lint

# Manual test - First login to get token
TOKEN=$(curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"user":{"email":"jake@jake.jake","password":"jakejake"}}' \
  | jq -r '.token')

# Update user profile
curl -X PUT http://localhost:8080/api/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Token $TOKEN" \
  -d '{
    "user": {
      "email": "jake@jake.jake",
      "bio": "I like to skateboard",
      "image": "https://i.stack.imgur.com/xHWG8.jpg"
    }
  }' | jq
```

## SQL Query to Add

```sql
-- name: UpdateUser :one
UPDATE users
SET
    email = COALESCE(sqlc.narg('email'), email),
    username = COALESCE(sqlc.narg('username'), username),
    password = COALESCE(sqlc.narg('password'), password),
    bio = COALESCE(sqlc.narg('bio'), bio),
    image = COALESCE(sqlc.narg('image'), image)
WHERE id = ?
RETURNING *;
```

## Notes

- All fields in the request are optional
- Only provided fields should be updated
- Must maintain existing authentication patterns
- Response format must match GET /api/user exactly
