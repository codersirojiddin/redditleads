package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// RouterDeps carries every handler + auth middleware the router needs to
// wire up. Kept as an interface-free struct of function values so this
// package doesn't need to import every domain package's concrete types
// (avoids import cycles: domain handlers already import this http package
// for JSON/error helpers).
type RouterDeps struct {
	// AllowedOrigin is the frontend origin allowed via CORS (e.g. FRONTEND_URL).
	// Empty string disables the CORS headers entirely.
	AllowedOrigin string

	AuthMiddleware func(http.Handler) http.Handler

	RegisterHandler http.HandlerFunc
	LoginHandler    http.HandlerFunc

	MeHandler http.HandlerFunc

	CreateProjectHandler http.HandlerFunc
	ListProjectsHandler  http.HandlerFunc
	GetProjectHandler    http.HandlerFunc
	DeleteProjectHandler http.HandlerFunc

	CreateKeywordHandler http.HandlerFunc
	ListKeywordsHandler  http.HandlerFunc
	DeleteKeywordHandler http.HandlerFunc

	CreateReportHandler http.HandlerFunc
	ListReportsHandler  http.HandlerFunc
	GetReportHandler    http.HandlerFunc
}

// NewRouter builds the full API route tree.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(corsMiddleware(deps.AllowedOrigin))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", deps.RegisterHandler)
		r.Post("/auth/login", deps.LoginHandler)

		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware)

			r.Get("/me", deps.MeHandler)

			r.Route("/projects", func(r chi.Router) {
				r.Post("/", deps.CreateProjectHandler)
				r.Get("/", deps.ListProjectsHandler)

				r.Route("/{projectID}", func(r chi.Router) {
					r.Get("/", deps.GetProjectHandler)
					r.Delete("/", deps.DeleteProjectHandler)

					r.Route("/keywords", func(r chi.Router) {
						r.Post("/", deps.CreateKeywordHandler)
						r.Get("/", deps.ListKeywordsHandler)
						r.Delete("/{keywordID}", deps.DeleteKeywordHandler)
					})

					r.Route("/reports", func(r chi.Router) {
						r.Post("/", deps.CreateReportHandler)
						r.Get("/", deps.ListReportsHandler)
					})
				})
			})

			r.Get("/reports/{reportID}", deps.GetReportHandler)
		})
	})

	return r
}

// corsMiddleware allows the configured frontend origin to call the API with
// credentials (Authorization header) from the browser. Kept dependency-free
// (no external cors package) since the rules are simple.
func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "600")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
