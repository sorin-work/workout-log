.PHONY: build run test test-web lint lint-web fmt db web-dev docker

TEST_DATABASE_URL ?= postgres://workout:workout@localhost:5432/workout_test?sslmode=disable

build:
	cd web && npm run build
	go build -o bin/workout-log ./cmd/workout-log

run:
	go run ./cmd/workout-log

test:
	TEST_DATABASE_URL=$(TEST_DATABASE_URL) go test ./...

test-web:
	cd web && npm test

lint:
	test -z "$$(gofmt -l cmd internal)"
	go vet ./...

lint-web:
	cd web && npm run lint && npm run format:check && npm run typecheck

fmt:
	gofmt -w cmd internal
	cd web && npm run format

# Starts only PostgreSQL from docker compose and creates the test database.
db:
	docker compose up -d --wait db
	docker compose exec db createdb -U workout workout_test 2>/dev/null || true

web-dev:
	cd web && npm run dev

docker:
	docker compose up -d --build
