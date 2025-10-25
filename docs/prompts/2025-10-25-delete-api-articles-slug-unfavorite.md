# DELETE /api/articles/:slug/favorite Implementation

## Status & Links

- [ ] In Progress
- PR: (to be created)

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
- [ ] Commit plan as first commit
- [ ] Push to remote
- [ ] Create DRAFT PR with `gh pr create --draft`
- [ ] Update plan with PR link
- [ ] Commit and push plan update

### Phase 1: Test First (RED)

- [ ] Add test `TestDeleteArticlesSlugFavorite_Success` in `article_test.go`
- [ ] Test happy path: unfavorite a favorited article and verify response
- [ ] Run test to confirm it fails: `go test -v -run TestDeleteArticlesSlugFavorite`
- [ ] Commit: "test: add failing test for DELETE /api/articles/:slug/favorite"
- [ ] Push to trigger CI

### Phase 2: Database Layer

- [ ] Add query to `internal/sqlite/query.sql`: `DeleteFavorite` - DELETE from favorites
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles

### Phase 3: Minimal Implementation (GREEN)

- [ ] Create `handleDeleteArticlesSlugFavorite` function in `article.go`
- [ ] Extract slug from path parameter
- [ ] Get authenticated user ID from context
- [ ] Use transaction for consistency
- [ ] Look up article by slug
- [ ] Delete favorite record
- [ ] Query article with updated favorited status
- [ ] Return article response
- [ ] Register route in `route()` function in `main.go`
- [ ] Run test: `go test -v -run TestDeleteArticlesSlugFavorite`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement DELETE /api/articles/:slug/favorite"
- [ ] Push to trigger CI

### Phase 4: Edge Cases & Validation (RED → GREEN)

- [ ] Add test for not found article
- [ ] Verify 404 handling works
- [ ] Add test for unauthorized access (no token)
- [ ] Verify middleware handles this
- [ ] Add test for idempotency (unfavoriting twice)
- [ ] Verify no error occurs
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add edge case tests for DELETE /api/articles/:slug/favorite"
- [ ] Push to trigger CI

### Phase 5: Refactor (if needed)

- [ ] Review code for duplication with POST favorite
- [ ] Consider extracting article response logic if duplicated 3+ times
- [ ] Ensure all tests still pass after refactoring
- [ ] Commit: "refactor: {description}" (if changes made)
- [ ] Push to trigger CI

### Phase 6: Verification & Finalize PR

- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Update plan status to "Completed"
- [ ] Update PR description
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
