# SpamAssassin MCP Server Development Guide

Complete guide for developers contributing to the SpamAssassin MCP server project.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Architecture](#project-architecture)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Contributing Workflow](#contributing-workflow)
- [Release Process](#release-process)
- [Debugging and Profiling](#debugging-and-profiling)

## Development Setup

### Prerequisites

- **Go 1.21+**: [Installation Guide](https://golang.org/doc/install)
- **Docker 20.10+**: [Installation Guide](https://docs.docker.com/get-docker/)
- **Docker Compose 2.0+**: [Installation Guide](https://docs.docker.com/compose/install/)
- **Git**: Version control
- **Make**: Build automation (optional)

### Local Development Environment

#### 1. Clone and Setup
```bash
# Clone the repository
git clone <repository-url>
cd spamassassin-mcp

# Install Go dependencies
go mod download

# Verify Go installation
go version
go mod verify
```

#### 2. Development Dependencies
```bash
# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/scs/v2/cmd/gosec@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Install testing tools
go install gotest.tools/gotestsum@latest
go install github.com/matryer/moq@latest
```

#### 3. IDE Configuration

**VS Code Settings** (`.vscode/settings.json`):
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "go.buildTags": "integration",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

**GoLand/IntelliJ Configuration**:
- Enable Go modules support
- Configure golangci-lint as external tool
- Set goimports as formatter
- Enable race detection in test configurations

### Running Development Environment

#### Option 1: Local Development (without Docker)
```bash
# Start SpamAssassin daemon (requires local installation)
sudo spamd --create-prefs --max-children 5 --listen-ip 127.0.0.1

# Run the MCP server
go run main.go

# In another terminal, test the server
curl http://localhost:8080/health
```

#### Option 2: Docker Development
```bash
# Start development environment
docker-compose up -d

# Access running container for debugging
docker-compose exec spamassassin-mcp /bin/bash

# View logs
docker-compose logs -f spamassassin-mcp
```

#### Option 3: Hybrid Development
```bash
# Start only SpamAssassin in Docker
docker-compose up -d spamd

# Run MCP server locally for faster iteration
SA_MCP_SPAMASSASSIN_HOST=localhost go run main.go
```

## Project Architecture

### Directory Structure
```
spamassassin-mcp/
├── main.go                    # Application entry point
├── go.mod                     # Go module definition
├── go.sum                     # Go module checksums
├── Dockerfile                 # Container definition
├── docker-compose.yml         # Local development
├── docker-compose.prod.yml    # Production deployment
├── 
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   │   └── config.go         # Config loading and validation
│   ├── handlers/             # MCP tool handlers
│   │   ├── handlers.go       # HTTP request handlers
│   │   └── handlers_test.go  # Handler unit tests
│   └── spamassassin/         # SpamAssassin integration
│       ├── client.go         # SpamAssassin client
│       ├── client_test.go    # Client unit tests
│       └── mock.go           # Mock client for testing
│
├── configs/                   # Configuration files
│   ├── config.yaml           # Default configuration
│   └── prod/                 # Production configurations
│
├── scripts/                   # Utility scripts
│   ├── entrypoint.sh         # Container entrypoint
│   ├── health-check.sh       # Health monitoring
│   └── build.sh              # Build automation
│
├── examples/                  # Usage examples
│   ├── sample-email.eml      # Test email samples
│   └── mcp-client-test.sh    # Testing script
│
├── docs/                      # Documentation
│   ├── API.md                # API reference
│   ├── DEPLOYMENT.md         # Deployment guide
│   ├── SECURITY.md           # Security documentation
│   └── DEVELOPMENT.md        # This file
│
└── tests/                     # Test files
    ├── integration/          # Integration tests
    ├── fixtures/             # Test data
    └── mocks/                # Mock implementations
```

### Package Architecture

#### `main` Package
- Application bootstrap and dependency injection
- Signal handling and graceful shutdown
- MCP server initialization and tool registration

#### `internal/config` Package
- Configuration loading from files and environment
- Validation and default value management
- Type-safe configuration structures

#### `internal/handlers` Package
- MCP tool implementation
- Request/response handling
- Input validation and error handling
- Rate limiting and security controls

#### `internal/spamassassin` Package
- SpamAssassin protocol client
- Email scanning and analysis
- Result parsing and formatting
- Connection management and retry logic

### Data Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ Claude Code │───▶│ MCP Server  │───▶│  Handlers   │───▶│SpamAssassin │
│   Client    │    │   (main)    │    │ (internal)  │    │   Client    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       ▲                   ▲                   ▲                   ▲
       │                   │                   │                   │
       │            ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
       └────────────│   Security  │◀───│ Rate Limiter│◀───│    spamd    │
                    │ Validation  │    │   & Auth    │    │  (daemon)   │
                    └─────────────┘    └─────────────┘    └─────────────┘
```

## Coding Standards

### Go Style Guidelines

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html) guidelines.

#### Key Principles

1. **Simplicity**: Prefer simple, readable code over clever optimizations
2. **Consistency**: Follow established patterns in the codebase
3. **Security**: Always validate inputs and handle errors securely
4. **Performance**: Write efficient code but prioritize correctness
5. **Documentation**: Document public APIs and complex logic

#### Code Formatting

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Lint code
golangci-lint run

# Check for security issues
gosec ./...
```

#### Naming Conventions

```go
// Good: Clear, descriptive names
type EmailScanner struct {
    client          *spamassassin.Client
    rateLimiter     *rate.Limiter
    securityConfig  config.SecurityConfig
}

func (s *EmailScanner) ScanEmailForSpam(ctx context.Context, email string) (*ScanResult, error) {
    // Implementation
}

// Bad: Unclear, abbreviated names
type ES struct {
    c  *spamassassin.Client
    rl *rate.Limiter
    sc config.SecurityConfig
}

func (s *ES) Scan(ctx context.Context, e string) (*ScanResult, error) {
    // Implementation
}
```

#### Error Handling

```go
// Good: Descriptive error messages with context
func (h *Handler) ScanEmail(ctx context.Context, params json.RawMessage) (any, error) {
    var req ScanEmailParams
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, fmt.Errorf("failed to parse scan email parameters: %w", err)
    }
    
    if err := h.validateEmailContent(req.Content); err != nil {
        return nil, fmt.Errorf("email validation failed: %w", err)
    }
    
    result, err := h.saClient.ScanEmail(req.Content, options)
    if err != nil {
        return nil, fmt.Errorf("SpamAssassin scan failed: %w", err)
    }
    
    return result, nil
}

// Bad: Generic error messages without context
func (h *Handler) ScanEmail(ctx context.Context, params json.RawMessage) (any, error) {
    var req ScanEmailParams
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, err
    }
    
    if err := h.validateEmailContent(req.Content); err != nil {
        return nil, err
    }
    
    result, err := h.saClient.ScanEmail(req.Content, options)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

#### Logging Standards

```go
import "github.com/sirupsen/logrus"

// Good: Structured logging with context
func (h *Handler) ScanEmail(ctx context.Context, params json.RawMessage) (any, error) {
    logger := logrus.WithFields(logrus.Fields{
        "operation": "scan_email",
        "request_id": getRequestID(ctx),
    })
    
    logger.Info("Processing email scan request")
    
    // ... processing ...
    
    logger.WithFields(logrus.Fields{
        "score": result.Score,
        "is_spam": result.IsSpam,
        "rules_count": len(result.RulesHit),
        "duration_ms": time.Since(start).Milliseconds(),
    }).Info("Email scan completed")
    
    return result, nil
}
```

### Documentation Standards

#### Package Documentation
```go
// Package handlers implements MCP tool handlers for the SpamAssassin MCP server.
//
// This package provides HTTP request handlers that implement the Model Context
// Protocol (MCP) tools for email security analysis. All handlers include
// comprehensive input validation, rate limiting, and security controls.
//
// Security Notice: This package implements defensive-only operations. No email
// transmission or malicious content generation capabilities are provided.
package handlers
```

#### Function Documentation
```go
// ScanEmail analyzes email content for spam probability using SpamAssassin.
//
// This function validates the input email content, applies rate limiting,
// and performs spam analysis using the configured SpamAssassin instance.
// All processing is done in-memory with no persistent storage.
//
// Parameters:
//   - ctx: Request context for timeout and cancellation
//   - params: JSON-encoded ScanEmailParams containing email content and options
//
// Returns:
//   - ScanEmailResult containing spam score, classification, and rule details
//   - Error if validation fails, rate limit exceeded, or scan fails
//
// Security: Input is validated for size (max 10MB) and format. No email
// content is logged or stored permanently.
func (h *Handler) ScanEmail(ctx context.Context, params json.RawMessage) (any, error) {
    // Implementation
}
```

#### Type Documentation
```go
// ScanEmailParams defines the parameters for email spam analysis.
//
// All fields are validated for security and format compliance before processing.
type ScanEmailParams struct {
    // Content is the raw email including headers (required, max 10MB)
    Content string `json:"content" description:"Raw email content including headers"`
    
    // Headers provides additional headers for analysis (optional)
    Headers map[string]string `json:"headers,omitempty" description:"Additional headers to analyze"`
    
    // CheckBayes enables Bayesian spam analysis (optional, default: false)
    CheckBayes bool `json:"check_bayes,omitempty" description:"Include Bayesian analysis"`
    
    // Verbose returns detailed rule explanations (optional, default: false)
    Verbose bool `json:"verbose,omitempty" description:"Return detailed rule explanations"`
}
```

## Testing Guidelines

### Testing Strategy

We use a multi-layered testing approach:

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **Container Tests**: Test containerized deployment
4. **Security Tests**: Test security controls and validation
5. **Performance Tests**: Test under load conditions

### Unit Testing

#### Test Structure
```go
func TestHandler_ScanEmail(t *testing.T) {
    tests := []struct {
        name    string
        params  ScanEmailParams
        mock    func(*MockSpamAssassinClient)
        want    *ScanEmailResult
        wantErr bool
    }{
        {
            name: "valid email returns scan result",
            params: ScanEmailParams{
                Content: "From: test@example.com\nTo: user@example.com\nSubject: Test\n\nTest content",
                Verbose: true,
            },
            mock: func(m *MockSpamAssassinClient) {
                m.EXPECT().ScanEmail(gomock.Any(), gomock.Any()).Return(&spamassassin.ScanResult{
                    Score:     2.1,
                    Threshold: 5.0,
                    IsSpam:    false,
                    RulesHit:  []spamassassin.RuleMatch{},
                }, nil)
            },
            want: &ScanEmailResult{
                Score:     2.1,
                Threshold: 5.0,
                IsSpam:    false,
                RulesHit:  []spamassassin.RuleMatch{},
            },
            wantErr: false,
        },
        {
            name: "invalid email format returns error",
            params: ScanEmailParams{
                Content: "invalid email content",
            },
            mock:    func(m *MockSpamAssassinClient) {},
            want:    nil,
            wantErr: true,
        },
        {
            name: "email too large returns error",
            params: ScanEmailParams{
                Content: strings.Repeat("x", 11*1024*1024), // 11MB
            },
            mock:    func(m *MockSpamAssassinClient) {},
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockClient := NewMockSpamAssassinClient(t)
            tt.mock(mockClient)
            
            handler := &Handler{
                saClient: mockClient,
                security: config.SecurityConfig{
                    MaxEmailSize: 10 * 1024 * 1024,
                },
                rateLimiter: rate.NewLimiter(rate.Inf, 1),
            }
            
            // Execute
            params, _ := json.Marshal(tt.params)
            got, err := handler.ScanEmail(context.Background(), params)
            
            // Assert
            if (err != nil) != tt.wantErr {
                t.Errorf("ScanEmail() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            
            if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ScanEmail() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Table-Driven Tests
```go
func TestValidateEmailAddress(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"valid email with subdomain", "user@mail.example.com", false},
        {"valid email with plus", "user+tag@example.com", false},
        {"invalid email no @", "userexample.com", true},
        {"invalid email no domain", "user@", true},
        {"invalid email no TLD", "user@example", true},
        {"empty email", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if err := validateEmailAddress(tt.email); (err != nil) != tt.wantErr {
                t.Errorf("validateEmailAddress() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Testing

#### SpamAssassin Integration
```go
func TestSpamAssassinIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup test SpamAssassin instance
    client, err := spamassassin.NewClient(config.SpamAssassinConfig{
        Host:    "localhost",
        Port:    783,
        Timeout: 30 * time.Second,
    })
    if err != nil {
        t.Skipf("SpamAssassin not available: %v", err)
    }

    tests := []struct {
        name     string
        email    string
        wantSpam bool
    }{
        {
            name: "obvious spam email",
            email: `From: spam@spammer.com
To: victim@example.com
Subject: FREE MONEY NOW!!!

Click here to claim your $1,000,000 prize!`,
            wantSpam: true,
        },
        {
            name: "legitimate email",
            email: `From: john@example.com
To: jane@example.com
Subject: Meeting Tomorrow

Hi Jane,
Just wanted to confirm our meeting tomorrow at 10am.
Best regards, John`,
            wantSpam: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := client.ScanEmail(tt.email, spamassassin.ScanOptions{
                Verbose: true,
            })
            
            if err != nil {
                t.Fatalf("ScanEmail() error = %v", err)
            }
            
            if result.IsSpam != tt.wantSpam {
                t.Errorf("ScanEmail() IsSpam = %v, want %v (score: %.2f)", 
                    result.IsSpam, tt.wantSpam, result.Score)
            }
        })
    }
}
```

### Container Testing

```go
func TestDockerContainer(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping container test in short mode")
    }

    // Start container
    cmd := exec.Command("docker-compose", "up", "-d")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to start container: %v", err)
    }
    defer func() {
        exec.Command("docker-compose", "down").Run()
    }()

    // Wait for service to be ready
    time.Sleep(10 * time.Second)

    // Test health check
    resp, err := http.Get("http://localhost:8080/health")
    if err != nil {
        t.Fatalf("Health check failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Health check status = %d, want %d", resp.StatusCode, http.StatusOK)
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run only unit tests
go test -short ./...

# Run integration tests
go test -tags=integration ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test
go test -run TestHandler_ScanEmail ./internal/handlers

# Run tests with verbose output
go test -v ./...

# Benchmark tests
go test -bench=. ./...
```

### Test Coverage Requirements

- **Unit Tests**: Minimum 80% coverage
- **Integration Tests**: Cover all major workflows
- **Security Tests**: Cover all input validation and security controls
- **Edge Cases**: Test error conditions and boundary values

## Contributing Workflow

### Getting Started

1. **Fork the Repository**
   ```bash
   # Fork on GitHub, then clone your fork
   git clone https://github.com/your-username/spamassassin-mcp.git
   cd spamassassin-mcp
   git remote add upstream https://github.com/original/spamassassin-mcp.git
   ```

2. **Set Up Development Environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Install development tools
   make install-tools
   
   # Verify setup
   make test
   ```

3. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Development Process

#### 1. Write Code
- Follow coding standards and guidelines
- Add comprehensive tests for new functionality
- Update documentation as needed
- Ensure security best practices

#### 2. Test Locally
```bash
# Run all tests
make test

# Run linting
make lint

# Run security checks
make security-check

# Test in Docker
make docker-test
```

#### 3. Commit Changes
```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat: add email reputation checking

- Add reputation checking against configured blocklists
- Implement domain validation and sanitization
- Add comprehensive test coverage
- Update API documentation

Fixes #123"
```

#### Commit Message Format
We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(handlers): add email reputation checking
fix(security): validate email size before processing
docs(api): update scan_email tool documentation
test(integration): add SpamAssassin connection tests
```

#### 4. Push and Create Pull Request
```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
gh pr create --title "Add email reputation checking" --body "Implements email reputation checking feature..."
```

### Pull Request Guidelines

#### PR Template
```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Security Considerations
- [ ] All inputs are validated
- [ ] No sensitive data is logged
- [ ] Security tests are added/updated
- [ ] Defensive-only operations maintained

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass locally
- [ ] Manual testing completed

## Documentation
- [ ] Code is self-documenting
- [ ] API documentation updated
- [ ] README updated if needed
- [ ] Security documentation updated if needed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] No new security vulnerabilities introduced
- [ ] Performance impact considered
- [ ] Backward compatibility maintained
```

#### Review Process

1. **Automated Checks**
   - All tests must pass
   - Code coverage must meet requirements
   - Linting and security checks must pass
   - Docker build must succeed

2. **Manual Review**
   - Code quality and style
   - Security implications
   - Performance impact
   - Documentation completeness

3. **Security Review**
   - Input validation
   - Error handling
   - Security boundary compliance
   - Defensive-only operation verification

### Code Review Guidelines

#### For Authors
- Keep PRs focused and reasonably sized
- Provide clear description and context
- Respond to feedback constructively
- Update tests and documentation

#### For Reviewers
- Review for security, performance, and maintainability
- Provide constructive feedback with examples
- Ask questions if unclear about intent
- Approve only when confident in changes

## Release Process

### Version Management

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Workflow

#### 1. Prepare Release
```bash
# Update version in relevant files
# Update CHANGELOG.md
# Tag release
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

#### 2. Build and Test
```bash
# Build release artifacts
make build-release

# Run full test suite
make test-all

# Security scan
make security-scan
```

#### 3. Container Images
```bash
# Build production image
docker build -t spamassassin-mcp:v1.2.3 .

# Security scan
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image spamassassin-mcp:v1.2.3

# Tag and push
docker tag spamassassin-mcp:v1.2.3 registry/spamassassin-mcp:v1.2.3
docker push registry/spamassassin-mcp:v1.2.3
```

#### 4. Documentation
- Update deployment documentation
- Create release notes
- Update security documentation if needed

## Debugging and Profiling

### Local Debugging

#### Using Delve Debugger
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug main application
dlv debug main.go

# Debug tests
dlv test ./internal/handlers -- -test.run TestHandler_ScanEmail
```

#### VS Code Debugging
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug MCP Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "env": {
        "SA_MCP_LOG_LEVEL": "debug"
      },
      "args": []
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/handlers",
      "args": [
        "-test.run",
        "TestHandler_ScanEmail"
      ]
    }
  ]
}
```

### Container Debugging

```bash
# Debug running container
docker-compose exec spamassassin-mcp /bin/bash

# Check processes
docker-compose exec spamassassin-mcp ps aux

# Monitor resources
docker stats spamassassin-mcp

# View logs with timestamps
docker-compose logs -f -t spamassassin-mcp
```

### Performance Profiling

#### CPU Profiling
```go
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // ... rest of application
}
```

```bash
# Collect CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze profile
go tool pprof -http=:8080 profile
```

#### Memory Profiling
```bash
# Collect memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Analyze memory usage
go tool pprof -http=:8080 heap
```

#### Trace Analysis
```bash
# Collect execution trace
curl http://localhost:6060/debug/pprof/trace?seconds=5 > trace.out

# Analyze trace
go tool trace trace.out
```

### Build Automation

#### Makefile
```makefile
.PHONY: build test lint security-check docker-build

# Variables
BINARY_NAME=mcp-server
DOCKER_IMAGE=spamassassin-mcp
VERSION?=latest

# Build
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o $(BINARY_NAME) main.go

# Testing
test:
	go test -race -coverprofile=coverage.out ./...

test-integration:
	go test -tags=integration ./...

# Code quality
lint:
	golangci-lint run

security-check:
	gosec ./...
	govulncheck ./...

# Docker
docker-build:
	docker build -t $(DOCKER_IMAGE):$(VERSION) .

docker-test:
	docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/scs/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Cleanup
clean:
	rm -f $(BINARY_NAME)
	docker system prune -f
```

### Continuous Integration

#### GitHub Actions Workflow
```yaml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache dependencies
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: make test
    
    - name: Run linting
      run: make lint
    
    - name: Security check
      run: make security-check
    
    - name: Docker build
      run: make docker-build
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

This completes the comprehensive development guide. The documentation covers all aspects of contributing to the project while maintaining security and quality standards.