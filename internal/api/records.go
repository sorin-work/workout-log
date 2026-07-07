package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

func (s *Server) handleListRecords(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r.Context())
	sets, err := s.store.LoggedSets(r.Context(), u.ID, 0)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.ComputeRecords(sets, domain.Formula(u.Formula)))
}

func (s *Server) handleExerciseRecords(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	userID := currentUser(r.Context()).ID
	ex, err := s.store.ExerciseByID(r.Context(), userID, id)
	if err != nil {
		writeStoreError(w, r, err, "Exercise")
		return
	}
	sets, err := s.store.LoggedSets(r.Context(), userID, id)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	recs := domain.ComputeRecords(sets, domain.Formula(currentUser(r.Context()).Formula))
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
	writeJSON(w, http.StatusOK, domain.ProgressSeries(sets, domain.Formula(currentUser(r.Context()).Formula)))
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
	if err != nil {
		writeStoreError(w, r, err, "Previous session")
		return
	}
	writeJSON(w, http.StatusOK, last)
}

func (s *Server) handleExerciseHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 50 {
			writeError(w, http.StatusBadRequest, "bad_request", "limit must be between 1 and 50.")
			return
		}
		limit = n
	}
	userID := currentUser(r.Context()).ID
	if _, err := s.store.ExerciseByID(r.Context(), userID, id); err != nil {
		writeStoreError(w, r, err, "Exercise")
		return
	}
	list, err := s.store.ExerciseHistory(r.Context(), userID, id, limit)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}
