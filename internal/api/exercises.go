package api

import (
	"errors"
	"net/http"
	"strings"

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

const maxExerciseNameLength = 40

type exerciseRequest struct {
	Name             string   `json:"name"`
	PrimaryMuscle    string   `json:"primaryMuscle"`
	SecondaryMuscles []string `json:"secondaryMuscles"`
	Equipment        string   `json:"equipment"`
}

func (in exerciseRequest) validate() (store.ExerciseInput, map[string]string) {
	out := store.ExerciseInput{
		Name:             strings.Join(strings.Fields(in.Name), " "),
		PrimaryMuscle:    in.PrimaryMuscle,
		SecondaryMuscles: []string{},
		Equipment:        in.Equipment,
	}
	fields := map[string]string{}
	switch {
	case out.Name == "":
		fields["name"] = "Name is required."
	case textLength(out.Name) > maxExerciseNameLength:
		fields["name"] = "Name must be at most 40 characters."
	}
	if !domain.IsMuscleGroup(out.PrimaryMuscle) {
		fields["primaryMuscle"] = "Choose a muscle group."
	}
	seen := map[string]bool{out.PrimaryMuscle: true}
	for _, m := range in.SecondaryMuscles {
		if !domain.IsMuscleGroup(m) {
			fields["secondaryMuscles"] = "Unknown muscle group " + m + "."
			break
		}
		if !seen[m] {
			seen[m] = true
			out.SecondaryMuscles = append(out.SecondaryMuscles, m)
		}
	}
	if !domain.IsEquipment(out.Equipment) {
		fields["equipment"] = "Choose the equipment."
	}
	return out, fields
}

func (s *Server) handleCreateExercise(w http.ResponseWriter, r *http.Request) {
	var req exerciseRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, fields := req.validate()
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	e, err := s.store.CreateExercise(r.Context(), currentUser(r.Context()).ID, in)
	if errors.Is(err, store.ErrConflict) {
		writeValidation(w, map[string]string{"name": "An exercise with this name already exists."})
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) handleUpdateExercise(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req exerciseRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, fields := req.validate()
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	if !s.ownsExercise(w, r, id) {
		return
	}
	e, err := s.store.UpdateExercise(r.Context(), currentUser(r.Context()).ID, id, in)
	switch {
	case errors.Is(err, store.ErrConflict):
		writeValidation(w, map[string]string{"name": "An exercise with this name already exists."})
	case err != nil:
		writeInternal(w, r, err)
	default:
		writeJSON(w, http.StatusOK, e)
	}
}

func (s *Server) handleArchiveExercise(archived bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "id")
		if !ok || !s.ownsExercise(w, r, id) {
			return
		}
		e, err := s.store.SetExerciseArchived(r.Context(), currentUser(r.Context()).ID, id, archived)
		if err != nil {
			writeInternal(w, r, err)
			return
		}
		writeJSON(w, http.StatusOK, e)
	}
}

// ownsExercise checks that the exercise exists for the user and is a custom
// one. Built-in exercises are read-only.
func (s *Server) ownsExercise(w http.ResponseWriter, r *http.Request, id int64) bool {
	e, err := s.store.ExerciseByID(r.Context(), currentUser(r.Context()).ID, id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeNotFound(w, "Exercise")
		return false
	case err != nil:
		writeInternal(w, r, err)
		return false
	case !e.Custom:
		writeError(w, http.StatusForbidden, "forbidden", "Built-in exercises cannot be changed.")
		return false
	}
	return true
}

type exerciseNotes struct {
	Notes string `json:"notes"`
}

func (s *Server) handleGetExerciseNotes(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	userID := currentUser(r.Context()).ID
	if _, err := s.store.ExerciseByID(r.Context(), userID, id); err != nil {
		writeStoreError(w, r, err, "Exercise")
		return
	}
	notes, err := s.store.ExerciseNotes(r.Context(), userID, id)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, exerciseNotes{Notes: notes})
}

func (s *Server) handlePutExerciseNotes(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req exerciseNotes
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Notes = strings.TrimSpace(req.Notes)
	if textLength(req.Notes) > 1000 {
		writeValidation(w, map[string]string{"notes": "Notes must be at most 1000 characters."})
		return
	}
	userID := currentUser(r.Context()).ID
	if _, err := s.store.ExerciseByID(r.Context(), userID, id); err != nil {
		writeStoreError(w, r, err, "Exercise")
		return
	}
	if err := s.store.SetExerciseNotes(r.Context(), userID, id, req.Notes); err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}
