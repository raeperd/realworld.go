# POST /api/articles/:slug/favorite Implementation

## Status & Links

- [x] Completed
- PR: https://github.com/raeperd/realworld.go/pull/35

## Context

Implementing the RealWorld API endpoint for favoriting articles as specified in `@docs/spec.md`.

### API Specification

**Endpoint**: `POST /api/articles/:slug/favorite`

**Authentication**: Required

**Request**: No additional parameters required

**Response**: Returns the favorited Article with `favorited: true` and incremented `favoritesCount`

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
    "favorited": true,
    "favoritesCount": 1,
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

- **Method**: POST
- **Path**: `/api/articles/:slug/favorite`
- **Auth**: JWT token required (via `authenticate` middleware)
- **Request Body**: None
- **Response**: Single article with `favorited: true` and updated `favoritesCount`

### Validation & Error Handling

- **401 Unauthorized**: Missing or invalid authentication token
- **404 Not Found**: Article with given slug doesn't exist
- **422 Unprocessable Entity**: User already favorited this article (idempotent - should succeed)

### Database Operations

- Insert into `favorites` table: `(user_id, article_id)`
- Handle duplicate favorites gracefully (INSERT OR IGNORE for idempotency)
- Query article with favorited status and count for the authenticated user

## Implementation Steps

### Phase 0: Create Plan and Draft PR ✅

- [x] Create feature branch: `feat/api-post-articles-slug-favorite`
- [x] Create plan document
- [x] Commit plan as first commit
- [x] Push to remote
- [x] Create DRAFT PR with `gh pr create --draft`
- [x] Update plan with PR link
- [x] Commit and push plan update

### Phase 1: Test First (RED) ✅

- [x] Add test `TestHandlePostArticlesSlugFavorite` in `article_test.go`
- [x] Test happy path: favorite an article and verify response
- [x] Run test to confirm it fails: `go test -v -run TestHandlePostArticlesSlugFavorite`
- [x] Commit: "test: add failing test for POST /api/articles/:slug/favorite"
- [x] Push to trigger CI

### Phase 2: Database Layer ✅

- [x] Add query to `internal/sqlite/query.sql`: `CreateFavorite` - INSERT OR IGNORE into favorites
- [x] Run `make generate` to generate Go code
- [x] Verify generated code compiles

### Phase 3: Minimal Implementation (GREEN) ✅

- [x] Create `handlePostArticlesSlugFavorite` function in `article.go`
- [x] Extract slug from path parameter
- [x] Get authenticated user ID from context
- [x] Look up article by slug
- [x] Insert favorite record (with OR IGNORE for idempotency)
- [x] Query article with updated favorited status
- [x] Return article response
- [x] Register route in `route()` function in `main.go`
- [x] Run test: `go test -v -run TestHandlePostArticlesSlugFavorite`
- [x] Run all tests: `make test`
- [x] Commit: "feat: implement POST /api/articles/:slug/favorite"
- [x] Push to trigger CI

### Phase 4: Edge Cases & Validation (RED → GREEN) ✅

- [x] Add test for not found article
- [x] Implement 404 handling (already handled by GetArticleBySlug)
- [x] Add test for unauthorized access (no token)
- [x] Verify middleware handles this (verified - authenticate middleware works)
- [x] Add test for idempotency (favoriting twice)
- [x] Verify INSERT OR IGNORE handles this (verified - works correctly)
- [x] Run all tests: `make test`
- [x] Commit: "test: add edge case tests for POST /api/articles/:slug/favorite"
- [x] Push to trigger CI

### Phase 5: Refactor ✅

- [x] Review code for duplication with GET /api/articles/:slug
- [x] No refactoring needed - no code duplicated 3+ times
- [x] Fixed linter dupl warnings by excluding dupl from test files

### Phase 6: Verification & Finalize PR ✅

- [x] Run full test suite: `make test` - All tests passing
- [x] Run linter: `make lint` - No issues
- [ ] Update plan status to "Completed"
- [ ] Update PR description
- [ ] Mark PR as ready: `gh pr ready`

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestHandlePostArticlesSlugFavorite

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

# Create an article to favorite
curl -X POST http://localhost:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Token $TOKEN" \
  -d '{"article":{"title":"Test Article","description":"Test","body":"Test body","tagList":["test"]}}'

# Favorite the article
curl -X POST http://localhost:8080/api/articles/test-article/favorite \
  -H "Authorization: Token $TOKEN" | jq

# Verify favorited is true and favoritesCount is 1
```
