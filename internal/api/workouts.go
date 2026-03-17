package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sorin-work/workout-log/internal/store"
)

const (
	maxWorkoutNameLength  = 80
	maxWorkoutNotesLength = 4000
)

type workoutRequest struct {
	Name       string     `json:"name"`
	Notes      string     `json:"notes"`
	StartedAt  *time.Time `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
}

func (s *Server) validateWorkout(req workoutRequest) (store.WorkoutInput, map[string]string) {
	in := store.WorkoutInput{
		Name:       strings.TrimSpace(req.Name),
		Notes:      strings.TrimSpace(req.Notes),
		FinishedAt: req.FinishedAt,
	}
	fields := map[string]string{}
	if in.Name == "" {
		in.Name = "Workout"
	}
	if len(in.Name) > maxWorkoutNameLength {
		fields["name"] = "Name must be at most 80 characters."
	}
	if len(in.Notes) > maxWorkoutNotesLength {
		fields["notes"] = "Notes must be at most 4000 characters."
	}
	now := s.now()
	if req.StartedAt == nil {
		in.StartedAt = now
	} else {
		in.StartedAt = *req.StartedAt
	}
	if in.StartedAt.After(now.Add(24 * time.Hour)) {
		fields["startedAt"] = "A workout cannot start in the future."
	}
	if in.FinishedAt != nil && in.FinishedAt.Before(in.StartedAt) {
		fields["finishedAt"] = "A workout cannot finish before it starts."
	}
	return in, fields
}

func (s *Server) handleCreateWorkout(w http.ResponseWriter, r *http.Request) {
	var req workoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, fields := s.validateWorkout(req)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	wo, err := s.store.CreateWorkout(r.Context(), currentUser(r.Context()).ID, in)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, wo)
}

// workoutPage is one page of the workout history.
type workoutPage struct {
	Items      []store.WorkoutSummary `json:"items"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	Total      int                    `json:"total"`
	TotalPages int                    `json:"totalPages"`
}

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

func (s *Server) handleListWorkouts(w http.ResponseWriter, r *http.Request) {
	page, size, ok := pagination(w, r)
	if !ok {
		return
	}
	items, total, err := s.store.ListWorkouts(r.Context(), currentUser(r.Context()).ID, size, (page-1)*size)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, workoutPage{
		Items: items, Page: page, PageSize: size, Total: total,
		TotalPages: (total + size - 1) / size,
	})
}

// pagination reads the 1-based page and pageSize query parameters.
func pagination(w http.ResponseWriter, r *http.Request) (page, size int, ok bool) {
	page, size = 1, defaultPageSize
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "bad_request", "page must be a positive number.")
			return 0, 0, false
		}
		page = n
	}
	if v := q.Get("pageSize"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxPageSize {
			writeError(w, http.StatusBadRequest, "bad_request", "pageSize must be between 1 and 100.")
			return 0, 0, false
		}
		size = n
	}
	return page, size, true
}

func (s *Server) handleGetWorkout(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	wo, err := s.store.WorkoutByID(r.Context(), currentUser(r.Context()).ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Workout")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, wo)
}

func (s *Server) handleUpdateWorkout(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req workoutRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, fields := s.validateWorkout(req)
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	wo, err := s.store.UpdateWorkout(r.Context(), currentUser(r.Context()).ID, id, in)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Workout")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, wo)
}

func (s *Server) handleDeleteWorkout(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	err := s.store.DeleteWorkout(r.Context(), currentUser(r.Context()).ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Workout")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleFinishWorkout(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	wo, err := s.store.FinishWorkout(r.Context(), currentUser(r.Context()).ID, id, s.now())
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Workout")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	changes, err := s.store.ApplyProgression(r.Context(), wo)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, finishResponse{Workout: wo, Progression: changes})
}

type finishResponse struct {
	Workout     store.Workout             `json:"workout"`
	Progression []store.ProgressionChange `json:"progression"`
}

func (s *Server) handleRepeatWorkout(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	wo, err := s.store.RepeatWorkout(r.Context(), currentUser(r.Context()).ID, id, s.now())
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Workout")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, wo)
}
