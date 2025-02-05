# syntax=docker/dockerfile:1

# Set language to Go, bullseye is the latest version of Debian
FROM golang:1.23-bullseye

# Set the port number the container should expose
# 443 is the default port for HTTPS
EXPOSE 443

# Copy the current directory contents into the container at /app
COPY . .

# Install dependencies
RUN go install

# Running the quickfeed application in development mode
CMD ["air", "-c", ".air.toml"]
