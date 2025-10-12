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

type UserWrapper[T UserPostRequestBody | UserResponseBody] struct {
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

func TestPostUsersLogin_Validation(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		email    string
		password string
	}{
		"email required": {
			email:    "",
			password: "testpass",
		},
		"password required": {
			email:    "test@test.com",
			password: "",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := httpPostUsersLogin(t, tc.email, tc.password)
			test.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)

			t.Cleanup(func() { _ = res.Body.Close() })
		})
	}
}

func TestPostUsersLogin_UserNotFound(t *testing.T) {
	t.Parallel()

	// Attempt login with email that doesn't exist
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("nonexistent_%s@example.com", unique)

	res := httpPostUsersLogin(t, email, "anypassword")
	test.Equal(t, http.StatusUnauthorized, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPostUsersLogin_WrongPassword(t *testing.T) {
	t.Parallel()

	// Setup: Create user via registration
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	email := fmt.Sprintf("wrongpw_test_%s@example.com", unique)
	correctPassword := "correctpass123"

	regReq := UserPostRequestBody{
		Username: "wrongpw_user_" + unique,
		Email:    email,
		Password: correctPassword,
	}
	regRes := httpPostUsers(t, regReq)
	test.Equal(t, http.StatusCreated, regRes.StatusCode)
	t.Cleanup(func() { _ = regRes.Body.Close() })

	// Test: Login with wrong password
	wrongPassword := "wrongpassword"
	loginRes := httpPostUsersLogin(t, email, wrongPassword)
	test.Equal(t, http.StatusUnauthorized, loginRes.StatusCode)
	t.Cleanup(func() { _ = loginRes.Body.Close() })
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
