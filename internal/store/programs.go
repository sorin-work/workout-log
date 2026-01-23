package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Program is an ordered rotation of templates, e.g. A/B/A, B/A/B.
type Program struct {
	ID           int64        `json:"id"`
	Name         string       `json:"name"`
	Active       bool         `json:"active"`
	NextPosition int          `json:"nextPosition"`
	Days         []ProgramDay `json:"days"`
}

// ProgramDay is one slot of the rotation.
type ProgramDay struct {
	Position     int    `json:"position"`
	TemplateID   int64  `json:"templateId"`
	TemplateName string `json:"templateName"`
}

// ProgramInput is the editable part of a program.
type ProgramInput struct {
	Name        string
	TemplateIDs []int64
}

// ListPrograms returns the user's programs, the active one first.
func (s *Store) ListPrograms(ctx context.Context, userID int64) ([]Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, active, next_position FROM programs
		WHERE user_id = $1 ORDER BY active DESC, lower(name), id`, userID)
	if err != nil {
		return nil, err
	}
	list, err := pgx.CollectRows(rows, scanProgram)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}
	ids := make([]int64, len(list))
	byID := map[int64]*Program{}
	for i := range list {
		ids[i] = list[i].ID
		byID[list[i].ID] = &list[i]
	}
	rows, err = s.pool.Query(ctx, `SELECT pd.program_id, pd.position, pd.template_id, t.name
		FROM program_days pd JOIN templates t ON t.id = pd.template_id
		WHERE pd.program_id = ANY($1) ORDER BY pd.program_id, pd.position`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var programID int64
		var d ProgramDay
		if err := rows.Scan(&programID, &d.Position, &d.TemplateID, &d.TemplateName); err != nil {
			return nil, err
		}
		p := byID[programID]
		p.Days = append(p.Days, d)
	}
	return list, rows.Err()
}

func scanProgram(row pgx.CollectableRow) (Program, error) {
	p := Program{Days: []ProgramDay{}}
	err := row.Scan(&p.ID, &p.Name, &p.Active, &p.NextPosition)
	return p, err
}

// ProgramByID returns one of the user's programs.
func (s *Store) ProgramByID(ctx context.Context, userID, id int64) (Program, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, active, next_position FROM programs
		WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return Program{}, err
	}
	p, err := pgx.CollectExactlyOneRow(rows, scanProgram)
	if err != nil {
		return Program{}, notFound(err)
	}
	rows, err = s.pool.Query(ctx, `SELECT pd.position, pd.template_id, t.name
		FROM program_days pd JOIN templates t ON t.id = pd.template_id
		WHERE pd.program_id = $1 ORDER BY pd.position`, id)
	if err != nil {
		return Program{}, err
	}
	p.Days, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (ProgramDay, error) {
		var d ProgramDay
		err := row.Scan(&d.Position, &d.TemplateID, &d.TemplateName)
		return d, err
	})
	return p, err
}

// CreateProgram stores a new, inactive program.
func (s *Store) CreateProgram(ctx context.Context, userID int64, in ProgramInput) (Program, error) {
	var id int64
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO programs (user_id, name) VALUES ($1, $2) RETURNING id`,
			userID, in.Name).Scan(&id); err != nil {
			return err
		}
		return saveProgramDays(ctx, tx, id, in.TemplateIDs)
	})
	if err != nil {
		return Program{}, err
	}
	return s.ProgramByID(ctx, userID, id)
}

// UpdateProgram renames a program and replaces its rotation.
func (s *Store) UpdateProgram(ctx context.Context, userID, id int64, in ProgramInput) (Program, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE programs SET name = $3, next_position = next_position % $4
			WHERE id = $1 AND user_id = $2`, id, userID, in.Name, max(len(in.TemplateIDs), 1))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		if _, err := tx.Exec(ctx, `DELETE FROM program_days WHERE program_id = $1`, id); err != nil {
			return err
		}
		return saveProgramDays(ctx, tx, id, in.TemplateIDs)
	})
	if err != nil {
		return Program{}, err
	}
	return s.ProgramByID(ctx, userID, id)
}

func saveProgramDays(ctx context.Context, tx pgx.Tx, programID int64, templateIDs []int64) error {
	for i, tid := range templateIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO program_days (program_id, position, template_id)
			VALUES ($1, $2, $3)`, programID, i, tid); err != nil {
			return err
		}
	}
	return nil
}

// DeleteProgram removes a program. Its templates are kept.
func (s *Store) DeleteProgram(ctx context.Context, userID, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM programs WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ActivateProgram makes a program the one being followed.
func (s *Store) ActivateProgram(ctx context.Context, userID, id int64) (Program, error) {
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE programs SET active = false WHERE user_id = $1 AND active AND id <> $2`,
			userID, id); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE programs SET active = true WHERE id = $1 AND user_id = $2`, id, userID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return nil
	})
	if err != nil {
		return Program{}, err
	}
	return s.ProgramByID(ctx, userID, id)
}

// StartProgramDay starts the next workout of a program and moves the
// rotation on by one.
func (s *Store) StartProgramDay(ctx context.Context, userID, id int64, at time.Time) (Workout, error) {
	p, err := s.ProgramByID(ctx, userID, id)
	if err != nil {
		return Workout{}, err
	}
	if len(p.Days) == 0 {
		return Workout{}, ErrNotFound
	}
	day := p.Days[p.NextPosition%len(p.Days)]
	w, err := s.StartWorkoutFromTemplate(ctx, userID, day.TemplateID, at)
	if err != nil {
		return Workout{}, err
	}
	_, err = s.pool.Exec(ctx, `UPDATE programs SET next_position = (next_position + 1) % $2 WHERE id = $1`,
		id, len(p.Days))
	return w, err
}
