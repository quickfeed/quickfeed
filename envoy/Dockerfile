FROM envoyproxy/envoy-alpine:v1.19-latest

# default envoy config if `--build-arg` is not informed.
ARG ENVOY_CONFIG=./envoy/envoy.yaml
COPY ${ENVOY_CONFIG} /etc/envoy/envoy.yaml

EXPOSE 8081 9090
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]