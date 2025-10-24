package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/raeperd/test"
)

func TestGetTags_EmptyList(t *testing.T) {
	t.Parallel()

	// Test: GET /api/tags without any tags in the database
	res := httpGetTags(t)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response has correct structure with empty tags array
	var response TagsResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.NotNil(t, response.Tags)
	test.Equal(t, 0, len(response.Tags))
}

func httpGetTags(t *testing.T) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, endpoint+"/api/tags", nil)
	test.Nil(t, err)

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type TagsResponseBody struct {
	Tags []string `json:"tags"`
}
