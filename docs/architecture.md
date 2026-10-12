# Architecture

```
cmd/workout-log/      main: config, database, migrations, demo seed, HTTP server
internal/
  api/                chi routes, handlers, auth middleware, rate limiting, static files
  auth/               bcrypt passwords and session tokens
  config/             environment variables
  database/           pgx pool and embedded goose migrations
  domain/             pure training logic: units, estimated 1RM, progression,
                      records, weekly volume, warm-ups, summaries
  seed/               demo account with a year of history
  store/              SQL queries (pgx), one file per area
  testdb/             isolated schema per test for integration tests
web/                  React 19 + Vite single-page app
```

## Request flow

The Go binary serves both the JSON API under `/api` and, when `WEB_DIR` is
set, the built web app with a fallback to `index.html` for client-side
routes. Handlers decode and validate the request, call the store, and use
`domain` for anything that is not plain storage (progression, records,
volume). The store returns `ErrNotFound` for rows that do not exist **or
belong to someone else**, so ownership checks live in the SQL `WHERE`
clauses rather than in handlers.

## Data

All weights are stored in kilograms as `NUMERIC(7,3)`; the web app converts
for users who prefer pounds. Timestamps are `timestamptz`; anything that
groups by day or week (calendar, weekly volume, summary, streak) uses the
user's time zone from their profile.

- `users`, `sessions`: accounts and hashed session tokens.
- `exercises`: the built-in library (`user_id IS NULL`) plus custom ones.
- `workouts`, `workout_sets`: what was done; a workout may point at the
  template it was started from.
- `templates`, `template_exercises`: plans with target sets, reps and weight,
  plus progression settings and the current failure count.
- `programs`, `program_days`: an ordered rotation of templates.
- `exercise_notes`: personal cues per user and exercise.

## Progression

`POST /api/workouts/{id}/finish` sets `finished_at` only if it is still
empty. The request that wins that conditional update applies progression in
a transaction (`store.ApplyProgression` with `domain.Progress`); repeated or
concurrent finishes do nothing more.

## Records and charts

Records, progress series and "new PR" detection are computed in Go from the
user's completed working sets (`store.LoggedSets`), using the formula the
user picked. This keeps the rules in one tested place (`domain`) at the cost
of reading all sets of a user, which is fine for one person's training log.

## Tests

`internal/domain` has plain unit tests. Store and API tests run against a
real PostgreSQL: `testdb.New` creates a schema per test, runs the migrations
and drops it afterwards, so packages run in parallel. The web app uses
Vitest with Testing Library and a mocked `fetch`.
