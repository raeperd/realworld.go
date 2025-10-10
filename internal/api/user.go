package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

func HandlePostUsers(w http.ResponseWriter, r *http.Request) {
	var request userPostRequestBody
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	if errs := request.Validate(); len(errs) > 0 {
		encodeErrorResponse(r.Context(), errs, w)
		return
	}

	encodeResponse(r.Context(), http.StatusCreated, userPostResponseBody{
		Email:    request.User.Email,
		Token:    "token",
		Username: request.User.Username,
		Bio:      "bio",
		Image:    "image",
	}, w)
}

func encodeErrorResponse(ctx context.Context, errs []error, w http.ResponseWriter) {
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

	encodeResponse(ctx, http.StatusUnprocessableEntity, errResp, w)
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
