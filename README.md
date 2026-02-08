# workout-log

A strength-training log for lifts, templates, programs and progress.

- Exercise library with 40+ built-in lifts and your own custom exercises
- Log workouts set by set: weight, reps, RPE, warm-ups
- Templates and programs with linear progression and automatic deloads
- Personal records and estimated one-rep maxes, with progress charts
- Kilograms or pounds

## Quick start with Docker

```bash
docker compose up --build
```

Open http://localhost:4413 and log in with the demo account
`demo@example.com` / `demo-password`.

## Local development

You need Go 1.25+, Node 22+ and PostgreSQL.

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
