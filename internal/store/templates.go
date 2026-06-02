package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Template is a reusable workout plan.
type Template struct {
	ID        int64              `json:"id"`
	Name      string             `json:"name"`
	Notes     string             `json:"notes"`
	Exercises []TemplateExercise `json:"exercises"`
}

// TemplateExercise is one exercise in a template with its target sets.
type TemplateExercise struct {
	ID           int64   `json:"id"`
	ExerciseID   int64   `json:"exerciseId"`
	ExerciseName string  `json:"exerciseName"`
	Position     int     `json:"position"`
	Sets         int     `json:"sets"`
	Reps         int     `json:"reps"`
	WeightKg     float64 `json:"weightKg"`
	RestSeconds  int     `json:"restSeconds"`
	// Progression: weight added after a successful session, and the
	// deload applied after DeloadAfter failed sessions in a row.
	IncrementKg   float64 `json:"incrementKg"`
	DeloadAfter   int     `json:"deloadAfter"`
	DeloadPercent float64 `json:"deloadPercent"`
	Failures      int     `json:"failures"`
}

// TemplateInput is a full template as sent by the client. Exercises with an
// ID update the existing row; the others are added. Rows not listed are
// removed.
type TemplateInput struct {
	Name      string
	Notes     string
	Exercises []TemplateExerciseInput
}

// TemplateExerciseInput is one exercise row of a TemplateInput.
type TemplateExerciseInput struct {
	ID            int64
	ExerciseID    int64
	Sets          int
	Reps          int
	WeightKg      float64
	RestSeconds   int
	IncrementKg   float64
	DeloadAfter   int
	DeloadPercent float64
}

// ListTemplates returns the user's templates with their exercises.
func (s *Store) ListTemplates(ctx context.Context, userID int64) ([]Template, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, notes FROM templates
		WHERE user_id = $1 ORDER BY lower(name), id`, userID)
	if err != nil {
		return nil, err
	}
	list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (Template, error) {
		var t Template
		err := row.Scan(&t.ID, &t.Name, &t.Notes)
		return t, err
	})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}
	ids := make([]int64, len(list))
	byID := make(map[int64]*Template, len(list))
	for i := range list {
		ids[i] = list[i].ID
		list[i].Exercises = []TemplateExercise{}
		byID[list[i].ID] = &list[i]
	}
	rows, err = s.pool.Query(ctx, `SELECT te.template_id, `+templateExerciseColumns+`
		FROM template_exercises te JOIN exercises e ON e.id = te.exercise_id
		WHERE te.template_id = ANY($1) ORDER BY te.template_id, te.position, te.id`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var templateID int64
		var te TemplateExercise
		if err := rows.Scan(append([]any{&templateID}, te.scanTargets()...)...); err != nil {
			return nil, err
		}
		t := byID[templateID]
		t.Exercises = append(t.Exercises, te)
	}
	return list, rows.Err()
}

const templateExerciseColumns = `te.id, te.exercise_id, e.name, te.position, te.sets, te.reps,
	te.weight_kg::float8, te.rest_seconds, te.increment_kg::float8, te.deload_after,
	te.deload_percent::float8, te.failures`

func (te *TemplateExercise) scanTargets() []any {
	return []any{&te.ID, &te.ExerciseID, &te.ExerciseName, &te.Position, &te.Sets, &te.Reps,
		&te.WeightKg, &te.RestSeconds, &te.IncrementKg, &te.DeloadAfter, &te.DeloadPercent, &te.Failures}
}

// TemplateByID returns one of the user's templates.
func (s *Store) TemplateByID(ctx context.Context, userID, id int64) (Template, error) {
	var t Template
	err := s.pool.QueryRow(ctx, `SELECT id, name, notes FROM templates
		WHERE id = $1 AND user_id = $2`, id, userID).Scan(&t.ID, &t.Name, &t.Notes)
	if err != nil {
		return Template{}, notFound(err)
	}
	t.Exercises, err = s.templateExercises(ctx, s.pool, id)
	return t, err
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (s *Store) templateExercises(ctx context.Context, q querier, templateID int64) ([]TemplateExercise, error) {
	rows, err := q.Query(ctx, `SELECT `+templateExerciseColumns+`
		FROM template_exercises te JOIN exercises e ON e.id = te.exercise_id
		WHERE te.template_id = $1 ORDER BY te.position, te.id`, templateID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TemplateExercise, error) {
		var te TemplateExercise
		err := row.Scan(te.scanTargets()...)
		return te, err
	})
}

// CreateTemplate stores a new template.
func (s *Store) CreateTemplate(ctx context.Context, userID int64, in TemplateInput) (Template, error) {
	var id int64
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO templates (user_id, name, notes)
			VALUES ($1, $2, $3) RETURNING id`, userID, in.Name, in.Notes).Scan(&id); err != nil {
			return err
		}
		return saveTemplateExercises(ctx, tx, id, in.Exercises)
	})
	if err != nil {
		return Template{}, err
	}
	return s.TemplateByID(ctx, userID, id)
}

// UpdateTemplate replaces the name, notes and exercises of a template.
func (s *Store) UpdateTemplate(ctx context.Context, userID, id int64, in TemplateInput) (Template, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE templates SET name = $3, notes = $4
			WHERE id = $1 AND user_id = $2`, id, userID, in.Name, in.Notes)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return saveTemplateExercises(ctx, tx, id, in.Exercises)
	})
	if err != nil {
		return Template{}, err
	}
	return s.TemplateByID(ctx, userID, id)
}

func saveTemplateExercises(ctx context.Context, tx pgx.Tx, templateID int64, list []TemplateExerciseInput) error {
	keep := []int64{}
	for _, te := range list {
		if te.ID != 0 {
			keep = append(keep, te.ID)
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM template_exercises
		WHERE template_id = $1 AND NOT (id = ANY($2))`, templateID, keep); err != nil {
		return err
	}
	for i, te := range list {
		if te.ID != 0 {
			tag, err := tx.Exec(ctx, `UPDATE template_exercises
				SET exercise_id = $3, position = $4, sets = $5, reps = $6, weight_kg = $7, rest_seconds = $8,
					increment_kg = $9, deload_after = $10, deload_percent = $11,
					-- Changing the weight by hand starts the failure count over.
					failures = CASE WHEN weight_kg = $7 THEN failures ELSE 0 END
				WHERE id = $1 AND template_id = $2`,
				te.ID, templateID, te.ExerciseID, i+1, te.Sets, te.Reps, te.WeightKg, te.RestSeconds,
				te.IncrementKg, te.DeloadAfter, te.DeloadPercent)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return ErrNotFound
			}
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO template_exercises
			(template_id, exercise_id, position, sets, reps, weight_kg, rest_seconds,
			 increment_kg, deload_after, deload_percent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			templateID, te.ExerciseID, i+1, te.Sets, te.Reps, te.WeightKg, te.RestSeconds,
			te.IncrementKg, te.DeloadAfter, te.DeloadPercent); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTemplate removes a template. Workouts started from it are kept.
func (s *Store) DeleteTemplate(ctx context.Context, userID, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM templates WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// StartWorkoutFromTemplate creates a workout with the template's sets filled
// in, ready to be ticked off.
func (s *Store) StartWorkoutFromTemplate(ctx context.Context, userID, templateID int64, at time.Time) (Workout, error) {
	t, err := s.TemplateByID(ctx, userID, templateID)
	if err != nil {
		return Workout{}, err
	}
	var workoutID int64
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO workouts (user_id, name, notes, started_at, template_id)
			VALUES ($1, $2, '', $3, $4) RETURNING id`, userID, t.Name, at, t.ID).Scan(&workoutID); err != nil {
			return err
		}
		pos := 0
		for _, te := range t.Exercises {
			for i := 0; i < te.Sets; i++ {
				pos++
				if _, err := tx.Exec(ctx, `INSERT INTO workout_sets
					(workout_id, exercise_id, position, reps, weight_kg)
					VALUES ($1, $2, $3, $4, $5)`, workoutID, te.ExerciseID, pos, te.Reps, te.WeightKg); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return Workout{}, err
	}
	return s.WorkoutByID(ctx, userID, workoutID)
}

// CopyTemplate duplicates a template with its exercises and progression
// settings. The copy starts without any failed sessions.
func (s *Store) CopyTemplate(ctx context.Context, userID, id int64, name string) (Template, error) {
	src, err := s.TemplateByID(ctx, userID, id)
	if err != nil {
		return Template{}, err
	}
	in := TemplateInput{Name: name, Notes: src.Notes}
	for _, te := range src.Exercises {
		in.Exercises = append(in.Exercises, TemplateExerciseInput{
			ExerciseID: te.ExerciseID, Sets: te.Sets, Reps: te.Reps, WeightKg: te.WeightKg,
			RestSeconds: te.RestSeconds, IncrementKg: te.IncrementKg,
			DeloadAfter: te.DeloadAfter, DeloadPercent: te.DeloadPercent,
		})
	}
	return s.CreateTemplate(ctx, userID, in)
}
