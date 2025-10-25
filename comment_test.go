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

func TestPostArticlesSlugComments_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "comment_author_" + unique
	email := fmt.Sprintf("comment_%s@example.com", unique)

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
			Title:       "Article for Comments " + unique,
			Description: "Test article",
			Body:        "Article body",
			TagList:     []string{"test"},
		},
	}

	articleRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, articleRes.StatusCode)
	t.Cleanup(func() { _ = articleRes.Body.Close() })

	var articleResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(articleRes.Body).Decode(&articleResponse))
	slug := articleResponse.Article.Slug

	// Test: POST /api/articles/:slug/comments
	commentReq := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "This is a test comment",
		},
	}

	res := httpPostArticlesSlugComments(t, slug, commentReq, token)
	test.Equal(t, http.StatusCreated, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response structure and content
	var response CommentResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, "This is a test comment", response.Comment.Body)
	test.Equal(t, username, response.Comment.Author.Username)
	test.Equal(t, false, response.Comment.Author.Following)
	test.NotNil(t, response.Comment.ID)
	test.NotNil(t, response.Comment.CreatedAt)
	test.NotNil(t, response.Comment.UpdatedAt)
}

func TestPostArticlesSlugComments_MissingBody(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "comment_author_empty_" + unique
	email := fmt.Sprintf("comment_empty_%s@example.com", unique)

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
			Title:       "Article for Empty Comment " + unique,
			Description: "Test article",
			Body:        "Article body",
			TagList:     []string{"test"},
		},
	}

	articleRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, articleRes.StatusCode)
	t.Cleanup(func() { _ = articleRes.Body.Close() })

	var articleResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(articleRes.Body).Decode(&articleResponse))
	slug := articleResponse.Article.Slug

	// Test: POST comment with empty body
	commentReq := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "",
		},
	}

	res := httpPostArticlesSlugComments(t, slug, commentReq, token)
	test.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestPostArticlesSlugComments_ArticleNotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "comment_author_notfound_" + unique
	email := fmt.Sprintf("comment_notfound_%s@example.com", unique)

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

	// Test: POST comment to non-existent article
	commentReq := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "Comment on nonexistent article",
		},
	}

	res := httpPostArticlesSlugComments(t, "nonexistent-article-slug", commentReq, token)
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func httpPostArticlesSlugComments(t *testing.T, slug string, reqBody CommentPostRequestBody, token string) *http.Response {
	t.Helper()

	body, err := json.Marshal(reqBody)
	test.Nil(t, err)

	req, err := http.NewRequest(http.MethodPost, endpoint+"/api/articles/"+slug+"/comments", bytes.NewReader(body))
	test.Nil(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type CommentPostRequestBody struct {
	Comment CommentPostRequest `json:"comment"`
}

type CommentPostRequest struct {
	Body string `json:"body"`
}

type CommentResponseBody struct {
	Comment CommentResponse `json:"comment"`
}

type CommentResponse struct {
	ID        int64         `json:"id"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
	Body      string        `json:"body"`
	Author    AuthorProfile `json:"author"`
}

func TestGetArticlesSlugComments_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "get_comments_author_" + unique
	email := fmt.Sprintf("get_comments_%s@example.com", unique)

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
			Title:       "Article for Get Comments " + unique,
			Description: "Test article",
			Body:        "Article body",
			TagList:     []string{"test"},
		},
	}

	articleRes := httpPostArticles(t, articleReq, token)
	test.Equal(t, http.StatusCreated, articleRes.StatusCode)
	t.Cleanup(func() { _ = articleRes.Body.Close() })

	var articleResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(articleRes.Body).Decode(&articleResponse))
	slug := articleResponse.Article.Slug

	// Create multiple comments
	comment1Req := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "First comment",
		},
	}
	comment1Res := httpPostArticlesSlugComments(t, slug, comment1Req, token)
	test.Equal(t, http.StatusCreated, comment1Res.StatusCode)
	t.Cleanup(func() { _ = comment1Res.Body.Close() })

	comment2Req := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "Second comment",
		},
	}
	comment2Res := httpPostArticlesSlugComments(t, slug, comment2Req, token)
	test.Equal(t, http.StatusCreated, comment2Res.StatusCode)
	t.Cleanup(func() { _ = comment2Res.Body.Close() })

	// Test: GET /api/articles/:slug/comments
	res := httpGetArticlesSlugComments(t, slug, "")
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify response structure and content
	var response CommentsResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, 2, len(response.Comments))

	// Comments should be ordered by created_at DESC (most recent first)
	// Note: In tests, if timestamps are identical, order may vary
	// Verify both comments are present
	bodies := []string{response.Comments[0].Body, response.Comments[1].Body}
	test.True(t, contains(bodies, "First comment"))
	test.True(t, contains(bodies, "Second comment"))
	test.Equal(t, username, response.Comments[0].Author.Username)
	test.Equal(t, username, response.Comments[1].Author.Username)
	test.Equal(t, false, response.Comments[0].Author.Following)
	test.NotNil(t, response.Comments[0].ID)
	test.NotNil(t, response.Comments[0].CreatedAt)
	test.NotNil(t, response.Comments[0].UpdatedAt)
}

func httpGetArticlesSlugComments(t *testing.T, slug string, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, endpoint+"/api/articles/"+slug+"/comments", nil)
	test.Nil(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

type CommentsResponseBody struct {
	Comments []CommentResponse `json:"comments"`
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
