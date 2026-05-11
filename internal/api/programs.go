package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/sorin-work/workout-log/internal/store"
)

const maxProgramDays = 14

type programRequest struct {
	Name        string  `json:"name"`
	TemplateIDs []int64 `json:"templateIds"`
}

func (s *Server) validateProgram(w http.ResponseWriter, r *http.Request, req programRequest) (store.ProgramInput, bool) {
	in := store.ProgramInput{Name: strings.TrimSpace(req.Name), TemplateIDs: req.TemplateIDs}
	fields := map[string]string{}
	switch {
	case in.Name == "":
		fields["name"] = "Name is required."
	case textLength(in.Name) > 60:
		fields["name"] = "Name must be at most 60 characters."
	}
	switch {
	case len(in.TemplateIDs) == 0:
		fields["templateIds"] = "Add at least one workout to the rotation."
	case len(in.TemplateIDs) > maxProgramDays:
		fields["templateIds"] = "A rotation can have at most 14 workouts."
	}
	userID := currentUser(r.Context()).ID
	for _, id := range in.TemplateIDs {
		if _, err := s.store.TemplateByID(r.Context(), userID, id); err != nil {
			if !errors.Is(err, store.ErrNotFound) {
				writeInternal(w, r, err)
				return in, false
			}
			fields["templateIds"] = "Choose templates you created."
			break
		}
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return in, false
	}
	return in, true
}

func (s *Server) handleListPrograms(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListPrograms(r.Context(), currentUser(r.Context()).ID)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateProgram(w http.ResponseWriter, r *http.Request) {
	var req programRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, ok := s.validateProgram(w, r, req)
	if !ok {
		return
	}
	p, err := s.store.CreateProgram(r.Context(), currentUser(r.Context()).ID, in)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleGetProgram(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	p, err := s.store.ProgramByID(r.Context(), currentUser(r.Context()).ID, id)
	s.writeProgram(w, r, p, err)
}

func (s *Server) handleUpdateProgram(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var req programRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	in, ok := s.validateProgram(w, r, req)
	if !ok {
		return
	}
	p, err := s.store.UpdateProgram(r.Context(), currentUser(r.Context()).ID, id, in)
	s.writeProgram(w, r, p, err)
}

func (s *Server) handleActivateProgram(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	p, err := s.store.ActivateProgram(r.Context(), currentUser(r.Context()).ID, id)
	s.writeProgram(w, r, p, err)
}

func (s *Server) writeProgram(w http.ResponseWriter, r *http.Request, p store.Program, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeNotFound(w, "Program")
	case err != nil:
		writeInternal(w, r, err)
	default:
		writeJSON(w, http.StatusOK, p)
	}
}

func (s *Server) handleDeleteProgram(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	err := s.store.DeleteProgram(r.Context(), currentUser(r.Context()).ID, id)
	if err != nil {
		writeStoreError(w, r, err, "Program")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleStartProgram(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	wo, err := s.store.StartProgramDay(r.Context(), currentUser(r.Context()).ID, id, s.now())
	if err != nil {
		writeStoreError(w, r, err, "Program")
		return
	}
	writeJSON(w, http.StatusCreated, wo)
}
