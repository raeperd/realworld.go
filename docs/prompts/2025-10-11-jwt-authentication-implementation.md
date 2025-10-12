# JWT Authentication Implementation - October 11, 2025

**Pull Request:** https://github.com/raeperd/realworld.go/pull/3

## Original User Request Context

### Initial Request
```
Plan to implement JWT authentication token in response of POST api/users response

checkout to new branch, feat-jwt 
currently, @user.go is not hashing password. fix it to do using TDD method
use @https://github.com/golang-jwt/jwt 
try simplest, concise as possible

see: @https://realworld-docs.netlify.app/specifications/backend/endpoints/ 
```

### User Clarifications
1. **JWT Secret Management**: Get secret by CLI options just like port in main.go and main_test.go
2. **Token Expiration**: 7 days expiration is acceptable
3. **Implementation Focus**: Focus on JWT without password hashing initially
4. **Testing Approach**: Simple JWT unit test would be enough
5. **Testing Standards**: Always add `t.Parallel()` - if test is not parallelizable, that's a problem to solve
6. **Package Structure**: Use `auth_test` package instead of `auth` when testing (external testing pattern)
7. **Methodology**: Follow Kent Beck's TDD methodology with Red → Green → Refactor + Tidy First principles

---

## Methodology

**Follow TDD principles from @TDD.md** - Write failing test → Make it pass → Refactor

## Project-Specific Notes

- JWT secret passed from main through handlers via dependency injection
- Follow RealWorld API specifications exactly for response format

## JWT Implementation Requirements

- Use HS256 signing method with `github.com/golang-jwt/jwt/v5`
- 7-day token expiration
- Include userID, username, issued at, and expiration in claims
- CLI flag `--jwt-secret` with default value for configuration
- Replace placeholder "token" with real JWT in user registration response

## Test-Driven Implementation Order

Each step follows: Write failing test → Make it pass → Refactor if needed

1. **[ ] Test: JWT token generation with valid inputs**
   - Write test for `GenerateToken(userID, username, secret)` function
   - Implement basic JWT generation with HS256 and 7-day expiration

2. **[ ] Test: JWT token parsing and validation**
   - Write test for `ParseToken(token, secret)` function
   - Implement token parsing with claims extraction

3. **[ ] Test: User registration returns real JWT token**
   - Write integration test verifying POST /api/users returns valid JWT
   - Add CLI flag, dependency injection, and replace placeholder token

4. **[ ] Test: End-to-end JWT validation**
   - Write test that validates returned JWT contains correct user data
   - Ensure complete integration works

## Verification Commands

```bash
# Setup
git checkout -b feat-jwt
go get github.com/golang-jwt/jwt/v5

# After each test cycle
go test ./...
go test -v ./internal/auth/
go test -v -run TestPostUsers_ReturnsValidJWT

# Final verification
go run main.go --jwt-secret="test-key"
```
