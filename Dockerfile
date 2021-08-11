FROM golang:1.16-alpine
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
RUN make devtools
RUN make proto
RUN make ui
RUN make install

EXPOSE 8080 9091
ENTRYPOINT ["quickfeed"]
