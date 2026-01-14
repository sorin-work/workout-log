-- +goose Up
ALTER TABLE template_exercises
    ADD COLUMN increment_kg   NUMERIC(6, 3) NOT NULL DEFAULT 0 CHECK (increment_kg >= 0),
    ADD COLUMN deload_after   INT NOT NULL DEFAULT 0 CHECK (deload_after >= 0),
    ADD COLUMN deload_percent NUMERIC(4, 1) NOT NULL DEFAULT 0 CHECK (deload_percent BETWEEN 0 AND 50),
    -- Failed sessions in a row at the current weight.
    ADD COLUMN failures       INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE template_exercises
    DROP COLUMN increment_kg,
    DROP COLUMN deload_after,
    DROP COLUMN deload_percent,
    DROP COLUMN failures;
