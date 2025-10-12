# POST /api/users/login Endpoint Implementation with TDD

A guide for implementing the user login endpoint following Test-Driven Development and Kent Beck's Tidy First principles.

## Status & Links
- **Status**: In Progress
- **PR Link**: (will be added when created)

## Context

### Original Request
Implement `POST /api/users/login` endpoint following TDD methodology as specified in the RealWorld API specification.

### API Specification (from spec.md)

**Endpoint**: `POST /api/users/login`

**Request Body**:
```json
{
  "user": {
    "email": "jake@jake.jake",
    "password": "jakejake"
  }
}
```

**Response**: Returns a User object (same format as registration)
```json
{
  "user": {
    "email": "jake@jake.jake",
    "token": "jwt.token.here",
    "username": "jake",
    "bio": "I work at statefarm",
    "image": null
  }
}
```

**Required fields**: email, password

**Status Codes**:
- `200 OK` - Successful authentication
- `401 Unauthorized` - Invalid credentials (email not found or wrong password)
- `422 Unprocessable Entity` - Validation errors (missing required fields)

---

Always follow the instructions in this plan. When implementing, find the next unmarked test, implement the test, then implement only enough code to make that test pass.

# METHODOLOGY

See **@docs/prompts/TDD.md** for complete TDD methodology. Follow the Red → Green → Refactor cycle strictly.

# FEATURE-SPECIFIC REQUIREMENTS

## Authentication Logic

1. **Email Lookup**: Use existing `GetUserByEmail` query from `internal/sqlite/query.sql`
2. **Password Verification**: Simple string comparison (plain text passwords for now)
3. **JWT Generation**: Reuse `auth.GenerateToken` with 7-day expiration
4. **Security**: Don't reveal whether email exists (generic "invalid credentials" message)

## Code Reuse Opportunities

- **Database**: Existing `GetUserByEmail` query
- **JWT**: Existing `auth.GenerateToken` function
- **Response Format**: Reuse `userPostResponseBody` from registration
- **Error Handling**: Reuse `encodeErrorResponse` pattern
- **Middleware Chain**: Already has CORS, accesslog, recovery

## Design Decisions

1. **Password Storage**: Currently plain text (no hashing yet) - compare directly
2. **Error Messages**: Generic "invalid credentials" for both non-existent user and wrong password
3. **Response Structure**: Identical to registration response (maintain consistency)
4. **Test Isolation**: Use unique email/username per test run for parallel test safety

# IMPLEMENTATION STEPS

Follow these steps in order. Each step follows: Write failing test → Make it pass → Refactor if needed

## Phase 1: Request Validation (RED → GREEN)

### Step 1: Write failing tests for validation
- [ ] Create `TestPostUsersLogin_Validation` in `user_test.go`
- [ ] Add test cases for:
  - Missing email returns 422
  - Missing password returns 422
- [ ] Create helper function `httpPostUsersLogin(t, email, password) *http.Response`
- [ ] Run test - should FAIL (RED phase)
- [ ] Commit: "test: add failing validation tests for POST /api/users/login"

### Step 2: Implement handler with validation only
- [ ] Create `HandlePostUsersLogin` function in `internal/api/user.go`
- [ ] Add `userLoginRequestBody` struct:
  ```go
  type userLoginRequestBody struct {
      User struct {
          Email    string `json:"email"`
          Password string `json:"password"`
      } `json:"user"`
  }
  ```
- [ ] Add `Validate()` method checking email and password not empty
- [ ] Return 422 with error messages for validation failures
- [ ] Register route in `main.go` route() function:
  ```go
  mux.HandleFunc("POST /api/users/login", api.HandlePostUsersLogin(db, jwtSecret))
  ```
- [ ] Run test - should PASS (GREEN phase)
- [ ] Run all tests to ensure nothing broke
- [ ] Commit: "feat: add POST /api/users/login with request validation"

## Phase 2: Non-Existent User (RED → GREEN)

### Step 3: Write failing test for user not found
- [ ] Add `TestPostUsersLogin_UserNotFound` test
- [ ] Attempt login with email that doesn't exist
- [ ] Verify returns 401 Unauthorized status
- [ ] Verify error response has appropriate message
- [ ] Run test - should FAIL (RED phase)
- [ ] Commit: "test: add failing test for non-existent user login"

### Step 4: Implement database lookup
- [ ] In `HandlePostUsersLogin`, query database using `queries.GetUserByEmail`
- [ ] If `sql.ErrNoRows`, return 401 Unauthorized with "invalid credentials" message
- [ ] If other database error, return 500 Internal Server Error
- [ ] Run test - should PASS (GREEN phase)
- [ ] Run all tests
- [ ] Commit: "feat: return 401 for non-existent user on login"

## Phase 3: Wrong Password (RED → GREEN)

### Step 5: Write failing test for wrong password
- [ ] Add `TestPostUsersLogin_WrongPassword` test
- [ ] First create a user via registration (setup)
- [ ] Attempt login with same email but wrong password
- [ ] Verify returns 401 Unauthorized status
- [ ] Run test - should FAIL (RED phase)
- [ ] Commit: "test: add failing test for wrong password on login"

### Step 6: Implement password verification
- [ ] Compare `request.User.Password` with `user.Password` from database
- [ ] If passwords don't match, return 401 with "invalid credentials"
- [ ] Run test - should PASS (GREEN phase)
- [ ] Run all tests
- [ ] Commit: "feat: verify password and return 401 on mismatch"

## Phase 4: Successful Login (RED → GREEN)

### Step 7: Write failing test for successful login
- [ ] Add `TestPostUsersLogin_Success` test
- [ ] Register user with known credentials
- [ ] Login with correct email and password
- [ ] Verify 200 OK status
- [ ] Verify response body contains:
  - Correct email
  - Correct username
  - Valid JWT token (not empty, not placeholder)
  - Bio and image fields (may be empty strings)
- [ ] Run test - should FAIL (RED phase)
- [ ] Commit: "test: add failing test for successful login"

### Step 8: Implement successful login response
- [ ] Generate JWT token using `auth.GenerateToken(user.ID, user.Username, jwtSecret)`
- [ ] Build response using existing response pattern
- [ ] Return 200 OK with user data wrapped in `{"user": {...}}`
- [ ] Run test - should PASS (GREEN phase)
- [ ] Run all tests
- [ ] Commit: "feat: return user with JWT token on successful login"

## Phase 5: JWT Token Validation (RED → GREEN)

### Step 9: Write test to verify JWT token contents
- [ ] Add `TestPostUsersLogin_ReturnsValidJWT` test
- [ ] Register user and login
- [ ] Parse the returned JWT token using `auth.ParseToken`
- [ ] Verify token contains correct username
- [ ] Verify token contains correct user_id (> 0)
- [ ] Run test - should PASS immediately if Step 8 used `auth.GenerateToken` correctly
- [ ] Commit: "test: verify login JWT contains correct user data"

## Phase 6: Refactoring (Tidy First)

### Step 10: Review for code duplication
- [ ] Compare `HandlePostUsersLogin` with `HandlePostUsers`
- [ ] Check if user response building can be extracted
- [ ] Check if error handling patterns can be shared
- [ ] If no significant duplication, skip extraction (YAGNI principle)
- [ ] Run all tests to verify behavior unchanged
- [ ] Commit: "refactor: extract common patterns" (only if refactoring done)

### Step 11: Update response body type if needed
- [ ] Check if `userPostResponseBody` should be renamed to `userResponseBody`
- [ ] Or create type alias if both handlers share the response format
- [ ] Update `responseBody` interface constraint to include login response
- [ ] Run all tests to verify no breakage
- [ ] Commit: "refactor: unify user response types" (only if changes made)

## Phase 7: Final Verification

### Step 12: Comprehensive testing
- [ ] Run full test suite: `make test`
- [ ] Run with race detection: `go test -race ./...`
- [ ] Run linter: `make lint`
- [ ] Fix any warnings or issues
- [ ] Verify all tests use `t.Parallel()`
- [ ] Commit: "fix: resolve linter warnings" (if any fixes)

### Step 13: Manual integration testing
- [ ] Start server: `make run`
- [ ] Register a test user (see commands below)
- [ ] Login with correct credentials
- [ ] Try login with wrong password
- [ ] Try login with non-existent email
- [ ] Verify all responses match specification

# VERIFICATION COMMANDS

```bash
# Setup branch
git checkout -b feat-login-endpoint
git status

# During TDD cycles (run after each step)
go test -v -run TestPostUsersLogin_Validation
go test -v -run TestPostUsersLogin_UserNotFound
go test -v -run TestPostUsersLogin_WrongPassword
go test -v -run TestPostUsersLogin_Success
go test -v -run TestPostUsersLogin_ReturnsValidJWT

# Run all user tests
go test -v -run TestPostUsers

# Full test suite
go test ./...
make test

# Final verification with race detection
go test -race ./...
make lint

# Manual testing (start server first: make run)
# 1. Register a user
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"user":{"username":"testuser","email":"test@example.com","password":"testpass123"}}'

# 2. Login with correct credentials (should return 200 with token)
curl -i -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"user":{"email":"test@example.com","password":"testpass123"}}'

# 3. Login with wrong password (should return 401)
curl -i -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"user":{"email":"test@example.com","password":"wrongpassword"}}'

# 4. Login with non-existent email (should return 401)
curl -i -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"user":{"email":"notfound@example.com","password":"anypassword"}}'

# 5. Missing required fields (should return 422)
curl -i -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"user":{"email":"test@example.com"}}'
```

# EXPECTED SCOPE

## Files to Modify

1. **internal/api/user.go** (~50-60 lines added)
   - `HandlePostUsersLogin` function (~40 lines)
   - `userLoginRequestBody` struct with Validate() method (~15 lines)
   - Reuse existing `userPostResponseBody` and `encodeErrorResponse`

2. **user_test.go** (~80-100 lines added)
   - `TestPostUsersLogin_Validation` (~20 lines)
   - `TestPostUsersLogin_UserNotFound` (~15 lines)
   - `TestPostUsersLogin_WrongPassword` (~20 lines)
   - `TestPostUsersLogin_Success` (~20 lines)
   - `TestPostUsersLogin_ReturnsValidJWT` (~20 lines)
   - `httpPostUsersLogin` helper function (~10 lines)

3. **main.go** (1 line added)
   - Register route in `route()` function

## Database Queries
**No new queries needed** - reuse existing:
- `GetUserByEmail` from `internal/sqlite/query.sql.go`

## Expected Metrics
- Total implementation: ~130-160 lines of code
- Functions added: 2 (handler + helper)
- Structs added: 1 (request body)
- Tests added: 5 test functions
- Commits: 8-12 commits following TDD cycle
- Test coverage: Should maintain or improve existing coverage

# CODE PATTERNS REFERENCE

## Handler Pattern (Follow HandlePostUsers)

```go
func HandlePostUsersLogin(db *sql.DB, jwtSecret string) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Decode request
        var request userLoginRequestBody
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        defer func() { _ = r.Body.Close() }()

        // 2. Validate request
        if errs := request.Validate(); len(errs) > 0 {
            encodeErrorResponse(r.Context(), http.StatusUnprocessableEntity, errs, w)
            return
        }

        // 3. Database transaction
        tx, err := db.BeginTx(r.Context(), nil)
        // ... handle transaction

        // 4. Query user by email
        queries := sqlite.New(tx)
        user, err := queries.GetUserByEmail(r.Context(), request.User.Email)
        // ... handle not found or error

        // 5. Verify password
        if user.Password != request.User.Password {
            encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("invalid credentials")}, w)
            return
        }

        // 6. Generate JWT
        token, err := auth.GenerateToken(user.ID, user.Username, jwtSecret)
        // ... handle error

        // 7. Commit transaction
        if err := tx.Commit(); err != nil {
            // ... handle error
        }

        // 8. Return success response
        encodeResponse(r.Context(), http.StatusOK, userPostResponseBody{
            Email:    user.Email,
            Token:    token,
            Username: user.Username,
            Bio:      user.Bio.String,
            Image:    user.Image.String,
        }, w)
    }
}
```

## Test Pattern (Follow TestPostUsers)

```go
func TestPostUsersLogin_Success(t *testing.T) {
    t.Parallel()

    // Setup: Create user via registration
    unique := fmt.Sprintf("%d", time.Now().UnixNano())
    email := fmt.Sprintf("login_test_%s@example.com", unique)
    password := "testpass123"

    regReq := UserPostRequestBody{
        Username: "login_user_" + unique,
        Email:    email,
        Password: password,
    }
    regRes := httpPostUsers(t, regReq)
    test.Equal(t, http.StatusCreated, regRes.StatusCode)
    t.Cleanup(func() { _ = regRes.Body.Close() })

    // Test: Login with same credentials
    loginRes := httpPostUsersLogin(t, email, password)
    test.Equal(t, http.StatusOK, loginRes.StatusCode)
    t.Cleanup(func() { _ = loginRes.Body.Close() })

    // Verify response
    var response UserResponseBody
    test.Nil(t, json.NewDecoder(loginRes.Body).Decode(&response))
    test.Equal(t, email, response.Email)
    test.Equal(t, regReq.Username, response.Username)
    test.NotEqual(t, "", response.Token)
}

func httpPostUsersLogin(t *testing.T, email, password string) *http.Response {
    t.Helper()

    requestBody := struct {
        User struct {
            Email    string `json:"email"`
            Password string `json:"password"`
        } `json:"user"`
    }{}
    requestBody.User.Email = email
    requestBody.User.Password = password

    body, err := json.Marshal(requestBody)
    test.Nil(t, err)

    res, err := http.Post(endpoint+"/api/users/login", "application/json", bytes.NewBuffer(body))
    test.Nil(t, err)

    return res
}
```

# COMMON PITFALLS TO AVOID

1. **Don't hash passwords yet** - Current implementation uses plain text, maintain consistency
2. **Don't reveal email existence** - Use generic "invalid credentials" for security
3. **Don't skip transaction** - Even for read-only operations, use transactions for consistency
4. **Don't forget t.Parallel()** - All tests must be parallelizable
5. **Don't forget unique test data** - Use timestamps/UUIDs to avoid conflicts
6. **Don't forget t.Cleanup()** - Always close response bodies in tests
7. **Don't skip error checking** - Check all database and encoding errors

# SUCCESS CRITERIA

- [ ] All tests pass with `t.Parallel()`
- [ ] Endpoint validates required fields (email, password)
- [ ] Returns 422 for missing email or password
- [ ] Returns 401 for non-existent users
- [ ] Returns 401 for wrong passwords
- [ ] Returns 200 with valid JWT on successful login
- [ ] JWT token contains correct user_id and username
- [ ] Response format matches RealWorld specification
- [ ] No linter warnings (`make lint`)
- [ ] No data races (`go test -race`)
- [ ] Code follows existing patterns from HandlePostUsers
- [ ] All commits follow TDD cycle (RED → GREEN → Refactor)
- [ ] Manual curl tests work as expected

---

**Remember**: Write one test at a time, make it run, then improve structure. Always run all tests after each change. Follow the TDD cycle strictly.
