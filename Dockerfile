FROM golang:1.24.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache make git bash

COPY go.mod go.sum makefile ./

RUN make deps

RUN make install-swagger

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/odin-dns .

EXPOSE 53
EXPOSE 8080

CMD ["./odin-dns"]
