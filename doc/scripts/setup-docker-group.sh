if [ "$(uname)" != "Linux" ]; then
	echo "This is only for linux"
	exit 1
fi

if [ -z "$(getent group docker)" ]; then
    # Create docker group
    sudo groupadd docker
fi

if ! systemctl is-active --quiet docker; then
    sudo systemctl start docker;
fi

if [ -z "$(getent group docker | grep $USER)" ]; then
    # Add user
    sudo usermod -aG docker $USER
    # Restart docker
    sudo systemctl restart docker;
fi

# Check if user does not have access, and if the group is configured correctly
# The three last statements are opposites of previous statements
if ! docker ps > /dev/null 2>&1; then
    echo "The group is configured, but you may need to restart your system for the changes to take effect"
else
    echo "Docker is running and properly configured"
fi
