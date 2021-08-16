FROM envoyproxy/envoy-alpine:v1.19-latest

ARG ENVOY_CONFIG=./envoy-localhost.yaml
ADD ${ENVOY_CONFIG} /etc/envoy/envoy.yaml

ARG DOMAIN=localhost
COPY ./certs /etc/letsencrypt/live/${DOMAIN}
RUN chmod 644 /etc/letsencrypt/live/${DOMAIN}/*.pem

CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
