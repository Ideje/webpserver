FROM golang:1.17 AS build

ENV GOPATH=/go
ENV GO111MODULE=on

WORKDIR /go/src
COPY go.mod webpserver/
COPY go.sum webpserver/

WORKDIR /go/src/webpserver

RUN go mod download

# disable HTTP/2 server support
ENV GODEBUG=http2server=0

WORKDIR /go/src
COPY . webpserver

WORKDIR /go/src/webpserver

RUN go get -d -v ./...

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install -a -tags netgo -ldflags '-s -w -extldflags "-static"'

FROM alpine:3.14
COPY --from=build /go/bin/webpserver /usr/sbin/webpserver

RUN mkdir -p /app/data
WORKDIR /app

ENTRYPOINT webpserver
