# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm

# Update and install build-base, npm and webpack
RUN apt-get update && apt-get install -y npm && npm install -g webpack

WORKDIR /quickfeed

# Set the port number the container should expose
EXPOSE 443

# Install air package for live reloading
# Pinning to a specific version to avoid breaking changes
RUN go install github.com/air-verse/air@v1.64

# Copy the local package files to the container's workspace
COPY go.mod go.sum ./
COPY ./kit/go.mod ./kit/go.sum ./kit/

# Download dependencies
RUN go mod download

# Running the quickfeed application in development mode
CMD ["air", "-c", ".air.toml"]
