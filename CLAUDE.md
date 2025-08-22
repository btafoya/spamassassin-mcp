# SpamAssassin MCP Server - Claude Code Configuration

## Project Overview
A secure, containerized Model Context Protocol (MCP) server that integrates SpamAssassin for defensive email security analysis. Provides Claude Code with comprehensive email analysis capabilities while maintaining strict security boundaries.

## Quick Commands

### Development
```bash
# Build and run locally
go build -o mcp-server main.go && ./mcp-server

# Format and validate code
go fmt ./... && go vet ./...

# Manage dependencies
go mod tidy && go mod download
```

### Container Operations
```bash
# Start services
docker-compose up -d

# View logs and health
docker-compose logs -f spamassassin-mcp
docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh

# Rebuild after changes
docker-compose up -d --build
```

### Testing MCP Integration
```bash
# Test email scanning
/scan_email --content "Subject: Test Email\n\nThis is a test email for analysis."

# Check sender reputation
/check_reputation --sender "test@example.com"

# Get current configuration
/get_config
```

## Task Completion Checklist
When completing development tasks:

1. **Code Quality**: `go fmt ./...` ’ `go vet ./...` ’ `go build`
2. **Container Test**: `docker-compose up -d --build`
3. **Health Check**: Verify `/usr/local/bin/health-check.sh` passes
4. **MCP Tools**: Test tool responses with sample inputs
5. **Security**: Ensure no sensitive data in logs or responses

## Tech Stack
- **Language**: Go 1.23.0+ with toolchain 1.24.4
- **Framework**: Model Context Protocol Go SDK v0.2.0
- **Config**: Viper with YAML/environment variables
- **Logging**: Logrus structured logging
- **Container**: Docker with security hardening
- **Integration**: SpamAssassin daemon via TCP

## Available MCP Tools
1. `scan_email` - Analyze email content for spam probability and rule matches
2. `check_reputation` - Check sender reputation and domain/IP blacklists  
3. `explain_score` - Detailed breakdown of spam score calculation
4. `get_config` - Retrieve current SpamAssassin configuration
5. `update_rules` - Update SpamAssassin rule definitions (defensive only)
6. `test_rules` - Test custom rules in safe sandbox environment

## Security Architecture
- **Defensive Only**: No email sending, relay, or malicious content generation
- **Container Security**: Non-root execution, read-only filesystem, resource limits
- **Input Validation**: Size limits (10MB), format validation, sanitization
- **Rate Limiting**: 60 requests/minute with burst of 10
- **Network Isolation**: Custom bridge network with controlled access

## Development Notes
- All email content validation happens at handler level
- SpamAssassin client uses connection pooling for performance
- Configuration supports both YAML files and environment variables
- Health checks verify both MCP server and SpamAssassin daemon status
- Logging configured for structured output with configurable levels

## Environment Variables
- `SA_MCP_LOG_LEVEL`: Logging level (debug, info, warn, error)
- `SA_MCP_SERVER_BIND_ADDR`: Server bind address (default: 0.0.0.0:8080)
- `SA_MCP_SPAMASSASSIN_HOST`: SpamAssassin daemon host (default: localhost)
- `SA_MCP_SPAMASSASSIN_PORT`: SpamAssassin daemon port (default: 783)
- `SA_MCP_SECURITY_MAX_EMAIL_SIZE`: Maximum email size in bytes (default: 10MB)