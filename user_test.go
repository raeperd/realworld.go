package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

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
			testEqual(t, http.StatusUnprocessableEntity, res.StatusCode)

			t.Cleanup(func() { _ = res.Body.Close() })
		})
	}
}

func TestPostUsers_CreateUser(t *testing.T) {
	t.Parallel()

	req := UserPostRequestBody{
		Username: "createuser",
		Email:    "createuser@test.com",
		Password: "testpass",
	}
	res := httpPostUsers(t, req)
	testEqual(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response UserResponseBody
	testNil(t, json.NewDecoder(res.Body).Decode(&response))
	testEqual(t, req.Username, response.Username)
	testEqual(t, req.Email, response.Email)

	res = httpPostUsers(t, req) // return conflict when user already exists
	testEqual(t, http.StatusConflict, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPostUsers_ReturnsValidJWT(t *testing.T) {
	t.Parallel()

	// Given
	req := UserPostRequestBody{
		Username: "jwtuser",
		Email:    "jwtuser@test.com",
		Password: "testpass",
	}

	// When
	res := httpPostUsers(t, req)
	testEqual(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response UserResponseBody
	testNil(t, json.NewDecoder(res.Body).Decode(&response))

	// Then - token should not be the placeholder
	if response.Token == "token" {
		t.Fatal("expected real JWT token, got placeholder")
	}

	// Then - token should be parseable as valid JWT
	claims, err := auth.ParseToken(response.Token, "test-secret") // TODO: use actual secret
	testNil(t, err)

	// Then - JWT should contain correct user data
	testEqual(t, req.Username, claims.Username)
	// UserID should be > 0 (actual DB ID, not placeholder)
	if claims.UserID <= 0 {
		t.Errorf("expected positive userID, got %d", claims.UserID)
	}
}

func httpPostUsers(t *testing.T, request UserPostRequestBody) *http.Response {
	t.Helper()

	body, err := json.Marshal(UserWrapper[UserPostRequestBody]{User: request})
	testNil(t, err)

	res, err := http.Post(endpoint+"/api/users", "application/json", bytes.NewBuffer(body))
	testNil(t, err)

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
