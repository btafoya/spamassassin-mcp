# SpamAssassin MCP Server

A secure, containerized Model Context Protocol (MCP) server that integrates SpamAssassin for defensive email security analysis.

## 🔒 Security-First Design

This server exclusively provides security analysis capabilities:
- ✅ Email spam detection and analysis
- ✅ Sender reputation checking
- ✅ Rule testing and validation
- ✅ Configuration inspection
- ❌ NO email sending/relay capabilities
- ❌ NO malicious content generation
- ❌ NO offensive security tools

## 🚀 Quick Start

```bash
# Pull the latest image from Docker Hub
docker pull your-dockerhub-username/spamassassin-mcp:latest

# Run the container
docker run -d \
  --name spamassassin-mcp \
  -p 8081:8080 \
  your-dockerhub-username/spamassassin-mcp:latest
```

## 🛠️ Configuration

The container can be configured using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SA_MCP_HOST_PORT` | `8081` | Host port for container deployment |
| `SA_MCP_LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `SA_MCP_SERVER_BIND_ADDR` | `0.0.0.0:8080` | Server bind address |
| `SA_MCP_SPAMASSASSIN_HOST` | `localhost` | SpamAssassin daemon host |
| `SA_MCP_SPAMASSASSIN_PORT` | `783` | SpamAssassin daemon port |
| `SA_MCP_SPAMASSASSIN_THRESHOLD` | `5.0` | Spam score threshold |

## 📋 Available Tools

### Email Analysis
- `scan_email` - Analyze email content for spam probability
- `check_reputation` - Check sender reputation and blacklists
- `explain_score` - Detailed spam score breakdown

### Configuration
- `get_config` - Retrieve current SpamAssassin configuration

## 🏥 Health Monitoring

The container includes automated health checks:
```bash
# Check container health
docker exec spamassassin-mcp /usr/local/bin/health-check.sh
```

## 📚 Documentation

For detailed documentation, please visit:
- [GitHub Repository](https://github.com/your-username/spamassassin-mcp)
- [API Reference](https://github.com/your-username/spamassassin-mcp/blob/main/docs/API.md)
- [Configuration Guide](https://github.com/your-username/spamassassin-mcp/blob/main/docs/CONFIGURATION.md)
- [Security Guide](https://github.com/your-username/spamassassin-mcp/blob/main/docs/SECURITY.md)

## 📄 License

MIT License - See [LICENSE](https://github.com/your-username/spamassassin-mcp/blob/main/LICENSE) file for details.
