package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sorin-work/workout-log/internal/domain"
	"github.com/sorin-work/workout-log/internal/store"
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
	rows, err := s.store.ExerciseWeeks(r.Context(), u.ID, from, loc.String())
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.WeeklyVolume(rows, from, now, loc))
}

func (s *Server) handleSummary(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r.Context())
	loc := userLocation(u.Timezone)
	// Everything since sign-up, so the total count is right; it is one row
	// per workout, which stays small.
	days, err := s.store.WorkoutDays(r.Context(), u.ID, time.Time{})
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, domain.Summarize(days, s.now(), loc))
}

type calendarDay struct {
	Date     string             `json:"date"`
	Workouts []store.WorkoutRef `json:"workouts"`
}

// handleCalendar lists the days of a month that have workouts, using the
// user's time zone to decide the day.
func (s *Server) handleCalendar(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r.Context())
	loc := userLocation(u.Timezone)
	month := r.URL.Query().Get("month")
	if month == "" {
		month = s.now().In(loc).Format("2006-01")
	}
	start, err := time.ParseInLocation("2006-01", month, loc)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "month must look like 2026-06.")
		return
	}
	refs, err := s.store.WorkoutsBetween(r.Context(), u.ID, start, start.AddDate(0, 1, 0))
	if err != nil {
		writeInternal(w, r, err)
		return
	}
	days := []calendarDay{}
	for _, ref := range refs {
		d := ref.StartedAt.In(loc).Format(time.DateOnly)
		if len(days) == 0 || days[len(days)-1].Date != d {
			days = append(days, calendarDay{Date: d})
		}
		last := &days[len(days)-1]
		last.Workouts = append(last.Workouts, ref)
	}
	writeJSON(w, http.StatusOK, days)
}
