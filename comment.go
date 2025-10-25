package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/raeperd/realworld.go/internal/sqlite"
)

func handlePostArticlesSlugComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request commentPostRequestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		if errs := request.Validate(); len(errs) > 0 {
			encodeErrorResponse(r.Context(), http.StatusUnprocessableEntity, errs, w)
			return
		}

		// Get authenticated user ID from context
		userID, ok := r.Context().Value(userIDKey).(int64)
		if !ok {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("unauthorized")}, w)
			return
		}

		// Get article slug from URL path
		slug := r.PathValue("slug")

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)

		// Get article by slug to verify it exists
		article, err := queries.GetArticleBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Create comment
		comment, err := queries.CreateComment(r.Context(), sqlite.CreateCommentParams{
			Body:      request.Comment.Body,
			ArticleID: article.ID,
			AuthorID:  userID,
		})
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get comment with author details
		commentWithAuthor, err := queries.GetCommentWithAuthor(r.Context(), comment.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Check if current user is following the comment author
		following := false
		if userID != commentWithAuthor.AuthorID {
			isFollowingInt, err := queries.IsFollowing(r.Context(), sqlite.IsFollowingParams{
				FollowerID: userID,
				FollowedID: commentWithAuthor.AuthorID,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			following = isFollowingInt == 1
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Build response
		response := commentResponseBody{
			Comment: commentPayload{
				ID:        commentWithAuthor.ID,
				CreatedAt: commentWithAuthor.CreatedAt.Format("2006-01-02T15:04:05.999Z"),
				UpdatedAt: commentWithAuthor.UpdatedAt.Format("2006-01-02T15:04:05.999Z"),
				Body:      commentWithAuthor.Body,
				Author: authorProfile{
					Username:  commentWithAuthor.AuthorUsername,
					Bio:       commentWithAuthor.AuthorBio.String,
					Image:     commentWithAuthor.AuthorImage.String,
					Following: following,
				},
			},
		}

		encodeResponse(r.Context(), http.StatusCreated, response, w)
	}
}

type commentPostRequestBody struct {
	Comment commentPostRequest `json:"comment"`
}

func (r commentPostRequestBody) Validate() []error {
	var errs []error
	if r.Comment.Body == "" {
		errs = append(errs, errors.New("body is required"))
	}
	return errs
}

type commentPostRequest struct {
	Body string `json:"body"`
}

type commentResponseBody struct {
	Comment commentPayload `json:"comment"`
}

type commentPayload struct {
	ID        int64         `json:"id"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
	Body      string        `json:"body"`
	Author    authorProfile `json:"author"`
}
