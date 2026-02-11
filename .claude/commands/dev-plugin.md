# /dev-plugin

Load the Crawl4AI Claude Code plugin codebase and system design into context, ready for development.

**Optional argument:** $ARGUMENTS — a one-line description of what you want to do (e.g., "add a new caching tool" or "fix the deep crawl timeout"). If empty, just load context and wait for instructions.

## Instructions

### Step 1: Read system design

Read the system design document to understand the full architecture:

```
python/.context/CLAUDE-PLUGIN-SYSTEM-DESIGN.md
```

### Step 2: Read all plugin source files

Read these files to have the full codebase in context:

**Core Python (MCP server + logic):**
- `python/crawl4ai_cloud/claude/mcp_server.py` — MCP tool definitions (9 tools)
- `python/crawl4ai_cloud/claude/core.py` — Backend-agnostic tool functions
- `python/crawl4ai_cloud/claude/config.py` — PluginConfig dataclass + load/save
- `python/crawl4ai_cloud/claude/backends/__init__.py` — CrawlBackend protocol + factory
- `python/crawl4ai_cloud/claude/backends/cloud.py` — Cloud API backend
- `python/crawl4ai_cloud/claude/backends/local.py` — OSS local backend
- `python/crawl4ai_cloud/claude/poll.py` — crawl4ai-poll CLI script
- `python/crawl4ai_cloud/claude/__init__.py` — Package exports

**Plugin distribution files:**
- `python/claude-plugin/.claude-plugin/plugin.json` — Plugin manifest
- `python/claude-plugin/.mcp.json` — MCP server config
- `python/claude-plugin/start.sh` — Bootstrap script
- `python/claude-plugin/hooks/hooks.json` — Session start hook
- `python/claude-plugin/scripts/check-setup.sh` — Setup check script
- `python/claude-plugin/skills/setup/SKILL.md` — /crawl4ai:setup skill
- `python/claude-plugin/skills/switch/SKILL.md` — /crawl4ai:switch skill
- `python/claude-plugin/skills/update/SKILL.md` — /crawl4ai:update skill

**Package config:**
- `python/pyproject.toml` — Entry points, dependencies, extras

**Marketplace:**
- `.claude-plugin/marketplace.json` — Marketplace catalog

### Step 3: Summarize what you loaded

After reading, briefly confirm:
- Number of files loaded
- Current plugin version (from plugin.json)
- Current SDK version (from pyproject.toml)
- Number of MCP tools
- Current branch and last commit

### Step 4: Handle the task

**If `$ARGUMENTS` is provided:**
- Analyze what the user wants based on their one-line description
- Identify which files need to change
- Propose a plan or start implementing depending on complexity
- For non-trivial changes, list the affected files and approach before coding

**If `$ARGUMENTS` is empty:**
- Say you're ready and list the key areas you can help with:
  - Adding new MCP tools
  - Modifying existing tools
  - Backend changes (cloud/local)
  - Plugin distribution (marketplace, skills, hooks)
  - Configuration changes
  - Bug fixes
- Wait for the user to tell you what they want

## Key reminders

- **Three-layer architecture**: mcp_server.py → core.py → backends. Always change all three layers when adding/modifying a tool.
- **Protocol**: New backend methods must be added to `CrawlBackend` protocol in `backends/__init__.py`.
- **Exports**: New core functions must be added to `__init__.py`.
- **Cloud API format**: Uses `"json_css"` not `"JsonCssExtractionStrategy"`. Check cloud API docs if unsure.
- **Deep crawl results**: Stored in S3 ZIP, not inline. `CrawlJob.results` is always `None`.
- **Python**: Use `python3` not `python` on this system.
- **Test API key**: Available in `~/.crawl4ai-cloud-sdk/.context/MEMORY.md`.
