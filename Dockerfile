# Multi-stage build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make
WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata postgresql-client
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /go/bin/goose /usr/local/bin/goose

EXPOSE 8080

CMD ["/bin/sh", "-c", "goose -dir ./migrations postgres \"user=$DB_USER password=$DB_PASSWORD host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable\" up && ./server"]
