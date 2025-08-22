# Suggested Commands for SpamAssassin MCP Server

## Development Commands

### Build and Run
```bash
# Build the Go binary
go build -o mcp-server main.go

# Run locally (requires SpamAssassin daemon)
./mcp-server

# Build with race detection for development
go build -race -o mcp-server main.go
```

### Dependencies Management
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
```

### Testing and Quality
```bash
# Run tests (when available)
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run static analysis (if golangci-lint installed)
golangci-lint run
```

## Container Operations

### Docker Commands
```bash
# Build container
docker build -t spamassassin-mcp .

# Run container standalone
docker run -p 8080:8080 spamassassin-mcp

# Start services with Docker Compose
docker-compose up -d

# Start with testing profile (includes spamd)
docker-compose --profile testing up -d

# View logs
docker-compose logs -f spamassassin-mcp

# Stop services
docker-compose down

# Rebuild and restart
docker-compose up -d --build
```

### Health and Monitoring
```bash
# Check container health
docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh

# Test SpamAssassin connectivity
docker-compose exec spamassassin-mcp nc -z localhost 783

# Monitor resource usage
docker stats spamassassin-mcp

# Check running processes
docker-compose exec spamassassin-mcp pgrep spamd
```

## Configuration and Environment

### Environment Variables
```bash
# Set log level for debugging
export SA_MCP_LOG_LEVEL=debug

# Modify server binding
export SA_MCP_SERVER_BIND_ADDR=localhost:8080

# Adjust security settings
export SA_MCP_SECURITY_MAX_EMAIL_SIZE=5242880  # 5MB
export SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=30
```

## Git Workflow
```bash
# Standard Git operations
git status
git add .
git commit -m "feat: descriptive commit message"
git push origin main

# Feature branch workflow
git checkout -b feature/new-tool
git push -u origin feature/new-tool
```

## Utility Commands (Linux)
```bash
# File operations
ls -la
find . -name "*.go" -type f
grep -r "pattern" .
grep -r "pattern" --include="*.go" .

# Process management
ps aux | grep mcp-server
pkill -f mcp-server
netstat -tlnp | grep 8080

# System monitoring
top
htop
df -h
free -h
```

## Claude Code Integration
```bash
# Add MCP server to Claude Code configuration
claude --mcp-server spamassassin tcp://localhost:8080

# Test MCP tools
/scan_email --content "Subject: Test\n\nTest email content"
/check_reputation --sender "test@example.com"
/get_config
```