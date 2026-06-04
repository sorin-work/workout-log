package api

import (
	"errors"
	"net/http"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

type setRequest struct {
	ExerciseID int64    `json:"exerciseId"`
	Reps       int      `json:"reps"`
	WeightKg   float64  `json:"weightKg"`
	RPE        *float64 `json:"rpe"`
	Completed  bool     `json:"completed"`
	Warmup     bool     `json:"warmup"`
}

// validateSet checks the request and that the exercise is one the user can
// log. It writes the error response itself and returns false on failure.
func (s *Server) validateSet(w http.ResponseWriter, r *http.Request, req setRequest) (store.SetData, bool) {
	fields := domain.ValidateSet(domain.SetInput{
		Reps: req.Reps, WeightKg: req.WeightKg, RPE: req.RPE,
		Completed: req.Completed, Warmup: req.Warmup,
	})
	if _, err := s.store.ExerciseByID(r.Context(), currentUser(r.Context()).ID, req.ExerciseID); err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			writeInternal(w, r, err)
			return store.SetData{}, false
		}
		fields["exerciseId"] = "Choose an exercise from the library."
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return store.SetData{}, false
	}
	return store.SetData{
		ExerciseID: req.ExerciseID, Reps: req.Reps, WeightKg: req.WeightKg,
		RPE: req.RPE, Completed: req.Completed, Warmup: req.Warmup,
	}, true
}

// loadWorkout fetches the workout in the URL, writing a 404 if the user does
// not own it.
func (s *Server) loadWorkout(w http.ResponseWriter, r *http.Request) (store.Workout, bool) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return store.Workout{}, false
	}
	wo, err := s.store.WorkoutByID(r.Context(), currentUser(r.Context()).ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Workout")
		return store.Workout{}, false
	}
	if err != nil {
		writeInternal(w, r, err)
		return store.Workout{}, false
	}
	return wo, true
}

func (s *Server) handleAddSet(w http.ResponseWriter, r *http.Request) {
	wo, ok := s.loadWorkout(w, r)
	if !ok {
		return
	}
	var req setRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	data, ok := s.validateSet(w, r, req)
	if !ok {
		return
	}
	set, err := s.store.AddSet(r.Context(), wo.ID, data)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, set)
}

func (s *Server) handleUpdateSet(w http.ResponseWriter, r *http.Request) {
	wo, ok := s.loadWorkout(w, r)
	if !ok {
		return
	}
	setID, ok := pathID(w, r, "setID")
	if !ok {
		return
	}
	var req setRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	data, ok := s.validateSet(w, r, req)
	if !ok {
		return
	}
	set, err := s.store.UpdateSet(r.Context(), wo.ID, setID, data)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Set")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, set)
}

func (s *Server) handleDeleteSet(w http.ResponseWriter, r *http.Request) {
	wo, ok := s.loadWorkout(w, r)
	if !ok {
		return
	}
	setID, ok := pathID(w, r, "setID")
	if !ok {
		return
	}
	err := s.store.DeleteSet(r.Context(), wo.ID, setID)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Set")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleAddWarmups inserts suggested warm-ups before the working sets of an
// exercise, based on its heaviest planned set.
func (s *Server) handleAddWarmups(w http.ResponseWriter, r *http.Request) {
	wo, ok := s.loadWorkout(w, r)
	if !ok {
		return
	}
	var req struct {
		ExerciseID int64 `json:"exerciseId"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	work := 0.0
	for _, st := range wo.Sets {
		if st.ExerciseID == req.ExerciseID && !st.Warmup && st.WeightKg > work {
			work = st.WeightKg
		}
	}
	if work == 0 {
		writeValidation(w, map[string]string{"exerciseId": "Log a working set with a weight first."})
		return
	}
	var data []store.SetData
	for _, ws := range domain.WarmupSets(work, 20, domain.SmallestPlateStep(domain.Kilograms)) {
		data = append(data, store.SetData{ExerciseID: req.ExerciseID, Reps: ws.Reps, WeightKg: ws.WeightKg, Warmup: true})
	}
	if len(data) == 0 {
		writeJSON(w, http.StatusOK, []store.Set{})
		return
	}
	sets, err := s.store.InsertWarmups(r.Context(), wo.ID, req.ExerciseID, data)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, sets)
}
