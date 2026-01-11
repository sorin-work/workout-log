package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sorin-work/workout-log/internal/store"
)

func templateBody(name string, rows ...map[string]any) map[string]any {
	if rows == nil {
		rows = []map[string]any{}
	}
	return map[string]any{"name": name, "exercises": rows}
}

func row(exID int64, sets, reps int, kg float64) map[string]any {
	return map[string]any{"exerciseId": exID, "sets": sets, "reps": reps, "weightKg": kg}
}

func createTemplate(t *testing.T, c *client, body map[string]any) store.Template {
	t.Helper()
	var tpl store.Template
	c.mustDo(http.MethodPost, "/api/templates", body, http.StatusCreated, &tpl)
	return tpl
}

func TestCreateTemplate(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("t@example.com")
	squat := exerciseID(t, c, "Back Squat")
	bench := exerciseID(t, c, "Bench Press")
	tpl := createTemplate(t, c, templateBody("Day A", row(squat, 5, 5, 100), row(bench, 3, 8, 70)))
	if tpl.Name != "Day A" || len(tpl.Exercises) != 2 {
		t.Fatalf("template: %+v", tpl)
	}
	first := tpl.Exercises[0]
	if first.ExerciseName != "Back Squat" || first.Position != 1 || first.RestSeconds != 120 {
		t.Fatalf("first exercise: %+v", first)
	}

	var list []store.Template
	c.mustDo(http.MethodGet, "/api/templates", nil, http.StatusOK, &list)
	if len(list) != 1 || len(list[0].Exercises) != 2 {
		t.Fatalf("list: %+v", list)
	}
}

func TestUpdateTemplateKeepsAddsAndRemovesRows(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("t@example.com")
	squat := exerciseID(t, c, "Back Squat")
	bench := exerciseID(t, c, "Bench Press")
	row1 := exerciseID(t, c, "Barbell Row")
	tpl := createTemplate(t, c, templateBody("Day A", row(squat, 5, 5, 100), row(bench, 3, 8, 70)))

	keep := row(squat, 5, 5, 102.5)
	keep["id"] = tpl.Exercises[0].ID
	var updated store.Template
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/templates/%d", tpl.ID),
		templateBody("Day A (heavy)", row(row1, 3, 5, 60), keep), http.StatusOK, &updated)

	if len(updated.Exercises) != 2 {
		t.Fatalf("exercises: %+v", updated.Exercises)
	}
	if updated.Exercises[0].ExerciseName != "Barbell Row" || updated.Exercises[1].ID != tpl.Exercises[0].ID {
		t.Fatalf("order or ids wrong: %+v", updated.Exercises)
	}
	if updated.Exercises[1].WeightKg != 102.5 {
		t.Fatalf("weight not updated: %+v", updated.Exercises[1])
	}
}

func TestTemplateValidation(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("t@example.com")
	res := c.do(http.MethodPost, "/api/templates", templateBody("", row(424242, 0, 5, 50)))
	if res.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d", res.Status)
	}
	f := errorFields(t, res)
	if f["name"] == "" || f["exercises.0.sets"] == "" || f["exercises.0.exerciseId"] == "" {
		t.Fatalf("fields: %v", f)
	}
}

func TestTemplatesArePrivate(t *testing.T) {
	a := newTestAPI(t)
	alice := a.signup("alice@example.com")
	bob := a.signup("bob@example.com")
	tpl := createTemplate(t, alice, templateBody("Alice A", row(exerciseID(t, alice, "Deadlift"), 1, 5, 140)))
	path := fmt.Sprintf("/api/templates/%d", tpl.ID)
	bob.mustDo(http.MethodGet, path, nil, http.StatusNotFound, nil)
	bob.mustDo(http.MethodPut, path, templateBody("Mine"), http.StatusNotFound, nil)
	bob.mustDo(http.MethodDelete, path, nil, http.StatusNotFound, nil)

	// Bob cannot attach Alice's template rows to his own template either.
	own := createTemplate(t, bob, templateBody("Bob A"))
	stolen := row(exerciseID(t, bob, "Deadlift"), 1, 5, 140)
	stolen["id"] = tpl.Exercises[0].ID
	bob.mustDo(http.MethodPut, fmt.Sprintf("/api/templates/%d", own.ID), templateBody("Bob A", stolen), http.StatusNotFound, nil)
}

func TestDeleteTemplate(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("t@example.com")
	tpl := createTemplate(t, c, templateBody("Temp"))
	c.mustDo(http.MethodDelete, fmt.Sprintf("/api/templates/%d", tpl.ID), nil, http.StatusNoContent, nil)
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/templates/%d", tpl.ID), nil, http.StatusNotFound, nil)
}

func TestStartWorkoutFromTemplate(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("t@example.com")
	squat := exerciseID(t, c, "Back Squat")
	bench := exerciseID(t, c, "Bench Press")
	tpl := createTemplate(t, c, templateBody("Day A", row(squat, 3, 5, 100), row(bench, 2, 8, 70)))

	var w store.Workout
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/templates/%d/start", tpl.ID), nil, http.StatusCreated, &w)
	if w.Name != "Day A" || w.TemplateID == nil || *w.TemplateID != tpl.ID {
		t.Fatalf("workout: %+v", w)
	}
	if len(w.Sets) != 5 {
		t.Fatalf("expected 5 planned sets, got %d", len(w.Sets))
	}
	if s := w.Sets[0]; s.ExerciseID != squat || s.Reps != 5 || s.WeightKg != 100 || s.Completed {
		t.Fatalf("first set: %+v", s)
	}
	if s := w.Sets[4]; s.ExerciseID != bench || s.Position != 5 {
		t.Fatalf("last set: %+v", s)
	}

	other := a.signup("other@example.com")
	other.mustDo(http.MethodPost, fmt.Sprintf("/api/templates/%d/start", tpl.ID), nil, http.StatusNotFound, nil)
}
