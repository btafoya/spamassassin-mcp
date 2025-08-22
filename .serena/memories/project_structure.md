# Project Structure

## Root Directory Layout
```
spamassassin-mcp/
├── main.go                 # MCP server entry point and tool registration
├── go.mod                  # Go module definition with dependencies
├── go.sum                  # Go module checksums
├── Dockerfile              # Multi-stage container build definition
├── docker-compose.yml      # Service orchestration configuration
├── .mcp.json              # MCP server metadata
├── CLAUDE.md              # Claude Code project configuration
├── README.md              # Comprehensive project documentation
├── LICENSE                # MIT license
└── .gitignore             # Git ignore patterns
```

## Source Code Structure
```
internal/                   # Internal Go packages (not importable)
├── config/
│   └── config.go          # Configuration management with Viper
├── handlers/
│   └── handlers.go        # MCP tool implementations
└── spamassassin/
    └── client.go          # SpamAssassin daemon client wrapper
```

## Configuration and Scripts
```
configs/
└── config.yaml            # Default YAML configuration

scripts/
├── entrypoint.sh          # Container initialization script
└── health-check.sh        # Health monitoring script
```

## Documentation
```
docs/                       # Detailed documentation (referenced in README)
├── API.md                 # MCP tools API reference
├── DEPLOYMENT.md          # Production deployment guide
├── SECURITY.md            # Security architecture documentation
├── DEVELOPMENT.md         # Development and contributing guide
├── TROUBLESHOOTING.md     # Common issues and solutions
└── CONFIGURATION.md       # Complete configuration reference
```

## Examples and Development
```
examples/                   # Usage examples and sample files
└── (sample email files and configuration examples)

.claude/                    # Claude Code configuration
.serena/                   # Serena project configuration
```

## Key Components

### Main Application (`main.go`)
- MCP server initialization
- Logging setup with logrus
- Tool registration for all available MCP tools
- Signal handling and graceful shutdown

### Configuration (`internal/config/`)
- YAML and environment variable configuration
- Viper-based configuration management
- Security settings and validation
- SpamAssassin connection parameters

### Handlers (`internal/handlers/`)
- MCP tool implementations:
  - `ScanEmail` - Email spam analysis
  - `CheckReputation` - Sender reputation checking
  - `ExplainScore` - Detailed score breakdown
  - `GetConfig` - Configuration retrieval
  - `UpdateRules` - Rule management
  - `TestRules` - Rule testing in safe environment
- Input validation and security checks
- Error handling and user-friendly responses

### SpamAssassin Client (`internal/spamassassin/`)
- TCP communication with SpamAssassin daemon
- Protocol handling for SPAMC protocol
- Connection pooling and timeout management
- Result parsing and error handling

### Container Configuration
- **Dockerfile**: Multi-stage build with security hardening
- **docker-compose.yml**: Service orchestration with security constraints
- **entrypoint.sh**: SpamAssassin daemon initialization and health checks
- **health-check.sh**: Container health monitoring

## Security Architecture
- Non-root container execution
- Read-only filesystem with specific writable tmpfs
- Network isolation with custom bridge
- Resource limits (512MB memory, 0.5 CPU)
- Rate limiting and input validation
- No persistent storage of email content