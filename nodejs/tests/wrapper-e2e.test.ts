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
  SiteCrawlJobStatus,
  ScanResult,
  ScanJobStatus,
  SiteScanConfig,
  SiteExtractConfig,
  GeneratedConfig,
  WrapperJob,
  AuthenticationError,
  NotFoundError,
  isScanJobComplete,
  isSiteCrawlJobComplete,
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

  test('with criteria', async () => {
    // AI-assisted: criteria triggers LLM config generation
    const result = await crawler.crawlSite('https://books.toscrape.com', {
      criteria: 'book listing pages',
      maxPages: 3,
      strategy: 'http',
    });
    expect(result.jobId).toBeDefined();
    expect(result.generatedConfig).toBeDefined();
    expect(result.generatedConfig!.scan).toBeDefined();
    expect(result.generatedConfig!.reasoning).toBeTruthy();
    expect(result.generatedConfig!.fallback).toBe(false);
  }, 60000);

  test('with criteria and extract (flagship)', async () => {
    // The big one: criteria + extract -> schema generated from sample URL
    const result = await crawler.crawlSite('https://books.toscrape.com', {
      criteria: 'book listing pages',
      maxPages: 3,
      strategy: 'http',
      extract: {
        query: 'book title and price',
        jsonExample: { title: '...', price: '£0.00' },
        method: 'auto',
      },
    });
    expect(result.jobId).toBeDefined();
    expect(result.generatedConfig).toBeDefined();
    expect(['css_schema', 'llm']).toContain(result.extractionMethodUsed);
    if (result.extractionMethodUsed === 'css_schema') {
      expect(result.schemaUsed).toBeDefined();
      expect(result.schemaUsed!.fields).toBeDefined();
    }
  }, 120000);

  test('scan config object accepted', async () => {
    const scanCfg: SiteScanConfig = {
      mode: 'map',
      patterns: ['*/catalogue/*'],
      scoreThreshold: 0.2,
    };
    const result = await crawler.crawlSite('https://books.toscrape.com', {
      maxPages: 3,
      strategy: 'http',
      scan: scanCfg,
    });
    expect(result.jobId).toBeDefined();
  }, 30000);

  test('unified polling with getSiteCrawlJob', async () => {
    const result = await crawler.crawlSite('https://books.toscrape.com', {
      criteria: 'book listings',
      maxPages: 3,
      strategy: 'http',
    });
    // Poll up to 5 times to see a phase transition
    for (let i = 0; i < 5; i++) {
      const status = await crawler.getSiteCrawlJob(result.jobId);
      expect(status.jobId).toBe(result.jobId);
      expect(['scan', 'crawl', 'done']).toContain(status.phase);
      expect(status.progress).toBeDefined();
      if (isSiteCrawlJobComplete(status)) break;
      await new Promise(r => setTimeout(r, 3000));
    }
  }, 60000);

  test('include without markdown strips it', async () => {
    // `include: ['links']` (no 'markdown') should trigger exclude_fields=['markdown']
    const result = await crawler.crawlSite('https://books.toscrape.com', {
      criteria: 'book listings',
      maxPages: 3,
      strategy: 'http',
      include: ['links'],
    });
    expect(result.jobId).toBeDefined();
  }, 60000);
});

// =============================================================================
// SCAN (AI-assisted)
// =============================================================================

describe('Scan', () => {
  test('basic legacy (no criteria)', async () => {
    const result = await crawler.scan('https://crawl4ai.com', { maxUrls: 10 });
    expect(result.success).toBe(true);
    expect(result.totalUrls).toBeGreaterThan(0);
    expect(result.domain).toBe('crawl4ai.com');
  }, 30000);

  test('with criteria', async () => {
    const result = await crawler.scan('https://docs.crawl4ai.com', {
      criteria: 'API reference and core documentation pages',
      maxUrls: 20,
    });
    expect(result.success).toBe(true);
    expect(['map', 'deep']).toContain(result.modeUsed);
    expect(result.generatedConfig).toBeDefined();
    expect(result.generatedConfig!.reasoning).toBeTruthy();
  }, 60000);

  test('with scan overrides', async () => {
    const result = await crawler.scan('https://docs.crawl4ai.com', {
      criteria: 'documentation pages',
      scan: { patterns: ['*/core/*'] },
      maxUrls: 10,
    });
    expect(result.success).toBe(true);
    expect(result.modeUsed).toBe('map');
  }, 60000);

  test('scan config object as option', async () => {
    const cfg: SiteScanConfig = { mode: 'map', patterns: ['*/docs/*'] };
    const result = await crawler.scan('https://docs.crawl4ai.com', {
      scan: cfg,
      maxUrls: 10,
    });
    expect(result.success).toBe(true);
  }, 30000);

  test('deep mode returns job and polls', async () => {
    const result = await crawler.scan('https://httpbin.org', {
      scan: { mode: 'deep', maxDepth: 1 },
      maxUrls: 5,
    });
    expect(result.jobId).toBeDefined();
    expect(result.modeUsed).toBe('deep');
    // Poll once to verify the endpoint responds
    const job = await crawler.getScanJob(result.jobId!);
    expect(job.jobId).toBe(result.jobId);
    // Cancel so we don't tie up worker slots
    const cancelled = await crawler.cancelScanJob(result.jobId!);
    expect(cancelled.jobId).toBe(result.jobId);
  }, 60000);
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

  test('extractMany now uses url + extraUrls (was urls=[]) — AUTO accepted as of 0.7.0', async () => {
    // Just confirm submission shape works; covered fully in v07-changes.test.ts.
    const job = await crawler.extractMany('https://example.com', {
      method: 'auto',
    });
    expect(job.jobId).toBeTruthy();
  }, 30000);
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
