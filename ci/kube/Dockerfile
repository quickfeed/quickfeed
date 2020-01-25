#docker file for kube
FROM golang:latest

LABEL maintainer="Hanif <mohamad.h@hotmail.no>"

WORKDIR /go/hello

COPY go.mod ./


RUN go mod download

COPY . .

RUN go build -o main .

#need to expose?
EXPOSE 8080
#what to have inside here when all backend has to run?
CMD ["./main"]