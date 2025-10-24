# POST /api/articles Implementation Plan

## Status & Links

Status: [ ] Not Started [x] In Progress [ ] Completed
PR: #26

## Context

Implementing the `POST /api/articles` endpoint as specified in `@docs/spec.md`.

This endpoint allows authenticated users to create new articles with title, description, body, and optional tags.

**API Specification (from spec.md):**

```
POST /api/articles

Authentication required, will return an Article

Required fields: title, description, body
Optional fields: tagList as an array of Strings
```

**Request Format:**
```json
{
  "article": {
    "title": "How to train your dragon",
    "description": "Ever wonder how?",
    "body": "You have to believe",
    "tagList": ["reactjs", "angularjs", "dragons"]
  }
}
```

**Response Format:**
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

Following Test-Driven Development as outlined in `@docs/prompts/TDD.md`:
- RED → GREEN → REFACTOR cycle
- Write failing test first
- Implement minimal code to pass
- Commit at each phase
- Push every commit immediately for CI validation

## Feature Requirements

### Endpoint Details
- **Method**: POST
- **Path**: `/api/articles`
- **Authentication**: Required (JWT token)
- **Request**: JSON object with "article" wrapper
- **Response**: JSON object with "article" wrapper containing created article

### Database Requirements

Need to create the following database structure:

1. **articles table**: Store article data (id, slug, title, description, body, author_id, timestamps)
2. **article_tags junction table**: Many-to-many relationship between articles and tags
3. **favorites table**: Track which users favorited which articles (for future endpoints)

**Key considerations:**
- Slug generation: Convert title to URL-friendly slug (lowercase, hyphens, unique)
- Tag handling: Create tags if they don't exist, associate with article
- Author: Use authenticated user's ID

### Validation & Error Handling
- Require authentication (401 if not authenticated)
- Validate required fields: title, description, body (422 if missing)
- Handle duplicate slug conflicts
- Validate tagList is array of strings if provided

### Business Logic
- Generate unique slug from title
- Create tags that don't exist yet
- Associate tags with article via junction table
- Return article with author profile embedded
- Return favorited=false and favoritesCount=0 (no favorites yet)

## Implementation Steps

### Phase 0: Database Schema & Queries
- [ ] Add articles table to `internal/sqlite/schema.sql`
- [ ] Add article_tags junction table
- [ ] Add favorites table (for future use)
- [ ] Add CreateArticle query to `internal/sqlite/query.sql`
- [ ] Add CreateTag query (with INSERT OR IGNORE for existing tags)
- [ ] Add AssociateArticleTag query
- [ ] Add GetArticleBySlug query (with author info)
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles: `go build`
- [ ] Commit: "feat: add articles database schema and queries"
- [ ] Push immediately: `git push`

### Phase 1: Test First (RED)
- [ ] Create `article_test.go` with package `main_test`
- [ ] Write failing integration test `TestPostArticles_Success` for happy path
- [ ] Include test for article with tags
- [ ] Run test to confirm it fails: `go test -v -run TestPostArticles`
- [ ] Commit: "test: add failing test for POST /api/articles"
- [ ] Push immediately: `git push`

### Phase 2: Minimal Implementation (GREEN)
- [ ] Create `article.go` with `handlePostArticles` function
- [ ] Implement article request/response types
- [ ] Implement slug generation function
- [ ] Implement validation for required fields
- [ ] Create article in database
- [ ] Handle tags creation and association
- [ ] Register route in `route()` function with authentication middleware
- [ ] Add article response types to `responseBody` interface union
- [ ] Run test to confirm it passes: `go test -v -run TestPostArticles`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement POST /api/articles endpoint"
- [ ] Push immediately: `git push`

### Phase 3: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for missing required fields (title, description, body)
- [ ] Verify 422 response with error messages
- [ ] Add test for unauthenticated request
- [ ] Verify 401 response
- [ ] Add test for duplicate slug handling (if needed)
- [ ] Add test for empty tagList
- [ ] Add test for article without tags
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add edge case tests for POST /api/articles"
- [ ] Push immediately: `git push`

### Phase 4: Refactor (if needed)
- [ ] Review code for duplication
- [ ] Extract slug generation if reusable
- [ ] Extract tag handling if complex
- [ ] Ensure all tests still pass after refactoring
- [ ] Commit: "refactor: {description}" (if changes made)
- [ ] Push immediately: `git push`

### Phase 5: Verification
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl
- [ ] Update this plan's status to "Completed"

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestPostArticles

# Run all tests
make test

# Lint check
make lint

# Manual test (need to create user and get token first)
# 1. Register user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"testuser","email":"test@example.com","password":"testpass"}}'

# 2. Extract token from response, then create article
curl -X POST http://localhost:8080/api/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Token {YOUR_TOKEN}" \
  -d '{
    "article": {
      "title": "Test Article",
      "description": "Test description",
      "body": "Test body content",
      "tagList": ["test", "example"]
    }
  }'
```

## Technical Notes

### Slug Generation
- Convert title to lowercase
- Replace spaces with hyphens
- Remove special characters
- Ensure uniqueness (append number if needed)
- Example: "How to Train Your Dragon" → "how-to-train-your-dragon"

### Tag Handling
- Use INSERT OR IGNORE to create tags only if they don't exist
- Get tag IDs after insert
- Associate article with tags via article_tags junction table

### Response Structure
- Article includes embedded author profile
- Author profile includes following status (false for now, since viewer is the author)
- favorited is always false for newly created articles
- favoritesCount is always 0 for newly created articles
- Dates in ISO 8601 format

### Database Schema Design

**articles table:**
- id: INTEGER PRIMARY KEY
- slug: TEXT NOT NULL UNIQUE
- title: TEXT NOT NULL
- description: TEXT NOT NULL
- body: TEXT NOT NULL
- author_id: INTEGER NOT NULL (FK to users)
- created_at: DATETIME
- updated_at: DATETIME

**article_tags table:**
- article_id: INTEGER (FK to articles)
- tag_id: INTEGER (FK to tags)
- PRIMARY KEY (article_id, tag_id)

**favorites table (for future):**
- user_id: INTEGER (FK to users)
- article_id: INTEGER (FK to articles)
- PRIMARY KEY (user_id, article_id)
- created_at: DATETIME
