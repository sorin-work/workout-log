package domain

import "testing"

func TestEstimateOneRepMaxEpley(t *testing.T) {
	cases := []struct {
		weight float64
		reps   int
		want   float64
	}{
		{100, 5, 116.7},
		{100, 10, 133.3},
		{60, 8, 76},
		{142.5, 3, 156.8},
	}
	for _, c := range cases {
		if got := EstimateOneRepMax(c.weight, c.reps, Epley); got != c.want {
			t.Errorf("Epley(%v x %d) = %v, want %v", c.weight, c.reps, got, c.want)
		}
	}
}

func TestEstimateOneRepMaxBrzycki(t *testing.T) {
	if got := EstimateOneRepMax(100, 5, Brzycki); got != 112.5 {
		t.Errorf("Brzycki(100 x 5) = %v, want 112.5", got)
	}
	if got := EstimateOneRepMax(100, 10, Brzycki); got != 133.3 {
		t.Errorf("Brzycki(100 x 10) = %v, want 133.3", got)
	}
}

func TestEstimateOneRepMaxUnusableSets(t *testing.T) {
	cases := []struct {
		weight float64
		reps   int
	}{
		{0, 5}, {-20, 5}, {100, 0}, {100, 13}, {100, 30},
	}
	for _, c := range cases {
		if got := EstimateOneRepMax(c.weight, c.reps, Epley); got != 0 {
			t.Errorf("EstimateOneRepMax(%v, %d) = %v, want 0", c.weight, c.reps, got)
		}
	}
}

func TestWeightForReps(t *testing.T) {
	if got := WeightForReps(120, 1); got != 120 {
		t.Errorf("1 rep at 120 1RM = %v", got)
	}
	if got := RoundTo(WeightForReps(120, 5), 0.5); got != 103 {
		t.Errorf("5 reps at 120 1RM = %v, want ~103", got)
	}
	if got := WeightForReps(0, 5); got != 0 {
		t.Errorf("no 1RM should give 0, got %v", got)
	}
}
