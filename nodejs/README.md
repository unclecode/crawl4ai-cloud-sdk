# Crawl4AI Cloud SDK for Node.js

Lightweight Node.js/TypeScript SDK for [Crawl4AI Cloud](https://api.crawl4ai.com). Mirrors the OSS API exactly.

> **Note:** This SDK is for **Crawl4AI Cloud** (api.crawl4ai.com), the managed cloud service. For the self-hosted open-source version, see [github.com/unclecode/crawl4ai](https://github.com/unclecode/crawl4ai).

[![npm version](https://badge.fury.io/js/crawl4ai-cloud.svg)](https://badge.fury.io/js/crawl4ai-cloud)

## Installation

```bash
# From npm (coming soon)
npm install crawl4ai-cloud

# From GitHub (available now)
npm install github:unclecode/crawl4ai-cloud-sdk#main
```

## Get Your API Key

1. Go to [api.crawl4ai.com](https://api.crawl4ai.com)
2. Sign up and get your API key

## Quick Start

```typescript
import { AsyncWebCrawler } from 'crawl4ai-cloud';

const crawler = new AsyncWebCrawler({ apiKey: 'sk_live_...' });

const result = await crawler.run('https://example.com');
console.log(result.markdown?.rawMarkdown);

await crawler.close();
```

## Features

### Single URL Crawl

```typescript
const result = await crawler.run('https://example.com');
console.log(result.success);
console.log(result.markdown?.rawMarkdown);
console.log(result.html);
```

### Batch Crawl

```typescript
const urls = ['https://example.com', 'https://httpbin.org/html'];

// Wait for results
const results = await crawler.runMany(urls, { wait: true });
for (const r of results as CrawlResult[]) {
  console.log(`${r.url}: ${r.success}`);
}

// Fire and forget (returns job)
const job = await crawler.runMany(urls, { wait: false });
console.log(`Job ID: ${(job as CrawlJob).id}`);
```

### Configuration

```typescript
import { AsyncWebCrawler, CrawlerRunConfig, BrowserConfig } from 'crawl4ai-cloud';

const config: CrawlerRunConfig = {
  wordCountThreshold: 10,
  excludeExternalLinks: true,
  screenshot: true,
};

const browserConfig: BrowserConfig = {
  viewportWidth: 1920,
  viewportHeight: 1080,
};

const result = await crawler.run('https://example.com', {
  config,
  browserConfig,
});
```

### Proxy Support

```typescript
// Shorthand
const result = await crawler.run(url, { proxy: 'datacenter' });
const result = await crawler.run(url, { proxy: 'residential' });

// Full config
const result = await crawler.run(url, {
  proxy: { mode: 'residential', country: 'US' }
});
```

### Deep Crawl

```typescript
const result = await crawler.deepCrawl('https://docs.example.com', {
  strategy: 'bfs',
  maxDepth: 2,
  maxUrls: 50,
  wait: true,
});
```

### Job Management

```typescript
// List jobs
const jobs = await crawler.listJobs({ status: 'completed', limit: 10 });

// Get job status
const job = await crawler.getJob(jobId);

// Wait for job
const completedJob = await crawler.waitJob(jobId, {
  pollInterval: 2.0,
  timeout: 300,
});

// Cancel job
await crawler.cancelJob(jobId);
```

## OSS Compatibility

The SDK provides `arun()` and `arunMany()` aliases for seamless migration:

```typescript
// These are equivalent
const result = await crawler.run(url);
const result = await crawler.arun(url);

const results = await crawler.runMany(urls);
const results = await crawler.arunMany(urls);
```

## Environment Variables

```bash
export CRAWL4AI_API_KEY=sk_live_...
```

```typescript
// API key auto-loaded from environment
const crawler = new AsyncWebCrawler({});
```

## Error Handling

```typescript
import {
  CloudError,
  AuthenticationError,
  RateLimitError,
  QuotaExceededError,
  NotFoundError,
} from 'crawl4ai-cloud';

try {
  const result = await crawler.run(url);
} catch (error) {
  if (error instanceof AuthenticationError) {
    console.log('Invalid API key');
  } else if (error instanceof RateLimitError) {
    console.log(`Rate limited. Retry after ${error.retryAfter}s`);
  } else if (error instanceof QuotaExceededError) {
    console.log('Quota exceeded');
  }
}
```

## TypeScript Support

Full TypeScript support with exported types:

```typescript
import type {
  CrawlResult,
  CrawlJob,
  MarkdownResult,
  CrawlerRunConfig,
  BrowserConfig,
  ProxyConfig,
} from 'crawl4ai-cloud';
```

## Links

- [Cloud Dashboard](https://api.crawl4ai.com) - Sign up & get your API key
- [Cloud API Docs](https://api.crawl4ai.com/docs) - Full API reference
- [OSS Repository](https://github.com/unclecode/crawl4ai) - Self-hosted option
- [Discord](https://discord.gg/jP8KfhDhyN) - Community & support

## License

Apache 2.0
