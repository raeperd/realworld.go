---
description: Implement RealWorld API endpoint with TDD approach
argument-hint: METHOD PATH (optional - auto-selects next endpoint if empty)
---

# API Implementation Command

⚠️ **CRITICAL RULES**:

**Rule 0: NEVER work on main branch**
- ALWAYS create feature branch FIRST: `git checkout -b feat/api-endpoint-name`
- Do ALL work in feature branch
- Create PR from feature branch to main

**Rule 1-3: Plan → Draft PR → Implement workflow**
1. **Create plan document first** (never write code before plan exists)
2. **Create draft PR immediately after plan** (enables CI visibility for TDD phases)
3. **Then begin TDD implementation** (RED → GREEN visible in PR with CI status)

**See `@CLAUDE.md` section "Implementation Plans & History" for complete workflow requirements.**

## Auto-Selection Mode

**Arguments provided**: $ARGUMENTS

**Instructions**:

If no arguments are provided (empty $ARGUMENTS):

**YOU MUST autonomously find and select the next endpoint - DO NOT ask the user which endpoint to implement!**

⚠️ **CRITICAL**: When `/api` is called without arguments:
- **DO**: Read the codebase, analyze what's implemented, and make your own decision
- **DO**: Present your selected endpoint with rationale
- **DO**: Ask only "Ready to implement [YOUR CHOICE]?"
- **DO NOT**: Ask "which endpoint should I implement?"
- **DO NOT**: Present multiple options for the user to choose from
- **DO NOT**: Wait for the user to tell you what to implement

1. **Read and analyze** (DO THIS FIRST - don't skip!):
   - Read `docs/spec.md` to list ALL available endpoints
   - Read `main.go` to identify already implemented endpoints
   - Read `internal/sqlite/schema.sql` to see available database tables
   - Check `docs/prompts/` for recently completed implementations

2. **Determine next endpoint** by comparing spec vs implementation:
   - List ALL unimplemented endpoints from spec
   - Analyze dependencies (e.g., articles need users, comments need articles)
   - Consider complexity (prefer simpler when dependencies are equal)
   - Consider feature grouping (complete related endpoints together)
   - Consider database readiness (prefer endpoints using existing schema)

3. **Select and present** your autonomous decision:
   - State the selected endpoint clearly: "Next endpoint to implement: **GET /api/tags**"
   - Explain rationale in 2-3 sentences covering:
     - Why this endpoint is the logical next step
     - What dependencies are satisfied
     - Expected complexity level
   - Show what prerequisites exist (database tables, auth system, etc.)

4. **Ask ONLY for confirmation to proceed**: "Ready to implement GET /api/tags?"
   - Do NOT ask "which endpoint should I implement?"
   - Do NOT present multiple options for the user to choose
   - You have already decided - only ask to confirm starting work

5. **On confirmation, proceed immediately** with full workflow:
   - Create branch and plan document → commit → push
   - Create draft PR with plan as first commit
   - Update plan with PR link → commit → push
   - Begin TDD: RED → GREEN → REFACTOR (each with commits and CI)

If arguments are provided ($1 and $2):
- Implement the specified endpoint **$1 $2** directly
- Still follow the complete plan → draft PR → implement workflow

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

**Use the plan creation workflow from `@.claude/commands/plan.md`:**

Create implementation plan for the endpoint with:
- **Title**: "{METHOD} {PATH}"
- **Description**: Brief summary from API spec

The plan document should be created at:
`docs/prompts/YYYY-MM-DD-{method}-{path-simplified}.md`

**Plan Structure** (as defined in plan.md):
- Status & Links (with PR reference)
- Context (original request, API spec reference)
- Methodology (reference to TDD.md)
- Feature Requirements (request/response, auth, validation, database)
- Implementation Steps (detailed TDD checklist)
- Verification Commands

**Follow plan.md workflow**:
1. Create branch and plan document
2. Commit: "docs: add implementation plan for {METHOD} {PATH}"
3. Push to create branch
4. Create draft PR with `gh pr create --draft`
5. Update plan with PR link
6. Commit and push: "docs: add PR link to plan"

This creates the plan document as the natural first commit in the PR.

**Implementation Steps:**
Detailed checklist following this pattern (TEST FIRST):

```markdown
### Phase 0: Create Plan and Draft PR (MANDATORY FIRST STEP)

⚠️ **DO NOT PROCEED TO PHASE 1 WITHOUT COMPLETING THIS PHASE!**

**Follow workflow from `@.claude/commands/plan.md`:**

1. **Create feature branch**:
   ```bash
   git checkout -b feat/api-{endpoint-name}
   ```

2. **Create plan document**: `docs/prompts/YYYY-MM-DD-{method}-{path}.md`
   - Include all sections: Status, Context, Methodology, Requirements, Steps, Verification
   - Mark status as "[ ] Not Started"
   - Leave PR field as "(to be created)"

3. **Commit plan as first commit**:
   ```bash
   git add docs/prompts/YYYY-MM-DD-{method}-{path}.md
   git commit -m "docs: add implementation plan for {METHOD} {PATH}"
   git push -u origin feat/api-{endpoint-name}
   ```

4. **Create DRAFT PR immediately**:
   ```bash
   gh pr create --draft \
     --title "feat: implement {METHOD} {PATH}" \
     --body "Implementation plan: docs/prompts/YYYY-MM-DD-{method}-{path}.md"
   ```

5. **Update plan with PR link**:
   - Change status to "[ ] In Progress"
   - Add PR link: "PR: https://github.com/{owner}/{repo}/pull/{number}"
   - Commit and push:
   ```bash
   git add docs/prompts/YYYY-MM-DD-{method}-{path}.md
   git commit -m "docs: add PR link to plan"
   git push
   ```

**Why this phase is mandatory**:
- ✅ Plan document is the first commit that creates the PR
- ✅ Each subsequent commit (RED/GREEN) triggers CI visible in PR
- ✅ Draft status prevents premature review
- ✅ Complete implementation history visible from plan to completion
- ✅ CI failures caught immediately on each push

**Verification before proceeding**:
- ✅ Branch exists and pushed to remote
- ✅ Plan document committed (first commit in branch)
- ✅ Draft PR created (visible on GitHub)
- ✅ Plan updated with PR link and pushed
- ✅ PR shows plan document as first commit

**ONLY AFTER THIS PHASE IS COMPLETE**, proceed to Phase 1 (Test First).

### Phase 1: Test First (RED)
- Create test file or add to existing test file
- Write failing integration test for happy path
- Run test to confirm it fails: `go test -v -run TestName`
- Commit: "test: add failing test for {endpoint}"
- Push to trigger CI (should show RED/failing status)

### Phase 2: Database Layer (if needed)
- Add SQL queries to `internal/sqlite/query.sql`
- Run `make generate` to generate Go code
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
- Add queries to `internal/sqlite/query.sql`
- Use meaningful query names (e.g., `-- name: GetArticleBySlug :one`)
- Run `make generate` after adding queries
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
1. **Read and analyze** (autonomously, without asking):
   - Read `docs/spec.md` to see all RealWorld API endpoints
   - Read `main.go` to identify what's already implemented
   - Read `internal/sqlite/schema.sql` to check database readiness
   - Check `docs/prompts/` for recent implementations
2. **Make autonomous decision** (no user input needed):
   - Compare spec vs implementation to find unimplemented endpoints
   - Analyze dependencies (articles need users, comments need articles)
   - Consider complexity and feature grouping
   - **SELECT the next logical endpoint yourself**
3. **Present your decision** with clear rationale:
   - "Next endpoint to implement: **GET /api/tags**"
   - Explain why (2-3 sentences): dependencies satisfied, complexity level, logical progression
   - Show prerequisites that exist (database, auth, related endpoints)
4. **Ask ONLY for confirmation**: "Ready to implement GET /api/tags?"
   - You've already decided - only confirming to start work
   - Do NOT ask "which endpoint?" or present multiple options
5. **On confirmation, proceed** with full TDD workflow:
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

When no arguments are provided, YOU must analyze and autonomously select the next endpoint:

**Analysis Steps** (do all of these yourself):
1. Read `docs/spec.md` to list all RealWorld API endpoints
2. Read `main.go` to identify already implemented endpoints
3. Read `internal/sqlite/schema.sql` to see available database tables
4. Check `docs/prompts/` for recently completed implementations
5. Analyze dependencies between endpoints
6. **Make the decision yourself** - do not ask the user which to implement

**Selection Criteria** (in order of importance):
1. **Dependencies satisfied** - Prerequisite endpoints/tables must exist
2. **Complexity** - Prefer simpler endpoints when dependencies are equal
3. **Feature completeness** - Complete related endpoints together when possible
4. **Database readiness** - Prefer endpoints using existing schema
5. **Logical progression** - Follow a sensible implementation order

**Present Your Decision** (not a question):
- State your selected endpoint clearly and confidently
- Explain the rationale (2-3 sentences) covering:
  - Why this is the logical next step
  - What dependencies are satisfied
  - Expected complexity level
- Show prerequisites (database tables, auth, related endpoints)
- Ask ONLY: "Ready to implement [YOUR SELECTED ENDPOINT]?"
  - This is asking permission to START work on YOUR decision
  - This is NOT asking the user to choose an endpoint

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
