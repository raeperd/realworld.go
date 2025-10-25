package main

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	"github.com/raeperd/realworld.go/internal/auth"
)

func main() {
	if err := run(context.Background(), os.Stdout, os.Args, Version); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// Version is set at build time using ldflags.
// It is optional and can be omitted if not required.
// Refer to [handleGetHealth] for more information.
var Version string

//go:embed internal/sqlite/schema.sql
var ddl string

// run initiates and starts the [http.Server], blocking until the context is canceled by OS signals.
// It listens on a port specified by the -port flag, defaulting to 8080.
// This function is inspired by techniques discussed in the [blog post] By Mat Ryer:
//
// [blog post]: https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years
func run(ctx context.Context, w io.Writer, args []string, version string) error {
	var port uint
	var jwtSecret string
	var dbPath string
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	fs.SetOutput(w)
	fs.UintVar(&port, "port", 8080, "port for HTTP API")
	fs.StringVar(&jwtSecret, "jwt-secret", "default-secret", "JWT signing secret")
	fs.StringVar(&dbPath, "db", "", "database connection string (empty for in-memory)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// NOTE: Removed `defer cancel()` since we want to control when to cancel the context
	// We'll call it explicitly after server shutdown

	// Initialize your resources here, for example:
	// - Database connections
	// - Message queue clients
	// - Cache clients
	// - External API clients
	// Example:
	// db, err := sql.Open(...)
	// if err != nil {
	//     return fmt.Errorf("database init: %w", err)
	// }

	slog.SetDefault(slog.New(slog.NewJSONHandler(w, nil)))

	// Use file database if provided, otherwise in-memory
	dbConnection := ":memory:"
	if dbPath != "" {
		dbConnection = dbPath
	}

	db, err := sql.Open("sqlite", dbConnection)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck

	// Limit to single connection to prevent SQLite locking issues with parallel tests
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys=ON"); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           route(slog.Default(), version, db, jwtSecret),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "server started", slog.Uint64("port", uint64(port)), slog.String("version", version))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.InfoContext(ctx, "shutting down server")

		// Create a new context for shutdown with timeout
		ctx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		// Shutdown the HTTP server first
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}

		// After server is shutdown, cancel the main context to close other resources
		cancel()

		// Add cleanup code here, in reverse order of initialization
		// Give each cleanup operation its own timeout if needed

		// Example cleanup sequence:
		// 1. Close application services that depend on other resources
		// if err := myService.Shutdown(ctx); err != nil {
		//     return fmt.Errorf("service shutdown: %w", err)
		// }

		// 2. Close message queue connections
		// if err := mqClient.Close(); err != nil {
		//     return fmt.Errorf("mq shutdown: %w", err)
		// }

		// 3. Close cache connections
		// if err := cacheClient.Close(); err != nil {
		//     return fmt.Errorf("cache shutdown: %w", err)
		// }

		// 4. Close database connections
		// if err := db.Close(); err != nil {
		//     return fmt.Errorf("database shutdown: %w", err)
		// }
		return nil
	}
}

// route sets up and returns an [http.Handler] for all the server routes.
// It is the single source of truth for all the routes.
// You can add custom [http.Handler] as needed.
func route(log *slog.Logger, version string, db *sql.DB, jwtSecret string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /health", handleGetHealth(version))
	mux.Handle("GET /openapi.yaml", handleGetOpenAPI(version))
	mux.Handle("/debug/", handleGetDebug())

	mux.HandleFunc("POST /api/users", handlePostUsers(db, jwtSecret))
	mux.HandleFunc("POST /api/users/login", handlePostUsersLogin(db, jwtSecret))
	mux.Handle("GET /api/user", authenticate(handleGetUser(db, jwtSecret), jwtSecret))
	mux.Handle("PUT /api/user", authenticate(handlePutUser(db, jwtSecret), jwtSecret))
	mux.Handle("GET /api/profiles/{username}", authenticateOptional(handleGetProfilesUsername(db), jwtSecret))
	mux.Handle("POST /api/profiles/{username}/follow", authenticate(handlePostProfilesUsernameFollow(db), jwtSecret))
	mux.Handle("DELETE /api/profiles/{username}/follow", authenticate(handleDeleteProfilesUsernameFollow(db), jwtSecret))
	mux.HandleFunc("GET /api/tags", handleGetTags(db))
	mux.Handle("POST /api/articles", authenticate(handlePostArticles(db), jwtSecret))
	mux.Handle("GET /api/articles/{slug}", authenticateOptional(handleGetArticlesSlug(db), jwtSecret))
	mux.Handle("PUT /api/articles/{slug}", authenticate(handlePutArticlesSlug(db), jwtSecret))
	mux.Handle("DELETE /api/articles/{slug}", authenticate(handleDeleteArticlesSlug(db), jwtSecret))
	mux.Handle("POST /api/articles/{slug}/comments", authenticate(handlePostArticlesSlugComments(db), jwtSecret))
	mux.Handle("GET /api/articles/{slug}/comments", authenticateOptional(handleGetArticlesSlugComments(db), jwtSecret))

	handler := cors(mux)
	handler = accesslog(handler, log)
	handler = recovery(handler, log)
	return handler
}

// handleGetHealth returns an [http.HandlerFunc] that responds with the health status of the service.
// It includes the service version, VCS revision, build time, and modified status.
// The service version can be set at build time using the VERSION variable (e.g., 'make build VERSION=v1.0.0').
func handleGetHealth(version string) http.HandlerFunc {
	type responseBody struct {
		Version        string    `json:"Version"`
		Uptime         string    `json:"Uptime"`
		LastCommitHash string    `json:"LastCommitHash"`
		LastCommitTime time.Time `json:"LastCommitTime"`
		DirtyBuild     bool      `json:"DirtyBuild"`
	}

	baseRes := responseBody{Version: version}
	buildInfo, _ := debug.ReadBuildInfo()
	for _, kv := range buildInfo.Settings {
		if kv.Value == "" {
			continue
		}
		switch kv.Key {
		case "vcs.revision":
			baseRes.LastCommitHash = kv.Value
		case "vcs.time":
			baseRes.LastCommitTime, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			baseRes.DirtyBuild = kv.Value == "true"
		}
	}

	up := time.Now()
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		res := baseRes // Create a copy for each request to avoid data race
		res.Uptime = time.Since(up).String()
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// handleGetDebug returns an [http.Handler] for debug routes, including pprof and expvar routes.
func handleGetDebug() http.Handler {
	mux := http.NewServeMux()

	// NOTE: this route is same as defined in net/http/pprof init function
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// NOTE: this route is same as defined in expvar init function
	mux.Handle("/debug/vars", expvar.Handler())
	return mux
}

// handleGetOpenAPI returns an [http.HandlerFunc] that serves the OpenAPI specification YAML file.
// The file is embedded in the binary using the go:embed directive.
func handleGetOpenAPI(version string) http.HandlerFunc {
	body := bytes.Replace(openAPI, []byte("${{ VERSION }}"), []byte(version), 1)
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}
}

// openAPI holds the embedded OpenAPI YAML file.
// Remove this and the api/openapi.yaml file if you prefer not to serve OpenAPI.
//
//go:embed api/openapi.yaml
var openAPI []byte

// accesslog is a middleware that logs request and response details,
// including latency, method, path, query parameters, IP address, response status, and bytes sent.
func accesslog(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := responseRecorder{ResponseWriter: w}

		next.ServeHTTP(&wr, r)

		log.InfoContext(r.Context(), "accessed",
			slog.String("latency", time.Since(start).String()),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.String("ip", r.RemoteAddr),
			slog.Int("status", wr.status),
			slog.Int("bytes", wr.numBytes))
	})
}

// recovery is a middleware that recovers from panics during HTTP handler execution and logs the error details.
// It must be the last middleware in the chain to ensure it captures all panics.
func recovery(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr := responseRecorder{ResponseWriter: w}
		defer func() {
			err := recover()
			if err == nil {
				return
			}

			if err, ok := err.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				// Handle the abort gracefully
				return
			}

			stack := make([]byte, 1024)
			n := runtime.Stack(stack, true)

			log.ErrorContext(r.Context(), "panic!",
				slog.Any("error", err),
				slog.String("stack", string(stack[:n])),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.String("ip", r.RemoteAddr))

			if wr.status > 0 {
				// response was already sent, nothing we can do
				return
			}

			// send error response
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}()
		next.ServeHTTP(&wr, r)
	})
}

// cors is a middleware that handles CORS (Cross-Origin Resource Sharing) for the API.
// It allows all origins to access the API endpoints, which is necessary for RealWorld frontend compatibility.
// TODO: Add Access-Control-Allow-Methods, Access-Control-Allow-Headers, Access-Control-Max-Age
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// authenticate is a middleware that validates JWT tokens and attaches user ID to the request context.
// It expects the token in the "Authorization: Token <jwt>" header format.
// Returns 401 Unauthorized if the token is missing or invalid.
func authenticate(next http.Handler, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("missing authorization header")}, w)
			return
		}

		// Extract token from "Token <jwt>" format
		tokenString := strings.TrimPrefix(authHeader, "Token ")
		if tokenString == authHeader {
			// "Token " prefix not found
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("invalid authorization header format")}, w)
			return
		}

		// Parse and validate token
		claims, err := auth.ParseToken(tokenString, jwtSecret)
		if err != nil {
			encodeErrorResponse(r.Context(), http.StatusUnauthorized, []error{errors.New("invalid or expired token")}, w)
			return
		}

		// Store user ID in context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authenticateOptional is a middleware that validates JWT tokens if present and attaches user ID to the request context.
// Unlike authenticate, this middleware does not return an error if the token is missing.
// If a token is provided but invalid, it continues without setting the user ID in context.
func authenticateOptional(next http.Handler, jwtSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without user ID
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from "Token <jwt>" format
		tokenString := strings.TrimPrefix(authHeader, "Token ")
		if tokenString == authHeader {
			// "Token " prefix not found, continue without user ID
			next.ServeHTTP(w, r)
			return
		}

		// Parse and validate token
		claims, err := auth.ParseToken(tokenString, jwtSecret)
		if err != nil {
			// Invalid token, continue without user ID
			next.ServeHTTP(w, r)
			return
		}

		// Store user ID in context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseRecorder is a wrapper around [http.ResponseWriter] that records the status and bytes written during the response.
// It implements the [http.ResponseWriter] interface by embedding the original ResponseWriter.
type responseRecorder struct {
	http.ResponseWriter
	status   int
	numBytes int
}

// Header implements the [http.ResponseWriter] interface.
func (re *responseRecorder) Header() http.Header {
	return re.ResponseWriter.Header()
}

// Write implements the [http.ResponseWriter] interface.
func (re *responseRecorder) Write(b []byte) (int, error) {
	re.numBytes += len(b)
	return re.ResponseWriter.Write(b)
}

// WriteHeader implements the [http.ResponseWriter] interface.
func (re *responseRecorder) WriteHeader(statusCode int) {
	re.status = statusCode
	re.ResponseWriter.WriteHeader(statusCode)
}
