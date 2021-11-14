FROM golang:1.17-alpine

# Install bash and git (and build-base to get gcc)
# (this is required when building FROM: golang:alpine)
RUN apk update && apk add --no-cache git bash build-base
RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.41.1

WORKDIR /quickfeed
