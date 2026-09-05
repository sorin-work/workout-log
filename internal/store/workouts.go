package store

import (
	"context"
	"sort"
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
	// Exercises lists what was trained, in order, with the heaviest set.
	Exercises []ExerciseSummary `json:"exercises"`
}

// ExerciseSummary is one exercise of a workout in the history list.
type ExerciseSummary struct {
	ExerciseID int64   `json:"exerciseId"`
	Name       string  `json:"name"`
	Sets       int     `json:"sets"`
	TopReps    int     `json:"topReps"`
	TopKg      float64 `json:"topKg"`
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

// WorkoutFilter narrows ListWorkouts.
type WorkoutFilter struct {
	// ExerciseID keeps only workouts with at least one set of the exercise.
	ExerciseID int64
}

// ListWorkouts returns a page of the user's workouts, newest first, with a
// summary of what was done in each, and the total number of workouts.
func (s *Store) ListWorkouts(ctx context.Context, userID int64, f WorkoutFilter, limit, offset int) ([]WorkoutSummary, int, error) {
	const match = `w.user_id = $1 AND ($2 = 0 OR EXISTS (
		SELECT 1 FROM workout_sets x WHERE x.workout_id = w.id AND x.exercise_id = $2))`
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM workouts w WHERE `+match, userID, f.ExerciseID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := s.pool.Query(ctx, `SELECT w.id, w.name, w.started_at, w.finished_at,
			count(DISTINCT ws.exercise_id),
			count(ws.id) FILTER (WHERE ws.completed AND NOT ws.warmup),
			coalesce(sum(ws.reps * ws.weight_kg) FILTER (WHERE ws.completed AND NOT ws.warmup), 0)::float8
		FROM workouts w LEFT JOIN workout_sets ws ON ws.workout_id = w.id
		WHERE `+match+`
		GROUP BY w.id
		ORDER BY w.started_at DESC, w.id DESC
		LIMIT $3 OFFSET $4`, userID, f.ExerciseID, limit, offset)
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
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if err := s.attachExerciseSummaries(ctx, out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// attachExerciseSummaries fills in the exercises of every workout in the
// page with one query, instead of one query per workout.
func (s *Store) attachExerciseSummaries(ctx context.Context, list []WorkoutSummary) error {
	if len(list) == 0 {
		return nil
	}
	ids := make([]int64, len(list))
	byID := make(map[int64]*WorkoutSummary, len(list))
	for i := range list {
		ids[i] = list[i].ID
		list[i].Exercises = []ExerciseSummary{}
		byID[list[i].ID] = &list[i]
	}
	rows, err := s.pool.Query(ctx, `SELECT DISTINCT ON (ws.workout_id, ws.exercise_id) ws.workout_id, ws.exercise_id, e.name,
			count(*) OVER (PARTITION BY ws.workout_id, ws.exercise_id), ws.reps, ws.weight_kg::float8,
			min(ws.position) OVER (PARTITION BY ws.workout_id, ws.exercise_id) AS first_pos
		FROM workout_sets ws JOIN exercises e ON e.id = ws.exercise_id
		WHERE ws.workout_id = ANY($1) AND ws.completed AND NOT ws.warmup
		ORDER BY ws.workout_id, ws.exercise_id, ws.weight_kg DESC, ws.reps DESC`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	firstPos := map[[2]int64]int{}
	for rows.Next() {
		var workoutID int64
		var e ExerciseSummary
		var pos int
		if err := rows.Scan(&workoutID, &e.ExerciseID, &e.Name, &e.Sets, &e.TopReps, &e.TopKg, &pos); err != nil {
			return err
		}
		firstPos[[2]int64{workoutID, e.ExerciseID}] = pos
		w := byID[workoutID]
		w.Exercises = append(w.Exercises, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range list {
		ex := list[i].Exercises
		id := list[i].ID
		sort.Slice(ex, func(a, b int) bool {
			return firstPos[[2]int64{id, ex[a].ExerciseID}] < firstPos[[2]int64{id, ex[b].ExerciseID}]
		})
	}
	return nil
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

// FinishWorkout stamps the finish time of a workout. finished is false when
// the workout was already finished, e.g. a double click or a retried
// request; only the call that actually finishes it should apply progression.
// The conditional update makes concurrent calls race safely: the row lock
// lets exactly one of them see finished_at IS NULL.
func (s *Store) FinishWorkout(ctx context.Context, userID, id int64, at time.Time) (w Workout, finished bool, err error) {
	tag, err := s.pool.Exec(ctx, `UPDATE workouts SET finished_at = greatest($3, started_at)
		WHERE id = $1 AND user_id = $2 AND finished_at IS NULL`, id, userID, at)
	if err != nil {
		return Workout{}, false, err
	}
	w, err = s.WorkoutByID(ctx, userID, id)
	if err != nil {
		return Workout{}, false, err
	}
	return w, tag.RowsAffected() == 1, nil
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

// CloseStaleWorkouts finishes workouts that were left open for longer than
// maxOpen, using the time of their last logged set (or the start time) as
// the finish time. Progression is not applied to these.
func (s *Store) CloseStaleWorkouts(ctx context.Context, maxOpen time.Duration) (int64, error) {
	tag, err := s.pool.Exec(ctx, `UPDATE workouts w
		SET finished_at = greatest(w.started_at, coalesce(
			(SELECT max(ws.created_at) FROM workout_sets ws WHERE ws.workout_id = w.id), w.started_at))
		WHERE w.finished_at IS NULL AND w.started_at < now() - $1::interval`, maxOpen.String())
	return tag.RowsAffected(), err
}
