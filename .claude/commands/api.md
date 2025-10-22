---
description: Implement RealWorld API endpoint with TDD approach
argument-hint: METHOD PATH (optional - auto-selects next endpoint if empty)
---

# API Implementation Command

## Auto-Selection Mode

**Arguments provided**: $ARGUMENTS

**Instructions**:

If no arguments are provided (empty $ARGUMENTS):
1. **Read the API specification** from `@docs/spec.md` to see all endpoints
2. **Check main.go** to see which endpoints are already implemented
3. **Analyze and select** the next logical endpoint to implement based on:
   - Dependencies (implement prerequisite endpoints first)
   - Complexity (start with simpler endpoints)
   - Feature grouping (complete related endpoints together)
4. **Present the selection** to user with rationale
5. **Ask for confirmation** before proceeding
6. **Proceed with implementation** using the workflow below

If arguments are provided ($1 and $2):
- Implement the specified endpoint **$1 $2** directly

---

## Implementation Workflow

You are implementing an endpoint using Test-Driven Development (TDD) following Kent Beck's principles.

### 1. Preparation Phase

Before starting implementation:

1. **Read the API specification** from `@docs/spec.md` for the target endpoint
2. **Check existing implementation plans** in `docs/prompts/` for similar endpoints
3. **Identify the database schema** requirements from `@internal/sqlite/schema.sql`
4. **Review existing code patterns** from similar handlers in the codebase

### 2. Create Implementation Plan

Create a detailed implementation plan document at:
`docs/prompts/YYYY-MM-DD-{method}-{path-simplified}.md`

The plan MUST include:

**Status & Links:**
```
Status: [ ] Not Started / [ ] In Progress / [x] Completed
PR: #[number] (when created)
```

**Context:**
- Original request and endpoint details
- Reference to API spec section

**Methodology:**
- Reference to `@docs/prompts/TDD.md` (don't repeat TDD content)
- Any endpoint-specific testing considerations

**Feature Requirements:**
- Request/response format from spec
- Authentication requirements
- Validation rules
- Database operations needed

**Implementation Steps:**
Detailed checklist following this pattern (TEST FIRST):

```markdown
### Phase 1: Test First (RED)
- Create test file or add to existing test file
- Write failing integration test for happy path
- Run test to confirm it fails: `go test -v -run TestName`
- Commit: "test: add failing test for {endpoint}"

### Phase 2: Database Layer (if needed)
- Add SQL queries to `internal/sqlite/query.sql`
- Run `make sqlc` to generate Go code
- Verify generated code compiles

### Phase 3: Minimal Implementation (GREEN)
- Create handler function with minimal implementation
- Add request/response types
- Add validation logic
- Register route in `route()` function
- Run test to confirm it passes: `go test -v -run TestName`
- Run all tests: `make test`
- Commit: "feat: implement {endpoint}"

### Phase 4: Edge Cases & Validation (RED → GREEN)
- Add test for validation error (missing fields)
- Implement validation to make test pass
- Add test for unauthorized access (if auth required)
- Implement auth check to make test pass
- Add test for not found scenario (if applicable)
- Implement not found handling to make test pass
- Run all tests: `make test`
- Commit: "test: add edge case tests for {endpoint}"

### Phase 5: Refactor (if needed)
- Review code for duplication
- Extract common patterns if found in 3+ places
- Ensure all tests still pass after each refactoring
- Commit: "refactor: {description}" (if changes made)

### Phase 6: Verification
- Run full test suite: `make test`
- Run linter: `make lint`
- Manual test with curl or HTTP client
- Update this plan's status to "Completed"

### Phase 7: Create Pull Request
- Push branch to remote: `git push -u origin <branch-name>`
- Create PR with `gh pr create` including:
  - Concise title: "feat: implement {METHOD} {PATH}"
  - Summary section with endpoint description
  - Changes section listing endpoint, database, handler, tests
  - Test plan section showing coverage and test results
  - Implementation details mentioning TDD workflow
- Update plan document with PR link
- Commit and push plan update
```

**Verification Commands:**
```bash
# Test the specific endpoint
go test -v -run Test{HandlerName}

# Run all tests
make test

# Lint check
make lint

# Manual test example
curl -X {METHOD} http://localhost:8080{PATH} \
  -H "Content-Type: application/json" \
  -H "Authorization: Token {token}" \
  -d '{request_body}'
```

### 3. Implementation Process

Follow TDD strictly:

**RED Phase:**
- Write the SIMPLEST test that exercises the endpoint
- Test MUST fail (verify by running it)
- Commit failing test before implementation

**GREEN Phase:**
- Write MINIMAL code to make test pass
- No gold-plating or premature abstraction
- Verify test passes
- Commit working implementation

**REFACTOR Phase:**
- Only refactor when tests are GREEN
- One refactoring at a time
- Run tests after each refactoring step
- Commit structural changes separately

### 4. Code Organization Guidelines

**File Structure:**
- For user-related endpoints: add to `user.go`
- For article-related endpoints: create `article.go`
- For profile-related endpoints: create `profile.go`
- For comment-related endpoints: create `comment.go`
- For tag-related endpoints: create `tag.go`

**Test File Naming:**
- Use `{feature}_test.go` (e.g., `article_test.go`)
- Use different package name: `package main_test`
- Always use `t.Parallel()` unless tests conflict

**Handler Function Naming:**
- Pattern: `handle{Method}{Resource}` (e.g., `handleGetArticles`, `handlePostArticles`)
- For path parameters: `handle{Method}{Resource}Slug` (e.g., `handleDeleteArticlesSlug`)

**Response Encoding:**
- Use existing `encodeResponse()` and `encodeErrorResponse()` functions
- Follow the exact JSON structure from API spec
- Ensure "user" wrapper or "article" wrapper as specified

### 5. Database Guidelines

**Query Organization:**
- Add queries to `internal/sqlite/queries.sql`
- Use meaningful query names (e.g., `-- name: GetArticleBySlug :one`)
- Run `make sqlc` after adding queries
- Handle `sql.ErrNoRows` appropriately

**Transaction Management:**
- Use transactions for write operations
- Always defer rollback: `defer func() { _ = tx.Rollback() }()`
- Commit only after all operations succeed
- Use `sqlite.New(tx)` to create queries with transaction

### 6. Testing Patterns

**Integration Tests:**
- Use real HTTP requests to actual server
- Setup test data in database before test
- Clean up or use isolated test database
- Test full request/response cycle

**Test Structure:**
```go
func TestHandleMethodResource(t *testing.T) {
    t.Parallel()

    // Arrange: setup test data

    // Act: make HTTP request

    // Assert: verify response
}
```

**Common Assertions:**
- Status code matches expected
- Response body matches spec format
- Database state changes as expected
- Error messages are informative

### 7. Commit Strategy

Make atomic commits following TDD phases:

1. **Test commit:** `test: add failing test for {endpoint}`
2. **Implementation commit:** `feat: implement {endpoint}`
3. **Edge cases commit:** `test: add edge case tests for {endpoint}`
4. **Refactor commit (if needed):** `refactor: {what was refactored}`

Each commit should:
- Have all tests passing (except RED phase test commits)
- Be independently reviewable
- Have clear, descriptive message

### 8. Common Patterns to Follow

**Authentication:**
```go
// Use authenticate middleware for protected endpoints
mux.Handle("GET /api/resource", authenticate(handleGetResource(db, jwtSecret), jwtSecret))
```

**Request Validation:**
```go
func (r requestBody) Validate() []error {
    var errs []error
    if r.Field == "" {
        errs = append(errs, errors.New("field is required"))
    }
    return errs
}
```

**Error Responses:**
```go
// Use existing helper
encodeErrorResponse(r.Context(), http.StatusUnprocessableEntity, errs, w)
```

**Success Responses:**
```go
// Use existing helper with proper wrapper
encodeResponse(r.Context(), http.StatusOK, responseBody{...}, w)
```

### 9. Execute Implementation

After creating the plan:

1. Ask user "Ready to implement $1 $2?"
2. On confirmation, follow the plan step-by-step
3. Update plan checkboxes as you complete each step
4. Make commits at appropriate points (test, implementation, refactor)
5. Report progress and any issues encountered

### 10. Final Verification and Pull Request

Before considering the task complete:

- All tests pass (`make test`)
- No linter warnings (`make lint`)
- Implementation matches API spec exactly
- Manual testing confirms correct behavior
- Plan document updated to "Completed" status
- All commits follow the commit strategy
- Pull request created with concise title and description
- Plan document updated with PR link

## Example Usage

### Mode 1: Auto-Selection (No Arguments)

```
User: /api
```

Expected flow:
1. Read spec and check implemented endpoints
2. Analyze and select next logical endpoint (e.g., "GET /api/tags - simple, no dependencies")
3. Present selection with rationale to user
4. Ask: "Ready to implement GET /api/tags?"
5. On confirmation, proceed with TDD workflow
6. Create plan, implement, test, commit, create PR

### Mode 2: Manual Selection (With Arguments)

```
User: /api POST /api/articles
```

Expected flow:
1. Read spec for POST /api/articles
2. Create plan at docs/prompts/2025-10-23-post-api-articles.md
3. Ask user for confirmation to proceed
4. Follow TDD cycle: test (RED) → implement (GREEN) → refactor
5. Make atomic commits at each phase
6. Verify and mark plan as completed
7. Create pull request with concise title and description
8. Update plan with PR link

## Endpoint Selection Strategy (Auto-Selection Mode)

When no arguments are provided, select the next endpoint based on:

**Priority Order**:
1. **Tags endpoint** (`GET /api/tags`) - Simplest, no auth, no dependencies
2. **Profile endpoints** (`GET /api/profiles/:username`) - Simple read operations
3. **Article read operations** (`GET /api/articles`, `GET /api/articles/:slug`) - Complex but read-only
4. **Follow/Unfollow** (`POST/DELETE /api/profiles/:username/follow`) - Requires profiles
5. **Article mutations** (`POST/PUT/DELETE /api/articles`) - Requires articles read
6. **Comments** (`GET/POST/DELETE /api/articles/:slug/comments`) - Requires articles
7. **Favorites** (`POST/DELETE /api/articles/:slug/favorite`) - Requires articles
8. **Feed** (`GET /api/articles/feed`) - Requires articles and follows

**Selection Criteria**:
- Dependencies satisfied (prerequisite endpoints implemented)
- Complexity (simpler endpoints first)
- Feature completeness (complete one feature before starting another)
- Database schema ready (tables exist or can be added)

## Notes

- Always reference `@docs/prompts/TDD.md` instead of repeating TDD methodology
- Keep plans concise and actionable
- Focus on one test at a time
- Never skip the failing test verification step
- Prefer integration tests over unit tests
- Use real database, not mocks
