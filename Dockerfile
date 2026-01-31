# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o whatsbox ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/whatsbox .

# Create data directory
RUN mkdir -p /app/data

# Default environment variables
ENV PORT=3000 \
    HOST=0.0.0.0 \
    DATABASE_PATH=/app/data/whatsbox.db \
    WA_SESSION_PATH=/app/data/wa_session.db \
    TEMP_DIR=/app/data/temp \
    LOG_LEVEL=info \
    LOG_FORMAT=json

EXPOSE 3000

VOLUME ["/app/data"]

CMD ["./whatsbox"]
