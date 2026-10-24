# API reference

All endpoints live under `/api` and speak JSON. Weights are always in
kilograms (`weightKg`); the web app converts for users who prefer pounds.
Timestamps are RFC 3339.

## Authentication

Sessions use an HttpOnly cookie named `wl_session`. Sign up or log in to get
one; every endpoint except health, sign-up and login needs it.

| Method | Path                | Body                                         | Notes                   |
| ------ | ------------------- | -------------------------------------------- | ----------------------- |
| POST   | `/auth/signup`      | `{email, password, displayName}`             | 201, sets the cookie    |
| POST   | `/auth/login`       | `{email, password}`                          | 401 on bad credentials  |
| POST   | `/auth/logout`      |                                              | 204                     |
| GET    | `/me`               |                                              | current user            |
| PUT    | `/me`               | `{displayName, unit, timezone, e1rmFormula}` | `unit` is `kg` or `lb`  |
| PUT    | `/me/password`      | `{currentPassword, newPassword}`             | logs out other sessions |
| DELETE | `/me`               | `{password}`                                 | deletes everything      |
| POST   | `/me/logout-others` |                                              | ends all other sessions |

Login and sign-up are limited to 10 requests per minute per IP (429 after).

## Errors

```json
{
  "error": "validation_failed",
  "message": "Some fields are invalid.",
  "fields": { "reps": "Reps must be between 0 and 100." }
}
```

`error` is one of `bad_request` (400), `unauthorized` (401), `forbidden`
(403), `not_found` (404), `in_use` (409), `validation_failed` (422), `rate_limited` (429)
and `internal` (500). Unknown JSON fields are rejected with a 400.

## Exercises

| Method  | Path                       | Notes                                                       |
| ------- | -------------------------- | ----------------------------------------------------------- |
| GET     | `/exercises`               | `q`, `muscle`, `equipment`, `archived=true`, `sort=popular` |
| POST    | `/exercises`               | `{name, primaryMuscle, secondaryMuscles, equipment}`        |
| GET     | `/exercises/{id}`          |                                                             |
| PUT     | `/exercises/{id}`          | custom exercises only (403 for built-ins)                   |
| DELETE  | `/exercises/{id}`          | only if never used (409 `in_use`; archive instead)          |
| POST    | `/exercises/{id}/archive`  | hides it from pickers, keeps history                        |
| POST    | `/exercises/{id}/restore`  |                                                             |
| GET     | `/exercises/{id}/records`  | personal records and rep maxes                              |
| GET     | `/exercises/{id}/progress` | one point per workout: estimated max, top set, volume       |
| GET     | `/exercises/{id}/last`     | previous session, `excludeWorkout=<id>`                     |
| GET     | `/exercises/{id}/history`  | recent sessions with their sets, `limit` (1-50)             |
| GET/PUT | `/exercises/{id}/notes`    | `{notes}`, personal cues                                    |

## Workouts and sets

| Method | Path                          | Notes                                                                        |
| ------ | ----------------------------- | ---------------------------------------------------------------------------- |
| GET    | `/workouts`                   | `page`, `pageSize` (max 100), `exerciseId`                                   |
| POST   | `/workouts`                   | `{name, notes, startedAt, finishedAt}`                                       |
| GET    | `/workouts/{id}`              | with sets                                                                    |
| PUT    | `/workouts/{id}`              |                                                                              |
| DELETE | `/workouts/{id}`              |                                                                              |
| POST   | `/workouts/{id}/finish`       | returns `{workout, progression, records}`; idempotent                        |
| POST   | `/workouts/{id}/repeat`       | new workout with the same sets                                               |
| POST   | `/workouts/{id}/sets`         | `{exerciseId, reps, weightKg, rpe, completed, warmup}` (max 200 per workout) |
| POST   | `/workouts/{id}/warmups`      | `{exerciseId}`: inserts suggested warm-ups before it                         |
| PUT    | `/workouts/{id}/order`        | `{setIds}`: every set of the workout, in the new order                       |
| PUT    | `/workouts/{id}/sets/{setId}` |                                                                              |
| DELETE | `/workouts/{id}/sets/{setId}` |                                                                              |

The history is a page object: `{items, page, pageSize, total, totalPages}`.

## Templates and programs

| Method         | Path                      | Notes                                                                                                                       |
| -------------- | ------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| GET            | `/templates`              |                                                                                                                             |
| POST           | `/templates`              | `{name, notes, exercises: [{id?, exerciseId, sets, reps, weightKg, restSeconds, incrementKg, deloadAfter, deloadPercent}]}` |
| GET/PUT/DELETE | `/templates/{id}`         | PUT keeps rows with an `id`, adds the others, drops the rest                                                                |
| POST           | `/templates/{id}/start`   | starts a workout with the planned sets and the template notes                                                               |
| POST           | `/templates/{id}/copy`    | duplicates the template as "… (copy)"                                                                                       |
| GET/POST       | `/programs`               | `{name, templateIds}`                                                                                                       |
| GET/PUT/DELETE | `/programs/{id}`          |                                                                                                                             |
| POST           | `/programs/{id}/activate` | only one program is active                                                                                                  |
| POST           | `/programs/{id}/start`    | starts the next template in the rotation                                                                                    |

### Progression

When a workout started from a template is finished, each template exercise
with an `incrementKg` gets heavier if all planned sets were completed with
the planned reps at the planned weight. Otherwise its failure count goes up;
after `deloadAfter` failures in a row the weight drops by `deloadPercent`
and is rounded to 2.5 kg.

## Stats

| Method | Path               | Notes                                                    |
| ------ | ------------------ | -------------------------------------------------------- |
| GET    | `/records`         | records for every exercise                               |
| GET    | `/stats/volume`    | `weeks` (1-104): sets and tonnage per muscle per week    |
| GET    | `/stats/summary`   | this week, last week, week streak                        |
| GET    | `/stats/calendar`  | `month=YYYY-MM`: days with workouts, in your time zone   |
| GET    | `/export/sets.csv` | every set as CSV                                         |
| POST   | `/import/sets.csv` | CSV in the export format, see below                      |
| GET    | `/health`          | 200 with the schema version, 503 if the database is down |
| GET    | `/version`         | build version                                            |

## CSV export and import

`GET /export/sets.csv` and `POST /import/sets.csv` (body: the file, as
`text/csv`, at most 5 MB) use the same columns:

| Column      | Example               | Notes                                             |
| ----------- | --------------------- | ------------------------------------------------- |
| `date`      | `2026-09-01 18:30:00` | workout start in your time zone                   |
| `workout`   | `Day A`               |                                                   |
| `exercise`  | `Back Squat`          | must match an exercise in your library (any case) |
| `set`       | `2`                   | position in the workout; ignored on import        |
| `reps`      | `5`                   |                                                   |
| `weight_kg` | `100`                 | always kilograms                                  |
| `rpe`       | `8.5`                 | optional                                          |
| `warmup`    | `false`               | optional, default `false`                         |
| `completed` | `true`                | optional, default `true`                          |

Rows with the same date and workout name become one finished workout.
A workout that already exists with the same start and name is skipped, so
importing the same file twice is harmless. If any line is invalid nothing is
imported and the 422 response lists the lines (`"line 7": "Unknown exercise
\"Unicorn Press\". Create it first."`). Text fields starting with `=`, `+`,
`-` or `@` are exported with a leading `'` so spreadsheets do not run them as
formulas; the import strips it again.
