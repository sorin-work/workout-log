.PHONY: build run test lint fmt

build:
	go build -o bin/workout-log ./cmd/workout-log

run:
	go run ./cmd/workout-log

test:
	go test ./...

lint:
	gofmt -l . | (! grep .)
	go vet ./...

fmt:
	gofmt -w .
