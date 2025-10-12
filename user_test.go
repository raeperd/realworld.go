package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/raeperd/test"

	"github.com/raeperd/realworld.go/internal/auth"
)

func TestPostUsers_Validation(t *testing.T) {
	t.Parallel()

	testcases := map[string]UserPostRequestBody{
		"username required": {
			Username: "",
			Email:    "test@test.com",
			Password: "test",
		},
		"email required": {
			Username: "test",
			Email:    "",
			Password: "test",
		},
		"password required": {
			Username: "test",
			Email:    "test@test.com",
			Password: "",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := httpPostUsers(t, tc)
			test.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

			t.Cleanup(func() { _ = res.Body.Close() })
		})
	}
}

func TestPostUsers_CreateUser(t *testing.T) {
	t.Parallel()

	// Generate unique username/email per test run to avoid conflicts
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	req := UserPostRequestBody{
		Username: "create_test_user_" + unique,
		Email:    fmt.Sprintf("create_test_%s@example.com", unique),
		Password: "testpass",
	}
	res := httpPostUsers(t, req)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response UserResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, req.Username, response.Username)
	test.Equal(t, req.Email, response.Email)

	res = httpPostUsers(t, req) // return conflict when user already exists
	test.Equal(t, http.StatusConflict, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPostUsers_ReturnsValidJWT(t *testing.T) {
	t.Parallel()

	// Given - generate unique username/email per test run to avoid conflicts
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	req := UserPostRequestBody{
		Username: "jwt_test_user_" + unique,
		Email:    fmt.Sprintf("jwt_test_%s@example.org", unique),
		Password: "testpass",
	}

	// When
	res := httpPostUsers(t, req)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response UserResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))

	// Then - token should not be the placeholder
	test.NotEqual(t, "token", response.Token)

	// Then - token should be parseable as valid JWT
	claims, err := auth.ParseToken(response.Token, "test-secret") // TODO: use actual secret
	test.Nil(t, err)

	// Then - JWT should contain correct user data
	test.Equal(t, req.Username, claims.Username)
	// UserID should be > 0 (actual DB ID, not placeholder)
	test.True(t, claims.UserID > 0)
}

func httpPostUsers(t *testing.T, request UserPostRequestBody) *http.Response {
	t.Helper()

	body, err := json.Marshal(UserWrapper[UserPostRequestBody]{User: request})
	test.Nil(t, err)

	res, err := http.Post(endpoint+"/api/users", "application/json", bytes.NewBuffer(body))
	test.Nil(t, err)

	return res
}

type UserWrapper[T UserPostRequestBody | UserLoginRequestBody | UserResponseBody] struct {
	User T `json:"user"`
}

type UserPostRequestBody struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponseBody struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// Login tests

func TestPostUsersLogin_Validation(t *testing.T) {
	t.Parallel()

	testcases := map[string]UserLoginRequestBody{
		"email required": {
			Email:    "",
			Password: "testpass",
		},
		"password required": {
			Email:    "test@test.com",
			Password: "",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := httpPostUsersLogin(t, tc)
			test.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

			t.Cleanup(func() { _ = res.Body.Close() })
		})
	}
}

func TestPostUsersLogin_Success(t *testing.T) {
	t.Parallel()

	// Generate unique username/email per test run to avoid conflicts
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	createReq := UserPostRequestBody{
		Username: "login_success_" + unique,
		Email:    fmt.Sprintf("login_success_%s@example.com", unique),
		Password: "testpass",
	}

	// Create user first
	res := httpPostUsers(t, createReq)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Now login with same credentials
	loginReq := UserLoginRequestBody{
		Email:    createReq.Email,
		Password: createReq.Password,
	}
	res = httpPostUsersLogin(t, loginReq)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response UserResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, createReq.Username, response.Username)
	test.Equal(t, createReq.Email, response.Email)
}

func TestPostUsersLogin_InvalidCredentials(t *testing.T) {
	t.Parallel()

	t.Run("non-existent email", func(t *testing.T) {
		t.Parallel()

		unique := fmt.Sprintf("%d", time.Now().UnixNano())
		loginReq := UserLoginRequestBody{
			Email:    fmt.Sprintf("nonexistent_%s@example.com", unique),
			Password: "testpass",
		}
		res := httpPostUsersLogin(t, loginReq)
		test.Equal(t, http.StatusUnauthorized, res.StatusCode)
		t.Cleanup(func() { _ = res.Body.Close() })
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()

		// Create user first
		unique := fmt.Sprintf("%d", time.Now().UnixNano())
		createReq := UserPostRequestBody{
			Username: "login_wrong_pass_" + unique,
			Email:    fmt.Sprintf("login_wrong_pass_%s@example.com", unique),
			Password: "correctpass",
		}
		res := httpPostUsers(t, createReq)
		test.Equal(t, http.StatusCreated, res.StatusCode)
		t.Cleanup(func() { _ = res.Body.Close() })

		// Try login with wrong password
		loginReq := UserLoginRequestBody{
			Email:    createReq.Email,
			Password: "wrongpass",
		}
		res = httpPostUsersLogin(t, loginReq)
		test.Equal(t, http.StatusUnauthorized, res.StatusCode)
		t.Cleanup(func() { _ = res.Body.Close() })
	})
}

func TestPostUsersLogin_ReturnsValidJWT(t *testing.T) {
	t.Parallel()

	// Create user first
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	createReq := UserPostRequestBody{
		Username: "login_jwt_" + unique,
		Email:    fmt.Sprintf("login_jwt_%s@example.com", unique),
		Password: "testpass",
	}
	res := httpPostUsers(t, createReq)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Login
	loginReq := UserLoginRequestBody{
		Email:    createReq.Email,
		Password: createReq.Password,
	}
	res = httpPostUsersLogin(t, loginReq)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response UserResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))

	// Token should be parseable as valid JWT
	claims, err := auth.ParseToken(response.Token, "test-secret")
	test.Nil(t, err)

	// JWT should contain correct user data
	test.Equal(t, createReq.Username, claims.Username)
	test.True(t, claims.UserID > 0)
}

func httpPostUsersLogin(t *testing.T, request UserLoginRequestBody) *http.Response {
	t.Helper()

	body, err := json.Marshal(UserWrapper[UserLoginRequestBody]{User: request})
	test.Nil(t, err)

	res, err := http.Post(endpoint+"/api/users/login", "application/json", bytes.NewBuffer(body))
	test.Nil(t, err)

	return res
}

type UserLoginRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
