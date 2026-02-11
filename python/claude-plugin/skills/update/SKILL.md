# /crawl4ai:update

Check for and apply updates to the Crawl4AI plugin and SDK.

## Instructions

1. **Check current versions:**
   - Read the installed plugin version: `cat ~/.claude/plugins/cache/crawl4ai-claude-plugins/crawl4ai/*/. claude-plugin/plugin.json 2>/dev/null` and extract the `version` field.
   - Check the installed SDK version: `python3 -c "import crawl4ai_cloud; print(crawl4ai_cloud.__version__)"`
   - If the SDK is not installed, note that.

2. **Check for plugin updates:**
   Run:
   ```
   /plugin marketplace update crawl4ai-claude-plugins
   ```
   This refreshes the marketplace catalog from the remote repository.

3. **Check if plugin needs reinstall:**
   After the marketplace update, compare the installed version with the latest available. If different, tell the user:
   ```
   Plugin update available: v{old} → v{new}
   Run: /plugin uninstall crawl4ai@crawl4ai-claude-plugins
   Then: /plugin install crawl4ai@crawl4ai-claude-plugins
   ```

4. **Update the SDK:**
   Run:
   ```bash
   pip3 install --upgrade crawl4ai-cloud-sdk[claude]
   ```
   Report the new version after upgrade.

5. **If using local mode, also update crawl4ai:**
   Check current config mode: `cat ~/.crawl4ai/claude_config.json`
   If mode is "local":
   ```bash
   pip3 install --upgrade crawl4ai
   ```

6. **Report summary:**
   Tell the user what was updated and whether a Claude Code restart is needed (restart is needed if the plugin version changed).
