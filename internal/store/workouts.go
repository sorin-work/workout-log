package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// Workout is a training session with its logged sets.
type Workout struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Notes      string     `json:"notes"`
	StartedAt  time.Time  `json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	TemplateID *int64     `json:"templateId"`
	Sets       []Set      `json:"sets"`
}

// WorkoutSummary is a workout in the history list.
type WorkoutSummary struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	StartedAt     time.Time  `json:"startedAt"`
	FinishedAt    *time.Time `json:"finishedAt"`
	ExerciseCount int        `json:"exerciseCount"`
	SetCount      int        `json:"setCount"`
	VolumeKg      float64    `json:"volumeKg"`
}

// Set is one logged set.
type Set struct {
	ID           int64    `json:"id"`
	ExerciseID   int64    `json:"exerciseId"`
	ExerciseName string   `json:"exerciseName"`
	Position     int      `json:"position"`
	Reps         int      `json:"reps"`
	WeightKg     float64  `json:"weightKg"`
	RPE          *float64 `json:"rpe"`
	Completed    bool     `json:"completed"`
	Warmup       bool     `json:"warmup"`
}

// WorkoutInput holds the editable fields of a workout.
type WorkoutInput struct {
	Name       string
	Notes      string
	StartedAt  time.Time
	FinishedAt *time.Time
}

const workoutColumns = `id, name, notes, started_at, finished_at, template_id`

func scanWorkout(row interface{ Scan(...any) error }) (Workout, error) {
	var w Workout
	err := row.Scan(&w.ID, &w.Name, &w.Notes, &w.StartedAt, &w.FinishedAt, &w.TemplateID)
	w.Sets = []Set{}
	return w, err
}

// CreateWorkout starts a new workout for the user.
func (s *Store) CreateWorkout(ctx context.Context, userID int64, in WorkoutInput) (Workout, error) {
	return scanWorkout(s.pool.QueryRow(ctx, `INSERT INTO workouts (user_id, name, notes, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING `+workoutColumns,
		userID, in.Name, in.Notes, in.StartedAt, in.FinishedAt))
}

// WorkoutByID returns one of the user's workouts with its sets in order.
func (s *Store) WorkoutByID(ctx context.Context, userID, id int64) (Workout, error) {
	w, err := scanWorkout(s.pool.QueryRow(ctx, `SELECT `+workoutColumns+` FROM workouts
		WHERE id = $1 AND user_id = $2`, id, userID))
	if err != nil {
		return Workout{}, notFound(err)
	}
	w.Sets, err = s.setsForWorkout(ctx, id)
	return w, err
}

func (s *Store) setsForWorkout(ctx context.Context, workoutID int64) ([]Set, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+setColumns+`
		FROM workout_sets ws JOIN exercises e ON e.id = ws.exercise_id
		WHERE ws.workout_id = $1 ORDER BY ws.position, ws.id`, workoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sets := []Set{}
	for rows.Next() {
		st, err := scanSet(rows)
		if err != nil {
			return nil, err
		}
		sets = append(sets, st)
	}
	return sets, rows.Err()
}

const setColumns = `ws.id, ws.exercise_id, e.name, ws.position, ws.reps, ws.weight_kg::float8, ws.rpe::float8, ws.completed, ws.warmup`

func scanSet(row interface{ Scan(...any) error }) (Set, error) {
	var st Set
	err := row.Scan(&st.ID, &st.ExerciseID, &st.ExerciseName, &st.Position, &st.Reps, &st.WeightKg, &st.RPE, &st.Completed, &st.Warmup)
	return st, err
}

// ListWorkouts returns a page of the user's workouts, newest first, with a
// summary of what was done in each, and the total number of workouts.
func (s *Store) ListWorkouts(ctx context.Context, userID int64, limit, offset int) ([]WorkoutSummary, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM workouts WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `SELECT w.id, w.name, w.started_at, w.finished_at,
			count(DISTINCT ws.exercise_id),
			count(ws.id) FILTER (WHERE ws.completed AND NOT ws.warmup),
			coalesce(sum(ws.reps * ws.weight_kg) FILTER (WHERE ws.completed AND NOT ws.warmup), 0)::float8
		FROM workouts w LEFT JOIN workout_sets ws ON ws.workout_id = w.id
		WHERE w.user_id = $1
		GROUP BY w.id
		ORDER BY w.started_at DESC, w.id DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []WorkoutSummary{}
	for rows.Next() {
		var w WorkoutSummary
		if err := rows.Scan(&w.ID, &w.Name, &w.StartedAt, &w.FinishedAt, &w.ExerciseCount, &w.SetCount, &w.VolumeKg); err != nil {
			return nil, 0, err
		}
		out = append(out, w)
	}
	return out, total, rows.Err()
}

// UpdateWorkout changes the name, notes and times of a workout.
func (s *Store) UpdateWorkout(ctx context.Context, userID, id int64, in WorkoutInput) (Workout, error) {
	_, err := scanWorkout(s.pool.QueryRow(ctx, `UPDATE workouts
		SET name = $3, notes = $4, started_at = $5, finished_at = $6
		WHERE id = $1 AND user_id = $2 RETURNING `+workoutColumns,
		id, userID, in.Name, in.Notes, in.StartedAt, in.FinishedAt))
	if err != nil {
		return Workout{}, notFound(err)
	}
	return s.WorkoutByID(ctx, userID, id)
}

// DeleteWorkout removes a workout and its sets.
func (s *Store) DeleteWorkout(ctx context.Context, userID, id int64) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM workouts WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FinishWorkout stamps the finish time of a workout.
func (s *Store) FinishWorkout(ctx context.Context, userID, id int64, at time.Time) (Workout, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE workouts SET finished_at = greatest($3, started_at)
		WHERE id = $1 AND user_id = $2`, id, userID, at)
	if err != nil {
		return Workout{}, err
	}
	if tag.RowsAffected() == 0 {
		return Workout{}, ErrNotFound
	}
	return s.WorkoutByID(ctx, userID, id)
}

// RepeatWorkout starts a new workout with the same exercises, reps and
// weights as an earlier one. Nothing is ticked off yet.
func (s *Store) RepeatWorkout(ctx context.Context, userID, id int64, at time.Time) (Workout, error) {
	src, err := s.WorkoutByID(ctx, userID, id)
	if err != nil {
		return Workout{}, err
	}
	var newID int64
	err = pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO workouts (user_id, name, started_at, template_id)
			VALUES ($1, $2, $3, $4) RETURNING id`, userID, src.Name, at, src.TemplateID).Scan(&newID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO workout_sets (workout_id, exercise_id, position, reps, weight_kg, warmup)
			SELECT $2, exercise_id, position, reps, weight_kg, warmup
			FROM workout_sets WHERE workout_id = $1`, id, newID)
		return err
	})
	if err != nil {
		return Workout{}, err
	}
	return s.WorkoutByID(ctx, userID, newID)
}
