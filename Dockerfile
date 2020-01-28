#----------------------------------------------------------------------extended
# Start from the latest golang base image
FROM golang:latest as builder

# Add Maintainer Info
LABEL maintainer="Hanif <mohamad.h@hotmail.no>"

# Set the Current Working Directory inside the container
WORKDIR /aguis/aguis

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o aguis .


######## Start a new stage from scratch #######
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /aguis/aguis .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./aguis"]
