# SpamAssassin MCP Server Deployment Guide

Complete guide for deploying the SpamAssassin MCP server in development, staging, and production environments.

## Table of Contents

- [Quick Start](#quick-start)
- [Production Deployment](#production-deployment)
- [Configuration Management](#configuration-management)
- [Scaling and Performance](#scaling-and-performance)
- [Monitoring and Observability](#monitoring-and-observability)
- [Backup and Recovery](#backup-and-recovery)
- [Maintenance](#maintenance)

## Quick Start

### Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- Minimum 2GB RAM, 1 CPU core
- 10GB available disk space
- Network access for rule updates (ports 80, 443)

### Development Deployment

```bash
# Clone or create project
git clone <repository-url>
cd spamassassin-mcp

# Start development environment
docker-compose up -d

# Verify health
docker-compose ps
docker-compose logs spamassassin-mcp

# Test connection
curl -f http://localhost:8080/health || echo "Use health check script"
docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh
```

### Connect Claude Code

```bash
# Add MCP server configuration
claude --mcp-server spamassassin tcp://localhost:8080

# Test basic functionality
/get_config
/scan_email --content "Subject: Test\n\nTest email content"
```

## Production Deployment

### System Requirements

#### Minimum Requirements
- **CPU**: 2 cores
- **RAM**: 4GB
- **Disk**: 20GB SSD
- **Network**: 1Gbps
- **OS**: Ubuntu 20.04+ / RHEL 8+ / Amazon Linux 2

#### Recommended Production
- **CPU**: 4+ cores
- **RAM**: 8GB+
- **Disk**: 50GB+ SSD with backup
- **Network**: Redundant network interfaces
- **OS**: Latest stable Linux distribution

### Production Docker Compose

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  spamassassin-mcp:
    build: 
      context: .
      dockerfile: Dockerfile
    container_name: spamassassin-mcp-prod
    restart: always
    ports:
      - "8080:8080"
    volumes:
      # Production configuration
      - ./configs/prod:/etc/spamassassin-mcp:ro
      # Persistent data
      - spamassassin-rules:/var/lib/spamassassin
      - spamassassin-logs:/var/log/spamassassin
      # Backup mount
      - /backup/spamassassin:/backup:rw
    environment:
      # Production environment variables
      - SA_MCP_LOG_LEVEL=warn
      - SA_MCP_SERVER_BIND_ADDR=0.0.0.0:8080
      - SA_MCP_SECURITY_RATE_LIMITING_REQUESTS_PER_MINUTE=120
      - SA_MCP_SECURITY_VALIDATION_ENABLED=true
    networks:
      - spamassassin-prod
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid,size=200m
      - /var/run:noexec,nosuid,size=100m
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '2.0'
        reservations:
          memory: 1G
          cpus: '1.0'
      restart_policy:
        condition: any
        delay: 10s
        max_attempts: 5
        window: 120s
    healthcheck:
      test: ["/usr/local/bin/health-check.sh"]
      interval: 30s
      timeout: 10s
      start_period: 60s
      retries: 3
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "5"

  # Production monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    restart: always
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    networks:
      - spamassassin-prod
    profiles:
      - monitoring

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    restart: always
    ports:
      - "3000:3000"
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana:/etc/grafana/provisioning
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=secure_password_here
    networks:
      - spamassassin-prod
    profiles:
      - monitoring

volumes:
  spamassassin-rules:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/spamassassin/rules
  spamassassin-logs:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /data/spamassassin/logs
  prometheus-data:
    driver: local
  grafana-data:
    driver: local

networks:
  spamassassin-prod:
    driver: bridge
    ipam:
      config:
        - subnet: 172.21.0.0/16
```

### Production Startup

```bash
# Create production directories
sudo mkdir -p /data/spamassassin/{rules,logs}
sudo mkdir -p /backup/spamassassin
sudo chown -R 1000:1000 /data/spamassassin

# Deploy production stack
docker-compose -f docker-compose.prod.yml up -d

# Verify deployment
docker-compose -f docker-compose.prod.yml ps
docker-compose -f docker-compose.prod.yml logs spamassassin-mcp-prod
```

## Configuration Management

### Environment-Specific Configuration

#### Development (`configs/config.yaml`)
```yaml
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "30s"

security:
  rate_limiting:
    requests_per_minute: 60
    burst_size: 10
  validation_enabled: true

log_level: "debug"
```

#### Production (`configs/prod/config.yaml`)
```yaml
server:
  bind_addr: "0.0.0.0:8080"
  timeout: "60s"

spamassassin:
  host: "localhost"
  port: 783
  timeout: "45s"
  threshold: 5.0

security:
  max_email_size: 10485760
  rate_limiting:
    requests_per_minute: 120
    burst_size: 20
  scan_timeout: "90s"
  validation_enabled: true
  
  # Production security lists
  allowed_senders:
    - "notifications@company.com"
    - "alerts@monitoring.com"
    
  blocked_domains:
    - "known-spam-domain.com"
    - "malicious-site.net"

log_level: "warn"
```

### Secrets Management

For production, use Docker secrets or external secret management:

```yaml
services:
  spamassassin-mcp:
    secrets:
      - mcp_config
      - ssl_cert
      - ssl_key
    environment:
      - SA_MCP_CONFIG_FILE=/run/secrets/mcp_config

secrets:
  mcp_config:
    external: true
  ssl_cert:
    external: true
  ssl_key:
    external: true
```

## Scaling and Performance

### Horizontal Scaling

#### Load Balancer Configuration (nginx)

```nginx
upstream spamassassin_backend {
    server spamassassin-mcp-1:8080 max_fails=3 fail_timeout=30s;
    server spamassassin-mcp-2:8080 max_fails=3 fail_timeout=30s;
    server spamassassin-mcp-3:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 80;
    server_name spamassassin-mcp.company.com;
    
    location / {
        proxy_pass http://spamassassin_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # Timeouts
        proxy_connect_timeout 5s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Rate limiting
        limit_req zone=api burst=20 nodelay;
        limit_req_status 429;
    }
    
    location /health {
        proxy_pass http://spamassassin_backend;
        access_log off;
    }
}

# Rate limiting zone
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=60r/m;
}
```

#### Multi-Instance Docker Compose

```yaml
services:
  spamassassin-mcp-1:
    extends:
      file: docker-compose.prod.yml
      service: spamassassin-mcp
    container_name: spamassassin-mcp-1
    
  spamassassin-mcp-2:
    extends:
      file: docker-compose.prod.yml
      service: spamassassin-mcp
    container_name: spamassassin-mcp-2
    
  spamassassin-mcp-3:
    extends:
      file: docker-compose.prod.yml
      service: spamassassin-mcp
    container_name: spamassassin-mcp-3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/certs:/etc/nginx/certs:ro
    depends_on:
      - spamassassin-mcp-1
      - spamassassin-mcp-2
      - spamassassin-mcp-3
```

### Performance Tuning

#### Container Resource Limits
```yaml
deploy:
  resources:
    limits:
      memory: 2G
      cpus: '2.0'
    reservations:
      memory: 1G
      cpus: '1.0'
```

#### SpamAssassin Optimization
```bash
# Increase SpamAssassin child processes
spamd --max-children 10 --max-spare 5 --min-spare 2

# Optimize memory usage
echo "bayes_auto_expire 1" >> /etc/spamassassin/local.cf
echo "bayes_journal_max_size 102400" >> /etc/spamassassin/local.cf
```

## Monitoring and Observability

### Health Checks

#### Application Health Check
```bash
#!/bin/bash
# /usr/local/bin/health-check.sh (enhanced)

check_application() {
    local response
    response=$(curl -f -s --max-time 5 http://localhost:8080/health 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo "✓ Application healthy: $response"
        return 0
    else
        echo "✗ Application health check failed"
        return 1
    fi
}

check_spamassassin() {
    if pgrep spamd > /dev/null; then
        echo "✓ SpamAssassin daemon running"
        return 0
    else
        echo "✗ SpamAssassin daemon not running"
        return 1
    fi
}

check_performance() {
    local cpu_usage memory_usage
    cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
    memory_usage=$(free | awk '/^Mem:/{printf "%.1f", $3/$2 * 100.0}')
    
    echo "📊 CPU: ${cpu_usage}%, Memory: ${memory_usage}%"
    
    if (( $(echo "$memory_usage > 90.0" | bc -l) )); then
        echo "⚠️ High memory usage detected"
        return 1
    fi
    
    return 0
}

main() {
    echo "=== SpamAssassin MCP Health Check ==="
    local exit_code=0
    
    check_application || exit_code=1
    check_spamassassin || exit_code=1
    check_performance || exit_code=1
    
    if [ $exit_code -eq 0 ]; then
        echo "✅ All checks passed"
    else
        echo "❌ Health check failed"
    fi
    
    return $exit_code
}

main "$@"
```

### Logging Configuration

#### Structured Logging Setup
```yaml
# docker-compose.yml
services:
  spamassassin-mcp:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "5"
        labels: "service,environment"
    labels:
      - "service=spamassassin-mcp"
      - "environment=production"
```

#### Log Aggregation with ELK Stack
```yaml
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
    volumes:
      - elasticsearch-data:/usr/share/elasticsearch/data
    
  logstash:
    image: docker.elastic.co/logstash/logstash:8.11.0
    volumes:
      - ./logstash/pipeline:/usr/share/logstash/pipeline:ro
    depends_on:
      - elasticsearch
    
  kibana:
    image: docker.elastic.co/kibana/kibana:8.11.0
    ports:
      - "5601:5601"
    environment:
      - ELASTICSEARCH_HOSTS=http://elasticsearch:9200
    depends_on:
      - elasticsearch
```

### Metrics and Monitoring

#### Prometheus Configuration (`monitoring/prometheus.yml`)
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "spamassassin_rules.yml"

scrape_configs:
  - job_name: 'spamassassin-mcp'
    static_configs:
      - targets: ['spamassassin-mcp:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'docker-containers'
    static_configs:
      - targets: ['docker-exporter:9323']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

#### Alert Rules (`monitoring/spamassassin_rules.yml`)
```yaml
groups:
  - name: spamassassin_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(spamassassin_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in SpamAssassin MCP server"
          
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(spamassassin_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High response time in SpamAssassin MCP server"
          
      - alert: ServiceDown
        expr: up{job="spamassassin-mcp"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SpamAssassin MCP server is down"
```

## Backup and Recovery

### Automated Backup Script

```bash
#!/bin/bash
# /usr/local/bin/backup-spamassassin.sh

BACKUP_DIR="/backup/spamassassin"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup SpamAssassin rules and configuration
echo "Starting backup at $(date)"

# Backup rules
docker-compose exec spamassassin-mcp tar czf /backup/rules_$DATE.tar.gz /var/lib/spamassassin

# Backup configuration
cp -r configs "$BACKUP_DIR/config_$DATE"

# Backup logs (last 7 days)
docker-compose exec spamassassin-mcp find /var/log/spamassassin -name "*.log" -mtime -7 -exec tar czf /backup/logs_$DATE.tar.gz {} +

# Cleanup old backups
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "config_*" -mtime +$RETENTION_DAYS -exec rm -rf {} +

echo "Backup completed at $(date)"
```

### Recovery Procedures

#### Complete System Recovery
```bash
# Stop services
docker-compose -f docker-compose.prod.yml down

# Restore from backup
cd /backup/spamassassin
tar xzf rules_YYYYMMDD_HHMMSS.tar.gz -C /data/spamassassin/
cp -r config_YYYYMMDD_HHMMSS/* configs/

# Restart services
docker-compose -f docker-compose.prod.yml up -d

# Verify recovery
docker-compose -f docker-compose.prod.yml exec spamassassin-mcp /usr/local/bin/health-check.sh
```

## Maintenance

### Regular Maintenance Tasks

#### Weekly Tasks
```bash
# Update SpamAssassin rules
docker-compose exec spamassassin-mcp sa-update --nogpg

# Restart services for rule reload
docker-compose restart spamassassin-mcp

# Clean up old logs
docker-compose exec spamassassin-mcp find /var/log/spamassassin -name "*.log" -mtime +7 -delete

# Check disk usage
df -h /data/spamassassin
```

#### Monthly Tasks
```bash
# Update container images
docker-compose pull
docker-compose up -d

# Optimize Bayes database
docker-compose exec spamassassin-mcp sa-learn --dump magic

# Review and rotate logs
docker system prune -f

# Security audit
docker-compose exec spamassassin-mcp find / -perm -4000 -ls
```

### Automated Maintenance with Cron

```bash
# /etc/cron.d/spamassassin-mcp

# Update rules weekly (Sunday 2 AM)
0 2 * * 0 root /usr/local/bin/backup-spamassassin.sh && docker-compose -f /opt/spamassassin-mcp/docker-compose.prod.yml exec spamassassin-mcp sa-update --nogpg

# Health check every 5 minutes
*/5 * * * * root docker-compose -f /opt/spamassassin-mcp/docker-compose.prod.yml exec spamassassin-mcp /usr/local/bin/health-check.sh || /usr/local/bin/alert-ops.sh

# Daily log rotation
0 0 * * * root docker-compose -f /opt/spamassassin-mcp/docker-compose.prod.yml exec spamassassin-mcp logrotate /etc/logrotate.conf
```

## Security Hardening

### Container Security
- Run as non-root user
- Read-only root filesystem
- No new privileges
- Security-optimized base images
- Regular security updates

### Network Security
- Isolated Docker networks
- Firewall rules limiting access
- TLS encryption for external communications
- Rate limiting and DDoS protection

### Operational Security
- Regular security audits
- Vulnerability scanning
- Access logging and monitoring
- Incident response procedures

## Troubleshooting Deployment Issues

### Common Issues

1. **Container Won't Start**
   ```bash
   # Check logs
   docker-compose logs spamassassin-mcp
   
   # Check resource usage
   docker stats
   
   # Verify configuration
   docker-compose config
   ```

2. **High Memory Usage**
   ```bash
   # Analyze memory usage
   docker-compose exec spamassassin-mcp ps aux --sort=-%mem
   
   # Restart with memory limits
   docker-compose down && docker-compose up -d
   ```

3. **SpamAssassin Daemon Issues**
   ```bash
   # Check spamd status
   docker-compose exec spamassassin-mcp pgrep spamd
   
   # Restart spamd
   docker-compose exec spamassassin-mcp pkill spamd
   docker-compose restart spamassassin-mcp
   ```

For more troubleshooting guidance, see [TROUBLESHOOTING.md](TROUBLESHOOTING.md).