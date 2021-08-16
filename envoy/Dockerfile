FROM envoyproxy/envoy-alpine:v1.19-latest

ARG ENVOY_CONFIG
ENV ENVOY_CONFIG=${ENVOY_CONFIG:-./envoy/envoy-localhost.yaml}
ADD ${ENVOY_CONFIG} /etc/envoy/envoy.yaml

ARG DOMAIN
ENV DOMAIN=${DOMAIN:-localhost}
COPY ./envoy/certs /etc/letsencrypt/live/${DOMAIN}
RUN chmod 644 /etc/letsencrypt/live/${DOMAIN}/*.pem

CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
