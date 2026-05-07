-- +goose Up
ALTER TABLE users
    ADD COLUMN e1rm_formula TEXT NOT NULL DEFAULT 'epley' CHECK (e1rm_formula IN ('epley', 'brzycki'));

-- +goose Down
ALTER TABLE users DROP COLUMN e1rm_formula;
