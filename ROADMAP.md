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

## Later

### 6. Workouts closed for you

On startup and then periodically, `main.go` calls `CloseStaleWorkouts` with
12 hours: workouts left open that long are finished at the time of their last
set. The comment says progression is not applied to these. Forgetting to tap
"Finish" after a good session therefore means the targets silently stay where
they were, and next time the app asks for the same weight again without
saying why.

- Mark such workouts: `auto_closed BOOLEAN` on `workouts`, set by
  `CloseStaleWorkouts`.
- The home page and the workout page show "This workout was closed for you.
  Apply progression?" with the changes it would make, computed with the same
  `domain.Progress` without writing.
- Applying uses the same path as a normal finish, guarded so it happens once,
  and records its changes in item 3's table, so it can be undone.
- Only for the latest workout of that template: if a newer one was finished
  since, applying the older one would progress from stale targets. Offer it
  only when nothing newer exists.
- Depends on item 3 for the change records.
- Tests: `internal/store/workouts_test.go` marks auto-closed workouts; an API
  test applies progression once and refuses when a newer workout exists.

### 7. Rep ranges and double progression

A template exercise has a fixed `reps` target and progression adds weight
when every set hits it. Most programs after the beginner stage use a range
instead ("3x8-12"): add reps each session at the same weight, and add weight
once every set reaches the top of the range, dropping back to the bottom.

- Data: `reps_max` on `template_exercises`, nullable. When set, `reps` is
  the bottom of the range.
- `domain.Progress` gets a second mode: success means every working set
  reached `reps_max`; then the weight goes up by the increment and the target
  reps go back to `reps`. Otherwise the target reps for next time are the
  lowest reps achieved plus one, capped at `reps_max`. Failures and deloads
  count sessions below the bottom of the range.
- The template editor shows "8-12"; the workout page shows the current rep
  target and the range.
- `SessionSucceeded` stays for fixed targets; the new rule lives next to it,
  with a `ProgressionState` that also remembers the current rep target
  (`target_reps` on the row).
- Tests: `internal/domain/progression_test.go` for a series of sessions
  climbing a range, the reset after a weight increase, and a deload.

### 8. Different targets per set

Every set of an exercise in a template has the same reps and weight
(`template_exercises.sets`, `reps`, `weight_kg`). Common structures cannot be
written down: a top set and back-off sets, pyramids, a last set "as many reps
as possible", or 5/3/1's 65/75/85% sets.

- Data: `template_sets` (`template_exercise_id`, `position`, `reps`,
  `weight_kg` or `percent` of the exercise's working weight, `amrap`). An
  exercise without rows keeps the current "N x reps at weight" behaviour, so
  old templates need no migration of their data.
- Starting a workout from the template (`handleStartTemplate`) creates the
  sets from these rows.
- Progression keeps working on the exercise's working weight; the per-set
  weights follow from their percentages. `SessionSucceeded` compares each
  logged set with its own target, and an AMRAP set counts as hit at its
  minimum reps.
- The editor in `TemplateEditPage.tsx` gets an "Edit sets" mode per exercise;
  the plain mode stays the default.
- Tricky: percentages of what? Of the template's working weight here; item 9
  adds a training max. Keep the field name generic (`percent`) and the base
  explicit (`percent_of`).
- Tests: `internal/domain` tests for success with mixed targets and AMRAP;
  API test for starting a workout with per-set targets.

### 9. Percentage programs from a training max

Popular intermediate programs (5/3/1, GZCLP's later stages, many powerlifting
blocks) are written as percentages of a training max that is raised every
cycle, not as a weight that goes up after each good session.

- Data: `training_maxes` (`user_id`, `exercise_id`, `weight_kg`,
  `updated_at`) and, on `programs`, a `cycle_weeks` and a list of weeks, each
  pointing at templates whose sets use `percent_of = 'training_max'`
  (item 8).
- After the last week of a cycle, the program raises each training max by its
  increment (the same `increment_kg` column), and the finish screen says so.
- "Set training max" from the exercise page: type it, or take 90% of the
  current estimated max (`EstimateOneRepMax`).
- Ship one built-in program definition (5/3/1 with its standard weeks) as a
  template that users copy, like the demo's 5x5, so nobody has to type the
  percentages.
- Depends on item 8.
- Tricky: programs today are a rotation with `next_position`. A cycle with
  weeks is a second kind of program; keep both and make the difference
  visible in `ProgramsPage.tsx`.
- Tests: `internal/domain` tests for the set weights of each week rounded to
  loadable weights, and the training max raise at the end of a cycle.

### 10. RPE targets

Sets record an RPE, and that is all it does: the estimated max ignores it, and
templates cannot ask for "a single at RPE 8, then 3x5 at 85% of it".

- An RPE-based estimate in `internal/domain/e1rm.go`: the usual reps-in-reserve
  table (reps plus 10 minus RPE as effective reps, for RPE 6 to 10), used when
  the set has an RPE and the user chose it as a third formula.
- Per-set targets from item 8 get an optional `rpe`. The workout page shows
  "@8" next to the target, and the weight suggestion for back-off sets is
  computed from the logged top set.
- Progression for RPE sets: none automatic. The user picks the weight; the app
  only suggests. That keeps `domain.Progress` untouched.
- Depends on item 8.
- Tests: `e1rm_test.go` for the RPE table, including RPE missing and RPE
  below 6 (not used).

### 11. Timed and distance sets

The library has "Plank" and "Hanging Leg Raise", and people do loaded carries
and sled pushes, but a set is only reps and weight (`workout_sets.reps` and
`weight_kg`). A plank is logged as "1 rep" or as seconds in the reps field,
and records then call 60 seconds "60 reps".

- Data: `measure` on `exercises` (`reps`, `time`, `distance`) and
  `duration_seconds` and `distance_m` on `workout_sets` and
  `template_exercises`, nullable.
- The set row (`SetRow.tsx`) shows the right fields for the exercise; the rest
  timer can count a timed set down.
- Records for timed exercises: longest hold, and heaviest for a given time.
  `ComputeRecords` skips non-rep exercises today only because `Reps <= 0`; make
  that explicit.
- Weekly volume counts sets for every kind, tonnage only for rep sets.
- Built-in exercises: set `Plank` to `time`; add a farmer's carry as
  `distance`.
- Tests: domain tests for records per measure; API validation of the fields
  per measure.

### 12. Body weight and weighted bodyweight lifts

Pull-ups, chin-ups and dips are in the library as `bodyweight` equipment, and
their weight is whatever was typed: 0 for body weight, or the added weight on
a belt. Their estimated max and tonnage are then meaningless, and a lighter
lifter's 10 pull-ups look like nothing.

- A small body weight log (`body_weights`: `user_id`, `measured_on`,
  `weight_kg`), entered on the home page or at the start of a workout. Not a
  health tracker: one number, used for the lifts.
- For exercises with `bodyweight` equipment, the logged `weight_kg` means
  added weight. Records, estimated maxes and tonnage use body weight on that
  date plus the added weight, labelled as such.
- Assisted variations (band or machine) as negative added weight, if it comes
  for free; otherwise later.
- Tricky: no body weight logged. Fall back to the added weight only, as
  today, and say "add your body weight for better numbers".
- Tests: domain tests for effective load with and without a body weight;
  records for a weighted pull-up.

### 13. Supersets and circuits

Exercises in a workout are a flat ordered list of sets (`position`).
Alternating two exercises with one rest, or a circuit of four, works by
remembering the order, and the rest timer starts after every set.

- Data: `group_key` on `template_exercises` and `workout_sets` (a small int,
  null for none). Exercises with the same key form a superset.
- The workout page shows grouped exercises together and moves through them
  round by round; the rest timer runs after the last exercise of a round
  only.
- `groupSets.ts` in `web/src/lib` already groups sets by exercise; the
  grouping by superset goes there too, with specs.
- Reordering (`handleReorderSets`) keeps groups together.
- Tests: `groupSets.test.ts` for rounds; API test for starting a template
  with a superset.

### 14. One side at a time

Bulgarian split squats, one-arm rows and walking lunges are logged as one set,
so it is not clear whether 10 reps means each side or in total.

- `unilateral` on `exercises` (set for the built-ins that are one-sided).
- Logging: reps are per side; the workout page says so. Optionally log left
  and right separately when they differ (`reps_left`, `reps_right`), shown
  only when the user expands the row.
- Volume and tonnage count both sides.
- Tests: domain test for tonnage of a unilateral set.

### 15. Your own bar and plates

The plate calculator uses a fixed 20 kg or 45 lb bar and a fixed plate set
(`BAR_WEIGHT` and `PLATES` in `web/src/lib/plates.ts`), and progression rounds
to 2.5 kg or 5 lb (`SmallestPlateStep`). A home gym with 0.5 kg microplates, a
15 kg bar, or dumbbells that go up in 2 kg steps gets suggestions it cannot
load.

- Data: `bar_kg` and `plates_kg` (an array) on `users`, and an optional
  `increment_step_kg` per exercise for dumbbells and machines.
- Settings: pick the bar and tick the plates you own.
- The plate calculator, warm-up suggestions (`internal/domain/warmup.go`) and
  progression rounding use them; item 2's helper takes the step from here.
- Tricky: warm-ups are computed on the server and plates on the client. Send
  the settings with the profile, and keep one rounding rule, in
  `internal/domain`, mirrored by a spec on the web side.
- Tests: domain tests for rounding with microplates; `plates.test.ts` for a
  custom set.

### 16. A rest timer that works in a pocket

The rest timer (`RestTimer.tsx`) beeps and vibrates through `lib/alert.ts`,
which only works while the page is in the foreground. With the phone locked
between sets, which is how most people rest, nothing happens.

- Keep the screen on during a workout with the Screen Wake Lock API, as a
  setting (it costs battery).
- When the timer is started, ask once for notification permission and
  schedule a notification at the end. Without a service worker, a
  notification from a background tab is unreliable; item 27's service worker
  makes it dependable.
- Show the remaining time in the tab title, as a fallback that always works.
- Tests: specs for the title and for the wake lock being requested and
  released, with the browser APIs mocked.

### 17. Program days on weekdays

A program is a rotation (`program_days` with `next_position`): the next
workout is whatever comes next, whenever you train. Many people plan by
weekday ("Monday A, Wednesday B, Friday A"), and the calendar cannot show
what is planned.

- Optional weekdays on `programs` (a `days` array). With them, "today's
  workout" on the home page says whether today is a training day, and the
  calendar (`MonthCalendar.tsx`) shows planned days as outlines.
- The rotation still decides which template is next; weekdays only decide
  when. A missed day does not skip a template.
- The week streak is unchanged.
- Tests: `internal/domain` test for the next planned date from a weekday set
  in the user's time zone, including a week boundary.

### 18. Swapping an exercise for a day

When the squat rack is taken, the workout page can add another exercise, but
the template's squat stays in the workout with no sets, and progression
counts it as skipped. The replacement does not progress at all.

- "Swap" on an exercise in a workout started from a template: pick another
  exercise; the planned sets move over with the same reps, and the weight
  comes from the substitute's last session (the `/last` endpoint already
  exists).
- Data: `replaces_exercise_id` on `workout_sets`, so the history shows what
  was swapped.
- Progression leaves the original alone (skipped, as now) and does not touch
  the substitute's template targets, since it has none.
- Optional per-exercise default substitutes in the exercise notes page.
- Tests: API test for a swap, progression skipping the original.

### 19. Notes and flags on a set

Workouts have notes and exercises have personal notes, but a single set
cannot say "belt", "paused", "left knee hurt" or "spotter helped". Those are
what people look back for when a number looks odd.

- `note` (short text) on `workout_sets`, and a few fixed flags as a small
  `TEXT[]`: `belt`, `paused`, `assisted`, `pain`.
- Shown in the history and on the exercise page; records can skip assisted
  sets.
- The CSV export and import carry both.
- Tests: validation of flags; records skip assisted sets.

### 20. Exercise descriptions and links

Built-in exercises have a name, muscles and equipment. A new lifter does not
know what a Pendlay row is, and custom exercises have nowhere to keep a setup
cue other than the personal notes.

- `description` and `url` on `exercises`. The built-ins get a short
  description in the migration (two or three sentences, written here, not
  copied from elsewhere). Links are the user's own, for custom exercises.
- No embedded videos and no hosted media: a link opens in a new tab.
- Tests: API returns and updates the fields; a link must be `https://`.

### 21. Import from other apps

The CSV import (`handleImportSets` in `internal/api/import.go`) reads only the
app's own export format, with `date`, `workout`, `exercise`, `reps` and
`weight_kg`. People moving from Strong, Hevy or FitNotes have years of history
in those apps' CSV exports.

- A `format` parameter and one parser per format in a new
  `internal/importer` package (pure functions from CSV rows to
  `(date, workout name, exercise, reps, weight, unit, rpe, note)`), with sample
  files under `testdata`.
- Exercise names are matched case-insensitively against the library and
  custom exercises, as today. Unknown names: a dry run lists them, and the
  user maps each one to an existing exercise or "create as custom".
- Units: those apps export in the user's unit; the importer converts to
  kilograms with `ToKg`.
- The existing duplicate rule (same start time and name skipped) stays, so
  importing twice is safe.
- Tricky: time zones. Strong exports local times without an offset; read
  them in the user's time zone, as the current importer does with
  `userLocation`.
- Tests: parser tests per format; API test for the mapping step.

### 22. Full account export and restore

The CSV export has every set but no templates, programs, progression state,
custom exercise details or notes. Moving to another server, or restoring one
account from a broken database, loses all of that.

- `GET /api/export/account.json`: profile settings, custom exercises,
  exercise notes, templates with their exercises and progression state,
  programs with their days and position, and workouts with sets.
- `POST /api/import/account.json` into an account without data: creates
  everything, remapping ids, in one transaction.
- The delete-account dialog (`DeleteAccount.tsx`) offers the export first.
- Tricky: built-in exercises are referenced by name, not id, so the file
  works on another server whose ids differ.
- Tests: an export and import round trip in `internal/api/export_test.go`
  produces the same records and the same next workout.

### 23. Getting back into a locked account

There is no password reset and no email at all. Someone who forgets their
password on a self-hosted instance has no way back in, and neither does the
person who runs it, short of SQL.

- A command: `workout-log reset-password <email>` prints a one-time link
  (or sets a temporary password) when run inside the container, documented
  in the README. No email needed.
- Optional email reset when `SMTP_URL` is set, with a single-use, hashed,
  one-hour token, using the same `sessions`-style hashing as
  `internal/auth`. Without SMTP the feature is hidden.
- Both sign out every session of the user, like "log out other devices" does
  for others.
- Tests: the command against a test schema; token single use and expiry.

### 24. Closing sign-up on a private server

`/api/auth/signup` is open to anyone who can reach the server. For a log run
for one person or a family, that means strangers can create accounts.

- `SIGNUP` setting: `open` (today's default), `invite` or `closed`. With
  `invite`, a signed-in user can create a one-time invite link in settings.
- The login page hides "Sign up" when it is closed; the endpoint answers 403.
- `SEED_DEMO` still works with sign-up closed.
- Tests: each mode in `internal/api/auth_test.go`, invites used once.

### 25. Record history

The records page shows the current best per exercise, and finishing a workout
reports new records (`RecordsSetIn`). There is no list of when each record
was set and by how much it moved, which is the most motivating view in most
lifting apps.

- Computed, not stored: walk the user's sets in time order and emit each time
  a record kind improves. A `RecordTimeline` in `internal/domain/records.go`
  next to `ComputeRecords`.
- On the exercise page: a list of PRs with dates, and markers on the progress
  chart.
- Tests: domain test for a sequence of sets, ties going to the earlier set as
  `better` already does.

### 26. Comparing training blocks

The progress chart shows one exercise over a time range and the volume page
shows weekly sets. Comparing "the 12 weeks of this program" with "the 12
before" (estimated max change, sessions, average sets per week) means reading
two charts side by side.

- A block is a date range, or the period a program was active (record
  `activated_at` and `deactivated_at` on `programs` from now on).
- A page that compares two blocks per exercise: start and end estimated max,
  best sets, sessions, weekly working sets, using the existing summary and
  volume code in `internal/domain`.
- Tests: domain tests for block summaries with gaps and deloads.

### 27. Logging without a signal

Gyms are often in basements. Every set is a request (`/api/workouts/{id}/sets`),
so logging fails when the signal does, and the rest timer is the only part
that keeps working.

- A service worker that caches the app shell, and an offline queue in
  IndexedDB for set changes in the current workout, replayed in order when the
  connection returns.
- The workout page works from the local copy and shows "3 changes waiting".
- Set ids: new sets get a client-generated id so later edits in the queue can
  refer to them; the API accepts and stores a `client_id` for idempotency, so a
  replay after a timeout does not create the set twice.
- Finishing a workout while offline is queued too, but progression and new
  records are shown only after the server has it.
- Tricky: the same workout open on two devices, one offline. Last write wins
  per set, which is acceptable for one person's log.
- Tests: the queue and replay as a pure module in `web/src/lib` with specs;
  API test for `client_id` idempotency.

### 28. Accessibility pass

The app has some ARIA labels (about thirty across `web/src`), but no one has
used it with a screen reader or keyboard only. Likely gaps from reading the
code:

- The set rows (`SetRow.tsx`) need labels per field that include the
  exercise and set number.
- The rest timer and the finish summary should be announced in a live region.
- The SVG charts (`components/charts`) need a text alternative, such as a
  table of the points.
- The calendar heat map relies on colour alone.
- Touch targets on the set row are small for sweaty hands, which is not
  strictly accessibility but goes with it.
- Tests: Testing Library specs for labels and live regions.

### 29. Browser end-to-end tests

Go tests call handlers and the web tests mock `fetch`. Nothing checks the
built app served by the Go binary with `WEB_DIR`: the SPA fallback, the
session cookie, the CSV download.

- Playwright in `web/e2e` against the Docker image with `SEED_DEMO=true`: log
  in, start the next program workout, log sets, finish and see the
  progression summary, open records, export CSV.
- A CI job after `docker`, reusing the built image.
- Tricky: the demo history is relative to today. Assert on structure.

### 30. Metrics and a readiness check

`/api/health` checks the database and reports the schema version; logs are
JSON lines. For someone hosting this for a few friends, request rates, errors
and latency would help.

- `/api/metrics` in Prometheus text format, off unless `METRICS=true`, with
  route patterns from chi as labels (not raw paths, which contain ids).
- The stale-workout closer and the session cleanup in `main.go` report how
  many rows they touched.
- Tests: disabled by default, counters move after a request.

## Maybe

### 31. Faster records with years of sets

`docs/architecture.md` says it plainly: records, progress series and new-PR
detection read all of a user's completed working sets (`store.LoggedSets`)
and compute in Go, which is fine for one person's log. Finishing a workout
calls `RecordsSetIn`, which runs `ComputeRecords` twice over every set.

- At four sessions a week with 25 sets each, that is about 5,000 sets a year;
  ten years is 50,000 rows read on every finish and every records page view.
  Probably still fast, but it grows without bound.
- If it matters: keep a `personal_records` table updated on finish and on set
  edits, and fall back to the full computation to rebuild it (after a formula
  change, or item 3's recalculation).
- Tricky: a second source of truth for records. A test should compare the
  stored records with a full `ComputeRecords` after a random series of
  inserts, edits and deletes.
- Why Maybe: I have no evidence it is slow. Measure with a seeded fixture of
  50,000 sets first, and assert on rows read, not timing.

### 32. A read-only link for a coach

Lifters often send their log to a coach or a training partner. Today that is
"copy as text" (`shareText.ts`) per workout.

- A revocable share link (`shares`: random token stored hashed, `scope` of
  all workouts or one program, `expires_at`) that opens a read-only view of
  history, records and charts without signing in.
- The read-only view reuses the pages with editing hidden, served from
  `/api/shared/<token>/...` routes that never touch the session.
- Why Maybe: an unauthenticated window into personal data, for a feature
  "copy as text" mostly covers.
- Tests: the token gives exactly its scope, expires, and is refused after
  revocation.

### 33. Coach accounts

A step further than item 32: a coach account that writes templates and
programs for several lifters and sees their logs.

- Links between users with consent from the lifter, templates owned by the
  coach and assigned to lifters, and comments on workouts.
- Why Maybe: it changes the ownership model everywhere. The store's
  `WHERE user_id = $1` checks (which `docs/architecture.md` describes as the
  ownership rule) would all need a second path, and getting one wrong leaks a
  lifter's data. It is a different product.

### 34. Bands and chains

Accommodating resistance (a band on the bar, chains) is common in
powerlifting. The weight on the bar is not the whole story, and today it can
only go in a note.

- An optional `extra` field on a set (band colour, chain weight), shown in
  history and kept out of records unless the user chooses otherwise.
- Why Maybe: a small group of lifters use it, and item 19's set notes and
  flags may be enough.

### 35. Languages and number formats

Everything is in English. Weights are entered with a dot or a comma (the
template editor replaces `,` with `.`), and numbers are shown with the
browser's default formatting in some places and fixed ones in others
(`web/src/lib/format.ts`).

- A `locale` setting for numbers and dates, applied through `Intl`
  consistently.
- Translations: a message catalogue per language.
- Why Maybe: the locale part is small and could be split out as an S; full
  translations need someone to keep them up to date.

### 36. Two-factor sign-in

Passwords are bcrypt hashed and sessions are hashed tokens with
"log out other devices". A training log is lower stakes than most accounts,
but a self-hosted server on the internet may still want a second factor.

- TOTP with recovery codes, as an option per user, with the QR code drawn in
  the browser; login becomes two steps when enabled, both rate limited.
- Why Maybe: little value for the data involved, and item 1's limits plus a
  good password cover most of the risk.
