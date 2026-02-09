# /crawl4ai:setup

Set up the Crawl4AI plugin for Claude Code.

## Instructions

Guide the user through setup:

1. **Check current config:**
   Read `~/.crawl4ai/claude_config.json` if it exists. Report current mode and whether API key is set.

2. **Choose mode:**
   Ask the user:
   - **Cloud mode** (default): Uses the Crawl4AI cloud API. Requires an API key from https://crawl4ai.com. Fast, no local browser needed.
   - **Local mode**: Uses OSS crawl4ai with a local Chromium browser. Free, no API key needed. Requires `pip install crawl4ai && crawl4ai-setup`.

3. **Configure based on choice:**

   **For cloud mode:**
   - Ask for API key (starts with `sk_live_` or `sk_test_`)
   - Write config to `~/.crawl4ai/claude_config.json`:
     ```json
     {"mode": "cloud", "api_key": "sk_live_xxx"}
     ```
   - Or suggest setting `CRAWL4AI_API_KEY` env var

   **For local mode:**
   - Check if crawl4ai is installed: `python3 -c "import crawl4ai; print(crawl4ai.__version__)"`
   - If not installed, run: `pip install crawl4ai && crawl4ai-setup`
   - Write config: `{"mode": "local"}`

4. **Verify setup:**
   Use the `crawl` MCP tool to crawl https://example.com. Confirm it returns markdown content.

5. **Report status:**
   Tell the user their setup is complete and which mode is active.
