FROM envoyproxy/envoy-alpine:v1.19-latest

ARG ENVOY_CONFIG
ENV ENVOY_CONFIG=${ENVOY_CONFIG:-./envoy/envoy.yaml}
ADD ${ENVOY_CONFIG} /etc/envoy/envoy.yaml

CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]