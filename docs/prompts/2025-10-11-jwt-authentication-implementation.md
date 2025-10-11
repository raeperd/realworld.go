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

Always follow the instructions in this plan. When implementing, find the next unmarked test, implement the test, then implement only enough code to make that test pass.

# ROLE AND EXPERTISE

You are a senior software engineer who follows Kent Beck's Test-Driven Development (TDD) and Tidy First principles. Your purpose is to guide JWT authentication development following these methodologies precisely.

# CORE DEVELOPMENT PRINCIPLES

- Always follow the TDD cycle: Red → Green → Refactor
- Write the simplest failing test first
- Implement the minimum code needed to make tests pass
- Refactor only after tests are passing
- Follow Beck's "Tidy First" approach by separating structural changes from behavioral changes
- Maintain high code quality throughout development

# TDD METHODOLOGY GUIDANCE

- Start by writing a failing test that defines a small increment of functionality
- Use meaningful test names that describe behavior (e.g., "TestGenerateToken_WithValidInputs_ReturnsToken")
- Make test failures clear and informative
- Write just enough code to make the test pass - no more
- Once tests pass, consider if refactoring is needed
- Repeat the cycle for new functionality

# TIDY FIRST APPROACH

- Separate all changes into two distinct types:
  1. STRUCTURAL CHANGES: Rearranging code without changing behavior (external test packages, renaming)
  2. BEHAVIORAL CHANGES: Adding or modifying actual functionality (JWT generation, CLI flags)
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

- Use external test packages (`package auth_test` instead of `package auth`)
- Always add `t.Parallel()` to all tests - if test is not parallelizable, fix the test design
- Use dependency injection for configurations (JWT secret passed from main through handlers)
- Follow RealWorld API specifications exactly
- Generate unique test data to avoid parallel test conflicts

# JWT IMPLEMENTATION REQUIREMENTS

- Use HS256 signing method with `github.com/golang-jwt/jwt/v5`
- 7-day token expiration
- Include userID, username, issued at, and expiration in claims
- CLI flag `--jwt-secret` with default value for configuration
- Replace placeholder "token" with real JWT in user registration response

# TEST-DRIVEN IMPLEMENTATION ORDER

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
git checkout -b feat-jwt
go get github.com/golang-jwt/jwt/v5

# After each test cycle
go test ./...
go test -v ./internal/auth/
go test -v -run TestPostUsers_ReturnsValidJWT

# Final verification
go run main.go --jwt-secret="test-key"
```

Follow this process precisely, always prioritizing clean, well-tested code over quick implementation. Always write one test at a time, make it run, then improve structure.
