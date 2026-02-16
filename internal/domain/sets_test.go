package domain

import "testing"

func ptr(f float64) *float64 { return &f }

func TestValidateSetAcceptsTypicalSets(t *testing.T) {
	sets := []SetInput{
		{Reps: 5, WeightKg: 100, Completed: true},
		{Reps: 8, WeightKg: 60, RPE: ptr(8.5), Completed: true},
		{Reps: 0, WeightKg: 80},
		{Reps: 12, WeightKg: 0, Completed: true},
		{Reps: 1, WeightKg: 220, RPE: ptr(10), Completed: true},
	}
	for _, s := range sets {
		if f := ValidateSet(s); len(f) != 0 {
			t.Errorf("ValidateSet(%+v) = %v", s, f)
		}
	}
}

func TestValidateSetRejectsBadValues(t *testing.T) {
	cases := map[string]struct {
		set   SetInput
		field string
	}{
		"negative reps":       {SetInput{Reps: -1}, "reps"},
		"too many reps":       {SetInput{Reps: 101}, "reps"},
		"negative weight":     {SetInput{Reps: 5, WeightKg: -2.5}, "weightKg"},
		"absurd weight":       {SetInput{Reps: 5, WeightKg: 1200}, "weightKg"},
		"rpe zero":            {SetInput{Reps: 5, RPE: ptr(0)}, "rpe"},
		"rpe above ten":       {SetInput{Reps: 5, RPE: ptr(11)}, "rpe"},
		"rpe percent typo":    {SetInput{Reps: 5, RPE: ptr(85)}, "rpe"},
		"rpe quarter step":    {SetInput{Reps: 5, RPE: ptr(7.25)}, "rpe"},
		"completed with zero": {SetInput{Reps: 0, Completed: true}, "reps"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if f := ValidateSet(c.set); f[c.field] == "" {
				t.Fatalf("expected an error on %s, got %v", c.field, f)
			}
		})
	}
}

func TestSetVolume(t *testing.T) {
	if v := SetVolume(SetInput{Reps: 5, WeightKg: 100, Completed: true}); v != 500 {
		t.Errorf("volume = %v, want 500", v)
	}
	if v := SetVolume(SetInput{Reps: 5, WeightKg: 60, Completed: true, Warmup: true}); v != 0 {
		t.Errorf("warm-up volume = %v, want 0", v)
	}
	if v := SetVolume(SetInput{Reps: 5, WeightKg: 100}); v != 0 {
		t.Errorf("incomplete set volume = %v, want 0", v)
	}
}
