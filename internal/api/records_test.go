package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/domain"
)

func TestRecordsEndpoint(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("r@example.com")
	squat := exerciseID(t, c, "Back Squat")
	w1 := createWorkout(t, c, "Mon", time.Now().Add(-72*time.Hour))
	addSet(t, c, w1.ID, squat, 5, 100, true)
	addSet(t, c, w1.ID, squat, 5, 120, false) // not completed
	w2 := createWorkout(t, c, "Thu", time.Now().Add(-2*time.Hour))
	addSet(t, c, w2.ID, squat, 3, 110, true)

	var recs []domain.ExerciseRecords
	c.mustDo(http.MethodGet, "/api/records", nil, http.StatusOK, &recs)
	if len(recs) != 1 {
		t.Fatalf("records: %+v", recs)
	}
	r := recs[0]
	if r.Heaviest.WeightKg != 110 || r.Heaviest.WorkoutID != w2.ID {
		t.Fatalf("heaviest should ignore the missed 120 kg set: %+v", r.Heaviest)
	}
	if r.EstimatedMax.Value != 121 || r.SetCount != 2 {
		t.Fatalf("records: %+v", r)
	}

	// Other users see nothing.
	other := a.signup("other@example.com")
	other.mustDo(http.MethodGet, "/api/records", nil, http.StatusOK, &recs)
	if len(recs) != 0 {
		t.Fatalf("other user sees records: %+v", recs)
	}
}

func TestExerciseRecordsEndpoint(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("r@example.com")
	bench := exerciseID(t, c, "Bench Press")
	var empty domain.ExerciseRecords
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/exercises/%d/records", bench), nil, http.StatusOK, &empty)
	if empty.ExerciseName != "Bench Press" || empty.EstimatedMax != nil {
		t.Fatalf("empty records: %+v", empty)
	}
	w := createWorkout(t, c, "Push", time.Now().Add(-time.Hour))
	addSet(t, c, w.ID, bench, 8, 70, true)
	var got domain.ExerciseRecords
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/exercises/%d/records", bench), nil, http.StatusOK, &got)
	if got.RepMaxes[8].WeightKg != 70 || got.RepMaxes[1].WeightKg != 70 {
		t.Fatalf("rep maxes: %+v", got.RepMaxes)
	}
	c.mustDo(http.MethodGet, "/api/exercises/999999/records", nil, http.StatusNotFound, nil)
}

func TestExerciseProgressEndpoint(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("r@example.com")
	dl := exerciseID(t, c, "Deadlift")
	for i, kg := range []float64{140, 145, 150} {
		w := createWorkout(t, c, "Pull", time.Now().Add(time.Duration(i-5)*24*time.Hour))
		addSet(t, c, w.ID, dl, 5, kg, true)
		addSet(t, c, w.ID, dl, 5, kg-20, true)
	}
	var pts []domain.ProgressPoint
	c.mustDo(http.MethodGet, fmt.Sprintf("/api/exercises/%d/progress", dl), nil, http.StatusOK, &pts)
	if len(pts) != 3 {
		t.Fatalf("points: %+v", pts)
	}
	if pts[0].TopWeightKg != 140 || pts[2].TopWeightKg != 150 || pts[2].Sets != 2 {
		t.Fatalf("points: %+v", pts)
	}
	if pts[2].VolumeKg != 5*150+5*130 {
		t.Fatalf("volume = %v", pts[2].VolumeKg)
	}
}
