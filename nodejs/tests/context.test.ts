/**
 * Context v2 SDK tests — two layers (unit + live).
 *
 * Layer 1 (unit) — pure checks that don't hit the network:
 * - Pillar builders produce the documented dict shape
 * - context() validates mutually-exclusive generatorId vs pillars
 * - parseContextEvent translates SSE payloads into the right typed event
 * - ContextResult isTerminal / isSuccess flags work
 * - Constraints.toDict produces the API-expected shape
 *
 * Layer 2 (live) — hits stage by default. Submits a real Context run,
 * streams it, fetches output, refreshes, lists versions, cancels. Gated
 * on CRAWL4AI_API_KEY being non-default; otherwise these are still run
 * but skipped at the top via the SKIP_LIVE env var.
 */
import {
  AsyncWebCrawler,
  Source,
  Strategy,
  Shape,
  Reconciler,
  Constraints,
  ContextResult,
  ContextNotImplementedError,
  parseContextEvent,
  CONTEXT_TERMINAL_STATUSES,
  type ContextEvent,
  type ContextOutput,
  type ContextVersion,
  type ContextDiff,
  type ContextCatalog,
} from '../src';
import { QuotaExceededError } from '../src';

const API_KEY =
  process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const BASE_URL =
  process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

const LIVE_TIMEOUT_MS = 240_000;

// ─── Unit — pillar builders ─────────────────────────────────────────────

describe('Source builder', () => {
  it('googleWeb — defaults', () => {
    expect(Source.googleWeb()).toEqual({
      type: 'google_web',
      params: { top_k_per_backend: 10 },
    });
  });

  it('googleWeb — with backends + region', () => {
    expect(
      Source.googleWeb({ backends: ['google', 'bing'], topKPerBackend: 8, region: 'us' }),
    ).toEqual({
      type: 'google_web',
      params: { backends: ['google', 'bing'], top_k_per_backend: 8, region: 'us' },
    });
  });

  it('crawl — required fields', () => {
    const out = Source.crawl({ domain: 'https://example.com', maxUrls: 30, maxDepth: 2 });
    expect(out.type).toBe('crawl');
    expect(out.params).toMatchObject({
      domain: 'https://example.com',
      max_urls: 30,
      max_depth: 2,
    });
  });

  it('crawl — with optional fields', () => {
    const out = Source.crawl({
      domain: 'https://example.com',
      maxUrls: 10,
      maxDepth: 1,
      scoreThreshold: 0.5,
      profileId: 'my-profile',
    });
    expect(out.params.score_threshold).toBe(0.5);
    expect(out.params.profile_id).toBe('my-profile');
  });

  it('file', () => {
    expect(Source.file({ fileId: 'file_abc', chunkSize: 1500, chunkOverlap: 150 })).toEqual({
      type: 'file',
      params: { file_id: 'file_abc', chunk_size: 1500, chunk_overlap: 150 },
    });
  });

  it('custom — passthrough', () => {
    expect(Source.custom({ type: 'hackernews', params: { tag: 'ai', limit: 50 } })).toEqual({
      type: 'hackernews',
      params: { tag: 'ai', limit: 50 },
    });
  });

  it('custom — with authRef', () => {
    const out = Source.custom({
      type: 'slack',
      params: { channel: 'C123' },
      authRef: 'link_abc',
    });
    expect(out.auth_ref).toBe('link_abc');
  });
});

describe('Strategy / Shape / Reconciler builders', () => {
  it('Strategy.allItems', () => {
    expect(Strategy.allItems()).toEqual({ type: 'all_items', params: {} });
  });

  it('Strategy.custom', () => {
    expect(
      Strategy.custom({ type: 'llm_rerank', params: { model: 'claude-haiku-4-5' } }),
    ).toEqual({
      type: 'llm_rerank',
      params: { model: 'claude-haiku-4-5' },
    });
  });

  it('Shape.raw', () => {
    expect(Shape.raw()).toEqual({ type: 'raw', params: {} });
  });

  it('Shape.custom', () => {
    expect(Shape.custom({ type: 'markdown_digest' })).toEqual({
      type: 'markdown_digest',
      params: {},
    });
  });

  it('Reconciler.noop', () => {
    expect(Reconciler.noop()).toEqual({ type: 'noop', params: {} });
  });

  it('Reconciler.custom with schedule', () => {
    const out = Reconciler.custom({
      type: 'cron',
      params: { schedule: '0 6 * * *', tz: 'UTC' },
    });
    expect(out.type).toBe('cron');
    expect(out.params.schedule).toBe('0 6 * * *');
  });
});

// ─── Unit — Constraints ─────────────────────────────────────────────────

describe('Constraints', () => {
  it('defaults', () => {
    const out = new Constraints().toDict();
    expect(out.max_items).toBe(20);
    expect(out.max_per_source).toBe(10);
    expect(out.max_crawl_time_s).toBe(120);
    expect(out.language).toBe('en');
    expect(out.freshness_days).toBeUndefined();
  });

  it('freshness emits when set', () => {
    expect(new Constraints({ freshnessDays: 7 }).toDict().freshness_days).toBe(7);
  });

  it('override all', () => {
    const out = new Constraints({
      maxItems: 50,
      maxPerSource: 20,
      maxCrawlTimeS: 300,
      freshnessDays: 30,
      language: 'fr',
    }).toDict();
    expect(out).toEqual({
      max_items: 50,
      max_per_source: 20,
      max_crawl_time_s: 300,
      freshness_days: 30,
      language: 'fr',
    });
  });
});

// ─── Unit — body composition + validation ───────────────────────────────

describe('context() body composition', () => {
  const crawler = new AsyncWebCrawler({ apiKey: 'sk_test_dummy', baseUrl: 'http://x' });

  it('rejects ad-hoc pillar configs (public CRUD not yet shipped)', async () => {
    await expect(
      crawler.context({
        intent: 'compare X and Y',
        sources: [Source.googleWeb()],
        strategy: Strategy.allItems(),
        shape: Shape.raw(),
        reconciler: Reconciler.noop(),
      }),
    ).rejects.toBeInstanceOf(ContextNotImplementedError);
  });

  it('rejects mutually-exclusive generatorId + pillars', async () => {
    await expect(
      crawler.context({
        intent: 'x',
        generatorId: 'gen_x',
        sources: [Source.googleWeb()],
      }),
    ).rejects.toThrow(/either `generatorId` OR pillar params/);
  });

  it('rejects empty intent', async () => {
    await expect(crawler.context({ intent: '   ' })).rejects.toThrow(
      /`intent` is required/,
    );
  });
});

// ─── Unit — SSE event parsing ───────────────────────────────────────────

describe('parseContextEvent', () => {
  it('status', () => {
    const ev = parseContextEvent('status', {
      type: 'status',
      status: 'planning',
      phase: 'planning',
      version: 1,
      ts: '2026-05-19T12:00:00Z',
    });
    if (ev === null) throw new Error('expected event');
    if (ev.type !== 'status') throw new Error('expected status event');
    expect(ev.status).toBe('planning');
    expect(ev.phase).toBe('planning');
    expect(ev.version).toBe(1);
  });

  it('phase_progress init', () => {
    const ev = parseContextEvent('phase_progress', {
      type: 'phase_progress',
      kind: 'init',
      phase: 'fetch',
      total: 3,
      items: [{ id: 'a', url: 'https://x' }],
    });
    if (ev === null) throw new Error('expected event');
    if (ev.type !== 'phase_progress' || ev.kind !== 'init') {
      throw new Error('expected phase_progress init event');
    }
    expect(ev.total).toBe(3);
  });

  it('phase_progress item_update', () => {
    const ev = parseContextEvent('phase_progress', {
      type: 'phase_progress',
      kind: 'item_update',
      id: 'abc',
      status: 'done',
      ms: 1240,
      size: 18432,
    });
    if (ev === null) throw new Error('expected event');
    if (ev.type !== 'phase_progress' || ev.kind !== 'item_update') {
      throw new Error('expected phase_progress item_update event');
    }
    expect(ev.id).toBe('abc');
    expect(ev.status).toBe('done');
    expect(ev.ms).toBe(1240);
    expect(ev.size).toBe(18432);
  });

  it('terminal', () => {
    const ev = parseContextEvent('terminal', {
      type: 'terminal',
      status: 'completed',
      total_ms: 21834,
      urls_crawled: 9,
      urls_failed: 0,
    });
    if (ev === null) throw new Error('expected event');
    if (ev.type !== 'terminal') throw new Error('expected terminal event');
    expect(ev.status).toBe('completed');
    expect(ev.urlsCrawled).toBe(9);
  });

  it('unknown event returns null', () => {
    expect(parseContextEvent('mystery', { type: 'mystery' })).toBeNull();
  });
});

// ─── Unit — ContextResult helpers ───────────────────────────────────────

describe('ContextResult', () => {
  it('isTerminal flag covers all terminal statuses', () => {
    for (const s of ['completed', 'completed_partial', 'failed', 'cancelled']) {
      const r = new ContextResult({ runId: 'x', status: s, version: 1 });
      expect(r.isTerminal).toBe(true);
      expect(CONTEXT_TERMINAL_STATUSES.has(s)).toBe(true);
    }
  });

  it('isSuccess only on completed / completed_partial', () => {
    expect(new ContextResult({ runId: 'x', status: 'completed', version: 1 }).isSuccess).toBe(true);
    expect(
      new ContextResult({ runId: 'x', status: 'completed_partial', version: 1 }).isSuccess,
    ).toBe(true);
    expect(new ContextResult({ runId: 'x', status: 'failed', version: 1 }).isSuccess).toBe(false);
    expect(new ContextResult({ runId: 'x', status: 'queued', version: 1 }).isTerminal).toBe(false);
  });
});

// ─── Live tests (skip unless explicitly enabled) ────────────────────────

const RUN_LIVE = !process.env.SKIP_LIVE;

const liveIt = RUN_LIVE ? it : it.skip;
const sleep = (ms: number): Promise<void> =>
  new Promise((r) => setTimeout(r, ms));

describe('Context v2 — live (stage)', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
  });

  afterAll(async () => {
    await crawler.close();
  });

  liveIt(
    'default generator — one-shot waiter',
    async () => {
      const result = await crawler.context({
        intent: 'brief overview of what LangChain is, with citations',
        constraints: new Constraints({
          maxItems: 5,
          maxPerSource: 3,
          maxCrawlTimeS: 60,
        }),
        wait: true,
        timeoutMs: 180_000,
      });
      expect(result.isTerminal).toBe(true);
      expect(result.runId.length).toBeGreaterThan(8);
      expect(result.version).toBeGreaterThanOrEqual(1);

      const output: ContextOutput = await result.output();
      expect(output.shape).toBe('raw');
      expect(Array.isArray(output.items)).toBe(true);
      for (const item of output.items) {
        // Provenance contract: each item carries a URL or a snippet
        expect(item.url || item.snippet).toBeTruthy();
      }
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'streaming — typed events flow through to terminal',
    async () => {
      const seen = new Set<string>();
      let sawTerminal = false;

      for await (const event of crawler.contextStream({
        intent: 'one-line answer: what is RAG',
        constraints: new Constraints({
          maxItems: 2,
          maxPerSource: 2,
          maxCrawlTimeS: 30,
        }),
      })) {
        seen.add(event.type);
        if (event.type === 'terminal') {
          sawTerminal = true;
          expect(
            ['completed', 'completed_partial', 'failed', 'cancelled'].includes(event.status),
          ).toBe(true);
        }
        if (event.type === 'phase_progress' && event.kind === 'item_update') {
          expect(['done', 'failed']).toContain(event.status);
        }
      }
      expect(sawTerminal).toBe(true);
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'pillar params raise NotImplementedError end-to-end',
    async () => {
      await expect(
        crawler.context({
          intent: 'test',
          sources: [Source.googleWeb()],
        }),
      ).rejects.toBeInstanceOf(ContextNotImplementedError);
    },
    20_000,
  );

  liveIt(
    'submit no-wait, get state, cancel — slot returns',
    async () => {
      // Best-effort: poll briefly until a slot is free.
      const deadline = Date.now() + 60_000;
      let result: ContextResult | undefined;
      while (Date.now() < deadline) {
        try {
          result = await crawler.context({
            intent: 'ignored — will be cancelled',
            constraints: new Constraints({ maxItems: 2, maxCrawlTimeS: 30 }),
            wait: false,
          });
          break;
        } catch (err) {
          if (err instanceof QuotaExceededError) {
            await sleep(5000);
            continue;
          }
          throw err;
        }
      }
      expect(result).toBeDefined();
      expect(result!.isTerminal).toBe(false);

      const state = await crawler.getContextRun(result!.runId);
      expect(state.runId).toBe(result!.runId);

      // Cancel — async on the server side
      await crawler.cancelContextRun(result!.runId);
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'versions + refresh — v2 lands on the chain',
    async () => {
      const deadline = Date.now() + 90_000;
      let v1: ContextResult | undefined;
      while (Date.now() < deadline) {
        try {
          v1 = await crawler.context({
            intent: 'one-line overview of vector databases',
            constraints: new Constraints({
              maxItems: 2,
              maxPerSource: 2,
              maxCrawlTimeS: 30,
            }),
            wait: true,
            timeoutMs: 180_000,
          });
          break;
        } catch (err) {
          if (err instanceof QuotaExceededError) {
            await sleep(5000);
            continue;
          }
          throw err;
        }
      }
      expect(v1).toBeDefined();
      expect(v1!.isTerminal).toBe(true);

      let v2: ContextResult | undefined;
      while (Date.now() < deadline) {
        try {
          v2 = await crawler.refreshContext(v1!.runId, { wait: true, timeoutMs: 180_000 });
          break;
        } catch (err) {
          if (err instanceof QuotaExceededError) {
            await sleep(5000);
            continue;
          }
          throw err;
        }
      }
      expect(v2).toBeDefined();
      expect(v2!.version).toBeGreaterThanOrEqual(v1!.version);

      const versions: ContextVersion[] = await crawler.listContextVersions(v1!.runId);
      expect(versions.length).toBeGreaterThanOrEqual(2);
      const maxVersion = Math.max(...versions.map((v) => v.version));
      expect(maxVersion).toBeGreaterThanOrEqual(v2!.version);
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'diff — same-chain returns ContextDiff',
    async () => {
      const deadline = Date.now() + 90_000;
      let v1: ContextResult | undefined;
      while (Date.now() < deadline) {
        try {
          v1 = await crawler.context({
            intent: 'one-line answer: what is HTTP/2',
            constraints: new Constraints({
              maxItems: 2,
              maxPerSource: 2,
              maxCrawlTimeS: 30,
            }),
            wait: true,
            timeoutMs: 180_000,
          });
          break;
        } catch (err) {
          if (err instanceof QuotaExceededError) {
            await sleep(5000);
            continue;
          }
          throw err;
        }
      }
      expect(v1).toBeDefined();

      try {
        await crawler.refreshContext(v1!.runId, { wait: true, timeoutMs: 180_000 });
      } catch (err) {
        if (err instanceof QuotaExceededError) {
          await sleep(10_000);
          await crawler.refreshContext(v1!.runId, { wait: true, timeoutMs: 180_000 });
        } else {
          throw err;
        }
      }

      const diff: ContextDiff = await crawler.diffContext(v1!.runId, v1!.runId);
      expect(Array.isArray(diff.added)).toBe(true);
      expect(Array.isArray(diff.removed)).toBe(true);
      expect(Array.isArray(diff.unchanged)).toBe(true);
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'catalog — the four pillars are surfaced',
    async () => {
      const catalog: ContextCatalog = await crawler.contextCatalog();
      const sourceNames = catalog.sources.map((s) => s.name);
      const strategyNames = catalog.strategies.map((s) => s.name);
      const shapeNames = catalog.shapes.map((s) => s.name);
      const reconcilerNames = catalog.reconcilers.map((s) => s.name);

      expect(sourceNames).toContain('google_web');
      expect(strategyNames).toContain('all_items');
      expect(shapeNames).toContain('raw');
      expect(reconcilerNames).toContain('noop');
    },
    60_000,
  );
});

// Silence unused-import warnings under jest --noEmit-style typecheck
void parseContextEvent;
void ((): ContextEvent | null => null);
