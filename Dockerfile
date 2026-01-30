# Build
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o gogate .

# Minimal Image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/gogate .
COPY config.toml .

EXPOSE 25565

# Use a custom config here
CMD ["./gogate"]
