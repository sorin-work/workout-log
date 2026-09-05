package store_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/sorin-work/workout-log/internal/store"
	"github.com/sorin-work/workout-log/internal/testdb"
)

// queryCounter counts queries sent to the database.
type queryCounter struct{ n atomic.Int64 }

func (q *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	q.n.Add(1)
	return ctx
}

func (q *queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestWorkoutHistoryUsesAFixedNumberOfQueries(t *testing.T) {
	counter := &queryCounter{}
	st := store.New(testdb.New(t, testdb.WithTracer(counter)))
	ctx := context.Background()
	u, err := st.CreateUser(ctx, "n1@example.com", "N", "hash")
	if err != nil {
		t.Fatal(err)
	}
	exercises, err := st.ListExercises(ctx, u.ID, store.ExerciseFilter{})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		w, err := st.CreateWorkout(ctx, u.ID, store.WorkoutInput{Name: "W", StartedAt: time.Now().Add(-time.Duration(i) * time.Hour)})
		if err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 3; j++ {
			if _, err := st.AddSet(ctx, w.ID, store.SetData{ExerciseID: exercises[j].ID, Reps: 5, WeightKg: float64(50 + j), Completed: true}); err != nil {
				t.Fatal(err)
			}
		}
	}

	counter.n.Store(0)
	page, _, err := st.ListWorkouts(ctx, u.ID, store.WorkoutFilter{}, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if n := counter.n.Load(); n > 3 {
		t.Fatalf("listing 20 workouts took %d queries", n)
	}
	if len(page) != 20 || len(page[0].Exercises) != 3 || page[0].Exercises[0].ExerciseID != exercises[0].ID {
		t.Fatalf("page: %d workouts, first has %+v", len(page), page[0].Exercises)
	}
}
