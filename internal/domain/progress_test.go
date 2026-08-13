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

func TestProgressSeriesOrdersByDateNotID(t *testing.T) {
	// Workout 9 was logged last but happened first.
	sets := []LoggedSet{
		{ExerciseID: 1, WorkoutID: 7, PerformedAt: day.AddDate(0, 0, 7), Reps: 5, WeightKg: 105},
		{ExerciseID: 1, WorkoutID: 8, PerformedAt: day.AddDate(0, 0, 14), Reps: 5, WeightKg: 107.5},
		{ExerciseID: 1, WorkoutID: 9, PerformedAt: day, Reps: 5, WeightKg: 100},
	}
	pts := ProgressSeries(sets, Epley)
	if pts[0].WorkoutID != 9 || pts[1].WorkoutID != 7 || pts[2].WorkoutID != 8 {
		t.Fatalf("order: %d, %d, %d", pts[0].WorkoutID, pts[1].WorkoutID, pts[2].WorkoutID)
	}
}
