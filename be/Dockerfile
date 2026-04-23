FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o falzo-api ./cmd/api

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/falzo-api ./falzo-api
EXPOSE 8000
CMD [ "./falzo-api" ]