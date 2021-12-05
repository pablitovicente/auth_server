## Build
FROM golang:1.17.4-buster AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN export CGO_ENABLED=0 && go build -o /auth_server

## Deploy
FROM alpine:latest

WORKDIR /

COPY --from=builder /auth_server /auth_server

EXPOSE 1323

ENTRYPOINT ["/auth_server"]