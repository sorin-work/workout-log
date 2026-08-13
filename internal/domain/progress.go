package domain

import (
	"math"
	"sort"
	"time"
)

// ProgressPoint summarises one exercise in one workout, for charts.
type ProgressPoint struct {
	WorkoutID    int64     `json:"workoutId"`
	PerformedAt  time.Time `json:"performedAt"`
	EstimatedMax float64   `json:"estimatedMax"`
	TopWeightKg  float64   `json:"topWeightKg"`
	VolumeKg     float64   `json:"volumeKg"`
	Sets         int       `json:"sets"`
	Reps         int       `json:"reps"`
}

// ProgressSeries turns logged sets of a single exercise into one point per
// workout.
func ProgressSeries(sets []LoggedSet, f Formula) []ProgressPoint {
	byWorkout := map[int64]*ProgressPoint{}
	for _, s := range sets {
		p := byWorkout[s.WorkoutID]
		if p == nil {
			p = &ProgressPoint{WorkoutID: s.WorkoutID, PerformedAt: s.PerformedAt}
			byWorkout[s.WorkoutID] = p
		}
		p.Sets++
		p.Reps += s.Reps
		p.VolumeKg += float64(s.Reps) * s.WeightKg
		p.TopWeightKg = math.Max(p.TopWeightKg, s.WeightKg)
		p.EstimatedMax = math.Max(p.EstimatedMax, EstimateOneRepMax(s.WeightKg, s.Reps, f))
	}
	out := make([]ProgressPoint, 0, len(byWorkout))
	for _, p := range byWorkout {
		p.VolumeKg = math.Round(p.VolumeKg*10) / 10
		out = append(out, *p)
	}
	// Order by when the workout happened. Ids only follow the order in which
	// workouts were logged, and back-filled workouts get newer ids.
	sort.Slice(out, func(i, j int) bool {
		if !out[i].PerformedAt.Equal(out[j].PerformedAt) {
			return out[i].PerformedAt.Before(out[j].PerformedAt)
		}
		return out[i].WorkoutID < out[j].WorkoutID
	})
	return out
}
