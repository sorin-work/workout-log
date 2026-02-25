package api

import (
	"net/http"
	"testing"

	"github.com/sorin-work/workout-log/internal/store"
)

func TestNewUsersDefaultToKilogramsAndUTC(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	var me store.User
	c.mustDo(http.MethodGet, "/api/me", nil, http.StatusOK, &me)
	if me.Unit != "kg" || me.Timezone != "UTC" {
		t.Fatalf("defaults: %+v", me)
	}
}

func TestUpdateProfile(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	var me store.User
	c.mustDo(http.MethodPut, "/api/me", map[string]any{
		"displayName": " Sam ", "unit": "lb", "timezone": "America/Chicago",
	}, http.StatusOK, &me)
	if me.DisplayName != "Sam" || me.Unit != "lb" || me.Timezone != "America/Chicago" {
		t.Fatalf("profile: %+v", me)
	}
	c.mustDo(http.MethodGet, "/api/me", nil, http.StatusOK, &me)
	if me.Unit != "lb" {
		t.Fatal("unit not persisted")
	}
}

func TestUpdateProfileValidation(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	res := c.do(http.MethodPut, "/api/me", map[string]any{
		"unit": "stone", "timezone": "Mars/Olympus_Mons",
	})
	if res.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d", res.Status)
	}
	f := errorFields(t, res)
	if f["unit"] == "" || f["timezone"] == "" {
		t.Fatalf("fields: %v", f)
	}
}

func TestChangePassword(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	other := a.client()
	other.mustDo(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "p@example.com", "password": "password123",
	}, http.StatusOK, nil)

	res := c.do(http.MethodPut, "/api/me/password", map[string]any{
		"currentPassword": "nope-nope", "newPassword": "new-password-1",
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["currentPassword"] == "" {
		t.Fatalf("wrong current password: %d %s", res.Status, res.Body)
	}
	res = c.do(http.MethodPut, "/api/me/password", map[string]any{
		"currentPassword": "password123", "newPassword": "short",
	})
	if errorFields(t, res)["newPassword"] == "" {
		t.Fatalf("short password: %d %s", res.Status, res.Body)
	}
	c.mustDo(http.MethodPut, "/api/me/password", map[string]any{
		"currentPassword": "password123", "newPassword": "new-password-1",
	}, http.StatusNoContent, nil)

	// This browser stays logged in, the other one is logged out.
	c.mustDo(http.MethodGet, "/api/me", nil, http.StatusOK, nil)
	other.mustDo(http.MethodGet, "/api/me", nil, http.StatusUnauthorized, nil)
	a.client().mustDo(http.MethodPost, "/api/auth/login", map[string]any{
		"email": "p@example.com", "password": "new-password-1",
	}, http.StatusOK, nil)
}
