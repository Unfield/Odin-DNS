FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o odin-dns -ldflags="-s -w" ./cmd/odin-dns/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/odin-dns .

EXPOSE 53/udp
EXPOSE 53/tcp
EXPOSE 8080/tcp

CMD ["./odin-dns"]
