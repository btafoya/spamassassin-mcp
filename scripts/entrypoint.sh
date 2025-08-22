#!/bin/bash
set -euo pipefail

# SpamAssassin MCP Server Entrypoint

# Logging function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" >&2
}

log "Starting SpamAssassin MCP Server entrypoint"

# Initialize SpamAssassin if needed
init_spamassassin() {
    log "Initializing SpamAssassin..."
    
    # Create SpamAssassin directories if they don't exist
    mkdir -p /var/lib/spamassassin/.spamassassin
    
    # Start spamd if not running
    if ! pgrep spamd > /dev/null; then
        log "Starting SpamAssassin daemon..."
        spamd \
            --create-prefs \
            --max-children 5 \
            --helper-home-dir /var/lib/spamassassin \
            --username spamassassin \
            --groupname spamassassin \
            --pidfile /var/run/spamd.pid \
            --daemonize \
            --allowed-ips 127.0.0.1,::1 \
            --listen-ip 127.0.0.1 \
            --port 783
            
        # Wait for spamd to be ready
        log "Waiting for SpamAssassin daemon to be ready..."
        for i in {1..30}; do
            if timeout 2 bash -c 'echo >/dev/tcp/127.0.0.1/783' 2>/dev/null; then
                log "SpamAssassin daemon is ready"
                break
            fi
            if [ $i -eq 30 ]; then
                log "ERROR: SpamAssassin daemon failed to start within 30 seconds"
                exit 1
            fi
            sleep 1
        done
    else
        log "SpamAssassin daemon already running"
    fi
}

# Update SpamAssassin rules
update_rules() {
    log "Updating SpamAssassin rules..."
    if command -v sa-update >/dev/null 2>&1; then
        sa-update --nogpg || log "Warning: Rule update failed (this is normal for first run)"
    else
        log "Warning: sa-update not available"
    fi
}

# Handle shutdown signals
cleanup() {
    log "Received shutdown signal, cleaning up..."
    if [ -f /var/run/spamd.pid ]; then
        local pid=$(cat /var/run/spamd.pid)
        if kill -0 "$pid" 2>/dev/null; then
            log "Stopping SpamAssassin daemon (PID: $pid)"
            kill -TERM "$pid"
            wait "$pid" 2>/dev/null || true
        fi
    fi
    exit 0
}

# Set up signal handlers
trap cleanup SIGTERM SIGINT

# Check if we're running as the correct user
if [ "$(id -u)" != "$(id -u spamassassin)" ]; then
    log "ERROR: Not running as spamassassin user"
    exit 1
fi

# Initialize SpamAssassin
init_spamassassin

# Update rules if requested
if [ "${UPDATE_RULES:-false}" = "true" ]; then
    update_rules
fi

# Validate configuration
log "Validating configuration..."
if [ ! -x /usr/local/bin/mcp-server ]; then
    log "ERROR: MCP server binary not found or not executable"
    exit 1
fi

# Start the MCP server
log "Starting MCP server with command: $*"
exec "$@"