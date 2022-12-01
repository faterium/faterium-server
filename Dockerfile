FROM golang:1.19-alpine as builder

WORKDIR /app

RUN apk add --update gcc make git libc-dev binutils-gold ca-certificates

COPY collections.json collections.json
COPY go.mod go.mod
COPY go.sum go.sum
COPY cmd cmd
COPY core core

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /release ./cmd/main.go

# Final stage
FROM alpine
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /release .
RUN apk add dbus avahi avahi-compat-libdns_sd
ENTRYPOINT ["/release"]
