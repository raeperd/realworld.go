# Implementation Plan: DELETE /api/articles/:slug/comments/:id

**Status**: [x] Completed
**PR**: #32

## Context

Implement the `DELETE /api/articles/:slug/comments/:id` endpoint to allow authenticated users to delete their own comments. This completes the comment CRUD operations after POST comments (#30) and GET comments (#31).

**API Specification Reference**: `@docs/spec.md` lines 226-230

## Methodology

Following Test-Driven Development as described in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Integration tests using real HTTP server and database
- Write failing test first, minimal implementation, then refactor

## Feature Requirements

### Endpoint
- **Method**: DELETE
- **Path**: /api/articles/:slug/comments/:id
- **Authentication**: Required (user must be authenticated)

### Request Format
No request body needed. Path parameters:
- `slug` (article slug)
- `id` (comment ID)

### Response Format
No response body (204 No Content on success)

### Validation Rules
- User must be authenticated (401 if not authenticated)
- Article must exist (404 if article not found)
- Comment must exist (404 if comment not found)
- User must be the comment author (403 if not the author)

### Business Logic
- Only the comment author can delete their comment
- Deleting a comment removes it from the database
- Returns 204 No Content on successful deletion
- Returns 404 if article or comment doesn't exist
- Returns 403 if user is not the comment author

### Database Operations

**Query needed** (`internal/sqlite/query.sql`):
```sql
-- name: GetCommentByID :one
SELECT id, body, article_id, author_id, created_at, updated_at
FROM comments
WHERE id = ?;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = ?;
```

Note: The comments table already exists from PR #30.

## Implementation Steps

### Phase 0: Feature Branch & Plan
- [x] Create feature branch: `git checkout -b feat/api-delete-articles-slug-comments-id`
- [ ] Create this plan document
- [ ] Commit plan: `git add docs/prompts && git commit -m "docs: add plan for DELETE /api/articles/:slug/comments/:id"`
- [ ] Push branch: `git push -u origin feat/api-delete-articles-slug-comments-id`
- [ ] Create draft PR: `gh pr create --draft --title "feat: implement DELETE /api/articles/:slug/comments/:id" --body "Implements comment deletion endpoint. See docs/prompts/2025-10-25-delete-api-articles-slug-comments-id.md"`
- [ ] Update plan with PR link and commit
- [ ] Push update: `git push`

### Phase 1: Database Queries
- [ ] Add `GetCommentByID` query to `internal/sqlite/query.sql`
- [ ] Add `DeleteComment` query to `internal/sqlite/query.sql`
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`
- [ ] Commit: `git add . && git commit -m "feat: add GetCommentByID and DeleteComment queries" && git push`

### Phase 2: Test First (RED)
- [ ] Add to existing `comment_test.go`
- [ ] Write `TestDeleteArticlesSlugCommentsID_Success` test
  - Create test user and article
  - Create test comment
  - Attempt to delete comment as the author
  - Verify 204 status
  - Verify comment no longer exists in database
- [ ] Run test to confirm it fails: `go test -v -run TestDeleteArticlesSlugCommentsID_Success`
- [ ] Commit: `git add . && git commit -m "test: add failing test for DELETE /api/articles/:slug/comments/:id" && git push`

### Phase 3: Minimal Implementation (GREEN)
- [ ] Add `handleDeleteArticlesSlugCommentsID` function to `comment.go`
- [ ] Get comment ID from path parameter
- [ ] Verify article exists by slug
- [ ] Fetch comment by ID from database
- [ ] Verify comment belongs to the article
- [ ] Verify current user is the comment author
- [ ] Delete comment from database
- [ ] Return 204 No Content
- [ ] Register route: `mux.Handle("DELETE /api/articles/{slug}/comments/{id}", authenticate(handleDeleteArticlesSlugCommentsID(db), jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestDeleteArticlesSlugCommentsID_Success`
- [ ] Run all tests: `make test`
- [ ] Commit: `git add . && git commit -m "feat: implement DELETE /api/articles/:slug/comments/:id" && git push`

### Phase 4: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for unauthorized (no auth token) - 401
- [ ] Verify unauthorized handling
- [ ] Commit: `git add . && git commit -m "test: add unauthorized test for DELETE comment" && git push`
- [ ] Add test for article not found - 404
- [ ] Verify article not found handling
- [ ] Commit: `git add . && git commit -m "test: add article not found test for DELETE comment" && git push`
- [ ] Add test for comment not found - 404
- [ ] Verify comment not found handling
- [ ] Commit: `git add . && git commit -m "test: add comment not found test for DELETE comment" && git push`
- [ ] Add test for forbidden (user not comment author) - 403
- [ ] Verify forbidden handling
- [ ] Commit: `git add . && git commit -m "test: add forbidden test for DELETE comment" && git push`
- [ ] Run all tests: `make test`

### Phase 5: Refactor (if needed)
- [ ] Review code for duplication with other delete handlers
- [ ] Extract common patterns if found in 3+ places
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit if changes made: `git add . && git commit -m "refactor: {description}" && git push`

### Phase 6: Verification & PR
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (see verification commands below)
- [ ] Update plan status to "Completed"
- [ ] Commit plan update: `git add docs/prompts && git commit -m "docs: mark DELETE /api/articles/:slug/comments/:id as completed" && git push`
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] Request user approval before merging

## Verification Commands

```bash
# Run specific test
go test -v -run TestDeleteArticlesSlugCommentsID

# Run all tests
make test

# Lint check
make lint

# Manual test - Delete comment as author (success)
curl -X DELETE http://localhost:8080/api/articles/test-article/comments/1 \
  -H "Authorization: Token YOUR_JWT_TOKEN"
# Should return 204 No Content

# Test unauthorized (no token)
curl -X DELETE http://localhost:8080/api/articles/test-article/comments/1
# Should return 401 Unauthorized

# Test article not found
curl -X DELETE http://localhost:8080/api/articles/nonexistent/comments/1 \
  -H "Authorization: Token YOUR_JWT_TOKEN"
# Should return 404 Not Found

# Test comment not found
curl -X DELETE http://localhost:8080/api/articles/test-article/comments/999 \
  -H "Authorization: Token YOUR_JWT_TOKEN"
# Should return 404 Not Found

# Test forbidden (delete another user's comment)
curl -X DELETE http://localhost:8080/api/articles/test-article/comments/1 \
  -H "Authorization: Token OTHER_USER_JWT_TOKEN"
# Should return 403 Forbidden
```

## Implementation Notes

- Use `authenticate` middleware to require authentication
- Extract comment ID from path parameter using `r.PathValue("id")`
- Convert comment ID string to int64
- Check both article existence and comment existence
- Verify comment belongs to the specified article (security check)
- Verify current user is the comment author (authorization check)
- Return 204 No Content on successful deletion (no response body)
- Use existing `encodeErrorResponse` helper for error responses
- Follow same authorization pattern as DELETE /api/articles/:slug
