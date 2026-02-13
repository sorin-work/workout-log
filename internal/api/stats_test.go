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
