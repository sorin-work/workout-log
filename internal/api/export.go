package api

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sorin-work/workout-log/internal/store"
)

var exportHeader = []string{"date", "workout", "exercise", "set", "reps", "weight_kg", "rpe", "warmup", "completed"}

// csvSafe stops spreadsheet apps from treating user text as a formula.
func csvSafe(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

func (s *Server) handleExportSets(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r.Context())
	loc := userLocation(u.Timezone)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="workout-log-`+s.now().In(loc).Format("2006-01-02")+`.csv"`)
	cw := csv.NewWriter(w)
	_ = cw.Write(exportHeader)
	err := s.store.ExportSets(r.Context(), u.ID, func(row store.ExportRow) error {
		rpe := ""
		if row.RPE != nil {
			rpe = strconv.FormatFloat(*row.RPE, 'f', -1, 64)
		}
		return cw.Write([]string{
			row.StartedAt.In(loc).Format(time.DateTime),
			csvSafe(row.WorkoutName),
			csvSafe(row.ExerciseName),
			strconv.Itoa(row.Position),
			strconv.Itoa(row.Reps),
			strconv.FormatFloat(row.WeightKg, 'f', -1, 64),
			rpe,
			strconv.FormatBool(row.Warmup),
			strconv.FormatBool(row.Completed),
		})
	})
	cw.Flush()
	if err != nil {
		// Headers are already sent; all we can do is log and cut the file short.
		writeInternalLogOnly(r, err)
	}
}
