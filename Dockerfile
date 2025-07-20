FROM golang:tip-alpine3.22

WORKDIR /app

COPY . .
RUN go mod download

RUN go test ./...