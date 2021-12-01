#!/bin/bash
if [ ! -d "/etc/dummycerts" ]
then
	sudo mkdir /etc/dummycerts
	echo "Creating dummy Certificate folder"
fi
#Folder for 
cd /etc/dummycerts/

#Creating certificates
sudo openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout RootCA.key -out RootCA.pem -subj "/C=US/CN=Example-Root-CA"
sudo openssl x509 -outform pem -in RootCA.pem -out RootCA.crt

#Link to cert files, kept in /etc/dummycerts
CRT=$(readlink -f ./RootCA.crt)
KEY=$(readlink -f ./RootCA.key)
PEM=$(readlink -f ./RootCA.pem)

echo "Creating config file in nginx/sites-available/default."

sudo bash -c 'cat > /etc/nginx/sites-available/default' <<EOL
server {
        listen 443 ssl http2;
        listen [::]:443 ssl http2;

        server_name https://127.0.0.1/auth/github/callback;
        location / {
                proxy_pass http://127.0.0.1:8081;
                proxy_redirect off;
                proxy_set_header Host \$host;
                proxy_set_header X-Real-IP \$remote_addr;
                proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Ssl on;
        }

        location /AutograderService/ {
                grpc_pass 127.0.0.1:8080;
                proxy_redirect off;
                proxy_set_header Host \$host;
                proxy_set_header X-Real-IP \$remote_addr;
                proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Ssl on;

                if (\$request_method = 'OPTIONS') {
                        add_header 'Access-Control-Allow-Origin' '*';
                        add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
                        add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web';
                        add_header 'Access-Control-Max-Age' 1728000;
                        add_header 'Content-Type' 'text/plain charset=UTF-8';
                        add_header 'Content-Length' 0;
                        return 204;
                        }
                if (\$request_method = 'POST') {
                        add_header 'Access-Control-Allow-Origin' '*';
                        add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS';
                        add_header 'Access-Control-Allow-Headers' 'DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Content-Transfer-Encoding,Custom-Header-1,X-Accept-Content-Transfer-Encoding,X-Accept-Response-Streaming,X-User-Agent,X-Grpc-Web';
                        add_header 'Access-Control-Expose-Headers' 'Content-Transfer-Encoding';
                }
        }
	ssl_certificate $CRT;
	ssl_certificate_key $KEY;
	ssl_trusted_certificate $PEM;
}
EOL

echo "The certificates and the nginx config file was created successfully. 'cat /etc/nginx/sites-available/default' to see the config"
echo -e  "\e[31mWarning\e[39m: These certificates are for development only and should never be used in production"

echo -e "\e[92m -- Reloading nginx config and restarting nginx -- \e[0m "
sudo nginx -t
sudo service nginx start

echo -e "\e[95mIt might say 'suspicious symbols' etc. But as long as the syntax is OK, it works.\e[0m"
