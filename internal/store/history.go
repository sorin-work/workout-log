package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/sorin-work/workout-log/internal/domain"
)

// LoggedSets returns the user's completed working sets, oldest first. When
// exerciseID is non-zero only that exercise is returned.
func (s *Store) LoggedSets(ctx context.Context, userID, exerciseID int64) ([]domain.LoggedSet, error) {
	rows, err := s.pool.Query(ctx, `SELECT ws.exercise_id, e.name, w.id, w.started_at, ws.reps, ws.weight_kg::float8
		FROM workout_sets ws
		JOIN workouts w ON w.id = ws.workout_id
		JOIN exercises e ON e.id = ws.exercise_id
		WHERE w.user_id = $1 AND ws.completed AND NOT ws.warmup
		  AND ($2 = 0 OR ws.exercise_id = $2)
		ORDER BY w.started_at, ws.position`, userID, exerciseID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.LoggedSet, error) {
		var l domain.LoggedSet
		err := row.Scan(&l.ExerciseID, &l.ExerciseName, &l.WorkoutID, &l.PerformedAt, &l.Reps, &l.WeightKg)
		return l, err
	})
}
