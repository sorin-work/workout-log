package domain

import (
	"sort"
	"time"
)

// LoggedSet is a completed working set with the context needed for records
// and charts.
type LoggedSet struct {
	ExerciseID   int64
	ExerciseName string
	WorkoutID    int64
	PerformedAt  time.Time
	Reps         int
	WeightKg     float64
}

// RecordSet points at the set that holds a record.
type RecordSet struct {
	Value       float64   `json:"value"`
	WeightKg    float64   `json:"weightKg"`
	Reps        int       `json:"reps"`
	WorkoutID   int64     `json:"workoutId"`
	PerformedAt time.Time `json:"performedAt"`
}

// ExerciseRecords are the personal records for one exercise.
type ExerciseRecords struct {
	ExerciseID   int64      `json:"exerciseId"`
	ExerciseName string     `json:"exerciseName"`
	EstimatedMax *RecordSet `json:"estimatedMax"`
	Heaviest     *RecordSet `json:"heaviest"`
	MostReps     *RecordSet `json:"mostReps"`
	BestVolume   *RecordSet `json:"bestVolume"`
	// RepMaxes[n] is the heaviest weight lifted for at least n reps.
	RepMaxes map[int]RecordSet `json:"repMaxes"`
	SetCount int               `json:"setCount"`
	LastDone time.Time         `json:"lastDone"`
}

// better reports whether a beats b, preferring the earlier set on ties so a
// record belongs to whoever set it first.
func better(value float64, at time.Time, cur *RecordSet) bool {
	if cur == nil {
		return true
	}
	if value != cur.Value {
		return value > cur.Value
	}
	return at.Before(cur.PerformedAt)
}

func record(value float64, s LoggedSet) *RecordSet {
	return &RecordSet{Value: value, WeightKg: s.WeightKg, Reps: s.Reps, WorkoutID: s.WorkoutID, PerformedAt: s.PerformedAt}
}

// ComputeRecords finds the personal records per exercise, sorted by exercise
// name.
func ComputeRecords(sets []LoggedSet, f Formula) []ExerciseRecords {
	byExercise := map[int64]*ExerciseRecords{}
	for _, s := range sets {
		if s.Reps <= 0 {
			continue
		}
		r := byExercise[s.ExerciseID]
		if r == nil {
			r = &ExerciseRecords{ExerciseID: s.ExerciseID, ExerciseName: s.ExerciseName, RepMaxes: map[int]RecordSet{}}
			byExercise[s.ExerciseID] = r
		}
		r.SetCount++
		if s.PerformedAt.After(r.LastDone) {
			r.LastDone = s.PerformedAt
		}
		if e := EstimateOneRepMax(s.WeightKg, s.Reps, f); e > 0 && better(e, s.PerformedAt, r.EstimatedMax) {
			r.EstimatedMax = record(e, s)
		}
		if s.WeightKg > 0 && better(s.WeightKg, s.PerformedAt, r.Heaviest) {
			r.Heaviest = record(s.WeightKg, s)
		}
		if better(float64(s.Reps), s.PerformedAt, r.MostReps) {
			r.MostReps = record(float64(s.Reps), s)
		}
		if v := float64(s.Reps) * s.WeightKg; v > 0 && better(v, s.PerformedAt, r.BestVolume) {
			r.BestVolume = record(v, s)
		}
		if s.WeightKg > 0 {
			for n := 1; n <= s.Reps && n <= MaxRepsForEstimate; n++ {
				cur, ok := r.RepMaxes[n]
				if !ok || better(s.WeightKg, s.PerformedAt, &cur) {
					r.RepMaxes[n] = *record(s.WeightKg, s)
				}
			}
		}
	}
	out := make([]ExerciseRecords, 0, len(byExercise))
	for _, r := range byExercise {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ExerciseName < out[j].ExerciseName })
	return out
}
