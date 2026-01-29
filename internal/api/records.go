package api

import (
	"errors"
	"net/http"

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
