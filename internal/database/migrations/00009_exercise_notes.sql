-- +goose Up
-- Personal cues per exercise ("grip one finger wider"), also for built-ins.
CREATE TABLE exercise_notes (
    user_id     BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    exercise_id BIGINT NOT NULL REFERENCES exercises (id) ON DELETE CASCADE,
    notes       TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, exercise_id)
);

-- +goose Down
DROP TABLE exercise_notes;
