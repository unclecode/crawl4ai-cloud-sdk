#!/usr/bin/env bash
# Check if Crawl4AI plugin is configured. Runs on SessionStart.

CONFIG_FILE="$HOME/.crawl4ai/claude_config.json"

if [ ! -f "$CONFIG_FILE" ]; then
    echo "[Crawl4AI] Plugin installed but not configured."
    echo "  1. Restart Claude Code (if you just installed the plugin)"
    echo "  2. Run /crawl4ai:setup to configure your API key or local mode"
else
    MODE=$(python3 -c "import json; print(json.load(open('$CONFIG_FILE')).get('mode','cloud'))" 2>/dev/null || echo "cloud")
    echo "[Crawl4AI] Ready ($MODE mode). 9 tools available. Run /crawl4ai:switch to change mode."
fi
