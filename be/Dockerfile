# syntax=docker/dockerfile:1

FROM golang:1.26.2-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/falzo-api \
    ./cmd/api

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 falzo \
    && adduser -S -D -H -u 10001 -G falzo falzo \
    && mkdir -p /app/logs \
    && chown -R falzo:falzo /app

WORKDIR /app
COPY --from=builder --chown=falzo:falzo /out/falzo-api ./falzo-api

USER falzo
EXPOSE 8080
ENTRYPOINT ["/app/falzo-api"]
