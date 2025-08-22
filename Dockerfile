# Multi-stage build for Go MCP server
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o mcp-server main.go

# Production image with SpamAssassin
FROM ubuntu:22.04

# Set environment variables
ENV DEBIAN_FRONTEND=noninteractive
ENV SA_MCP_LOG_LEVEL=info
ENV SA_MCP_SERVER_BIND_ADDR=0.0.0.0:8080

# Install SpamAssassin and dependencies
RUN apt-get update && apt-get install -y \
    spamassassin \
    spamc \
    curl \
    ca-certificates \
    tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create application user
RUN groupadd -r spamassassin && \
    useradd -r -g spamassassin -d /home/spamassassin -s /bin/bash spamassassin

# Create necessary directories
RUN mkdir -p \
    /etc/spamassassin-mcp \
    /var/lib/spamassassin \
    /var/log/spamassassin \
    /home/spamassassin

# Copy MCP server binary
COPY --from=builder /app/mcp-server /usr/local/bin/mcp-server

# Copy configuration files
COPY configs/ /etc/spamassassin-mcp/
COPY scripts/ /usr/local/bin/

# Set permissions
RUN chmod +x /usr/local/bin/mcp-server \
    /usr/local/bin/entrypoint.sh \
    /usr/local/bin/health-check.sh

# Set ownership
RUN chown -R spamassassin:spamassassin \
    /var/lib/spamassassin \
    /var/log/spamassassin \
    /home/spamassassin \
    /etc/spamassassin-mcp

# Expose MCP server port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD /usr/local/bin/health-check.sh

# Switch to non-root user
USER spamassassin
WORKDIR /home/spamassassin

# Start with entrypoint script
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["mcp-server"]