# Script runs between each reload of the server
# View .air.toml to alter configuration

if pgrep quickfeed > /dev/null; then
    pkill quickfeed
fi

# Following can be replaced with 'make install', but would have to update the tmp directory for air

go build -o ./tmp-air/quickfeed .

if [ "$(uname)" == "Linux" ]; then
    setcap cap_net_bind_service=+ep ./tmp-air/quickfeed
fi
