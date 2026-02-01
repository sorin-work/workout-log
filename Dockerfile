# syntax=docker/dockerfile:1

FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.25-alpine AS api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/workout-log ./cmd/workout-log

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=api /out/workout-log /app/workout-log
COPY --from=web /src/web/dist /app/web
ENV WEB_DIR=/app/web PORT=8080
USER app
EXPOSE 8080
ENTRYPOINT ["/app/workout-log"]
