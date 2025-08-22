#!/bin/bash
set -euo pipefail

# SpamAssassin MCP Server Health Check

# Configuration
MCP_HOST=${SA_MCP_SERVER_BIND_ADDR:-"0.0.0.0:8080"}
MCP_PORT=$(echo "$MCP_HOST" | cut -d':' -f2)
TIMEOUT=5

# Extract just the port if host is specified
if [[ "$MCP_HOST" == *":"* ]]; then
    MCP_PORT=$(echo "$MCP_HOST" | cut -d':' -f2)
else
    MCP_PORT="8080"
fi

# Health check functions
check_mcp_server() {
    # Check if MCP server port is open using TCP test
    if ! timeout 3 bash -c "echo >/dev/tcp/localhost/$MCP_PORT" 2>/dev/null; then
        echo "ERROR: MCP server not responding on port $MCP_PORT"
        return 1
    fi
    
    # Try to connect via HTTP (basic connectivity test)
    if command -v curl >/dev/null 2>&1; then
        # Test HTTP connection to the root endpoint
        if curl -f -s --max-time "$TIMEOUT" "http://localhost:$MCP_PORT/" >/dev/null 2>&1; then
            return 0
        fi
        # Fallback: just test that the port accepts HTTP requests
        if curl -s --max-time "$TIMEOUT" "http://localhost:$MCP_PORT/mcp" >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    return 0
}

check_spamassassin() {
    # Check if SpamAssassin daemon is running
    if ! pgrep spamd >/dev/null 2>&1; then
        echo "WARNING: SpamAssassin daemon not running"
        return 1
    fi
    
    # Check if spamd port is accessible
    if ! timeout 3 bash -c "echo >/dev/tcp/localhost/783" 2>/dev/null; then
        echo "WARNING: SpamAssassin daemon not responding on port 783"
        return 1
    fi
    
    return 0
}

check_resources() {
    # Check memory usage
    local mem_usage
    mem_usage=$(free | awk '/^Mem:/{printf "%.1f", $3/$2 * 100.0}')
    
    if (( $(echo "$mem_usage > 90.0" | bc -l) )); then
        echo "WARNING: High memory usage: ${mem_usage}%"
    fi
    
    # Check disk space
    local disk_usage
    disk_usage=$(df /var/lib/spamassassin | awk 'NR==2{print $(NF-1)}' | sed 's/%//')
    
    if [ "$disk_usage" -gt 90 ]; then
        echo "WARNING: High disk usage: ${disk_usage}%"
    fi
    
    return 0
}

# Main health check
main() {
    local exit_code=0
    
    echo "Performing health check..."
    
    # Check MCP server
    if ! check_mcp_server; then
        exit_code=1
    else
        echo "✓ MCP server healthy"
    fi
    
    # Check SpamAssassin (non-fatal)
    if ! check_spamassassin; then
        echo "⚠ SpamAssassin issues detected"
        # Don't fail health check for SpamAssassin issues
    else
        echo "✓ SpamAssassin healthy"
    fi
    
    # Check resources (non-fatal)
    check_resources
    
    if [ $exit_code -eq 0 ]; then
        echo "✓ Overall health check passed"
    else
        echo "✗ Health check failed"
    fi
    
    return $exit_code
}

# Install bc if not available (for floating point comparison)
if ! command -v bc >/dev/null 2>&1; then
    # Fallback without bc
    check_resources() {
        return 0
    }
fi

# Run health check
main "$@"