package domain

import (
	"testing"
	"time"
)

func TestWeekStart(t *testing.T) {
	berlin, _ := time.LoadLocation("Europe/Berlin")
	cases := []struct {
		in   time.Time
		loc  *time.Location
		want string
	}{
		{time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC), time.UTC, "2026-02-09"},  // Monday
		{time.Date(2026, 2, 15, 23, 0, 0, 0, time.UTC), time.UTC, "2026-02-09"}, // Sunday night
		{time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC), berlin, "2026-02-16"},  // already Monday in Berlin
		{time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), time.UTC, "2025-12-29"},  // across the year
	}
	for _, c := range cases {
		if got := WeekStart(c.in, c.loc).Format(time.DateOnly); got != c.want {
			t.Errorf("WeekStart(%v, %v) = %s, want %s", c.in, c.loc, got, c.want)
		}
	}
}

func TestWeeklyVolume(t *testing.T) {
	mon := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	rows := []ExerciseWeek{
		{WeekStart: mon, PrimaryMuscle: "chest", SecondaryMuscles: []string{"triceps", "shoulders"}, Sets: 5, VolumeKg: 2000},
		{WeekStart: mon, PrimaryMuscle: "triceps", Sets: 3, VolumeKg: 600},
		{WeekStart: mon.AddDate(0, 0, 14), PrimaryMuscle: "quads", SecondaryMuscles: []string{"glutes"}, Sets: 5, VolumeKg: 2500},
	}
	weeks := WeeklyVolume(rows, mon.AddDate(0, 0, 3), mon.AddDate(0, 0, 16), time.UTC)
	if len(weeks) != 3 {
		t.Fatalf("weeks: %+v", weeks)
	}
	w0 := weeks[0]
	if w0.WeekStart != "2026-02-02" || w0.TotalSets != 8 {
		t.Fatalf("week 0: %+v", w0)
	}
	if got := w0.Muscles["triceps"]; got.Sets != 5.5 || got.VolumeKg != 600 {
		t.Errorf("triceps: %+v", got)
	}
	if got := w0.Muscles["chest"]; got.Sets != 5 || got.VolumeKg != 2000 {
		t.Errorf("chest: %+v", got)
	}
	if len(weeks[1].Muscles) != 0 {
		t.Errorf("empty week should have no muscles: %+v", weeks[1])
	}
	if weeks[2].Muscles["glutes"].Sets != 2.5 {
		t.Errorf("glutes: %+v", weeks[2].Muscles)
	}
}
