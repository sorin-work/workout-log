# Roadmap

Things I might add to workout-log. The app covers the loop I need for barbell
training today (log sets, follow a program, progress, see records and
volume), so nothing here is promised or scheduled. These are ideas, written
down with enough detail that I (or anyone else) can pick one up later without
starting from zero. Items may change, move between groups or be dropped.

What the app does today is in the [README](README.md) and the notes under
[docs/](docs/). It stays a strength-training log: cardio, nutrition and
general health tracking are out of scope, and items that drift that way say
so.

- **Next**: small or well understood, mostly gaps in what already exists. I
  would start with these.
- **Later**: worth doing, but bigger or waiting on a decision.
- **Maybe**: not sure yet that they are worth the extra complexity.

Sizes are rough: **S** is an evening or two, **M** a few evenings up to a
couple of weekends, **L** several weeks of spare time.

Each item is meant to fit in one pull request and to be covered by the test
suite. There are no browser tests. `internal/domain` has plain unit tests for
the training rules, and `internal/store`, `internal/api` and `internal/seed`
run integration tests against PostgreSQL, each test in its own schema built by
`testdb.New` from the real migrations. The web app has Vitest and Testing
Library specs with `fetch` mocked. So rules worth testing (progression,
records, units, rounding, schedules) belong in `internal/domain`, without I/O,
and not in a handler, a store query or a React component; pure web helpers
belong in `web/src/lib`. Most items add columns or tables and bring a goose
migration under `internal/database/migrations` (the next one is
`00011_*.sql`, with `-- +goose Up` and `-- +goose Down`). All weights stay in
kilograms in the database, as `NUMERIC(7,3)`.

When an item ships, its row in the overview says `Shipped` in the Group column,
its section moves under a `## Shipped` heading at the end of this file and
shrinks to one paragraph on what was built and what was left out, and the same
pull request updates the README, the notes under `docs/` and an `Unreleased`
entry in the [CHANGELOG](CHANGELOG.md).

"Depends on" lists real prerequisites only. Items that merely touch the same
file are not listed there; the bigger overlaps are mentioned in the items.

## Overview

| #   | Item                                         | Group | Area        | Size | Depends on |
| --- | -------------------------------------------- | ----- | ----------- | ---- | ---------- |
| 1   | Rate limits that a header cannot dodge       | Next  | security    | S    | -          |
| 2   | Progression in pounds                        | Next  | progression | S    | -          |
| 3   | Progression after editing a finished workout | Next  | progression | M    | -          |
| 4   | Security headers                             | Next  | security    | S    | -          |
| 5   | Scheduled backups and a restore guide        | Next  | ops         | S    | -          |
| 6   | Workouts closed for you                      | Later | progression | S    | 3          |
| 7   | Rep ranges and double progression            | Later | progression | M    | -          |
| 8   | Different targets per set                    | Later | templates   | M    | -          |
| 9   | Percentage programs from a training max      | Later | programs    | L    | 8          |
| 10  | RPE targets                                  | Later | progression | M    | 8          |
| 11  | Timed and distance sets                      | Later | logging     | M    | -          |
| 12  | Body weight and weighted bodyweight lifts    | Later | logging     | M    | -          |
| 13  | Supersets and circuits                       | Later | logging     | M    | -          |
| 14  | One side at a time                           | Later | logging     | S    | -          |
| 15  | Your own bar and plates                      | Later | logging     | S    | -          |
| 16  | A rest timer that works in a pocket          | Later | logging     | S    | -          |
| 17  | Program days on weekdays                     | Later | programs    | S    | -          |
| 18  | Swapping an exercise for a day               | Later | logging     | M    | -          |
| 19  | Notes and flags on a set                     | Later | logging     | S    | -          |
| 20  | Exercise descriptions and links              | Later | exercises   | S    | -          |
| 21  | Import from other apps                       | Later | import      | M    | -          |
| 22  | Full account export and restore              | Later | accounts    | M    | -          |
| 23  | Getting back into a locked account           | Later | accounts    | S    | -          |
| 24  | Closing sign-up on a private server          | Later | accounts    | S    | -          |
| 25  | Record history                               | Later | records     | S    | -          |
| 26  | Comparing training blocks                    | Later | stats       | M    | -          |
| 27  | Logging without a signal                     | Later | logging     | L    | -          |
| 28  | Accessibility pass                           | Later | ui          | M    | -          |
| 29  | Browser end-to-end tests                     | Later | tests       | M    | -          |
| 30  | Metrics and a readiness check                | Later | ops         | S    | -          |
| 31  | Faster records with years of sets            | Maybe | performance | M    | -          |
| 32  | A read-only link for a coach                 | Maybe | sharing     | M    | -          |
| 33  | Coach accounts                               | Maybe | sharing     | L    | 32         |
| 34  | Bands and chains                             | Maybe | logging     | S    | -          |
| 35  | Languages and number formats                 | Maybe | ui          | M    | -          |
| 36  | Two-factor sign-in                           | Maybe | accounts    | M    | -          |

## Next

### 1. Rate limits that a header cannot dodge

`NewServer` in `internal/api/server.go` uses chi's `middleware.RealIP`, which
replaces `r.RemoteAddr` with the value of `True-Client-IP`, `X-Real-IP` or
`X-Forwarded-For` from the request, whoever sent it. The compose setup has no
proxy in front: the Go server is published directly on `WEB_PORT`. The login
and sign-up limiter (`rateLimiter.middleware`, keyed on `r.RemoteAddr`) can
therefore be skipped by sending a different `X-Forwarded-For` on every
request, and the request log records whatever address the client claims.

- Configuration: `TRUSTED_PROXIES`, a list of CIDRs, empty by default. Only
  when the direct peer is in that list are the forwarded headers read, and then
  the rightmost address that is not itself trusted. A small
  `clientIP(r, trusted)` in `internal/api` replaces `middleware.RealIP`.
- The README section that says "set `SECURE_COOKIES=true` when served over
  HTTPS" gets a line about `TRUSTED_PROXIES` for the same setups.
- The limiter itself: login and sign-up share one bucket of 10 a minute per
  address (`loginAttemptsPerMinute`). Add a second key on the email for login
  failures, so one account cannot be guessed at from many addresses, and
  count only failures for it.
- Tricky: `http.Request.RemoteAddr` includes the port; `SplitHostPort` is
  already handled in the middleware, keep it in the new helper.
- Tests: `internal/api/ratelimit_test.go` with a spoofed header from an
  untrusted peer (ignored), from a trusted one (used), and the per-email key.

### 2. Progression in pounds

Weights are stored in kilograms and the web app converts for pound users, but
progression does not know about the unit. `ApplyProgression` in
`internal/store/progression.go` always passes
`domain.SmallestPlateStep(domain.Kilograms)` (2.5 kg) as `RoundingKg`, so a
pound user's deload is rounded to a multiple of 2.5 kg, which is 5.51 lb: a
225 lb squat deloaded by 10% becomes 203.9 lb instead of 205. The increment is
entered in pounds and stored in kilograms (5 lb is 2.268 kg), which is exact
enough on its own, but after a rounded deload the targets stop landing on
loadable weights and the plate calculator shows a remainder.

- `ApplyProgression` takes the user's unit and rounds in that unit: convert
  to pounds, round to `SmallestPlateStep(Pounds)` (5 lb), convert back. A
  `domain.RoundInUnit(kg, unit, step)` helper keeps the conversion in
  `internal/domain` where it can be tested.
- The same helper for increments: after adding the increment, snap to the
  unit's step, so float noise from `ToKg` never accumulates.
- Item 15 replaces the fixed step with the user's own smallest plates; this
  item only makes the existing rule right for pounds.
- Tricky: a user who switches units keeps targets that were rounded in the
  old unit. Do not rewrite them; the next progression rounds in the new unit.
- Tests: `internal/domain/progression_test.go` for deloads and increments in
  pounds; `internal/api/progression_test.go` for a pound user finishing a
  failed session three times.

### 3. Progression after editing a finished workout

Finishing a workout applies progression once (`FinishWorkout` with the
conditional update, then `ApplyProgression`). Everything after that is
ignored: fixing a mistyped rep count on a finished workout, deleting a
finished workout that was logged by mistake (`DeleteWorkout` deletes the row
and nothing else), or adding a forgotten set. The template keeps the target
and failure count the wrong data produced, and the only fix is editing the
template by hand.

- Data: a `progression_changes` table that records what each finish did
  (`workout_id`, `template_exercise_id`, `previous_weight_kg`,
  `previous_failures`, `weight_kg`, `failures`). The `ProgressionChange`
  struct already holds most of it.
- Deleting a finished workout reverts its changes, but only while they are the
  latest for that template exercise. If a later workout already progressed
  from them, say so and leave the targets alone.
- Editing sets of a finished workout whose changes are still the latest
  offers "Recalculate progression": revert, then apply again with the edited
  sets, in one transaction.
- The finish screen (`ProgressionSummary.tsx`) can then offer "Undo" too.
- Tricky: two templates can share an exercise, and a workout can be started
  from a template that was edited since. Store the change against the template
  exercise row, and revert only fields that still hold the values this
  workout wrote.
- Tricky: `workouts.template_id` is `ON DELETE SET NULL`. A deleted template
  has nothing to revert; the history rows go with it (`ON DELETE CASCADE` on
  the template exercise).
- Tests: `internal/api/progression_test.go` for delete and revert, a later
  workout blocking the revert, recalculation after an edit.

### 4. Security headers

The Go server answers every request itself, API and web app, and sets only
`Cache-Control` (`noStore` in `server.go`, and `static.go` for assets). There
is no `Content-Security-Policy`, no `X-Content-Type-Options`, no frame
protection and no HSTS, even when `SECURE_COOKIES=true` says it is behind
HTTPS.

- A `securityHeaders` middleware on the whole router:
  `X-Content-Type-Options: nosniff`, `Referrer-Policy: same-origin`,
  `X-Frame-Options: DENY` plus `frame-ancestors 'none'`, and a CSP of
  `default-src 'self'` with what the Vite build actually needs (check for
  inline styles from the charts; the SVG charts are hand-written, so probably
  `style-src 'self' 'unsafe-inline'` at most).
- HSTS only when `SECURE_COOKIES=true`.
- `Permissions-Policy` that leaves only what the rest timer uses (vibration
  needs nothing; item 16's wake lock and notifications need nothing either).
- Tricky: a CSP that breaks the app is worse than none. Ship it as
  `Content-Security-Policy-Report-Only` first? There is no report endpoint, so
  rather test the built app once with the header on, by hand, and say so in
  the pull request.
- Tests: `internal/api/server_test.go` checks the headers on an API route and
  on a static file, and HSTS only with secure cookies.

### 5. Scheduled backups and a restore guide

A year of training history lives in the `db-data` volume, and the README only
says that `docker compose down -v` starts fresh. The CSV export covers sets but
not templates, programs or notes (item 22).

- A `backup` service in `docker-compose.yml` under a `backup` profile, using
  the `postgres:17-alpine` image: `pg_dump -Fc` once a day into a `./backups`
  bind mount, named by date, deleting dumps older than `BACKUP_KEEP_DAYS`
  (default 30). Copying the folder elsewhere is the user's job.
- `docs/backups.md`: start it, take one by hand, restore into a fresh volume,
  check a dump with `pg_restore --list`.
- Upgrades: goose runs the embedded migrations on start
  (`internal/database`). The guide says to take a dump before upgrading; goose
  has down migrations, but I do not test them against real data.
- Tricky: restoring with the app running lets it migrate or seed a half
  restored database. Stop `app` first.
- Tests: none in Go. A CI step that dumps and restores the test database in
  the `api` job would prove the commands.
