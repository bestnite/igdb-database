FROM golang:1.24-alpine AS builder
LABEL authors="nite"

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG version=dev
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o igdb-database .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/igdb-database /app/igdb-database

ENTRYPOINT [ "./igdb-database"]