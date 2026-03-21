package domain

import "time"

// WorkoutDay is the minimum needed about a workout for the dashboard.
type WorkoutDay struct {
	StartedAt time.Time
	Sets      int
	VolumeKg  float64
}

// WeekTotals sums up one week of training.
type WeekTotals struct {
	Workouts int     `json:"workouts"`
	Sets     int     `json:"sets"`
	VolumeKg float64 `json:"volumeKg"`
}

// Summary is the dashboard overview.
type Summary struct {
	ThisWeek      WeekTotals `json:"thisWeek"`
	LastWeek      WeekTotals `json:"lastWeek"`
	WeekStreak    int        `json:"weekStreak"`
	TotalWorkouts int        `json:"totalWorkouts"`
}

// Summarize builds the dashboard numbers. The streak counts weeks in a row
// with at least one workout; the current week does not break it while it is
// still going.
func Summarize(workouts []WorkoutDay, now time.Time, loc *time.Location) Summary {
	thisWeek := WeekStart(now, loc)
	lastWeek := thisWeek.AddDate(0, 0, -7)
	trained := map[string]bool{}
	var s Summary
	for _, w := range workouts {
		s.TotalWorkouts++
		ws := WeekStart(w.StartedAt, loc)
		trained[ws.Format(time.DateOnly)] = true
		var t *WeekTotals
		switch {
		case ws.Equal(thisWeek):
			t = &s.ThisWeek
		case ws.Equal(lastWeek):
			t = &s.LastWeek
		default:
			continue
		}
		t.Workouts++
		t.Sets += w.Sets
		t.VolumeKg += w.VolumeKg
	}
	week := thisWeek
	if !trained[week.Format(time.DateOnly)] {
		week = lastWeek
	}
	for trained[week.Format(time.DateOnly)] {
		s.WeekStreak++
		week = week.AddDate(0, 0, -7)
	}
	return s
}
