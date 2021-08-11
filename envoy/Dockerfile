FROM envoyproxy/envoy-alpine:v1.19-latest

ADD ./envoy/envoy.yaml /etc/envoy/envoy.yaml

ARG DOMAIN
COPY ./envoy/certs /etc/letsencrypt/live/${DOMAIN}

RUN chmod 644 /etc/letsencrypt/live/${DOMAIN}/fullchain.pem
RUN chmod 644 /etc/letsencrypt/live/${DOMAIN}/key.pem

CMD ["envoy", "-c", "/etc/envoy/envoy.yaml"]
