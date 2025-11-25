-- +goose Up
CREATE TABLE workouts (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    started_at  TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT workouts_finished_after_start CHECK (finished_at IS NULL OR finished_at >= started_at)
);

CREATE INDEX workouts_user_started_idx ON workouts (user_id, started_at DESC);

CREATE TABLE workout_sets (
    id          BIGSERIAL PRIMARY KEY,
    workout_id  BIGINT NOT NULL REFERENCES workouts (id) ON DELETE CASCADE,
    exercise_id BIGINT NOT NULL REFERENCES exercises (id),
    position    INT NOT NULL,
    reps        INT NOT NULL CHECK (reps >= 0),
    -- Weights are always stored in kilograms; the UI converts for lb users.
    weight_kg   NUMERIC(7, 3) NOT NULL DEFAULT 0 CHECK (weight_kg >= 0),
    rpe         NUMERIC(3, 1),
    completed   BOOLEAN NOT NULL DEFAULT false,
    warmup      BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX workout_sets_workout_idx ON workout_sets (workout_id, position);
CREATE INDEX workout_sets_exercise_idx ON workout_sets (exercise_id);

-- +goose Down
DROP TABLE workout_sets;
DROP TABLE workouts;
