# GET /api/articles/:slug Implementation Plan

## Status & Links

Status: [ ] Not Started [ ] In Progress [x] Completed
PR: #27

## Context

Implementing the `GET /api/articles/:slug` endpoint as specified in `@docs/spec.md`.

This endpoint allows anyone (authenticated or not) to retrieve a single article by its slug.

**API Specification (from spec.md):**

```
GET /api/articles/:slug

No authentication required, will return single article
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
- **Method**: GET
- **Path**: `/api/articles/:slug`
- **Authentication**: Optional (affects `favorited` and `following` fields)
- **Response**: JSON object with "article" wrapper containing article details

### Database Requirements

Queries already exist in `internal/sqlite/query.sql`:
1. **GetArticleBySlug**: Fetch article with author info (lines 52-60)
2. **GetArticleTagsByArticleID**: Fetch tags for the article (lines 62-67)
3. **GetFavoritesCount**: Count favorites for the article (lines 69-70)
4. **IsFavorited**: Check if current user favorited (lines 72-73)
5. **IsFollowing**: Check if current user follows author (query.sql:28-29)

**Key considerations:**
- Return 404 if article with slug not found
- Return favorited=false and following=false if no authenticated user
- Return actual favorited/following status if authenticated user
- Dates in ISO 8601 format (RFC3339)

### Validation & Error Handling
- No authentication required (but optional)
- Return 404 if article not found
- Handle database errors gracefully

### Business Logic
- Fetch article by slug with author info
- Fetch article tags
- Fetch favorites count
- If authenticated: check if user favorited and follows author
- If not authenticated: favorited=false, following=false
- Return complete article with embedded author profile

## Implementation Steps

### Phase 1: Test First (RED)
- [x] Create branch: `git checkout -b feat/get-api-articles-slug`
- [x] Commit and push plan document
- [x] Create draft PR
- [x] Update plan with PR link
- [ ] Add failing test `TestGetArticlesSlug_Success` in `article_test.go`
- [ ] Test fetches article created in setup
- [ ] Verify all fields in response
- [ ] Run test to confirm it fails: `go test -v -run TestGetArticlesSlug`
- [ ] Commit: "test: add failing test for GET /api/articles/:slug"
- [ ] Push immediately: `git push`

### Phase 2: Minimal Implementation (GREEN)
- [ ] Add `handleGetArticlesSlug` function in `article.go`
- [ ] Parse slug from URL path parameter
- [ ] Fetch article using existing `GetArticleBySlug` query
- [ ] Fetch tags using `GetArticleTagsByArticleID`
- [ ] Fetch favorites count using `GetFavoritesCount`
- [ ] Check favorited status if authenticated
- [ ] Check following status if authenticated
- [ ] Return 404 if article not found
- [ ] Register route in `route()` with optional authentication
- [ ] Run test to confirm it passes: `go test -v -run TestGetArticlesSlug`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement GET /api/articles/:slug endpoint"
- [ ] Push immediately: `git push`

### Phase 3: Edge Cases & Validation (RED → GREEN)
- [ ] Add test for unauthenticated request
- [ ] Verify favorited=false and following=false for unauthenticated
- [ ] Add test for authenticated request
- [ ] Verify favorited and following reflect actual status
- [ ] Add test for article not found (404)
- [ ] Implement 404 handling
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add edge case tests for GET /api/articles/:slug"
- [ ] Push immediately: `git push`

### Phase 4: Refactor (if needed)
- [ ] Review code for duplication with POST /api/articles
- [ ] Consider extracting article response building logic if duplicated
- [ ] Ensure all tests still pass after refactoring
- [ ] Commit: "refactor: {description}" (if changes made)
- [ ] Push immediately: `git push`

### Phase 5: Verification
- [ ] Run full test suite: `make test`
- [ ] Run linter: `make lint`
- [ ] Manual test with curl (authenticated and unauthenticated)
- [ ] Update this plan's status to "Completed"
- [ ] Mark PR as ready for review

## Verification Commands

```bash
# Test the specific endpoint
go test -v -run TestGetArticlesSlug

# Run all tests
make test

# Lint check
make lint

# Manual test - unauthenticated
curl -X GET http://localhost:8080/api/articles/how-to-train-your-dragon

# Manual test - authenticated (need token from login)
curl -X GET http://localhost:8080/api/articles/how-to-train-your-dragon \
  -H "Authorization: Token {YOUR_TOKEN}"

# Manual test - article not found
curl -X GET http://localhost:8080/api/articles/nonexistent-slug
```

## Technical Notes

### URL Path Parameter Extraction
```go
slug := r.PathValue("slug")
```

### Authentication Detection
Use `authenticateOptional` middleware to:
- Continue without user ID if no token
- Set user ID in context if valid token provided
- Continue without user ID if invalid token

### Response Building
Reuse `articleResponse` and `authorProfile` types from `article.go`.

### 404 Handling
```go
if errors.Is(err, sql.ErrNoRows) {
    encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
    return
}
```

### Conditional Favorited/Following Checks
```go
userID, ok := r.Context().Value(userIDKey).(int64)
if ok {
    // Check actual favorited and following status
} else {
    // Default to false
}
```

### Query Reuse
All necessary queries already exist:
- `GetArticleBySlug` returns article with author info joined
- `GetArticleTagsByArticleID` returns []string of tag names
- `GetFavoritesCount` returns count
- `IsFavorited` returns boolean (if authenticated)
- `IsFollowing` returns boolean (if authenticated)
