# GET /api/articles Implementation Plan

## Status & Links

- **Status**: [x] Completed
- **PR**: https://github.com/raeperd/realworld.go/pull/33
- **Created**: 2025-10-25
- **Completed**: 2025-10-25

## Context

Implementing the `GET /api/articles` endpoint to list articles with filtering capabilities.

**From API Specification** (`@docs/spec.md`):
- Endpoint: `GET /api/articles`
- Returns most recent articles globally by default
- Query Parameters:
  - `tag` - Filter by tag (e.g., `?tag=AngularJS`)
  - `author` - Filter by author username (e.g., `?author=jake`)
  - `favorited` - Filter by user who favorited (e.g., `?favorited=jake`)
  - `limit` - Limit number of articles (default: 20)
  - `offset` - Skip number of articles (default: 0)
- Authentication: Optional
- Returns: Multiple articles ordered by most recent first
- Response format: Does NOT include article body (performance optimization since 2024/08/16)

## Methodology

This implementation follows Test-Driven Development as defined in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Write failing test first
- Minimal implementation to pass
- Refactor only when tests pass
- Separate commits for structural vs behavioral changes

## Feature Requirements

### Request Format
```
GET /api/articles?tag=dragons&author=jake&favorited=john&limit=10&offset=0
Authorization: Token jwt.token.here (optional)
```

### Response Format
```json
{
  "articles": [{
    "slug": "how-to-train-your-dragon",
    "title": "How to train your dragon",
    "description": "Ever wonder how?",
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
  }],
  "articlesCount": 1
}
```

**Note**: Response does NOT include `body` field for performance reasons.

### Authentication
- Optional authentication via `authenticateOptional` middleware
- If authenticated: `favorited` and `following` fields reflect current user's state
- If not authenticated: `favorited` and `following` are always `false`

### Query Parameters & Validation
- `tag`: Filter by tag name (exact match)
- `author`: Filter by author username (exact match)
- `favorited`: Filter by username who favorited the articles
- `limit`: Integer, default 20, must be positive
- `offset`: Integer, default 0, must be non-negative

### Database Schema Requirements
- ✅ `articles` table exists with slug, title, description, author_id, created_at, updated_at
- ✅ `users` table for author and favorited user lookup
- ✅ `tags` and `article_tags` tables for tag filtering
- ✅ `favorites` table for favorited filtering and favoritesCount
- ✅ `follows` table for following status

### Error Handling
- No specific validation errors expected (all query params are optional)
- Empty results return `{"articles": [], "articlesCount": 0}`

## Implementation Steps

### Phase 0: Create Plan and Draft PR (MANDATORY FIRST STEP)

⚠️ **DO NOT PROCEED TO PHASE 1 WITHOUT COMPLETING THIS PHASE!**

- [x] Create feature branch: `feat/api-get-articles`
- [ ] Create plan document: `docs/prompts/2025-10-25-get-api-articles.md`
- [ ] Commit plan as first commit
- [ ] Push to create remote branch
- [ ] Create DRAFT PR with `gh pr create --draft`
- [ ] Update plan with PR link
- [ ] Commit and push plan update

**Verification before proceeding**:
- ✅ Branch exists and pushed to remote
- ✅ Plan document committed (first commit in branch)
- ✅ Draft PR created (visible on GitHub)
- ✅ Plan updated with PR link and pushed
- ✅ PR shows plan document as first commit

### Phase 1: Test First (RED)

- [ ] Create test in `article_test.go` for basic list articles (no filters)
- [ ] Test creates sample articles via POST /api/articles
- [ ] Test verifies response structure matches spec (no body field)
- [ ] Test verifies articles ordered by most recent first
- [ ] Run test to confirm it fails: `go test -v -run TestHandleGetArticles`
- [ ] Commit: "test: add failing test for GET /api/articles"
- [ ] Push to trigger CI (should show RED/failing status)

### Phase 2: Database Layer

- [ ] Add SQL query `ListArticles` to `internal/sqlite/query.sql`
  - Select articles with author info, tags, favorites count
  - JOIN users for author details
  - LEFT JOIN for tags aggregation
  - LEFT JOIN for favorites count
  - ORDER BY created_at DESC
  - Support LIMIT and OFFSET
- [ ] Add conditional WHERE clauses for filtering:
  - Tag filter: JOIN article_tags and tags
  - Author filter: JOIN users on username
  - Favorited filter: JOIN favorites and users on username
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles
- [ ] Commit: "feat: add database queries for listing articles"
- [ ] Push to trigger CI

### Phase 3: Minimal Implementation (GREEN)

- [ ] Create `handleGetArticles` function in `article.go`
- [ ] Add request parsing for query parameters (tag, author, favorited, limit, offset)
- [ ] Set defaults: limit=20, offset=0
- [ ] Call database query with filters
- [ ] Build response with articles array (excluding body field)
- [ ] Include articlesCount in response
- [ ] Register route in `route()`: `mux.Handle("GET /api/articles", authenticateOptional(handleGetArticles(db), jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestHandleGetArticles`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement GET /api/articles basic listing"
- [ ] Push to trigger CI (should show GREEN/passing status)

### Phase 4: Edge Cases & Filtering Tests (RED → GREEN)

- [ ] Add test for tag filtering
- [ ] Add test for author filtering
- [ ] Add test for favorited filtering
- [ ] Add test for limit parameter
- [ ] Add test for offset parameter
- [ ] Add test for combined filters (tag + author)
- [ ] Add test for authenticated user (favorited and following fields)
- [ ] Add test for empty results
- [ ] Implement filtering logic to make tests pass
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add filtering and edge case tests for GET /api/articles"
- [ ] Push to trigger CI

### Phase 5: Refactor (if needed)

- [ ] Review code for duplication with GET /api/articles/:slug
- [ ] Consider extracting common article response building logic
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit: "refactor: extract article response building" (if changes made)
- [ ] Push to trigger CI

### Phase 6: Verification & Finalize PR

- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl:
  - List all articles
  - Filter by tag
  - Filter by author
  - Filter by favorited
  - Test limit and offset
  - Test with authentication token
- [ ] Update this plan's status to "Completed"
- [ ] Update PR description with implementation summary
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] **DO NOT merge** - Wait for review

## Verification Commands

```bash
# Run specific test
go test -v -run TestHandleGetArticles

# Run all tests
make test

# Lint check
make lint

# Manual testing examples

# List all articles (default limit 20, offset 0)
curl http://localhost:8080/api/articles

# Filter by tag
curl "http://localhost:8080/api/articles?tag=dragons"

# Filter by author
curl "http://localhost:8080/api/articles?author=jake"

# Filter by favorited user
curl "http://localhost:8080/api/articles?favorited=john"

# Pagination
curl "http://localhost:8080/api/articles?limit=10&offset=5"

# Combined filters
curl "http://localhost:8080/api/articles?tag=dragons&author=jake&limit=5"

# With authentication (shows favorited and following status)
curl http://localhost:8080/api/articles \
  -H "Authorization: Token <jwt-token>"
```

## Implementation Notes

- **Important**: Response must NOT include `body` field (spec change from 2024/08/16)
- Default limit is 20, default offset is 0
- Articles ordered by most recent first (created_at DESC)
- Authentication is optional - affects `favorited` and `following` fields only
- Empty results should return valid response with empty array
- All query parameters are optional - no validation errors needed
