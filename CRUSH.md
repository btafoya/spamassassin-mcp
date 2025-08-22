# CRUSH.md - Development Guidelines for SpamAssassin MCP

## Build Commands
```bash
# Build the project
go build -o mcp-server main.go

# Format code
go fmt ./...

# Vet code
go vet ./...

# Build with container
docker compose up -d --build
```

## Test Commands
```bash
# Run all tests (when implemented)
go test ./...

# Run a specific test file (when implemented)
go test internal/handlers/handlers_test.go

# Run tests with verbose output (when implemented)
go test -v ./...
```

## Lint Commands
```bash
# Code validation
go vet ./...

# Format check (see if files are formatted)
go fmt -l ./...
```

## Code Style Guidelines

### General Principles
- Follow standard Go conventions from `gofmt`, `go vet`, and Go Code Review Comments
- Use clear, descriptive names for variables, functions, and types
- Keep functions small and focused on a single responsibility

### Naming Conventions
- **Variables**: camelCase (e.g., `emailContent`, `maxRetries`)
- **Constants**: CamelCase for exported, camelCase for unexported
- **Functions**: CamelCase for exported, camelCase for unexported
- **Structs**: CamelCase for exported types
- **Interfaces**: Often end with "er" suffix (e.g., `Handler`, `Scanner`)

### Imports Organization
1. Standard library packages
2. Third-party packages
3. Local packages

### Error Handling
- Always handle errors explicitly
- Use descriptive error messages with context
- Wrap errors with additional context when appropriate: `fmt.Errorf("operation failed: %w", err)`

### Documentation
- Use standard Go doc comments for exported functions and types
- Start comments with the name of the item being documented