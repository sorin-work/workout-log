package api

import (
	"context"
	"net/http"
	"strings"
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

func TestLoginTrimsEmail(t *testing.T) {
	a := newTestAPI(t)
	a.signup("trim@example.com")
	a.client().mustDo(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "  Trim@Example.com ", "password": "password123",
	}, http.StatusOK, nil)
}

func TestEndedSessionAsksToLogInAgain(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("old@example.com")
	// End the session behind the browser's back, as expiry or a password
	// change on another device would.
	u, err := a.store.UserByEmail(context.Background(), "old@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.store.DeleteOtherSessions(context.Background(), u.ID, []byte("none")); err != nil {
		t.Fatal(err)
	}
	res := c.do(http.MethodGet, "/api/me", nil)
	if res.Status != http.StatusUnauthorized || !strings.Contains(string(res.Body), "expired") {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
}

func TestGarbageSessionCookie(t *testing.T) {
	a := newTestAPI(t)
	req, _ := http.NewRequest(http.MethodGet, a.srv.URL+"/api/me", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: "not-a-real-token"})
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestSignupNameLimitCountsCharacters(t *testing.T) {
	a := newTestAPI(t)
	res := a.client().do(http.MethodPost, "/api/auth/signup", map[string]any{
		"email": "jp@example.com", "password": "password123", "displayName": strings.Repeat("あ", 30),
	})
	if res.Status != http.StatusCreated {
		t.Fatalf("30 Japanese characters rejected: %d %s", res.Status, res.Body)
	}
}
