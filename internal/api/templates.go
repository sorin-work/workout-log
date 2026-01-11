package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

const maxTemplateExercises = 20

type templateRequest struct {
	Name      string `json:"name"`
	Notes     string `json:"notes"`
	Exercises []struct {
		ID          int64   `json:"id"`
		ExerciseID  int64   `json:"exerciseId"`
		Sets        int     `json:"sets"`
		Reps        int     `json:"reps"`
		WeightKg    float64 `json:"weightKg"`
		RestSeconds *int    `json:"restSeconds"`
	} `json:"exercises"`
}

// validateTemplate checks the request and writes a 422 on failure.
func (s *Server) validateTemplate(w http.ResponseWriter, r *http.Request, req templateRequest) (store.TemplateInput, bool) {
	in := store.TemplateInput{
		Name:  strings.TrimSpace(req.Name),
		Notes: strings.TrimSpace(req.Notes),
	}
	fields := map[string]string{}
	switch {
	case in.Name == "":
		fields["name"] = "Name is required."
	case len(in.Name) > 60:
		fields["name"] = "Name must be at most 60 characters."
	}
	if len(req.Exercises) > maxTemplateExercises {
		fields["exercises"] = "A template can have at most 20 exercises."
	}
	userID := currentUser(r.Context()).ID
	for i, e := range req.Exercises {
		key := fmt.Sprintf("exercises.%d", i)
		rest := 120
		if e.RestSeconds != nil {
			rest = *e.RestSeconds
		}
		switch {
		case e.Sets < 1 || e.Sets > 20:
			fields[key+".sets"] = "Sets must be between 1 and 20."
		case e.Reps < 1 || e.Reps > domain.MaxReps:
			fields[key+".reps"] = "Reps must be between 1 and 100."
		case e.WeightKg < 0 || e.WeightKg > domain.MaxWeightKg:
			fields[key+".weightKg"] = "Weight must be between 0 and 1000 kg."
		case rest < 0 || rest > 1800:
			fields[key+".restSeconds"] = "Rest must be between 0 and 30 minutes."
		}
		if _, err := s.store.ExerciseByID(r.Context(), userID, e.ExerciseID); err != nil {
			if !errors.Is(err, store.ErrNotFound) {
				writeInternal(w, r, err)
				return in, false
			}
			fields[key+".exerciseId"] = "Choose an exercise from the library."
		}
		in.Exercises = append(in.Exercises, store.TemplateExerciseInput{
			ID: e.ID, ExerciseID: e.ExerciseID, Sets: e.Sets, Reps: e.Reps,
			WeightKg: e.WeightKg, RestSeconds: rest,
		})
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return in, false
	}
	return in, true
}

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListTemplates(r.Context(), currentUser(r.Context()).ID)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	t, err := s.store.TemplateByID(r.Context(), currentUser(r.Context()).ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Template")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req templateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, ok := s.validateTemplate(w, r, req)
	if !ok {
		return
	}
	t, err := s.store.CreateTemplate(r.Context(), currentUser(r.Context()).ID, in)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req templateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, ok := s.validateTemplate(w, r, req)
	if !ok {
		return
	}
	t, err := s.store.UpdateTemplate(r.Context(), currentUser(r.Context()).ID, id, in)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Template")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	err := s.store.DeleteTemplate(r.Context(), currentUser(r.Context()).ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Template")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStartTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	wo, err := s.store.StartWorkoutFromTemplate(r.Context(), currentUser(r.Context()).ID, id, s.now())
	if errors.Is(err, store.ErrNotFound) {
		writeNotFound(w, "Template")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, wo)
}
