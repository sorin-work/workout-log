package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/store"
)

func listExercises(t *testing.T, c *client, query string) []store.Exercise {
	t.Helper()
	var list []store.Exercise
	c.mustDo(http.MethodGet, "/api/exercises"+query, nil, http.StatusOK, &list)
	return list
}

func names(list []store.Exercise) map[string]bool {
	m := map[string]bool{}
	for _, e := range list {
		m[e.Name] = true
	}
	return m
}

func createExercise(t *testing.T, c *client, name, muscle string) store.Exercise {
	t.Helper()
	var e store.Exercise
	c.mustDo(http.MethodPost, "/api/exercises", map[string]any{
		"name": name, "primaryMuscle": muscle, "equipment": "machine",
	}, http.StatusCreated, &e)
	return e
}

func TestListBuiltinExercises(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	list := listExercises(t, c, "")
	if len(list) < 40 {
		t.Fatalf("got %d built-in exercises", len(list))
	}
	for i := 1; i < len(list); i++ {
		if strings.ToLower(list[i-1].Name) > strings.ToLower(list[i].Name) {
			t.Fatalf("not sorted: %q before %q", list[i-1].Name, list[i].Name)
		}
	}
}

func TestSearchExercises(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	got := names(listExercises(t, c, "?q=squat"))
	for _, want := range []string{"Back Squat", "Front Squat", "Goblet Squat", "Bulgarian Split Squat"} {
		if !got[want] {
			t.Errorf("search for squat is missing %s", want)
		}
	}
	if got["Deadlift"] {
		t.Error("search for squat returned Deadlift")
	}
	if n := len(listExercises(t, c, "?q=100%25")); n != 0 {
		t.Errorf("percent sign should be matched literally, got %d results", n)
	}
}

func TestFilterExercisesByMuscleIncludesSecondary(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	got := names(listExercises(t, c, "?muscle=triceps"))
	if !got["Triceps Pushdown"] || !got["Bench Press"] {
		t.Fatalf("triceps filter missing primary or secondary exercises: %v", got)
	}
	c.mustDo(http.MethodGet, "/api/exercises?muscle=wings", nil, http.StatusBadRequest, nil)
}

func TestCustomExercisesArePrivate(t *testing.T) {
	a := newTestAPI(t)
	alice := a.signup("alice@example.com")
	bob := a.signup("bob@example.com")
	e := createExercise(t, alice, "Belt Squat", "quads")
	if !e.Custom {
		t.Fatal("custom flag not set")
	}
	if names(listExercises(t, bob, ""))["Belt Squat"] {
		t.Fatal("bob can see alice's exercise")
	}
	bob.mustDo(http.MethodGet, fmt.Sprintf("/api/exercises/%d", e.ID), nil, http.StatusNotFound, nil)
	// Bob may use the same name for his own exercise.
	createExercise(t, bob, "Belt Squat", "quads")
}

func TestCreateExerciseValidation(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	res := c.do(http.MethodPost, "/api/exercises", map[string]any{
		"name": "   ", "primaryMuscle": "wings", "equipment": "spaceship",
		"secondaryMuscles": []string{"elbows"},
	})
	if res.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d", res.Status)
	}
	f := errorFields(t, res)
	for _, k := range []string{"name", "primaryMuscle", "equipment", "secondaryMuscles"} {
		if f[k] == "" {
			t.Errorf("missing error for %s: %v", k, f)
		}
	}
	res = c.do(http.MethodPost, "/api/exercises", map[string]any{
		"name": "bench press", "primaryMuscle": "chest", "equipment": "barbell",
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["name"] == "" {
		t.Fatalf("built-in name should conflict: %d %s", res.Status, res.Body)
	}
}

func TestUpdateAndArchiveExercise(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	e := createExercise(t, c, "Pec Deck", "chest")

	var updated store.Exercise
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/exercises/%d", e.ID), map[string]any{
		"name": "Pec  Deck Fly", "primaryMuscle": "chest", "equipment": "machine",
		"secondaryMuscles": []string{"shoulders", "chest", "shoulders"},
	}, http.StatusOK, &updated)
	if updated.Name != "Pec Deck Fly" || len(updated.SecondaryMuscles) != 1 {
		t.Fatalf("update not applied: %+v", updated)
	}

	c.mustDo(http.MethodPost, fmt.Sprintf("/api/exercises/%d/archive", e.ID), nil, http.StatusOK, nil)
	if names(listExercises(t, c, ""))["Pec Deck Fly"] {
		t.Fatal("archived exercise still listed")
	}
	if !names(listExercises(t, c, "?archived=true"))["Pec Deck Fly"] {
		t.Fatal("archived exercise missing with archived=true")
	}
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/exercises/%d/restore", e.ID), nil, http.StatusOK, nil)
	if !names(listExercises(t, c, ""))["Pec Deck Fly"] {
		t.Fatal("restored exercise not listed")
	}
}

func TestBuiltinExercisesAreReadOnly(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	bench := listExercises(t, c, "?q=Bench%20Press")[0]
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/exercises/%d", bench.ID), map[string]any{
		"name": "My Bench", "primaryMuscle": "chest", "equipment": "barbell",
	}, http.StatusForbidden, nil)
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/exercises/%d/archive", bench.ID), nil, http.StatusForbidden, nil)
}

func TestExerciseNameLimitCountsCharacters(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	for _, name := range []string{
		"Kniebeuge mit Sicherheitsstange (hoch)", // 38 characters
		"Développé couché prise serrée haltères",
		"ダンベルベンチプレス（インクライン）",
	} {
		if n := len([]rune(name)); n > 40 || len(name) <= 40 && name[0] != 'K' {
			t.Fatalf("bad fixture %q: %d characters, %d bytes", name, n, len(name))
		}
		createExercise(t, c, name, "chest")
	}
	res := c.do(http.MethodPost, "/api/exercises", map[string]any{
		"name": strings.Repeat("é", 41), "primaryMuscle": "chest", "equipment": "machine",
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["name"] == "" {
		t.Fatalf("41 characters accepted: %d %s", res.Status, res.Body)
	}
}

func TestExerciseNotes(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	bench := exerciseID(t, c, "Bench Press")
	path := fmt.Sprintf("/api/exercises/%d/notes", bench)
	var n exerciseNotes
	c.mustDo(http.MethodGet, path, nil, http.StatusOK, &n)
	if n.Notes != "" {
		t.Fatalf("notes = %q", n.Notes)
	}
	c.mustDo(http.MethodPut, path, map[string]any{"notes": " Feet flat, pinkies on the rings "}, http.StatusOK, nil)
	c.mustDo(http.MethodGet, path, nil, http.StatusOK, &n)
	if n.Notes != "Feet flat, pinkies on the rings" {
		t.Fatalf("notes = %q", n.Notes)
	}
	// Notes are personal even for built-in exercises.
	var other exerciseNotes
	a.signup("other@example.com").mustDo(http.MethodGet, path, nil, http.StatusOK, &other)
	if other.Notes != "" {
		t.Fatalf("other user sees %q", other.Notes)
	}
	c.mustDo(http.MethodPut, path, map[string]any{"notes": ""}, http.StatusOK, nil)
	c.mustDo(http.MethodGet, path, nil, http.StatusOK, &n)
	if n.Notes != "" {
		t.Fatalf("notes not cleared: %q", n.Notes)
	}
}

func TestExercisesSortedByUse(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("ex@example.com")
	w := createWorkout(t, c, "A", time.Now().Add(-time.Hour))
	curl := exerciseID(t, c, "Hammer Curl")
	for i := 0; i < 3; i++ {
		addSet(t, c, w.ID, curl, 10, 14, true)
	}
	addSet(t, c, w.ID, exerciseID(t, c, "Plank"), 1, 0, true)
	list := listExercises(t, c, "?sort=popular")
	if list[0].Name != "Hammer Curl" || list[1].Name != "Plank" {
		t.Fatalf("first two: %s, %s", list[0].Name, list[1].Name)
	}
	if list[2].Name > list[3].Name {
		t.Fatalf("the rest should stay alphabetical: %s, %s", list[2].Name, list[3].Name)
	}
}
