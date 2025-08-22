# Docker Hub Setup Guide

This document explains how to automatically set the Docker Hub overview from the README.md file.

## Files Created

1. `scripts/extract-dockerhub-info.sh` - Extracts key information from README.md and formats it for Docker Hub
2. `scripts/update-dockerhub-manual.sh` - Manual script to update Docker Hub overview
3. `.github/workflows/update-dockerhub.yml` - GitHub Actions workflow to automatically update Docker Hub description
4. `dockerhub-overview.md` - Generated Docker Hub overview file

## How It Works

1. The `extract-dockerhub-info.sh` script reads the README.md file and extracts key sections
2. It formats this information into a Docker Hub-friendly format in `dockerhub-overview.md`
3. The GitHub Actions workflow automatically runs this script when README.md is updated
4. If Docker Hub credentials are provided, it automatically updates the Docker Hub description

## Setting Up Automatic Updates

To enable automatic Docker Hub description updates:

1. Generate a Docker Hub access token:
   - Log in to Docker Hub
   - Go to Account Settings > Security
   - Click "New Access Token"
   - Give it a descriptive name (e.g., "GitHub Actions")
   - Set permissions to "Read & Write"
   - Copy the generated token

2. Set up Docker Hub credentials as GitHub Secrets:
   - Go to your GitHub repository settings
   - Click "Secrets and variables" > "Actions"
   - Add two new repository secrets:
     - `DOCKERHUB_USERNAME`: Your Docker Hub username
     - `DOCKERHUB_PASSWORD`: Your Docker Hub access token

## Manual Updates

You can also manually generate and update the Docker Hub overview:

```bash
# Generate Docker Hub overview
./scripts/extract-dockerhub-info.sh

# Manually update Docker Hub (requires Docker Hub credentials)
./scripts/update-dockerhub-manual.sh your-dockerhub-username your-dockerhub-access-token
```

## Customization

To customize the Docker Hub overview:

1. Edit `scripts/extract-dockerhub-info.sh` to modify the content extraction logic
2. Modify the template in the script to change the format
3. Update the GitHub Actions workflow if needed

The workflow automatically triggers on changes to:
- README.md
- The workflow file itself
- The extraction script