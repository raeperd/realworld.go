# CORS Middleware Implementation - October 12, 2025

**Status:** Completed

**Pull Request:** https://github.com/raeperd/realworld.go/pull/7

## Original User Request Context

### Initial Request
```
plan to add CORS middleware in @main.go in TDD way
first checkout to new branch feat-cors
```

### Project Context
- RealWorld API specification requires CORS for frontend interoperability
- README.md already claims "CORS Support: Cross-origin resource sharing enabled" (line 42)
- Currently only OpenAPI endpoint has a single CORS header (`Access-Control-Allow-Origin: *`)
- Need proper CORS middleware following existing patterns (accesslog, recovery)

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

# CORS IMPLEMENTATION REQUIREMENTS

## Required CORS Headers

### Preflight Requests (OPTIONS method)
- `Access-Control-Allow-Origin: *` - Allow all origins for demo app
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS` - All methods used by RealWorld
- `Access-Control-Allow-Headers: Content-Type, Authorization` - Headers used by RealWorld API
- `Access-Control-Max-Age: 86400` - Cache preflight for 24 hours
- Return `200 OK` status

### Actual Requests (GET, POST, PUT, DELETE)
- `Access-Control-Allow-Origin: *` - Allow all origins

## Middleware Design
- Follow existing middleware pattern (function returning http.Handler)
- Process OPTIONS requests and return early with 200 OK
- Add CORS headers to all requests (including non-OPTIONS)
- Should be added to middleware chain in route() function
- Must work with responseRecorder (used by accesslog middleware)

# TEST-DRIVEN IMPLEMENTATION ORDER

Each step follows: Write failing test → Make it pass → Refactor if needed

## Phase 1: Preflight Handling

### 1. [ ] Test: CORS middleware handles OPTIONS preflight request
   - Write test `TestCORSMiddleware_PreflightRequest` in main_test.go
   - Verify OPTIONS request returns 200 OK
   - Verify Access-Control-Allow-Origin header is set
   - Verify Access-Control-Allow-Methods header is set
   - Verify Access-Control-Allow-Headers header is set
   - Verify Access-Control-Max-Age header is set

### 2. [ ] Implementation: Basic CORS middleware for OPTIONS
   - Create `cors()` middleware function in main.go
   - Check if request method is OPTIONS
   - Set required CORS headers
   - Return 200 OK for OPTIONS
   - Call next handler for non-OPTIONS
   - **Run test** to confirm it passes (GREEN)
   - **Commit**: "test: add failing test for CORS preflight handling" (RED phase commit)
   - **Commit**: "feat: implement basic CORS preflight handling" (GREEN phase commit)

## Phase 2: Actual Request Headers

### 3. [ ] Test: CORS middleware adds headers to actual requests
   - Write test `TestCORSMiddleware_ActualRequest` in main_test.go
   - Test GET request includes Access-Control-Allow-Origin header
   - Test POST request includes Access-Control-Allow-Origin header
   - Test PUT request includes Access-Control-Allow-Origin header
   - Test DELETE request includes Access-Control-Allow-Origin header

### 4. [ ] Implementation: CORS headers on all requests
   - Enhance `cors()` middleware to set headers before calling next handler
   - **Run test** to confirm it passes (GREEN)
   - **Commit**: "test: add failing test for CORS headers on actual requests" (RED phase commit)
   - **Commit**: "feat: add CORS headers to all requests" (GREEN phase commit)

## Phase 3: Integration & Cleanup

### 5. [ ] Structural: Integrate middleware and remove duplicates
   - Add `cors()` middleware to route() function chain (before accesslog)
   - Remove duplicate `Access-Control-Allow-Origin` header from handleGetOpenAPI
   - **Run all tests** to verify behavior unchanged
   - **Commit**: "refactor: integrate CORS middleware and remove duplicate header"

### 6. [ ] Integration Test: Verify CORS works end-to-end
   - Write test `TestGetOpenAPI_HasCORSHeaders` to verify OpenAPI endpoint
   - Write test `TestPostUsers_HasCORSHeaders` to verify API endpoint
   - **Run all tests** including integration tests
   - **Commit**: "test: add integration tests for CORS headers"

## Phase 4: Verification

### 7. [ ] Final verification
   - Run full test suite: `make test`
   - Run linter: `make lint`
   - Manual verification with curl
   - All tests pass with `t.Parallel()`

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

Based on TestAccessLogMiddleware and TestRecoveryMiddleware:

```go
func TestCORSMiddleware_PreflightRequest(t *testing.T) {
    t.Parallel()

    handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest(http.MethodOptions, "/test", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    test.Equal(t, http.StatusOK, rec.Code)
    test.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
    // ... more assertions
}
```

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

# EXPECTED CHANGES

## Files to Modify
- `main.go` - Add cors() middleware function (~20 lines)
- `main_test.go` - Add TestCORSMiddleware tests (~80-100 lines)

## Code Metrics
- Total lines added: ~100-120 lines
- Functions added: 1 (cors middleware)
- Tests added: 3-4 test functions
- Commits expected: 6-8 commits following TDD cycle

# CORS SPECIFICATION REFERENCE

RealWorld API requires CORS for:
- Frontend applications (React, Angular, Vue) running on different origins
- Swagger UI documentation interface
- Development environments with different ports

Standard CORS headers:
- `Access-Control-Allow-Origin`: Which origins can access
- `Access-Control-Allow-Methods`: Which HTTP methods are allowed
- `Access-Control-Allow-Headers`: Which headers can be sent
- `Access-Control-Max-Age`: How long to cache preflight response

# SUCCESS CRITERIA

- [x] All tests pass with `t.Parallel()`
- [x] CORS middleware handles OPTIONS preflight requests
- [x] CORS headers present on all API endpoints
- [x] Duplicate header removed from handleGetOpenAPI
- [x] Integration tests verify end-to-end functionality
- [x] Linter passes with no warnings
- [x] README.md claim "CORS Support" is now accurate
- [x] Code follows existing middleware patterns
- [x] Commits follow TDD cycle (RED → GREEN → Refactor)

Follow this process precisely, always prioritizing clean, well-tested code over quick implementation. Always write one test at a time, make it run, then improve structure.
