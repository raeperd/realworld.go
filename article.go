package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/raeperd/realworld.go/internal/sqlite"
)

func handlePostArticles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request articlePostRequestBody
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

		// Generate slug from title
		slug := generateSlug(request.Article.Title)

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)

		// Create article
		article, err := queries.CreateArticle(r.Context(), sqlite.CreateArticleParams{
			Slug:        slug,
			Title:       request.Article.Title,
			Description: request.Article.Description,
			Body:        request.Article.Body,
			AuthorID:    userID,
		})
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Handle tags if provided
		if len(request.Article.TagList) > 0 {
			for _, tagName := range request.Article.TagList {
				// Get or create tag (upsert)
				tag, err := queries.GetOrCreateTag(r.Context(), tagName)
				if err != nil {
					encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
					return
				}

				// Associate tag with article
				err = queries.AssociateArticleTag(r.Context(), sqlite.AssociateArticleTagParams{
					ArticleID: article.ID,
					TagID:     tag.ID,
				})
				if err != nil {
					encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
					return
				}
			}
		}

		// Get author details
		author, err := queries.GetUserByID(r.Context(), userID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get tags for response
		tags, err := queries.GetArticleTagsByArticleID(r.Context(), article.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Ensure tags is never null in JSON response
		if tags == nil {
			tags = []string{}
		}

		encodeResponse(r.Context(), http.StatusCreated, articleResponseBody{
			Article: articleResponse{
				Slug:           article.Slug,
				Title:          article.Title,
				Description:    article.Description,
				Body:           article.Body,
				TagList:        tags,
				CreatedAt:      article.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
				UpdatedAt:      article.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				Favorited:      false,
				FavoritesCount: 0,
				Author: authorProfile{
					Username:  author.Username,
					Bio:       author.Bio.String,
					Image:     author.Image.String,
					Following: false, // Author viewing their own article
				},
			},
		}, w)
	}
}

func generateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters, keep only alphanumeric and hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

type articlePostRequestBody struct {
	Article articlePostRequest `json:"article"`
}

type articlePostRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}

func (r articlePostRequestBody) Validate() []error {
	var errs []error
	if r.Article.Title == "" {
		errs = append(errs, errors.New("title is required"))
	}
	if r.Article.Description == "" {
		errs = append(errs, errors.New("description is required"))
	}
	if r.Article.Body == "" {
		errs = append(errs, errors.New("body is required"))
	}
	return errs
}

type articleResponseBody struct {
	Article articleResponse `json:"article"`
}

type articleResponse struct {
	Slug           string        `json:"slug"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	Body           string        `json:"body"`
	TagList        []string      `json:"tagList"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
	Favorited      bool          `json:"favorited"`
	FavoritesCount int64         `json:"favoritesCount"`
	Author         authorProfile `json:"author"`
}

type authorProfile struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

func handleGetArticlesSlug(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		queries := sqlite.New(db)

		// Get article by slug
		article, err := queries.GetArticleBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get tags for the article
		tags, err := queries.GetArticleTagsByArticleID(r.Context(), article.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Ensure tags is never null in JSON response
		if tags == nil {
			tags = []string{}
		}

		// Get favorites count
		favoritesCount, err := queries.GetFavoritesCount(r.Context(), article.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Check if user is authenticated
		userID, authenticated := r.Context().Value(userIDKey).(int64)

		// Check favorited status if authenticated
		favorited := false
		if authenticated {
			favoritedInt, err := queries.IsFavorited(r.Context(), sqlite.IsFavoritedParams{
				UserID:    userID,
				ArticleID: article.ID,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			favorited = favoritedInt > 0
		}

		// Check following status if authenticated
		following := false
		if authenticated {
			followingInt, err := queries.IsFollowing(r.Context(), sqlite.IsFollowingParams{
				FollowerID: userID,
				FollowedID: article.AuthorID,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			following = followingInt > 0
		}

		encodeResponse(r.Context(), http.StatusOK, articleResponseBody{
			Article: articleResponse{
				Slug:           article.Slug,
				Title:          article.Title,
				Description:    article.Description,
				Body:           article.Body,
				TagList:        tags,
				CreatedAt:      article.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
				UpdatedAt:      article.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				Favorited:      favorited,
				FavoritesCount: favoritesCount,
				Author: authorProfile{
					Username:  article.AuthorUsername,
					Bio:       article.AuthorBio.String,
					Image:     article.AuthorImage.String,
					Following: following,
				},
			},
		}, w)
	}
}

func handlePutArticlesSlug(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		var request articlePutRequestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		// Get authenticated user ID from context
		userID, ok := r.Context().Value(userIDKey).(int64)
		if !ok {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("unauthorized")}, w)
			return
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)

		// Get existing article by slug
		existingArticle, err := queries.GetArticleBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Check if user is the author
		if existingArticle.AuthorID != userID {
			encodeErrorResponse(r.Context(), http.StatusForbidden, []error{errors.New("not authorized to update this article")}, w)
			return
		}

		// Prepare update parameters
		updateParams := sqlite.UpdateArticleParams{
			ID: existingArticle.ID,
		}

		// Handle slug regeneration if title is updated
		if request.Article.Title != nil {
			newSlug := generateSlug(*request.Article.Title)
			updateParams.Slug = sql.NullString{String: newSlug, Valid: true}
			updateParams.Title = sql.NullString{String: *request.Article.Title, Valid: true}
		}

		if request.Article.Description != nil {
			updateParams.Description = sql.NullString{String: *request.Article.Description, Valid: true}
		}

		if request.Article.Body != nil {
			updateParams.Body = sql.NullString{String: *request.Article.Body, Valid: true}
		}

		// Update article
		article, err := queries.UpdateArticle(r.Context(), updateParams)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get author details
		author, err := queries.GetUserByID(r.Context(), userID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get tags for response
		tags, err := queries.GetArticleTagsByArticleID(r.Context(), article.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get favorites count
		favoritesCount, err := queries.GetFavoritesCount(r.Context(), article.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Ensure tags is never null in JSON response
		if tags == nil {
			tags = []string{}
		}

		encodeResponse(r.Context(), http.StatusOK, articleResponseBody{
			Article: articleResponse{
				Slug:           article.Slug,
				Title:          article.Title,
				Description:    article.Description,
				Body:           article.Body,
				TagList:        tags,
				CreatedAt:      article.CreatedAt.Format("2006-01-02T15:04:05.000Z"),
				UpdatedAt:      article.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				Favorited:      false,
				FavoritesCount: favoritesCount,
				Author: authorProfile{
					Username:  author.Username,
					Bio:       author.Bio.String,
					Image:     author.Image.String,
					Following: false, // Author viewing their own article
				},
			},
		}, w)
	}
}

type articlePutRequestBody struct {
	Article articlePutRequest `json:"article"`
}

type articlePutRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Body        *string `json:"body,omitempty"`
}

func handleDeleteArticlesSlug(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		// Get authenticated user ID from context
		userID, ok := r.Context().Value(userIDKey).(int64)
		if !ok {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("unauthorized")}, w)
			return
		}

		queries := sqlite.New(db)

		// Get existing article by slug
		article, err := queries.GetArticleBySlug(r.Context(), slug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("article not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Check if user is the author
		if article.AuthorID != userID {
			encodeErrorResponse(r.Context(), http.StatusForbidden, []error{errors.New("not authorized to delete this article")}, w)
			return
		}

		// Delete the article
		err = queries.DeleteArticle(r.Context(), article.ID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Return 200 OK with no content
		w.WriteHeader(http.StatusOK)
	}
}
