package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMalformedIDsAreNotFound(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("e@example.com")
	for _, p := range []string{"/api/workouts/abc", "/api/workouts/-1", "/api/templates/0", "/api/exercises/1.5"} {
		c.mustDo(http.MethodGet, p, nil, http.StatusNotFound, nil)
	}
}

func TestUnknownFieldsAreRejected(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("e@example.com")
	res := c.do(http.MethodPost, "/api/workouts", map[string]any{"name": "A", "userId": 1})
	if res.Status != http.StatusBadRequest || !strings.Contains(string(res.Body), "userId") {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
}

func TestWorkoutNotesLimit(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("e@example.com")
	res := c.do(http.MethodPost, "/api/workouts", map[string]any{"notes": strings.Repeat("x", 4001)})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["notes"] == "" {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
}

func TestArchivedExercisesKeepTheirHistory(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("e@example.com")
	ex := createExercise(t, c, "Old Machine", "back")
	w := createWorkout(t, c, "Back", time.Now().Add(-time.Hour))
	addSet(t, c, w.ID, ex.ID, 10, 50, true)
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/exercises/%d/archive", ex.ID), nil, http.StatusOK, nil)

	var got struct {
		Sets []struct {
			ExerciseName string `json:"exerciseName"`
		} `json:"sets"`
	}
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/workouts/%d", w.ID), nil, http.StatusOK, &got)
	if len(got.Sets) != 1 || got.Sets[0].ExerciseName != "Old Machine" {
		t.Fatalf("sets: %+v", got.Sets)
	}
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/exercises/%d/records", ex.ID), nil, http.StatusOK, nil)
}

func TestLargeBodiesAreRejected(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("e@example.com")
	res := c.do(http.MethodPost, "/api/workouts", map[string]any{"notes": strings.Repeat("x", 2<<20)})
	if res.Status != http.StatusBadRequest || !strings.Contains(string(res.Body), "too large") {
		t.Fatalf("status %d: %.200s", res.Status, res.Body)
	}
}
