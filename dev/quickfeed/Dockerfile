FROM golang

WORKDIR /quickfeed
COPY . /quickfeed
RUN ls -la
RUN go install

CMD ["quickfeed", "-help"]

# COPY cmd/gosecret cmd/gosecret
# RUN go build -o bin/gosecret ./cmd/gosecret
# RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.30.0
