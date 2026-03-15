package store

import (
	"context"
	"time"
)

// ExportRow is one set in the CSV export.
type ExportRow struct {
	StartedAt    time.Time
	WorkoutName  string
	ExerciseName string
	Position     int
	Reps         int
	WeightKg     float64
	RPE          *float64
	Warmup       bool
	Completed    bool
}

// ExportSets streams every set of the user, oldest workout first, to fn.
func (s *Store) ExportSets(ctx context.Context, userID int64, fn func(ExportRow) error) error {
	rows, err := s.pool.Query(ctx, `SELECT w.started_at, w.name, e.name, ws.position, ws.reps,
			ws.weight_kg::float8, ws.rpe::float8, ws.warmup, ws.completed
		FROM workout_sets ws
		JOIN workouts w ON w.id = ws.workout_id
		JOIN exercises e ON e.id = ws.exercise_id
		WHERE w.user_id = $1
		ORDER BY w.started_at, w.id, ws.position`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var r ExportRow
		if err := rows.Scan(&r.StartedAt, &r.WorkoutName, &r.ExerciseName, &r.Position, &r.Reps,
			&r.WeightKg, &r.RPE, &r.Warmup, &r.Completed); err != nil {
			return err
		}
		if err := fn(r); err != nil {
			return err
		}
	}
	return rows.Err()
}
