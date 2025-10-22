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

### 2. Create Implementation Plan and Draft PR

Create a detailed implementation plan document at:
`docs/prompts/YYYY-MM-DD-{method}-{path-simplified}.md`

The plan MUST include:

**Status & Links:**
```
Status: [ ] Not Started / [ ] In Progress / [x] Completed
PR: (to be added after PR creation)
```

**After creating the plan:**
1. Commit the plan document: "docs: add implementation plan for {METHOD} {PATH}"
2. Push to create branch and trigger first commit
3. Create draft PR with `gh pr create --draft`
4. Update plan with PR link
5. Commit and push: "docs: add PR link to plan"

This makes the plan document the natural first commit that creates the PR.

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
### Phase 0: Create Plan and Draft PR
- Create feature branch: `git checkout -b {branch-name}`
- Create implementation plan document at `docs/prompts/YYYY-MM-DD-{feature}.md`
- Commit plan: "docs: add implementation plan for {METHOD} {PATH}"
- Push branch: `git push -u origin {branch-name}`
- Create DRAFT PR with `gh pr create --draft` including:
  - Title: "feat: implement {METHOD} {PATH}"
  - Body referencing the plan document
  - Mark as draft to indicate work in progress
- This first commit with plan naturally creates the PR
- Update plan document with PR link
- Commit and push: "docs: add PR link to plan"

### Phase 1: Test First (RED)
- Create test file or add to existing test file
- Write failing integration test for happy path
- Run test to confirm it fails: `go test -v -run TestName`
- Commit: "test: add failing test for {endpoint}"
- Push to trigger CI (should show RED/failing status)

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
- Push to trigger CI (should show GREEN/passing status)

### Phase 4: Edge Cases & Validation (RED → GREEN)
- Add test for validation error (missing fields)
- Implement validation to make test pass
- Add test for unauthorized access (if auth required)
- Implement auth check to make test pass
- Add test for not found scenario (if applicable)
- Implement not found handling to make test pass
- Run all tests: `make test`
- Commit: "test: add edge case tests for {endpoint}"
- Push to trigger CI

### Phase 5: Refactor (if needed)
- Review code for duplication
- Extract common patterns if found in 3+ places
- Ensure all tests still pass after each refactoring
- Commit: "refactor: {description}" (if changes made)
- Push to trigger CI

### Phase 6: Verification & Finalize PR
- Run full test suite: `make test`
- Run linter: `make lint`
- Manual test with curl or HTTP client
- Update this plan's status to "Completed"
- Update PR description with final details
- Mark PR as ready for review: `gh pr ready`
- **DO NOT merge the PR** - Let the user review and merge when ready
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
2. Analyze dependencies and select next logical endpoint
3. Present selection with clear rationale explaining:
   - Why this endpoint is chosen
   - What dependencies are satisfied
   - Complexity level and implementation effort
4. Ask: "Ready to implement [SELECTED ENDPOINT]?"
5. On confirmation, proceed with TDD workflow:
   - Create branch and plan document → commit → push
   - Create draft PR (plan commit is first in PR)
   - Update plan with PR link → commit → push
   - Write failing test → commit → push → CI shows RED
   - Implement code → commit → push → CI shows GREEN
   - Add edge cases → commit → push → CI validates
   - Refactor if needed → commit → push → CI validates
   - Mark PR ready for review with `gh pr ready`

### Mode 2: Manual Selection (With Arguments)

```
User: /api POST /api/articles
```

Expected flow:
1. Read spec for POST /api/articles
2. Create plan at docs/prompts/2025-10-23-post-api-articles.md
3. Ask user for confirmation to proceed
4. Create branch and commit plan → push (first commit)
5. Create draft PR (plan document naturally creates PR)
6. Update plan with PR link → commit → push
7. Follow TDD cycle with CI visibility:
   - Write failing test → commit → push → CI RED
   - Implement code → commit → push → CI GREEN
   - Add edge cases → commit → push → CI validates
   - Refactor if needed → commit → push → CI validates
8. Mark PR ready for review with `gh pr ready` (not merged automatically)

## Endpoint Selection Strategy (Auto-Selection Mode)

When no arguments are provided, analyze and select the next endpoint intelligently:

**Analysis Steps**:
1. Read `@docs/spec.md` to list all RealWorld API endpoints
2. Check `@main.go` to identify already implemented endpoints
3. Check `@internal/sqlite/schema.sql` to see available database tables
4. Analyze dependencies between endpoints

**Selection Criteria** (in order of importance):
1. **Dependencies satisfied** - Prerequisite endpoints/tables must exist
2. **Complexity** - Prefer simpler endpoints when dependencies are equal
3. **Feature completeness** - Complete related endpoints together when possible
4. **Database readiness** - Prefer endpoints using existing schema

**Present Selection**:
- Show selected endpoint with clear rationale
- Explain why this endpoint is the logical next step
- Mention any dependencies that are satisfied
- Ask user for confirmation before proceeding

## Benefits of Plan-First Draft PR Approach

- **Natural First Commit**: Plan document is the logical first commit that creates the PR
- **Clear Documentation**: PR starts with implementation plan, providing context
- **CI Visibility**: Each TDD phase (RED → GREEN) is visible in PR commit history with CI status
- **Early Feedback**: Catch build/lint issues immediately on each commit
- **Progress Tracking**: PR shows implementation progress in real-time from plan to completion
- **Draft Status**: Clearly indicates work in progress, preventing premature review
- **Atomic Commits**: Each commit has clear purpose with CI validation
- **Complete History**: PR contains full story: plan → RED → GREEN → refactor → done

## Notes

- Always reference `@docs/prompts/TDD.md` instead of repeating TDD methodology
- Keep plans concise and actionable
- Focus on one test at a time
- Never skip the failing test verification step
- Prefer integration tests over unit tests
- Use real database, not mocks
- Push after each significant commit to trigger CI
