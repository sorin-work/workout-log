package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// ImportedWorkout is a finished workout read from a file.
type ImportedWorkout struct {
	Name      string
	StartedAt time.Time
	Sets      []SetData
}

// ImportWorkouts stores the workouts in one transaction. A workout with the
// same name and start time as an existing one is skipped.
func (s *Store) ImportWorkouts(ctx context.Context, userID int64, list []ImportedWorkout) (created, sets, skipped int, err error) {
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		for _, w := range list {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM workouts
				WHERE user_id = $1 AND started_at = $2 AND name = $3)`, userID, w.StartedAt, w.Name).Scan(&exists); err != nil {
				return err
			}
			if exists {
				skipped++
				continue
			}
			var id int64
			if err := tx.QueryRow(ctx, `INSERT INTO workouts (user_id, name, started_at, finished_at)
				VALUES ($1, $2, $3, $3) RETURNING id`, userID, w.Name, w.StartedAt).Scan(&id); err != nil {
				return err
			}
			batch := &pgx.Batch{}
			for i, d := range w.Sets {
				batch.Queue(`INSERT INTO workout_sets (workout_id, exercise_id, position, reps, weight_kg, rpe, completed, warmup)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, id, d.ExerciseID, i+1, d.Reps, d.WeightKg, d.RPE, d.Completed, d.Warmup)
			}
			if err := tx.SendBatch(ctx, batch).Close(); err != nil {
				return err
			}
			created++
			sets += len(w.Sets)
		}
		return nil
	})
	return created, sets, skipped, err
}
