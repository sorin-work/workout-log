// Package api wires the HTTP routes of the workout log.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/sorin-work/workout-log/internal/store"
)

// Options configures a Server.
type Options struct {
	SessionTTL time.Duration
	// SecureCookies marks the session cookie as Secure (HTTPS only).
	SecureCookies bool
	// WebDir, when set, is the directory of the built web app to serve.
	WebDir string
}

// Version is set at build time with -ldflags "-X .../internal/api.Version=...".
var Version = "dev"

// loginAttemptsPerMinute limits sign-up and login requests per client IP.
const loginAttemptsPerMinute = 10

// Server holds the dependencies shared by the HTTP handlers.
type Server struct {
	store  *store.Store
	opts   Options
	router chi.Router
	now    func() time.Time
}

// NewServer builds the router with all routes registered.
func NewServer(st *store.Store, opts Options) *Server {
	if opts.SessionTTL == 0 {
		opts.SessionTTL = 30 * 24 * time.Hour
	}
	s := &Server{store: st, opts: opts, now: time.Now}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Use(requestLogger)
		r.Use(middleware.Compress(5, "application/json", "text/csv"))
		r.Use(noStore)
		r.Get("/health", s.handleHealth)
		r.Get("/version", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"version": Version})
		})

		r.Group(func(r chi.Router) {
			r.Use(newRateLimiter(loginAttemptsPerMinute, time.Minute).middleware)
			r.Post("/auth/signup", s.handleSignup)
			r.Post("/auth/login", s.handleLogin)
		})
		r.Post("/auth/logout", s.handleLogout)

		r.Group(func(r chi.Router) {
			r.Use(s.requireUser)
			r.Route("/me", s.profileRoutes)
			r.Route("/exercises", s.exerciseRoutes)
			r.Route("/templates", s.templateRoutes)
			r.Route("/programs", s.programRoutes)
			r.Route("/workouts", s.workoutRoutes)
			r.Get("/records", s.handleListRecords)
			r.Get("/stats/volume", s.handleWeeklyVolume)
			r.Get("/stats/summary", s.handleSummary)
			r.Get("/stats/calendar", s.handleCalendar)
			r.Get("/export/sets.csv", s.handleExportSets)
			r.Post("/import/sets.csv", s.handleImportSets)
		})

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusNotFound, "not_found", "No such endpoint.")
		})
	})
	if opts.WebDir != "" {
		r.Handle("/*", spaHandler(opts.WebDir))
	}
	s.router = r
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.store.Ping(ctx); err != nil {
		slog.Warn("health check failed", "err", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "database": "down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "ok"})
}

// noStore keeps browsers and proxies from caching API responses, which hold
// personal data and change all the time.
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
