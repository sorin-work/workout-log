package api

import (
	"context"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/sorin-work/workout-log/internal/auth"
	"github.com/sorin-work/workout-log/internal/store"
)

const sessionCookie = "wl_session"

type ctxKey int

const userKey ctxKey = iota

// currentUser returns the user set by requireUser.
func currentUser(ctx context.Context) store.User {
	u, _ := ctx.Value(userKey).(store.User)
	return u
}

type credentials struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	in.Email = strings.TrimSpace(in.Email)
	in.DisplayName = strings.TrimSpace(in.DisplayName)
	fields := map[string]string{}
	if addr, err := mail.ParseAddress(in.Email); err != nil || addr.Address != in.Email {
		fields["email"] = "Enter a valid email address."
	}
	if len(in.Password) < auth.MinPasswordLength {
		fields["password"] = "Password must be at least 8 characters."
	} else if len(in.Password) > 72 {
		fields["password"] = "Password must be at most 72 characters."
	}
	if textLength(in.DisplayName) > 60 {
		fields["displayName"] = "Name must be at most 60 characters."
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	u, err := s.store.CreateUser(r.Context(), in.Email, in.DisplayName, hash)
	if errors.Is(err, store.ErrConflict) {
		writeValidation(w, map[string]string{"email": "An account with this email already exists."})
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	if err := s.startSession(w, r, u.ID); err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if !decodeJSON(w, r, &in) {
		return
	}
	u, err := s.store.UserByEmail(r.Context(), strings.TrimSpace(in.Email))
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect.")
		return
	}
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	if err := auth.CheckPassword(u.PasswordHash, in.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is incorrect.")
		return
	}
	if err := s.startSession(w, r, u.ID); err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(sessionCookie); err == nil {
		if err := s.store.DeleteSession(r.Context(), auth.HashToken(c.Value)); err != nil {
			writeInternal(w, r, err)
			return
		}
	}
	s.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, currentUser(r.Context()))
}

func (s *Server) startSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, hash, err := auth.NewSessionToken()
	if err != nil {
		return err
	}
	expires := s.now().Add(s.opts.SessionTTL)
	if err := s.store.CreateSession(r.Context(), hash, userID, expires); err != nil {
		return err
	}
	http.SetCookie(w, s.sessionCookie(token, expires))
	return nil
}

// requireUser rejects requests without a valid session cookie.
func (s *Server) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil || c.Value == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Please log in.")
			return
		}
		u, err := s.store.UserBySession(r.Context(), auth.HashToken(c.Value))
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Your session has expired. Please log in again.")
			return
		}
		if err != nil {
			writeInternal(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	})
}

func (s *Server) handleLogoutOthers(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(sessionCookie)
	n, err := s.store.DeleteOtherSessions(r.Context(), currentUser(r.Context()).ID, auth.HashToken(c.Value))
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"loggedOut": n})
}

// sessionCookie builds the session cookie with the same attributes for
// setting and clearing it; browsers only replace a cookie when they match.
func (s *Server) sessionCookie(value string, expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.opts.SecureCookies,
	}
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	c := s.sessionCookie("", time.Unix(0, 0))
	c.MaxAge = -1
	http.SetCookie(w, c)
}
