FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:3.21

RUN adduser -D appuser

WORKDIR /data
COPY --from=builder /app/server /app/server
COPY .env /data/.env

RUN chown -R appuser:appuser /data

USER appuser
EXPOSE 7300

ENTRYPOINT ["/app/server"]
