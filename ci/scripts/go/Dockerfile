FROM golang:1.15-alpine

# Install bash and git (and build-base to get gcc)
# (this is required when building FROM: golang:alpine)
RUN apk update && apk add --no-cache git bash build-base

WORKDIR /quickfeed

RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.30.0
