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

func TestGetProfilesUsername_NotFound(t *testing.T) {
	t.Parallel()

	// Test: GET /api/profiles/{username} for non-existent user
	nonExistentUsername := fmt.Sprintf("nonexistent_%d", time.Now().UnixNano())
	res := httpGetProfile(t, nonExistentUsername, "")
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestGetProfilesUsername_WithAuth(t *testing.T) {
	t.Parallel()

	// Setup: Create two users - viewer and target
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create viewer user
	viewerUsername := "viewer_" + unique
	viewerEmail := fmt.Sprintf("viewer_%s@example.com", unique)
	viewerReq := UserPostRequestBody{
		Username: viewerUsername,
		Email:    viewerEmail,
		Password: "testpass123",
	}
	viewerRes := httpPostUsers(t, viewerReq)
	test.Equal(t, http.StatusCreated, viewerRes.StatusCode)
	t.Cleanup(func() { _ = viewerRes.Body.Close() })

	var viewerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(viewerRes.Body).Decode(&viewerResponse))
	viewerToken := viewerResponse.Token

	// Create target user
	targetUsername := "target_" + unique
	targetEmail := fmt.Sprintf("target_%s@example.com", unique)
	targetReq := UserPostRequestBody{
		Username: targetUsername,
		Email:    targetEmail,
		Password: "testpass123",
	}
	targetRes := httpPostUsers(t, targetReq)
	test.Equal(t, http.StatusCreated, targetRes.StatusCode)
	t.Cleanup(func() { _ = targetRes.Body.Close() })

	// Test: GET /api/profiles/{username} with authentication
	res := httpGetProfile(t, targetUsername, viewerToken)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response
	var response ProfileResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, targetUsername, response.Profile.Username)
	test.Equal(t, false, response.Profile.Following) // Should be false until follow feature is implemented
}
