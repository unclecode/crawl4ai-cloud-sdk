#!/usr/bin/env bash
# Crawl4AI health check + credits remaining
set -euo pipefail

API_URL="${CRAWL4AI_API_URL:-$(python3 -c 'import json; print(json.load(open("'"$HOME"'/.crawl4ai/claude_config.json")).get("api_base_url","https://api.crawl4ai.com"))' 2>/dev/null || echo 'https://api.crawl4ai.com')}"
API_KEY="${CRAWL4AI_API_KEY:-$(python3 -c 'import json; print(json.load(open("'"$HOME"'/.crawl4ai/claude_config.json")).get("api_key",""))' 2>/dev/null || echo '')}"

echo "=== Health ==="
curl -s "$API_URL/health" 2>/dev/null | python3 -m json.tool 2>/dev/null || echo '{"status": "unreachable"}'

if [ -n "$API_KEY" ]; then
    echo ""
    echo "=== Storage ==="
    curl -s -H "X-API-Key: $API_KEY" "$API_URL/v1/crawl/storage" 2>/dev/null | python3 -m json.tool 2>/dev/null || echo '{"error": "failed to fetch storage"}'
else
    echo ""
    echo "No API key configured. Run /crawl4ai:setup"
fi
