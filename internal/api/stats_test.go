package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/domain"
)

func sumSets(weeks []domain.WeekVolume, muscle string) float64 {
	total := 0.0
	for _, w := range weeks {
		total += w.Muscles[muscle].Sets
	}
	return total
}

func TestWeeklyVolume(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("v@example.com")
	bench := exerciseID(t, c, "Bench Press")
	pushdown := exerciseID(t, c, "Triceps Pushdown")

	w1 := createWorkout(t, c, "Push", time.Now().Add(-8*24*time.Hour))
	for i := 0; i < 3; i++ {
		addSet(t, c, w1.ID, bench, 5, 80, true)
	}
	addSet(t, c, w1.ID, bench, 5, 80, false) // missed, does not count
	w2 := createWorkout(t, c, "Push", time.Now().Add(-15*24*time.Hour))
	addSet(t, c, w2.ID, bench, 5, 80, true)
	addSet(t, c, w2.ID, pushdown, 12, 30, true)
	addSet(t, c, w2.ID, pushdown, 12, 30, true)

	var weeks []domain.WeekVolume
	c.mustDo(http.MethodGet, "/api/stats/volume?weeks=4", nil, http.StatusOK, &weeks)
	if len(weeks) != 4 {
		t.Fatalf("weeks: %+v", weeks)
	}
	if got := sumSets(weeks, "chest"); got != 4 {
		t.Errorf("chest sets = %v, want 4", got)
	}
	// 4 bench sets at half a set each plus 2 pushdowns.
	if got := sumSets(weeks, "triceps"); got != 4 {
		t.Errorf("triceps sets = %v, want 4", got)
	}
	last := weeks[len(weeks)-1]
	if last.TotalSets != 0 {
		t.Errorf("this week should be empty: %+v", last)
	}
}

func TestWeeklyVolumeParams(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("v@example.com")
	var weeks []domain.WeekVolume
	c.mustDo(http.MethodGet, "/api/stats/volume", nil, http.StatusOK, &weeks)
	if len(weeks) != 12 {
		t.Fatalf("default should be 12 weeks, got %d", len(weeks))
	}
	c.mustDo(http.MethodGet, "/api/stats/volume?weeks=0", nil, http.StatusBadRequest, nil)
	c.mustDo(http.MethodGet, "/api/stats/volume?weeks=abc", nil, http.StatusBadRequest, nil)
}

func TestSummaryEndpoint(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("s@example.com")
	var empty domain.Summary
	c.mustDo(http.MethodGet, "/api/stats/summary", nil, http.StatusOK, &empty)
	if empty.TotalWorkouts != 0 || empty.WeekStreak != 0 {
		t.Fatalf("empty summary: %+v", empty)
	}
	w := createWorkout(t, c, "A", time.Now().Add(-7*24*time.Hour))
	addSet(t, c, w.ID, exerciseID(t, c, "Deadlift"), 5, 100, true)
	var s domain.Summary
	c.mustDo(http.MethodGet, "/api/stats/summary", nil, http.StatusOK, &s)
	if s.TotalWorkouts != 1 || s.LastWeek.Sets != 1 || s.LastWeek.VolumeKg != 500 || s.WeekStreak != 1 {
		t.Fatalf("summary: %+v", s)
	}
}

func TestCalendarUsesTheUsersDay(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("cal@example.com")
	c.mustDo(http.MethodPut, "/api/me", map[string]any{"unit": "kg", "timezone": "America/New_York"}, http.StatusOK, nil)
	// 02:30 UTC on 4 June is still the evening of 3 June in New York.
	createWorkout(t, c, "Late", time.Date(2026, 6, 4, 2, 30, 0, 0, time.UTC))
	createWorkout(t, c, "Early", time.Date(2026, 6, 3, 12, 0, 0, 0, time.UTC))
	createWorkout(t, c, "May", time.Date(2026, 5, 31, 23, 0, 0, 0, time.UTC))

	var days []calendarDay
	c.mustDo(http.MethodGet, "/api/stats/calendar?month=2026-06", nil, http.StatusOK, &days)
	if len(days) != 1 || days[0].Date != "2026-06-03" || len(days[0].Workouts) != 2 {
		t.Fatalf("days: %+v", days)
	}
	c.mustDo(http.MethodGet, "/api/stats/calendar?month=2026-05", nil, http.StatusOK, &days)
	if len(days) != 1 || days[0].Date != "2026-05-31" {
		t.Fatalf("may: %+v", days)
	}
	c.mustDo(http.MethodGet, "/api/stats/calendar?month=june", nil, http.StatusBadRequest, nil)
}
