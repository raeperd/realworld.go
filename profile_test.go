package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/raeperd/test"
)

func TestGetProfilesUsername_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user via registration
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "profile_user_" + unique
	email := fmt.Sprintf("profile_%s@example.com", unique)

	regReq := UserPostRequestBody{
		Username: username,
		Email:    email,
		Password: "testpass123",
	}
	regRes := httpPostUsers(t, regReq)
	test.Equal(t, http.StatusCreated, regRes.StatusCode)
	t.Cleanup(func() { _ = regRes.Body.Close() })

	// Test: GET /api/profiles/{username} without authentication
	res := httpGetProfile(t, username, "")
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response contains correct profile data
	var response ProfileResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, username, response.Profile.Username)
	test.Equal(t, "", response.Profile.Bio)
	test.Equal(t, "", response.Profile.Image)
	test.Equal(t, false, response.Profile.Following)
}

func httpGetProfile(t *testing.T, username, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, endpoint+"/api/profiles/"+username, nil)
	test.Nil(t, err)

	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type ProfileResponseBody struct {
	Profile struct {
		Username  string `json:"username"`
		Bio       string `json:"bio"`
		Image     string `json:"image"`
		Following bool   `json:"following"`
	} `json:"profile"`
}
