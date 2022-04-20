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
COPY . /quickfeed
WORKDIR /quickfeed
RUN make devtools
RUN make proto
RUN npm install -g webpack
RUN npm install -g webpack-cli
RUN make ui
RUN make install

ARG GRPC_PORT
ENV GRPC_PORT=${GRPC_PORT:-9090}

ARG HTTP_PORT
ENV HTTP_PORT=${HTTP_PORT:-8081}

EXPOSE ${HTTP_PORT} ${GRPC_PORT}
ENTRYPOINT ["/quickfeed/run.sh"]
