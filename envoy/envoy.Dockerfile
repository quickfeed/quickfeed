FROM envoyproxy/envoy:latest
COPY envoy.yaml /etc/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy.yaml