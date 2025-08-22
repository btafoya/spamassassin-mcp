#!/bin/bash

# Extract Docker Hub information from README.md
# This script creates a Docker Hub overview by extracting key sections from the README

echo "Extracting Docker Hub information from README.md..."

# Create Docker Hub overview file
cat > dockerhub-overview.md << 'EOF'
# SpamAssassin MCP Server

A secure, containerized Model Context Protocol (MCP) server that integrates SpamAssassin for defensive email security analysis. This server provides Claude Code with comprehensive email analysis capabilities while maintaining strict security boundaries.

## 🔒 Security-First Design

**Defensive Operations Only** - This server exclusively provides security analysis capabilities:
- ✅ Email spam detection and analysis
- ✅ Sender reputation checking
- ✅ Rule testing and validation
- ✅ Configuration inspection
- ❌ NO email sending/relay capabilities
- ❌ NO malicious content generation
- ❌ NO offensive security tools

## 🚀 Quick Start

### Using Docker

```bash
# Pull the latest image from Docker Hub
docker pull your-dockerhub-username/spamassassin-mcp:latest

# Run the container
docker run -d \
  --name spamassassin-mcp \
  -p 8081:8080 \
  your-dockerhub-username/spamassassin-mcp:latest
```

### Using Docker Compose

```bash
# Clone the repository
git clone https://github.com/your-username/spamassassin-mcp.git
cd spamassassin-mcp

# Copy and customize configuration
cp .env.example .env
# Edit .env to customize ports and settings

# Start the containers
docker compose up -d

# Check health
docker compose logs spamassassin-mcp
```

## 🛠️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SA_MCP_HOST_PORT` | `8081` | Host port for container deployment |
| `SA_MCP_LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `SA_MCP_SERVER_BIND_ADDR` | `0.0.0.0:8080` | Server bind address (container internal) |
| `SA_MCP_SPAMASSASSIN_HOST` | `localhost` | SpamAssassin daemon host |
| `SA_MCP_SPAMASSASSIN_PORT` | `783` | SpamAssassin daemon port |
| `SA_MCP_SPAMASSASSIN_THRESHOLD` | `5.0` | Spam score threshold |
| `SA_MCP_SECURITY_MAX_EMAIL_SIZE` | `10485760` | Max email size (10MB) |

### Security Settings

The server includes comprehensive security measures:
- **Rate Limiting**: 60 requests/minute with burst of 10
- **Input Validation**: Email format and size validation
- **Container Security**: Non-root execution, read-only filesystem
- **Network Isolation**: Custom bridge network
- **Resource Limits**: Memory and CPU constraints

## 📋 Available Tools

### Email Analysis

- `scan_email` - Analyze email content for spam probability and rule matches
- `check_reputation` - Check sender reputation and domain/IP blacklists
- `explain_score` - Detailed breakdown of spam score calculation

### Configuration Management

- `get_config` - Retrieve current SpamAssassin configuration and status

### Rule Testing

- `test_rules` - Test custom rules against sample emails in a safe environment

## 🏥 Health Monitoring

The container includes automated health checks:

```bash
# Check container health
docker exec spamassassin-mcp /usr/local/bin/health-check.sh

# View health status
docker ps
```

## 📚 Documentation

For detailed documentation, please visit the GitHub repository:
- [GitHub Repository](https://github.com/your-username/spamassassin-mcp)
- [API Reference](https://github.com/your-username/spamassassin-mcp/blob/main/docs/API.md)
- [Configuration Guide](https://github.com/your-username/spamassassin-mcp/blob/main/docs/CONFIGURATION.md)
- [Security Guide](https://github.com/your-username/spamassassin-mcp/blob/main/docs/SECURITY.md)
- [Deployment Guide](https://github.com/your-username/spamassassin-mcp/blob/main/docs/DEPLOYMENT.md)

## 📄 License

MIT License - See [LICENSE](https://github.com/your-username/spamassassin-mcp/blob/main/LICENSE) file for details.
EOF

echo "Docker Hub overview generated at dockerhub-overview.md"
echo "Please remember to:"
echo "1. Replace 'your-dockerhub-username' with your actual Docker Hub username"
echo "2. Replace 'your-username' with your actual GitHub username"