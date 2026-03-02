package store

import "context"

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
