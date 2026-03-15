package api

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestExportSetsCSV(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("csv@example.com")
	c.mustDo(http.MethodPut, "/api/me", map[string]any{"unit": "kg", "timezone": "Europe/Berlin"}, http.StatusOK, nil)
	ex := createExercise(t, c, "=HYPERLINK(1)", "quads")
	started := time.Date(2026, 3, 10, 17, 30, 0, 0, time.UTC)
	w := createWorkout(t, c, "Legs, heavy", started)
	addSet(t, c, w.ID, exerciseID(t, c, "Back Squat"), 5, 102.5, true)
	c.mustDo(http.MethodPost, fmt.Sprintf("/api/workouts/%d/sets", w.ID), map[string]any{
		"exerciseId": ex.ID, "reps": 10, "weightKg": 60, "rpe": 8.5,
	}, http.StatusCreated, nil)

	res := c.mustDo(http.MethodGet, "/api/export/sets.csv", nil, http.StatusOK, nil)
	records, err := csv.NewReader(strings.NewReader(string(res.Body))).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || records[0][0] != "date" {
		t.Fatalf("csv: %v", records)
	}
	first := records[1]
	if first[0] != "2026-03-10 18:30:00" || first[1] != "Legs, heavy" || first[5] != "102.5" || first[8] != "true" {
		t.Fatalf("first row: %v", first)
	}
	second := records[2]
	if second[2] != "'=HYPERLINK(1)" || second[6] != "8.5" || second[8] != "false" {
		t.Fatalf("second row: %v", second)
	}
}
