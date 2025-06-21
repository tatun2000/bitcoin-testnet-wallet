FROM golang:1.24.4-alpine3.22 as builder
WORKDIR /wallet/cmd
COPY . /wallet
RUN --mount=type=cache,target=/root/.cache/go-build go build -o wallet .

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /wallet/cmd/wallet /app/wallet

RUN chmod +x /app/wallet && \
    echo "/app/wallet" >> /etc/shells && \
    adduser -D -s /app/wallet walletuser && \
    chown -R walletuser /app

USER walletuser

ENTRYPOINT ["/app/wallet"]
