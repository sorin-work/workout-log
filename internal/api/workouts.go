package api

import (
	"errors"
	"net/http"
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

func (s *Server) handleListWorkouts(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListWorkouts(r.Context(), currentUser(r.Context()).ID)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
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
