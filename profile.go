package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/raeperd/realworld.go/internal/sqlite"
)

func handleGetProfilesUsername(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")

		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)
		user, err := queries.GetUserByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("profile not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Check if authenticated user is following this profile
		following := false
		if viewerID, ok := r.Context().Value(userIDKey).(int64); ok {
			isFollowing, err := queries.IsFollowing(r.Context(), sqlite.IsFollowingParams{
				FollowerID: viewerID,
				FollowedID: user.ID,
			})
			if err != nil {
				encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
				return
			}
			following = isFollowing > 0
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusOK, profileGetResponseWrapper{
			Profile: profileGetResponseBody{
				Username:  user.Username,
				Bio:       user.Bio.String,
				Image:     user.Image.String,
				Following: following,
			},
		}, w)
	}
}

type profileGetResponseWrapper struct {
	Profile profileGetResponseBody `json:"profile"`
}

type profileGetResponseBody struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

func handlePostProfilesUsernameFollow(db *sql.DB, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.PathValue("username")
		followerID, ok := r.Context().Value(userIDKey).(int64)
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

		// Get user to follow
		followedUser, err := queries.GetUserByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				encodeErrorResponse(r.Context(), http.StatusNotFound, []error{errors.New("profile not found")}, w)
				return
			}
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		// Prevent self-follow
		if followedUser.ID == followerID {
			encodeErrorResponse(r.Context(), http.StatusUnprocessableEntity, []error{errors.New("cannot follow yourself")}, w)
			return
		}

		// Create follow relationship
		err = queries.CreateFollow(r.Context(), sqlite.CreateFollowParams{
			FollowerID: followerID,
			FollowedID: followedUser.ID,
		})
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusOK, profileGetResponseWrapper{
			Profile: profileGetResponseBody{
				Username:  followedUser.Username,
				Bio:       followedUser.Bio.String,
				Image:     followedUser.Image.String,
				Following: true,
			},
		}, w)
	}
}
