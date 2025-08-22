# SpamAssassin MCP Server - Project Overview

## Purpose
A secure, containerized Model Context Protocol (MCP) server that integrates SpamAssassin for defensive email security analysis. This server provides Claude Code with comprehensive email analysis capabilities while maintaining strict security boundaries.

## Key Features
- **Defensive Security Only**: Email spam detection, sender reputation checking, rule testing, configuration inspection
- **Security-First Design**: No email sending/relay capabilities, no malicious content generation
- **Containerized Architecture**: Docker-based deployment with comprehensive security measures
- **MCP Integration**: Full Model Context Protocol compatibility for Claude Code integration

## Tech Stack
- **Language**: Go 1.23.0+ (toolchain 1.24.4)
- **Framework**: Model Context Protocol Go SDK v0.2.0
- **Dependencies**: 
  - Viper for configuration management
  - Logrus for structured logging
  - golang.org/x/time for rate limiting
- **Containerization**: Docker with multi-stage builds
- **External Integration**: SpamAssassin daemon for email analysis

## Architecture
- **MCP Server**: Main entry point (`main.go`) with tool registration
- **Configuration**: YAML-based config with environment variable overrides
- **Handlers**: MCP tool implementations for email analysis
- **SpamAssassin Client**: Wrapper for SpamAssassin daemon communication
- **Security**: Input validation, rate limiting, container hardening

## Available MCP Tools
1. `scan_email` - Analyze email content for spam probability
2. `check_reputation` - Check sender reputation and blacklists
3. `explain_score` - Detailed spam score breakdown
4. `get_config` - Retrieve current configuration
5. `update_rules` - Update SpamAssassin rule definitions
6. `test_rules` - Test custom rules in safe environment

## Security Measures
- Rate limiting (60 req/min, burst of 10)
- Input validation and size limits (10MB max)
- Container security (non-root, read-only filesystem)
- Network isolation with custom bridge
- Resource limits and health monitoring