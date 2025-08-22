#!/bin/bash
set -euo pipefail

# SpamAssassin MCP Server Test Script
# This script demonstrates how to interact with the MCP server

MCP_HOST=${1:-"localhost:8080"}
SAMPLE_EMAIL="examples/sample-email.eml"

echo "Testing SpamAssassin MCP Server at $MCP_HOST"
echo "=============================================="

# Check if sample email exists
if [ ! -f "$SAMPLE_EMAIL" ]; then
    echo "Error: Sample email not found at $SAMPLE_EMAIL"
    exit 1
fi

# Test 1: Scan Email
echo "Test 1: Scanning sample email..."
echo "Email content:"
echo "$(head -5 $SAMPLE_EMAIL)"
echo "..."
echo

# In a real MCP environment, you would use the MCP protocol
# For testing, we're showing the expected JSON payload

cat << EOF
Expected scan_email request:
{
  "tool": "scan_email",
  "params": {
    "content": "$(cat $SAMPLE_EMAIL | sed 's/"/\\"/g' | tr '\n' '\\' | sed 's/\\/\\n/g')",
    "verbose": true,
    "check_bayes": true
  }
}
EOF

echo
echo "Expected response:"
cat << EOF
{
  "score": 15.3,
  "threshold": 5.0,
  "is_spam": true,
  "rules_hit": [
    {"name": "URGENT_SUBJECT", "score": 2.5, "description": "Subject contains urgent words"},
    {"name": "MONEY_PRIZE", "score": 3.0, "description": "Contains money prize claims"},
    {"name": "SUSPICIOUS_LINK", "score": 2.8, "description": "Contains suspicious links"}
  ],
  "summary": "Email classified as spam due to multiple indicators...",
  "timestamp": "2024-01-01T12:00:00Z"
}
EOF

echo
echo "=============================================="

# Test 2: Check Reputation
echo "Test 2: Checking sender reputation..."
cat << EOF
Expected check_reputation request:
{
  "tool": "check_reputation",
  "params": {
    "sender": "suspicious@spam-domain.com",
    "domain": "spam-domain.com"
  }
}
EOF

echo
echo "Expected response:"
cat << EOF
{
  "sender": "suspicious@spam-domain.com",
  "domain": "spam-domain.com",
  "reputation": "bad",
  "blocked": true,
  "reasons": ["Domain spam-domain.com is blocked"],
  "details": {
    "check_time": "2024-01-01T12:00:00Z",
    "source": "spamassassin-mcp"
  }
}
EOF

echo
echo "=============================================="

# Test 3: Get Configuration
echo "Test 3: Getting server configuration..."
cat << EOF
Expected get_config request:
{
  "tool": "get_config",
  "params": {}
}
EOF

echo
echo "Expected response:"
cat << EOF
{
  "version": "3.4.x",
  "threshold": 5.0,
  "bayes_enabled": true,
  "rule_count": 1000,
  "settings": {
    "host": "localhost",
    "port": 783,
    "timeout": "30s"
  }
}
EOF

echo
echo "=============================================="

# Test 4: Explain Score
echo "Test 4: Explaining spam score..."
cat << EOF
Expected explain_score request:
{
  "tool": "explain_score",
  "params": {
    "email_content": "Subject: Free Money!\\n\\nClaim your prize now!"
  }
}
EOF

echo
echo "Expected response:"
cat << EOF
{
  "final_score": 8.5,
  "rule_details": [
    {"name": "MONEY_SUBJECT", "score": 3.0, "description": "Subject mentions money"},
    {"name": "PRIZE_CLAIM", "score": 2.5, "description": "Body contains prize claims"},
    {"name": "EXCLAMATION", "score": 1.0, "description": "Excessive exclamation marks"}
  ],
  "explanation": "Final Score: 8.50 (Threshold: 5.00)\\nClassification: SPAM\\n\\nRules Triggered:\\n  MONEY_SUBJECT: 3.00 - Subject mentions money\\n  PRIZE_CLAIM: 2.50 - Body contains prize claims\\n  EXCLAMATION: 1.00 - Excessive exclamation marks"
}
EOF

echo
echo "=============================================="
echo "Test script completed!"
echo
echo "To connect with Claude Code:"
echo "  claude --mcp-server spamassassin tcp://$MCP_HOST"
echo
echo "To check server health:"
echo "  curl -f http://$MCP_HOST/health || docker-compose exec spamassassin-mcp /usr/local/bin/health-check.sh"