# SpamAssassin MCP Server Configuration Reference

Complete reference for all configuration options, environment variables, and deployment settings.

## Table of Contents

- [Configuration Overview](#configuration-overview)
- [Configuration Sources](#configuration-sources)
- [Server Configuration](#server-configuration)
- [SpamAssassin Configuration](#spamassassin-configuration)
- [Security Configuration](#security-configuration)
- [Environment Variables](#environment-variables)
- [Docker Configuration](#docker-configuration)
- [Production Configuration](#production-configuration)
- [Configuration Validation](#configuration-validation)

## Configuration Overview

The SpamAssassin MCP server uses a hierarchical configuration system with the following precedence (highest to lowest):

1. **Environment Variables** (highest priority)
2. **YAML Configuration Files**
3. **Built-in Defaults** (lowest priority)

### Default Configuration Locations

```
/etc/spamassassin-mcp/config.yaml     # Primary configuration file
./config.yaml                         # Local development override
./configs/config.yaml                 # Project configuration
```

### Configuration Structure

```yaml
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "30s"

spamassassin:
  host: "localhost"
  port: 783
  timeout: "30s"
  threshold: 5.0

security:
  max_email_size: 10485760
  rate_limiting:
    requests_per_minute: 60
    burst_size: 10
  scan_timeout: "60s"
  validation_enabled: true
  allowed_senders: []
  blocked_domains: []

log_level: "info"
```

## Configuration Sources

### YAML Configuration Files

#### Development Configuration (`configs/config.yaml`)
```yaml
# Development environment configuration
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "30s"

spamassassin:
  host: "localhost"
  port: 783
  timeout: "30s"
  threshold: 5.0

security:
  max_email_size: 10485760  # 10MB
  rate_limiting:
    requests_per_minute: 60
    burst_size: 10
  scan_timeout: "60s"
  validation_enabled: true
  
  # Development allowlists (examples)
  allowed_senders:
    - "test@example.com"
    - "dev@localhost"
    
  # Development blocklists (examples)
  blocked_domains:
    - "spam-test.com"
    - "malware-test.net"

log_level: "debug"
```

#### Production Configuration (`configs/prod/config.yaml`)
```yaml
# Production environment configuration
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "60s"

spamassassin:
  host: "localhost"
  port: 783
  timeout: "45s"
  threshold: 5.0

security:
  max_email_size: 10485760  # 10MB
  rate_limiting:
    requests_per_minute: 120  # Higher for production
    burst_size: 20
  scan_timeout: "90s"
  validation_enabled: true
  
  # Production security lists
  allowed_senders:
    - "notifications@company.com"
    - "alerts@monitoring.company.com"
    - "security@company.com"
    
  blocked_domains:
    - "known-spam-domain.com"
    - "malicious-site.net"
    - "phishing-domain.org"

log_level: "warn"
```

#### High-Security Configuration (`configs/secure/config.yaml`)
```yaml
# High-security environment configuration
server:
  bind_addr: "127.0.0.1:8080"  # Localhost only
  timeout: "30s"

spamassassin:
  host: "localhost"
  port: 783
  timeout: "30s"
  threshold: 3.0  # Lower threshold = more sensitive

security:
  max_email_size: 5242880  # 5MB (reduced for security)
  rate_limiting:
    requests_per_minute: 30  # Stricter rate limiting
    burst_size: 5
  scan_timeout: "30s"
  validation_enabled: true
  
  # Minimal allowlist for high security
  allowed_senders:
    - "admin@company.com"
    
  # Extensive blocklist
  blocked_domains:
    - "suspicious-domain1.com"
    - "suspicious-domain2.net"
    - "known-malware.org"

log_level: "info"  # Avoid debug to prevent data leakage
```

## Server Configuration

### `server` Section

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `bind_addr` | string | `"0.0.0.0:8080"` | Address and port to bind the MCP server |
| `timeout` | duration | `"30s"` | HTTP server read/write timeout |

#### Examples

```yaml
# Listen on all interfaces, port 8080
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "30s"

# Listen on localhost only (more secure)
server:
  bind_addr: "127.0.0.1:8080"
  timeout: "60s"

# Custom port
server:
  bind_addr: "0.0.0.0:9090"
  timeout: "45s"
```

#### Security Considerations

- Use `127.0.0.1` for localhost-only access
- Increase timeout for high-latency environments
- Consider firewall rules for external access

## SpamAssassin Configuration

### `spamassassin` Section

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | string | `"localhost"` | SpamAssassin daemon hostname |
| `port` | int | `783` | SpamAssassin daemon port |
| `timeout` | duration | `"30s"` | Connection timeout for SpamAssassin |
| `threshold` | float64 | `5.0` | Spam score threshold |

#### Examples

```yaml
# Default SpamAssassin configuration
spamassassin:
  host: "localhost"
  port: 783
  timeout: "30s"
  threshold: 5.0

# Remote SpamAssassin instance
spamassassin:
  host: "spamassassin.company.com"
  port: 783
  timeout: "60s"
  threshold: 4.5

# High-sensitivity configuration
spamassassin:
  host: "localhost"
  port: 783
  timeout: "45s"
  threshold: 3.0  # Lower threshold = more sensitive
```

#### Threshold Guidelines

| Threshold | Sensitivity | Use Case |
|-----------|-------------|----------|
| `2.0-3.0` | Very High | High-security environments |
| `3.0-4.0` | High | Corporate email filtering |
| `4.0-5.0` | Medium | Balanced filtering (default) |
| `5.0-7.0` | Low | Minimal false positives |
| `7.0+` | Very Low | Research/analysis only |

## Security Configuration

### `security` Section

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `max_email_size` | int64 | `10485760` | Maximum email size in bytes (10MB) |
| `rate_limiting.requests_per_minute` | int | `60` | Requests allowed per minute |
| `rate_limiting.burst_size` | int | `10` | Burst capacity for rate limiting |
| `scan_timeout` | duration | `"60s"` | Maximum time for email scan |
| `validation_enabled` | bool | `true` | Enable input validation |
| `allowed_senders` | []string | `[]` | Whitelist of allowed email senders |
| `blocked_domains` | []string | `[]` | Blacklist of blocked domains |

#### Security Examples

```yaml
# High-security configuration
security:
  max_email_size: 5242880  # 5MB
  rate_limiting:
    requests_per_minute: 30
    burst_size: 5
  scan_timeout: "30s"
  validation_enabled: true
  allowed_senders:
    - "admin@company.com"
  blocked_domains:
    - "suspicious-domain.com"

# High-throughput configuration
security:
  max_email_size: 20971520  # 20MB
  rate_limiting:
    requests_per_minute: 300
    burst_size: 50
  scan_timeout: "120s"
  validation_enabled: true
  allowed_senders: []
  blocked_domains: []

# Development configuration
security:
  max_email_size: 1048576  # 1MB
  rate_limiting:
    requests_per_minute: 10
    burst_size: 3
  scan_timeout: "10s"
  validation_enabled: false  # Disable for testing
  allowed_senders:
    - "test@localhost"
  blocked_domains:
    - "test-spam.com"
```

#### Rate Limiting Configuration

```yaml
# Conservative rate limiting
rate_limiting:
  requests_per_minute: 30
  burst_size: 5

# Balanced rate limiting (default)
rate_limiting:
  requests_per_minute: 60
  burst_size: 10

# High-throughput rate limiting
rate_limiting:
  requests_per_minute: 300
  burst_size: 50

# Development (very permissive)
rate_limiting:
  requests_per_minute: 1000
  burst_size: 100
```

## Environment Variables

All configuration options can be overridden using environment variables with the `SA_MCP_` prefix.

### Environment Variable Format

Configuration paths are converted to environment variables using underscores:
- `server.bind_addr` → `SA_MCP_SERVER_BIND_ADDR`
- `spamassassin.host` → `SA_MCP_SPAMASSASSIN_HOST`
- `security.rate_limiting.requests_per_minute` → `SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE`

### Complete Environment Variable List

#### Server Configuration
```bash
SA_MCP_SERVER_BIND_ADDR="0.0.0.0:8080"
SA_MCP_SERVER_TIMEOUT="30s"
```

#### SpamAssassin Configuration
```bash
SA_MCP_SPAMASSASSIN_HOST="localhost"
SA_MCP_SPAMASSASSIN_PORT="783"
SA_MCP_SPAMASSASSIN_TIMEOUT="30s"
SA_MCP_SPAMASSASSIN_THRESHOLD="5.0"
```

#### Security Configuration
```bash
SA_MCP_SECURITY_MAX_EMAIL_SIZE="10485760"
SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE="60"
SA_MCP_SECURITY_RATE_LIMITING_BURST_SIZE="10"
SA_MCP_SECURITY_SCAN_TIMEOUT="60s"
SA_MCP_SECURITY_VALIDATION_ENABLED="true"
```

#### Logging Configuration
```bash
SA_MCP_LOG_LEVEL="info"
```

#### Array Environment Variables

For array values (like `allowed_senders` and `blocked_domains`), use comma-separated values:

```bash
SA_MCP_SECURITY_ALLOWED_SENDERS="admin@company.com,support@company.com,alerts@company.com"
SA_MCP_SECURITY_BLOCKED_DOMAINS="spam-domain.com,malicious-site.net,phishing-domain.org"
```

### Environment Variable Examples

#### Development Environment
```bash
# Development .env file
SA_MCP_LOG_LEVEL=debug
SA_MCP_SERVER_BIND_ADDR=0.0.0.0:8080
SA_MCP_SPAMASSASSIN_HOST=localhost
SA_MCP_SPAMASSASSIN_THRESHOLD=5.0
SA_MCP_SECURITY_VALIDATION_ENABLED=false
SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=1000
```

#### Production Environment
```bash
# Production environment variables
SA_MCP_LOG_LEVEL=warn
SA_MCP_SERVER_BIND_ADDR=0.0.0.0:8080
SA_MCP_SERVER_TIMEOUT=60s
SA_MCP_SPAMASSASSIN_HOST=spamassassin-service
SA_MCP_SPAMASSASSIN_TIMEOUT=45s
SA_MCP_SPAMASSASSIN_THRESHOLD=4.5
SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=120
SA_MCP_SECURITY_RATE_LIMITING_BURST_SIZE=20
SA_MCP_SECURITY_SCAN_TIMEOUT=90s
SA_MCP_SECURITY_ALLOWED_SENDERS="notifications@company.com,alerts@company.com"
SA_MCP_SECURITY_BLOCKED_DOMAINS="known-spam.com,malicious.net"
```

## Docker Configuration

### Docker Compose Environment

```yaml
# docker-compose.yml
services:
  spamassassin-mcp:
    environment:
      # Server configuration
      - SA_MCP_LOG_LEVEL=info
      - SA_MCP_SERVER_BIND_ADDR=0.0.0.0:8080
      - SA_MCP_SERVER_TIMEOUT=30s
      
      # SpamAssassin configuration
      - SA_MCP_SPAMASSASSIN_HOST=localhost
      - SA_MCP_SPAMASSASSIN_PORT=783
      - SA_MCP_SPAMASSASSIN_TIMEOUT=30s
      - SA_MCP_SPAMASSASSIN_THRESHOLD=5.0
      
      # Security configuration
      - SA_MCP_SECURITY_MAX_EMAIL_SIZE=10485760
      - SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=60
      - SA_MCP_SECURITY_RATE_LIMITING_BURST_SIZE=10
      - SA_MCP_SECURITY_SCAN_TIMEOUT=60s
      - SA_MCP_SECURITY_VALIDATION_ENABLED=true
```

### Docker Secrets Integration

```yaml
# docker-compose.yml with secrets
services:
  spamassassin-mcp:
    secrets:
      - source: mcp_config
        target: /etc/spamassassin-mcp/config.yaml
        mode: 0444
    environment:
      - SA_MCP_CONFIG_FILE=/etc/spamassassin-mcp/config.yaml

secrets:
  mcp_config:
    external: true
```

### Environment Files

#### `.env.development`
```bash
SA_MCP_LOG_LEVEL=debug
SA_MCP_SERVER_BIND_ADDR=0.0.0.0:8080
SA_MCP_SPAMASSASSIN_THRESHOLD=5.0
SA_MCP_SECURITY_VALIDATION_ENABLED=false
SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=1000
```

#### `.env.production`
```bash
SA_MCP_LOG_LEVEL=warn
SA_MCP_SERVER_TIMEOUT=60s
SA_MCP_SPAMASSASSIN_TIMEOUT=45s
SA_MCP_SPAMASSASSIN_THRESHOLD=4.5
SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=120
SA_MCP_SECURITY_RATE_LIMITING_BURST_SIZE=20
SA_MCP_SECURITY_SCAN_TIMEOUT=90s
```

## Production Configuration

### High-Availability Configuration

```yaml
# High-availability production configuration
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "60s"

spamassassin:
  host: "spamassassin-cluster.company.com"
  port: 783
  timeout: "45s"
  threshold: 4.5

security:
  max_email_size: 15728640  # 15MB
  rate_limiting:
    requests_per_minute: 200
    burst_size: 30
  scan_timeout: "120s"
  validation_enabled: true
  
  allowed_senders:
    - "notifications@company.com"
    - "alerts@monitoring.company.com"
    - "security@company.com"
    - "noreply@company.com"
    
  blocked_domains:
    # Known spam/malware domains
    - "known-spam-domain1.com"
    - "known-spam-domain2.net"
    - "malware-distribution.org"
    - "phishing-site1.com"
    - "phishing-site2.net"

log_level: "warn"
```

### Multi-Environment Configuration

#### Environment-Specific Overrides

```bash
# config/environments/development.yaml
log_level: "debug"
security:
  validation_enabled: false
  rate_limiting:
    requests_per_minute: 1000

# config/environments/staging.yaml
log_level: "info"
spamassassin:
  threshold: 4.0
security:
  rate_limiting:
    requests_per_minute: 100

# config/environments/production.yaml
log_level: "warn"
server:
  timeout: "60s"
spamassassin:
  threshold: 4.5
  timeout: "45s"
security:
  rate_limiting:
    requests_per_minute: 200
    burst_size: 30
```

### Load Balancer Configuration

For load-balanced deployments:

```yaml
# Load balancer backend configuration
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "30s"  # Keep lower for health checks

security:
  rate_limiting:
    requests_per_minute: 300  # Higher per instance
    burst_size: 50
  scan_timeout: "90s"

# Health check endpoint optimization
health_check:
  enabled: true
  path: "/health"
  timeout: "5s"
```

## Configuration Validation

### Validation Rules

The server validates configuration on startup:

1. **Network Configuration**: Validates bind addresses and ports
2. **Timeout Values**: Ensures reasonable timeout ranges
3. **Size Limits**: Validates email size limits
4. **Rate Limits**: Ensures positive rate limiting values
5. **Security Settings**: Validates security configuration

### Validation Script

```bash
#!/bin/bash
# validate-config.sh

echo "Validating SpamAssassin MCP configuration..."

# Check if configuration files exist
CONFIG_FILES=(
    "configs/config.yaml"
    "configs/prod/config.yaml"
)

for file in "${CONFIG_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✓ Found $file"
        # Validate YAML syntax
        if ! yaml-validator "$file" 2>/dev/null; then
            echo "✗ Invalid YAML syntax in $file"
            exit 1
        fi
    else
        echo "⚠ Missing $file (using defaults)"
    fi
done

# Test configuration loading
if docker-compose exec spamassassin-mcp mcp-server --validate-config; then
    echo "✓ Configuration validation passed"
else
    echo "✗ Configuration validation failed"
    exit 1
fi

# Test SpamAssassin connectivity
if docker-compose exec spamassassin-mcp nc -z localhost 783; then
    echo "✓ SpamAssassin connectivity OK"
else
    echo "✗ Cannot connect to SpamAssassin"
    exit 1
fi

echo "Configuration validation completed successfully"
```

### Common Configuration Errors

#### Invalid YAML Syntax
```yaml
# Bad: Missing quotes around time values
timeout: 30s

# Good: Proper string quoting
timeout: "30s"
```

#### Invalid Environment Variables
```bash
# Bad: Invalid boolean value
SA_MCP_SECURITY_VALIDATION_ENABLED=yes

# Good: Proper boolean values
SA_MCP_SECURITY_VALIDATION_ENABLED=true
```

#### Resource Limits
```yaml
# Bad: Unrealistic values
security:
  max_email_size: 1000000000  # 1GB - too large
  scan_timeout: "10m"         # 10 minutes - too long

# Good: Reasonable limits
security:
  max_email_size: 10485760    # 10MB
  scan_timeout: "60s"         # 60 seconds
```

### Configuration Testing

#### Unit Tests
```bash
# Test configuration loading
go test ./internal/config -v

# Test with different environments
SA_MCP_LOG_LEVEL=debug go test ./internal/config -v
```

#### Integration Tests
```bash
# Test full configuration stack
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Test with production configuration
docker-compose -f docker-compose.prod.yml config
```

### Environment-Specific Validation

```bash
# Validate development configuration
SA_MCP_LOG_LEVEL=debug mcp-server --validate-config

# Validate production configuration
SA_MCP_LOG_LEVEL=warn \
SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=120 \
mcp-server --validate-config
```

## Troubleshooting Configuration Issues

### Common Issues

1. **Server Won't Start**
   - Check YAML syntax
   - Verify port availability
   - Check file permissions

2. **SpamAssassin Connection Failed**
   - Verify SpamAssassin is running
   - Check host/port configuration
   - Test network connectivity

3. **Rate Limiting Too Strict**
   - Increase `requests_per_minute`
   - Increase `burst_size`
   - Check client connection patterns

4. **High Memory Usage**
   - Reduce `max_email_size`
   - Decrease `scan_timeout`
   - Optimize SpamAssassin configuration

### Debug Configuration

```yaml
# Debug configuration for troubleshooting
log_level: "debug"
server:
  timeout: "120s"  # Longer timeout for debugging
spamassassin:
  timeout: "60s"
security:
  validation_enabled: false  # Disable for testing
  rate_limiting:
    requests_per_minute: 1000  # Very permissive
    burst_size: 100
```

### Configuration Monitoring

```bash
# Monitor configuration changes
watch -n 5 'docker-compose exec spamassassin-mcp env | grep SA_MCP_'

# Check effective configuration
docker-compose exec spamassassin-mcp mcp-server --show-config
```

For additional configuration support, refer to the [Troubleshooting Guide](TROUBLESHOOTING.md) or [Development Guide](DEVELOPMENT.md).