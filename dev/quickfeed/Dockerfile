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

RUN go list -f "{{range .Imports}}{{.}} {{end}}" tools.go
RUN make devtools && make ui && go list && make proto && make install

EXPOSE 8080
ENTRYPOINT ["quickfeed"]

# COPY cmd/gosecret cmd/gosecret
# RUN go build -o bin/gosecret ./cmd/gosecret
# RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.30.0

