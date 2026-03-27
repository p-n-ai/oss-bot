#!/usr/bin/env bash
set -euo pipefail

# Test the webhook handler with a simulated GitHub issue_comment event.
# Usage: ./scripts/test-webhook.sh [bot-url]
#   bot-url defaults to http://localhost:8090

BOT_URL="${1:-http://localhost:8090}"
SECRET="${OSS_GITHUB_WEBHOOK_SECRET:-test-secret}"

BODY='{"action":"created","comment":{"body":"@oss-bot add teaching notes for F3-02","user":{"login":"testuser"}},"issue":{"number":1},"repository":{"full_name":"p-n-ai/oss"}}'

# Compute HMAC-SHA256 signature
SIGNATURE=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print $2}')

echo "Sending test webhook to $BOT_URL/webhook"
echo "Command: @oss-bot add teaching notes for F3-02"
echo ""

curl -s -X POST "$BOT_URL/webhook" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: issue_comment" \
  -H "X-Hub-Signature-256: sha256=$SIGNATURE" \
  -d "$BODY" \
  -w "\nHTTP Status: %{http_code}\n"
