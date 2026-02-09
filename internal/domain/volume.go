package domain

import (
	"sort"
	"time"
)

// SecondaryShare is how much a set counts towards a secondary muscle. A
// bench press set is one set for the chest and half a set for the triceps.
const SecondaryShare = 0.5

// ExerciseWeek is the completed working sets of one exercise in one week.
type ExerciseWeek struct {
	WeekStart        time.Time
	PrimaryMuscle    string
	SecondaryMuscles []string
	Sets             int
	VolumeKg         float64
}

// MuscleVolume is the training volume of one muscle group in a week.
type MuscleVolume struct {
	Sets     float64 `json:"sets"`
	VolumeKg float64 `json:"volumeKg"`
}

// WeekVolume is the volume per muscle group for the week starting on
// WeekStart (a Monday).
type WeekVolume struct {
	WeekStart string                  `json:"weekStart"`
	Muscles   map[string]MuscleVolume `json:"muscles"`
	TotalSets float64                 `json:"totalSets"`
}

// WeekStart returns the Monday 00:00 of the week containing t, in loc.
func WeekStart(t time.Time, loc *time.Location) time.Time {
	t = t.In(loc)
	offset := (int(t.Weekday()) + 6) % 7 // Monday = 0
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
	return d.AddDate(0, 0, -offset)
}

// WeeklyVolume spreads exercise sets over muscle groups and returns one
// entry per week from the week of from to the week of to, including weeks
// without training. Weeks are keyed by their Monday as YYYY-MM-DD.
func WeeklyVolume(rows []ExerciseWeek, from, to time.Time, loc *time.Location) []WeekVolume {
	byWeek := map[string]*WeekVolume{}
	var keys []string
	for w := WeekStart(from, loc); !w.After(to); w = w.AddDate(0, 0, 7) {
		k := w.Format(time.DateOnly)
		byWeek[k] = &WeekVolume{WeekStart: k, Muscles: map[string]MuscleVolume{}}
		keys = append(keys, k)
	}
	for _, r := range rows {
		wv := byWeek[r.WeekStart.Format(time.DateOnly)]
		if wv == nil {
			continue
		}
		p := wv.Muscles[r.PrimaryMuscle]
		p.Sets += float64(r.Sets)
		p.VolumeKg += r.VolumeKg
		wv.Muscles[r.PrimaryMuscle] = p
		wv.TotalSets += float64(r.Sets)
		for _, m := range r.SecondaryMuscles {
			if m == r.PrimaryMuscle {
				continue
			}
			s := wv.Muscles[m]
			s.Sets += float64(r.Sets) * SecondaryShare
			wv.Muscles[m] = s
		}
	}
	sort.Strings(keys)
	out := make([]WeekVolume, len(keys))
	for i, k := range keys {
		out[i] = *byWeek[k]
	}
	return out
}
