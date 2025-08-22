# SpamAssassin MCP Server

A secure, containerized Model Context Protocol (MCP) server that integrates SpamAssassin for defensive email security analysis. This server provides Claude Code with comprehensive email analysis capabilities while maintaining strict security boundaries.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/Docker-supported-blue.svg)](https://www.docker.com/)
[![Security](https://img.shields.io/badge/Security-Defensive--Only-green.svg)](docs/SECURITY.md)

## 🛡️ Security-First Design

**Defensive Operations Only** - This server exclusively provides security analysis capabilities:
- ✅ Email spam detection and analysis
- ✅ Sender reputation checking
- ✅ Rule testing and validation
- ✅ Configuration inspection
- ❌ NO email sending/relay capabilities
- ❌ NO malicious content generation
- ❌ NO offensive security tools

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Claude Code with MCP support

### 1. Build and Start
```bash
# Clone or create the project directory
cd spamassassin-mcp

# Build and start the containers
docker-compose up -d

# Check health
docker-compose logs spamassassin-mcp
```

### 2. Connect Claude Code
```bash
# Add to your MCP configuration
claude --mcp-server spamassassin tcp://localhost:8080
```

### 3. Test the Integration
```bash
# Scan a sample email
/scan_email --content "Subject: Test Email

This is a test email for spam analysis."

# Check sender reputation
/check_reputation --sender "test@example.com"

# Get current configuration
/get_config
```

## 🔧 Available Tools

### Email Analysis

#### `scan_email`
Analyze email content for spam probability and rule matches.

**Parameters:**
- `content` (required): Raw email content including headers
- `headers` (optional): Additional headers to analyze
- `check_bayes` (optional): Include Bayesian analysis
- `verbose` (optional): Return detailed rule explanations

**Example:**
```json
{
  "content": "Subject: Urgent Action Required\\n\\nClick here to claim your prize!",
  "verbose": true,
  "check_bayes": true
}
```

#### `check_reputation`
Check sender reputation and domain/IP blacklists.

**Parameters:**
- `sender` (required): Email sender address
- `domain` (optional): Sender domain
- `ip` (optional): Sender IP address

#### `explain_score`
Explain how a spam score was calculated with detailed breakdown.

### Configuration Management

#### `get_config`
Retrieve current SpamAssassin configuration and status.

#### `update_rules`
Update SpamAssassin rule definitions (defensive updates only).

**Parameters:**
- `source` (optional): Rule source (official/custom)
- `force` (optional): Force update even if recent

### Rule Testing

#### `test_rules`
Test custom rules against sample emails in a safe environment.

**Parameters:**
- `rules` (required): Custom rule definitions
- `test_emails` (required): Array of sample emails to test

## 📁 Project Structure

```
spamassassin-mcp/
├── main.go                 # MCP server entry point
├── go.mod                  # Go module definition
├── Dockerfile              # Multi-stage container build
├── docker-compose.yml      # Service orchestration
├── internal/
│   ├── config/            # Configuration management
│   ├── handlers/          # MCP tool handlers
│   └── spamassassin/      # SpamAssassin client wrapper
├── configs/
│   └── config.yaml        # Server configuration
├── scripts/
│   ├── entrypoint.sh      # Container initialization
│   └── health-check.sh    # Health monitoring
└── README.md
```

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SA_MCP_LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `SA_MCP_SERVER_BIND_ADDR` | `0.0.0.0:8080` | Server bind address |
| `SA_MCP_SPAMASSASSIN_HOST` | `localhost` | SpamAssassin daemon host |
| `SA_MCP_SPAMASSASSIN_PORT` | `783` | SpamAssassin daemon port |
| `SA_MCP_SPAMASSASSIN_THRESHOLD` | `5.0` | Spam score threshold |
| `SA_MCP_SECURITY_MAX_EMAIL_SIZE` | `10485760` | Max email size (10MB) |

### Security Settings

The server includes comprehensive security measures:

- **Rate Limiting**: 60 requests/minute with burst of 10
- **Input Validation**: Email format and size validation
- **Content Sanitization**: Safe handling of email content
- **Container Security**: Non-root execution, read-only filesystem
- **Network Isolation**: Custom bridge network
- **Resource Limits**: Memory and CPU constraints

## 🔍 Usage Examples

### Basic Email Scanning
```bash
# Simple spam check
/scan_email --content "$(cat suspicious_email.eml)"

# Detailed analysis with Bayes
/scan_email --content "$(cat email.eml)" --verbose --check_bayes
```

### Reputation Analysis
```bash
# Check sender reputation
/check_reputation --sender "unknown@suspicious-domain.com"

# Check domain and IP
/check_reputation --domain "suspicious-domain.com" --ip "192.168.1.100"
```

### Rule Development
```bash
# Test custom rules
/test_rules --rules "header LOCAL_TEST Subject =~ /test/i
describe LOCAL_TEST Test rule
score LOCAL_TEST 2.0" --test_emails '["Subject: test email\n\nThis is a test."]'
```

### Score Analysis
```bash
# Get detailed score explanation
/explain_score --email_content "Subject: Free Money!\n\nClaim your prize now!"
```

## 🏥 Health Monitoring

### Health Check Endpoint
The container includes automated health checks:

```bash
# Check container health
docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh

# View health status
docker ps
```

### Logs and Monitoring
```bash
# View server logs
docker-compose logs -f spamassassin-mcp

# Monitor resource usage
docker stats spamassassin-mcp
```

## 🔒 Security Considerations

### Defensive Posture
- Server exclusively provides analysis capabilities
- No email transmission or relay functionality
- Input validation and sanitization on all endpoints
- Rate limiting to prevent abuse
- Comprehensive logging for audit trails

### Container Security
- Runs as non-root user (`spamassassin`)
- Read-only root filesystem
- No new privileges allowed
- Resource limits enforced
- Network isolation

### Data Handling
- No persistent storage of email content
- Temporary analysis only
- Configurable retention policies
- GDPR/privacy-compliant design

## 🚧 Development

### Building from Source
```bash
# Install dependencies
go mod download

# Build binary
go build -o mcp-server main.go

# Run locally (requires SpamAssassin)
./mcp-server
```

### Testing
```bash
# Run with testing profile (includes spamd)
docker-compose --profile testing up -d

# Test SpamAssassin connectivity
docker-compose exec spamassassin-mcp nc -z localhost 783
```

## 📚 Documentation

- **[API Reference](docs/API.md)** - Complete MCP tools documentation
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment instructions
- **[Security Guide](docs/SECURITY.md)** - Security architecture and best practices
- **[Development Guide](docs/DEVELOPMENT.md)** - Contributing and development setup
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Configuration Reference](docs/CONFIGURATION.md)** - Complete configuration options

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions welcome! Please read our [Contributing Guide](docs/DEVELOPMENT.md#contributing) and ensure all changes maintain the security-first, defensive-only design principles.

## 📞 Support

For issues and questions:
1. **First Steps**: Check the [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
2. **Logs**: `docker-compose logs spamassassin-mcp`
3. **Health Check**: `docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh`
4. **SpamAssassin Status**: `docker-compose exec spamassassin-mcp pgrep spamd`

## 🔗 Related Projects

- [Model Context Protocol](https://github.com/modelcontextprotocol) - Official MCP specifications
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) - Claude's official CLI
- [SpamAssassin](https://spamassassin.apache.org/) - Open source spam filtering

---

**⚠️ Security Notice**: This server is designed exclusively for defensive security analysis. It does not provide capabilities for sending emails, generating spam content, or any offensive security operations. All operations are logged and auditable.