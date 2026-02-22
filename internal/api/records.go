package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

func (s *Server) handleListRecords(w http.ResponseWriter, r *http.Request) {
	sets, err := s.store.LoggedSets(r.Context(), currentUser(r.Context()).ID, 0)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.ComputeRecords(sets, domain.Epley))
}

func (s *Server) handleExerciseRecords(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	userID := currentUser(r.Context()).ID
	ex, err := s.store.ExerciseByID(r.Context(), userID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Exercise")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	sets, err := s.store.LoggedSets(r.Context(), userID, id)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	recs := domain.ComputeRecords(sets, domain.Epley)
	if len(recs) == 0 {
		writeJSON(w, http.StatusOK, domain.ExerciseRecords{
			ExerciseID: ex.ID, ExerciseName: ex.Name, RepMaxes: map[int]domain.RecordSet{},
		})
		return
	}
	writeJSON(w, http.StatusOK, recs[0])
}

func (s *Server) handleExerciseProgress(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	userID := currentUser(r.Context()).ID
	if _, err := s.store.ExerciseByID(r.Context(), userID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeNotFound(w, "Exercise")
			return
		}
		writeInternal(w, r, err)
		return
	}
	sets, err := s.store.LoggedSets(r.Context(), userID, id)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.ProgressSeries(sets, domain.Epley))
}

func (s *Server) handleLastSession(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var exclude int64
	if v := r.URL.Query().Get("excludeWorkout"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "excludeWorkout must be a workout id.")
			return
		}
		exclude = n
	}
	last, err := s.store.LastSession(r.Context(), currentUser(r.Context()).ID, id, exclude)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Previous session")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, last)
}
