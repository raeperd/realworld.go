package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/raeperd/realworld.go/internal/auth"
	"github.com/raeperd/realworld.go/internal/sqlite"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const userIDKey contextKey = "userID"

func handlePostUsers(db *sql.DB, jwtSecret string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request userPostRequestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		if errs := request.Validate(); len(errs) > 0 {
			encodeErrorResponse(r.Context(), http.StatusUnprocessableEntity, errs, w)
			return
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)
		user, err := queries.GetUserByEmail(r.Context(), request.User.Email)
		if err == nil {
			// User found, return conflict
			encodeErrorResponse(r.Context(), http.StatusConflict, []error{fmt.Errorf("user with email %s already exists", request.User.Email)}, w)
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			// Database error (like "no such table"), return internal server error
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		// User not found (sql.ErrNoRows), proceed to create

		user, err = queries.CreateUser(r.Context(), sqlite.CreateUserParams{
			Username: request.User.Username,
			Email:    request.User.Email,
			Password: request.User.Password,
		})
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Generate JWT token
		token, err := auth.GenerateToken(user.ID, user.Username, jwtSecret)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusCreated, userPostResponseBody{
			Email:    user.Email,
			Token:    token,
			Username: user.Username,
			Bio:      user.Bio.String,
			Image:    user.Image.String,
		}, w)
	}
}

func handlePostUsersLogin(db *sql.DB, jwtSecret string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var request userLoginRequestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		if errs := request.Validate(); len(errs) > 0 {
			encodeErrorResponse(r.Context(), http.StatusUnprocessableEntity, errs, w)
			return
		}

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)
		user, err := queries.GetUserByEmail(r.Context(), request.User.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// User not found - return 401 with generic message
				encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("invalid credentials")}, w)
				return
			}
			// Database error
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Verify password (plain text comparison for now)
		if user.Password != request.User.Password {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("invalid credentials")}, w)
			return
		}

		// Generate JWT token
		token, err := auth.GenerateToken(user.ID, user.Username, jwtSecret)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusOK, userPostResponseBody{
			Email:    user.Email,
			Token:    token,
			Username: user.Username,
			Bio:      user.Bio.String,
			Image:    user.Image.String,
		}, w)
	}
}

func encodeErrorResponse(ctx context.Context, status int, errs []error, w http.ResponseWriter) {
	errResp := errorResponseBody{
		Errors: struct {
			Body []string `json:"body"`
		}{
			Body: make([]string, len(errs)),
		},
	}
	for i, err := range errs {
		errResp.Errors.Body[i] = err.Error()
	}

	encodeResponse(ctx, status, errResp, w)
}

func encodeResponse[T responseBody](ctx context.Context, status int, body T, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.ErrorContext(ctx, "failed to encode response body", slog.String("error", err.Error()))
	}
}

type responseBody interface {
	userPostResponseBody | errorResponseBody
}

type userPostRequestBody struct {
	User struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

func (u userPostRequestBody) Validate() []error {
	var errs []error
	if u.User.Email == "" {
		errs = append(errs, errors.New("email is required"))
	}
	if u.User.Password == "" {
		errs = append(errs, errors.New("password is required"))
	}
	if u.User.Username == "" {
		errs = append(errs, errors.New("username is required"))
	}
	return errs
}

type userPostResponseBody struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

type errorResponseBody struct {
	Errors struct {
		Body []string `json:"body"`
	} `json:"errors"`
}

type userLoginRequestBody struct {
	User struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	} `json:"user"`
}

func (u userLoginRequestBody) Validate() []error {
	var errs []error
	if u.User.Email == "" {
		errs = append(errs, errors.New("email is required"))
	}
	if u.User.Password == "" {
		errs = append(errs, errors.New("password is required"))
	}
	return errs
}

func handleGetUser(db *sql.DB, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context (set by authenticate middleware)
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
		user, err := queries.GetUserByID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("user not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Generate fresh JWT token for response
		token, err := auth.GenerateToken(user.ID, user.Username, jwtSecret)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusOK, userPostResponseBody{
			Email:    user.Email,
			Token:    token,
			Username: user.Username,
			Bio:      user.Bio.String,
			Image:    user.Image.String,
		}, w)
	}
}

func handlePutUser(db *sql.DB, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from context (set by authenticate middleware)
		userID, ok := r.Context().Value(userIDKey).(int64)
		if !ok {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("unauthorized")}, w)
			return
		}

		var request userPutRequestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)
		user, err := queries.UpdateUser(r.Context(), sqlite.UpdateUserParams{
			ID:       userID,
			Email:    sql.NullString{String: request.User.Email, Valid: request.User.Email != ""},
			Username: sql.NullString{String: request.User.Username, Valid: request.User.Username != ""},
			Password: sql.NullString{String: request.User.Password, Valid: request.User.Password != ""},
			Bio:      sql.NullString{String: request.User.Bio, Valid: request.User.Bio != ""},
			Image:    sql.NullString{String: request.User.Image, Valid: request.User.Image != ""},
		})
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Generate fresh JWT token for response
		token, err := auth.GenerateToken(user.ID, user.Username, jwtSecret)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusOK, userPostResponseBody{
			Email:    user.Email,
			Token:    token,
			Username: user.Username,
			Bio:      user.Bio.String,
			Image:    user.Image.String,
		}, w)
	}
}

type userPutRequestBody struct {
	User struct {
		Email    string `json:"email,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Bio      string `json:"bio,omitempty"`
		Image    string `json:"image,omitempty"`
	} `json:"user"`
}
