# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependency downloads separately from source changes.
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.20

RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /bin/server /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]
