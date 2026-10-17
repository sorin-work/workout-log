package domain

import "slices"

// ProgressionRule describes linear progression for one exercise in a
// template: add IncrementKg after a successful session, and after
// DeloadAfter failed sessions in a row drop the weight by DeloadPercent.
type ProgressionRule struct {
	IncrementKg   float64
	DeloadAfter   int
	DeloadPercent float64
	// RoundingKg is the smallest weight jump available, usually 2.5 kg.
	RoundingKg float64
}

// ProgressionState is what a template remembers between sessions.
type ProgressionState struct {
	WeightKg float64
	Failures int
}

// Outcome says what a finished session did to the target weight.
type Outcome string

const (
	OutcomeIncreased Outcome = "increased"
	OutcomeKept      Outcome = "kept"
	OutcomeDeloaded  Outcome = "deloaded"
	OutcomeSkipped   Outcome = "skipped"
)

// Target is the planned work for an exercise in a template.
type Target struct {
	Sets int
	Reps int
}

// SessionSucceeded reports whether the logged sets hit the target: at least
// target.Sets completed working sets with target.Reps or more at the planned
// weight or heavier.
func SessionSucceeded(target Target, weightKg float64, sets []SetInput) bool {
	hit := 0
	for _, s := range sets {
		if !s.Working() {
			continue
		}
		if s.Reps >= target.Reps && s.WeightKg >= weightKg-1e-9 {
			hit++
		}
	}
	return hit >= target.Sets
}

// Progress applies the rule to a finished session and returns the new state.
// A session without any working sets for the exercise leaves the state alone.
func Progress(rule ProgressionRule, state ProgressionState, target Target, sets []SetInput) (ProgressionState, Outcome) {
	if !slices.ContainsFunc(sets, SetInput.Working) {
		return state, OutcomeSkipped
	}
	if SessionSucceeded(target, state.WeightKg, sets) {
		if rule.IncrementKg <= 0 {
			return ProgressionState{WeightKg: state.WeightKg}, OutcomeKept
		}
		return ProgressionState{WeightKg: state.WeightKg + rule.IncrementKg}, OutcomeIncreased
	}
	failures := state.Failures + 1
	if rule.DeloadAfter > 0 && failures >= rule.DeloadAfter && rule.DeloadPercent > 0 {
		return ProgressionState{WeightKg: deload(state.WeightKg, rule)}, OutcomeDeloaded
	}
	return ProgressionState{WeightKg: state.WeightKg, Failures: failures}, OutcomeKept
}

func deload(weight float64, rule ProgressionRule) float64 {
	w := weight * (1 - rule.DeloadPercent/100)
	if rule.RoundingKg > 0 {
		w = RoundTo(w, rule.RoundingKg)
	}
	if w < 0 {
		return 0
	}
	return w
}
