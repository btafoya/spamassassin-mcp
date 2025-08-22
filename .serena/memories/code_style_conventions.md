# Code Style and Conventions

## Go Code Style

### General Principles
- Follow standard Go conventions from `gofmt`, `go vet`, and the Go Code Review Comments
- Use clear, descriptive names for variables, functions, and types
- Prefer composition over inheritance
- Keep functions small and focused on a single responsibility

### Naming Conventions
- **Packages**: Short, lowercase, single words when possible
- **Variables**: camelCase (e.g., `emailContent`, `maxRetries`)
- **Constants**: CamelCase for exported, camelCase for unexported
- **Functions**: CamelCase for exported, camelCase for unexported
- **Structs**: CamelCase for exported types
- **Interfaces**: Often end with "er" suffix (e.g., `Handler`, `Scanner`)

### Code Organization
```go
// Package declaration
package handlers

// Imports (standard library first, then third-party, then local)
import (
    "context"
    "fmt"
    
    "github.com/sirupsen/logrus"
    
    "spamassassin-mcp/internal/config"
)

// Constants
const (
    MaxEmailSize = 10485760 // 10MB
    DefaultTimeout = "30s"
)

// Types (exported first, then unexported)
type Handler struct {
    client SpamAssassinClient
    config *config.Config
    logger *logrus.Logger
}

// Functions (exported first, then unexported)
```

### Error Handling
- Always handle errors explicitly
- Use descriptive error messages with context
- Wrap errors with additional context when appropriate
- Log errors at appropriate levels

```go
// Good error handling
result, err := client.ScanEmail(ctx, content)
if err != nil {
    h.logger.WithError(err).Error("Failed to scan email")
    return nil, fmt.Errorf("email scan failed: %w", err)
}
```

### Documentation
- Use standard Go doc comments for exported functions and types
- Start comments with the name of the item being documented
- Keep comments concise but informative

```go
// ScanEmail analyzes email content for spam probability using SpamAssassin.
// It returns a detailed analysis including score, rules matched, and explanations.
func (h *Handler) ScanEmail(ctx context.Context, params ScanEmailParams) (*ScanEmailResult, error) {
    // Implementation
}
```

## Project-Specific Conventions

### Configuration
- Use Viper for configuration management
- Support both YAML files and environment variables
- Environment variables prefixed with `SA_MCP_`
- Provide sensible defaults for all configuration options

### Logging
- Use logrus for structured logging
- Include relevant context fields in log entries
- Log levels: debug, info, warn, error
- Never log sensitive email content in production

### Security
- Validate all inputs at entry points
- Use rate limiting for all public endpoints
- Implement proper timeout handling
- Follow principle of least privilege

### MCP Integration
- All tools must be registered in `registerTools()` function
- Use consistent parameter and result structures
- Implement proper context handling for cancellation
- Return descriptive errors for user feedback

## File Structure Conventions

### Directory Layout
```
internal/
├── config/          # Configuration management
├── handlers/        # MCP tool implementations  
└── spamassassin/    # SpamAssassin client wrapper

configs/             # Configuration files
scripts/             # Shell scripts and utilities
examples/            # Usage examples
docs/                # Documentation
```

### File Naming
- Use lowercase with underscores for scripts: `health_check.sh`
- Use lowercase for Go packages: `config`, `handlers`
- Use descriptive names that indicate purpose

## Testing Conventions (Future)
- Test files end with `_test.go`
- Use table-driven tests where appropriate
- Mock external dependencies (SpamAssassin client)
- Achieve >80% code coverage for critical paths
- Include integration tests for MCP functionality