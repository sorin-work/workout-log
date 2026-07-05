package api

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

// progressionTemplate creates a 3x5 squat template that adds 2.5 kg per
// successful session and deloads 10% after two failures, plus a bench row
// without progression.
func progressionTemplate(t *testing.T, c *client, squatKg float64) store.Template {
	t.Helper()
	squat := row(exerciseID(t, c, "Back Squat"), 3, 5, squatKg)
	squat["incrementKg"] = 2.5
	squat["deloadAfter"] = 2
	squat["deloadPercent"] = 10
	bench := row(exerciseID(t, c, "Bench Press"), 3, 5, 60)
	return createTemplate(t, c, templateBody("A", squat, bench))
}

// runSession starts the template, marks the squat sets as done with the given
// reps (one entry per set) and finishes the workout.
func runSession(t *testing.T, c *client, tpl store.Template, squatReps ...int) finishResponse {
	t.Helper()
	var w store.Workout
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/templates/%d/start", tpl.ID), nil, http.StatusCreated, &w)
	i := 0
	for _, s := range w.Sets {
		if s.ExerciseID != tpl.Exercises[0].ExerciseID || i >= len(squatReps) {
			continue
		}
		c.mustDo(http.MethodPut, fmt.Sprintf("/api/workouts/%d/sets/%d", w.ID, s.ID), map[string]any{
			"exerciseId": s.ExerciseID, "reps": squatReps[i], "weightKg": s.WeightKg, "completed": true,
		}, http.StatusOK, nil)
		i++
	}
	var res finishResponse
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/workouts/%d/finish", w.ID), nil, http.StatusOK, &res)
	return res
}

func squatTarget(t *testing.T, c *client, tpl store.Template) store.TemplateExercise {
	t.Helper()
	var got store.Template
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/templates/%d", tpl.ID), nil, http.StatusOK, &got)
	return got.Exercises[0]
}

func TestFinishingASuccessfulSessionAddsWeight(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	tpl := progressionTemplate(t, c, 100)
	res := runSession(t, c, tpl, 5, 5, 5)
	if len(res.Progression) != 1 {
		t.Fatalf("bench was not trained and should be skipped: %+v", res.Progression)
	}
	ch := res.Progression[0]
	if ch.Outcome != domain.OutcomeIncreased || ch.PreviousWeightKg != 100 || ch.WeightKg != 102.5 {
		t.Fatalf("change: %+v", ch)
	}
	if got := squatTarget(t, c, tpl); got.WeightKg != 102.5 {
		t.Fatalf("template weight = %v", got.WeightKg)
	}
}

func TestMissedRepsCountAsAFailure(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	tpl := progressionTemplate(t, c, 100)
	res := runSession(t, c, tpl, 5, 5, 3)
	if ch := res.Progression[0]; ch.Outcome != domain.OutcomeKept || ch.Failures != 1 {
		t.Fatalf("change: %+v", ch)
	}
	if got := squatTarget(t, c, tpl); got.WeightKg != 100 || got.Failures != 1 {
		t.Fatalf("target: %+v", got)
	}
}

func TestDeloadAfterRepeatedFailures(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	tpl := progressionTemplate(t, c, 100)
	runSession(t, c, tpl, 5, 4, 4)
	res := runSession(t, c, tpl, 5, 3)
	if ch := res.Progression[0]; ch.Outcome != domain.OutcomeDeloaded || ch.WeightKg != 90 {
		t.Fatalf("change: %+v", ch)
	}
	if got := squatTarget(t, c, tpl); got.WeightKg != 90 || got.Failures != 0 {
		t.Fatalf("target: %+v", got)
	}
}

func TestEditingTheWeightResetsFailures(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	tpl := progressionTemplate(t, c, 100)
	runSession(t, c, tpl, 2)
	cur := squatTarget(t, c, tpl)
	if cur.Failures != 1 {
		t.Fatalf("failures = %d", cur.Failures)
	}
	edited := row(cur.ExerciseID, 3, 5, 95)
	edited["id"] = cur.ID
	edited["incrementKg"] = 2.5
	c.mustDo(http.MethodPut, fmt.Sprintf("/api/templates/%d", tpl.ID), templateBody("A", edited), http.StatusOK, nil)
	if got := squatTarget(t, c, tpl); got.Failures != 0 || got.WeightKg != 95 {
		t.Fatalf("target: %+v", got)
	}
}

func TestFinishingTwiceProgressesOnce(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	tpl := progressionTemplate(t, c, 100)
	res := runSession(t, c, tpl, 5, 5, 5)
	path := fmt.Sprintf("/api/workouts/%d/finish", res.Workout.ID)

	var again finishResponse
	c.mustDo(http.MethodPost, path, nil, http.StatusOK, &again)
	if len(again.Progression) != 0 || again.Workout.FinishedAt == nil {
		t.Fatalf("second finish: %+v", again)
	}
	if got := squatTarget(t, c, tpl); got.WeightKg != 102.5 {
		t.Fatalf("weight after finishing twice = %v, want 102.5", got.WeightKg)
	}
}

func TestConcurrentFinishProgressesOnce(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("p@example.com")
	tpl := progressionTemplate(t, c, 100)
	var w store.Workout
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/templates/%d/start", tpl.ID), nil, http.StatusCreated, &w)
	for _, s := range w.Sets {
		c.mustDo(http.MethodPut, fmt.Sprintf("/api/workouts/%d/sets/%d", w.ID, s.ID), map[string]any{
			"exerciseId": s.ExerciseID, "reps": s.Reps, "weightKg": s.WeightKg, "completed": true,
		}, http.StatusOK, nil)
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.do(http.MethodPost, fmt.Sprintf("/api/workouts/%d/finish", w.ID), nil)
		}()
	}
	wg.Wait()
	if got := squatTarget(t, c, tpl); got.WeightKg != 102.5 {
		t.Fatalf("weight after five simultaneous finishes = %v, want 102.5", got.WeightKg)
	}
}
