/**
 * E2E tests for the 0.8.1 multi-URL fan-out + url_statuses[] flow (Node.js).
 *
 * Covers:
 * - Wrapper async GET responses parse `urlStatuses` and `downloadUrl`.
 * - getPerUrlResult fetches one URL's CrawlResult (recipe-agnostic).
 * - wait=true hydrates job.results from per-URL S3 (auto-hydrate path).
 * - Single-URL submits leave urlStatuses as undefined.
 */

import { AsyncWebCrawler, UrlStatus, WrapperJob } from '../src';
import { wrapperJobFromDict } from '../src/models';

const API_KEY = process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const BASE_URL = process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';


describe('WrapperJob shape (unit)', () => {
  it('parses url_statuses + download_url', () => {
    const data = {
      job_id: 'job_abc',
      status: 'completed',
      progress: { total: 2, completed: 2, failed: 0 },
      progress_percent: 100,
      url_statuses: [
        { index: 0, url: 'https://a.com', status: 'done', duration_ms: 100, error: null },
        { index: 1, url: 'https://b.com', status: 'failed', duration_ms: 500, error: 'timeout' },
      ],
      download_url: 'https://...zip',
      created_at: '2026-05-03T00:00:00Z',
    };
    const job = wrapperJobFromDict(data);
    expect(job.urlStatuses).toBeDefined();
    expect(job.urlStatuses!.length).toBe(2);
    expect(job.urlStatuses![0].status).toBe('done');
    expect(job.urlStatuses![0].durationMs).toBe(100);
    expect(job.urlStatuses![1].error).toBe('timeout');
    expect(job.downloadUrl).toBe('https://...zip');
    expect(job.results).toBeUndefined();
  });

  it('leaves urlStatuses undefined for single-URL responses', () => {
    const data = {
      job_id: 'job_single',
      status: 'completed',
      progress: { total: 1, completed: 1, failed: 0 },
      created_at: '2026-05-03T00:00:00Z',
    };
    const job = wrapperJobFromDict(data);
    expect(job.urlStatuses).toBeUndefined();
    expect(job.results).toBeUndefined();
  });
});


describe('url_statuses E2E (stage)', () => {
  jest.setTimeout(120_000);

  it('scrapeMany wait=true hydrates results from per-URL S3', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
    const job = await crawler.scrapeMany(
      ['https://example.com', 'https://example.org'],
      { strategy: 'http', wait: true, timeout: 60 },
    );
    expect(['completed', 'partial']).toContain(job.status);
    expect(job.urlStatuses).toBeDefined();
    expect(job.urlStatuses!.length).toBe(2);
    expect(job.results).toBeDefined();
    expect(job.results!.length).toBe(2);
  });

  it('extractMany wait=true hydrates extracted_content', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
    const job = await crawler.extractMany('https://example.com', {
      extraUrls: ['https://example.org'],
      strategy: 'http',
      query: 'page title',
      method: 'auto',
      wait: true,
      timeout: 120,
    });
    expect(['completed', 'partial']).toContain(job.status);
    expect(job.urlStatuses).toBeDefined();
    expect(job.urlStatuses!.length).toBe(2);
    expect(job.results).toBeDefined();
    expect(job.results!.length).toBe(2);
  });

  it('getPerUrlResult is recipe-agnostic', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
    const job = await crawler.scrapeMany(
      ['https://example.com', 'https://example.org'],
      { strategy: 'http', wait: true, timeout: 60 },
    );
    const result = await crawler.getPerUrlResult(job.jobId, 0);
    expect(result.url).toBeDefined();
    if (result.success) {
      expect(result.markdown).toBeDefined();
    }
  });

  it('single-URL submit returns urlStatuses=undefined', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
    const job = await crawler.scrapeMany(
      ['https://example.com'],
      { strategy: 'http', wait: true, timeout: 60 },
    );
    expect(['completed', 'partial']).toContain(job.status);
    expect(job.urlStatuses).toBeUndefined();
    expect(job.results).toBeUndefined();
  });
});
