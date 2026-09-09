package api

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
)

const maxImportBytes = 5 << 20

type importResult struct {
	Workouts int `json:"workouts"`
	Sets     int `json:"sets"`
	Skipped  int `json:"skipped"`
}

// handleImportSets reads a CSV in the export format. Each distinct date and
// workout name becomes one workout; workouts that already exist with the
// same start time and name are skipped, so importing a file twice is safe.
func (s *Server) handleImportSets(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r.Context())
	loc := userLocation(u.Timezone)
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBytes)
	cr := csv.NewReader(r.Body)
	cr.FieldsPerRecord = -1
	header, err := cr.Read()
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "The file is empty or not a CSV file.")
		return
	}
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))] = i
	}
	for _, need := range []string{"date", "workout", "exercise", "reps", "weight_kg"} {
		if _, ok := col[need]; !ok {
			writeError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("Missing column %q.", need))
			return
		}
	}
	exercises, err := s.store.ListExercises(r.Context(), u.ID, store.ExerciseFilter{IncludeArchived: true})
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	byName := map[string]int64{}
	for _, e := range exercises {
		byName[strings.ToLower(e.Name)] = e.ID
	}

	get := func(rec []string, name string) string {
		if i, ok := col[name]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}
	var workouts []store.ImportedWorkout
	index := map[string]int{}
	fields := map[string]string{}
	line := 1
	for {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		line++
		if err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", fmt.Sprintf("Line %d is not valid CSV.", line))
			return
		}
		key := fmt.Sprintf("line %d", line)
		at, err := time.ParseInLocation(time.DateTime, get(rec, "date"), loc)
		if err != nil {
			fields[key] = "Date must look like 2026-09-09 18:30:00."
			continue
		}
		exID, ok := byName[strings.ToLower(strings.TrimPrefix(get(rec, "exercise"), "'"))]
		if !ok {
			fields[key] = fmt.Sprintf("Unknown exercise %q. Create it first.", get(rec, "exercise"))
			continue
		}
		reps, err1 := strconv.Atoi(get(rec, "reps"))
		kg, err2 := strconv.ParseFloat(get(rec, "weight_kg"), 64)
		if err1 != nil || err2 != nil {
			fields[key] = "Reps and weight_kg must be numbers."
			continue
		}
		set := store.SetData{ExerciseID: exID, Reps: reps, WeightKg: kg, Completed: true}
		if v := get(rec, "rpe"); v != "" {
			rpe, err := strconv.ParseFloat(v, 64)
			if err != nil {
				fields[key] = "RPE must be a number."
				continue
			}
			set.RPE = &rpe
		}
		if v := get(rec, "warmup"); v != "" {
			set.Warmup, _ = strconv.ParseBool(v)
		}
		if v := get(rec, "completed"); v != "" {
			set.Completed, _ = strconv.ParseBool(v)
		}
		if f := domain.ValidateSet(domain.SetInput{Reps: set.Reps, WeightKg: set.WeightKg, RPE: set.RPE,
			Completed: set.Completed, Warmup: set.Warmup}); len(f) > 0 {
			for _, msg := range f {
				fields[key] = msg
				break
			}
			continue
		}
		name := strings.TrimPrefix(get(rec, "workout"), "'")
		if name == "" {
			name = "Workout"
		}
		k := at.UTC().Format(time.RFC3339) + "\x00" + name
		i, ok := index[k]
		if !ok {
			i = len(workouts)
			index[k] = i
			workouts = append(workouts, store.ImportedWorkout{Name: name, StartedAt: at})
		}
		workouts[i].Sets = append(workouts[i].Sets, set)
		if len(fields) > 20 {
			break
		}
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}
	created, sets, skipped, err := s.store.ImportWorkouts(r.Context(), u.ID, workouts)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, importResult{Workouts: created, Sets: sets, Skipped: skipped})
}
