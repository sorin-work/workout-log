# workout-log

A strength-training log for lifts, templates, programs and progress. It is
made for the barbell and dumbbell work itself: sets, reps, weight and RPE,
the program you follow and whether the numbers go up. It is not a general
health or cardio tracker.

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

## Stack

| Part     | Tech                                                              |
| -------- | ----------------------------------------------------------------- |
| API      | Go 1.26, chi, pgx, goose migrations (embedded), slog              |
| Database | PostgreSQL 17                                                     |
| Web      | React 19, TypeScript, Vite, React Router, hand-written SVG charts |
| Tests    | `go test` against a real PostgreSQL, Vitest + Testing Library     |

## Quick start with Docker

```bash
docker compose up --build
```

Open http://localhost:4413 and log in with the demo account:

- **Email:** `demo@example.com`
- **Password:** `demo-password`

The demo user is created on first start with a year of Beginner 5x5 history
(two templates in an A/B program with progression and the odd deload), a
custom exercise and some notes. Set `SEED_DEMO=false` to start empty.

| Service | URL / port            | Notes                                          |
| ------- | --------------------- | ---------------------------------------------- |
| app     | http://localhost:4413 | Go server: API under `/api` and the web app    |
| db      | `127.0.0.1:5432`      | PostgreSQL 17, user, password and db `workout` |

Change the ports with `WEB_PORT` and `DB_PORT`, e.g.
`WEB_PORT=8080 docker compose up`. Data lives in the `db-data` volume;
`docker compose down -v` starts fresh.

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

`make` wraps the common commands:

| Command         | What it does                                         |
| --------------- | ---------------------------------------------------- |
| `make db`       | start PostgreSQL in Docker and create `workout_test` |
| `make run`      | run the API                                          |
| `make web-dev`  | run the Vite dev server                              |
| `make test`     | Go unit and integration tests                        |
| `make test-web` | web tests                                            |
| `make lint`     | `gofmt` and `go vet`                                 |
| `make lint-web` | ESLint, Prettier and TypeScript                      |
| `make build`    | build the web app and the binary into `bin/`         |

## Tests

```bash
make db
export TEST_DATABASE_URL=postgres://workout:workout@localhost:5432/workout_test?sslmode=disable
go test ./...
cd web && npm test          # or npm run test:coverage
```

- `internal/domain`: unit tests for units, estimated maxes, progression,
  deloads, records, weekly volume, warm-ups and streaks.
- `internal/store`, `internal/api`, `internal/seed`: integration tests
  against PostgreSQL. Each test gets its own schema with all migrations, so
  packages run in parallel; they are skipped when `TEST_DATABASE_URL` is not
  set.
- `web/`: Vitest with Testing Library for helpers, components and pages,
  with `fetch` mocked.

CI (`.github/workflows/ci.yml`) runs formatting, vet, all tests with a
PostgreSQL service, the web lint/type/test/build steps and a Docker build.

## Project layout

```
cmd/workout-log/   entry point
internal/api/      HTTP handlers and routes
internal/domain/   training logic without I/O
internal/store/    SQL queries
internal/database/ connection and migrations (migrations/*.sql)
internal/seed/     demo data
web/               React app
docs/              API reference and architecture notes
```

See [docs/api.md](docs/api.md) for the endpoints and
[docs/architecture.md](docs/architecture.md) for how it fits together.

## License

MIT
