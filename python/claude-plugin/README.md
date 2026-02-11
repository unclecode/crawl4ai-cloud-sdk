# Crawl4AI Plugin for Claude Code

Give Claude the ability to crawl websites, extract structured data, take screenshots, and discover URLs — all from natural conversation.

**9 tools** for web crawling and data extraction, with two backend options:

- **Cloud mode** — Fast, no local browser needed. Uses the [Crawl4AI Cloud API](https://api.crawl4ai.com).
- **Local mode** — Free, fully private. Uses the open-source [Crawl4AI](https://github.com/unclecode/crawl4ai) library with a local Chromium browser.

## Installation

### From inside Claude Code

```
/plugin marketplace add unclecode/crawl4ai-cloud-sdk
/plugin install crawl4ai@crawl4ai-claude-plugins
```

Restart Claude Code, then run `/crawl4ai:setup` to configure.

### From the terminal

```bash
claude plugin marketplace add unclecode/crawl4ai-cloud-sdk
claude plugin install crawl4ai@crawl4ai-claude-plugins --scope user
claude
```

On first launch, the plugin automatically installs its Python dependencies. No manual `pip install` needed.

## Setup

After installing and restarting Claude Code, run:

```
/crawl4ai:setup
```

This walks you through choosing a mode and configuring credentials.

### Cloud mode (recommended)

1. Create an account at [api.crawl4ai.com](https://api.crawl4ai.com)
2. Copy your API key (starts with `sk_live_` or `sk_test_`)
3. Run `/crawl4ai:setup`, choose **cloud**, paste your key

### Local mode

1. Run `/crawl4ai:setup`, choose **local**
2. The setup installs [crawl4ai](https://github.com/unclecode/crawl4ai) and a local Chromium browser automatically
3. No API key needed — everything runs on your machine

## Slash Commands

| Command | Description |
|---------|-------------|
| `/crawl4ai:setup` | Configure the plugin (mode, API key) |
| `/crawl4ai:switch` | Toggle between cloud and local mode |
| `/crawl4ai:update` | Check for plugin and SDK updates |

## Tools

The plugin provides 9 MCP tools that Claude can use automatically based on your requests.

### crawl

Crawl a single page or start a multi-page deep crawl.

**Single page:**

> Crawl https://example.com and summarize the content

> Get the markdown from https://docs.crawl4ai.com/core/quickstart

**Deep crawl (multi-page):**

> Deep crawl https://docs.crawl4ai.com with max 10 pages

> Crawl https://example.com/blog with BFS strategy, depth 2, max 20 pages

Deep crawl is non-blocking — it returns a job ID immediately. Claude will poll the job status and fetch results when ready.

### extract

Extract structured data from a page using a CSS or XPath schema.

> Extract all book titles and prices from https://books.toscrape.com

> Extract every quote and author from https://quotes.toscrape.com

> Extract the product name, price, and rating from this page using the schema I gave you

### schema

Generate an extraction schema automatically using AI.

> Generate a schema to extract product listings from https://books.toscrape.com

> Create a CSS extraction schema for news articles on https://news.ycombinator.com

Use the generated schema with the `extract` tool for repeatable, structured extraction.

### map

Discover URLs on a domain via sitemap, Common Crawl, or both.

> Find all URLs on https://docs.crawl4ai.com

> Map https://example.com using sitemap, filter for /blog/*

> Discover URLs on https://crawl4ai.com from Common Crawl related to "pricing"

### screenshot

Take a screenshot of any web page.

> Take a screenshot of https://example.com

> Screenshot https://crawl4ai.com, wait for the hero section to load

> Screenshot just the navigation bar on https://docs.crawl4ai.com

### job_status

Check the progress of a deep crawl job.

> What's the status of job scan_abc123?

> Check if my deep crawl is done

### fetch

Download and parse results from a completed deep crawl.

> Fetch the results from that deep crawl

> Download the crawl results and save them to /tmp/results.json

For large result sets, use the `save_to` parameter to write results to a file instead of returning them inline.

### profile_list

List available browser profiles for authenticated crawling.

> Show my browser profiles

### profile_create

Create a new browser profile for sites that require login.

> Create a browser profile for authenticated crawling

## Example Conversations

### Basic crawling

> **You:** Crawl https://docs.crawl4ai.com/core/quickstart and give me the key points
>
> Claude crawls the page and returns a summary of the quickstart guide.

### Structured data extraction

> **You:** I need to scrape all books from https://books.toscrape.com — titles, prices, and ratings
>
> Claude generates a schema, extracts the data, and presents it in a table.

### Multi-page deep crawl

> **You:** Deep crawl https://docs.crawl4ai.com, max 15 pages. Save results to /tmp/docs.json
>
> Claude starts the crawl, polls in the background, fetches results when done, and saves to file.

### URL discovery

> **You:** What pages exist on https://crawl4ai.com? I'm looking for documentation pages.
>
> Claude maps the domain and filters for doc-related URLs.

### Screenshot

> **You:** Take a full-page screenshot of https://example.com
>
> Claude captures the screenshot and displays it.

### Schema generation + extraction pipeline

> **You:** Figure out how to extract job listings from https://example.com/careers, then extract them all
>
> Claude generates a schema from the page structure, then uses it to extract all listings.

## How Deep Crawl Works

Deep crawl runs in three phases:

1. **Start** — `crawl` tool with `deep_crawl=true` sends the request and returns a job ID immediately (no blocking).
2. **Poll** — Claude checks `job_status` periodically, or runs `crawl4ai-poll` in the background.
3. **Fetch** — Once complete, `fetch` downloads the results ZIP from S3, parses each page, and returns the content.

For large sites, use `save_to` with the fetch tool to write results to a file:

> Deep crawl https://docs.crawl4ai.com max 50 pages, save results to ./crawl_output.json

### Background polling CLI

The plugin includes `crawl4ai-poll` for background job monitoring:

```bash
crawl4ai-poll --job-id scan_xxx --api-key $CRAWL4AI_API_KEY
```

Claude Code can run this in the background automatically while continuing the conversation.

## Configuration

Config is stored at `~/.crawl4ai/claude_config.json`:

```json
{
  "mode": "cloud",
  "api_key": "sk_live_..."
}
```

### Environment variables

| Variable | Description |
|----------|-------------|
| `CRAWL4AI_API_KEY` | API key (overrides config file) |
| `CRAWL4AI_MODE` | `cloud` or `local` (overrides config file) |
| `CRAWL4AI_API_BASE_URL` | Custom API base URL |

## Cloud vs Local

| | Cloud | Local |
|---|---|---|
| **Speed** | Fast (parallel cloud browsers) | Depends on your machine |
| **Setup** | API key only | Installs Chromium (~400MB) |
| **Cost** | Pay per crawl | Free |
| **Privacy** | Pages processed on cloud | Everything stays local |
| **Deep crawl** | Full support with S3 results | Depends on OSS version |
| **Auth crawling** | Via cloud profiles | Via local browser profiles |

## Links

- **Cloud API** — [api.crawl4ai.com](https://api.crawl4ai.com)
- **Cloud SDK** — [github.com/unclecode/crawl4ai-cloud-sdk](https://github.com/unclecode/crawl4ai-cloud-sdk)
- **OSS Library** — [github.com/unclecode/crawl4ai](https://github.com/unclecode/crawl4ai)
- **Discord** — [discord.gg/jP8KfhDhyN](https://discord.gg/jP8KfhDhyN)

## License

Apache-2.0
