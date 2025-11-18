package api

import (
	"errors"
	"net/http"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

func (s *Server) handleListExercises(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ExerciseFilter{
		Query:           q.Get("q"),
		Muscle:          q.Get("muscle"),
		Equipment:       q.Get("equipment"),
		IncludeArchived: q.Get("archived") == "true",
	}
	if f.Muscle != "" && !domain.IsMuscleGroup(f.Muscle) {
		writeError(w, http.StatusBadRequest, "bad_request", "Unknown muscle group.")
		return
	}
	if f.Equipment != "" && !domain.IsEquipment(f.Equipment) {
		writeError(w, http.StatusBadRequest, "bad_request", "Unknown equipment type.")
		return
	}
	list, err := s.store.ListExercises(r.Context(), currentUser(r.Context()).ID, f)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetExercise(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	e, err := s.store.ExerciseByID(r.Context(), currentUser(r.Context()).ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Exercise")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}
