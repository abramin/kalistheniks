FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/kalistheniks ./cmd/api

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates wget \
 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/kalistheniks /usr/local/bin/kalistheniks

ENV ADDR=":8080" \
    DB_DSN="postgres://kalistheniks:kalistheniks@db:5432/kalistheniks?sslmode=disable" \
    JWT_SECRET="change-me" \
    ENV="production"

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 CMD wget -qO- http://127.0.0.1:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/kalistheniks"]
