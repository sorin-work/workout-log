-- +goose Up
-- Records, progress and "last time" look up completed sets of one exercise
-- across a user's workouts.
CREATE INDEX workout_sets_exercise_workout_idx ON workout_sets (exercise_id, workout_id) WHERE completed;
DROP INDEX workout_sets_exercise_idx;

-- Progression and "delete template" look up workouts by template.
CREATE INDEX workouts_template_idx ON workouts (template_id) WHERE template_id IS NOT NULL;

-- Open workouts are closed by a timer.
CREATE INDEX workouts_open_idx ON workouts (started_at) WHERE finished_at IS NULL;

-- +goose Down
DROP INDEX workouts_open_idx;
DROP INDEX workouts_template_idx;
CREATE INDEX workout_sets_exercise_idx ON workout_sets (exercise_id);
DROP INDEX workout_sets_exercise_workout_idx;
