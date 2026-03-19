FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /book-keeper .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /book-keeper /usr/local/bin/book-keeper

RUN mkdir -p /data

ENTRYPOINT ["book-keeper"]
