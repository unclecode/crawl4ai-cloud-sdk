# Crawl4AI Cloud SDK Documentation

Multi-language SDK for [Crawl4AI Cloud API](https://api.crawl4ai.com).

Mirrors the OSS library API exactly — copy your existing code, add an API key, done.

---

## Installation

| Language | Package | Install |
|----------|---------|---------|
| [Python](./python) | `crawl4ai-cloud` | `pip install crawl4ai-cloud` |
| [Node.js](./nodejs) | `crawl4ai-cloud` | `npm install crawl4ai-cloud` |
| [Go](./go) | `crawl4ai` | `go get github.com/unclecode/crawl4ai-cloud-sdk/go` |

---

## Quick Start

### Python

```python
from crawl4ai_cloud import AsyncWebCrawler

async with AsyncWebCrawler(api_key="sk_live_...") as crawler:
    result = await crawler.run("https://example.com")
    print(result.markdown.raw_markdown)
```

### Node.js

```typescript
import { AsyncWebCrawler } from 'crawl4ai-cloud';

const crawler = new AsyncWebCrawler({ apiKey: 'sk_live_...' });
const result = await crawler.run('https://example.com');
console.log(result.markdown?.rawMarkdown);
await crawler.close();
```

### Go

```go
import "github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"

crawler, _ := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
    APIKey: "sk_live_...",
})
result, _ := crawler.Run("https://example.com", nil)
fmt.Println(result.Markdown.RawMarkdown)
```

---

## Migration from OSS

Zero learning curve — your existing code works with minimal changes:

```python
# Before (OSS - local)
from crawl4ai import AsyncWebCrawler
async with AsyncWebCrawler() as crawler:
    result = await crawler.arun(url)

# After (Cloud - just change import + add key)
from crawl4ai_cloud import AsyncWebCrawler
async with AsyncWebCrawler(api_key="sk_...") as crawler:
    result = await crawler.run(url)  # arun() also works!
```

---

## Cloud vs OSS

| | OSS (Self-hosted) | Cloud |
|---|---|---|
| **Setup** | Install Playwright, browsers, dependencies | `pip install crawl4ai-cloud` |
| **Size** | ~800MB+ | ~10MB |
| **Scaling** | Manage your own infrastructure | We handle it |
| **Proxies** | BYO | Built-in (datacenter, residential) |
| **Best for** | Full control, custom deployments | Fast integration, production workloads |

---

## Not Using Cloud?

### Self-Hosted (OSS)
Want full control? Use our open-source library:
- **Repository**: [github.com/unclecode/crawl4ai](https://github.com/unclecode/crawl4ai)
- **50k+ stars**, battle-tested by the community
- Full Docker support, deploy anywhere

### Professional Services
Need custom solutions, enterprise support, or dedicated infrastructure?
- Join our [Discord](https://discord.gg/jP8KfhDhyN) and reach out
- We offer consulting, custom deployments, and enterprise plans

---

## API Reference

Full API documentation: [api.crawl4ai.com/docs](https://api.crawl4ai.com/docs)

---

## Language-Specific Docs

- [Python SDK](./python/README.md)
- [Node.js SDK](./nodejs/README.md)
- [Go SDK](./go/README.md)
