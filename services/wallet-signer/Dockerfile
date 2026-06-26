
FROM rust:1.75-alpine AS builder

RUN apk add --no-cache musl-dev openssl-dev

WORKDIR /app

COPY Cargo.toml Cargo.lock ./
RUN cargo fetch

COPY src ./src
RUN cargo build --release

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata

RUN addgroup -g 1000 -S app && \
    adduser -u 1000 -S app -G app

WORKDIR /app

COPY --from=builder /app/target/release/wallet-signer /app/wallet-signer

USER app

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/wallet-signer", "--health"]

EXPOSE 8088

ENTRYPOINT ["/app/wallet-signer"]

