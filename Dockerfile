# Quickfeed builder
FROM golang:1.16-alpine AS builder
RUN apk update && apk add --no-cache \
        ca-certificates \
        build-base \
        curl \
        bash \
        sudo \
        make \
        protobuf \
        protobuf-dev \
        npm
ADD . /quickfeed
WORKDIR /quickfeed
RUN make devtools && make ui && make proto && make install

EXPOSE 8080
ENTRYPOINT ["quickfeed"]