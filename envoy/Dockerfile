FROM envoyproxy/envoy-alpine:v1.19-latest

ARG ENVOY_CONFIG
ENV ENVOY_CONFIG=${ENVOY_CONFIG:-./envoy/envoy.yaml}
ADD ${ENVOY_CONFIG} /etc/envoy/envoy.yaml

EXPOSE 8081 9090
CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]