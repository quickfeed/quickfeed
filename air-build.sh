# Script runs between each reload of the server
# View .air.toml to alter configuration

# Build binary with version control system disabled
go build -buildvcs=false -o ./tmp-air/quickfeed .

# Set capabilities to bind to privileged ports
# This is only necessary on Linux and when not running in a Docker container
if [ "$(uname)" == "Linux" ] && [ ! -f /.dockerenv ]; then
    setcap cap_net_bind_service=+ep ./tmp-air/quickfeed
fi
