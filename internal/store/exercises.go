package store

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Exercise is a lift in the library. Built-in exercises have no owner.
type Exercise struct {
	ID               int64      `json:"id"`
	UserID           *int64     `json:"-"`
	Name             string     `json:"name"`
	PrimaryMuscle    string     `json:"primaryMuscle"`
	SecondaryMuscles []string   `json:"secondaryMuscles"`
	Equipment        string     `json:"equipment"`
	Custom           bool       `json:"custom"`
	ArchivedAt       *time.Time `json:"archivedAt"`
}

// ExerciseFilter narrows ListExercises.
type ExerciseFilter struct {
	Query           string
	Muscle          string
	Equipment       string
	IncludeArchived bool
	// Popular sorts the exercises the user logged most in the last 90 days
	// first, then by name.
	Popular bool
}

const exerciseColumns = `id, user_id, name, primary_muscle, secondary_muscles, equipment, archived_at`

func scanExercise(row interface{ Scan(...any) error }) (Exercise, error) {
	var e Exercise
	err := row.Scan(&e.ID, &e.UserID, &e.Name, &e.PrimaryMuscle, &e.SecondaryMuscles, &e.Equipment, &e.ArchivedAt)
	e.Custom = e.UserID != nil
	if e.SecondaryMuscles == nil {
		e.SecondaryMuscles = []string{}
	}
	return e, err
}

// ListExercises returns the built-in exercises plus the user's own ones,
// sorted by name.
func (s *Store) ListExercises(ctx context.Context, userID int64, f ExerciseFilter) ([]Exercise, error) {
	where := []string{"(user_id IS NULL OR user_id = $1)"}
	args := []any{userID}
	if q := strings.TrimSpace(f.Query); q != "" {
		args = append(args, "%"+escapeLike(q)+"%")
		where = append(where, "name ILIKE $"+itoa(len(args)))
	}
	if f.Muscle != "" {
		args = append(args, f.Muscle)
		n := itoa(len(args))
		where = append(where, "(primary_muscle = $"+n+" OR $"+n+" = ANY(secondary_muscles))")
	}
	if f.Equipment != "" {
		args = append(args, f.Equipment)
		where = append(where, "equipment = $"+itoa(len(args)))
	}
	if !f.IncludeArchived {
		where = append(where, "archived_at IS NULL")
	}
	order := "lower(name)"
	if f.Popular {
		order = `(SELECT count(*) FROM workout_sets ws JOIN workouts w ON w.id = ws.workout_id
			WHERE ws.exercise_id = exercises.id AND w.user_id = $1
			  AND w.started_at > now() - interval '90 days') DESC, lower(name)`
	}
	rows, err := s.pool.Query(ctx, `SELECT `+exerciseColumns+` FROM exercises
		WHERE `+strings.Join(where, " AND ")+` ORDER BY `+order, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Exercise{}
	for rows.Next() {
		e, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ExerciseByID returns an exercise the user can see.
func (s *Store) ExerciseByID(ctx context.Context, userID, id int64) (Exercise, error) {
	e, err := scanExercise(s.pool.QueryRow(ctx, `SELECT `+exerciseColumns+` FROM exercises
		WHERE id = $1 AND (user_id IS NULL OR user_id = $2)`, id, userID))
	return e, notFound(err)
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// ExerciseInput holds the editable fields of a custom exercise.
type ExerciseInput struct {
	Name             string
	PrimaryMuscle    string
	SecondaryMuscles []string
	Equipment        string
}

// CreateExercise adds a custom exercise for a user. It returns ErrConflict
// if the user already has an exercise with that name, or one exists in the
// built-in library.
func (s *Store) CreateExercise(ctx context.Context, userID int64, in ExerciseInput) (Exercise, error) {
	if taken, err := s.builtinNameTaken(ctx, in.Name); err != nil || taken {
		if err == nil {
			err = ErrConflict
		}
		return Exercise{}, err
	}
	e, err := scanExercise(s.pool.QueryRow(ctx, `INSERT INTO exercises
		(user_id, name, primary_muscle, secondary_muscles, equipment)
		VALUES ($1, $2, $3, $4, $5) RETURNING `+exerciseColumns,
		userID, in.Name, in.PrimaryMuscle, in.SecondaryMuscles, in.Equipment))
	if isUniqueViolation(err) {
		return Exercise{}, ErrConflict
	}
	return e, err
}

// UpdateExercise changes a custom exercise owned by the user.
func (s *Store) UpdateExercise(ctx context.Context, userID, id int64, in ExerciseInput) (Exercise, error) {
	if taken, err := s.builtinNameTaken(ctx, in.Name); err != nil || taken {
		if err == nil {
			err = ErrConflict
		}
		return Exercise{}, err
	}
	e, err := scanExercise(s.pool.QueryRow(ctx, `UPDATE exercises
		SET name = $3, primary_muscle = $4, secondary_muscles = $5, equipment = $6
		WHERE id = $1 AND user_id = $2 RETURNING `+exerciseColumns,
		id, userID, in.Name, in.PrimaryMuscle, in.SecondaryMuscles, in.Equipment))
	if isUniqueViolation(err) {
		return Exercise{}, ErrConflict
	}
	return e, notFound(err)
}

// SetExerciseArchived archives or restores a custom exercise. Archived
// exercises keep their history but are hidden from pickers.
func (s *Store) SetExerciseArchived(ctx context.Context, userID, id int64, archived bool) (Exercise, error) {
	e, err := scanExercise(s.pool.QueryRow(ctx, `UPDATE exercises
		SET archived_at = CASE WHEN $3 THEN coalesce(archived_at, now()) END
		WHERE id = $1 AND user_id = $2 RETURNING `+exerciseColumns, id, userID, archived))
	return e, notFound(err)
}

func (s *Store) builtinNameTaken(ctx context.Context, name string) (bool, error) {
	var taken bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM exercises
		WHERE user_id IS NULL AND lower(name) = lower($1))`, name).Scan(&taken)
	return taken, err
}

// ExerciseNotes returns the user's notes for an exercise, or "".
func (s *Store) ExerciseNotes(ctx context.Context, userID, exerciseID int64) (string, error) {
	var notes string
	err := s.pool.QueryRow(ctx, `SELECT notes FROM exercise_notes WHERE user_id = $1 AND exercise_id = $2`,
		userID, exerciseID).Scan(&notes)
	if err := notFound(err); err == ErrNotFound {
		return "", nil
	}
	return notes, err
}

// SetExerciseNotes saves the user's notes for an exercise. Empty notes are
// removed.
func (s *Store) SetExerciseNotes(ctx context.Context, userID, exerciseID int64, notes string) error {
	if notes == "" {
		_, err := s.pool.Exec(ctx, `DELETE FROM exercise_notes WHERE user_id = $1 AND exercise_id = $2`, userID, exerciseID)
		return err
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO exercise_notes (user_id, exercise_id, notes) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, exercise_id) DO UPDATE SET notes = excluded.notes, updated_at = now()`,
		userID, exerciseID, notes)
	return err
}

// ErrInUse is returned when deleting an exercise that has history.
var ErrInUse = errors.New("in use")

// DeleteExercise removes a custom exercise that was never logged or used in
// a template. Exercises with history should be archived instead.
func (s *Store) DeleteExercise(ctx context.Context, userID, id int64) error {
	var used bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM workout_sets WHERE exercise_id = $1)
		OR EXISTS (SELECT 1 FROM template_exercises WHERE exercise_id = $1)`, id).Scan(&used); err != nil {
		return err
	}
	if used {
		return ErrInUse
	}
	tag, err := s.pool.Exec(ctx, `DELETE FROM exercises WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
