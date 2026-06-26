
# Stage 1: Build
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/auth-service ./cmd/auth-service

# Stage 2: Final lightweight image
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -S app -G app

WORKDIR /app

COPY --from=builder /go/bin/auth-service /app/auth-service
COPY --from=builder /app/configs /app/configs

RUN chown -R app:app /app

USER app

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/auth-service", "healthcheck"]

EXPOSE 8080

ENTRYPOINT ["/app/auth-service"]

