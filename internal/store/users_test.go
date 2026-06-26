package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sorin-work/workout-log/internal/store"
	"github.com/sorin-work/workout-log/internal/testdb"
)

func TestCreateUserRejectsDuplicateEmail(t *testing.T) {
	st := store.New(testdb.New(t))
	ctx := context.Background()
	if _, err := st.CreateUser(ctx, "lifter@example.com", "Lifter", "hash"); err != nil {
		t.Fatal(err)
	}
	_, err := st.CreateUser(ctx, "Lifter@Example.com", "Other", "hash")
	if !errors.Is(err, store.ErrConflict) {
		t.Fatalf("err = %v, want ErrConflict", err)
	}
}

func TestSessionsExpire(t *testing.T) {
	st := store.New(testdb.New(t))
	ctx := context.Background()
	u, err := st.CreateUser(ctx, "a@example.com", "A", "hash")
	if err != nil {
		t.Fatal(err)
	}
	live, old := []byte("live-token-hash"), []byte("old-token-hash")
	if err := st.CreateSession(ctx, live, u.ID, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateSession(ctx, old, u.ID, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if got, err := st.UserBySession(ctx, live); err != nil || got.ID != u.ID {
		t.Fatalf("live session: user %v, err %v", got.ID, err)
	}
	if _, err := st.UserBySession(ctx, old); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expired session: err = %v, want ErrNotFound", err)
	}
	n, err := st.DeleteExpiredSessions(ctx)
	if err != nil || n != 1 {
		t.Fatalf("deleted %d sessions, err %v", n, err)
	}
}

func TestCloseStaleWorkouts(t *testing.T) {
	st := store.New(testdb.New(t))
	ctx := context.Background()
	u, err := st.CreateUser(ctx, "s@example.com", "S", "hash")
	if err != nil {
		t.Fatal(err)
	}
	old, _ := st.CreateWorkout(ctx, u.ID, store.WorkoutInput{Name: "Forgot", StartedAt: time.Now().Add(-30 * time.Hour)})
	fresh, _ := st.CreateWorkout(ctx, u.ID, store.WorkoutInput{Name: "Now", StartedAt: time.Now().Add(-time.Hour)})
	n, err := st.CloseStaleWorkouts(ctx, 12*time.Hour)
	if err != nil || n != 1 {
		t.Fatalf("closed %d, err %v", n, err)
	}
	w, _ := st.WorkoutByID(ctx, u.ID, old.ID)
	if w.FinishedAt == nil || !w.FinishedAt.Equal(w.StartedAt) {
		t.Fatalf("stale workout: %+v", w)
	}
	w, _ = st.WorkoutByID(ctx, u.ID, fresh.ID)
	if w.FinishedAt != nil {
		t.Fatal("recent workout was closed")
	}
}
