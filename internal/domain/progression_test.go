package domain

import "testing"

var fiveByFive = Target{Sets: 5, Reps: 5}

var standardRule = ProgressionRule{IncrementKg: 2.5, DeloadAfter: 3, DeloadPercent: 10, RoundingKg: 2.5}

func sets(n, reps int, kg float64) []SetInput {
	out := make([]SetInput, n)
	for i := range out {
		out[i] = SetInput{Reps: reps, WeightKg: kg, Completed: true}
	}
	return out
}

func TestProgressIncreasesAfterAllSetsCompleted(t *testing.T) {
	next, out := Progress(standardRule, ProgressionState{WeightKg: 100}, fiveByFive, sets(5, 5, 100))
	if out != OutcomeIncreased || next.WeightKg != 102.5 || next.Failures != 0 {
		t.Fatalf("got %v %+v", out, next)
	}
}

func TestProgressCountsFailures(t *testing.T) {
	logged := append(sets(4, 5, 100), SetInput{Reps: 3, WeightKg: 100, Completed: true})
	next, out := Progress(standardRule, ProgressionState{WeightKg: 100, Failures: 1}, fiveByFive, logged)
	if out != OutcomeKept || next.WeightKg != 100 || next.Failures != 2 {
		t.Fatalf("got %v %+v", out, next)
	}
}

func TestProgressDeloadsAfterRepeatedFailures(t *testing.T) {
	next, out := Progress(standardRule, ProgressionState{WeightKg: 100, Failures: 2}, fiveByFive, sets(3, 5, 100))
	if out != OutcomeDeloaded || next.WeightKg != 90 || next.Failures != 0 {
		t.Fatalf("got %v %+v", out, next)
	}
	next, _ = Progress(standardRule, ProgressionState{WeightKg: 50, Failures: 2}, fiveByFive, sets(1, 5, 50))
	if next.WeightKg != 45 {
		t.Fatalf("deload from 50 = %v, want 45", next.WeightKg)
	}
}

func TestProgressIgnoresWarmupsAndLighterSets(t *testing.T) {
	logged := sets(5, 5, 100)
	logged[0].Warmup = true
	if SessionSucceeded(fiveByFive, 100, logged) {
		t.Fatal("warm-up counted as a working set")
	}
	if SessionSucceeded(fiveByFive, 100, sets(5, 5, 95)) {
		t.Fatal("lighter sets counted as hitting the target")
	}
	if !SessionSucceeded(fiveByFive, 100, sets(6, 6, 100)) {
		t.Fatal("extra sets and reps should still count as success")
	}
}

func TestProgressSkipsExercisesNotTrained(t *testing.T) {
	state := ProgressionState{WeightKg: 100, Failures: 1}
	planned := sets(5, 5, 100)
	for i := range planned {
		planned[i].Completed = false
	}
	next, out := Progress(standardRule, state, fiveByFive, planned)
	if out != OutcomeSkipped || next != state {
		t.Fatalf("got %v %+v", out, next)
	}
}

func TestProgressWithoutIncrementOrDeload(t *testing.T) {
	rule := ProgressionRule{}
	next, out := Progress(rule, ProgressionState{WeightKg: 20}, fiveByFive, sets(5, 5, 20))
	if out != OutcomeKept || next.WeightKg != 20 {
		t.Fatalf("got %v %+v", out, next)
	}
	next, out = Progress(rule, ProgressionState{WeightKg: 20, Failures: 10}, fiveByFive, sets(1, 5, 20))
	if out != OutcomeKept || next.Failures != 11 {
		t.Fatalf("got %v %+v", out, next)
	}
}
