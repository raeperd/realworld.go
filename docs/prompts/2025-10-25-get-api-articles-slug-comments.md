# Implementation Plan: GET /api/articles/:slug/comments

**Status**: [x] Completed
**PR**: #31

## Context

Implement the `GET /api/articles/:slug/comments` endpoint to retrieve all comments for a specific article. This follows the POST comments endpoint (#30) and completes the read functionality for comments.

**API Specification Reference**: `@docs/spec.md` lines 219-223

## Methodology

Following Test-Driven Development as described in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Integration tests using real HTTP server and database
- Write failing test first, minimal implementation, then refactor

## Feature Requirements

### Endpoint
- **Method**: GET
- **Path**: /api/articles/:slug/comments
- **Authentication**: Optional (affects `following` field in author profiles)

### Request Format
No request body needed. Path parameter: `slug` (article slug)

### Response Format
```json
{
  "comments": [{
    "id": 1,
    "createdAt": "2016-02-18T03:22:56.637Z",
    "updatedAt": "2016-02-18T03:22:56.637Z",
    "body": "It takes a Jacobian",
    "author": {
      "username": "jake",
      "bio": "I work at statefarm",
      "image": "https://i.stack.imgur.com/xHWG8.jpg",
      "following": false
    }
  }]
}
```

### Validation Rules
- Article must exist (404 if article not found)
- Returns empty array if article has no comments
- If authenticated, `following` reflects current user's relationship to comment authors
- If not authenticated, `following` is always false

### Business Logic
- Returns all comments for the specified article
- Comments ordered by creation date (most recent first)
- Each comment includes full author profile
- Following status calculated per comment author based on current user

### Database Operations

**Query needed** (`internal/sqlite/query.sql`):
```sql
-- name: GetCommentsByArticleSlug :many
SELECT
    c.id,
    c.body,
    c.created_at,
    c.updated_at,
    c.author_id,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM comments c
JOIN articles a ON c.article_id = a.id
JOIN users u ON c.author_id = u.id
WHERE a.slug = ?
ORDER BY c.created_at DESC;
```

Note: The comments table and CreateComment query already exist from PR #30.

## Implementation Steps

### Phase 0: Feature Branch & Plan
- [x] Create feature branch: `git checkout -b feat/get-api-articles-slug-comments`
- [ ] Create this plan document
- [ ] Commit plan: `git add docs/prompts && git commit -m "docs: add plan for GET /api/articles/:slug/comments"`
- [ ] Push branch: `git push -u origin feat/get-api-articles-slug-comments`
- [ ] Create draft PR: `gh pr create --draft --title "feat: implement GET /api/articles/:slug/comments" --body "Implements comment retrieval endpoint. See docs/prompts/2025-10-25-get-api-articles-slug-comments.md"`
- [ ] Update plan with PR link and commit
- [ ] Push update: `git push`

### Phase 1: Database Query
- [ ] Add `GetCommentsByArticleSlug` query to `internal/sqlite/query.sql`
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`
- [ ] Commit: `git add . && git commit -m "feat: add GetCommentsByArticleSlug query" && git push`

### Phase 2: Test First (RED)
- [ ] Add to existing `comment_test.go` (or create if doesn't exist)
- [ ] Write `TestGetArticlesSlugComments_Success` test
  - Create test article
  - Create 2-3 test comments
  - Verify 200 status, array structure, comment data, author info
- [ ] Run test to confirm it fails: `go test -v -run TestGetArticlesSlugComments_Success`
- [ ] Commit: `git add . && git commit -m "test: add failing test for GET /api/articles/:slug/comments" && git push`

### Phase 3: Minimal Implementation (GREEN)
- [ ] Add `handleGetArticlesSlugComments` function to `comment.go`
- [ ] Define response types (reuse comment/author types if exist)
- [ ] Get comments by article slug from database
- [ ] For each comment, determine following status (if authenticated)
- [ ] Build and return response with comments array
- [ ] Register route: `mux.HandleFunc("GET /api/articles/{slug}/comments", handleGetArticlesSlugComments(db, jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestGetArticlesSlugComments_Success`
- [ ] Run all tests: `make test`
- [ ] Commit: `git add . && git commit -m "feat: implement GET /api/articles/:slug/comments" && git push`

### Phase 4: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for article not found (404)
- [ ] Verify error handling for not found
- [ ] Commit: `git add . && git commit -m "test: add not found test for GET comments" && git push`
- [ ] Add test for empty comments (200 with empty array)
- [ ] Verify empty array is returned correctly
- [ ] Commit: `git add . && git commit -m "test: add empty comments test" && git push`
- [ ] Add test for following status when authenticated
- [ ] Verify following field is calculated correctly
- [ ] Commit: `git add . && git commit -m "test: add following status test for comments" && git push`
- [ ] Run all tests: `make test`

### Phase 5: Refactor (if needed)
- [ ] Review code for duplication with POST comments handler
- [ ] Extract common patterns if found in 3+ places
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit if changes made: `git add . && git commit -m "refactor: {description}" && git push`

### Phase 6: Verification & PR
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (see verification commands below)
- [ ] Update plan status to "Completed"
- [ ] Commit plan update: `git add docs/prompts && git commit -m "docs: mark GET /api/articles/:slug/comments as completed" && git push`
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] Request user approval before merging

## Verification Commands

```bash
# Run specific test
go test -v -run TestGetArticlesSlugComments

# Run all tests
make test

# Lint check
make lint

# Manual test - Get comments without authentication
curl -X GET http://localhost:8080/api/articles/test-article/comments
# Should return 200 OK with comments array

# Manual test - Get comments with authentication
curl -X GET http://localhost:8080/api/articles/test-article/comments \
  -H "Authorization: Token YOUR_JWT_TOKEN"
# Should return 200 OK with following status calculated

# Test article not found
curl -X GET http://localhost:8080/api/articles/nonexistent/comments
# Should return 404 Not Found

# Test empty comments
curl -X GET http://localhost:8080/api/articles/article-with-no-comments/comments
# Should return 200 OK with empty array []
```

## Implementation Notes

- Reuse comment and author response types from POST comments handler
- Handle optional authentication using `optionalAuthenticate` middleware or check for auth in handler
- Return empty array (not null) when no comments exist
- Calculate following status only if user is authenticated
- Use existing `encodeResponse` helper for consistent JSON formatting
- Order comments by created_at DESC for most recent first
- Handle sql.ErrNoRows when article not found (return 404)
