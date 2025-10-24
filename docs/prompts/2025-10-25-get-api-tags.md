# GET /api/tags Implementation Plan

## Status & Links

Status: [ ] Not Started [ ] In Progress [x] Completed
PR: #24

## Context

Implementing the `GET /api/tags` endpoint as specified in `@docs/spec.md`.

This is a simple endpoint that returns a list of all unique tags used in articles. No authentication is required.

**API Specification (from spec.md):**

```
GET /api/tags

No authentication required, returns a List of Tags
```

**Expected Response Format:**

```json
{
  "tags": [
    "reactjs",
    "angularjs"
  ]
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
- **Path**: `/api/tags`
- **Authentication**: None required
- **Response**: JSON object with "tags" array containing string tags

### Database Requirements

Need to check if we have a tags table in the schema. Based on the Article structure showing `tagList`, tags are likely stored either:
1. As a separate `tags` table with article associations
2. As a JSON array in articles table

We'll need to add the appropriate SQL query to fetch distinct tags.

### Validation & Error Handling
- No request validation needed (no parameters)
- Should return empty array if no tags exist
- Standard error response for server errors

## Implementation Steps

### Phase 0: Schema Investigation & Database Setup
- [ ] Check `internal/sqlite/schema.sql` for tags table structure
- [ ] Add SQL query to `internal/sqlite/queries.sql` for fetching all distinct tags
- [ ] Run `make generate` to generate Go code
- [ ] Verify generated code compiles

### Phase 1: Test First (RED)
- [ ] Create `tag_test.go` with package `main_test`
- [ ] Write failing integration test `TestHandleGetTags` for happy path (empty tags)
- [ ] Run test to confirm it fails: `go test -v -run TestHandleGetTags`
- [ ] Commit: "test: add failing test for GET /api/tags"
- [ ] Push immediately: `git push`

### Phase 2: Minimal Implementation (GREEN)
- [ ] Create `tag.go` with `handleGetTags` function
- [ ] Implement response type matching spec format
- [ ] Register route in `route()` function: `mux.HandleFunc("GET /api/tags", handleGetTags(db))`
- [ ] Run test to confirm it passes: `go test -v -run TestHandleGetTags`
- [ ] Run all tests: `make test`
- [ ] Commit: "feat: implement GET /api/tags endpoint"
- [ ] Push immediately: `git push`

### Phase 3: Additional Test Cases (RED → GREEN)
- [ ] Add test for tags with actual data (if relevant)
- [ ] Implement any needed changes to make test pass
- [ ] Run all tests: `make test`
- [ ] Commit: "test: add test cases for GET /api/tags"
- [ ] Push immediately: `git push`

### Phase 4: Refactor (if needed)
- [ ] Review code for duplication
- [ ] Extract common patterns if found
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
go test -v -run TestHandleGetTags

# Run all tests
make test

# Lint check
make lint

# Manual test
curl http://localhost:8080/api/tags

# With server running (in another terminal)
make run
# Then:
curl -v http://localhost:8080/api/tags
```

## Notes

- This endpoint is particularly simple: no auth, no parameters, just returns data
- The main question is how tags are stored in the database
- Since articles haven't been implemented yet, we may need to add the tags table structure first
- Response should always return a valid JSON object with "tags" array (empty if no tags)
