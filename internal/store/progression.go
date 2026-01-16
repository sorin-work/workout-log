package store

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/sorin-work/workout-log/internal/domain"
)

// ProgressionChange reports what finishing a workout did to one exercise of
// the template it was started from.
type ProgressionChange struct {
	TemplateExerciseID int64          `json:"templateExerciseId"`
	ExerciseID         int64          `json:"exerciseId"`
	ExerciseName       string         `json:"exerciseName"`
	Outcome            domain.Outcome `json:"outcome"`
	PreviousWeightKg   float64        `json:"previousWeightKg"`
	WeightKg           float64        `json:"weightKg"`
	Failures           int            `json:"failures"`
}

// ApplyProgression updates the targets of the template a workout was started
// from, based on the sets logged in it. Workouts without a template, or whose
// template was deleted, change nothing.
func (s *Store) ApplyProgression(ctx context.Context, w Workout) ([]ProgressionChange, error) {
	changes := []ProgressionChange{}
	if w.TemplateID == nil {
		return changes, nil
	}
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		list, err := s.templateExercises(ctx, tx, *w.TemplateID)
		if err != nil {
			return err
		}
		for _, te := range list {
			var logged []domain.SetInput
			for _, st := range w.Sets {
				if st.ExerciseID == te.ExerciseID {
					logged = append(logged, domain.SetInput{
						Reps: st.Reps, WeightKg: st.WeightKg, RPE: st.RPE,
						Completed: st.Completed, Warmup: st.Warmup,
					})
				}
			}
			rule := domain.ProgressionRule{
				IncrementKg:   te.IncrementKg,
				DeloadAfter:   te.DeloadAfter,
				DeloadPercent: te.DeloadPercent,
				RoundingKg:    domain.SmallestPlateStep(domain.Kilograms),
			}
			next, outcome := domain.Progress(rule,
				domain.ProgressionState{WeightKg: te.WeightKg, Failures: te.Failures},
				domain.Target{Sets: te.Sets, Reps: te.Reps}, logged)
			if outcome == domain.OutcomeSkipped {
				continue
			}
			if _, err := tx.Exec(ctx, `UPDATE template_exercises SET weight_kg = $2, failures = $3
				WHERE id = $1`, te.ID, next.WeightKg, next.Failures); err != nil {
				return err
			}
			changes = append(changes, ProgressionChange{
				TemplateExerciseID: te.ID,
				ExerciseID:         te.ExerciseID,
				ExerciseName:       te.ExerciseName,
				Outcome:            outcome,
				PreviousWeightKg:   te.WeightKg,
				WeightKg:           next.WeightKg,
				Failures:           next.Failures,
			})
		}
		return nil
	})
	return changes, err
}
