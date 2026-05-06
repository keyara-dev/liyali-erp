# Railway backend build
# Build context is the repo root; all paths are relative to here.

FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make ca-certificates tzdata

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .

RUN go mod tidy && go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o main .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o migrate \
    -tags migrate \
    ./cmd/migrate

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget && \
    addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/config ./config
COPY --from=builder /app/database ./database

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
