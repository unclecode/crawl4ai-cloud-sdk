# Crawl4AI Cloud SDK for Node.js

One-line web scraping, screenshots, and structured extraction — powered by [Crawl4AI Cloud](https://api.crawl4ai.com).

[![npm version](https://badge.fury.io/js/crawl4ai-cloud.svg)](https://badge.fury.io/js/crawl4ai-cloud) v0.7.0

## What's new in 0.7.0

- **`scrape()` + `scrapeMany()`** are the canonical methods (replacing `markdown` / `markdownMany`). Same shape, same response — they hit the new `/v1/scrape(/async)` endpoints.
- `markdown()` / `markdownMany()` are kept as **deprecated aliases** that route to `/v1/scrape`. Removed in 0.8.0.
- **`extractMany()`** signature fixed: now takes `extractMany(url, { extraUrls: […], … })` (was `extractMany(urls[], …)`, which the server rejected). `method: 'auto'` is now allowed for batch.
- **`sources`** field added to `scan()` and `map()` (`'primary'` | `'extended'`). Replaces `mode`, which is kept as a deprecated alias.
- **`crawlSite()` and `deepCrawl()`** target deprecated endpoints. They still work and emit a `DeprecationWarning`. **Migrate to the composable scan + scrape/extract chain** — see below.

## Install

```bash
npm install crawl4ai-cloud
```

## Quick Start

```typescript
import { AsyncWebCrawler } from 'crawl4ai-cloud';

const c = new AsyncWebCrawler({ apiKey: 'sk_live_...' });

// Scrape a page → clean markdown (+ optional links/media/metadata/tables)
const page = await c.scrape('https://example.com');
console.log(page.markdown);

// Take a full-page screenshot
const shot = await c.screenshot('https://example.com');
// shot.screenshot is base64 PNG

// Extract structured data — AUTO picks css_schema vs llm.
// wait: true auto-hydrates job.results from per-URL S3.
const job = await c.extractMany('https://news.ycombinator.com', {
  method: 'auto',
  query: 'get each story title, URL, and points',
  wait: true,
});
for (const r of job.results ?? []) console.log(r.extractedContent);

// Discover all URLs on a domain
const sitemap = await c.map('https://docs.python.org', { sources: 'primary' });
sitemap.urls.slice(0, 10).forEach(u => console.log(u.url));
```

## Crawl a whole site — composable two-step

There's no single bundled "crawl this site" method anymore. Compose it:

```typescript
// Step 1 — discover URLs
const scan = await c.scan('https://books.toscrape.com', {
  criteria: 'all book detail pages',
  maxUrls: 20,
});
const urls = scan.urls.map(u => u.url);

// Step 2A — pipe to scrapeMany for markdown
const mdJob = await c.scrapeMany(urls, { strategy: 'http', wait: true });

// Step 2B — OR pipe to extractMany for structured fields. The base url is
// the schema TEMPLATE — schema is generated once, then re-applied across
// extraUrls for free in css_schema mode (10-100× cheaper than per-page LLM).
const [base, ...rest] = urls;
const exJob = await c.extractMany(base, {
  extraUrls: rest,
  method: 'auto',
  query: 'book title, price, rating',
  wait: true,
});
for (const r of exJob.results ?? []) console.log(r.extractedContent);
```

## Multi-URL fan-out — what happens under the hood

`scrapeMany(urls)`, `screenshotMany(urls)`, and `extractMany(url, { extraUrls })` all decompose into N independent child jobs that distribute across the worker pool. Throughput scales with pool size, not with how many URLs you submitted.

```typescript
const job = await c.scrapeMany(urls, { wait: false });             // returns immediately
// Poll to inspect per-URL state without downloading the data:
const status = await c.getScrapeJob(job.jobId);                    // WrapperJob
for (const u of status.urlStatuses ?? []) {
  console.log(u.index, u.url, u.status, u.durationMs, u.error);
}

// Fetch one URL's full result on demand (recipe-agnostic — works for any wrapper)
const result = await c.getPerUrlResult(job.jobId, 0);              // CrawlResult
console.log(result.markdown);
```

`wait: true` does this for you: poll → wait until terminal → call `getPerUrlResult` for each URL in parallel → populate `job.results`. Failed URLs become `CrawlResult` stubs (`success=false` + `errorMessage`) so `job.results.length === job.urlStatuses.length`.

## Wrapper API Reference

| Method | What it does | Endpoint |
|---|---|---|
| `scrape(url)` | Page → clean markdown + optional `include: ['links'\|'media'\|'metadata'\|'tables']` | `POST /v1/scrape` |
| `scrapeMany(urls)` | Async batch scrape (≤100 URLs) | `POST /v1/scrape/async` |
| `screenshot(url)` | Base64 PNG (and optional PDF) | `POST /v1/screenshot` |
| `screenshotMany(urls)` | Async batch screenshot | `POST /v1/screenshot/async` |
| `extract(url, { query })` | Sync extract — single URL | `POST /v1/extract` |
| `extractMany(url, { extraUrls })` | Async extract — base + followers, schema reused for free in css_schema mode | `POST /v1/extract/async` |
| `scan(url, { sources, criteria })` | URL discovery — sources=primary/extended, optional AI criteria | `POST /v1/scan` |
| `map(url, { sources })` | Simpler URL discovery (no criteria) | `POST /v1/map` |
| `enrich({ query })` | Multi-entity table from a brief, list of entities, or list of URLs | `POST /v1/enrich/async` |

### Deprecated (still work, emit warnings)

| Method | Migration |
|---|---|
| `markdown(url)` / `markdownMany(urls)` | Use `scrape` / `scrapeMany`. |
| `crawlSite(url, { criteria, extract })` | Use `scan` + `extractMany`. See "Crawl a whole site" above. |
| `deepCrawl(url, …)` | Use `scan(url, { scan: { mode: 'deep' } })` + the chain above. |
| `scan(url, { mode })` / `map(url, { mode })` | Use `sources: 'primary'` or `sources: 'extended'`. |

### Enrich v2 (build a data table)

Multi-phase enrichment. Give a brief, a list of entities, or a list of URLs. The job walks through `planning → resolving_urls → extracting → merging`, with optional human-review pauses at `plan_ready` and `urls_ready`.

```typescript
// 1. Agent one-shot — give a brief, get a table back
const result = await crawler.enrich({
  query: 'licensed nurseries in North York Toronto with extended hours',
  country: 'ca',
  topKPerEntity: 3,
});
for (const row of result.phaseData.rows ?? []) {
  console.log(row.inputKey, row.fields);
  for (const [f, c] of Object.entries(row.certainty)) {
    console.log(`  ${f}: ${c.toFixed(2)}  (from ${row.sources[f].url})`);
  }
}
console.log(`crawls=${result.usage.crawls}, llm=${JSON.stringify(result.usage.llmTotals)}`);

// 2. Pre-resolved URLs — skip planning + URL resolution
const r2 = await crawler.enrich({
  urls: ['https://example.com/a', 'https://example.com/b'],
  features: ['price', 'hours'],   // string shortcut for [{name: 'price'}, ...]
});

// 3. Human review flow — pause for editing the plan
const job = await crawler.enrich({
  query: 'best Italian restaurants in Brooklyn',
  country: 'us',
  autoConfirmPlan: false,
  autoConfirmUrls: false,
  wait: false,
});

let paused = await crawler.waitEnrichJob(job.jobId, { until: 'plan_ready' });
console.log(paused.phaseData.plan?.entities, paused.phaseData.plan?.features);

await crawler.resumeEnrichJob(job.jobId, {
  entities: [{ name: "Lucali" }, { name: "Roberta's" }],
  features: [{ name: 'address' }, { name: 'hours' }],
});
paused = await crawler.waitEnrichJob(job.jobId, { until: 'urls_ready' });
await crawler.resumeEnrichJob(job.jobId);   // accept server's URL picks
const final = await crawler.waitEnrichJob(job.jobId);

// 4. Live progress via SSE
for await (const event of crawler.streamEnrichJob(job.jobId)) {
  if (event.type === 'phase')   console.log('→', event.status);
  if (event.type === 'row')     console.log('✓', event.row?.inputKey);
  if (event.type === 'complete') break;
}

// Job management
const jobs = await crawler.listEnrichJobs({ limit: 5 });
await crawler.cancelEnrichJob(job.jobId);
```

**Vocabulary:** **entities** are row identifiers; **criteria** are search-side filters; **features** are extraction columns. **Per-purpose usage** is reported in `result.usage.llmTokensByPurpose` with five buckets: `plan_intent`, `url_plan`, `paywall_classify`, `extract`, `merge_tiebreak` (only buckets that ran appear).

## Wrapper API Reference

| Method | Endpoint | Returns | Description |
|--------|----------|---------|-------------|
| `markdown(url)` | `POST /v1/markdown` | `MarkdownResponse` | Clean markdown from any page |
| `screenshot(url)` | `POST /v1/screenshot` | `ScreenshotResponse` | Full-page screenshot (PNG) or PDF |
| `extract(url)` | `POST /v1/extract` | `ExtractResponse` | LLM or schema-based structured extraction |
| `map(url)` | `POST /v1/map` | `MapResponse` | Simple URL discovery (always sync) |
| `scan(url, {criteria})` | `POST /v1/scan` | `ScanResult` | **AI-assisted** URL discovery with plain-English criteria |
| `crawlSite(url, {criteria, extract})` | `POST /v1/crawl/site` | `SiteCrawlResponse` | **AI-assisted** whole-site crawl (always async) |
| `enrich(opts)` | `POST /v1/enrich/async` | `EnrichJobStatus` | Multi-phase enrichment from query / entities / URLs |

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

### Map Options (legacy, sync-only)

```typescript
const res = await c.map('https://example.com', {
  mode: 'deep',            // 'default' or 'deep' (DomainMapper source depth)
  maxUrls: 500,
  includeSubdomains: true,
  query: 'blog posts',     // relevance scoring
  scoreThreshold: 0.5,
});
```

### Scan Options (AI-assisted)

```typescript
// AI picks everything from criteria
const result = await c.scan('https://docs.crawl4ai.com', {
  criteria: 'API reference pages',
  maxUrls: 50,
});

// Explicit overrides on top of LLM output
const result = await c.scan('https://example.com', {
  criteria: 'product pages',
  scan: {
    mode: 'auto',          // 'auto' | 'map' | 'deep'
    patterns: ['*/p/*'],
    includeSubdomains: false,
  },
});

// Async deep scan (waits for completion)
const result = await c.scan('https://directory.example.com', {
  criteria: 'company profile pages',
  scan: { mode: 'deep', maxDepth: 3 },
  wait: true,              // poll until done
  pollInterval: 3,         // seconds
});
```

### Site Crawl Options (AI-assisted flagship)

```typescript
const job = await c.crawlSite('https://books.toscrape.com', {
  // AI-assisted fields (new in v0.4.0)
  criteria: 'all book listing pages',
  scan: { mode: 'auto' },                        // optional explicit override
  extract: {
    query: 'book title, price, rating',
    jsonExample: { title: '...', price: '£0.00', rating: 0 },
    method: 'auto',                              // 'auto' | 'llm' | 'schema'
  },
  include: ['markdown', 'links'],                // drop 'markdown' to strip it
  // Standard fields
  maxPages: 50,
  strategy: 'http',                              // 'browser' or 'http'
  fit: true,
  // Legacy fields (still supported, prefer the above)
  discovery: 'bfs',                              // 'map' (default) / 'bfs' / 'dfs' / 'best_first'
  pattern: '/docs/*',
  maxDepth: 3,
  // Waiting
  wait: true,                                    // poll until done
  pollInterval: 5,                               // seconds
  timeout: 300,                                  // seconds
});

// job.generatedConfig — LLM's scan + extract decisions
// job.extractionMethodUsed — "css_schema" | "llm"
// job.schemaUsed — generated CSS schema (reusable!)
```

**Drop markdown with `include`**: pass `include: ['links', 'media']` (without `'markdown'`) and the worker strips markdown from every result — saves bandwidth for extract-only crawls.

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

### Scan Jobs (AI-assisted deep scans)

```typescript
// Poll status + URLs discovered so far
const job = await c.getScanJob(jobId);
console.log(`Status: ${job.status}, found: ${job.totalUrls}`);

// Cancel a running deep scan (preserves partial results)
await c.cancelScanJob(jobId);
```

### Site Crawl Jobs (unified scan + crawl polling)

```typescript
const status = await c.getSiteCrawlJob(jobId);
console.log(`${status.phase}: ${status.progress.urlsCrawled}/${status.progress.total}`);
// status.phase walks: "scan" -> "crawl" -> "done"
// status.downloadUrl is set when phase === "done" and status === "completed"
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
