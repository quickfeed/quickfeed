FROM envoyproxy/envoy:latest
RUN apt-get update
COPY ./envoy/envoy.yaml /etc/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy.yaml