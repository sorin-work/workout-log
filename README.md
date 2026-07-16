# workout-log

A strength-training log for lifts, templates, programs and progress.

- Exercise library with 40+ built-in lifts, your own custom exercises and
  personal notes per exercise
- Log workouts set by set: weight, reps, RPE and warm-ups, with a rest timer,
  plate calculator, suggested warm-ups and what you did last time
- Templates and programs that rotate through them, with linear progression
  and automatic deloads
- Personal records, estimated one-rep maxes (Epley or Brzycki) and progress
  charts per exercise
- Weekly sets per muscle group, a training calendar and a week streak
- Kilograms or pounds, your own time zone, CSV export

## How progression works

Each exercise in a template can have an increment, e.g. +2.5 kg. When you
finish a workout started from that template:

1. If every planned set was completed with the planned reps at the planned
   weight (or more), the next target goes up by the increment.
2. Otherwise the miss is counted. After _deload after_ misses in a row the
   target drops by the deload percentage, rounded to the nearest 2.5 kg, and
   the count starts over.
3. Exercises you skipped entirely are left alone.

The finish screen shows what changed and any new personal records.

## Quick start with Docker

```bash
docker compose up --build
```

Open http://localhost:4413 and log in with the demo account
`demo@example.com` / `demo-password`.

## Configuration

The server is configured with environment variables:

| Variable         | Default | Description                                                  |
| ---------------- | ------- | ------------------------------------------------------------ |
| `DATABASE_URL`   |         | PostgreSQL connection string (required)                      |
| `PORT`           | `8080`  | HTTP port                                                    |
| `WEB_DIR`        |         | Directory with the built web app to serve; API only if empty |
| `LOG_LEVEL`      | `info`  | `debug`, `info`, `warn` or `error` (JSON logs on stdout)     |
| `SESSION_TTL`    | `720h`  | How long a login lasts                                       |
| `SECURE_COOKIES` | `false` | Set to `true` when served over HTTPS                         |
| `SEED_DEMO`      | `false` | Create the demo account on startup (`true` in compose)       |

Docker compose also reads `WEB_PORT` (default `4413`), `DB_PORT` (`5432`)
and `LOG_LEVEL`.

## Local development

You need Go 1.26+, Node 22+ and PostgreSQL.

```bash
# database
docker compose up -d db   # or use your own PostgreSQL

# API on :8080 (migrations run on startup)
export DATABASE_URL=postgres://workout:workout@localhost:5432/workout?sslmode=disable
go run ./cmd/workout-log

# web app on :5173, proxies /api to :8080
cd web && npm install && npm run dev
```

## Tests

```bash
export TEST_DATABASE_URL=postgres://workout:workout@localhost:5432/workout_test?sslmode=disable
go test ./...
cd web && npm test
```

Integration tests create a throwaway schema per test and are skipped when
`TEST_DATABASE_URL` is not set.

## License

MIT
