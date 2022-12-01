FROM golang:1.19-alpine as builder

WORKDIR /app

RUN apk add --update gcc make git libc-dev binutils-gold ca-certificates

COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY core core

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /release ./cmd/main.go

# Final stage
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /release .
ENTRYPOINT ["/release"]
