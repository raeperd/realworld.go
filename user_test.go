package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/raeperd/realworld.go/internal/auth"
	"github.com/raeperd/test"
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
