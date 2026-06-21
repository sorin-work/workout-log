package seed_test

import (
	"context"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/seed"
	"github.com/sorin-work/workout-log/internal/store"
	"github.com/sorin-work/workout-log/internal/testdb"
)

func TestDemoSeed(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	now := time.Date(2026, 2, 7, 12, 0, 0, 0, time.UTC)
	created, err := seed.Demo(ctx, pool, now)
	if err != nil || !created {
		t.Fatalf("created=%v err=%v", created, err)
	}
	again, err := seed.Demo(ctx, pool, now)
	if err != nil || again {
		t.Fatalf("second run should be a no-op: created=%v err=%v", again, err)
	}

	st := store.New(pool)
	u, err := st.UserByEmail(ctx, seed.DemoEmail)
	if err != nil {
		t.Fatal(err)
	}
	page, total, err := st.ListWorkouts(ctx, u.ID, store.WorkoutFilter{}, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total < 120 {
		t.Fatalf("only %d demo workouts", total)
	}
	if page[0].StartedAt.After(now) {
		t.Fatalf("demo workout in the future: %v", page[0].StartedAt)
	}
	tpls, err := st.ListTemplates(ctx, u.ID)
	if err != nil || len(tpls) != 2 {
		t.Fatalf("templates: %v %v", tpls, err)
	}
	for _, te := range tpls[0].Exercises {
		if te.ExerciseName == "Back Squat" && te.WeightKg <= 60 {
			t.Fatalf("squat never progressed: %+v", te)
		}
	}
	progs, err := st.ListPrograms(ctx, u.ID)
	if err != nil || len(progs) != 1 || !progs[0].Active {
		t.Fatalf("programs: %+v %v", progs, err)
	}
}
