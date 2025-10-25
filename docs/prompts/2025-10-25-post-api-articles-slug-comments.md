# Implementation Plan: POST /api/articles/:slug/comments

**Status**: [ ] Not Started
**PR**: (will be added after PR creation)

## Context

Implement the `POST /api/articles/:slug/comments` endpoint to allow authenticated users to add comments to articles. This is the first of three comment-related endpoints (POST, GET, DELETE).

**API Specification Reference**: `@docs/spec.md` lines 201-217

## Methodology

Following Test-Driven Development as described in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Write failing test first
- Minimal implementation to pass
- Refactor only when tests are green

**Testing approach**: Integration tests using real HTTP server and database.

## Feature Requirements

### Endpoint
- **Method**: POST
- **Path**: /api/articles/:slug/comments
- **Authentication**: Required

### Request Format
```json
{
  "comment": {
    "body": "His name was my name too."
  }
}
```

**Required fields**: body

### Response Format
```json
{
  "comment": {
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
  }
}
```

### Validation Rules
- Authentication required (401 if not authenticated)
- Article must exist (404 if article not found)
- Body field is required (422 if missing)
- Body must not be empty (422 if empty)

### Business Logic
- Comment is associated with the article via article_id
- Comment author is the authenticated user
- Created and updated timestamps are set automatically
- Following status is determined by current user's relationship to comment author

### Database Operations

**Schema addition needed** (`internal/sqlite/schema.sql`):
```sql
CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    body text NOT NULL,
    article_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TRIGGER update_comments_updated_at
    AFTER UPDATE OF body ON comments
    FOR EACH ROW
BEGIN
    UPDATE comments
    SET updated_at = DATETIME('now')
    WHERE rowid = NEW.rowid;
END;

CREATE INDEX idx_comments_article_id ON comments(article_id);
CREATE INDEX idx_comments_author_id ON comments(author_id);
```

**Query needed** (`internal/sqlite/query.sql`):
```sql
-- name: CreateComment :one
INSERT INTO comments (body, article_id, author_id)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetCommentWithAuthor :one
SELECT
    c.*,
    u.username as author_username,
    u.bio as author_bio,
    u.image as author_image
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.id = ?;
```

## Implementation Steps

### Phase 0: Feature Branch & Plan
- [x] Create feature branch: `git checkout -b feat/post-api-articles-slug-comments`
- [ ] Create this plan document
- [ ] Commit plan: `git add docs/prompts && git commit -m "docs: add plan for POST /api/articles/:slug/comments"`
- [ ] Push branch: `git push -u origin feat/post-api-articles-slug-comments`
- [ ] Create draft PR: `gh pr create --draft --title "feat: implement POST /api/articles/:slug/comments" --body "Implements comment creation endpoint. See docs/prompts/2025-10-25-post-api-articles-slug-comments.md"`
- [ ] Update plan with PR link and commit
- [ ] Push update: `git push`

### Phase 1: Database Schema
- [ ] Add comments table to `internal/sqlite/schema.sql`
- [ ] Add trigger for updated_at
- [ ] Add indexes for article_id and author_id
- [ ] Run `make generate` to recreate database
- [ ] Verify schema compiles: `go build`
- [ ] Commit: `git add . && git commit -m "feat: add comments table schema" && git push`

### Phase 2: Database Queries
- [ ] Add `CreateComment` query to `internal/sqlite/query.sql`
- [ ] Add `GetCommentWithAuthor` query to `internal/sqlite/query.sql`
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`
- [ ] Commit: `git add . && git commit -m "feat: add comment SQL queries" && git push`

### Phase 3: Test First (RED)
- [ ] Create `comment_test.go` in `comment_test` package
- [ ] Add test setup helper to create test article
- [ ] Write `TestPostArticlesSlugComments_Success` test
- [ ] Test should verify 200 status, comment body, author info, timestamps
- [ ] Run test to confirm it fails: `go test -v -run TestPostArticlesSlugComments_Success`
- [ ] Commit: `git add . && git commit -m "test: add failing test for POST /api/articles/:slug/comments" && git push`

### Phase 4: Minimal Implementation (GREEN)
- [ ] Create `comment.go` file with `handlePostArticlesSlugComments` function
- [ ] Define request/response types
- [ ] Implement basic validation (body required)
- [ ] Get article by slug (return 404 if not found)
- [ ] Create comment in database
- [ ] Fetch comment with author info
- [ ] Build and return response
- [ ] Register route: `mux.Handle("POST /api/articles/{slug}/comments", authenticate(handlePostArticlesSlugComments(db), jwtSecret))`
- [ ] Run test to confirm it passes: `go test -v -run TestPostArticlesSlugComments_Success`
- [ ] Run all tests: `make test`
- [ ] Commit: `git add . && git commit -m "feat: implement POST /api/articles/:slug/comments" && git push`

### Phase 5: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for missing body field (422)
- [ ] Implement validation to make test pass
- [ ] Commit: `git add . && git commit -m "test: add validation test for missing body" && git push`
- [ ] Add test for article not found (404)
- [ ] Verify error handling for not found
- [ ] Commit: `git add . && git commit -m "test: add not found test for POST comments" && git push`
- [ ] Add test for empty body (422)
- [ ] Implement empty body validation
- [ ] Commit: `git add . && git commit -m "test: add empty body validation test" && git push`
- [ ] Run all tests: `make test`

### Phase 6: Refactor (if needed)
- [ ] Review code for duplication with other handlers
- [ ] Extract common patterns if found in 3+ places
- [ ] Ensure all tests still pass after each refactoring
- [ ] Commit if changes made: `git add . && git commit -m "refactor: {description}" && git push`

### Phase 7: Verification & PR
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (see verification commands below)
- [ ] Update plan status to "Completed"
- [ ] Commit plan update: `git add docs/prompts && git commit -m "docs: mark POST /api/articles/:slug/comments as completed" && git push`
- [ ] Mark PR as ready for review: `gh pr ready`
- [ ] Request user approval before merging

## Verification Commands

```bash
# Run specific test
go test -v -run TestPostArticlesSlugComments

# Run all tests
make test

# Lint check
make lint

# Manual test - First get an auth token
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "email": "test@test.com",
      "password": "password"
    }
  }'

# Then create a comment
curl -X POST http://localhost:8080/api/articles/test-article/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "comment": {
      "body": "This is a test comment"
    }
  }'
# Should return 200 OK with comment object

# Test missing body
curl -X POST http://localhost:8080/api/articles/test-article/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "comment": {}
  }'
# Should return 422 with validation error

# Test article not found
curl -X POST http://localhost:8080/api/articles/nonexistent/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Token YOUR_JWT_TOKEN" \
  -d '{
    "comment": {
      "body": "Comment on nonexistent article"
    }
  }'
# Should return 404 Not Found
```

## Implementation Notes

- Use similar pattern to article handlers for request/response structure
- Follow existing timestamp formatting (ISO 8601 with RFC3339)
- Ensure "following" status is calculated based on current user context
- Use transaction for consistency when creating comment
- Return full comment object with author profile embedded
- Handle sql.ErrNoRows when article not found
- Validate body is not empty string, not just missing field
