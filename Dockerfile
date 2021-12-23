## Build
FROM golang:1.17.4-buster AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY pkg ./pkg
COPY certs ./certs

RUN export CGO_ENABLED=0 && go build -o /auth_server

## Deploy
FROM alpine:latest

WORKDIR /app

COPY --from=builder /auth_server /app/auth_server
COPY /config.json /app/config.json
COPY /certs /app/certs

EXPOSE 3000

ENTRYPOINT ["/app/auth_server"]