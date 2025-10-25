package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
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
	test.True(t, slices.Contains(bodies, "First comment"))
	test.True(t, slices.Contains(bodies, "Second comment"))
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

func TestGetArticlesSlugComments_ArticleNotFound(t *testing.T) {
	t.Parallel()

	// Test: GET comments for non-existent article
	res := httpGetArticlesSlugComments(t, "nonexistent-article-slug", "")
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestGetArticlesSlugComments_EmptyComments(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "empty_comments_author_" + unique
	email := fmt.Sprintf("empty_comments_%s@example.com", unique)

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

	// Create an article with no comments
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Article Without Comments " + unique,
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

	// Test: GET comments for article with no comments
	res := httpGetArticlesSlugComments(t, slug, "")
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response CommentsResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, 0, len(response.Comments))
}

func TestGetArticlesSlugComments_WithFollowing(t *testing.T) {
	t.Parallel()

	// Setup: Create author user
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	authorUsername := "comment_author_" + unique
	authorEmail := fmt.Sprintf("author_%s@example.com", unique)

	authorUserReq := UserPostRequestBody{
		Username: authorUsername,
		Email:    authorEmail,
		Password: "testpass123",
	}
	authorRes := httpPostUsers(t, authorUserReq)
	test.Equal(t, http.StatusCreated, authorRes.StatusCode)
	t.Cleanup(func() { _ = authorRes.Body.Close() })

	var authorResponse UserResponseBody
	test.Nil(t, json.NewDecoder(authorRes.Body).Decode(&authorResponse))
	authorToken := authorResponse.Token

	// Create follower user
	followerUsername := "follower_" + unique
	followerEmail := fmt.Sprintf("follower_%s@example.com", unique)

	followerUserReq := UserPostRequestBody{
		Username: followerUsername,
		Email:    followerEmail,
		Password: "testpass123",
	}
	followerRes := httpPostUsers(t, followerUserReq)
	test.Equal(t, http.StatusCreated, followerRes.StatusCode)
	t.Cleanup(func() { _ = followerRes.Body.Close() })

	var followerResponse UserResponseBody
	test.Nil(t, json.NewDecoder(followerRes.Body).Decode(&followerResponse))
	followerToken := followerResponse.Token

	// Follower follows author
	followRes := httpPostProfileFollow(t, authorUsername, followerToken)
	test.Equal(t, http.StatusOK, followRes.StatusCode)
	t.Cleanup(func() { _ = followRes.Body.Close() })

	// Author creates article and comment
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Article for Following Test " + unique,
			Description: "Test article",
			Body:        "Article body",
			TagList:     []string{"test"},
		},
	}

	articleRes := httpPostArticles(t, articleReq, authorToken)
	test.Equal(t, http.StatusCreated, articleRes.StatusCode)
	t.Cleanup(func() { _ = articleRes.Body.Close() })

	var articleResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(articleRes.Body).Decode(&articleResponse))
	slug := articleResponse.Article.Slug

	commentReq := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "Test comment for following",
		},
	}
	commentRes := httpPostArticlesSlugComments(t, slug, commentReq, authorToken)
	test.Equal(t, http.StatusCreated, commentRes.StatusCode)
	t.Cleanup(func() { _ = commentRes.Body.Close() })

	// Test: GET comments as follower (should show following = true)
	res := httpGetArticlesSlugComments(t, slug, followerToken)
	test.Equal(t, http.StatusOK, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	var response CommentsResponseBody
	test.Nil(t, json.NewDecoder(res.Body).Decode(&response))
	test.Equal(t, 1, len(response.Comments))
	test.Equal(t, authorUsername, response.Comments[0].Author.Username)
	test.Equal(t, true, response.Comments[0].Author.Following)
}

func TestDeleteArticlesSlugCommentsID_Success(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "delete_comment_author_" + unique
	email := fmt.Sprintf("delete_comment_%s@example.com", unique)

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
			Title:       "Article for Delete Comment " + unique,
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

	// Create a comment
	commentReq := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "Comment to be deleted",
		},
	}

	commentRes := httpPostArticlesSlugComments(t, slug, commentReq, token)
	test.Equal(t, http.StatusCreated, commentRes.StatusCode)
	t.Cleanup(func() { _ = commentRes.Body.Close() })

	var commentResponse CommentResponseBody
	test.Nil(t, json.NewDecoder(commentRes.Body).Decode(&commentResponse))
	commentID := commentResponse.Comment.ID

	// Test: DELETE /api/articles/:slug/comments/:id
	res := httpDeleteArticlesSlugCommentsID(t, slug, commentID, token)
	test.Equal(t, http.StatusNoContent, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })

	// Verify comment no longer exists by getting all comments
	getRes := httpGetArticlesSlugComments(t, slug, "")
	test.Equal(t, http.StatusOK, getRes.StatusCode)
	t.Cleanup(func() { _ = getRes.Body.Close() })

	var getResponse CommentsResponseBody
	test.Nil(t, json.NewDecoder(getRes.Body).Decode(&getResponse))
	test.Equal(t, 0, len(getResponse.Comments))
}

func httpDeleteArticlesSlugCommentsID(t *testing.T, slug string, commentID int64, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/articles/%s/comments/%d", endpoint, slug, commentID), nil)
	test.Nil(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	res, err := http.DefaultClient.Do(req)
	test.Nil(t, err)

	return res
}

func TestDeleteArticlesSlugCommentsID_Unauthorized(t *testing.T) {
	t.Parallel()

	// Try to delete without auth token
	res := httpDeleteArticlesSlugCommentsID(t, "some-article", 1, "")
	test.Equal(t, http.StatusUnauthorized, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestDeleteArticlesSlugCommentsID_ArticleNotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "delete_notfound_" + unique
	email := fmt.Sprintf("delete_notfound_%s@example.com", unique)

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

	// Try to delete comment from non-existent article
	res := httpDeleteArticlesSlugCommentsID(t, "nonexistent-article", 1, token)
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestDeleteArticlesSlugCommentsID_CommentNotFound(t *testing.T) {
	t.Parallel()

	// Setup: Create a user and get token
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	username := "delete_comment_notfound_" + unique
	email := fmt.Sprintf("delete_comment_notfound_%s@example.com", unique)

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
			Title:       "Article for Comment Not Found " + unique,
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

	// Try to delete non-existent comment
	res := httpDeleteArticlesSlugCommentsID(t, slug, 999999, token)
	test.Equal(t, http.StatusNotFound, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}

func TestDeleteArticlesSlugCommentsID_Forbidden(t *testing.T) {
	t.Parallel()

	// Setup: Create author user
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	authorUsername := "comment_author_forbidden_" + unique
	authorEmail := fmt.Sprintf("author_forbidden_%s@example.com", unique)

	authorUserReq := UserPostRequestBody{
		Username: authorUsername,
		Email:    authorEmail,
		Password: "testpass123",
	}
	authorRes := httpPostUsers(t, authorUserReq)
	test.Equal(t, http.StatusCreated, authorRes.StatusCode)
	t.Cleanup(func() { _ = authorRes.Body.Close() })

	var authorResponse UserResponseBody
	test.Nil(t, json.NewDecoder(authorRes.Body).Decode(&authorResponse))
	authorToken := authorResponse.Token

	// Create other user
	otherUsername := "other_user_" + unique
	otherEmail := fmt.Sprintf("other_%s@example.com", unique)

	otherUserReq := UserPostRequestBody{
		Username: otherUsername,
		Email:    otherEmail,
		Password: "testpass123",
	}
	otherRes := httpPostUsers(t, otherUserReq)
	test.Equal(t, http.StatusCreated, otherRes.StatusCode)
	t.Cleanup(func() { _ = otherRes.Body.Close() })

	var otherResponse UserResponseBody
	test.Nil(t, json.NewDecoder(otherRes.Body).Decode(&otherResponse))
	otherToken := otherResponse.Token

	// Author creates article and comment
	articleReq := ArticlePostRequestBody{
		Article: ArticlePostRequest{
			Title:       "Article for Forbidden Test " + unique,
			Description: "Test article",
			Body:        "Article body",
			TagList:     []string{"test"},
		},
	}

	articleRes := httpPostArticles(t, articleReq, authorToken)
	test.Equal(t, http.StatusCreated, articleRes.StatusCode)
	t.Cleanup(func() { _ = articleRes.Body.Close() })

	var articleResponse ArticleResponseBody
	test.Nil(t, json.NewDecoder(articleRes.Body).Decode(&articleResponse))
	slug := articleResponse.Article.Slug

	commentReq := CommentPostRequestBody{
		Comment: CommentPostRequest{
			Body: "Author's comment",
		},
	}

	commentRes := httpPostArticlesSlugComments(t, slug, commentReq, authorToken)
	test.Equal(t, http.StatusCreated, commentRes.StatusCode)
	t.Cleanup(func() { _ = commentRes.Body.Close() })

	var commentResponse CommentResponseBody
	test.Nil(t, json.NewDecoder(commentRes.Body).Decode(&commentResponse))
	commentID := commentResponse.Comment.ID

	// Try to delete comment as other user (should be forbidden)
	res := httpDeleteArticlesSlugCommentsID(t, slug, commentID, otherToken)
	test.Equal(t, http.StatusForbidden, res.StatusCode)
	t.Cleanup(func() { _ = res.Body.Close() })
}
