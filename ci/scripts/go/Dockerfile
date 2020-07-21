FROM golang:alpine

# Install bash and git
# (this is required when building FROM: golang:alpine)
RUN apk update && apk add --no-cache git bash

WORKDIR /quickfeed

COPY cmd/gosecret cmd/gosecret
RUN go build -o bin/gosecret ./cmd/gosecret
