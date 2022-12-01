FROM golang:1.19-alpine

WORKDIR /app

RUN apk add --update gcc make git libc-dev binutils-gold ca-certificates

COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY core core

RUN CGO_ENABLED=0 GOOS=linux go build -o ./release ./cmd/main.go

ENTRYPOINT ["./release"]
