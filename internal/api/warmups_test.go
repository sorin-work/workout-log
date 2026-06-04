package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/store"
)

func TestAddWarmups(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("wu@example.com")
	bench := exerciseID(t, c, "Bench Press")
	squat := exerciseID(t, c, "Back Squat")
	w := createWorkout(t, c, "A", time.Now())
	addSet(t, c, w.ID, squat, 5, 100, true)
	addSet(t, c, w.ID, bench, 5, 80, false)
	addSet(t, c, w.ID, bench, 5, 80, false)

	var added []store.Set
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/workouts/%d/warmups", w.ID), map[string]any{"exerciseId": bench},
		http.StatusCreated, &added)
	if len(added) != 4 || added[0].WeightKg != 20 || !added[0].Warmup || added[3].WeightKg != 65 {
		t.Fatalf("warm-ups: %+v", added)
	}
	var got store.Workout
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/workouts/%d", w.ID), nil, http.StatusOK, &got)
	if len(got.Sets) != 7 || got.Sets[0].ExerciseID != squat || !got.Sets[1].Warmup || got.Sets[5].Warmup {
		t.Fatalf("order after insert: %+v", got.Sets)
	}
	for i, s := range got.Sets {
		if s.Position != i+1 {
			t.Fatalf("positions not contiguous: %+v", got.Sets)
		}
	}
	res := c.do(http.MethodPost, fmt.Sprintf("/api/workouts/%d/warmups", w.ID), map[string]any{"exerciseId": exerciseID(t, c, "Deadlift")})
	if res.Status != http.StatusUnprocessableEntity {
		t.Fatalf("no working set: %d", res.Status)
	}
}
