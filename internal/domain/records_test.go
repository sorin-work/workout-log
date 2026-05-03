package domain

import (
	"testing"
	"time"
)

var day = time.Date(2026, 1, 5, 18, 0, 0, 0, time.UTC)

func logged(exID int64, name string, daysLater, reps int, kg float64) LoggedSet {
	return LoggedSet{
		ExerciseID: exID, ExerciseName: name, WorkoutID: int64(daysLater + 1),
		PerformedAt: day.AddDate(0, 0, daysLater), Reps: reps, WeightKg: kg,
	}
}

func TestComputeRecords(t *testing.T) {
	sets := []LoggedSet{
		logged(1, "Squat", 0, 5, 100),
		logged(1, "Squat", 2, 3, 110),
		logged(1, "Squat", 4, 8, 90),
		logged(2, "Bench", 1, 5, 80),
	}
	recs := ComputeRecords(sets, Epley)
	if len(recs) != 2 || recs[0].ExerciseName != "Bench" {
		t.Fatalf("records: %+v", recs)
	}
	sq := recs[1]
	if sq.SetCount != 3 || !sq.LastDone.Equal(day.AddDate(0, 0, 4)) {
		t.Fatalf("squat meta: %+v", sq)
	}
	// 90x8 = 114, 110x3 = 121, 100x5 = 116.7
	if sq.EstimatedMax.Value != 121 || sq.EstimatedMax.Reps != 3 {
		t.Errorf("estimated max: %+v", sq.EstimatedMax)
	}
	if sq.Heaviest.WeightKg != 110 {
		t.Errorf("heaviest: %+v", sq.Heaviest)
	}
	if sq.MostReps.Reps != 8 || sq.BestVolume.Value != 720 {
		t.Errorf("reps/volume: %+v %+v", sq.MostReps, sq.BestVolume)
	}
	if sq.RepMaxes[3].WeightKg != 110 || sq.RepMaxes[5].WeightKg != 100 || sq.RepMaxes[8].WeightKg != 90 {
		t.Errorf("rep maxes: %+v", sq.RepMaxes)
	}
	if _, ok := sq.RepMaxes[9]; ok {
		t.Error("no 9 rep max was set")
	}
}

func TestHeavySingleIsNotInflated(t *testing.T) {
	recs := ComputeRecords([]LoggedSet{
		logged(1, "Deadlift", 0, 1, 200),
		logged(1, "Deadlift", 1, 5, 170),
	}, Epley)
	// 170x5 estimates 198.3; the 200 single is the real max.
	if got := recs[0].EstimatedMax; got.Value != 200 || got.Reps != 1 {
		t.Fatalf("estimated max: %+v", got)
	}
}

func TestRecordsPreferTheFirstTimeOnTies(t *testing.T) {
	recs := ComputeRecords([]LoggedSet{
		logged(1, "Squat", 3, 5, 100),
		logged(1, "Squat", 1, 5, 100),
	}, Epley)
	if got := recs[0].Heaviest.PerformedAt; !got.Equal(day.AddDate(0, 0, 1)) {
		t.Fatalf("tie went to %v", got)
	}
}

func TestBodyweightExercisesTrackReps(t *testing.T) {
	recs := ComputeRecords([]LoggedSet{
		logged(5, "Pull-Up", 0, 8, 0),
		logged(5, "Pull-Up", 3, 11, 0),
	}, Epley)
	r := recs[0]
	if r.EstimatedMax != nil || r.Heaviest != nil || r.BestVolume != nil {
		t.Fatalf("bodyweight sets should only set rep records: %+v", r)
	}
	if r.MostReps.Reps != 11 {
		t.Fatalf("most reps: %+v", r.MostReps)
	}
}
