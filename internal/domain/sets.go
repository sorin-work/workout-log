package domain

import "math"

// Limits for a single logged set.
const (
	MaxReps     = 100
	MaxWeightKg = 1000
	MinRPE      = 1
	MaxRPE      = 10
)

// SetInput is a set as entered by the user, with the weight in kilograms.
type SetInput struct {
	Reps      int
	WeightKg  float64
	RPE       *float64
	Completed bool
	Warmup    bool
}

// ValidateSet checks a set and returns a message per invalid field.
func ValidateSet(s SetInput) map[string]string {
	fields := map[string]string{}
	if s.Reps < 0 || s.Reps > MaxReps {
		fields["reps"] = "Reps must be between 0 and 100."
	}
	if s.WeightKg < 0 || s.WeightKg > MaxWeightKg || math.IsNaN(s.WeightKg) {
		fields["weightKg"] = "Weight must be between 0 and 1000 kg."
	}
	if s.RPE != nil {
		r := *s.RPE
		switch {
		case r < MinRPE:
			fields["rpe"] = "RPE must be between 1 and 10."
		case math.Mod(r*2, 1) != 0:
			fields["rpe"] = "RPE goes in half steps, e.g. 7.5."
		}
	}
	if s.Completed && s.Reps == 0 {
		fields["reps"] = "A completed set needs at least one rep."
	}
	return fields
}

// SetVolume is the tonnage of a set: reps times weight. Warm-ups and sets
// that were not completed do not count.
func SetVolume(s SetInput) float64 {
	if s.Warmup || !s.Completed {
		return 0
	}
	return float64(s.Reps) * s.WeightKg
}
