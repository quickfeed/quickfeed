FROM envoyproxy/envoy:v1.14.1
COPY ./envoy/envoy.yaml /etc/envoy.yaml
ADD ./fullchain.pem /etc/letsencrypt/live/ag3.ux.uis.no/fullchain.pem
ADD ./privkey.pem /etc/letsencrypt/live/ag3.ux.uis.no/privkey.pem
CMD /usr/local/bin/envoy -c /etc/envoy.yaml