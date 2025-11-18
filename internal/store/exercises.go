package store

import (
	"context"
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
	rows, err := s.pool.Query(ctx, `SELECT `+exerciseColumns+` FROM exercises
		WHERE `+strings.Join(where, " AND ")+` ORDER BY lower(name)`, args...)
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
