# ====================== BUILD STAGE ======================
FROM golang:1.24-bookworm AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /run-app ./cmd/server

# ====================== RUNTIME STAGE ======================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /run-app /app/run-app
COPY --from=builder /usr/src/app/web /app/web

RUN mkdir -p /app/web/static /app/pb_data

EXPOSE 3000

CMD ["/app/run-app", "serve"]
