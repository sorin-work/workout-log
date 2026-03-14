package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/sorin-work/workout-log/internal/auth"
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

type passwordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req passwordRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	u := currentUser(r.Context())
	if err := auth.CheckPassword(u.PasswordHash, req.CurrentPassword); err != nil {
		writeValidation(w, map[string]string{"currentPassword": "Current password is incorrect."})
		return
	}
	switch {
	case len(req.NewPassword) < auth.MinPasswordLength:
		writeValidation(w, map[string]string{"newPassword": "Password must be at least 8 characters."})
		return
	case len(req.NewPassword) > 72:
		writeValidation(w, map[string]string{"newPassword": "Password must be at most 72 characters."})
		return
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	c, _ := r.Cookie(sessionCookie)
	if err := s.store.UpdatePassword(r.Context(), u.ID, hash, auth.HashToken(c.Value)); err != nil {
		writeInternal(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	u := currentUser(r.Context())
	if err := auth.CheckPassword(u.PasswordHash, req.Password); err != nil {
		writeValidation(w, map[string]string{"password": "Password is incorrect."})
		return
	}
	if err := s.store.DeleteUser(r.Context(), u.ID); err != nil {
		writeInternal(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: s.opts.SecureCookies,
	})
	w.WriteHeader(http.StatusNoContent)
}
