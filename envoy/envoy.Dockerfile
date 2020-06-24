FROM envoyproxy/envoy:v1.14.1
COPY ./envoy/envoy.yaml /etc/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy.yaml