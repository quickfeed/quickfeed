FROM envoyproxy/envoy:v1.14.1
COPY ./envoy/envoy.yaml /etc/envoy.yaml
ADD ./certfile /etc/certfile
ADD ./keyfile /etc/keyfile
CMD /usr/local/bin/envoy -c /etc/envoy.yaml