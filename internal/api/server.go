// Package api wires the HTTP routes of the workout log.
package api

import (
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
}

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
		r.Get("/health", s.handleHealth)

		r.Post("/auth/signup", s.handleSignup)
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/logout", s.handleLogout)

		r.Group(func(r chi.Router) {
			r.Use(s.requireUser)
			r.Get("/me", s.handleMe)
			r.Put("/me", s.handleUpdateProfile)

			r.Get("/exercises", s.handleListExercises)
			r.Post("/exercises", s.handleCreateExercise)
			r.Get("/exercises/{id}", s.handleGetExercise)
			r.Put("/exercises/{id}", s.handleUpdateExercise)
			r.Post("/exercises/{id}/archive", s.handleArchiveExercise(true))
			r.Post("/exercises/{id}/restore", s.handleArchiveExercise(false))

			r.Get("/workouts", s.handleListWorkouts)
			r.Post("/workouts", s.handleCreateWorkout)
			r.Get("/workouts/{id}", s.handleGetWorkout)
			r.Put("/workouts/{id}", s.handleUpdateWorkout)
			r.Delete("/workouts/{id}", s.handleDeleteWorkout)
			r.Post("/workouts/{id}/finish", s.handleFinishWorkout)
			r.Post("/workouts/{id}/sets", s.handleAddSet)
			r.Put("/workouts/{id}/sets/{setID}", s.handleUpdateSet)
			r.Delete("/workouts/{id}/sets/{setID}", s.handleDeleteSet)
		})

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusNotFound, "not_found", "No such endpoint.")
		})
	})
	s.router = r
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
