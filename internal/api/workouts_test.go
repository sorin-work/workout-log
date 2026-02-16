package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/store"
)

func createWorkout(t *testing.T, c *client, name string, started time.Time) store.Workout {
	t.Helper()
	var w store.Workout
	c.mustDo(http.MethodPost, "/api/workouts", map[string]any{
		"name": name, "startedAt": started,
	}, http.StatusCreated, &w)
	return w
}

func exerciseID(t *testing.T, c *client, name string) int64 {
	t.Helper()
	for _, e := range listExercises(t, c, "") {
		if e.Name == name {
			return e.ID
		}
	}
	t.Fatalf("exercise %q not found", name)
	return 0
}

func addSet(t *testing.T, c *client, workoutID, exID int64, reps int, kg float64, completed bool) store.Set {
	t.Helper()
	var s store.Set
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/workouts/%d/sets", workoutID), map[string]any{
		"exerciseId": exID, "reps": reps, "weightKg": kg, "completed": completed,
	}, http.StatusCreated, &s)
	return s
}

func TestCreateWorkoutDefaults(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	var w store.Workout
	c.mustDo(http.MethodPost, "/api/workouts", map[string]any{}, http.StatusCreated, &w)
	if w.Name != "Workout" || w.StartedAt.IsZero() || w.FinishedAt != nil {
		t.Fatalf("unexpected defaults: %+v", w)
	}
}

func TestWorkoutValidation(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	start := time.Now().Add(-time.Hour)
	res := c.do(http.MethodPost, "/api/workouts", map[string]any{
		"startedAt": start, "finishedAt": start.Add(-time.Minute),
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["finishedAt"] == "" {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
	res = c.do(http.MethodPost, "/api/workouts", map[string]any{
		"startedAt": time.Now().Add(72 * time.Hour),
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["startedAt"] == "" {
		t.Fatalf("future start accepted: %d %s", res.Status, res.Body)
	}
}

func TestLogSetsAndSummary(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	squat := exerciseID(t, c, "Back Squat")
	bench := exerciseID(t, c, "Bench Press")
	w := createWorkout(t, c, "Day A", time.Now().Add(-time.Hour))

	addSet(t, c, w.ID, squat, 5, 60, true) // warm-up flagged below
	addSet(t, c, w.ID, squat, 5, 100, true)
	addSet(t, c, w.ID, squat, 5, 100, false)
	last := addSet(t, c, w.ID, bench, 8, 70, true)
	if last.Position != 4 || last.ExerciseName != "Bench Press" {
		t.Fatalf("unexpected set: %+v", last)
	}

	var got store.Workout
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/workouts/%d", w.ID), nil, http.StatusOK, &got)
	if len(got.Sets) != 4 || got.Sets[0].Position != 1 {
		t.Fatalf("sets: %+v", got.Sets)
	}

	warm := got.Sets[0]
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/workouts/%d/sets/%d", w.ID, warm.ID), map[string]any{
		"exerciseId": squat, "reps": 5, "weightKg": 60, "completed": true, "warmup": true,
	}, http.StatusOK, nil)

	var page workoutPage
	c.mustDo(http.MethodGet, "/api/workouts", nil, http.StatusOK, &page)
	if len(page.Items) != 1 {
		t.Fatalf("list: %+v", page)
	}
	s := page.Items[0]
	// 5x100 + 8x70; the warm-up and the missed set do not count.
	if s.ExerciseCount != 2 || s.SetCount != 2 || s.VolumeKg != 1060 {
		t.Fatalf("summary: %+v", s)
	}
}

func TestSetValidation(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	w := createWorkout(t, c, "Day A", time.Now())
	res := c.do(http.MethodPost, fmt.Sprintf("/api/workouts/%d/sets", w.ID), map[string]any{
		"exerciseId": 999999, "reps": -3, "weightKg": 50, "rpe": 0,
	})
	if res.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d", res.Status)
	}
	f := errorFields(t, res)
	for _, k := range []string{"exerciseId", "reps", "rpe"} {
		if f[k] == "" {
			t.Errorf("missing error for %s: %v", k, f)
		}
	}
}

func TestCannotUseAnotherUsersCustomExercise(t *testing.T) {
	a := newTestAPI(t)
	alice := a.signup("alice@example.com")
	bob := a.signup("bob@example.com")
	ex := createExercise(t, alice, "Belt Squat", "quads")
	w := createWorkout(t, bob, "Legs", time.Now())
	res := bob.do(http.MethodPost, fmt.Sprintf("/api/workouts/%d/sets", w.ID), map[string]any{
		"exerciseId": ex.ID, "reps": 5, "weightKg": 50,
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["exerciseId"] == "" {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
}

func TestWorkoutsArePrivate(t *testing.T) {
	a := newTestAPI(t)
	alice := a.signup("alice@example.com")
	bob := a.signup("bob@example.com")
	w := createWorkout(t, alice, "Private", time.Now())
	path := fmt.Sprintf("/api/workouts/%d", w.ID)
	bob.mustDo(http.MethodGet, path, nil, http.StatusNotFound, nil)
	bob.mustDo(http.MethodPut, path, map[string]any{"name": "Mine now"}, http.StatusNotFound, nil)
	bob.mustDo(http.MethodDelete, path, nil, http.StatusNotFound, nil)
	bob.mustDo(http.MethodPost, path+"/sets", map[string]any{
		"exerciseId": exerciseID(t, bob, "Deadlift"), "reps": 1, "weightKg": 200,
	}, http.StatusNotFound, nil)
}

func TestDeleteSetAndWorkout(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	w := createWorkout(t, c, "Day B", time.Now())
	set := addSet(t, c, w.ID, exerciseID(t, c, "Deadlift"), 5, 140, true)
	path := fmt.Sprintf("/api/workouts/%d", w.ID)
	c.mustDo(http.MethodDelete, fmt.Sprintf("%s/sets/%d", path, set.ID), nil, http.StatusNoContent, nil)
	c.mustDo(http.MethodDelete, fmt.Sprintf("%s/sets/%d", path, set.ID), nil, http.StatusNotFound, nil)
	c.mustDo(http.MethodDelete, path, nil, http.StatusNoContent, nil)
	c.mustDo(http.MethodGet, path, nil, http.StatusNotFound, nil)
}

func TestUpdateWorkout(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	start := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	w := createWorkout(t, c, "Day A", start)
	var got store.Workout
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/workouts/%d", w.ID), map[string]any{
		"name": "  Heavy day ", "notes": "felt strong", "startedAt": start,
		"finishedAt": start.Add(70 * time.Minute),
	}, http.StatusOK, &got)
	if got.Name != "Heavy day" || got.Notes != "felt strong" || got.FinishedAt == nil {
		t.Fatalf("update not applied: %+v", got)
	}
}

func TestFinishWorkout(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	w := createWorkout(t, c, "Day A", time.Now().Add(-45*time.Minute))
	var res finishResponse
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/workouts/%d/finish", w.ID), nil, http.StatusOK, &res)
	got := res.Workout
	if len(res.Progression) != 0 {
		t.Fatalf("a workout without a template should not progress anything: %+v", res.Progression)
	}
	if got.FinishedAt == nil || got.FinishedAt.Before(got.StartedAt) {
		t.Fatalf("finishedAt = %v", got.FinishedAt)
	}
	other := a.signup("other@example.com")
	other.mustDo(http.MethodPost, fmt.Sprintf("/api/workouts/%d/finish", w.ID), nil, http.StatusNotFound, nil)
}

func TestWorkoutHistoryPages(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	base := time.Now().Add(-30 * 24 * time.Hour)
	for i := 0; i < 4; i++ {
		createWorkout(t, c, fmt.Sprintf("Workout %d", i+1), base.Add(time.Duration(i)*24*time.Hour))
	}
	var p1, p2 workoutPage
	c.mustDo(http.MethodGet, "/api/workouts?pageSize=2", nil, http.StatusOK, &p1)
	c.mustDo(http.MethodGet, "/api/workouts?pageSize=2&page=2", nil, http.StatusOK, &p2)
	if p1.Total != 4 || p1.TotalPages != 2 {
		t.Fatalf("page 1 meta: %+v", p1)
	}
	if p1.Items[0].Name != "Workout 4" || p1.Items[1].Name != "Workout 3" {
		t.Fatalf("page 1 not newest first: %+v", p1.Items)
	}
	if len(p2.Items) != 2 || p2.Items[1].Name != "Workout 1" {
		t.Fatalf("page 2: %+v", p2.Items)
	}
	c.mustDo(http.MethodGet, "/api/workouts?page=0", nil, http.StatusBadRequest, nil)
	c.mustDo(http.MethodGet, "/api/workouts?pageSize=500", nil, http.StatusBadRequest, nil)
}

func TestRPEAboveTenIsRejected(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("w@example.com")
	w := createWorkout(t, c, "Day A", time.Now())
	res := c.do(http.MethodPost, fmt.Sprintf("/api/workouts/%d/sets", w.ID), map[string]any{
		"exerciseId": exerciseID(t, c, "Deadlift"), "reps": 3, "weightKg": 180, "rpe": 95, "completed": true,
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["rpe"] == "" {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
}
