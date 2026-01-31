package domain

import "testing"

func TestProgressSeriesOnePointPerWorkout(t *testing.T) {
	sets := []LoggedSet{
		logged(1, "Squat", 0, 5, 100),
		logged(1, "Squat", 0, 5, 100),
		logged(1, "Squat", 0, 3, 105),
		logged(1, "Squat", 3, 5, 102.5),
	}
	pts := ProgressSeries(sets, Epley)
	if len(pts) != 2 {
		t.Fatalf("points: %+v", pts)
	}
	first := pts[0]
	if first.Sets != 3 || first.Reps != 13 || first.TopWeightKg != 105 || first.VolumeKg != 1315 {
		t.Fatalf("first point: %+v", first)
	}
	if first.EstimatedMax != 116.7 {
		t.Fatalf("estimated max = %v", first.EstimatedMax)
	}
	if !pts[1].PerformedAt.After(first.PerformedAt) {
		t.Fatal("points out of order")
	}
}

func TestProgressSeriesEmpty(t *testing.T) {
	if pts := ProgressSeries(nil, Epley); len(pts) != 0 {
		t.Fatalf("points: %+v", pts)
	}
}
