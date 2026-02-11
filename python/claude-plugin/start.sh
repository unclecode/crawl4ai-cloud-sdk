#!/usr/bin/env bash
set -euo pipefail

if ! python3 -c "import crawl4ai_cloud.claude.mcp_server" 2>/dev/null; then
    echo "Installing crawl4ai-cloud-sdk..." >&2
    if command -v uv &>/dev/null; then
        uv pip install --quiet "crawl4ai-cloud-sdk[claude]" >&2
    else
        pip3 install --quiet "crawl4ai-cloud-sdk[claude]" >&2
    fi
fi

exec python3 -m crawl4ai_cloud.claude.mcp_server "$@"
