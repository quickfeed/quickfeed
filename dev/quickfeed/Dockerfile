FROM golang:1.15-alpine

# TODO: delete packages after tests
RUN apk update && apk add --no-cache \
        ca-certificates \
        build-base \
        vim \
        curl \
        bash \
        net-tools \
        make \
        protobuf \
        protobuf-dev \
        npm

ADD . /quickfeed
WORKDIR /quickfeed

RUN make devtools && make ui && make proto && make install

EXPOSE 8080
ENTRYPOINT ["quickfeed"]
