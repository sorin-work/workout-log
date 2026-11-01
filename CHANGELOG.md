# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses
[Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-11-01

First stable release.

### Features

- Accounts with session cookies, password change, log out other devices,
  account deletion and rate-limited login and sign-up.
- Exercise library with 44 built-in lifts, custom exercises (archive,
  restore, delete unused ones), personal notes and most-used-first search.
- Workout logging with weight, reps, RPE and warm-ups; rest timer with sound,
  plate calculator, suggested warm-ups, "last time" and today's best
  estimated 1RM; reorder exercises, repeat a workout, copy it as text.
- Templates with target sets, reps, weight and rest, linear progression and
  automatic deloads; programs that rotate through templates.
- Personal records, rep maxes and new-PR detection on finish; Epley or
  Brzycki estimated 1RM.
- Progress charts per exercise with time ranges, recent sessions, weekly sets
  and tonnage per muscle group with a heat map and target band, training
  calendar, weekly summary and streak.
- Kilograms or pounds, per-user time zone, CSV export and import.
- Demo account with a year of 5x5 history.
- Docker image serving the API and web app, compose setup with PostgreSQL,
  healthchecks and restart policies; CI for lint, tests, build and Docker.

### Fixed during development

- Sets with an RPE above 10 were accepted.
- The oldest workouts were missing from the history pages.
- Sets could be edited or deleted through another workout's URL.
- HTML in workout notes was rendered instead of shown as text.
- Heavy singles showed an estimated max above the weight lifted.
- Names with accents or non-Latin letters hit length limits early.
- Weights entered in pounds came back slightly different.
- Deloads dropped the weight further than the deload percentage.
- Finishing a workout twice added the progression increment twice.
- Progress charts zig-zagged after back-filling old workouts.
- Weekly volume put late-evening workouts in the wrong week.
- The workout history ran one query per workout.
- Picking an exercise in the template editor submitted the form.
