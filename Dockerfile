FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./

ARG BUILD_VERSION=0.0.0
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${BUILD_VERSION}" -o /book-keeper .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /book-keeper /usr/local/bin/book-keeper

RUN mkdir -p /data

ENTRYPOINT ["book-keeper"]
