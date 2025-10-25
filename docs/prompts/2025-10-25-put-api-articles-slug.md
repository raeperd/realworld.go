# Implementation Plan: PUT /api/articles/:slug

**Status**: [x] Completed
**PR**: #28

## Context

Implement the `PUT /api/articles/:slug` endpoint to allow authenticated users to update their own articles. This endpoint completes the basic CRUD operations for articles (Create, Read, Update).

**API Specification Reference**: `@docs/spec.md` lines 176-193

## Methodology

Following Test-Driven Development as described in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Write failing test first
- Minimal implementation to pass
- Refactor only when tests are green

**Testing approach**: Integration tests using real HTTP server and database.

## Feature Requirements

### Endpoint
- **Method**: PUT
- **Path**: /api/articles/:slug
- **Authentication**: Required (only article author can update)

### Request Format
```json
{
  "article": {
    "title": "Did you train your dragon?",
    "description": "Updated description",
    "body": "Updated body"
  }
}
```

### Response Format
Returns the updated Article (same format as GET /api/articles/:slug)

### Validation Rules
- Authentication required (401 if not authenticated)
- Authorization required (403 if user is not the article author)
- Article must exist (404 if not found)
- All fields are optional (partial update support)
- Title validation: if provided, must not be empty
- Description validation: if provided, must not be empty
- Body validation: if provided, must not be empty

### Business Logic
- **Slug regeneration**: When title is updated, slug must be regenerated
- **Partial updates**: Only provided fields should be updated
- **Timestamp**: `updated_at` is automatically updated by database trigger
- **Tags**: Not supported in this endpoint (tags are set only at creation)

### Database Operations
Need new query in `internal/sqlite/query.sql`:
```sql
-- name: UpdateArticle :one
UPDATE articles
SET
    slug = COALESCE(sqlc.narg('slug'), slug),
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    body = COALESCE(sqlc.narg('body'), body)
WHERE id = sqlc.arg('id')
RETURNING *;
```

## Implementation Steps

### Phase 0: Feature Branch & Plan
- [x] Create feature branch: `git checkout -b feat/put-api-articles-slug`
- [x] Create this plan document
- [ ] Commit plan: `git add docs/prompts && git commit -m "docs: add plan for PUT /api/articles/:slug"`
- [ ] Push branch: `git push -u origin feat/put-api-articles-slug`
- [ ] Create draft PR: `gh pr create --draft --title "feat: implement PUT /api/articles/:slug" --body "Implements article update endpoint. See docs/prompts/2025-10-25-put-api-articles-slug.md"`
- [ ] Update plan with PR link and commit

### Phase 1: Database Layer
- [ ] Add `UpdateArticle` query to `internal/sqlite/query.sql`
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`
- [ ] Commit: `git add . && git commit -m "feat: add UpdateArticle SQL query" && git push`

### Phase 2: Test First (RED)
- [ ] Add test to `article_test.go`: `TestHandlePutArticlesSlug`
- [ ] Test should verify successful update with title change (slug regeneration)
- [ ] Run test to confirm it fails: `go test -v -run TestHandlePutArticlesSlug`
- [ ] Commit: `git add . && git commit -m "test: add failing test for PUT /api/articles/:slug" && git push`

### Phase 3: Minimal Implementation (GREEN)
- [ ] Create `handlePutArticlesSlug` function in `article.go`
- [ ] Add request/response types (reuse existing `articleResponseBody`)
- [ ] Implement basic update logic with slug regeneration
- [ ] Register route in `route()` function: `mux.Handle("PUT /api/articles/{slug}", authenticate(handlePutArticlesSlug(db), jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestHandlePutArticlesSlug`
- [ ] Run all tests: `make test`
- [ ] Commit: `git add . && git commit -m "feat: implement PUT /api/articles/:slug" && git push`

### Phase 4: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for article not found (404)
- [ ] Implement not found handling to make test pass
- [ ] Commit: `git add . && git commit -m "test: add not found test for PUT /api/articles/:slug" && git push`
- [ ] Add test for unauthorized update (403 - different user)
- [ ] Implement author check to make test pass
- [ ] Commit: `git add . && git commit -m "test: add authorization test for PUT /api/articles/:slug" && git push`
- [ ] Add test for partial update (only some fields)
- [ ] Verify partial update works correctly
- [ ] Commit if changes needed: `git add . && git commit -m "test: add partial update test" && git push`
- [ ] Add test for empty field validation
- [ ] Implement validation to make test pass
- [ ] Commit: `git add . && git commit -m "test: add validation tests" && git push`
- [ ] Run all tests: `make test`

### Phase 5: Refactor (if needed)
- [ ] Review code for duplication with POST handler
- [ ] Extract common patterns if found (e.g., article response building)
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit if changes made: `git add . && git commit -m "refactor: extract common article response builder" && git push`

### Phase 6: Verification & PR
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (see verification commands below)
- [ ] Update plan status to "Completed"
- [ ] Commit plan update: `git add docs/prompts && git commit -m "docs: mark PUT /api/articles/:slug as completed" && git push`
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] Request user review before merging

## Verification Commands

```bash
# Run specific test
go test -v -run TestHandlePutArticlesSlug

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
      "title": "Original Title",
      "description": "Original description",
      "body": "Original body",
      "tagList": ["test"]
    }
  }'

# Then update it
curl -X PUT http://localhost:8080/api/articles/original-title \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "article": {
      "title": "Updated Title"
    }
  }'

# Verify slug changed to "updated-title" and title updated

# Test partial update
curl -X PUT http://localhost:8080/api/articles/updated-title \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "article": {
      "body": "New body content"
    }
  }'

# Test unauthorized (different user's token)
curl -X PUT http://localhost:8080/api/articles/updated-title \
  -H "Content-Type: application/json" \
  -H "Authorization: Token OTHER_USER_TOKEN" \
  -d '{
    "article": {
      "title": "Hacked"
    }
  }'
# Should return 403 Forbidden

# Test not found
curl -X PUT http://localhost:8080/api/articles/nonexistent \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "article": {
      "title": "Doesnt matter"
    }
  }'
# Should return 404 Not Found
```

## Implementation Notes

- Use similar pattern to `handleGetArticlesSlug` for fetching article
- Reuse `generateSlug()` function for slug regeneration
- Use transaction to ensure atomic update
- Follow the pattern from `handlePutUser` for partial updates with COALESCE
- Return full article response including author profile and tags
- Check author_id from database matches userID from context for authorization
