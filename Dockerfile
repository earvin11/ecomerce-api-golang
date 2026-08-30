# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/bin/api ./cmd/api

FROM alpine:3.20

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /app/bin/api ./api

USER app

EXPOSE 8080

ENTRYPOINT ["./api"]