# syntax=docker/dockerfile:1

FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS api
ARG VERSION=1.0.0
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/sorin-work/workout-log/internal/api.Version=${VERSION}" -o /out/workout-log ./cmd/workout-log

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=api /out/workout-log /app/workout-log
COPY --from=web /src/web/dist /app/web
ENV WEB_DIR=/app/web PORT=8080
USER app
EXPOSE 8080
ENTRYPOINT ["/app/workout-log"]
