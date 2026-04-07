/**
 * E2E Adversarial Tests for Wrapper API SDK Methods (Node.js).
 * Real HTTP tests against stage.crawl4ai.com. No mocks.
 */

import {
  AsyncWebCrawler,
  MarkdownResponse,
  ScreenshotResponse,
  ExtractResponse,
  MapResponse,
  SiteCrawlResponse,
  WrapperJob,
  AuthenticationError,
  NotFoundError,
} from '../src';

const API_KEY = process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const FREE_KEY = process.env.CRAWL4AI_FREE_KEY ||
  'sk_live_oA8wznhyeNNCjhoSwVrEPAJjkQodQSwPq_stsg-gL3c';
const BASE_URL = process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

let crawler: AsyncWebCrawler;

beforeAll(() => {
  crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
});

afterAll(async () => {
  await crawler.close();
});

// =============================================================================
// MARKDOWN
// =============================================================================

describe('Markdown', () => {
  test('basic', async () => {
    const md = await crawler.markdown('https://example.com', { strategy: 'http' });
    expect(md.success).toBe(true);
    expect(md.url).toBe('https://example.com');
    expect(md.markdown!.length).toBeGreaterThan(50);
  }, 30000);

  test('fit markdown', async () => {
    const md = await crawler.markdown('https://httpbin.org/html', { strategy: 'http', fit: true });
    expect(md.success).toBe(true);
    expect(md.fitMarkdown).toBeDefined();
  }, 30000);

  test('include fields', async () => {
    const md = await crawler.markdown('https://books.toscrape.com', {
      strategy: 'http', include: ['links', 'media', 'metadata'],
    });
    expect(md.success).toBe(true);
    expect(md.links).toBeDefined();
    expect(md.media).toBeDefined();
    expect(md.metadata).toBeDefined();
  }, 60000);

  test('credits returned', async () => {
    const md = await crawler.markdown('https://example.com', { strategy: 'http' });
    expect(md.usage).toBeDefined();
    expect(md.usage!.creditsUsed).toBeGreaterThan(0);
  }, 30000);

  test('crawler_config passthrough', async () => {
    const md = await crawler.markdown('https://books.toscrape.com', {
      strategy: 'browser',
      crawlerConfig: { css_selector: 'article.product_pod', wait_until: 'domcontentloaded' },
    });
    expect(md.success).toBe(true);
  }, 60000);
});

// =============================================================================
// SCREENSHOT
// =============================================================================

describe('Screenshot', () => {
  test('basic', async () => {
    const ss = await crawler.screenshot('https://example.com');
    expect(ss.success).toBe(true);
    expect(ss.screenshot!.length).toBeGreaterThan(1000);
  }, 60000);

  test('pdf', async () => {
    const ss = await crawler.screenshot('https://example.com', { pdf: true });
    expect(ss.success).toBe(true);
    expect(ss.pdf!.length).toBeGreaterThan(1000);
  }, 60000);

  test('viewport only', async () => {
    const ss = await crawler.screenshot('https://example.com', { fullPage: false });
    expect(ss.success).toBe(true);
  }, 60000);
});

// =============================================================================
// EXTRACT
// =============================================================================

describe('Extract', () => {
  test('auto', async () => {
    const data = await crawler.extract('https://books.toscrape.com', {
      query: 'extract all books with title and price',
    });
    expect(data.success).toBe(true);
    expect(data.data!.length).toBeGreaterThan(0);
    expect(['css_schema', 'llm']).toContain(data.methodUsed);
  }, 120000);

  test('llm method', async () => {
    const data = await crawler.extract('https://example.com', {
      method: 'llm', query: 'what is this page about',
    });
    expect(data.success).toBe(true);
    expect(data.methodUsed).toBe('llm');
  }, 120000);
});

// =============================================================================
// MAP
// =============================================================================

describe('Map', () => {
  test('basic', async () => {
    const result = await crawler.map('https://crawl4ai.com', { maxUrls: 10 });
    expect(result.success).toBe(true);
    expect(result.totalUrls).toBeGreaterThan(0);
    expect(result.domain).toBe('crawl4ai.com');
  }, 60000);

  test('with query scoring', async () => {
    const result = await crawler.map('https://docs.crawl4ai.com', {
      query: 'extraction', maxUrls: 5, scoreThreshold: 0.1,
    });
    expect(result.success).toBe(true);
  }, 60000);
});

// =============================================================================
// SITE CRAWL
// =============================================================================

describe('Site Crawl', () => {
  test('basic', async () => {
    const result = await crawler.crawlSite('https://books.toscrape.com', {
      maxPages: 3, strategy: 'http',
    });
    expect(result.jobId).toBeDefined();
    expect(result.strategy).toBe('map');
  }, 30000);
});

// =============================================================================
// ASYNC LIFECYCLE
// =============================================================================

describe('Async Lifecycle', () => {
  test('markdown many with wait', async () => {
    const job = await crawler.markdownMany(
      ['https://example.com', 'https://httpbin.org/html'],
      { strategy: 'http', wait: true, timeout: 60 },
    );
    expect(['completed', 'partial']).toContain(job.status);
    expect(job.progress!.completed).toBe(2);
  }, 90000);

  test('poll and list', async () => {
    const job = await crawler.markdownMany(['https://example.com'], { strategy: 'http' });
    expect(job.jobId).toBeDefined();

    // Wait then poll
    await new Promise(r => setTimeout(r, 5000));
    const status = await crawler.getMarkdownJob(job.jobId);
    expect(status.jobId).toBe(job.jobId);

    // List
    const jobs = await crawler.listMarkdownJobs({ limit: 5 });
    expect(jobs.length).toBeGreaterThan(0);
  }, 30000);
});

// =============================================================================
// JOB CANCEL
// =============================================================================

describe('Job Cancel', () => {
  test('cancel markdown job', async () => {
    const job = await crawler.markdownMany(
      ['https://httpbin.org/delay/10', 'https://httpbin.org/delay/10', 'https://httpbin.org/delay/10'],
      { strategy: 'http' },
    );
    await new Promise(r => setTimeout(r, 1000));
    const cancelled = await crawler.cancelMarkdownJob(job.jobId);
    expect(cancelled).toBe(true);

    const status = await crawler.getMarkdownJob(job.jobId);
    expect(status.status).toBe('cancelled');
  }, 30000);
});

// =============================================================================
// NAMESPACE ISOLATION
// =============================================================================

describe('Namespace Isolation', () => {
  test('cross-namespace 404', async () => {
    const job = await crawler.markdownMany(['https://example.com'], { strategy: 'http' });
    await new Promise(r => setTimeout(r, 2000));

    await expect(crawler.getScreenshotJob(job.jobId)).rejects.toThrow();
  }, 15000);
});

// =============================================================================
// ERROR CASES
// =============================================================================

describe('Errors', () => {
  test('bad auth', async () => {
    const bad = new AsyncWebCrawler({ apiKey: 'sk_live_bad_key_000', baseUrl: BASE_URL });
    await expect(bad.markdown('https://example.com')).rejects.toThrow();
    await bad.close();
  }, 15000);

  test('nonexistent job', async () => {
    await expect(crawler.getMarkdownJob('job_doesnotexist000000')).rejects.toThrow();
  }, 15000);

  test('extract auto batch rejected', async () => {
    await expect(
      crawler.extractMany(['https://example.com'], { method: 'auto' as any })
    ).rejects.toThrow('AUTO');
  }, 5000);
});

// =============================================================================
// ADVERSARIAL
// =============================================================================

describe('Adversarial', () => {
  test('sql in query', async () => {
    const data = await crawler.extract('https://example.com', {
      method: 'llm', query: "'; DROP TABLE users; --",
    });
    expect(data).toBeDefined();
  }, 120000);

  test('xss in config', async () => {
    const md = await crawler.markdown('https://example.com', {
      strategy: 'http',
      crawlerConfig: { css_selector: '<script>alert(1)</script>' },
    });
    expect(md).toBeDefined();
  }, 30000);

  test('unicode url', async () => {
    const md = await crawler.markdown('https://example.com/\u00e9\u00e8', { strategy: 'http' });
    expect(md).toBeDefined();
  }, 30000);
});

// =============================================================================
// FREE TIER
// =============================================================================

describe('Free Tier', () => {
  test('markdown', async () => {
    const free = new AsyncWebCrawler({ apiKey: FREE_KEY, baseUrl: BASE_URL });
    const md = await free.markdown('https://example.com', { strategy: 'http' });
    expect(md.success).toBe(true);
    expect(md.usage!.creditsUsed).toBeGreaterThan(0);
    await free.close();
  }, 30000);
});
