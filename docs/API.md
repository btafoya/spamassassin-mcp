# SpamAssassin MCP Server API Reference

Complete documentation for all MCP tools provided by the SpamAssassin MCP server.

## Overview

The SpamAssassin MCP server provides 6 defensive security tools through the Model Context Protocol. All tools are designed for analysis and defensive security operations only.

## Security Notice

⚠️ **Defensive Operations Only**: All tools are designed exclusively for security analysis. No offensive capabilities are provided.

## Authentication & Rate Limiting

- **Rate Limiting**: 60 requests per minute with burst capacity of 10
- **No Authentication**: Currently runs in trusted environment
- **Request Size Limits**: Maximum email size 10MB

## Tools Reference

### Email Analysis Tools

#### `scan_email`

Analyze email content for spam probability and rule matches using SpamAssassin.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | ✅ | Raw email content including headers |
| `headers` | object | ❌ | Additional headers to analyze |
| `check_bayes` | boolean | ❌ | Include Bayesian analysis (default: false) |
| `verbose` | boolean | ❌ | Return detailed rule explanations (default: false) |

**Request Example:**
```json
{
  "tool": "scan_email",
  "params": {
    "content": "From: sender@example.com\nTo: recipient@example.com\nSubject: Test Email\n\nThis is a test email.",
    "verbose": true,
    "check_bayes": true
  }
}
```

**Response:**
```json
{
  "score": 2.1,
  "threshold": 5.0,
  "is_spam": false,
  "rules_hit": [
    {
      "name": "MISSING_HEADERS",
      "score": 1.0,
      "description": "Missing some standard email headers"
    },
    {
      "name": "SHORT_EMAIL",
      "score": 1.1,
      "description": "Email body is very short"
    }
  ],
  "summary": "Email analysis completed with detailed rule explanations...",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid email format or content too large
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: SpamAssassin processing error

---

#### `check_reputation`

Check sender reputation and domain/IP blacklists against configured security policies.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `sender` | string | ✅ | Email sender address |
| `domain` | string | ❌ | Sender domain (auto-extracted if not provided) |
| `ip` | string | ❌ | Sender IP address |

**Request Example:**
```json
{
  "tool": "check_reputation",
  "params": {
    "sender": "suspicious@spam-domain.com",
    "domain": "spam-domain.com",
    "ip": "192.168.1.100"
  }
}
```

**Response:**
```json
{
  "sender": "suspicious@spam-domain.com",
  "domain": "spam-domain.com",
  "ip": "192.168.1.100",
  "reputation": "bad",
  "blocked": true,
  "reasons": [
    "Domain spam-domain.com is blocked"
  ],
  "details": {
    "check_time": "2024-01-01T12:00:00Z",
    "source": "spamassassin-mcp"
  }
}
```

**Reputation Values:**
- `good`: Sender is whitelisted or trusted
- `bad`: Sender is blacklisted or suspicious
- `unknown`: No reputation data available

---

#### `explain_score`

Provide detailed explanation of how a spam score was calculated, including rule breakdown and reasoning.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `email_content` | string | ✅ | Raw email content to analyze |

**Request Example:**
```json
{
  "tool": "explain_score",
  "params": {
    "email_content": "Subject: URGENT: Free Money!\n\nClick here to claim your $1,000,000 prize!"
  }
}
```

**Response:**
```json
{
  "final_score": 12.5,
  "rule_details": [
    {
      "name": "URGENT_SUBJECT",
      "score": 3.0,
      "description": "Subject contains urgent language"
    },
    {
      "name": "MONEY_PRIZE",
      "score": 4.5,
      "description": "Content mentions money prizes"
    },
    {
      "name": "SUSPICIOUS_LINK",
      "score": 2.0,
      "description": "Contains potentially suspicious links"
    },
    {
      "name": "EXCLAMATION_MARKS",
      "score": 1.5,
      "description": "Excessive use of exclamation marks"
    },
    {
      "name": "SHORT_BODY",
      "score": 1.5,
      "description": "Very short email body"
    }
  ],
  "bayes_score": 0.95,
  "network_tests": [
    "DNSBL_CHECK_PASSED",
    "SPF_CHECK_FAILED"
  ],
  "explanation": "Final Score: 12.50 (Threshold: 5.00)\nClassification: SPAM\n\nRules Triggered:\n  URGENT_SUBJECT: 3.00 - Subject contains urgent language\n  MONEY_PRIZE: 4.50 - Content mentions money prizes\n  SUSPICIOUS_LINK: 2.00 - Contains potentially suspicious links\n  EXCLAMATION_MARKS: 1.50 - Excessive use of exclamation marks\n  SHORT_BODY: 1.50 - Very short email body\n\nBayesian Analysis: 95% spam probability\n\nNetwork Tests:\n  ✓ DNSBL check passed\n  ✗ SPF validation failed"
}
```

### Configuration Management Tools

#### `get_config`

Retrieve current SpamAssassin configuration and server status information.

**Parameters:** None

**Request Example:**
```json
{
  "tool": "get_config",
  "params": {}
}
```

**Response:**
```json
{
  "version": "3.4.6",
  "threshold": 5.0,
  "bayes_enabled": true,
  "rule_count": 1247,
  "settings": {
    "host": "localhost",
    "port": 783,
    "timeout": "30s",
    "max_email_size": 10485760,
    "rate_limit_per_minute": 60
  }
}
```

---

#### `update_rules`

Update SpamAssassin rule definitions from official sources (defensive updates only).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source` | string | ❌ | Rule source: "official" or "custom" (default: "official") |
| `force` | boolean | ❌ | Force update even if recent (default: false) |

**Request Example:**
```json
{
  "tool": "update_rules",
  "params": {
    "source": "official",
    "force": false
  }
}
```

**Response:**
```json
{
  "status": "success",
  "message": "Rules updated successfully",
  "timestamp": "2024-01-01T12:00:00Z",
  "rules_updated": 1247,
  "last_update": "2024-01-01T11:30:00Z"
}
```

### Rule Testing Tools

#### `test_rules`

Test custom SpamAssassin rules against sample emails in a safe, isolated environment.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `rules` | string | ✅ | Custom rule definitions in SpamAssassin format |
| `test_emails` | array | ✅ | Array of sample email strings to test |

**Request Example:**
```json
{
  "tool": "test_rules",
  "params": {
    "rules": "header LOCAL_TEST Subject =~ /test/i\ndescribe LOCAL_TEST Test rule for subject\nscore LOCAL_TEST 2.0",
    "test_emails": [
      "Subject: This is a test email\n\nTest content here.",
      "Subject: Normal email\n\nRegular content here."
    ]
  }
}
```

**Response:**
```json
{
  "results": [
    {
      "email": "Subject: This is a test email...",
      "score": 2.0,
      "is_spam": false,
      "rules_matched": ["LOCAL_TEST"]
    },
    {
      "email": "Subject: Normal email...",
      "score": 0.0,
      "is_spam": false,
      "rules_matched": []
    }
  ],
  "summary": "Tested 2 emails against custom rules. 1 email matched the test rule."
}
```

## Error Handling

### Common Error Codes

| Code | Description | Solution |
|------|-------------|----------|
| `400` | Bad Request - Invalid parameters | Check parameter format and requirements |
| `413` | Request Entity Too Large | Reduce email size (max 10MB) |
| `429` | Too Many Requests | Wait before retrying (rate limit: 60/min) |
| `500` | Internal Server Error | Check server logs and SpamAssassin status |
| `503` | Service Unavailable | SpamAssassin daemon not responding |

### Error Response Format

```json
{
  "error": {
    "code": "400",
    "message": "Invalid email format",
    "details": "Email content must include valid headers",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

## Rate Limiting

All tools are subject to rate limiting:

- **Limit**: 60 requests per minute
- **Burst**: Up to 10 requests in quick succession
- **Reset**: Rate limit window resets every minute
- **Headers**: Rate limit information included in response headers

## Request/Response Headers

### Request Headers
- `Content-Type: application/json`
- `User-Agent: claude-code-mcp-client/1.0`

### Response Headers
- `Content-Type: application/json`
- `X-RateLimit-Limit: 60`
- `X-RateLimit-Remaining: 59`
- `X-RateLimit-Reset: 1640995200`

## Security Considerations

### Input Validation
- All email content is validated for format and size
- Headers are sanitized to prevent injection attacks
- IP addresses and domains are validated with regex patterns
- Custom rules are parsed safely in isolated environment

### Data Privacy
- No email content is permanently stored
- All processing is done in memory
- Logs contain only metadata, not email content
- Temporary files are automatically cleaned up

### Audit Logging
- All API calls are logged with timestamps
- Security events are logged at WARN level
- Rate limit violations are tracked
- Errors include correlation IDs for debugging

## Usage Examples

### Claude Code Integration

```bash
# Connect to MCP server
claude --mcp-server spamassassin tcp://localhost:8080

# Scan suspicious email
/scan_email --content "$(cat suspicious.eml)" --verbose

# Check sender reputation
/check_reputation --sender "unknown@suspicious-domain.com"

# Get detailed score explanation
/explain_score --email_content "Subject: Free Money!\n\nClaim your prize!"

# Test custom rules
/test_rules --rules "header CUSTOM_RULE Subject =~ /urgent/i" --test_emails '["Subject: Urgent message\n\nContent"]'
```

### Programmatic Usage

For direct integration with other tools, the server accepts standard HTTP POST requests to the MCP endpoint with JSON payloads following the MCP specification.

## Performance Considerations

- **Response Times**: Typical scan takes 50-200ms
- **Throughput**: ~300 emails/minute sustained
- **Memory Usage**: ~100MB base + 1MB per concurrent scan
- **Disk Usage**: ~50MB for rules + log files

## Monitoring and Health Checks

- **Health Endpoint**: Built-in health check script
- **Metrics**: Response times, error rates, rule hit counts
- **Alerts**: Automatic alerts for high error rates or long response times
- **Logging**: Structured JSON logs for easy parsing