# Task Completion Checklist

## When a development task is completed, ensure the following steps are performed:

### Code Quality
- [ ] **Format Code**: Run `go fmt ./...` to ensure consistent formatting
- [ ] **Vet Code**: Run `go vet ./...` to catch potential issues
- [ ] **Build Verification**: Ensure `go build -o mcp-server main.go` succeeds
- [ ] **Dependency Check**: Run `go mod tidy` to clean up dependencies

### Container Validation
- [ ] **Docker Build**: Verify `docker build -t spamassassin-mcp .` succeeds
- [ ] **Container Start**: Test `docker-compose up -d` works correctly
- [ ] **Health Check**: Verify `/usr/local/bin/health-check.sh` passes
- [ ] **Service Connectivity**: Confirm SpamAssassin daemon connectivity with `nc -z localhost 783`

### Security Validation
- [ ] **Input Validation**: Ensure all user inputs are properly validated
- [ ] **Rate Limiting**: Verify rate limiting is working correctly
- [ ] **Error Handling**: Check that errors don't leak sensitive information
- [ ] **Configuration**: Review that no sensitive data is exposed in logs

### MCP Integration Testing
- [ ] **Tool Registration**: Verify all tools are properly registered
- [ ] **Parameter Validation**: Test tool parameters with invalid inputs
- [ ] **Response Format**: Ensure responses follow MCP specification
- [ ] **Error Responses**: Verify proper error handling and user-friendly messages

### Documentation
- [ ] **Code Comments**: Ensure exported functions have proper Go doc comments
- [ ] **Configuration**: Update config documentation if new options added
- [ ] **API Changes**: Update README.md if tool interfaces changed
- [ ] **Examples**: Provide usage examples for new functionality

### Environment Testing
- [ ] **Local Development**: Test with `./mcp-server` locally
- [ ] **Container Environment**: Test full Docker Compose stack
- [ ] **Resource Usage**: Monitor memory/CPU usage is within limits
- [ ] **Log Output**: Verify appropriate log levels and no sensitive data leakage

### Before Committing
- [ ] **Clean Build**: Ensure fresh build from clean state works
- [ ] **Git Status**: Review all changes with `git status` and `git diff`
- [ ] **Commit Message**: Use descriptive commit message following conventional commits
- [ ] **Security Review**: Double-check no secrets or sensitive data in commit

### Production Readiness (if applicable)
- [ ] **Performance**: Verify response times meet requirements
- [ ] **Monitoring**: Ensure health checks and logging are adequate
- [ ] **Recovery**: Test graceful shutdown and restart procedures
- [ ] **Security Scan**: Run container security scan if available

## Common Commands to Run After Task Completion

```bash
# Code quality checks
go fmt ./...
go vet ./...
go build -o mcp-server main.go

# Dependency management
go mod tidy
go mod verify

# Container validation
docker-compose down
docker-compose up -d --build
docker-compose logs spamassassin-mcp

# Health verification
docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh

# Git workflow
git add .
git status
git commit -m "feat/fix/docs: descriptive message"
```

## Quality Gates
All tasks must pass these quality gates before being considered complete:
1. Code builds without errors or warnings
2. Container starts successfully and passes health checks
3. All MCP tools respond correctly to test inputs
4. No sensitive information is logged or exposed
5. Documentation is updated to reflect changes