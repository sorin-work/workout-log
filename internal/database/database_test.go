package database_test

import (
	"context"
	"testing"

	"github.com/sorin-work/workout-log/internal/database"
	"github.com/sorin-work/workout-log/internal/testdb"
)

func TestMigrationsCreateTables(t *testing.T) {
	pool := testdb.New(t)
	for _, table := range []string{"users", "sessions", "exercises", "workouts", "workout_sets", "templates", "template_exercises", "programs", "program_days", "exercise_notes"} {
		var exists bool
		err := pool.QueryRow(context.Background(),
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables
			 WHERE table_schema = current_schema() AND table_name = $1)`, table).Scan(&exists)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Errorf("table %s was not created", table)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	pool := testdb.New(t)
	if err := database.Migrate(context.Background(), pool); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
}
