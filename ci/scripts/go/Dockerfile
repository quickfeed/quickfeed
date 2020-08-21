FROM golang:1.15-alpine

# Install bash and git (and build-base to get gcc)
# (this is required when building FROM: golang:alpine)
RUN apk update && apk add --no-cache git bash build-base

WORKDIR /quickfeed

COPY cmd/gosecret cmd/gosecret
RUN go build -o bin/gosecret ./cmd/gosecret
