# syntax=docker/dockerfile:1

FROM golang:1.25

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /mc-admin-ui

EXPOSE 8080

ENV GIN_MODE=release

# Run
CMD ["/mc-admin-ui"]