FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.21

RUN adduser -D appuser && \
    apk add --no-cache wget

WORKDIR /data
COPY --from=builder /app/server /app/server

RUN chown -R appuser:appuser /data

USER appuser
EXPOSE 7300

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:7300/health || exit 1

ENTRYPOINT ["/app/server"]
