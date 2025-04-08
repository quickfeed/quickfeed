FROM golang:1.24-alpine

# Install bash and git (and build-base to get gcc)
# (this is required when building FROM: golang:alpine)
RUN apk update && apk add --no-cache git=~2.47 bash=~5.2.37 build-base=~0.5 golangci-lint=~1.61

WORKDIR /quickfeed
