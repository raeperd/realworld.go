# DELETE /api/articles/:slug/favorite Implementation

## Status & Links

- [x] Completed
- PR: https://github.com/raeperd/realworld.go/pull/36

## Context

Implementing the RealWorld API endpoint for unfavoriting articles as specified in `@docs/spec.md`.

### API Specification

**Endpoint**: `DELETE /api/articles/:slug/favorite`

**Authentication**: Required

**Request**: No additional parameters required

**Response**: Returns the unfavorited Article with `favorited: false` and decremented `favoritesCount`

```json
{
  "article": {
    "slug": "how-to-train-your-dragon",
    "title": "How to train your dragon",
    "description": "Ever wonder how?",
    "body": "It takes a Jacobian",
    "tagList": ["dragons", "training"],
    "createdAt": "2016-02-18T03:22:56.637Z",
    "updatedAt": "2016-02-18T03:48:35.824Z",
    "favorited": false,
    "favoritesCount": 0,
    "author": {
      "username": "jake",
      "bio": "I work at statefarm",
      "image": "https://i.stack.imgur.com/xHWG8.jpg",
      "following": false
    }
  }
}
```

## Methodology

Following Test-Driven Development as described in `@docs/prompts/TDD.md`.

## Feature Requirements

### Request/Response Format

- **Method**: DELETE
- **Path**: `/api/articles/:slug/favorite`
- **Auth**: JWT token required (via `authenticate` middleware)
- **Request Body**: None
- **Response**: Single article with `favorited: false` and updated `favoritesCount`

### Validation & Error Handling

- **401 Unauthorized**: Missing or invalid authentication token
- **404 Not Found**: Article with given slug doesn't exist
- **Idempotency**: Unfavoriting an already-unfavorited article should succeed (no error)

### Database Operations

- Delete from `favorites` table: `WHERE user_id = ? AND article_id = ?`
- Query article with favorited status and count for the authenticated user
- Use transaction for consistency

## Implementation Steps

### Phase 0: Create Plan and Draft PR ✅

- [x] Create feature branch: `feat/api-delete-articles-slug-unfavorite`
- [x] Create plan document
- [x] Commit plan as first commit
- [x] Push to remote
- [x] Create DRAFT PR with `gh pr create --draft`
- [x] Update plan with PR link
- [x] Commit and push plan update

### Phase 1: Test First (RED) ✅

- [x] Add test `TestDeleteArticlesSlugFavorite_Success` in `article_test.go`
- [x] Test happy path: unfavorite a favorited article and verify response
- [x] Run test to confirm it fails: `go test -v -run TestDeleteArticlesSlugFavorite`
- [x] Commit: "test: add failing test for DELETE /api/articles/:slug/favorite"
- [x] Push to trigger CI

### Phase 2: Database Layer ✅

- [x] Add query to `internal/sqlite/query.sql`: `DeleteFavorite` - DELETE from favorites
- [x] Run `make generate` to generate Go code
- [x] Verify generated code compiles

### Phase 3: Minimal Implementation (GREEN) ✅

- [x] Create `handleDeleteArticlesSlugFavorite` function in `article.go`
- [x] Extract slug from path parameter
- [x] Get authenticated user ID from context
- [x] Use transaction for consistency
- [x] Look up article by slug
- [x] Delete favorite record
- [x] Query article with updated favorited status
- [x] Return article response
- [x] Register route in `route()` function in `main.go`
- [x] Run test: `go test -v -run TestDeleteArticlesSlugFavorite`
- [x] Run all tests: `make test`
- [x] Commit: "feat: implement DELETE /api/articles/:slug/favorite"
- [x] Push to trigger CI

### Phase 4: Edge Cases & Validation (RED → GREEN) ✅

- [x] Add test for not found article
- [x] Verify 404 handling works
- [x] Add test for unauthorized access (no token)
- [x] Verify middleware handles this
- [x] Add test for idempotency (unfavoriting twice)
- [x] Verify no error occurs
- [x] Run all tests: `make test`
- [x] Commit: "test: add edge case tests for DELETE /api/articles/:slug/favorite"
- [x] Push to trigger CI

### Phase 5: Refactor ✅

- [x] Review code for duplication with POST favorite
- [x] Intentional duplication - added nolint comments (similar to profile handlers)
- [x] Ensure all tests still pass after changes
- [x] Commit: "chore: add nolint comments for intentional duplication in favorite handlers"
- [x] Push to trigger CI

### Phase 6: Verification & Finalize PR ✅

- [x] Run full test suite: `make test` - All tests passing
- [x] Run linter: `make lint` - No issues
- [ ] Update plan status to "Completed"
- [ ] Mark PR as ready: `gh pr ready`

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestDeleteArticlesSlugFavorite

# Run all tests
make test

# Lint check
make lint

# Manual test
# First, register and login to get a token
TOKEN=$(curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"test","email":"test@test.com","password":"password"}}' \
  | jq -r '.user.token')

# Create an article
curl -X POST http://localhost:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Token $TOKEN" \
  -d '{"article":{"title":"Test Article","description":"Test","body":"Test body","tagList":["test"]}}'

# Favorite the article first
curl -X POST http://localhost:8080/api/articles/test-article/favorite \
  -H "Authorization: Token $TOKEN" | jq

# Unfavorite the article
curl -X DELETE http://localhost:8080/api/articles/test-article/favorite \
  -H "Authorization: Token $TOKEN" | jq

# Verify favorited is false and favoritesCount is 0
```
