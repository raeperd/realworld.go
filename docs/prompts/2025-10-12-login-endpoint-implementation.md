# Login Endpoint Implementation - October 12, 2025

## User Request Context

### Initial Request
```
Plan to implement POST /api/users/login in TDD way
see @TDD.md @spec.md for reference
```

### User Clarifications
1. Implement login first with plain text password comparison
2. Leave bcrypt password hashing as TODO for later

---

## Methodology

**Follow TDD principles from @TDD.md** - Write failing test → Make it pass → Refactor

## Login-Specific Requirements

### Endpoint: `POST /api/users/login`

**Request**: `{"user":{"email":"...","password":"..."}}`  
**Response**: Same as registration - `{"user":{"email":"...","token":"...","username":"...","bio":"...","image":"..."}}`

**Status Codes**:
- 200 OK - Successful login
- 401 Unauthorized - Invalid email or password
- 422 Unprocessable Entity - Validation errors

**Security Note**: Uses plain text password comparison (TODO: implement bcrypt hashing)

## Implementation Steps

Each step follows: Write failing test → Make it pass → Refactor

### Phase 1: Write Failing Tests (RED)
- [x] `TestPostUsersLogin_Validation` - test missing email, missing password (expect 422)
- [x] `TestPostUsersLogin_Success` - create user, then login with same credentials (expect 200)
- [x] `TestPostUsersLogin_InvalidCredentials` - test wrong email/password (expect 401)
- [x] `TestPostUsersLogin_ReturnsValidJWT` - verify JWT is valid and contains correct user data
- [x] Add `httpPostUsersLogin()` helper function following `httpPostUsers()` pattern

### Phase 2: Minimal Implementation (GREEN)
- [x] Add SQL query to `internal/sqlite/query.sql`:
```sql
-- name: GetUserByEmailAndPassword :one
SELECT * FROM users WHERE email = ? AND password = ?;
```
- [x] Run `sqlc generate`
- [x] Create `userLoginRequestBody` with `Validate()` method in `internal/api/user.go`
- [x] Implement `HandlePostUsersLogin(db, jwtSecret)` - decode, validate, query, generate JWT, return response
- [x] Register route in `main.go`: `mux.HandleFunc("POST /api/users/login", api.HandlePostUsersLogin(db, jwtSecret))`
- [x] Run `go test ./...` - all tests should PASS

### Phase 3: Refactor
- [x] Check for code duplication - reused existing response types and helpers
- [x] Add TODO comment for password hashing implementation
- [x] Update `responseBody` type union if needed - no changes needed, reusing `userPostResponseBody`

## Verification Commands

```bash
# After each step
go test ./...

# Specific tests
go test -v -run TestPostUsersLogin

# With race detection
go test -race ./...

# Generate sqlc code
sqlc generate
```

