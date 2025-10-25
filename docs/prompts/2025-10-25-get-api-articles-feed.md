# GET /api/articles/feed Implementation Plan

## Status & Links

- **Status**: [x] In Progress
- **PR**: https://github.com/raeperd/realworld.go/pull/34
- **Created**: 2025-10-25
- **Completed**: -

## Context

Implementing the `GET /api/articles/feed` endpoint to retrieve articles from users the current user follows.

**From API Specification** (`@docs/spec.md`):
- Endpoint: `GET /api/articles/feed`
- Returns articles created by followed users, ordered by most recent first
- Query Parameters:
  - `limit` - Limit number of articles (default: 20)
  - `offset` - Skip number of articles (default: 0)
- Authentication: **Required** (unlike GET /api/articles which is optional)
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
GET /api/articles/feed?limit=10&offset=0
Authorization: Token jwt.token.here (REQUIRED)
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
      "following": true
    }
  }],
  "articlesCount": 1
}
```

**Note**: Response does NOT include `body` field for performance reasons.

### Authentication
- **Required** authentication via `authenticate` middleware
- Returns 401 if no valid token provided
- Current user ID extracted from JWT token
- `favorited` field reflects if current user favorited the article
- `following` field will always be `true` (since feed only shows followed users' articles)

### Query Parameters & Validation
- `limit`: Integer, default 20, must be positive
- `offset`: Integer, default 0, must be non-negative

### Database Schema Requirements
- ✅ `articles` table exists with slug, title, description, author_id, created_at, updated_at
- ✅ `users` table for author details
- ✅ `follows` table to filter articles by followed users
- ✅ `tags` and `article_tags` tables for tag lists
- ✅ `favorites` table for favorited status and favoritesCount

### Business Logic
- Query articles WHERE author_id IN (SELECT followed_id FROM follows WHERE follower_id = current_user_id)
- Order by created_at DESC (most recent first)
- Include author profile with following=true
- Include favorited status for current user
- Include favorites count
- Include tag list

### Error Handling
- 401 Unauthorized if token missing or invalid
- Empty feed returns `{"articles": [], "articlesCount": 0}`

## Implementation Steps

### Phase 0: Create Plan and Draft PR (MANDATORY FIRST STEP)

⚠️ **DO NOT PROCEED TO PHASE 1 WITHOUT COMPLETING THIS PHASE!**

- [x] Create feature branch: `feat/api-articles-feed`
- [ ] Create plan document: `docs/prompts/2025-10-25-get-api-articles-feed.md`
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

- [ ] Create test in `article_test.go` for feed endpoint
- [ ] Test setup:
  - Create two users (user1, user2)
  - User1 follows user2
  - User2 creates article
  - User1 requests feed
- [ ] Test verifies:
  - Returns user2's article in feed
  - Response structure matches spec (no body field)
  - Articles ordered by most recent first
  - articlesCount is correct
- [ ] Run test to confirm it fails: `go test -v -run TestHandleGetArticlesFeed`
- [ ] Commit: "test: add failing test for GET /api/articles/feed"
- [ ] Push to trigger CI (should show RED/failing status)

### Phase 2: Database Layer

- [ ] Add SQL query `ListArticlesFeed` to `internal/sqlite/query.sql`
  - Select articles where author_id in followed users
  - JOIN follows table: WHERE author_id IN (SELECT followed_id FROM follows WHERE follower_id = ?)
  - JOIN users for author details
  - LEFT JOIN for tags aggregation
  - LEFT JOIN for favorites count
  - LEFT JOIN favorites for current user's favorited status
  - ORDER BY created_at DESC
  - Support LIMIT and OFFSET
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles
- [ ] Commit: "feat: add database query for articles feed"
- [ ] Push to trigger CI

### Phase 3: Minimal Implementation (GREEN)

- [ ] Create `handleGetArticlesFeed` function in `article.go`
- [ ] Extract current user ID from context
- [ ] Parse query parameters (limit, offset) with defaults
- [ ] Call database query with current user ID, limit, offset
- [ ] Build response with articles array (excluding body field)
- [ ] Include articlesCount in response
- [ ] Register route in `route()`: `mux.Handle("GET /api/articles/feed", authenticate(handleGetArticlesFeed(db), jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestHandleGetArticlesFeed`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement GET /api/articles/feed"
- [ ] Push to trigger CI (should show GREEN/passing status)

### Phase 4: Edge Cases & Validation (RED → GREEN)

- [ ] Add test for empty feed (user not following anyone)
- [ ] Add test for limit parameter
- [ ] Add test for offset parameter
- [ ] Add test for multiple articles from different followed users
- [ ] Add test for favorited status in feed
- [ ] Add test for unauthorized access (401)
- [ ] Implement edge case handling to make tests pass
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add edge case tests for GET /api/articles/feed"
- [ ] Push to trigger CI

### Phase 5: Refactor (if needed)

- [ ] Review code for duplication with GET /api/articles
- [ ] Consider extracting common article response building logic
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit: "refactor: extract common article response building" (if changes made)
- [ ] Push to trigger CI

### Phase 6: Verification & Finalize PR

- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl:
  - Test feed with authentication
  - Test limit and offset
  - Test empty feed
  - Test unauthorized access (401)
- [ ] Update this plan's status to "Completed"
- [ ] Update PR description with implementation summary
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] **DO NOT merge** - Wait for review

## Verification Commands

```bash
# Run specific test
go test -v -run TestHandleGetArticlesFeed

# Run all tests
make test

# Lint check
make lint

# Manual testing examples

# Get feed (requires authentication)
curl http://localhost:8080/api/articles/feed \
  -H "Authorization: Token <jwt-token>"

# Pagination
curl "http://localhost:8080/api/articles/feed?limit=10&offset=5" \
  -H "Authorization: Token <jwt-token>"

# Unauthorized access (should return 401)
curl http://localhost:8080/api/articles/feed
```

## Implementation Notes

- **Important**: Authentication is REQUIRED (not optional like GET /api/articles)
- Response must NOT include `body` field (spec change from 2024/08/16)
- Default limit is 20, default offset is 0
- Articles ordered by most recent first (created_at DESC)
- `following` field will always be `true` for articles in feed
- Empty feed should return valid response with empty array
- Query filters by followed users using JOIN on follows table
