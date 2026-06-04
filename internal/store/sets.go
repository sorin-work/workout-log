package store

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// SetData holds the fields of a set being added or changed.
type SetData struct {
	ExerciseID int64
	Reps       int
	WeightKg   float64
	RPE        *float64
	Completed  bool
	Warmup     bool
}

// AddSet appends a set to the end of a workout.
func (s *Store) AddSet(ctx context.Context, workoutID int64, d SetData) (Set, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `INSERT INTO workout_sets
		(workout_id, exercise_id, position, reps, weight_kg, rpe, completed, warmup)
		VALUES ($1, $2,
			(SELECT coalesce(max(position), 0) + 1 FROM workout_sets WHERE workout_id = $1),
			$3, $4, $5, $6, $7)
		RETURNING id`,
		workoutID, d.ExerciseID, d.Reps, d.WeightKg, d.RPE, d.Completed, d.Warmup).Scan(&id)
	if err != nil {
		return Set{}, err
	}
	return s.setByID(ctx, id)
}

// UpdateSet changes a set of the given workout. Callers must have checked
// that the workout belongs to the user.
func (s *Store) UpdateSet(ctx context.Context, workoutID, setID int64, d SetData) (Set, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE workout_sets
		SET exercise_id = $3, reps = $4, weight_kg = $5, rpe = $6, completed = $7, warmup = $8
		WHERE id = $1 AND workout_id = $2`,
		setID, workoutID, d.ExerciseID, d.Reps, d.WeightKg, d.RPE, d.Completed, d.Warmup)
	if err != nil {
		return Set{}, err
	}
	if tag.RowsAffected() == 0 {
		return Set{}, ErrNotFound
	}
	return s.setByID(ctx, setID)
}

// DeleteSet removes a set from the given workout.
func (s *Store) DeleteSet(ctx context.Context, workoutID, setID int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM workout_sets WHERE id = $1 AND workout_id = $2`, setID, workoutID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) setByID(ctx context.Context, id int64) (Set, error) {
	st, err := scanSet(s.pool.QueryRow(ctx, `SELECT `+setColumns+`
		FROM workout_sets ws JOIN exercises e ON e.id = ws.exercise_id
		WHERE ws.id = $1`, id))
	return st, notFound(err)
}

// InsertWarmups adds warm-up sets for an exercise right before its first set
// in the workout, shifting the later sets down.
func (s *Store) InsertWarmups(ctx context.Context, workoutID, exerciseID int64, warmups []SetData) ([]Set, error) {
	var ids []int64
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		var first int
		err := tx.QueryRow(ctx, `SELECT coalesce(min(position), coalesce(max(position), 0) + 1)
			FROM workout_sets WHERE workout_id = $1 AND (exercise_id = $2 OR NOT EXISTS (
				SELECT 1 FROM workout_sets WHERE workout_id = $1 AND exercise_id = $2))`,
			workoutID, exerciseID).Scan(&first)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE workout_sets SET position = position + $3
			WHERE workout_id = $1 AND position >= $2`, workoutID, first, len(warmups)); err != nil {
			return err
		}
		for i, d := range warmups {
			var id int64
			if err := tx.QueryRow(ctx, `INSERT INTO workout_sets
				(workout_id, exercise_id, position, reps, weight_kg, warmup)
				VALUES ($1, $2, $3, $4, $5, true) RETURNING id`,
				workoutID, exerciseID, first+i, d.Reps, d.WeightKg).Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := make([]Set, 0, len(ids))
	for _, id := range ids {
		st, err := s.setByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, st)
	}
	return out, nil
}
