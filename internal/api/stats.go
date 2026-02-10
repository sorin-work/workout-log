package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sorin-work/workout-log/internal/domain"
)

const (
	defaultVolumeWeeks = 12
	maxVolumeWeeks     = 104
)

// userLocation returns the user's time zone, falling back to UTC.
func userLocation(tz string) *time.Location {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (s *Server) handleWeeklyVolume(w http.ResponseWriter, r *http.Request) {
	weeks := defaultVolumeWeeks
	if v := r.URL.Query().Get("weeks"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxVolumeWeeks {
			writeError(w, http.StatusBadRequest, "bad_request", "weeks must be between 1 and 104.")
			return
		}
		weeks = n
	}
	u := currentUser(r.Context())
	loc := userLocation(u.Timezone)
	now := s.now()
	from := domain.WeekStart(now, loc).AddDate(0, 0, -7*(weeks-1))
	rows, err := s.store.ExerciseWeeks(r.Context(), u.ID, from)
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.WeeklyVolume(rows, from, now, loc))
}
