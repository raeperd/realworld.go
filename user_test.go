package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
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
		Username: "test",
		Email:    "test@test.com",
		Password: "test",
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
