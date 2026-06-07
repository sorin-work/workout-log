package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sorin-work/workout-log/internal/domain"
)

// ExerciseWeeks returns completed working sets per exercise and week since
// from, for the weekly volume report.
func (s *Store) ExerciseWeeks(ctx context.Context, userID int64, from time.Time) ([]domain.ExerciseWeek, error) {
	rows, err := s.pool.Query(ctx, `SELECT date_trunc('week', w.started_at), e.primary_muscle, e.secondary_muscles,
			count(*), sum(ws.reps * ws.weight_kg)::float8
		FROM workout_sets ws
		JOIN workouts w ON w.id = ws.workout_id
		JOIN exercises e ON e.id = ws.exercise_id
		WHERE w.user_id = $1 AND w.started_at >= $2 AND ws.completed AND NOT ws.warmup
		GROUP BY 1, e.id
		ORDER BY 1`, userID, from)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ExerciseWeek, error) {
		var r domain.ExerciseWeek
		err := row.Scan(&r.WeekStart, &r.PrimaryMuscle, &r.SecondaryMuscles, &r.Sets, &r.VolumeKg)
		return r, err
	})
}

// WorkoutDays returns every workout of the user since from with its
// completed working sets and volume.
func (s *Store) WorkoutDays(ctx context.Context, userID int64, from time.Time) ([]domain.WorkoutDay, error) {
	rows, err := s.pool.Query(ctx, `SELECT w.started_at,
			count(ws.id) FILTER (WHERE ws.completed AND NOT ws.warmup),
			coalesce(sum(ws.reps * ws.weight_kg) FILTER (WHERE ws.completed AND NOT ws.warmup), 0)::float8
		FROM workouts w LEFT JOIN workout_sets ws ON ws.workout_id = w.id
		WHERE w.user_id = $1 AND w.started_at >= $2
		GROUP BY w.id ORDER BY w.started_at`, userID, from)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.WorkoutDay, error) {
		var d domain.WorkoutDay
		err := row.Scan(&d.StartedAt, &d.Sets, &d.VolumeKg)
		return d, err
	})
}

// WorkoutRef is a workout id, name and start time.
type WorkoutRef struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartedAt time.Time `json:"startedAt"`
}

// WorkoutsBetween lists the user's workouts that started in [from, to).
func (s *Store) WorkoutsBetween(ctx context.Context, userID int64, from, to time.Time) ([]WorkoutRef, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, started_at FROM workouts
		WHERE user_id = $1 AND started_at >= $2 AND started_at < $3 ORDER BY started_at`, userID, from, to)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[WorkoutRef])
}
