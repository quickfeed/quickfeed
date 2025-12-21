FROM golang:1.24-alpine
RUN apk update && apk add --no-cache git bash build-base
WORKDIR /quickfeed
COPY go.mod go.sum ./
RUN go mod download
