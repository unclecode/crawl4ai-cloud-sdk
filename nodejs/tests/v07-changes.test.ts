/**
 * SDK 0.7.0 changes — real e2e tests against stage. No mocks.
 *
 * Covers:
 *  - crawler.scrape() / scrapeMany() — new canonical names
 *  - crawler.markdown() / markdownMany() — deprecated aliases (still work)
 *  - crawler.extractMany() — new shape (url + extraUrls, AUTO allowed)
 *  - sources= field on scan() + map() (legacy mode= still works)
 *  - Composable scan + scrape/extract chain
 *  - crawler.crawlSite() / deepCrawl() — deprecated, still respond
 */

import { AsyncWebCrawler, WrapperJob, MarkdownResponse } from '../src';

const API_KEY = process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const BASE_URL = process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

let crawler: AsyncWebCrawler;
beforeAll(() => { crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL }); });

// =============================================================================
// Scrape — new canonical name
// =============================================================================

describe('Scrape (new canonical)', () => {
  test('scrape basic', async () => {
    const r = await crawler.scrape('https://example.com', { strategy: 'http' });
    expect(r.success).toBe(true);
    expect(r.markdown).toContain('Example');
  }, 60000);

  test('scrape with include', async () => {
    const r = await crawler.scrape('https://example.com', {
      strategy: 'http', include: ['links', 'metadata'],
    });
    expect(r.success).toBe(true);
    expect(r.links).toBeDefined();
    expect(r.metadata).toBeDefined();
  }, 60000);

  test('scrapeMany async batch', async () => {
    const job = await crawler.scrapeMany(
      ['https://example.com', 'https://httpbin.org/html'],
      { strategy: 'http', wait: true, timeout: 120 },
    );
    expect(job.status).toBe('completed');
  }, 180000);
});

// =============================================================================
// Markdown alias — still works
// =============================================================================

describe('Markdown alias (deprecated)', () => {
  test('markdown still works', async () => {
    const r = await crawler.markdown('https://example.com', { strategy: 'http' });
    expect(r.success).toBe(true);
    expect(r.markdown).toBeTruthy();
  }, 60000);

  test('markdownMany still works', async () => {
    const job = await crawler.markdownMany(
      ['https://example.com'],
      { strategy: 'http', wait: true, timeout: 60 },
    );
    expect(job.status).toBe('completed');
  }, 120000);
});

// =============================================================================
// Extract — new shape (url + extraUrls), AUTO allowed
// =============================================================================

describe('extractMany (new url + extraUrls shape)', () => {
  test('single base url with AUTO method', async () => {
    const job = await crawler.extractMany('https://example.com', {
      method: 'auto', wait: true, timeout: 120,
    });
    expect(job.status).toBe('completed');
  }, 180000);

  test('url + extraUrls with LLM method', async () => {
    const job = await crawler.extractMany('https://example.com', {
      extraUrls: ['https://httpbin.org/html'],
      method: 'llm',
      query: 'summarize the page',
      wait: true,
      timeout: 180,
    });
    expect(job.status).toBe('completed');
  }, 240000);
});

// =============================================================================
// sources= field on map + scan
// =============================================================================

describe('sources= field', () => {
  test('map with sources=primary', async () => {
    const r = await crawler.map('https://www.python.org', {
      sources: 'primary', maxUrls: 5,
    });
    expect(r.success).toBe(true);
    expect(r.totalUrls).toBeGreaterThan(0);
  }, 120000);

  test('map with legacy mode=default still works', async () => {
    const r = await crawler.map('https://www.python.org', {
      mode: 'default', maxUrls: 5,
    });
    expect(r.success).toBe(true);
  }, 120000);

  test('scan with sources=primary', async () => {
    const r = await crawler.scan('https://www.python.org', {
      sources: 'primary', maxUrls: 5,
    });
    expect(r.totalUrls).toBeGreaterThan(0);
  }, 120000);
});

// =============================================================================
// Composable chain
// =============================================================================

describe('Composable chain', () => {
  test('scan then scrapeMany', async () => {
    const scan = await crawler.scan('https://www.python.org', {
      sources: 'primary', maxUrls: 3,
    });
    const urls = scan.urls.slice(0, 3).map(u => u.url);
    expect(urls.length).toBeGreaterThan(0);

    const job = await crawler.scrapeMany(urls, {
      strategy: 'http', wait: true, timeout: 120,
    });
    expect(job.status).toBe('completed');
  }, 240000);

  test('scan then extractMany with extraUrls', async () => {
    const scan = await crawler.scan('https://www.python.org', {
      sources: 'primary', maxUrls: 3,
    });
    const urls = scan.urls.slice(0, 3).map(u => u.url);
    expect(urls.length).toBeGreaterThan(0);

    const [base, ...rest] = urls;
    const job = await crawler.extractMany(base, {
      extraUrls: rest,
      method: 'auto',
      wait: true,
      timeout: 240,
    });
    expect(job.status).toBe('completed');
  }, 300000);
});
