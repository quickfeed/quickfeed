#!/bin/bash

# builds a new container from Envoy image if called with an argument
# othervise starts the Envoy container

if [ "$#" -eq 1 ]; then
    echo "got 1 arg"
    docker build -t ag_envoy -f ./envoy/envoy.Dockerfile .
else
    docker run --name=envoy -p 8082:8082 --net=host ag_envoy 
fi