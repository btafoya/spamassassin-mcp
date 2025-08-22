# SpamAssassin MCP Server Security Guide

Comprehensive security documentation covering architecture, threat model, and best practices for the SpamAssassin MCP server.

## Table of Contents

- [Security Philosophy](#security-philosophy)
- [Threat Model](#threat-model)
- [Security Architecture](#security-architecture)
- [Access Control](#access-control)
- [Input Validation](#input-validation)
- [Container Security](#container-security)
- [Network Security](#network-security)
- [Data Protection](#data-protection)
- [Monitoring and Incident Response](#monitoring-and-incident-response)
- [Security Checklist](#security-checklist)

## Security Philosophy

### Defensive-Only Design

The SpamAssassin MCP server is built with a **defensive-only security posture**:

- ✅ **Analysis and Detection**: Email spam analysis, reputation checking, rule validation
- ✅ **Educational Tools**: Score explanations, rule testing in safe environments
- ✅ **Configuration Management**: Read-only configuration inspection, defensive rule updates
- ❌ **NO Email Transmission**: Cannot send, relay, or forward emails
- ❌ **NO Malicious Content**: Cannot generate spam, phishing, or malicious content
- ❌ **NO Offensive Operations**: No penetration testing or attack simulation capabilities

### Security Principles

1. **Fail Secure**: System fails to a secure state when errors occur
2. **Least Privilege**: Minimal permissions and access rights
3. **Defense in Depth**: Multiple layers of security controls
4. **Zero Trust**: Verify all inputs and validate all operations
5. **Transparency**: Full audit logging and monitoring
6. **Privacy by Design**: Minimal data collection and processing

## Threat Model

### Assets to Protect

| Asset | Value | Threats |
|-------|--------|---------|
| **Email Content** | High | Data leakage, unauthorized access, content manipulation |
| **System Configuration** | Medium | Unauthorized changes, privilege escalation |
| **Service Availability** | High | DoS attacks, resource exhaustion |
| **Host System** | Critical | Container escape, privilege escalation |
| **Network Communications** | Medium | Interception, man-in-the-middle attacks |

### Attack Vectors

#### 1. Malicious Email Content
**Threat**: Specially crafted emails designed to exploit vulnerabilities
**Mitigations**:
- Strict input validation and size limits (10MB max)
- Email format validation using Go's `mail` package
- Content sanitization before processing
- Timeout protection (60s scan limit)
- Process isolation in containers

#### 2. API Abuse
**Threat**: Excessive requests or malformed API calls
**Mitigations**:
- Rate limiting (60 requests/minute, burst 10)
- Request size validation
- Parameter type checking
- Structured error responses (no data leakage)

#### 3. Container Escape
**Threat**: Breaking out of container to access host system
**Mitigations**:
- Non-root user execution (`spamassassin` user)
- Read-only root filesystem
- No new privileges (`--security-opt no-new-privileges`)
- Resource limits (CPU, memory)
- Capability dropping

#### 4. Network Attacks
**Threat**: Network-based attacks against the service
**Mitigations**:
- Isolated Docker networks
- Minimal exposed ports (only 8080)
- Rate limiting at application and network levels
- Health check monitoring

#### 5. Configuration Tampering
**Threat**: Unauthorized modification of security settings
**Mitigations**:
- Read-only configuration mounts
- Configuration validation on startup
- Immutable container deployment
- Audit logging of configuration access

## Security Architecture

### Multi-Layer Security Model

```
┌─────────────────────────────────────────────┐
│                Load Balancer                │
│         (Rate Limiting, DDoS Protection)   │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│              Network Layer                  │
│    (Firewall, Network Isolation, TLS)      │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           Application Layer                 │
│  (Input Validation, Rate Limiting, Auth)   │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│            Container Layer                  │
│   (Non-root, Read-only, Resource Limits)   │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│              Host Layer                     │
│    (OS Hardening, Monitoring, Logging)     │
└─────────────────────────────────────────────┘
```

### Security Boundaries

1. **Network Boundary**: Firewall, load balancer, network isolation
2. **Application Boundary**: Input validation, rate limiting, authentication
3. **Container Boundary**: Process isolation, resource limits, capability restrictions
4. **System Boundary**: OS-level security, audit logging, monitoring

## Access Control

### Authentication and Authorization

Currently the server operates in a trusted environment without authentication. For production deployment:

#### Recommended Authentication Methods

1. **API Key Authentication**
```go
type APIKeyAuth struct {
    validKeys map[string]bool
}

func (a *APIKeyAuth) Authenticate(req *http.Request) bool {
    apiKey := req.Header.Get("X-API-Key")
    return a.validKeys[apiKey]
}
```

2. **JWT Token Authentication**
```go
func validateJWT(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(secretKey), nil
    })
}
```

3. **mTLS (Mutual TLS)**
```yaml
# nginx configuration for mTLS
server {
    ssl_client_certificate /etc/nginx/client-ca.crt;
    ssl_verify_client on;
    ssl_verify_depth 2;
}
```

### Role-Based Access Control (Future Enhancement)

Planned RBAC implementation:

| Role | Permissions | Use Case |
|------|-------------|----------|
| **Analyzer** | scan_email, explain_score | Email analysis only |
| **Administrator** | All tools + config management | Full system access |
| **Monitor** | get_config, health checks | Monitoring and diagnostics |
| **Developer** | test_rules, scan_email | Rule development and testing |

## Input Validation

### Email Content Validation

```go
func (h *Handler) validateEmailContent(content string) error {
    // Size validation
    if len(content) > int(h.security.MaxEmailSize) {
        return fmt.Errorf("email size exceeds limit of %d bytes", h.security.MaxEmailSize)
    }

    // Content validation
    if content == "" {
        return fmt.Errorf("email content cannot be empty")
    }

    // Format validation using Go's mail package
    if _, err := mail.ReadMessage(strings.NewReader(content)); err != nil {
        return fmt.Errorf("invalid email format: %w", err)
    }

    // Header sanitization
    content = sanitizeHeaders(content)
    
    return nil
}

func sanitizeHeaders(content string) string {
    // Remove potentially dangerous headers
    dangerousHeaders := []string{
        "X-Original-To:",
        "Return-Path:",
        "Envelope-To:",
    }
    
    lines := strings.Split(content, "\n")
    var sanitized []string
    
    for _, line := range lines {
        safe := true
        for _, header := range dangerousHeaders {
            if strings.HasPrefix(strings.ToLower(line), strings.ToLower(header)) {
                safe = false
                break
            }
        }
        if safe {
            sanitized = append(sanitized, line)
        }
    }
    
    return strings.Join(sanitized, "\n")
}
```

### Parameter Validation

```go
var (
    emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    ipRegex    = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
    domainRegex = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func validateEmailAddress(email string) error {
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("invalid email address format")
    }
    return nil
}

func validateIPAddress(ip string) error {
    if !ipRegex.MatchString(ip) {
        return fmt.Errorf("invalid IP address format")
    }
    
    // Additional validation for private/reserved ranges
    if net.ParseIP(ip) == nil {
        return fmt.Errorf("invalid IP address")
    }
    
    return nil
}
```

### Rate Limiting Implementation

```go
type RateLimiter struct {
    limiter *rate.Limiter
    clients map[string]*rate.Limiter
    mutex   sync.RWMutex
}

func NewRateLimiter(requestsPerMinute int, burstSize int) *RateLimiter {
    return &RateLimiter{
        limiter: rate.NewLimiter(
            rate.Every(time.Minute/time.Duration(requestsPerMinute)),
            burstSize,
        ),
        clients: make(map[string]*rate.Limiter),
    }
}

func (rl *RateLimiter) Allow(clientID string) bool {
    rl.mutex.RLock()
    clientLimiter, exists := rl.clients[clientID]
    rl.mutex.RUnlock()
    
    if !exists {
        rl.mutex.Lock()
        clientLimiter = rate.NewLimiter(rl.limiter.Limit(), rl.limiter.Burst())
        rl.clients[clientID] = clientLimiter
        rl.mutex.Unlock()
    }
    
    return clientLimiter.Allow()
}
```

## Container Security

### Dockerfile Security Best Practices

```dockerfile
# Use specific version, not latest
FROM golang:1.21-alpine AS builder

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Build with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o mcp-server main.go

# Production stage
FROM ubuntu:22.04

# Install only necessary packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    spamassassin \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create application user with specific UID/GID
RUN groupadd -r -g 1000 spamassassin && \
    useradd -r -u 1000 -g spamassassin -d /home/spamassassin -s /bin/bash spamassassin

# Set secure file permissions
COPY --from=builder --chown=spamassassin:spamassassin /app/mcp-server /usr/local/bin/mcp-server
RUN chmod 755 /usr/local/bin/mcp-server

# Switch to non-root user
USER spamassassin

# Security labels
LABEL security.policy="defensive-only"
LABEL security.risk="low"
LABEL security.compliance="CIS-Docker-Benchmark"
```

### Runtime Security Configuration

```yaml
# docker-compose.yml security configuration
services:
  spamassassin-mcp:
    security_opt:
      - no-new-privileges:true
      - apparmor:docker-default
      - seccomp:default
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid,size=100m
      - /var/run:noexec,nosuid,size=50m
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - SETGID
      - SETUID
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
          pids: 100
        reservations:
          memory: 256M
          cpus: '0.25'
```

### Container Scanning

Regular security scanning with tools like:

```bash
# Vulnerability scanning
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/src aquasec/trivy image spamassassin-mcp:latest

# Container security benchmark
docker run --rm --net host --pid host --userns host --cap-add audit_control \
  -e DOCKER_CONTENT_TRUST=$DOCKER_CONTENT_TRUST \
  -v /var/lib:/var/lib:ro \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /usr/lib/systemd:/usr/lib/systemd:ro \
  -v /etc:/etc:ro \
  docker/docker-bench-security
```

## Network Security

### Network Isolation

```yaml
networks:
  spamassassin-internal:
    driver: bridge
    internal: true
    ipam:
      config:
        - subnet: 172.20.0.0/24
          
  spamassassin-external:
    driver: bridge
    ipam:
      config:
        - subnet: 172.21.0.0/24
```

### TLS Configuration

Production deployment should use TLS:

```yaml
# nginx TLS configuration
server {
    listen 443 ssl http2;
    ssl_certificate /etc/nginx/certs/server.crt;
    ssl_certificate_key /etc/nginx/certs/server.key;
    
    # Strong SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # HSTS
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    
    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Referrer-Policy "strict-origin-when-cross-origin";
}
```

### Firewall Rules

```bash
# iptables rules for production
iptables -A INPUT -p tcp --dport 8080 -s 10.0.0.0/8 -j ACCEPT
iptables -A INPUT -p tcp --dport 8080 -j DROP
iptables -A INPUT -p tcp --dport 22 -s 192.168.1.0/24 -j ACCEPT
iptables -A INPUT -p tcp --dport 22 -j DROP
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -j DROP
```

## Data Protection

### Data Classification

| Data Type | Classification | Retention | Protection |
|-----------|----------------|-----------|------------|
| **Email Content** | Confidential | Processing only | In-memory only, immediate disposal |
| **Scan Results** | Internal | Session only | Structured logging without content |
| **Configuration** | Internal | Persistent | Encrypted at rest, access controls |
| **Audit Logs** | Internal | 90 days | Encrypted, tamper-evident |
| **Performance Metrics** | Internal | 30 days | Aggregated, anonymized |

### Data Handling Procedures

#### Email Content Processing
```go
type SecureEmailProcessor struct {
    maxSize int64
    timeout time.Duration
}

func (p *SecureEmailProcessor) ProcessEmail(content string) (*ScanResult, error) {
    // Input validation
    if err := p.validateInput(content); err != nil {
        return nil, err
    }
    
    // Process in memory only
    ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
    defer cancel()
    
    result, err := p.scanWithTimeout(ctx, content)
    
    // Explicitly clear sensitive data
    content = ""
    runtime.GC()
    
    return result, err
}

func (p *SecureEmailProcessor) scanWithTimeout(ctx context.Context, content string) (*ScanResult, error) {
    done := make(chan *ScanResult, 1)
    errChan := make(chan error, 1)
    
    go func() {
        result, err := p.performScan(content)
        if err != nil {
            errChan <- err
            return
        }
        done <- result
    }()
    
    select {
    case result := <-done:
        return result, nil
    case err := <-errChan:
        return nil, err
    case <-ctx.Done():
        return nil, fmt.Errorf("scan timeout exceeded")
    }
}
```

#### Secure Logging
```go
type SecureLogger struct {
    logger *logrus.Logger
}

func (l *SecureLogger) LogScanRequest(userID string, emailSize int, scanType string) {
    l.logger.WithFields(logrus.Fields{
        "user_id":    hashUserID(userID),  // Hash for privacy
        "email_size": emailSize,
        "scan_type":  scanType,
        "timestamp":  time.Now().UTC(),
        "session_id": generateSessionID(),
    }).Info("Email scan requested")
}

func hashUserID(userID string) string {
    h := sha256.Sum256([]byte(userID + secretSalt))
    return hex.EncodeToString(h[:8]) // First 8 bytes for logging
}
```

### Encryption at Rest

For sensitive configuration data:

```yaml
# Using Docker secrets
secrets:
  api_keys:
    external: true
  tls_cert:
    external: true
  
services:
  spamassassin-mcp:
    secrets:
      - source: api_keys
        target: /run/secrets/api_keys
        mode: 0400
```

## Monitoring and Incident Response

### Security Monitoring

#### Real-time Monitoring
```go
type SecurityMonitor struct {
    alertThresholds map[string]float64
    alertManager    AlertManager
}

func (m *SecurityMonitor) MonitorRequest(req *Request) {
    // Monitor for suspicious patterns
    if m.detectAnomalousRequest(req) {
        m.alertManager.SendAlert(SecurityAlert{
            Type:      "anomalous_request",
            Severity:  "medium",
            Details:   req.Summary(),
            Timestamp: time.Now(),
        })
    }
    
    // Monitor rate limiting violations
    if req.RateLimited {
        m.alertManager.SendAlert(SecurityAlert{
            Type:      "rate_limit_violation",
            Severity:  "low",
            ClientID:  req.ClientID,
            Timestamp: time.Now(),
        })
    }
}
```

#### Security Metrics
```go
var (
    securityRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "security_requests_total",
            Help: "Total number of security-related requests",
        },
        []string{"type", "status"},
    )
    
    suspiciousActivityTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "suspicious_activity_total",
            Help: "Total number of suspicious activities detected",
        },
        []string{"type", "severity"},
    )
)
```

### Incident Response Plan

#### 1. Detection Phase
- Automated monitoring alerts
- Log analysis and correlation
- User-reported issues
- External threat intelligence

#### 2. Response Phase
```bash
#!/bin/bash
# Emergency response script

INCIDENT_TYPE=$1
SEVERITY=$2

case $SEVERITY in
    "critical")
        # Immediate isolation
        docker-compose down
        iptables -A INPUT -j DROP
        ;;
    "high")
        # Rate limiting increase
        # Enable additional monitoring
        ;;
    "medium")
        # Enhanced logging
        # Notify security team
        ;;
esac

# Log incident
echo "$(date): Incident $INCIDENT_TYPE with severity $SEVERITY" >> /var/log/security-incidents.log
```

#### 3. Recovery Phase
- Service restoration procedures
- Security patch application
- Configuration updates
- Lessons learned documentation

### Audit Logging

```go
type AuditLogger struct {
    logger *logrus.Logger
}

func (a *AuditLogger) LogSecurityEvent(event SecurityEvent) {
    a.logger.WithFields(logrus.Fields{
        "event_type":    event.Type,
        "user_id":      hashString(event.UserID),
        "source_ip":    hashString(event.SourceIP),
        "action":       event.Action,
        "resource":     event.Resource,
        "result":       event.Result,
        "timestamp":    event.Timestamp.UTC(),
        "session_id":   event.SessionID,
        "risk_score":   event.RiskScore,
    }).Warn("Security event detected")
}

type SecurityEvent struct {
    Type      string    `json:"type"`
    UserID    string    `json:"user_id"`
    SourceIP  string    `json:"source_ip"`
    Action    string    `json:"action"`
    Resource  string    `json:"resource"`
    Result    string    `json:"result"`
    Timestamp time.Time `json:"timestamp"`
    SessionID string    `json:"session_id"`
    RiskScore float64   `json:"risk_score"`
}
```

## Security Checklist

### Deployment Security Checklist

#### Pre-Deployment
- [ ] Security code review completed
- [ ] Vulnerability scanning passed
- [ ] Container security benchmarks met
- [ ] Network segmentation configured
- [ ] Firewall rules implemented
- [ ] TLS certificates installed and valid
- [ ] Secrets management configured
- [ ] Monitoring and alerting enabled
- [ ] Backup and recovery procedures tested
- [ ] Incident response plan reviewed

#### Runtime Security
- [ ] Services running as non-root users
- [ ] Resource limits enforced
- [ ] Rate limiting active
- [ ] Input validation functioning
- [ ] Audit logging operational
- [ ] Security monitoring alerts working
- [ ] Regular security updates scheduled
- [ ] Access controls verified

#### Ongoing Security
- [ ] Security patches applied within 24 hours
- [ ] Vulnerability scans run weekly
- [ ] Log analysis performed daily
- [ ] Security metrics reviewed monthly
- [ ] Incident response tested quarterly
- [ ] Security training completed annually
- [ ] Third-party security audit completed annually

### Security Configuration Validation

```bash
#!/bin/bash
# Security validation script

echo "=== SpamAssassin MCP Security Validation ==="

# Check container security
echo "Checking container security..."
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  aquasec/trivy image spamassassin-mcp:latest

# Verify non-root execution
USER_CHECK=$(docker-compose exec spamassassin-mcp whoami)
if [ "$USER_CHECK" != "spamassassin" ]; then
    echo "ERROR: Container not running as spamassassin user"
    exit 1
fi

# Check file permissions
PERM_CHECK=$(docker-compose exec spamassassin-mcp stat -c "%a" /usr/local/bin/mcp-server)
if [ "$PERM_CHECK" != "755" ]; then
    echo "ERROR: Incorrect binary permissions"
    exit 1
fi

# Verify rate limiting
echo "Testing rate limiting..."
for i in {1..70}; do
    curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/health
done | grep "429" > /dev/null
if [ $? -eq 0 ]; then
    echo "✓ Rate limiting working"
else
    echo "⚠ Rate limiting may not be working"
fi

# Check TLS configuration
echo "Checking TLS configuration..."
if command -v testssl.sh >/dev/null 2>&1; then
    testssl.sh --quiet --color 0 https://localhost:8080
fi

echo "✓ Security validation completed"
```

## Compliance Considerations

### GDPR Compliance
- Minimal data collection (no PII storage)
- Data processing transparency
- Right to erasure (automatic data disposal)
- Data protection by design and default

### SOC 2 Type II
- Access controls and authentication
- System availability and performance monitoring
- Security monitoring and incident response
- Change management procedures

### ISO 27001
- Information security management system
- Risk assessment and treatment
- Security controls implementation
- Continuous improvement process

## Security Contact

For security issues or vulnerabilities:
- **Email**: security@company.com
- **PGP Key**: Available at security@company.com
- **Response Time**: 24 hours for critical issues

For non-critical security questions, please refer to this documentation or create an issue in the project repository.