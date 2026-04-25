/**
 * Enrich v2 E2E tests — runs against stage.crawl4ai.com.
 *
 * Covers the seven-method surface:
 *    enrich(...)         POST /v1/enrich/async
 *    getEnrichJob        GET  /v1/enrich/jobs/{id}
 *    waitEnrichJob       poll loop, optional `until=` phase
 *    resumeEnrichJob     POST /v1/enrich/jobs/{id}/continue
 *    streamEnrichJob     GET  /v1/enrich/jobs/{id}/stream  (SSE)
 *    cancelEnrichJob     DELETE /v1/enrich/jobs/{id}
 *    listEnrichJobs      GET  /v1/enrich/jobs
 *
 * Usage:
 *    npx jest tests/enrich-e2e.test.ts --verbose
 */

import {
  AsyncWebCrawler,
  EnrichJobStatus,
  EnrichRow,
  EnrichEvent,
  EnrichJobListItem,
  isEnrichJobComplete,
  isEnrichJobSuccessful,
} from '../src';

const API_KEY =
  process.env.CRAWL4AI_API_KEY ||
  'sk_live_V89kxHtmkxw0jJORu_sWzyuvGw6TKHaJhoNGK8gGdqU';
const BASE_URL =
  process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

let crawler: AsyncWebCrawler;

beforeAll(() => {
  crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
});

afterAll(async () => {
  await crawler.close();
});

// ─── 1. URLs-only mode (fastest path) ────────────────────────────────

describe('Enrich v2 — URLs-only', () => {
  test('single URL + two features', async () => {
    const result = await crawler.enrich({
      urls: ['https://kidocode.com'],
      features: [
        { name: 'company_name' },
        { name: 'contact_email', description: 'primary contact email' },
      ],
      strategy: 'http',
      wait: true,
      timeout: 180,
    });
    expect(isEnrichJobComplete(result)).toBe(true);
    expect(isEnrichJobSuccessful(result)).toBe(true);

    const rows = result.phaseData.rows;
    expect(rows).toBeDefined();
    expect(rows!.length).toBeGreaterThanOrEqual(1);

    const row = rows![0];
    expect(row.url === 'https://kidocode.com' || row.groupId === 'https://kidocode.com').toBe(true);
    expect(Object.keys(row.fields).length).toBeGreaterThan(0);
    const hasCompany = Object.keys(row.fields).some(k => k.toLowerCase().startsWith('company'));
    expect(hasCompany).toBe(true);

    // Usage envelope sanity
    expect(result.usage.crawls).toBeGreaterThanOrEqual(1);
    expect(result.usage.llmTokensByPurpose.extract).toBeDefined();
    expect(result.usage.llmTokensByPurpose.extract.input).toBeGreaterThan(0);
    expect(result.usage.llmTokensByPurpose.extract.output).toBeGreaterThan(0);
  }, 240_000);

  test('string features shorthand', async () => {
    const result = await crawler.enrich({
      urls: ['https://example.com'],
      features: ['title', 'description'],
      strategy: 'http',
      wait: true,
      timeout: 120,
    });
    expect(isEnrichJobComplete(result)).toBe(true);
    expect(result.phaseData.rows).toBeDefined();
    expect(result.phaseData.rows!.length).toBeGreaterThanOrEqual(1);
  }, 180_000);
});

// ─── 2. Job lifecycle ────────────────────────────────────────────────

describe('Enrich v2 — Job lifecycle', () => {
  test('fire-and-forget then getEnrichJob', async () => {
    const job = await crawler.enrich({
      urls: ['https://kidocode.com'],
      features: [{ name: 'company_name' }],
      strategy: 'http',
      wait: false,
    });
    expect(job.jobId).toMatch(/^enr_/);
    expect(['queued', 'extracting', 'merging', 'completed']).toContain(job.status);

    const latest = await crawler.getEnrichJob(job.jobId);
    expect(latest.jobId).toBe(job.jobId);
  }, 60_000);

  test('waitEnrichJob until terminal', async () => {
    const job = await crawler.enrich({
      urls: ['https://example.com'],
      features: [{ name: 'title' }],
      strategy: 'http',
      wait: false,
    });
    const terminal = await crawler.waitEnrichJob(job.jobId, { timeout: 120 });
    expect(isEnrichJobComplete(terminal)).toBe(true);
    expect(isEnrichJobSuccessful(terminal)).toBe(true);
  }, 180_000);

  test('listEnrichJobs returns recent jobs', async () => {
    const jobs: EnrichJobListItem[] = await crawler.listEnrichJobs({ limit: 5 });
    expect(Array.isArray(jobs)).toBe(true);
    expect(jobs.length).toBeGreaterThanOrEqual(1);
    expect(jobs.every(j => j.jobId.startsWith('enr_'))).toBe(true);
  }, 30_000);

  test('cancelEnrichJob mid-flight', async () => {
    // Start a query-based job — slowest path so we have time to cancel
    const job = await crawler.enrich({
      query: 'top BBQ restaurants in Austin Texas with outdoor seating',
      country: 'us',
      topKPerEntity: 2,
      wait: false,
    });
    expect(job.jobId).toMatch(/^enr_/);

    const cancelled = await crawler.cancelEnrichJob(job.jobId);
    expect(cancelled).toBe(true);

    await new Promise(r => setTimeout(r, 2000));
    const latest = await crawler.getEnrichJob(job.jobId);
    expect(latest.status).toBe('cancelled');
  }, 60_000);
});

// ─── 3. Review flow: pause + resume with edits ───────────────────────

describe('Enrich v2 — Review flow', () => {
  test('pause at plan_ready, then resume', async () => {
    const job = await crawler.enrich({
      query: 'best Italian restaurants in Brooklyn New York',
      country: 'us',
      topKPerEntity: 1,
      autoConfirmPlan: false,   // pause here
      autoConfirmUrls: true,    // run after we resume
      wait: false,
    });

    const paused = await crawler.waitEnrichJob(job.jobId, {
      until: 'plan_ready', timeout: 120,
    });
    expect(paused.status).toBe('plan_ready');
    expect(paused.phaseData.plan).toBeDefined();
    expect(paused.phaseData.plan!.entities.length).toBeGreaterThanOrEqual(1);
    expect(paused.phaseData.plan!.features.length).toBeGreaterThanOrEqual(1);
    expect(paused.usage.llmTokensByPurpose.plan_intent).toBeDefined();

    // Trim entities + features so the rest of the pipeline is fast
    const editedEntities = [{ name: paused.phaseData.plan!.entities[0].name }];
    const editedFeatures = [{ name: 'address' }];
    const resumed = await crawler.resumeEnrichJob(job.jobId, {
      entities: editedEntities,
      features: editedFeatures,
    });
    expect(resumed.status).not.toBe('plan_ready');

    const final = await crawler.waitEnrichJob(job.jobId, { timeout: 300 });
    expect(isEnrichJobComplete(final)).toBe(true);
    if (final.phaseData.rows) {
      expect(final.phaseData.rows.length).toBeLessThanOrEqual(1);
    }
  }, 420_000);
});

// ─── 4. SSE streaming ────────────────────────────────────────────────

describe('Enrich v2 — SSE stream', () => {
  test('snapshot + complete events arrive', async () => {
    const job = await crawler.enrich({
      urls: ['https://example.com'],
      features: [{ name: 'title' }],
      strategy: 'http',
      wait: false,
    });

    const seen: string[] = [];
    const collect = (async () => {
      for await (const event of crawler.streamEnrichJob(job.jobId)) {
        seen.push(event.type);
        if (event.type === 'snapshot') {
          expect(event.snapshot).toBeDefined();
          expect(event.snapshot!.jobId).toBe(job.jobId);
        }
        if (event.type === 'complete') return;
      }
    })();
    await Promise.race([
      collect,
      new Promise((_, rej) => setTimeout(() => rej(new Error('stream timeout')), 120_000)),
    ]);

    expect(seen).toContain('snapshot');
    expect(seen).toContain('complete');
  }, 180_000);
});
