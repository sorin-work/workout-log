package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sorin-work/workout-log/internal/store"
)

func abProgram(t *testing.T, c *client) (store.Program, store.Template, store.Template) {
	t.Helper()
	a := createTemplate(t, c, templateBody("Day A", row(exerciseID(t, c, "Back Squat"), 3, 5, 100)))
	b := createTemplate(t, c, templateBody("Day B", row(exerciseID(t, c, "Deadlift"), 1, 5, 140)))
	var p store.Program
	c.mustDo(http.MethodPost, "/api/programs", map[string]any{
		"name": "A/B", "templateIds": []int64{a.ID, b.ID},
	}, http.StatusCreated, &p)
	return p, a, b
}

func TestProgramRotation(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("prog@example.com")
	p, _, _ := abProgram(t, c)
	if len(p.Days) != 2 || p.Days[0].TemplateName != "Day A" || p.Active {
		t.Fatalf("program: %+v", p)
	}
	want := []string{"Day A", "Day B", "Day A"}
	for i, name := range want {
		var w store.Workout
		c.mustDo(http.MethodPost, fmt.Sprintf("/api/programs/%d/start", p.ID), nil, http.StatusCreated, &w)
		if w.Name != name {
			t.Fatalf("workout %d = %q, want %q", i+1, w.Name, name)
		}
	}
}

func TestOnlyOneActiveProgram(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("prog@example.com")
	p1, ta, tb := abProgram(t, c)
	var p2 store.Program
	c.mustDo(http.MethodPost, "/api/programs", map[string]any{
		"name": "B only", "templateIds": []int64{tb.ID, ta.ID},
	}, http.StatusCreated, &p2)
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/programs/%d/activate", p1.ID), nil, http.StatusOK, nil)
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/programs/%d/activate", p2.ID), nil, http.StatusOK, nil)

	var list []store.Program
	c.mustDo(http.MethodGet, "/api/programs", nil, http.StatusOK, &list)
	if len(list) != 2 || list[0].ID != p2.ID || !list[0].Active || list[1].Active {
		t.Fatalf("programs: %+v", list)
	}
	if len(list[0].Days) != 2 || list[0].Days[0].TemplateName != "Day B" {
		t.Fatalf("days not loaded: %+v", list[0])
	}
}

func TestShrinkingAProgramKeepsNextDayInRange(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("prog@example.com")
	p, ta, _ := abProgram(t, c)
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/programs/%d/start", p.ID), nil, http.StatusCreated, nil)
	var updated store.Program
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/programs/%d", p.ID), map[string]any{
		"name": "A only", "templateIds": []int64{ta.ID},
	}, http.StatusOK, &updated)
	if updated.NextPosition != 0 || len(updated.Days) != 1 {
		t.Fatalf("program: %+v", updated)
	}
}

func TestProgramValidationAndPrivacy(t *testing.T) {
	a := newTestAPI(t)
	alice := a.signup("alice@example.com")
	bob := a.signup("bob@example.com")
	p, ta, _ := abProgram(t, alice)

	res := bob.do(http.MethodPost, "/api/programs", map[string]any{
		"name": "Stolen", "templateIds": []int64{ta.ID},
	})
	if res.Status != http.StatusUnprocessableEntity || errorFields(t, res)["templateIds"] == "" {
		t.Fatalf("status %d: %s", res.Status, res.Body)
	}
	res = bob.do(http.MethodPost, "/api/programs", map[string]any{"name": "", "templateIds": []int64{}})
	if f := errorFields(t, res); f["name"] == "" || f["templateIds"] == "" {
		t.Fatalf("fields: %v", f)
	}
	path := fmt.Sprintf("/api/programs/%d", p.ID)
	bob.mustDo(http.MethodGet, path, nil, http.StatusNotFound, nil)
	bob.mustDo(http.MethodPost, path+"/start", nil, http.StatusNotFound, nil)
	bob.mustDo(http.MethodPost, path+"/activate", nil, http.StatusNotFound, nil)
	bob.mustDo(http.MethodDelete, path, nil, http.StatusNotFound, nil)
	alice.mustDo(http.MethodDelete, path, nil, http.StatusNoContent, nil)
}
