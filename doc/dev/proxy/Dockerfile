FROM envoyproxy/envoy:v1.17.0

# envoy configuration
COPY ./envoy/envoy.yaml /etc/envoy/envoy.yaml
CMD /usr/local/bin/envoy -c /etc/envoy/envoy.yaml -l trace --log-path /tmp/envoy_info.log

# nginx configuration
RUN apt update && apt install nginx -y
RUN rm /etc/nginx/sites-enabled/default
COPY ./dev/proxy/nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY ./dev/proxy/nginx/site-available/quickfeed.conf /etc/nginx/sites-available/quickfeed
RUN ln -s /etc/nginx/sites-available/quickfeed /etc/nginx/sites-enabled/quickfeed

# TODOS:
# Install certbot
# Fix nginx config
# Enable service, expose ports
# RUN service nginx enable && service nginx start
