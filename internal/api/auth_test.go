package api

import (
	"net/http"
	"testing"
)

func TestSignupLogsIn(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("lifter@example.com")
	var me struct {
		Email string `json:"email"`
	}
	c.mustDo(http.MethodGet, "/api/me", nil, http.StatusOK, &me)
	if me.Email != "lifter@example.com" {
		t.Fatalf("me.email = %q", me.Email)
	}
}

func TestSignupValidation(t *testing.T) {
	a := newTestAPI(t)
	res := a.client().do(http.MethodPost, "/api/auth/signup", map[string]any{
		"email": "not-an-email", "password": "short",
	})
	if res.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
	f := errorFields(t, res)
	if f["email"] == "" || f["password"] == "" {
		t.Fatalf("expected email and password errors, got %v", f)
	}
}

func TestSignupDuplicateEmail(t *testing.T) {
	a := newTestAPI(t)
	a.signup("dup@example.com")
	res := a.client().do(http.MethodPost, "/api/auth/signup", map[string]any{
		"email": "dup@example.com", "password": "password123",
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["email"] == "" {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
}

func TestLogin(t *testing.T) {
	a := newTestAPI(t)
	a.signup("login@example.com")

	c := a.client()
	res := c.do(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "login@example.com", "password": "wrong-password",
	})
	if res.Status != http.StatusUnauthorized {
		t.Fatalf("wrong password: status %d", res.Status)
	}
	c.mustDo(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "LOGIN@example.com", "password": "password123",
	}, http.StatusOK, nil)
	c.mustDo(http.MethodGet, "/api/me", nil, http.StatusOK, nil)
}

func TestLoginUnknownEmail(t *testing.T) {
	a := newTestAPI(t)
	res := a.client().do(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "nobody@example.com", "password": "password123",
	})
	if res.Status != http.StatusUnauthorized {
		t.Fatalf("status %d", res.Status)
	}
}

func TestLogoutEndsSession(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("bye@example.com")
	c.mustDo(http.MethodPost, "/api/auth/logout", nil, http.StatusNoContent, nil)
	c.mustDo(http.MethodGet, "/api/me", nil, http.StatusUnauthorized, nil)
}

func TestMeRequiresSession(t *testing.T) {
	a := newTestAPI(t)
	a.client().mustDo(http.MethodGet, "/api/me", nil, http.StatusUnauthorized, nil)
}

func TestLogoutOtherDevices(t *testing.T) {
	a := newTestAPI(t)
	phone := a.signup("multi@example.com")
	laptop := a.client()
	laptop.mustDo(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "multi@example.com", "password": "password123",
	}, http.StatusOK, nil)
	var res map[string]int
	phone.mustDo(http.MethodPost, "/api/me/logout-others", nil, http.StatusOK, &res)
	if res["loggedOut"] != 1 {
		t.Fatalf("result: %v", res)
	}
	phone.mustDo(http.MethodGet, "/api/me", nil, http.StatusOK, nil)
	laptop.mustDo(http.MethodGet, "/api/me", nil, http.StatusUnauthorized, nil)
}
