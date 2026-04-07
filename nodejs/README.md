# Crawl4AI Cloud SDK for Node.js

One-line web scraping, screenshots, and structured extraction -- powered by [Crawl4AI Cloud](https://api.crawl4ai.com).

[![npm version](https://badge.fury.io/js/crawl4ai-cloud.svg)](https://badge.fury.io/js/crawl4ai-cloud) v0.3.0

## Install

```bash
npm install crawl4ai-cloud
```

## Quick Start

```typescript
import { AsyncWebCrawler } from 'crawl4ai-cloud';

const c = new AsyncWebCrawler({ apiKey: 'sk_live_...' });
```

### Get Markdown

```typescript
const { markdown } = await c.markdown('https://example.com');
console.log(markdown);
```

### Take a Screenshot

```typescript
const { screenshot } = await c.screenshot('https://example.com');
// screenshot is a base64-encoded PNG
```

### Extract Structured Data

```typescript
const { data } = await c.extract('https://news.ycombinator.com', {
  query: 'Extract all story titles and their URLs',
});
console.log(data); // [{ title: '...', url: '...' }, ...]
```

### Map a Site

```typescript
const { urls, totalUrls } = await c.map('https://docs.example.com');
console.log(`Found ${totalUrls} pages`);
```

### Crawl an Entire Site

```typescript
const job = await c.crawlSite('https://docs.example.com', {
  maxPages: 50,
});
console.log(`Job ${job.jobId} queued, ${job.discoveredUrls} pages found`);
```

## Wrapper API Reference

| Method | Endpoint | Returns | Description |
|--------|----------|---------|-------------|
| `markdown(url)` | `POST /v1/markdown` | `MarkdownResponse` | Clean markdown from any page |
| `screenshot(url)` | `POST /v1/screenshot` | `ScreenshotResponse` | Full-page screenshot (PNG) or PDF |
| `extract(url)` | `POST /v1/extract` | `ExtractResponse` | LLM or schema-based structured extraction |
| `map(url)` | `POST /v1/map` | `MapResponse` | Discover all URLs under a domain |
| `crawlSite(url)` | `POST /v1/crawl/site` | `SiteCrawlResponse` | Crawl a full site (async job) |

### Markdown Options

```typescript
const res = await c.markdown('https://example.com', {
  strategy: 'browser',   // 'browser' (default) or 'http'
  fit: true,              // return cleaned/fitted markdown (default true)
  include: ['article'],   // CSS selectors to include
  proxy: { mode: 'residential', country: 'US' },
  bypassCache: true,
});
// res.markdown, res.fitMarkdown, res.links, res.media, res.metadata
```

### Screenshot Options

```typescript
const res = await c.screenshot('https://example.com', {
  fullPage: true,         // full page vs viewport (default true)
  pdf: true,              // also generate PDF
  waitFor: 'css:.loaded', // wait for element before capture
});
// res.screenshot (base64 PNG), res.pdf (base64 PDF)
```

### Extract Options

```typescript
// Auto mode -- LLM picks the best approach
const res = await c.extract('https://example.com/products', {
  query: 'Extract product name, price, and rating',
  method: 'auto', // 'auto' (default), 'llm', or 'schema'
});

// Schema mode -- CSS/XPath selectors, no LLM cost
const res = await c.extract('https://example.com/products', {
  method: 'schema',
  schema: {
    name: 'Products',
    baseSelector: '.product-card',
    fields: [
      { name: 'title', selector: 'h2', type: 'text' },
      { name: 'price', selector: '.price', type: 'text' },
    ],
  },
});
```

### Map Options

```typescript
const res = await c.map('https://example.com', {
  mode: 'deep',            // 'default' or 'deep'
  maxUrls: 500,
  includeSubdomains: true,
  query: 'blog posts',     // relevance scoring
  scoreThreshold: 0.5,
});
```

### Site Crawl Options

```typescript
const job = await c.crawlSite('https://docs.example.com', {
  maxPages: 100,
  discovery: 'bfs',        // 'map' (default), 'bfs', 'dfs', 'best_first'
  strategy: 'browser',     // 'browser' (default) or 'http'
  fit: true,
  pattern: '/docs/*',      // URL pattern filter
  maxDepth: 3,
});
```

## Async / Batch Operations

For multiple URLs, use the batch variants. They return a `WrapperJob` you can poll or wait on.

### markdownMany

```typescript
const job = await c.markdownMany(
  ['https://a.com', 'https://b.com', 'https://c.com'],
  { wait: true, timeout: 120 },
);
console.log(`Job ${job.jobId}: ${job.status}`);
```

### screenshotMany

```typescript
const job = await c.screenshotMany(
  ['https://a.com', 'https://b.com'],
  { fullPage: true, wait: true },
);
```

### extractMany

```typescript
const job = await c.extractMany(
  ['https://a.com/products', 'https://b.com/products'],
  {
    method: 'llm', // 'auto' not supported for batch -- use 'llm' or 'schema'
    query: 'Extract product name and price',
    wait: true,
  },
);
```

### Batch Options (shared)

All batch methods accept these additional options:

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `wait` | `boolean` | `false` | Block until job completes |
| `pollInterval` | `number` | `2` | Seconds between status checks (when `wait: true`) |
| `timeout` | `number` | -- | Max seconds to wait |
| `webhookUrl` | `string` | -- | URL to POST results to on completion |
| `priority` | `number` | `5` | Job priority (1=high, 5=normal, 10=low) |

## Job Management

### Wrapper Jobs (markdown, screenshot, extract)

```typescript
// Get job status
const job = await c.getMarkdownJob(jobId);
const job = await c.getScreenshotJob(jobId);
const job = await c.getExtractJob(jobId);

// List jobs
const jobs = await c.listMarkdownJobs({ status: 'completed', limit: 10 });
const jobs = await c.listScreenshotJobs({ limit: 5 });
const jobs = await c.listExtractJobs();

// Cancel
await c.cancelMarkdownJob(jobId);
await c.cancelScreenshotJob(jobId);
await c.cancelExtractJob(jobId);
```

### Crawl Jobs (from runMany / deepCrawl)

```typescript
const jobs = await c.listJobs({ status: 'completed', limit: 10 });
const job = await c.getJob(jobId);
const job = await c.waitJob(jobId, { pollInterval: 2, timeout: 300 });
await c.cancelJob(jobId);

// Download results as ZIP
const url = await c.downloadUrl(jobId);
```

## Power User: Config Passthrough

Every wrapper method accepts optional `crawlerConfig` and `browserConfig` for full control over the underlying crawl engine.

```typescript
const res = await c.markdown('https://spa-app.com', {
  crawlerConfig: {
    waitFor: 'css:.content-loaded',
    jsCode: 'window.scrollTo(0, document.body.scrollHeight)',
    scanFullPage: true,
    pageTimeout: 60000,
  },
  browserConfig: {
    viewportWidth: 1920,
    viewportHeight: 1080,
    headers: { 'Accept-Language': 'en-US' },
  },
  proxy: { mode: 'residential', country: 'US' },
});
```

This works on `markdown`, `screenshot`, `extract`, and their batch variants. The `map` method accepts `proxy` only (no browser needed).

## Full Power Mode

For advanced use cases -- custom extraction strategies, full HTML access, media/link parsing, session management -- use the core `run()` / `runMany()` / `deepCrawl()` methods directly.

### Single Crawl

```typescript
const result = await c.run('https://example.com', {
  config: { screenshot: true, excludeExternalLinks: true },
  browserConfig: { viewportWidth: 1920 },
  proxy: 'datacenter',
});
console.log(result.markdown?.rawMarkdown);
console.log(result.screenshot); // base64
console.log(result.links);
console.log(result.media);
```

### Batch Crawl

```typescript
const job = await c.runMany(
  ['https://a.com', 'https://b.com'],
  { wait: true, priority: 1 },
);
```

### Deep Crawl

```typescript
const result = await c.deepCrawl('https://docs.example.com', {
  strategy: 'bfs',
  maxDepth: 3,
  maxUrls: 200,
  wait: true,
});
```

### Scan (URL Discovery Only)

```typescript
const scan = await c.scan('https://example.com', { maxUrls: 100 });
console.log(`Found ${scan.totalUrls} URLs across ${scan.hostsFound} hosts`);
```

## Error Handling

```typescript
import {
  CloudError,
  AuthenticationError,
  RateLimitError,
  QuotaExceededError,
  NotFoundError,
  ValidationError,
  TimeoutError,
} from 'crawl4ai-cloud';

try {
  const res = await c.markdown('https://example.com');
} catch (err) {
  if (err instanceof AuthenticationError) {
    console.log('Bad API key');
  } else if (err instanceof RateLimitError) {
    console.log(`Rate limited, retry after ${err.retryAfter}s`);
  } else if (err instanceof QuotaExceededError) {
    console.log('Credits exhausted');
  } else if (err instanceof TimeoutError) {
    console.log('Request timed out');
  } else if (err instanceof CloudError) {
    console.log(`API error ${err.statusCode}: ${err.message}`);
  }
}
```

## Environment Variables

```bash
export CRAWL4AI_API_KEY=sk_live_...
```

```typescript
// API key auto-loaded from CRAWL4AI_API_KEY
const c = new AsyncWebCrawler();
```

## Links

- [Cloud Dashboard](https://api.crawl4ai.com) -- Sign up and get your API key
- [API Docs](https://api.crawl4ai.com/docs) -- Full REST API reference
- [Open Source](https://github.com/unclecode/crawl4ai) -- Self-hosted crawler
- [Discord](https://discord.gg/jP8KfhDhyN) -- Community and support

## License

Apache 2.0
