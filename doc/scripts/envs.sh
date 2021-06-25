# Make a copy of this file and update the variables.
# WARNING: Do not push your file with secret values to GitHub or elsewhere.

# GitHub OAUTH App keys
export GITHUB_KEY="KEY"
export GITHUB_SECRET="SECRET"

# Local GitHub OAUTH App keys
export LOCAL_GITHUB_KEY="KEY"
export LOCAL_GITHUB_SECRET="SECRET"

# Envoy Config
export ENVOY_CONFIG="envoy/envoy-example.yaml"
export GRPC_PORT="9090"
export HTTP_PORT="8081"

# Configuration for remote
export DOMAIN="www.example.com"
export CERT="/etc/letsencrypt/live/www.example.com/fullchain.pem"
export CERT_KEY="/etc/letsencrypt/live/www.example.com/privkey.pem"

# Configuration for local
export LOCAL_DOMAIN="127.0.0.1"
# You can generate local certificates using the script in /doc/local-setup/gen-local-certificate.sh
export LOCAL_CERT="/etc/dummycerts/RootCA.crt"
export LOCAL_CERT_KEY="/etc/dummycerts/RootCA.key"
