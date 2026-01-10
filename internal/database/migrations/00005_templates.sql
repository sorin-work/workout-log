-- +goose Up
CREATE TABLE templates (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    notes      TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX templates_user_idx ON templates (user_id);

CREATE TABLE template_exercises (
    id           BIGSERIAL PRIMARY KEY,
    template_id  BIGINT NOT NULL REFERENCES templates (id) ON DELETE CASCADE,
    exercise_id  BIGINT NOT NULL REFERENCES exercises (id),
    position     INT NOT NULL,
    sets         INT NOT NULL CHECK (sets BETWEEN 1 AND 20),
    reps         INT NOT NULL CHECK (reps BETWEEN 1 AND 100),
    weight_kg    NUMERIC(7, 3) NOT NULL DEFAULT 0 CHECK (weight_kg >= 0),
    rest_seconds INT NOT NULL DEFAULT 120 CHECK (rest_seconds BETWEEN 0 AND 1800)
);

CREATE INDEX template_exercises_template_idx ON template_exercises (template_id, position);

ALTER TABLE workouts ADD COLUMN template_id BIGINT REFERENCES templates (id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE workouts DROP COLUMN template_id;
DROP TABLE template_exercises;
DROP TABLE templates;
