package main

import (
	"database/sql"
	"net/http"

	"github.com/raeperd/realworld.go/internal/sqlite"
)

func handleGetTags(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}
		defer func() { _ = tx.Rollback() }()

		queries := sqlite.New(tx)
		tags, err := queries.GetAllTags(r.Context())
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		if err := tx.Commit(); err != nil {
			encodeErrorResponse(r.Context(), http.StatusInternalServerError, []error{err}, w)
			return
		}

		encodeResponse(r.Context(), http.StatusOK, tagsResponseBody{Tags: tags}, w)
	}
}

type tagsResponseBody struct {
	Tags []string `json:"tags"`
}
