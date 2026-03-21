package domain

import (
	"testing"
	"time"
)

func TestSummarize(t *testing.T) {
	now := time.Date(2026, 3, 18, 20, 0, 0, 0, time.UTC) // Wednesday
	at := func(days int) time.Time { return now.AddDate(0, 0, days) }
	workouts := []WorkoutDay{
		{StartedAt: at(-1), Sets: 15, VolumeKg: 5000},
		{StartedAt: at(-2), Sets: 12, VolumeKg: 4000},
		{StartedAt: at(-7), Sets: 10, VolumeKg: 3000},
		{StartedAt: at(-14), Sets: 10, VolumeKg: 3000},
		{StartedAt: at(-35), Sets: 10, VolumeKg: 3000}, // gap before this one
	}
	s := Summarize(workouts, now, time.UTC)
	if s.ThisWeek != (WeekTotals{Workouts: 2, Sets: 27, VolumeKg: 9000}) {
		t.Errorf("this week: %+v", s.ThisWeek)
	}
	if s.LastWeek.Workouts != 1 || s.TotalWorkouts != 5 {
		t.Errorf("summary: %+v", s)
	}
	if s.WeekStreak != 3 {
		t.Errorf("streak = %d, want 3", s.WeekStreak)
	}
}

func TestStreakSurvivesAnEmptyCurrentWeek(t *testing.T) {
	now := time.Date(2026, 3, 16, 9, 0, 0, 0, time.UTC) // Monday morning
	s := Summarize([]WorkoutDay{{StartedAt: now.AddDate(0, 0, -3)}, {StartedAt: now.AddDate(0, 0, -10)}}, now, time.UTC)
	if s.WeekStreak != 2 {
		t.Fatalf("streak = %d, want 2", s.WeekStreak)
	}
	if Summarize(nil, now, time.UTC).WeekStreak != 0 {
		t.Fatal("no workouts should mean no streak")
	}
}
