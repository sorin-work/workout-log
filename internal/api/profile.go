package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

type profileRequest struct {
	DisplayName string `json:"displayName"`
	Unit        string `json:"unit"`
	Timezone    string `json:"timezone"`
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req profileRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	p := store.Profile{
		DisplayName: strings.TrimSpace(req.DisplayName),
		Unit:        req.Unit,
		Timezone:    strings.TrimSpace(req.Timezone),
	}
	fields := map[string]string{}
	if len(p.DisplayName) > 60 {
		fields["displayName"] = "Name must be at most 60 characters."
	}
	if _, err := domain.ParseUnit(p.Unit); err != nil {
		fields["unit"] = "Unit must be kg or lb."
	}
	if p.Timezone == "" {
		p.Timezone = "UTC"
	}
	if _, err := time.LoadLocation(p.Timezone); err != nil || p.Timezone == "Local" {
		fields["timezone"] = "Unknown time zone."
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	u, err := s.store.UpdateProfile(r.Context(), currentUser(r.Context()).ID, p)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}
