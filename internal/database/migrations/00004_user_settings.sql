-- +goose Up
ALTER TABLE users
    ADD COLUMN unit     TEXT NOT NULL DEFAULT 'kg' CHECK (unit IN ('kg', 'lb')),
    ADD COLUMN timezone TEXT NOT NULL DEFAULT 'UTC';

-- +goose Down
ALTER TABLE users DROP COLUMN unit, DROP COLUMN timezone;
