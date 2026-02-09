# /crawl4ai:switch

Switch between cloud and local backends for Crawl4AI.

## Instructions

1. **Read current config:**
   Read `~/.crawl4ai/claude_config.json`. Report the current mode.

2. **Switch to the other mode:**
   - If currently "cloud", switch to "local"
   - If currently "local", switch to "cloud"
   - Or if the user specifies a mode, switch to that

3. **Handle prerequisites:**
   - **Switching to cloud:** Ask for API key if not already set (check env var CRAWL4AI_API_KEY too)
   - **Switching to local:** Check `python3 -c "import crawl4ai"`. If missing: `pip install crawl4ai && crawl4ai-setup`

4. **Write new config:**
   Update `~/.crawl4ai/claude_config.json` with the new mode.

5. **Verify:**
   Use the `crawl` MCP tool to crawl https://example.com with the new backend. Confirm it works.

6. **Report:**
   Tell the user the switch is complete and the new active mode.
