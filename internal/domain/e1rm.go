package domain

import "math"

// MaxRepsForEstimate is the highest rep count we trust for estimating a
// one-rep max. Beyond it the formulas drift too far to be useful.
const MaxRepsForEstimate = 12

// Formula names an estimated-1RM formula.
type Formula string

const (
	Epley   Formula = "epley"
	Brzycki Formula = "brzycki"
)

// EstimateOneRepMax estimates the one-rep max for a set of reps at weight.
// It returns 0 when the set cannot be used for an estimate (no weight, no
// reps or more than MaxRepsForEstimate reps). A single rep is the 1RM itself.
// The result is rounded to 0.1.
func EstimateOneRepMax(weight float64, reps int, f Formula) float64 {
	if weight <= 0 || reps <= 0 || reps > MaxRepsForEstimate {
		return 0
	}
	var est float64
	switch f {
	case Brzycki:
		est = weight * 36 / float64(37-reps)
	default:
		est = weight * (1 + float64(reps)/30)
	}
	return math.Round(est*10) / 10
}

// WeightForReps inverts the Epley formula: the weight that should be
// possible for reps given an estimated 1RM. Used for training-max targets.
func WeightForReps(oneRepMax float64, reps int) float64 {
	if oneRepMax <= 0 || reps <= 0 {
		return 0
	}
	if reps == 1 {
		return oneRepMax
	}
	return oneRepMax / (1 + float64(reps)/30)
}
