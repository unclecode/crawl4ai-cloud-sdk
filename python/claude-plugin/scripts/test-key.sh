#!/usr/bin/env bash
# Validate Crawl4AI API key with a minimal call
set -euo pipefail

API_URL="${CRAWL4AI_API_URL:-$(python3 -c 'import json; print(json.load(open("'"$HOME"'/.crawl4ai/claude_config.json")).get("api_base_url","https://api.crawl4ai.com"))' 2>/dev/null || echo 'https://api.crawl4ai.com')}"
API_KEY="${CRAWL4AI_API_KEY:-$(python3 -c 'import json; print(json.load(open("'"$HOME"'/.crawl4ai/claude_config.json")).get("api_key",""))' 2>/dev/null || echo '')}"

if [ -z "$API_KEY" ]; then
    echo '{"success": false, "error": "No API key. Set CRAWL4AI_API_KEY or run /crawl4ai:setup"}'
    exit 1
fi

echo "Testing key: ${API_KEY:0:20}..."
RESULT=$(curl -s -X POST "$API_URL/v1/markdown" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "strategy": "http"}' \
  --max-time 30)

SUCCESS=$(echo "$RESULT" | python3 -c "import json,sys; print(json.load(sys.stdin).get('success','?'))" 2>/dev/null || echo "false")
CREDITS=$(echo "$RESULT" | python3 -c "import json,sys; u=json.load(sys.stdin).get('usage',{}); print(f'used={u.get(\"credits_used\",\"?\")}, remaining={u.get(\"credits_remaining\",\"?\")}')" 2>/dev/null || echo "unknown")

if [ "$SUCCESS" = "True" ]; then
    echo "API key valid. Credits: $CREDITS"
else
    echo "API key test failed."
    echo "$RESULT" | python3 -m json.tool 2>/dev/null || echo "$RESULT"
fi
