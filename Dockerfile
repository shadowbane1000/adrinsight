# Build stage
FROM golang:bookworm AS builder

RUN apt-get update && apt-get install -y libsqlite3-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -tags fts5 -o /app/adr-insight ./cmd/adr-insight

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y libsqlite3-0 ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/adr-insight .
RUN mkdir -p /data/adr

EXPOSE 8081

CMD ["./adr-insight", "serve", "--adr-dir", "/data/adr", "--db", "/data/adr-insight.db"]
