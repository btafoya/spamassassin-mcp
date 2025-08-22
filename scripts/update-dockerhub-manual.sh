#!/bin/bash

# Manual script to update Docker Hub overview
# Usage: ./scripts/update-dockerhub-manual.sh <dockerhub_username> <dockerhub_password>

if [ $# -ne 2 ]; then
    echo "Usage: $0 <dockerhub_username> <dockerhub_password>"
    exit 1
fi

DOCKERHUB_USERNAME=$1
DOCKERHUB_PASSWORD=$2
DOCKERHUB_REPOSITORY="spamassassin-mcp"

echo "Generating Docker Hub overview..."
chmod +x scripts/extract-dockerhub-info.sh
./scripts/extract-dockerhub-info.sh

echo "Replacing placeholders..."
sed -i "s/your-dockerhub-username/$DOCKERHUB_USERNAME/g" dockerhub-overview.md
sed -i "s/your-username/btafoya/g" dockerhub-overview.md

echo "Docker Hub overview generated. You can now manually copy the content of dockerhub-overview.md"
echo "to your Docker Hub repository description at:"
echo "https://hub.docker.com/repository/docker/$DOCKERHUB_USERNAME/$DOCKERHUB_REPOSITORY/general"

echo ""
echo "To update Docker Hub automatically, set the following secrets in your GitHub repository:"
echo "- DOCKERHUB_USERNAME = $DOCKERHUB_USERNAME"
echo "- DOCKERHUB_PASSWORD = <your Docker Hub access token>"