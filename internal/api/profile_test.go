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
