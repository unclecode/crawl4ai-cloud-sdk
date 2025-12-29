# Crawl4AI Cloud SDK

<div align="center">

**Multi-language SDK for [Crawl4AI Cloud API](https://api.crawl4ai.com)**

Mirrors the OSS library API exactly — copy your existing code, add an API key, done.

[![GitHub Stars](https://img.shields.io/github/stars/unclecode/crawl4ai-cloud?style=social)](https://github.com/unclecode/crawl4ai-cloud/stargazers)
[![Discord](https://img.shields.io/badge/Discord-Join%20Us-5865F2?logo=discord&logoColor=white)](https://discord.gg/jP8KfhDhyN)
[![Twitter Follow](https://img.shields.io/twitter/follow/crawl4ai?style=social)](https://x.com/crawl4ai)

</div>

---

## Why Cloud?

| | OSS (Self-hosted) | Cloud |
|---|---|---|
| **Setup** | Install Playwright, browsers, dependencies | `pip install crawl4ai-cloud` |
| **Size** | ~800MB+ | ~10MB |
| **Scaling** | Manage your own infrastructure | We handle it |
| **Proxies** | BYO | Built-in (datacenter, residential) |
| **Best for** | Full control, custom deployments | Fast integration, production workloads |

## Get Your API Key

1. Go to **[api.crawl4ai.com](https://api.crawl4ai.com)**
2. Sign up and get your API key (`sk_live_...`)
3. Install your preferred SDK below

---

## SDKs

| Language | Package | Install |
|----------|---------|---------|
| [Python](./python) | `crawl4ai-cloud` | `pip install crawl4ai-cloud` |
| [Node.js](./nodejs) | `crawl4ai-cloud` | `npm install crawl4ai-cloud` |
| [Go](./go) | `crawl4ai` | `go get github.com/unclecode/crawl4ai-cloud/go` |

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
import "github.com/unclecode/crawl4ai-cloud/go/pkg/crawl4ai"

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

## Not Using Cloud?

We've got you covered:

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

## Community

<p align="center">
  <a href="https://github.com/unclecode/crawl4ai">
    <img src="https://img.shields.io/badge/OSS%20Repo-Star%20Us-yellow?style=for-the-badge&logo=github" alt="Star OSS Repo" />
  </a>
  <a href="https://discord.gg/jP8KfhDhyN">
    <img src="https://img.shields.io/badge/Discord-Join%20Us-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Join Discord" />
  </a>
  <a href="https://x.com/crawl4ai">
    <img src="https://img.shields.io/badge/Twitter-Follow-000000?style=for-the-badge&logo=x&logoColor=white" alt="Follow on X" />
  </a>
  <a href="https://www.linkedin.com/company/crawl4ai">
    <img src="https://img.shields.io/badge/LinkedIn-Follow-0077B5?style=for-the-badge&logo=linkedin&logoColor=white" alt="Follow on LinkedIn" />
  </a>
</p>

---

## Documentation

- [Crawl4AI Docs](https://docs.crawl4ai.com)
- [Cloud API Reference](https://api.crawl4ai.com/docs)
- [OSS Repository](https://github.com/unclecode/crawl4ai)

## License

Apache 2.0
