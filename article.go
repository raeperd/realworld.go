package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

func handleGetArticles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		queryParams := r.URL.Query()
		limit := int64(20) // default
		offset := int64(0) // default
		tag := queryParams.Get("tag")
		author := queryParams.Get("author")
		favorited := queryParams.Get("favorited")

		if limitStr := queryParams.Get("limit"); limitStr != "" {
			if parsedLimit, err := parseInt64(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			if parsedOffset, err := parseInt64(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		queries := sqlite.New(db)

		// Build WHERE clause for filtering
		var articles []sqlite.ListArticlesRow
		var totalCount int64
		var err error

		// If filters are provided, use raw SQL (sqlc doesn't handle dynamic WHERE well)
		if tag != "" || author != "" || favorited != "" {
			articles, totalCount, err = listArticlesWithFilters(r.Context(), db, tag, author, favorited, limit, offset)
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
		} else {
			// Use sqlc-generated query for simple case
			articles, err = queries.ListArticles(r.Context(), sqlite.ListArticlesParams{
				Limit:  limit,
				Offset: offset,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}

			// Get total count
			totalCount, err = queries.CountArticles(r.Context())
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
		}

		// Check if user is authenticated
		userID, authenticated := r.Context().Value(userIDKey).(int64)

		// Build article IDs for batch queries
		articleIDs := make([]int64, len(articles))
		for i := range articles {
			articleIDs[i] = articles[i].ID
		}

		// Get favorites counts for all articles
		favoritesMap := make(map[int64]int64)
		if len(articleIDs) > 0 {
			favoritesCounts, err := queries.GetFavoritesByArticleIDs(r.Context(), articleIDs)
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, fc := range favoritesCounts {
				favoritesMap[fc.ArticleID] = fc.Count
			}
		}

		// Get favorited status if authenticated
		favoritedMap := make(map[int64]bool)
		if authenticated && len(articleIDs) > 0 {
			favoritedArticles, err := queries.CheckFavoritedByUser(r.Context(), sqlite.CheckFavoritedByUserParams{
				UserID:     userID,
				ArticleIds: articleIDs,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, articleID := range favoritedArticles {
				favoritedMap[articleID] = true
			}
		}

		// Get author IDs for following check
		authorIDs := make([]int64, len(articles))
		for i := range articles {
			authorIDs[i] = articles[i].AuthorID
		}

		// Get following status if authenticated
		followingMap := make(map[int64]bool)
		if authenticated && len(authorIDs) > 0 {
			followedAuthors, err := queries.GetFollowingByIDs(r.Context(), sqlite.GetFollowingByIDsParams{
				FollowerID:  userID,
				FollowedIds: authorIDs,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, followedID := range followedAuthors {
				followingMap[followedID] = true
			}
		}

		// Batch fetch tags for all articles
		tagsMap := make(map[int64][]string)
		if len(articleIDs) > 0 {
			articleTags, err := queries.GetArticleTagsByArticleIDs(r.Context(), articleIDs)
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, at := range articleTags {
				tagsMap[at.ArticleID] = append(tagsMap[at.ArticleID], at.Name)
			}
		}

		// Build response
		responseArticles := make([]articleListResponse, len(articles))
		for i := range articles {
			// Get tags for this article from map
			tags := tagsMap[articles[i].ID]

			// Ensure tags is never null
			if tags == nil {
				tags = []string{}
			}

			responseArticles[i] = articleListResponse{
				Slug:           articles[i].Slug,
				Title:          articles[i].Title,
				Description:    articles[i].Description,
				TagList:        tags,
				CreatedAt:      articles[i].CreatedAt.Format("2006-01-02T15:04:05.000Z"),
				UpdatedAt:      articles[i].UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				Favorited:      favoritedMap[articles[i].ID],
				FavoritesCount: favoritesMap[articles[i].ID],
				Author: authorProfile{
					Username:  articles[i].AuthorUsername,
					Bio:       articles[i].AuthorBio.String,
					Image:     articles[i].AuthorImage.String,
					Following: followingMap[articles[i].AuthorID],
				},
			}
		}

		encodeResponse(r.Context(), http.StatusOK, articlesResponseBody{
			Articles:      responseArticles,
			ArticlesCount: totalCount,
		}, w)
	}
}

func listArticlesWithFilters(ctx context.Context, db *sql.DB, tag, author, favorited string, limit, offset int64) ([]sqlite.ListArticlesRow, int64, error) {
	// Build base query
	query := `
		SELECT DISTINCT
			a.id,
			a.slug,
			a.title,
			a.description,
			a.created_at,
			a.updated_at,
			a.author_id,
			u.username as author_username,
			u.bio as author_bio,
			u.image as author_image
		FROM articles a
		JOIN users u ON a.author_id = u.id`

	countQuery := `
		SELECT COUNT(DISTINCT a.id)
		FROM articles a
		JOIN users u ON a.author_id = u.id`

	// Build WHERE clauses
	var whereClauses []string
	var args []any

	if tag != "" {
		query += `
		JOIN article_tags at ON a.id = at.article_id
		JOIN tags t ON at.tag_id = t.id`
		countQuery += `
		JOIN article_tags at ON a.id = at.article_id
		JOIN tags t ON at.tag_id = t.id`
		whereClauses = append(whereClauses, "t.name = ?")
		args = append(args, tag)
	}

	if author != "" {
		whereClauses = append(whereClauses, "u.username = ?")
		args = append(args, author)
	}

	if favorited != "" {
		query += `
		JOIN favorites f ON a.id = f.article_id
		JOIN users fu ON f.user_id = fu.id`
		countQuery += `
		JOIN favorites f ON a.id = f.article_id
		JOIN users fu ON f.user_id = fu.id`
		whereClauses = append(whereClauses, "fu.username = ?")
		args = append(args, favorited)
	}

	if len(whereClauses) > 0 {
		whereClause := " WHERE " + strings.Join(whereClauses, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	query += `
		ORDER BY a.created_at DESC
		LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	// Get total count first
	var totalCount int64
	countArgs := args[:len(args)-2] // Remove limit and offset for count
	err := db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Query articles
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var articles []sqlite.ListArticlesRow
	for rows.Next() {
		var article sqlite.ListArticlesRow
		err := rows.Scan(
			&article.ID,
			&article.Slug,
			&article.Title,
			&article.Description,
			&article.CreatedAt,
			&article.UpdatedAt,
			&article.AuthorID,
			&article.AuthorUsername,
			&article.AuthorBio,
			&article.AuthorImage,
		)
		if err != nil {
			return nil, 0, err
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return articles, totalCount, nil
}

func parseInt64(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func handleGetArticlesFeed(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID from context (required for feed)
		userID, ok := r.Context().Value(userIDKey).(int64)
		if !ok {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("unauthorized")}, w)
			return
		}

		// Parse query parameters
		queryParams := r.URL.Query()
		limit := int64(20) // default
		offset := int64(0) // default

		if limitStr := queryParams.Get("limit"); limitStr != "" {
			if parsedLimit, err := parseInt64(limitStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			if parsedOffset, err := parseInt64(offsetStr); err == nil && parsedOffset >= 0 {
				offset = parsedOffset
			}
		}

		queries := sqlite.New(db)

		// Get articles from followed users
		articles, err := queries.ListArticlesFeed(r.Context(), sqlite.ListArticlesFeedParams{
			FollowerID: userID,
			Limit:      limit,
			Offset:     offset,
		})
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Get total count of articles in feed
		totalCount, err := queries.CountArticlesFeed(r.Context(), userID)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Build article IDs for batch queries
		articleIDs := make([]int64, len(articles))
		for i := range articles {
			articleIDs[i] = articles[i].ID
		}

		// Get favorites counts for all articles
		favoritesMap := make(map[int64]int64)
		if len(articleIDs) > 0 {
			favoritesCounts, err := queries.GetFavoritesByArticleIDs(r.Context(), articleIDs)
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, fc := range favoritesCounts {
				favoritesMap[fc.ArticleID] = fc.Count
			}
		}

		// Get favorited status
		favoritedMap := make(map[int64]bool)
		if len(articleIDs) > 0 {
			favoritedArticles, err := queries.CheckFavoritedByUser(r.Context(), sqlite.CheckFavoritedByUserParams{
				UserID:     userID,
				ArticleIds: articleIDs,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, articleID := range favoritedArticles {
				favoritedMap[articleID] = true
			}
		}

		// Batch fetch tags for all articles
		tagsMap := make(map[int64][]string)
		if len(articleIDs) > 0 {
			articleTags, err := queries.GetArticleTagsByArticleIDs(r.Context(), articleIDs)
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			for _, at := range articleTags {
				tagsMap[at.ArticleID] = append(tagsMap[at.ArticleID], at.Name)
			}
		}

		// Build response
		responseArticles := make([]articleListResponse, len(articles))
		for i := range articles {
			// Get tags for this article from map
			tags := tagsMap[articles[i].ID]

			// Ensure tags is never null
			if tags == nil {
				tags = []string{}
			}

			responseArticles[i] = articleListResponse{
				Slug:           articles[i].Slug,
				Title:          articles[i].Title,
				Description:    articles[i].Description,
				TagList:        tags,
				CreatedAt:      articles[i].CreatedAt.Format("2006-01-02T15:04:05.000Z"),
				UpdatedAt:      articles[i].UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				Favorited:      favoritedMap[articles[i].ID],
				FavoritesCount: favoritesMap[articles[i].ID],
				Author: authorProfile{
					Username:  articles[i].AuthorUsername,
					Bio:       articles[i].AuthorBio.String,
					Image:     articles[i].AuthorImage.String,
					Following: true, // Always true in feed (articles are from followed users)
				},
			}
		}

		encodeResponse(r.Context(), http.StatusOK, articlesResponseBody{
			Articles:      responseArticles,
			ArticlesCount: totalCount,
		}, w)
	}
}

type articlesResponseBody struct {
	Articles      []articleListResponse `json:"articles"`
	ArticlesCount int64                 `json:"articlesCount"`
}

type articleListResponse struct {
	Slug           string        `json:"slug"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	TagList        []string      `json:"tagList"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
	Favorited      bool          `json:"favorited"`
	FavoritesCount int64         `json:"favoritesCount"`
	Author         authorProfile `json:"author"`
}
