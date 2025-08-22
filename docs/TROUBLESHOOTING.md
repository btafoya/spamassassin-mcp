# SpamAssassin MCP Server Troubleshooting Guide

Comprehensive troubleshooting guide for common issues, debugging techniques, and problem resolution.

## Table of Contents

- [Quick Diagnostic Commands](#quick-diagnostic-commands)
- [Common Issues](#common-issues)
- [Installation Problems](#installation-problems)
- [Runtime Issues](#runtime-issues)
- [Performance Problems](#performance-problems)
- [Security Issues](#security-issues)
- [Integration Problems](#integration-problems)
- [Debugging Techniques](#debugging-techniques)
- [Log Analysis](#log-analysis)
- [FAQ](#faq)

## Quick Diagnostic Commands

### Health Check Sequence
```bash
# 1. Check container status
docker-compose ps

# 2. Run health check
docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh

# 3. Check logs
docker-compose logs --tail=50 spamassassin-mcp

# 4. Test connectivity
curl -f http://localhost:8080/health || echo "Connection failed"

# 5. Check SpamAssassin daemon
docker-compose exec spamassassin-mcp pgrep spamd
```

### System Information
```bash
# Container resource usage
docker stats spamassassin-mcp

# Disk usage
docker-compose exec spamassassin-mcp df -h

# Memory usage
docker-compose exec spamassassin-mcp free -h

# Process list
docker-compose exec spamassassin-mcp ps aux
```

## Common Issues

### 1. Container Won't Start

#### Symptoms
- Container exits immediately
- "Container spamassassin-mcp is unhealthy"
- Docker compose fails to start services

#### Diagnosis
```bash
# Check container logs
docker-compose logs spamassassin-mcp

# Check Docker daemon
systemctl status docker

# Verify image integrity
docker images | grep spamassassin-mcp

# Check disk space
df -h
```

#### Solutions

**Insufficient Resources**
```bash
# Check available resources
docker system df
docker system prune -f

# Increase memory limits in docker-compose.yml
deploy:
  resources:
    limits:
      memory: 1G  # Increase from 512M
```

**Permission Issues**
```bash
# Fix file permissions
sudo chown -R 1000:1000 /data/spamassassin
sudo chmod -R 755 scripts/
```

**Configuration Errors**
```bash
# Validate configuration
docker-compose config

# Check for syntax errors
yamllint docker-compose.yml
```

### 2. SpamAssassin Daemon Not Starting

#### Symptoms
- "SpamAssassin daemon not running" in health check
- Connection refused on port 783
- Scan operations fail with connection errors

#### Diagnosis
```bash
# Check spamd process
docker-compose exec spamassassin-mcp pgrep -f spamd

# Check port binding
docker-compose exec spamassassin-mcp netstat -ln | grep 783

# Test spamd directly
docker-compose exec spamassassin-mcp echo "test" | spamc
```

#### Solutions

**Manual Start**
```bash
# Enter container
docker-compose exec spamassassin-mcp /bin/bash

# Start spamd manually
spamd --create-prefs --max-children 5 --listen-ip 127.0.0.1 --port 783 --pidfile /var/run/spamd.pid

# Check status
pgrep spamd && echo "spamd running"
```

**Configuration Fix**
```bash
# Check SpamAssassin configuration
docker-compose exec spamassassin-mcp spamassassin --lint

# Update rules
docker-compose exec spamassassin-mcp sa-update --nogpg

# Restart container
docker-compose restart spamassassin-mcp
```

### 3. High Memory Usage

#### Symptoms
- Container killed by OOM killer
- Slow response times
- Memory usage > 90%

#### Diagnosis
```bash
# Monitor memory usage
docker stats --no-stream spamassassin-mcp

# Check memory breakdown
docker-compose exec spamassassin-mcp cat /proc/meminfo

# Analyze process memory
docker-compose exec spamassassin-mcp ps aux --sort=-%mem
```

#### Solutions

**Optimize SpamAssassin**
```bash
# Reduce child processes
# In entrypoint.sh, modify spamd command:
spamd --max-children 3 --max-spare 2 --min-spare 1

# Clear Bayes database if too large
docker-compose exec spamassassin-mcp sa-learn --clear
```

**Container Limits**
```yaml
# Increase memory limits
deploy:
  resources:
    limits:
      memory: 2G
    reservations:
      memory: 1G
```

### 4. Rate Limiting Issues

#### Symptoms
- "429 Too Many Requests" errors
- Clients being blocked unexpectedly
- Legitimate requests failing

#### Diagnosis
```bash
# Check rate limit configuration
docker-compose exec spamassassin-mcp env | grep RATE

# Monitor request patterns
docker-compose logs spamassassin-mcp | grep "rate limit"

# Test rate limiting
for i in {1..70}; do curl -s http://localhost:8080/health; done
```

#### Solutions

**Adjust Rate Limits**
```yaml
# In docker-compose.yml
environment:
  - SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=120
  - SA_MCP_SECURITY_RATE_LIMITING_BURST_SIZE=20
```

**Per-Client Rate Limiting**
```go
// Implement IP-based rate limiting in handlers
func (h *Handler) getRateLimiter(clientIP string) *rate.Limiter {
    h.mutex.RLock()
    limiter, exists := h.clients[clientIP]
    h.mutex.RUnlock()
    
    if !exists {
        h.mutex.Lock()
        limiter = rate.NewLimiter(h.rate, h.burst)
        h.clients[clientIP] = limiter
        h.mutex.Unlock()
    }
    
    return limiter
}
```

### 5. Connection Timeout Issues

#### Symptoms
- "Connection timeout" errors
- Long response times
- Intermittent failures

#### Diagnosis
```bash
# Check network connectivity
docker-compose exec spamassassin-mcp ping -c 3 localhost

# Test port connectivity
docker-compose exec spamassassin-mcp nc -z localhost 783

# Monitor connection times
time curl http://localhost:8080/health
```

#### Solutions

**Increase Timeouts**
```yaml
# In configuration
spamassassin:
  timeout: "60s"  # Increase from 30s
  
security:
  scan_timeout: "120s"  # Increase from 60s
```

**Connection Pooling**
```go
// Implement connection pooling in SpamAssassin client
type ClientPool struct {
    pool    chan net.Conn
    factory func() (net.Conn, error)
    timeout time.Duration
}

func (p *ClientPool) Get() (net.Conn, error) {
    select {
    case conn := <-p.pool:
        return conn, nil
    default:
        return p.factory()
    }
}
```

## Installation Problems

### Docker Installation Issues

#### Docker Not Starting
```bash
# Check Docker daemon status
sudo systemctl status docker

# Restart Docker daemon
sudo systemctl restart docker

# Check Docker logs
sudo journalctl -u docker.service
```

#### Docker Compose Version
```bash
# Check version (need 2.0+)
docker-compose --version

# Upgrade Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

### Build Issues

#### Build Context Too Large
```bash
# Add .dockerignore file
echo "node_modules\n.git\n*.log" > .dockerignore

# Clean build context
docker system prune -f
```

#### Go Module Issues
```bash
# Clear module cache
go clean -modcache

# Verify modules
go mod verify

# Download dependencies
go mod download
```

## Runtime Issues

### Service Discovery Problems

#### Claude Code Connection Issues
```bash
# Verify MCP server is listening
netstat -ln | grep 8080

# Test MCP protocol
echo '{"method":"tools/list","params":{}}' | nc localhost 8080

# Check Claude Code configuration
claude --config-path
```

#### Network Isolation
```bash
# Check Docker networks
docker network ls

# Inspect network configuration
docker network inspect spamassassin-mcp_default

# Test inter-container communication
docker-compose exec spamassassin-mcp ping spamd
```

### Configuration Issues

#### Environment Variables Not Loading
```bash
# Check environment in container
docker-compose exec spamassassin-mcp env | grep SA_MCP

# Validate configuration
docker-compose exec spamassassin-mcp cat /etc/spamassassin-mcp/config.yaml

# Test configuration parsing
docker-compose exec spamassassin-mcp mcp-server --validate-config
```

#### File Permission Problems
```bash
# Check file ownership
docker-compose exec spamassassin-mcp ls -la /etc/spamassassin-mcp/

# Fix permissions
docker-compose exec spamassassin-mcp chown -R spamassassin:spamassassin /var/lib/spamassassin
```

## Performance Problems

### Slow Email Scanning

#### Symptoms
- Scan times > 10 seconds
- Timeout errors
- High CPU usage during scans

#### Solutions

**SpamAssassin Optimization**
```bash
# Disable slow tests
echo "score DNS_FROM_OPENWHOIS 0" >> /etc/spamassassin/local.cf
echo "score DNS_FROM_RFC_IGNORANT 0" >> /etc/spamassassin/local.cf

# Optimize Bayes
echo "bayes_auto_expire 1" >> /etc/spamassassin/local.cf
echo "bayes_journal_max_size 102400" >> /etc/spamassassin/local.cf
```

**Parallel Processing**
```go
// Implement concurrent scanning
func (h *Handler) ScanEmailsBatch(emails []string) ([]*ScanResult, error) {
    results := make([]*ScanResult, len(emails))
    var wg sync.WaitGroup
    
    for i, email := range emails {
        wg.Add(1)
        go func(idx int, content string) {
            defer wg.Done()
            result, err := h.saClient.ScanEmail(content, options)
            if err == nil {
                results[idx] = result
            }
        }(i, email)
    }
    
    wg.Wait()
    return results, nil
}
```

### Memory Leaks

#### Detection
```bash
# Monitor memory over time
while true; do
    docker stats --no-stream spamassassin-mcp | grep memory
    sleep 30
done

# Profile memory usage
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

#### Solutions
```go
// Explicit garbage collection after large operations
func (h *Handler) ScanEmail(ctx context.Context, params json.RawMessage) (any, error) {
    defer runtime.GC()
    
    // ... processing
    
    // Clear large variables
    params = nil
    
    return result, nil
}
```

## Security Issues

### Input Validation Failures

#### Malformed Email Content
```bash
# Test with various malformed inputs
curl -X POST http://localhost:8080/scan_email \
  -H "Content-Type: application/json" \
  -d '{"content":"invalid\x00\x01email"}'
```

#### Large Email Attacks
```bash
# Test size limits
dd if=/dev/zero bs=1M count=15 | base64 > large_email.txt
curl -X POST http://localhost:8080/scan_email \
  -H "Content-Type: application/json" \
  -d "{\"content\":\"$(cat large_email.txt)\"}"
```

### Container Security

#### Privilege Escalation Check
```bash
# Verify non-root execution
docker-compose exec spamassassin-mcp whoami

# Check capabilities
docker-compose exec spamassassin-mcp cat /proc/self/status | grep Cap

# Verify read-only filesystem
docker-compose exec spamassassin-mcp touch /test_write || echo "Read-only OK"
```

## Integration Problems

### Claude Code Integration

#### Connection Issues
```bash
# Test MCP protocol manually
cat << EOF | nc localhost 8080
{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}
EOF
```

#### Tool Registration Problems
```bash
# Check available tools
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}'
```

### Load Balancer Integration

#### Health Check Configuration
```nginx
# nginx health check configuration
location /health {
    proxy_pass http://spamassassin_backend/health;
    proxy_connect_timeout 5s;
    proxy_send_timeout 10s;
    proxy_read_timeout 10s;
    access_log off;
}
```

## Debugging Techniques

### Enable Debug Logging

#### Application Debug Mode
```yaml
# In docker-compose.yml
environment:
  - SA_MCP_LOG_LEVEL=debug
```

#### SpamAssassin Debug Mode
```bash
# Enable SpamAssassin debugging
docker-compose exec spamassassin-mcp spamd -D --max-children 1
```

### Network Debugging

#### Packet Capture
```bash
# Capture network traffic
docker-compose exec spamassassin-mcp tcpdump -i any -w /tmp/capture.pcap port 783

# Analyze captured packets
docker-compose exec spamassassin-mcp tcpdump -r /tmp/capture.pcap -A
```

#### Connection Tracing
```bash
# Trace system calls
docker-compose exec spamassassin-mcp strace -p $(pgrep mcp-server) -e trace=network
```

### Performance Profiling

#### CPU Profiling
```bash
# Enable pprof endpoint
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof
```

#### Memory Profiling
```bash
# Memory heap analysis
curl http://localhost:6060/debug/pprof/heap > mem.prof
go tool pprof mem.prof
```

## Log Analysis

### Log Locations
```bash
# Application logs
docker-compose logs spamassassin-mcp

# SpamAssassin logs
docker-compose exec spamassassin-mcp ls -la /var/log/spamassassin/

# System logs
docker-compose exec spamassassin-mcp journalctl -f
```

### Log Analysis Tools

#### Error Pattern Detection
```bash
# Find error patterns
docker-compose logs spamassassin-mcp | grep -E "(ERROR|FATAL|panic)"

# Count error types
docker-compose logs spamassassin-mcp | grep ERROR | cut -d' ' -f5- | sort | uniq -c

# Timeline analysis
docker-compose logs -t spamassassin-mcp | grep ERROR
```

#### Performance Analysis
```bash
# Response time analysis
docker-compose logs spamassassin-mcp | grep "scan completed" | \
  awk '{print $NF}' | sed 's/ms//' | sort -n

# Request rate analysis
docker-compose logs spamassassin-mcp | grep "scan request" | \
  cut -d' ' -f1-2 | uniq -c
```

### Structured Log Parsing

#### JSON Log Processing
```bash
# Extract specific fields from JSON logs
docker-compose logs spamassassin-mcp | \
  jq -r 'select(.level=="ERROR") | "\(.time) \(.msg)"'

# Aggregate error types
docker-compose logs spamassassin-mcp | \
  jq -r 'select(.level=="ERROR") | .error_type' | sort | uniq -c
```

## FAQ

### General Questions

**Q: Why is the server responding slowly?**
A: Check SpamAssassin daemon status, memory usage, and network connectivity. Large emails or complex rule sets can increase processing time.

**Q: How do I update SpamAssassin rules?**
A: Use the `update_rules` MCP tool or run `sa-update --nogpg` in the container.

**Q: Can I customize the spam threshold?**
A: Yes, set `SA_MCP_SPAMASSASSIN_THRESHOLD` environment variable or update the configuration file.

### Security Questions

**Q: Is email content stored permanently?**
A: No, all email content is processed in-memory only and immediately discarded after analysis.

**Q: How do I enable authentication?**
A: Currently not implemented. For production, implement API key authentication or place behind an authenticated proxy.

**Q: What data is logged?**
A: Only metadata (request times, scores, rule counts) is logged. Email content is never logged.

### Performance Questions

**Q: How many emails can be processed per minute?**
A: Approximately 300-400 emails/minute depending on email size and complexity.

**Q: How do I scale for higher throughput?**
A: Deploy multiple instances behind a load balancer or increase SpamAssassin child processes.

**Q: Why do some scans timeout?**
A: Large emails or network issues can cause timeouts. Increase timeout values or check network connectivity.

### Configuration Questions

**Q: Where are configuration files located?**
A: Default configuration is in `/etc/spamassassin-mcp/config.yaml` in the container.

**Q: How do I override default settings?**
A: Use environment variables with `SA_MCP_` prefix or mount custom configuration files.

**Q: Can I use custom SpamAssassin rules?**
A: Yes, mount custom rule files or use the `test_rules` tool to validate custom rules.

### Integration Questions

**Q: How do I connect Claude Code to the server?**
A: Use `claude --mcp-server spamassassin tcp://localhost:8080`

**Q: Can I use this with other MCP clients?**
A: Yes, the server implements the standard MCP protocol and works with any MCP-compatible client.

**Q: How do I monitor the server health?**
A: Use the built-in health check script or monitor the `/health` endpoint.

## Emergency Procedures

### Server Recovery
```bash
# Emergency stop
docker-compose down

# Clean restart
docker-compose down --volumes
docker-compose up -d

# Factory reset
docker-compose down --volumes --remove-orphans
docker system prune -f
docker-compose up -d
```

### Data Recovery
```bash
# Restore from backup
docker-compose down
tar xzf /backup/spamassassin/rules_backup.tar.gz -C /data/spamassassin/
docker-compose up -d
```

### Security Incident Response
```bash
# Immediate isolation
iptables -A INPUT -p tcp --dport 8080 -j DROP

# Stop services
docker-compose down

# Analyze logs
docker-compose logs spamassassin-mcp > incident_logs.txt

# Preserve evidence
cp -r configs/ incident_config_backup/
```

For additional support, please refer to the project documentation or create an issue in the repository.