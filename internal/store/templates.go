package store

import (
	"context"

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
	ID          int64
	ExerciseID  int64
	Sets        int
	Reps        int
	WeightKg    float64
	RestSeconds int
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
	for i := range list {
		if list[i].Exercises, err = s.templateExercises(ctx, s.pool, list[i].ID); err != nil {
			return nil, err
		}
	}
	return list, nil
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
	rows, err := q.Query(ctx, `SELECT te.id, te.exercise_id, e.name, te.position, te.sets, te.reps,
			te.weight_kg::float8, te.rest_seconds
		FROM template_exercises te JOIN exercises e ON e.id = te.exercise_id
		WHERE te.template_id = $1 ORDER BY te.position, te.id`, templateID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (TemplateExercise, error) {
		var te TemplateExercise
		err := row.Scan(&te.ID, &te.ExerciseID, &te.ExerciseName, &te.Position, &te.Sets, &te.Reps, &te.WeightKg, &te.RestSeconds)
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
				SET exercise_id = $3, position = $4, sets = $5, reps = $6, weight_kg = $7, rest_seconds = $8
				WHERE id = $1 AND template_id = $2`,
				te.ID, templateID, te.ExerciseID, i+1, te.Sets, te.Reps, te.WeightKg, te.RestSeconds)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return ErrNotFound
			}
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO template_exercises
			(template_id, exercise_id, position, sets, reps, weight_kg, rest_seconds)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			templateID, te.ExerciseID, i+1, te.Sets, te.Reps, te.WeightKg, te.RestSeconds); err != nil {
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
