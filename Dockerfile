# syntax=docker/dockerfile:1

FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Produce a stripped static binary for the runtime stage.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mc-admin-ui

FROM alpine:3.20 AS runtime
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder --chown=app:app /app/mc-admin-ui /usr/local/bin/mc-admin-ui
COPY --from=builder --chown=app:app /app/templates ./templates

ENV GIN_MODE=release

EXPOSE 8080

USER app

CMD ["mc-admin-ui"]