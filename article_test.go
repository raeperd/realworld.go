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

func TestGetArticlesSlug_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and article
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "article_reader_" + unique
	email := fmt.Sprintf("reader_%s@example.com", unique)

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

	// Create an article
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Test Article " + unique,
			Description: "Test description",
			Body:        "Test body content",
			TagList:     []string{"test", "golang"},
		},
	}

	createRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	slug := createResponse.Article.Slug

	// Test: GET /api/articles/:slug without authentication
	res := httpGetArticlesSlug(t, slug, "")
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response
	var response ArticleResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, slug, response.Article.Slug)
	test.Equal(t, "Test Article "+unique, response.Article.Title)
	test.Equal(t, "Test description", response.Article.Description)
	test.Equal(t, "Test body content", response.Article.Body)
	test.Equal(t, 2, len(response.Article.TagList))
	test.Equal(t, false, response.Article.Favorited)
	test.Equal(t, int64(0), response.Article.FavoritesCount)
	test.Equal(t, username, response.Article.Author.Username)
	test.Equal(t, false, response.Article.Author.Following)
	test.NotNil(t, response.Article.CreatedAt)
	test.NotNil(t, response.Article.UpdatedAt)
}

func TestGetArticlesSlug_NotFound(t *testing.T) {
	t.Parallel()

	// Test: GET /api/articles/:slug with non-existent slug
	res := httpGetArticlesSlug(t, "nonexistent-article-slug", "")
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestGetArticlesSlug_Authenticated(t *testing.T) {
	t.Parallel()

	// Setup: Create two users - one author and one reader
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	authorUsername := "author_" + unique
	authorEmail := fmt.Sprintf("author_%s@example.com", unique)

	authorReq := UserPostRequestBody{
		Username: authorUsername,
		Email:    authorEmail,
		Password: "testpass123",
	}
	authorRes := httpPostUsers(t, authorReq)
	test.Equal(t, http.StatusCreated, authorRes.StatusCode)
	t.Cleanup(func() { _ = authorRes.Body.Close() })

	var authorResponse UserResponseBody
	test.Nil(t, json.NewDecoder(authorRes.Body).Decode(&authorResponse))
	authorToken := authorResponse.Token

	// Create an article as the author
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Authenticated Test Article " + unique,
			Description: "Test description",
			Body:        "Test body content",
			TagList:     []string{"auth", "test"},
		},
	}

	createRes := httpPostArticles(t, articleReq, authorToken)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	slug := createResponse.Article.Slug

	// Create a reader user
	readerUsername := "reader_" + unique
	readerEmail := fmt.Sprintf("reader_%s@example.com", unique)

	readerReq := UserPostRequestBody{
		Username: readerUsername,
		Email:    readerEmail,
		Password: "testpass123",
	}
	readerRes := httpPostUsers(t, readerReq)
	test.Equal(t, http.StatusCreated, readerRes.StatusCode)
	t.Cleanup(func() { _ = readerRes.Body.Close() })

	var readerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(readerRes.Body).Decode(&readerResponse))
	readerToken := readerResponse.Token

	// Test: GET /api/articles/:slug as authenticated reader
	res := httpGetArticlesSlug(t, slug, readerToken)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response - authenticated user should get correct favorited/following status
	var response ArticleResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, slug, response.Article.Slug)
	test.Equal(t, "Authenticated Test Article "+unique, response.Article.Title)
	test.Equal(t, false, response.Article.Favorited)        // Reader hasn't favorited
	test.Equal(t, false, response.Article.Author.Following) // Reader doesn't follow author
	test.Equal(t, authorUsername, response.Article.Author.Username)
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

func httpGetArticlesSlug(t *testing.T, slug string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, endpoint+"/api/articles/"+slug, nil)
	test.Nil(t, err)
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

func TestPutArticlesSlug_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and article
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "article_updater_" + unique
	email := fmt.Sprintf("updater_%s@example.com", unique)

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

	// Create an article
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Original Title " + unique,
			Description: "Original description",
			Body:        "Original body content",
			TagList:     []string{"original"},
		},
	}

	createRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	originalSlug := createResponse.Article.Slug

	// Test: Update article with new title (should regenerate slug)
	updateReq := ArticlePutRequestBody{
		Article: ArticlePutRequest{
			Title:       stringPtr("Updated Title " + unique),
			Description: stringPtr("Updated description"),
			Body:        stringPtr("Updated body content"),
		},
	}

	res := httpPutArticlesSlug(t, originalSlug, updateReq, token)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response
	var response ArticleResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.NotEqual(t, originalSlug, response.Article.Slug) // Slug should change
	test.Equal(t, "Updated Title "+unique, response.Article.Title)
	test.Equal(t, "Updated description", response.Article.Description)
	test.Equal(t, "Updated body content", response.Article.Body)
	test.Equal(t, username, response.Article.Author.Username)
	test.NotNil(t, response.Article.UpdatedAt)
}

func httpPutArticlesSlug(t *testing.T, slug string, reqBody ArticlePutRequestBody, token string) *http.Response {
	t.Helper()

	body, err := json.Marshal(reqBody)
	test.Nil(t, err)

	req, err := http.NewRequest(http.MethodPut, endpoint+"/api/articles/"+slug, bytes.NewReader(body))
	test.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type ArticlePutRequestBody struct {
	Article ArticlePutRequest `json:"article"`
}

type ArticlePutRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Body        *string `json:"body,omitempty"`
}

func stringPtr(s string) *string {
	return &s
}

func TestPutArticlesSlug_NotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create a user
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
	token := userResponse.Token

	// Test: Try to update non-existent article
	updateReq := ArticlePutRequestBody{
		Article: ArticlePutRequest{
			Title: stringPtr("New Title"),
		},
	}

	res := httpPutArticlesSlug(t, "nonexistent-slug", updateReq, token)
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPutArticlesSlug_Forbidden(t *testing.T) {
	t.Parallel()

	// Setup: Create two users
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

	// User 1 - article author
	author := "author_" + unique
	authorEmail := fmt.Sprintf("author_%s@example.com", unique)

	authorReq := UserPostRequestBody{
		Username: author,
		Email:    authorEmail,
		Password: "testpass123",
	}
	authorRes := httpPostUsers(t, authorReq)
	test.Equal(t, http.StatusCreated, authorRes.StatusCode)
	t.Cleanup(func() { _ = authorRes.Body.Close() })

	var authorResponse UserResponseBody
	test.Nil(t, json.NewDecoder(authorRes.Body).Decode(&authorResponse))
	authorToken := authorResponse.Token

	// Create an article as author
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Author's Article " + unique,
			Description: "Description",
			Body:        "Body content",
			TagList:     []string{},
		},
	}

	createRes := httpPostArticles(t, articleReq, authorToken)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	slug := createResponse.Article.Slug

	// User 2 - different user
	otherUser := "other_" + unique
	otherEmail := fmt.Sprintf("other_%s@example.com", unique)

	otherReq := UserPostRequestBody{
		Username: otherUser,
		Email:    otherEmail,
		Password: "testpass123",
	}
	otherRes := httpPostUsers(t, otherReq)
	test.Equal(t, http.StatusCreated, otherRes.StatusCode)
	t.Cleanup(func() { _ = otherRes.Body.Close() })

	var otherResponse UserResponseBody
	test.Nil(t, json.NewDecoder(otherRes.Body).Decode(&otherResponse))
	otherToken := otherResponse.Token

	// Test: Try to update article as different user
	updateReq := ArticlePutRequestBody{
		Article: ArticlePutRequest{
			Title: stringPtr("Hacked Title"),
		},
	}

	res := httpPutArticlesSlug(t, slug, updateReq, otherToken)
	test.Equal(t, http.StatusForbidden, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPutArticlesSlug_PartialUpdate(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and article
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "partial_user_" + unique
	email := fmt.Sprintf("partial_%s@example.com", unique)

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

	// Create an article
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Original Title " + unique,
			Description: "Original description",
			Body:        "Original body content",
			TagList:     []string{"original"},
		},
	}

	createRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	slug := createResponse.Article.Slug

	// Test: Update only body (partial update)
	updateReq := ArticlePutRequestBody{
		Article: ArticlePutRequest{
			Body: stringPtr("Updated body only"),
		},
	}

	res := httpPutArticlesSlug(t, slug, updateReq, token)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response - slug and title should remain unchanged
	var response ArticleResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, slug, response.Article.Slug)                          // Slug unchanged
	test.Equal(t, "Original Title "+unique, response.Article.Title)     // Title unchanged
	test.Equal(t, "Original description", response.Article.Description) // Description unchanged
	test.Equal(t, "Updated body only", response.Article.Body)           // Body updated
}

func TestDeleteArticlesSlug_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and article
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "article_deleter_" + unique
	email := fmt.Sprintf("deleter_%s@example.com", unique)

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

	// Create an article
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Article to Delete " + unique,
			Description: "Will be deleted",
			Body:        "Body content",
			TagList:     []string{"delete", "test"},
		},
	}

	createRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	slug := createResponse.Article.Slug

	// Test: Delete the article
	res := httpDeleteArticlesSlug(t, slug, token)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify article is deleted by trying to get it
	getRes := httpGetArticlesSlug(t, slug, "")
	test.Equal(t, http.StatusNotFound, getRes.StatusCode)
	t.Cleanup(func() { _ = getRes.Body.Close() })
}

func httpDeleteArticlesSlug(t *testing.T, slug string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, endpoint+"/api/articles/"+slug, nil)
	test.Nil(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

func TestDeleteArticlesSlug_NotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create a user
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
	token := userResponse.Token

	// Test: Try to delete non-existent article
	res := httpDeleteArticlesSlug(t, "nonexistent-slug", token)
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestDeleteArticlesSlug_Forbidden(t *testing.T) {
	t.Parallel()

	// Setup: Create two users
	unique := fmt.Sprintf("%d", time.Now().UnixNano())

	// User 1 - article author
	author := "author_" + unique
	authorEmail := fmt.Sprintf("author_%s@example.com", unique)

	authorReq := UserPostRequestBody{
		Username: author,
		Email:    authorEmail,
		Password: "testpass123",
	}
	authorRes := httpPostUsers(t, authorReq)
	test.Equal(t, http.StatusCreated, authorRes.StatusCode)
	t.Cleanup(func() { _ = authorRes.Body.Close() })

	var authorResponse UserResponseBody
	test.Nil(t, json.NewDecoder(authorRes.Body).Decode(&authorResponse))
	authorToken := authorResponse.Token

	// Create an article as author
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Author's Article " + unique,
			Description: "Description",
			Body:        "Body content",
			TagList:     []string{},
		},
	}

	createRes := httpPostArticles(t, articleReq, authorToken)
	test.Equal(t, http.StatusCreated, createRes.StatusCode)
	t.Cleanup(func() { _ = createRes.Body.Close() })

	var createResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(createRes.Body).Decode(&createResponse))
	slug := createResponse.Article.Slug

	// User 2 - different user
	otherUser := "other_" + unique
	otherEmail := fmt.Sprintf("other_%s@example.com", unique)

	otherReq := UserPostRequestBody{
		Username: otherUser,
		Email:    otherEmail,
		Password: "testpass123",
	}
	otherRes := httpPostUsers(t, otherReq)
	test.Equal(t, http.StatusCreated, otherRes.StatusCode)
	t.Cleanup(func() { _ = otherRes.Body.Close() })

	var otherResponse UserResponseBody
	test.Nil(t, json.NewDecoder(otherRes.Body).Decode(&otherResponse))
	otherToken := otherResponse.Token

	// Test: Try to delete article as different user
	res := httpDeleteArticlesSlug(t, slug, otherToken)
	test.Equal(t, http.StatusForbidden, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestGetArticles_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and multiple articles
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "article_lister_" + unique
	email := fmt.Sprintf("lister_%s@example.com", unique)

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

	// Create first article
	article1Req := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "First Article " + unique,
			Description: "First description",
			Body:        "First body",
			TagList:     []string{"golang", "test"},
		},
	}
	article1Res := httpPostArticles(t, article1Req, token)
	test.Equal(t, http.StatusCreated, article1Res.StatusCode)
	t.Cleanup(func() { _ = article1Res.Body.Close() })

	// Small delay to ensure different created_at timestamps
	time.Sleep(10 * time.Millisecond)

	// Create second article
	article2Req := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Second Article " + unique,
			Description: "Second description",
			Body:        "Second body",
			TagList:     []string{"rust", "tutorial"},
		},
	}
	article2Res := httpPostArticles(t, article2Req, token)
	test.Equal(t, http.StatusCreated, article2Res.StatusCode)
	t.Cleanup(func() { _ = article2Res.Body.Close() })

	// Test: GET /api/articles without authentication
	res := httpGetArticles(t, "", "")
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response structure
	var response ArticlesResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))

	// Should have at least 2 articles (our created ones)
	test.True(t, len(response.Articles) >= 2)
	test.True(t, response.ArticlesCount >= 2)

	// Find our articles in the response (they should be the most recent)
	var firstArticle, secondArticle *ArticleListResponse
	for i := range response.Articles {
		if response.Articles[i].Title == "First Article "+unique {
			firstArticle = &response.Articles[i]
		}
		if response.Articles[i].Title == "Second Article "+unique {
			secondArticle = &response.Articles[i]
		}
	}

	test.NotNil(t, firstArticle)
	test.NotNil(t, secondArticle)

	// Verify first article structure (should NOT include body field)
	test.Equal(t, "first-article-"+unique, firstArticle.Slug)
	test.Equal(t, "First Article "+unique, firstArticle.Title)
	test.Equal(t, "First description", firstArticle.Description)
	test.Equal(t, 2, len(firstArticle.TagList))
	test.Equal(t, false, firstArticle.Favorited)
	test.Equal(t, int64(0), firstArticle.FavoritesCount)
	test.Equal(t, username, firstArticle.Author.Username)
	test.NotNil(t, firstArticle.CreatedAt)
	test.NotNil(t, firstArticle.UpdatedAt)

	// Verify second article
	test.Equal(t, "second-article-"+unique, secondArticle.Slug)
	test.Equal(t, "Second Article "+unique, secondArticle.Title)
	test.Equal(t, "Second description", secondArticle.Description)
	test.Equal(t, 2, len(secondArticle.TagList))
}

func httpGetArticles(t *testing.T, queryParams string, token string) *http.Response {
	t.Helper()

	url := endpoint + "/api/articles"
	if queryParams != "" {
		url += "?" + queryParams
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	test.Nil(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type ArticlesResponseBody struct {
	Articles      []ArticleListResponse `json:"articles"`
	ArticlesCount int64                 `json:"articlesCount"`
}

type ArticleListResponse struct {
	Slug           string        `json:"slug"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	TagList        []string      `json:"tagList"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
	Favorited      bool          `json:"favorited"`
	FavoritesCount int64         `json:"favoritesCount"`
	Author         AuthorProfile `json:"author"`
	// Note: Body field is intentionally omitted (per spec update 2024/08/16)
}
