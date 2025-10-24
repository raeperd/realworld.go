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

//nolint:dupl // Test setup duplication is acceptable for clarity
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
	test.Equal(t, false, response.Profile.Following) // Should be false when not following
}

//nolint:dupl // Test setup duplication is acceptable for clarity
func TestGetProfilesUsername_WithAuthAfterFollow(t *testing.T) {
	t.Parallel()

	// Setup: Create two users - viewer and target
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

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

	// Follow the target user
	followRes := httpPostProfileFollow(t, targetUsername, viewerToken)
	test.Equal(t, http.StatusOK, followRes.StatusCode)
	t.Cleanup(func() { _ = followRes.Body.Close() })

	// Test: GET /api/profiles/{username} with authentication after following
	res := httpGetProfile(t, targetUsername, viewerToken)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify following status is true
	var response ProfileResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, targetUsername, response.Profile.Username)
	test.Equal(t, true, response.Profile.Following) // Should be true after following
}

//nolint:dupl // Test setup duplication is acceptable for clarity
func TestPostProfilesUsernameFollow_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create two users - follower and followed
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create follower user
	followerUsername := "follower_" + unique
	followerEmail := fmt.Sprintf("follower_%s@example.com", unique)
	followerReq := UserPostRequestBody{
		Username: followerUsername,
		Email:    followerEmail,
		Password: "testpass123",
	}
	followerRes := httpPostUsers(t, followerReq)
	test.Equal(t, http.StatusCreated, followerRes.StatusCode)
	t.Cleanup(func() { _ = followerRes.Body.Close() })

	var followerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(followerRes.Body).Decode(&followerResponse))
	followerToken := followerResponse.Token

	// Create followed user
	followedUsername := "followed_" + unique
	followedEmail := fmt.Sprintf("followed_%s@example.com", unique)
	followedReq := UserPostRequestBody{
		Username: followedUsername,
		Email:    followedEmail,
		Password: "testpass123",
	}
	followedRes := httpPostUsers(t, followedReq)
	test.Equal(t, http.StatusCreated, followedRes.StatusCode)
	t.Cleanup(func() { _ = followedRes.Body.Close() })

	// Test: POST /api/profiles/{username}/follow
	res := httpPostProfileFollow(t, followedUsername, followerToken)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response contains profile with following: true
	var response ProfileResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, followedUsername, response.Profile.Username)
	test.Equal(t, true, response.Profile.Following)
}

func httpPostProfileFollow(t *testing.T, username, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, endpoint+"/api/profiles/"+username+"/follow", nil)
	test.Nil(t, err)

	req.Header.Set("Authorization", "Token "+token)

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

func TestPostProfilesUsernameFollow_NotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create follower user
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	followerUsername := "follower_" + unique
	followerEmail := fmt.Sprintf("follower_%s@example.com", unique)
	followerReq := UserPostRequestBody{
		Username: followerUsername,
		Email:    followerEmail,
		Password: "testpass123",
	}
	followerRes := httpPostUsers(t, followerReq)
	test.Equal(t, http.StatusCreated, followerRes.StatusCode)
	t.Cleanup(func() { _ = followerRes.Body.Close() })

	var followerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(followerRes.Body).Decode(&followerResponse))
	followerToken := followerResponse.Token

	// Test: Attempt to follow non-existent user
	nonExistentUsername := fmt.Sprintf("nonexistent_%d", time.Now().UnixNano())
	res := httpPostProfileFollow(t, nonExistentUsername, followerToken)
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPostProfilesUsernameFollow_FollowSelf(t *testing.T) {
	t.Parallel()

	// Setup: Create user
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "user_" + unique
	email := fmt.Sprintf("user_%s@example.com", unique)
	userReq := UserPostRequestBody{
		Username: username,
		Email:    email,
		Password: "testpass123",
	}
	userRes := httpPostUsers(t, userReq)
	test.Equal(t, http.StatusCreated, userRes.StatusCode)
	t.Cleanup(func() { _ = userRes.Body.Close() })

	var userResponse UserResponseBody
	test.Nil(t, json.NewDecoder(userRes.Body).Decode(&userResponse))
	userToken := userResponse.Token

	// Test: Attempt to follow self
	res := httpPostProfileFollow(t, username, userToken)
	test.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

//nolint:dupl // Test setup duplication is acceptable for clarity
func TestPostProfilesUsernameFollow_AlreadyFollowing(t *testing.T) {
	t.Parallel()

	// Setup: Create two users
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

	followerUsername := "follower_" + unique
	followerEmail := fmt.Sprintf("follower_%s@example.com", unique)
	followerReq := UserPostRequestBody{
		Username: followerUsername,
		Email:    followerEmail,
		Password: "testpass123",
	}
	followerRes := httpPostUsers(t, followerReq)
	test.Equal(t, http.StatusCreated, followerRes.StatusCode)
	t.Cleanup(func() { _ = followerRes.Body.Close() })

	var followerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(followerRes.Body).Decode(&followerResponse))
	followerToken := followerResponse.Token

	followedUsername := "followed_" + unique
	followedEmail := fmt.Sprintf("followed_%s@example.com", unique)
	followedReq := UserPostRequestBody{
		Username: followedUsername,
		Email:    followedEmail,
		Password: "testpass123",
	}
	followedRes := httpPostUsers(t, followedReq)
	test.Equal(t, http.StatusCreated, followedRes.StatusCode)
	t.Cleanup(func() { _ = followedRes.Body.Close() })

	// First follow
	firstRes := httpPostProfileFollow(t, followedUsername, followerToken)
	test.Equal(t, http.StatusOK, firstRes.StatusCode)
	t.Cleanup(func() { _ = firstRes.Body.Close() })

	// Second follow (should be idempotent)
	secondRes := httpPostProfileFollow(t, followedUsername, followerToken)
	test.Equal(t, http.StatusOK, secondRes.StatusCode)
	t.Cleanup(func() { _ = secondRes.Body.Close() })

	var response ProfileResponseBody
	test.Nil(t, json.NewDecoder(secondRes.Body).Decode(&response))
	test.Equal(t, followedUsername, response.Profile.Username)
	test.Equal(t, true, response.Profile.Following)
}

func TestPostProfilesUsernameFollow_Unauthorized(t *testing.T) {
	t.Parallel()

	// Setup: Create a user to follow
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "user_" + unique
	email := fmt.Sprintf("user_%s@example.com", unique)
	userReq := UserPostRequestBody{
		Username: username,
		Email:    email,
		Password: "testpass123",
	}
	userRes := httpPostUsers(t, userReq)
	test.Equal(t, http.StatusCreated, userRes.StatusCode)
	t.Cleanup(func() { _ = userRes.Body.Close() })

	// Test: Attempt to follow without authorization
	res := httpPostProfileFollow(t, username, "")
	test.Equal(t, http.StatusUnauthorized, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

//nolint:dupl // Test setup duplication is acceptable for clarity
func TestDeleteProfilesUsernameFollow_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create two users - follower and followed
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create follower user
	followerUsername := "follower_" + unique
	followerEmail := fmt.Sprintf("follower_%s@example.com", unique)
	followerReq := UserPostRequestBody{
		Username: followerUsername,
		Email:    followerEmail,
		Password: "testpass123",
	}
	followerRes := httpPostUsers(t, followerReq)
	test.Equal(t, http.StatusCreated, followerRes.StatusCode)
	t.Cleanup(func() { _ = followerRes.Body.Close() })

	var followerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(followerRes.Body).Decode(&followerResponse))
	followerToken := followerResponse.Token

	// Create followed user
	followedUsername := "followed_" + unique
	followedEmail := fmt.Sprintf("followed_%s@example.com", unique)
	followedReq := UserPostRequestBody{
		Username: followedUsername,
		Email:    followedEmail,
		Password: "testpass123",
	}
	followedRes := httpPostUsers(t, followedReq)
	test.Equal(t, http.StatusCreated, followedRes.StatusCode)
	t.Cleanup(func() { _ = followedRes.Body.Close() })

	// First follow the user
	followRes := httpPostProfileFollow(t, followedUsername, followerToken)
	test.Equal(t, http.StatusOK, followRes.StatusCode)
	t.Cleanup(func() { _ = followRes.Body.Close() })

	// Test: DELETE /api/profiles/{username}/follow
	res := httpDeleteProfileFollow(t, followedUsername, followerToken)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response contains profile with following: false
	var response ProfileResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, followedUsername, response.Profile.Username)
	test.Equal(t, false, response.Profile.Following)
}

func httpDeleteProfileFollow(t *testing.T, username, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, endpoint+"/api/profiles/"+username+"/follow", nil)
	test.Nil(t, err)

	req.Header.Set("Authorization", "Token "+token)

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}
