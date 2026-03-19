FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /falzo-api ./cmd/api

FROM alpine:3.22

WORKDIR /app

RUN adduser -D -u 10001 appuser

COPY --from=builder /falzo-api /usr/local/bin/falzo-api

ENV APP_ENV=production
ENV HTTP_ADDR=:8080

EXPOSE 8080

USER appuser

CMD ["falzo-api"]
