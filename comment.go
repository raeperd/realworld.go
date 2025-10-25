package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

func handleGetArticlesSlugComments(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get article slug from URL path
		slug := r.PathValue("slug")

		// Get optional user ID from context (for following status)
		userID, _ := r.Context().Value(userIDKey).(int64)

		queries := sqlite.New(db)

		// Verify article exists first
		_, err := queries.GetArticleBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get comments by article slug
		comments, err := queries.GetCommentsByArticleSlug(r.Context(), slug)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Build following status map to avoid N+1 queries
		followingMap := make(map[int64]bool)
		if userID != 0 && len(comments) > 0 {
			// Collect unique author IDs (excluding current user)
			authorIDsMap := make(map[int64]struct{})
			for i := range comments {
				if comments[i].AuthorID != userID {
					authorIDsMap[comments[i].AuthorID] = struct{}{}
				}
			}

			// Convert map to slice for batch query
			if len(authorIDsMap) > 0 {
				authorIDs := make([]int64, 0, len(authorIDsMap))
				for id := range authorIDsMap {
					authorIDs = append(authorIDs, id)
				}

				// Single batch query to get all following relationships
				followedIDs, err := queries.GetFollowingByIDs(r.Context(), sqlite.GetFollowingByIDsParams{
					FollowerID:  userID,
					FollowedIds: authorIDs,
				})
				if err != nil {
					encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
					return
				}

				// Build map for O(1) lookups
				for _, followedID := range followedIDs {
					followingMap[followedID] = true
				}
			}
		}

		// Build response with comments
		commentPayloads := make([]commentPayload, len(comments))
		for i := range comments {
			// Get following status from map (defaults to false if not present)
			following := followingMap[comments[i].AuthorID]

			commentPayloads[i] = commentPayload{
				ID:        comments[i].ID,
				CreatedAt: comments[i].CreatedAt.Format("2006-01-02T15:04:05.999Z"),
				UpdatedAt: comments[i].UpdatedAt.Format("2006-01-02T15:04:05.999Z"),
				Body:      comments[i].Body,
				Author: authorProfile{
					Username:  comments[i].AuthorUsername,
					Bio:       comments[i].AuthorBio.String,
					Image:     comments[i].AuthorImage.String,
					Following: following,
				},
			}
		}

		response := commentsResponseBody{
			Comments: commentPayloads,
		}

		encodeResponse(r.Context(), http.StatusOK, response, w)
	}
}

type commentsResponseBody struct {
	Comments []commentPayload `json:"comments"`
}

func handleDeleteArticlesSlugCommentsID(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		commentIDStr := r.PathValue("id")

		// Parse comment ID from path
		var commentID int64
		if _, err := fmt.Sscanf(commentIDStr, "%d", &commentID); err != nil {
			encodeErrorResponse(r.Context(), http.StatusBadRequest, []error{errors.New("invalid comment ID")}, w)
			return
		}

		// Get authenticated user ID from context
		userID, ok := r.Context().Value(userIDKey).(int64)
		if !ok {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("unauthorized")}, w)
			return
		}

		queries := sqlite.New(db)

		// Verify article exists
		article, err := queries.GetArticleBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get comment by ID
		comment, err := queries.GetCommentByID(r.Context(), commentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("comment not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Verify comment belongs to the article (security check)
		if comment.ArticleID != article.ID {
			encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("comment not found")}, w)
			return
		}

		// Verify user is the comment author (authorization check)
		if comment.AuthorID != userID {
			encodeErrorResponse(r.Context(), http.StatusForbidden, []error{errors.New("not authorized to delete this comment")}, w)
			return
		}

		// Delete the comment
		err = queries.DeleteComment(r.Context(), commentID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Return 204 No Content
		w.WriteHeader(http.StatusNoContent)
	}
}
