# Implementation Plan: DELETE /api/articles/:slug

**Status**: [ ] Not Started / [x] In Progress / [ ] Completed
**PR**: #29

## Context

Implement the `DELETE /api/articles/:slug` endpoint to allow authenticated users to delete their own articles. This endpoint completes the CRUD operations for articles (Create, Read, Update, Delete).

**API Specification Reference**: `@docs/spec.md` lines 196-199

## Methodology

Following Test-Driven Development as described in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Write failing test first
- Minimal implementation to pass
- Refactor only when tests are green

**Testing approach**: Integration tests using real HTTP server and database.

## Feature Requirements

### Endpoint
- **Method**: DELETE
- **Path**: /api/articles/:slug
- **Authentication**: Required (only article author can delete)

### Request Format
No request body required.

### Response Format
- **Success**: 200 OK with no content body (or 204 No Content)
- **Error responses**: Standard error format

### Validation Rules
- Authentication required (401 if not authenticated)
- Authorization required (403 if user is not the article author)
- Article must exist (404 if not found)

### Business Logic
- **Cascading deletes**: Database foreign key constraints handle deletion of:
  - Article tags associations (article_tags table)
  - Favorites (favorites table)
  - Comments will be handled when comment endpoints are implemented
- **Permanent deletion**: No soft delete, article is permanently removed

### Database Operations

The database schema already has CASCADE DELETE constraints:
```sql
-- From schema.sql
FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
```

Need new query in `internal/sqlite/query.sql`:
```sql
-- name: DeleteArticle :exec
DELETE FROM articles WHERE id = ?;
```

## Implementation Steps

### Phase 0: Feature Branch & Plan
- [x] Create feature branch: `git checkout -b feat/delete-api-articles-slug`
- [x] Create this plan document
- [ ] Commit plan: `git add docs/prompts && git commit -m "docs: add plan for DELETE /api/articles/:slug"`
- [ ] Push branch: `git push -u origin feat/delete-api-articles-slug`
- [ ] Create draft PR: `gh pr create --draft --title "feat: implement DELETE /api/articles/:slug" --body "Implements article deletion endpoint. See docs/prompts/2025-10-25-delete-api-articles-slug.md"`
- [ ] Update plan with PR link and commit

### Phase 1: Database Layer
- [ ] Add `DeleteArticle` query to `internal/sqlite/query.sql`
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`
- [ ] Commit: `git add . && git commit -m "feat: add DeleteArticle SQL query" && git push`

### Phase 2: Test First (RED)
- [ ] Add test to `article_test.go`: `TestDeleteArticlesSlug_Success`
- [ ] Test should verify successful deletion (200 or 204 status)
- [ ] Verify article is actually deleted from database
- [ ] Run test to confirm it fails: `go test -v -run TestDeleteArticlesSlug_Success`
- [ ] Commit: `git add . && git commit -m "test: add failing test for DELETE /api/articles/:slug" && git push`

### Phase 3: Minimal Implementation (GREEN)
- [ ] Create `handleDeleteArticlesSlug` function in `article.go`
- [ ] Implement delete logic with authorization check
- [ ] Register route in `route()` function: `mux.Handle("DELETE /api/articles/{slug}", authenticate(handleDeleteArticlesSlug(db), jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestDeleteArticlesSlug_Success`
- [ ] Run all tests: `make test`
- [ ] Commit: `git add . && git commit -m "feat: implement DELETE /api/articles/:slug" && git push`

### Phase 4: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for article not found (404)
- [ ] Implement not found handling to make test pass
- [ ] Commit: `git add . && git commit -m "test: add not found test for DELETE /api/articles/:slug" && git push`
- [ ] Add test for unauthorized deletion (403 - different user)
- [ ] Implement author check to make test pass
- [ ] Commit: `git add . && git commit -m "test: add authorization test for DELETE /api/articles/:slug" && git push`
- [ ] Run all tests: `make test`

### Phase 5: Refactor (if needed)
- [ ] Review code for duplication with other handlers
- [ ] Extract common patterns if found
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit if changes made: `git add . && git commit -m "refactor: {description}" && git push`

### Phase 6: Verification & PR
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (see verification commands below)
- [ ] Update plan status to "Completed"
- [ ] Commit plan update: `git add docs/prompts && git commit -m "docs: mark DELETE /api/articles/:slug as completed" && git push`
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] Request user approval before merging

## Verification Commands

```bash
# Run specific test
go test -v -run TestDeleteArticlesSlug

# Run all tests
make test

# Lint check
make lint

# Manual test - First create an article
curl -X POST http://localhost:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "article": {
      "title": "Article to Delete",
      "description": "Will be deleted",
      "body": "Test body"
    }
  }'

# Then delete it
curl -X DELETE http://localhost:8080/api/articles/article-to-delete \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -v
# Should return 200 OK or 204 No Content

# Verify it's deleted - try to get it
curl -X GET http://localhost:8080/api/articles/article-to-delete \
  -v
# Should return 404 Not Found

# Test unauthorized (different user's token)
curl -X DELETE http://localhost:8080/api/articles/some-article \
  -H "Authorization: Token OTHER_USER_TOKEN" \
  -v
# Should return 403 Forbidden

# Test not found
curl -X DELETE http://localhost:8080/api/articles/nonexistent \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -v
# Should return 404 Not Found
```

## Implementation Notes

- Use similar pattern to `handlePutArticlesSlug` for fetching article and checking authorization
- DELETE should return 200 OK with no body (simpler than 204 for consistency)
- Database CASCADE DELETE will automatically clean up:
  - article_tags entries
  - favorites entries
  - comments (when implemented)
- Check author_id from database matches userID from context for authorization
- Use transaction for consistency (though single DELETE is atomic)
