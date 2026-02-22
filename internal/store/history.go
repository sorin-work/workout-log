package store

import (
	"context"
	"time"

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

// LastSession is what the user did for an exercise the previous time.
type LastSession struct {
	WorkoutID   int64     `json:"workoutId"`
	PerformedAt time.Time `json:"performedAt"`
	Sets        []Set     `json:"sets"`
}

// LastSession returns the most recent workout, other than excludeWorkoutID,
// that has completed sets of the exercise. ErrNotFound means never done.
func (s *Store) LastSession(ctx context.Context, userID, exerciseID, excludeWorkoutID int64) (LastSession, error) {
	var last LastSession
	err := s.pool.QueryRow(ctx, `SELECT w.id, w.started_at FROM workouts w
		WHERE w.user_id = $1 AND w.id <> $3 AND EXISTS (
			SELECT 1 FROM workout_sets ws
			WHERE ws.workout_id = w.id AND ws.exercise_id = $2 AND ws.completed)
		ORDER BY w.started_at DESC LIMIT 1`, userID, exerciseID, excludeWorkoutID).Scan(&last.WorkoutID, &last.PerformedAt)
	if err != nil {
		return LastSession{}, notFound(err)
	}
	rows, err := s.pool.Query(ctx, `SELECT `+setColumns+`
		FROM workout_sets ws JOIN exercises e ON e.id = ws.exercise_id
		WHERE ws.workout_id = $1 AND ws.exercise_id = $2 AND ws.completed
		ORDER BY ws.position`, last.WorkoutID, exerciseID)
	if err != nil {
		return LastSession{}, err
	}
	last.Sets, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (Set, error) { return scanSet(row) })
	return last, err
}
