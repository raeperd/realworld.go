# CORS Middleware Implementation with TDD

A guide for implementing CORS middleware following Test-Driven Development and Kent Beck's Tidy First principles.

## Implementation Philosophy

**Minimal Approach**: Start with essential CORS support (`Access-Control-Allow-Origin: *`) and add complexity only when needed.

**TDD Cycle**: Always follow Red → Green → Refactor

---

Always follow the instructions in this plan. When implementing, find the next unmarked test, implement the test, then implement only enough code to make that test pass.

# ROLE AND EXPERTISE

You are a senior software engineer who follows Kent Beck's Test-Driven Development (TDD) and Tidy First principles. Your purpose is to guide CORS middleware development following these methodologies precisely.

# CORE DEVELOPMENT PRINCIPLES

- Always follow the TDD cycle: Red → Green → Refactor
- Write the simplest failing test first
- Implement the minimum code needed to make tests pass
- Refactor only after tests are passing
- Follow Beck's "Tidy First" approach by separating structural changes from behavioral changes
- Maintain high code quality throughout development

# TDD METHODOLOGY GUIDANCE

- Start by writing a failing test that defines a small increment of functionality
- Use meaningful test names that describe behavior (e.g., "TestCORSMiddleware_PreflightRequest")
- Make test failures clear and informative
- Write just enough code to make the test pass - no more
- Once tests pass, consider if refactoring is needed
- Repeat the cycle for new functionality

# TIDY FIRST APPROACH

- Separate all changes into two distinct types:
  1. STRUCTURAL CHANGES: Rearranging code without changing behavior (removing duplicate headers, reordering middleware)
  2. BEHAVIORAL CHANGES: Adding or modifying actual functionality (CORS middleware implementation)
- Never mix structural and behavioral changes in the same commit
- Always make structural changes first when both are needed
- Validate structural changes do not alter behavior by running tests before and after

# COMMIT DISCIPLINE

- Only commit when:
  1. ALL tests are passing
  2. ALL compiler/linter warnings have been resolved
  3. The change represents a single logical unit of work
  4. Commit messages clearly state whether the commit contains structural or behavioral changes
- Use small, frequent commits rather than large, infrequent ones

# GO-SPECIFIC STANDARDS

- Use external test package (`package main_test` instead of `package main`)
- Always add `t.Parallel()` to all tests - if test is not parallelizable, fix the test design
- Follow middleware pattern from existing accesslog and recovery middleware
- Use httptest.NewRequest and httptest.NewRecorder for middleware unit tests
- Follow RealWorld API specifications exactly

# CORS IMPLEMENTATION REQUIREMENTS (ACTUAL)

## Implemented CORS Headers (Minimal Approach)

### Preflight Requests (OPTIONS method)
- ✅ `Access-Control-Allow-Origin: *` - Allow all origins for demo app
- ✅ Return `200 OK` status
- ⏳ TODO: `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- ⏳ TODO: `Access-Control-Allow-Headers: Content-Type, Authorization`
- ⏳ TODO: `Access-Control-Max-Age: 86400` (cache preflight for 24 hours)

### Actual Requests (GET, POST, PUT, DELETE)
- ✅ `Access-Control-Allow-Origin: *` - Allow all origins

**Design Decision**: Keep implementation minimal for this iteration. Additional headers can be added later when needed.

## Middleware Design (As Implemented)
- ✅ Follow existing middleware pattern (function returning http.Handler)
- ✅ Process OPTIONS requests and return early with 200 OK
- ✅ Add CORS headers to all requests (including non-OPTIONS)
- ✅ Added to middleware chain in route() function (before accesslog)
- ✅ Works with responseRecorder (used by accesslog middleware)

# TEST-DRIVEN IMPLEMENTATION STEPS

Follow these steps in order. Each step follows: Write failing test → Make it pass → Refactor if needed

## Phase 1: Preflight Handling (RED → GREEN)

### Step 1: Write failing test for OPTIONS preflight
- [ ] Create `TestCORSPreflightRequest` test function
- [ ] Create `httpOptions()` helper function (or use httptest if preferred)
- [ ] Verify OPTIONS request returns 200 OK
- [ ] Verify Access-Control-Allow-Origin header is set to "*"
- [ ] Add TODO comment for additional headers (minimal approach)
- [ ] Run test - should FAIL (RED phase)
- [ ] Commit: "test: add failing test for CORS preflight handling"

### Step 2: Implement basic CORS middleware
- [ ] Create `cors()` middleware function
- [ ] Set `Access-Control-Allow-Origin: *` for all requests
- [ ] Handle OPTIONS method by returning early with 200 OK
- [ ] Integrate into route() middleware chain (before accesslog)
- [ ] Run test - should PASS (GREEN phase)
- [ ] Commit: "feat: implement basic CORS middleware for OPTIONS requests"

## Phase 2: Verify All Request Types

### Step 3: Test CORS headers on actual requests
- [ ] Create `TestCORSActualRequests` with subtests for GET and POST
- [ ] Verify headers present on different HTTP methods
- [ ] Test may pass immediately if middleware already sets headers globally
- [ ] Commit: "test: verify CORS headers on actual requests"

## Phase 3: Refactoring & Cleanup

### Step 4: Remove duplicate headers (Tidy First)
- [ ] Search for duplicate `Access-Control-Allow-Origin` headers in code
- [ ] Remove duplicates (middleware now handles it globally)
- [ ] Run all tests to verify behavior unchanged
- [ ] Commit: "refactor: remove duplicate CORS headers"

### Step 5: Fix any linter issues
- [ ] Run linter (`make lint` or equivalent)
- [ ] Fix any warnings (e.g., unclosed response bodies)
- [ ] Commit: "fix: resolve linter warnings"

## Phase 4: Final Verification

### Step 6: Comprehensive testing
- [ ] Run full test suite with race detection
- [ ] Run linter
- [ ] Manual verification with curl or browser
- [ ] All tests pass with `t.Parallel()`

**Common Issues to Watch For:**
- Data races in concurrent handlers (use `-race` flag)
- Unclosed response bodies in tests
- Pre-existing duplicate CORS headers

# MIDDLEWARE PATTERN REFERENCE

Based on existing middleware (accesslog, recovery):

```go
func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        // Handle OPTIONS method
        // Call next handler
    })
}
```

Integration in route():
```go
handler := cors(mux)
handler = accesslog(handler, log)
handler = recovery(handler, log)
```

# TEST PATTERN REFERENCE

## Option 1: Integration Test (Real HTTP Requests)

**Recommended for consistency with existing integration tests.**

```go
// TestCORSPreflightRequest tests CORS middleware handles OPTIONS preflight requests
func TestCORSPreflightRequest(t *testing.T) {
    t.Parallel()

    res := httpOptions(t, "/api/users")
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        t.Errorf("want 200, got %d", res.StatusCode)
    }
    if got := res.Header.Get("Access-Control-Allow-Origin"); got != "*" {
        t.Errorf("want *, got %s", got)
    }
}

// Helper function for OPTIONS requests
func httpOptions(t *testing.T, path string) *http.Response {
    t.Helper()
    req, _ := http.NewRequest(http.MethodOptions, endpoint+path, nil)
    res, _ := http.DefaultClient.Do(req)
    return res
}
```

## Option 2: Unit Test (httptest)

**Faster but tests middleware in isolation.**

```go
func TestCORSMiddleware(t *testing.T) {
    t.Parallel()

    handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest(http.MethodOptions, "/test", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("want 200, got %d", rec.Code)
    }
    if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
        t.Errorf("want *, got %s", got)
    }
}
```

**Choose based on your project's testing style.**

# EXAMPLE WORKFLOW

When approaching each test:
1. Write a simple failing test for a small part of the feature
2. Run test to confirm it fails (RED)
3. Implement the bare minimum to make it pass (GREEN)
4. Run all tests to confirm they pass
5. Make any necessary structural changes (Tidy First), running tests after each change
6. Commit structural changes separately if any
7. Commit behavioral change
8. Move to next test

# VERIFICATION COMMANDS

```bash
# Setup
git checkout -b feat-cors
git status

# After each test cycle
go test ./...
go test -v -run TestCORSMiddleware

# During development
go test -v ./...  # Run all tests
make test        # Run with race detection and coverage

# Manual verification after implementation
curl -i -X OPTIONS http://localhost:8080/api/users
curl -i http://localhost:8080/health
curl -i http://localhost:8080/openapi.yaml

# Final verification
make test
make lint
```

# EXPECTED SCOPE

## Files to Modify
- Main server file (e.g., `main.go`) - Add cors() middleware function (~15-20 lines)
- Test file (e.g., `main_test.go`) - Add test functions (~40-50 lines)

## Expected Metrics
- Total implementation: ~60-70 lines of code
- Functions added: 2 (cors middleware, test helper)
- Tests added: 2 test functions minimum
- Commits: 5-7 commits following TDD cycle
- Coverage: Should increase slightly (depends on existing coverage)

# CORS SPECIFICATION REFERENCE

## Why CORS is Needed

Web APIs need CORS when:
- Frontend applications run on different origins (different domains/ports)
- Interactive documentation interfaces (Swagger/OpenAPI UI)
- Development environments with different ports (frontend:3000, backend:8080)

## Standard CORS Headers

**Minimal (Start Here)**:
- `Access-Control-Allow-Origin: *` - Allow all origins (use specific origin in production)

**Complete (Add When Needed)**:
- `Access-Control-Allow-Methods` - Which HTTP methods are allowed (GET, POST, PUT, DELETE, OPTIONS)
- `Access-Control-Allow-Headers` - Which headers can be sent (Content-Type, Authorization)
- `Access-Control-Max-Age` - How long to cache preflight response (86400 = 24 hours)

# SUCCESS CRITERIA

- [ ] All tests pass with `t.Parallel()`
- [ ] CORS middleware handles OPTIONS preflight requests
- [ ] CORS headers present on all API endpoints
- [ ] Duplicate headers removed from existing handlers
- [ ] Integration tests verify end-to-end functionality
- [ ] Linter passes with no warnings
- [ ] No data races detected (`go test -race`)
- [ ] Code follows existing middleware patterns
- [ ] Commits follow TDD cycle (RED → GREEN → Refactor)

---

**Remember**: Prioritize clean, well-tested code over quick implementation. Always write one test at a time, make it run, then improve structure.
