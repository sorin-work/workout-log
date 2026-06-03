package domain

// WarmupSet is a suggested warm-up before the working sets.
type WarmupSet struct {
	Reps     int     `json:"reps"`
	WeightKg float64 `json:"weightKg"`
}

// warmupSteps are the usual ramp: bar, then 40%, 60% and 80% of the work
// weight with fewer reps as it gets heavier.
var warmupSteps = []struct {
	share float64
	reps  int
}{{0.4, 5}, {0.6, 3}, {0.8, 2}}

// WarmupSets suggests warm-ups for a working weight on a barbell. Steps that
// round to the bar or to the working weight itself are left out, so light
// work weights get fewer warm-ups.
func WarmupSets(workKg, barKg, stepKg float64) []WarmupSet {
	out := []WarmupSet{}
	if workKg <= barKg {
		return out
	}
	out = append(out, WarmupSet{Reps: 10, WeightKg: barKg})
	last := barKg
	for _, s := range warmupSteps {
		w := RoundTo(workKg*s.share, stepKg)
		if w <= last || w >= workKg {
			continue
		}
		out = append(out, WarmupSet{Reps: s.reps, WeightKg: w})
		last = w
	}
	return out
}
