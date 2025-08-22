#!/bin/bash

# Manual script to update Docker Hub overview
# Usage: ./scripts/update-dockerhub-manual.sh <docker_username> <docker_password>

if [ $# -ne 2 ]; then
    echo "Usage: $0 <docker_username> <docker_password>"
    exit 1
fi

DOCKER_USERNAME=$1
DOCKER_PASSWORD=$2
DOCKER_REPOSITORY="spamassassin-mcp"

echo "Generating Docker Hub overview..."
chmod +x scripts/extract-dockerhub-info.sh
./scripts/extract-dockerhub-info.sh

echo "Replacing placeholders..."
sed -i "s/your-dockerhub-username/$DOCKER_USERNAME/g" dockerhub-overview.md
sed -i "s/your-username/btafoya/g" dockerhub-overview.md

echo "Docker Hub overview generated. You can now manually copy the content of dockerhub-overview.md"
echo "to your Docker Hub repository description at:"
echo "https://hub.docker.com/repository/docker/$DOCKER_USERNAME/$DOCKER_REPOSITORY/general"

echo ""
echo "To update Docker Hub automatically, set the following secrets in your GitHub repository:"
echo "- DOCKER_USERNAME = $DOCKER_USERNAME"
echo "- DOCKER_PASSWORD = <your Docker Hub access token>"