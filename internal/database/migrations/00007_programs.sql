-- +goose Up
CREATE TABLE programs (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    active        BOOLEAN NOT NULL DEFAULT false,
    -- Index into program_days of the workout that comes next.
    next_position INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- At most one program is followed at a time.
CREATE UNIQUE INDEX programs_one_active_idx ON programs (user_id) WHERE active;

CREATE TABLE program_days (
    id          BIGSERIAL PRIMARY KEY,
    program_id  BIGINT NOT NULL REFERENCES programs (id) ON DELETE CASCADE,
    position    INT NOT NULL,
    template_id BIGINT NOT NULL REFERENCES templates (id) ON DELETE CASCADE,
    UNIQUE (program_id, position)
);

-- +goose Down
DROP TABLE program_days;
DROP TABLE programs;
