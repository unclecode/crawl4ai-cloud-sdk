/**
 * Enrich API E2E tests -- runs against stage.crawl4ai.com. No mocks.
 *
 * Usage:
 *    npx jest tests/enrich-e2e.test.ts --verbose
 */

import {
  AsyncWebCrawler,
  EnrichJobStatus,
  EnrichRow,
  EnrichFieldSource,
  EnrichJobProgress,
  isEnrichJobComplete,
  isEnrichJobSuccessful,
  enrichJobStatusFromDict,
} from '../src';

const API_KEY =
  process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const BASE_URL =
  process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

let crawler: AsyncWebCrawler;

beforeAll(() => {
  crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
});

afterAll(async () => {
  await crawler.close();
});

// =============================================================================
// CORE ENRICHMENT
// =============================================================================

describe('Enrich - Happy Path', () => {
  test('basic enrich (1 URL, 2 fields, depth 0)', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com'],
      [
        { name: 'Company Name' },
        { name: 'Email', description: 'contact email' },
      ],
      { maxDepth: 0, enableSearch: false, strategy: 'browser', timeout: 120 },
    );

    expect(result.jobId).toBeTruthy();
    expect(isEnrichJobComplete(result)).toBe(true);
    expect(isEnrichJobSuccessful(result)).toBe(true);
    expect(result.rows).toBeDefined();
    expect(result.rows!.length).toBe(1);

    const row = result.rows![0];
    expect(row.url).toBe('https://kidocode.com');
    expect(row.fields['Company Name']).toBeTruthy();
    expect(['complete', 'partial']).toContain(row.status);
    expect(row.depthUsed).toBe(0);
  }, 180000);

  test('enrich with depth (1 URL, 3 fields, depth 1)', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com'],
      [
        { name: 'Company Name' },
        { name: 'Email', description: 'primary contact email' },
        { name: 'Phone', description: 'phone number' },
      ],
      { maxDepth: 1, maxLinks: 3, enableSearch: false, timeout: 120 },
    );

    expect(isEnrichJobComplete(result)).toBe(true);
    expect(result.rows!.length).toBe(1);

    const row = result.rows![0];
    expect(row.fields['Company Name']).toBeTruthy();
    // With depth 1, should find more fields
    const found = Object.values(row.fields).filter((v) => v).length;
    expect(found).toBeGreaterThanOrEqual(2);
  }, 180000);

  test('multiple URLs', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com', 'https://httpbin.org'],
      [{ name: 'Title', description: 'page or company title' }],
      { maxDepth: 0, enableSearch: false, timeout: 120 },
    );

    expect(isEnrichJobComplete(result)).toBe(true);
    expect(result.progress.total).toBe(2);
    expect(result.rows).toBeDefined();
    expect(result.rows!.length).toBe(2);

    const urls = new Set(result.rows!.map((r) => r.url));
    expect(
      urls.has('https://kidocode.com') || urls.has('https://httpbin.org'),
    ).toBe(true);
  }, 180000);
});

// =============================================================================
// SOURCE ATTRIBUTION
// =============================================================================

describe('Enrich - Source Attribution', () => {
  test('sources present for found fields', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com'],
      [{ name: 'Company Name' }, { name: 'Email' }],
      { maxDepth: 0, enableSearch: false, timeout: 120 },
    );

    const row = result.rows![0];
    for (const [fieldName, value] of Object.entries(row.fields)) {
      if (value) {
        expect(row.sources[fieldName]).toBeDefined();
        const src: EnrichFieldSource = row.sources[fieldName];
        expect(['direct', 'depth', 'search']).toContain(src.method);
        expect(src.url).toBeTruthy();
      }
    }
  }, 180000);
});

// =============================================================================
// JOB MANAGEMENT
// =============================================================================

describe('Enrich - Job Management', () => {
  test('fire and forget + manual poll', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com'],
      [{ name: 'Company Name' }],
      { maxDepth: 0, enableSearch: false, wait: false },
    );

    expect(result.jobId).toMatch(/^enr_/);
    expect(result.status).toBe('pending');

    // Poll until done
    let status: EnrichJobStatus = result;
    for (let i = 0; i < 30; i++) {
      status = await crawler.getEnrichJob(result.jobId);
      if (isEnrichJobComplete(status)) break;
      await new Promise((r) => setTimeout(r, 2000));
    }

    expect(isEnrichJobComplete(status)).toBe(true);
    expect(status.rows).toBeDefined();
  }, 180000);

  test('list jobs', async () => {
    const jobs = await crawler.listEnrichJobs({ limit: 5 });
    expect(Array.isArray(jobs)).toBe(true);
    // We created jobs in earlier tests, so there should be at least one
    expect(jobs.length).toBeGreaterThanOrEqual(1);
    expect(jobs[0].jobId).toBeTruthy();
  }, 30000);

  test('cancel job', async () => {
    const result = await crawler.enrich(
      ['https://example.com', 'https://httpbin.org', 'https://kidocode.com'],
      [
        { name: 'Title' },
        { name: 'Description', description: 'page description' },
        { name: 'Email' },
      ],
      { maxDepth: 1, enableSearch: true, wait: false },
    );

    expect(result.jobId).toMatch(/^enr_/);

    // Cancel
    const cancelled = await crawler.cancelEnrichJob(result.jobId);
    expect(cancelled).toBe(true);

    // Verify cancelled
    const status = await crawler.getEnrichJob(result.jobId);
    expect(status.status).toBe('cancelled');
  }, 60000);
});

// =============================================================================
// PROGRESS & TOKEN USAGE
// =============================================================================

describe('Enrich - Progress & Token Usage', () => {
  test('progress tracking', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com'],
      [{ name: 'Company Name' }],
      { maxDepth: 0, enableSearch: false, timeout: 120 },
    );

    expect(result.progress).toBeDefined();
    expect(result.progress.total).toBe(1);
    expect(result.progress.completed + result.progress.failed).toBe(1);
    expect(result.progressPercent).toBe(100);
  }, 180000);

  test('token usage tracked per row', async () => {
    const result = await crawler.enrich(
      ['https://kidocode.com'],
      [{ name: 'Company Name' }, { name: 'Email' }],
      { maxDepth: 0, enableSearch: false, timeout: 120 },
    );

    const row = result.rows![0];
    expect(row.tokenUsage).toBeDefined();
    expect(row.tokenUsage!['total_tokens']).toBeGreaterThan(0);
  }, 180000);
});
