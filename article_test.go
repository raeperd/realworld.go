package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/raeperd/test"
)

func TestPostArticles_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "article_author_" + unique
	email := fmt.Sprintf("article_%s@example.com", unique)

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
	token := userResponse.Token

	// Test: POST /api/articles with authentication
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "How to train your dragon",
			Description: "Ever wonder how?",
			Body:        "You have to believe",
			TagList:     []string{"dragons", "training"},
		},
	}

	res := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response structure and content
	var response ArticleResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, "how-to-train-your-dragon", response.Article.Slug)
	test.Equal(t, "How to train your dragon", response.Article.Title)
	test.Equal(t, "Ever wonder how?", response.Article.Description)
	test.Equal(t, "You have to believe", response.Article.Body)
	test.Equal(t, 2, len(response.Article.TagList))
	test.Equal(t, false, response.Article.Favorited)
	test.Equal(t, int64(0), response.Article.FavoritesCount)
	test.Equal(t, username, response.Article.Author.Username)
	test.Equal(t, false, response.Article.Author.Following)
	test.NotNil(t, response.Article.CreatedAt)
	test.NotNil(t, response.Article.UpdatedAt)
}

func TestPostArticles_WithoutTags(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "article_author_no_tags_" + unique
	email := fmt.Sprintf("article_no_tags_%s@example.com", unique)

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
	token := userResponse.Token

	// Test: Create article without tags
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Simple article without tags " + unique,
			Description: "Description here",
			Body:        "Body content",
			TagList:     []string{},
		},
	}

	res := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response ArticleResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, 0, len(response.Article.TagList))
}

func TestPostArticles_Validation(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "validation_user_" + unique
	email := fmt.Sprintf("validation_%s@example.com", unique)

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
	token := userResponse.Token

	testcases := map[string]ArticlePostRequestBody{
		"title required": {
			Article: ArticlePostRequest{
				Title:       "",
				Description: "Description",
				Body:        "Body content",
			},
		},
		"description required": {
			Article: ArticlePostRequest{
				Title:       "Title",
				Description: "",
				Body:        "Body content",
			},
		},
		"body required": {
			Article: ArticlePostRequest{
				Title:       "Title",
				Description: "Description",
				Body:        "",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := httpPostArticles(t, tc, token)
			test.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
			t.Cleanup(func() { _ = res.Body.Close() })
		})
	}
}

func TestPostArticles_Unauthorized(t *testing.T) {
	t.Parallel()

	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Test Article",
			Description: "Description",
			Body:        "Body content",
		},
	}

	// Test without token
	res := httpPostArticles(t, articleReq, "")
	test.Equal(t, http.StatusUnauthorized, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func httpPostArticles(t *testing.T, reqBody ArticlePostRequestBody, token string) *http.Response {
	t.Helper()

	body, err := json.Marshal(reqBody)
	test.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, endpoint+"/api/articles", bytes.NewReader(body))
	test.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type ArticlePostRequestBody struct {
	Article ArticlePostRequest `json:"article"`
}

type ArticlePostRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

type ArticleResponseBody struct {
	Article ArticleResponse `json:"article"`
}

type ArticleResponse struct {
	Slug           string        `json:"slug"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Body           string        `json:"body"`
	TagList        []string      `json:"tagList"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
	Favorited      bool          `json:"favorited"`
	FavoritesCount int64         `json:"favoritesCount"`
	Author         AuthorProfile `json:"author"`
}

type AuthorProfile struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}
