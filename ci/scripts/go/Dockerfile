FROM golang:latest

WORKDIR /quickfeed

# Install bash and git
# (this is required if building FROM: golang:alpine)
# RUN apk add --no-cache bash git

# Compile gosecret binary
COPY cmd/gosecret cmd/gosecret
RUN go build -o bin/gosecret ./cmd/gosecret

# Populate module cache based on go.{mod,sum} files in kit module
COPY kit kit
RUN cd kit && go mod download
