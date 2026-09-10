package api

import (
	"net/http"
	"testing"
)

const importCSV = `date,workout,exercise,set,reps,weight_kg,rpe,warmup,completed
2026-09-01 18:30:00,Day A,Back Squat,1,5,20,,true,true
2026-09-01 18:30:00,Day A,Back Squat,2,5,100,8,false,true
2026-09-01 18:30:00,Day A,bench press,3,5,80,,false,true
2026-09-03 18:30:00,Day B,Deadlift,1,5,140,9.5,false,true
`

func TestImportSets(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("imp@example.com")
	var res importResult
	r := c.upload("/api/import/sets.csv", "text/csv", importCSV)
	if r.Status != http.StatusOK {
		t.Fatalf("status %d: %s", r.Status, r.Body)
	}
	r.decode(t, &res)
	if res != (importResult{Workouts: 2, Sets: 4}) {
		t.Fatalf("result: %+v", res)
	}
	var page workoutPage
	c.mustDo(http.MethodGet, "/api/workouts", nil, http.StatusOK, &page)
	if page.Total != 2 || page.Items[1].Name != "Day A" || page.Items[1].SetCount != 2 || page.Items[1].FinishedAt == nil {
		t.Fatalf("imported: %+v", page.Items)
	}

	// Importing the same file again changes nothing.
	c.upload("/api/import/sets.csv", "text/csv", importCSV).decode(t, &res)
	if res != (importResult{Skipped: 2}) {
		t.Fatalf("second import: %+v", res)
	}
}

func TestExportThenImportRoundTrip(t *testing.T) {
	a := newTestAPI(t)
	src := a.signup("src@example.com")
	src.upload("/api/import/sets.csv", "text/csv", importCSV)
	exported := src.mustDo(http.MethodGet, "/api/export/sets.csv", nil, http.StatusOK, nil)

	dst := a.signup("dst@example.com")
	var res importResult
	dst.upload("/api/import/sets.csv", "text/csv", string(exported.Body)).decode(t, &res)
	if res.Workouts != 2 || res.Sets != 4 {
		t.Fatalf("round trip: %+v", res)
	}
}

func TestImportReportsBadLines(t *testing.T) {
	a := newTestAPI(t)
	c := a.signup("imp@example.com")
	bad := `date,workout,exercise,reps,weight_kg,rpe
2026-09-01,Day A,Back Squat,5,100,
2026-09-01 18:30:00,Day A,Unicorn Press,5,100,
2026-09-01 18:30:00,Day A,Back Squat,five,100,
2026-09-01 18:30:00,Day A,Back Squat,5,100,12
`
	r := c.upload("/api/import/sets.csv", "text/csv", bad)
	if r.Status != http.StatusUnprocessableEntity {
		t.Fatalf("status %d: %s", r.Status, r.Body)
	}
	f := errorFields(t, r)
	for _, k := range []string{"line 2", "line 3", "line 4", "line 5"} {
		if f[k] == "" {
			t.Errorf("no error for %s: %v", k, f)
		}
	}
	var page workoutPage
	c.mustDo(http.MethodGet, "/api/workouts", nil, http.StatusOK, &page)
	if page.Total != 0 {
		t.Fatal("a file with errors must not import anything")
	}
	r = c.upload("/api/import/sets.csv", "text/csv", "name,reps\nx,1\n")
	if r.Status != http.StatusBadRequest {
		t.Fatalf("missing columns: %d", r.Status)
	}
}
