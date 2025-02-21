# syntax=docker/dockerfile:1

FROM golang:1.23-alpine

# Update and install build-base, npm and webpack
RUN apk update && apk add --no-cache bash build-base npm && npm install webpack

WORKDIR /quickfeed

# Set the port number the container should expose
EXPOSE 443

# Install air package for live reloading
RUN go install github.com/air-verse/air@latest

# Copy the local package files to the container's workspace
COPY go.mod go.sum ./
COPY ./kit/go.mod ./kit/go.sum ./kit/

# Download dependencies
RUN go mod download

# Running the quickfeed application in development mode
CMD ["air", "-c", ".air.toml"]
