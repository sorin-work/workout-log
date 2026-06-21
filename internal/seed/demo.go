// Package seed creates a demo account with realistic training history.
package seed

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sorin-work/workout-log/internal/auth"
	"github.com/sorin-work/workout-log/internal/domain"
)

// Demo account credentials, documented in the README.
const (
	DemoEmail    = "demo@example.com"
	DemoPassword = "demo-password"
)

// Weeks of history generated for the demo user.
const demoWeeks = 52

type lift struct {
	name      string
	sets      int
	reps      int
	startKg   float64
	increment float64
	warmups   bool
}

var dayA = []lift{
	{"Back Squat", 5, 5, 60, 2.5, true},
	{"Bench Press", 5, 5, 50, 2.5, true},
	{"Barbell Row", 5, 5, 50, 2.5, false},
	{"Dumbbell Curl", 3, 10, 12, 0, false},
}

var dayB = []lift{
	{"Back Squat", 5, 5, 60, 2.5, true},
	{"Overhead Press", 5, 5, 35, 2.5, true},
	{"Deadlift", 1, 5, 80, 5, true},
	{"Pull-Up", 3, 8, 0, 0, false},
	{"Hanging Leg Raise", 3, 12, 0, 0, false},
	{"Landmine Press", 3, 10, 20, 0, false},
}

// Demo creates the demo user unless it already exists. now anchors the
// generated history so the most recent workout is a day or two ago.
func Demo(ctx context.Context, pool *pgxpool.Pool, now time.Time) (created bool, err error) {
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE lower(email) = $1)`,
		DemoEmail).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}
	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return false, err
	}
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		return generate(ctx, tx, hash, now)
	})
	return err == nil, err
}

type generator struct {
	ctx       context.Context
	tx        pgx.Tx
	userID    int64
	exercises map[string]int64
	rng       *rand.Rand
	loc       *time.Location
}

func generate(ctx context.Context, tx pgx.Tx, hash string, now time.Time) error {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return err
	}
	g := &generator{ctx: ctx, tx: tx, exercises: map[string]int64{}, rng: rand.New(rand.NewPCG(4413, 2026)), loc: loc}
	if err := tx.QueryRow(ctx, `INSERT INTO users (email, display_name, password_hash, unit, timezone)
		VALUES ($1, 'Demo Lifter', $2, 'kg', 'Europe/Berlin') RETURNING id`, DemoEmail, hash).Scan(&g.userID); err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT id, name FROM exercises WHERE user_id IS NULL`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		g.exercises[name] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// A custom exercise, so the demo shows one in the library.
	var landmine int64
	if err := tx.QueryRow(ctx, `INSERT INTO exercises (user_id, name, primary_muscle, secondary_muscles, equipment)
		VALUES ($1, 'Landmine Press', 'shoulders', '{chest,triceps}', 'barbell') RETURNING id`, g.userID).Scan(&landmine); err != nil {
		return err
	}
	g.exercises["Landmine Press"] = landmine
	if _, err := tx.Exec(ctx, `INSERT INTO exercise_notes (user_id, exercise_id, notes)
		VALUES ($1, $2, 'Pinkies on the rings, feet flat, slight arch.')`, g.userID, g.exercises["Bench Press"]); err != nil {
		return err
	}

	tplA, err := g.template("Day A", dayA)
	if err != nil {
		return err
	}
	tplB, err := g.template("Day B", dayB)
	if err != nil {
		return err
	}

	state := map[string]*domain.ProgressionState{}
	for _, l := range append(append([]lift{}, dayA...), dayB...) {
		state[l.name] = &domain.ProgressionState{WeightKg: l.startKg}
	}

	// Monday, Wednesday and Friday evenings, oldest first, ending before today.
	today := time.Date(now.In(loc).Year(), now.In(loc).Month(), now.In(loc).Day(), 0, 0, 0, 0, loc)
	start := today.AddDate(0, 0, -7*demoWeeks)
	n := 0
	for d := start; d.Before(today); d = d.AddDate(0, 0, 1) {
		wd := d.Weekday()
		if wd != time.Monday && wd != time.Wednesday && wd != time.Friday {
			continue
		}
		if g.rng.IntN(12) == 0 { // a missed session now and then
			continue
		}
		day, tplID, name := dayA, tplA, "Day A"
		if n%2 == 1 {
			day, tplID, name = dayB, tplB, "Day B"
		}
		n++
		startedAt := d.Add(18*time.Hour + time.Duration(g.rng.IntN(60))*time.Minute)
		if err := g.workout(name, tplID, startedAt, day, state); err != nil {
			return err
		}
	}

	// Leave the templates where the simulated lifter is now.
	for _, tpl := range []struct {
		id    int64
		lifts []lift
	}{{tplA, dayA}, {tplB, dayB}} {
		for _, l := range tpl.lifts {
			st := state[l.name]
			if _, err := tx.Exec(ctx, `UPDATE template_exercises SET weight_kg = $3, failures = $4
				WHERE template_id = $1 AND exercise_id = $2`, tpl.id, g.exercises[l.name], st.WeightKg, st.Failures); err != nil {
				return err
			}
		}
	}

	var programID int64
	if err := tx.QueryRow(ctx, `INSERT INTO programs (user_id, name, active, next_position)
		VALUES ($1, 'Beginner 5x5', true, $2) RETURNING id`, g.userID, n%2).Scan(&programID); err != nil {
		return err
	}
	for i, id := range []int64{tplA, tplB} {
		if _, err := tx.Exec(ctx, `INSERT INTO program_days (program_id, position, template_id)
			VALUES ($1, $2, $3)`, programID, i, id); err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) template(name string, lifts []lift) (int64, error) {
	var id int64
	if err := g.tx.QueryRow(g.ctx, `INSERT INTO templates (user_id, name) VALUES ($1, $2) RETURNING id`,
		g.userID, name).Scan(&id); err != nil {
		return 0, err
	}
	for i, l := range lifts {
		exID, ok := g.exercises[l.name]
		if !ok {
			return 0, fmt.Errorf("seed: unknown exercise %q", l.name)
		}
		deloadAfter, deloadPct := 0, 0.0
		if l.increment > 0 {
			deloadAfter, deloadPct = 3, 10
		}
		if _, err := g.tx.Exec(g.ctx, `INSERT INTO template_exercises
			(template_id, exercise_id, position, sets, reps, weight_kg, increment_kg, deload_after, deload_percent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			id, exID, i+1, l.sets, l.reps, l.startKg, l.increment, deloadAfter, deloadPct); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (g *generator) workout(name string, tplID int64, startedAt time.Time, lifts []lift, state map[string]*domain.ProgressionState) error {
	finishedAt := startedAt.Add(time.Duration(55+g.rng.IntN(35)) * time.Minute)
	var workoutID int64
	if err := g.tx.QueryRow(g.ctx, `INSERT INTO workouts (user_id, name, notes, started_at, finished_at, template_id)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		g.userID, name, g.note(), startedAt, finishedAt, tplID).Scan(&workoutID); err != nil {
		return err
	}
	batch := &pgx.Batch{}
	pos := 0
	add := func(exID int64, reps int, kg float64, rpe *float64, completed, warmup bool) {
		pos++
		batch.Queue(`INSERT INTO workout_sets (workout_id, exercise_id, position, reps, weight_kg, rpe, completed, warmup)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, workoutID, exID, pos, reps, kg, rpe, completed, warmup)
	}
	for _, l := range lifts {
		exID := g.exercises[l.name]
		st := state[l.name]
		if l.warmups && st.WeightKg >= 40 {
			add(exID, 5, 20, nil, true, true)
			add(exID, 3, domain.RoundTo(st.WeightKg*0.6, 2.5), nil, true, true)
		}
		logged := make([]domain.SetInput, 0, l.sets)
		// Harder as the weight climbs: the chance of missing reps grows.
		missChance := 0.04 + st.WeightKg/2500
		for i := 0; i < l.sets; i++ {
			reps := l.reps
			if l.startKg == 0 {
				reps = l.reps - 2 + g.rng.IntN(5)
			} else if g.rng.Float64() < missChance {
				reps = l.reps - 1 - g.rng.IntN(2)
			}
			var rpe *float64
			if i == l.sets-1 && l.increment > 0 {
				v := 7 + float64(g.rng.IntN(6))/2
				if reps < l.reps {
					v = 10
				}
				rpe = &v
			}
			add(exID, reps, st.WeightKg, rpe, true, false)
			logged = append(logged, domain.SetInput{Reps: reps, WeightKg: st.WeightKg, Completed: true})
		}
		if l.increment > 0 {
			rule := domain.ProgressionRule{IncrementKg: l.increment, DeloadAfter: 3, DeloadPercent: 10, RoundingKg: 2.5}
			next, _ := domain.Progress(rule, *st, domain.Target{Sets: l.sets, Reps: l.reps}, logged)
			*st = next
		}
	}
	return g.tx.SendBatch(g.ctx, batch).Close()
}

var notes = []string{
	"Felt strong today.",
	"Short on sleep, kept it moving.",
	"Gym was packed, waited for the rack.",
	"Knees a bit stiff in the warm-up, fine after.",
	"Great session!\nTry a wider grip on rows next time.",
}

func (g *generator) note() string {
	if g.rng.IntN(5) != 0 {
		return ""
	}
	return notes[g.rng.IntN(len(notes))]
}
