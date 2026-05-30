package domain

import (
	"testing"
	"time"
)

func TestRecordsSkipSetsWithoutReps(t *testing.T) {
	recs := ComputeRecords([]LoggedSet{logged(1, "Squat", 0, 0, 100)}, Epley)
	if len(recs) != 0 {
		t.Fatalf("records: %+v", recs)
	}
}

func TestVolumeIgnoresPrimaryListedAsSecondary(t *testing.T) {
	mon := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	weeks := WeeklyVolume([]ExerciseWeek{
		{WeekStart: mon, PrimaryMuscle: "back", SecondaryMuscles: []string{"back", "biceps"}, Sets: 4, VolumeKg: 1000},
	}, mon, mon, time.UTC)
	if got := weeks[0].Muscles["back"].Sets; got != 4 {
		t.Fatalf("back sets = %v, want 4", got)
	}
}

func TestVolumeDropsRowsOutsideTheRange(t *testing.T) {
	mon := time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC)
	weeks := WeeklyVolume([]ExerciseWeek{
		{WeekStart: mon.AddDate(0, 0, -7), PrimaryMuscle: "quads", Sets: 5},
	}, mon, mon, time.UTC)
	if len(weeks) != 1 || weeks[0].TotalSets != 0 {
		t.Fatalf("weeks: %+v", weeks)
	}
}

func TestSummaryUsesTheUsersWeek(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	// Sunday 20:00 UTC is already Monday morning in Tokyo.
	now := time.Date(2026, 5, 24, 20, 0, 0, 0, time.UTC)
	s := Summarize([]WorkoutDay{{StartedAt: now.Add(-time.Hour), Sets: 3}}, now, tokyo)
	if s.ThisWeek.Workouts != 1 {
		t.Fatalf("summary: %+v", s)
	}
}

func TestRoundToPounds(t *testing.T) {
	if got := RoundTo(FromKg(100, Pounds), SmallestPlateStep(Pounds)); got != 220 {
		t.Fatalf("100 kg in plate-able pounds = %v", got)
	}
}
