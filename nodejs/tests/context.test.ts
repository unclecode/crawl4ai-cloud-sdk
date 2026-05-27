/**
 * Context v2 SDK tests — unit + live.
 *
 * Unit (always run):
 *   - All pillar builders produce the documented wire dict shape
 *   - Synthesizer.markdown / llm validation
 *   - Strategy.llmRerank
 *   - Inline pipeline body composition
 *   - contextOutputFromDict shape-specific sugar (markdown / files / data)
 *   - SSE event parsing
 *
 * Live (skipped when SKIP_LIVE=1) — hits stage. Submits real Context runs.
 */
import {
  AsyncWebCrawler,
  Source,
  Strategy,
  Synthesizer,
  Shape,
  Reconciler,
  Constraints,
  ContextResult,
  contextOutputFromDict,
  parseContextEvent,
  CONTEXT_TERMINAL_STATUSES,
  type ContextEvent,
  type ContextOutput,
  type ContextCatalog,
} from '../src';
import { QuotaExceededError } from '../src';

const API_KEY =
  process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const BASE_URL =
  process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

const LIVE_TIMEOUT_MS = 240_000;

// ─── Unit — Source builders ─────────────────────────────────────────────

describe('Source builder', () => {
  it('googleWeb — defaults', () => {
    expect(Source.googleWeb()).toEqual({
      type: 'google_web',
      params: { top_k_per_backend: 10 },
    });
  });

  it('googleWeb — full', () => {
    expect(
      Source.googleWeb({ backends: ['google', 'bing'], topKPerBackend: 8, region: 'us' }),
    ).toEqual({
      type: 'google_web',
      params: { backends: ['google', 'bing'], top_k_per_backend: 8, region: 'us' },
    });
  });

  it('googleDrive — search default', () => {
    const out = Source.googleDrive();
    expect(out.type).toBe('google_drive');
    expect(out.params).toEqual({ mode: 'search', folder_id: '' });
  });

  it('googleDrive — folder', () => {
    expect(Source.googleDrive({ mode: 'folder', folderId: 'abc' })).toEqual({
      type: 'google_drive',
      params: { mode: 'folder', folder_id: 'abc' },
    });
  });

  it('googleDrive — folder requires folderId', () => {
    expect(() => Source.googleDrive({ mode: 'folder' })).toThrow(/folderId is required/);
  });

  it('googleDrive — with authRef', () => {
    expect(Source.googleDrive({ authRef: 'link_abc' }).auth_ref).toBe('link_abc');
  });

  it('gmail — search default', () => {
    const out = Source.gmail();
    expect(out.type).toBe('gmail');
    expect(out.params.mode).toBe('search');
    expect(out.params.include_spam_trash).toBe(false);
  });

  it('gmail — label', () => {
    expect(
      Source.gmail({
        mode: 'label',
        labelId: 'Label_42',
        after: '2026/01/01',
        before: '2026/05/01',
        includeSpamTrash: true,
      }),
    ).toEqual({
      type: 'gmail',
      params: {
        mode: 'label',
        label_id: 'Label_42',
        after: '2026/01/01',
        before: '2026/05/01',
        include_spam_trash: true,
      },
    });
  });

  it('gmail — label requires labelId', () => {
    expect(() => Source.gmail({ mode: 'label' })).toThrow(/labelId is required/);
  });

  it('crawl', () => {
    const out = Source.crawl({
      domain: 'https://example.com',
      maxUrls: 30,
      maxDepth: 2,
      scoreThreshold: 0.5,
      profileId: 'my-profile',
    });
    expect(out.params).toMatchObject({
      domain: 'https://example.com',
      max_urls: 30,
      max_depth: 2,
      score_threshold: 0.5,
      profile_id: 'my-profile',
    });
  });

  it('file', () => {
    expect(Source.file({ fileId: 'file_abc' })).toEqual({
      type: 'file',
      params: { file_id: 'file_abc', chunk_size: 2000, chunk_overlap: 200 },
    });
  });

  it('custom', () => {
    expect(Source.custom({ type: 'hackernews', params: { tag: 'ai' }, authRef: 'link_x' })).toEqual({
      type: 'hackernews',
      params: { tag: 'ai' },
      auth_ref: 'link_x',
    });
  });
});

// ─── Unit — Strategy builders ───────────────────────────────────────────

describe('Strategy builder', () => {
  it('allItems', () => {
    expect(Strategy.allItems()).toEqual({ type: 'all_items', params: {} });
  });

  it('llmRerank — defaults', () => {
    const out = Strategy.llmRerank();
    expect(out.type).toBe('llm_rerank');
    expect(out.params).toMatchObject({
      top_n: 0,
      instruction: '',
      score_threshold: 0,
      batch_size: 20,
      max_concurrency: 4,
      content_aware: false,
      content_chars: 4000,
    });
  });

  it('llmRerank — full', () => {
    const out = Strategy.llmRerank({
      topN: 5,
      instruction: 'Prefer official docs.',
      model: 'anthropic/claude-sonnet-4-6',
      scoreThreshold: 0.3,
      contentAware: true,
      contentChars: 6000,
    });
    expect(out.params).toMatchObject({
      top_n: 5,
      instruction: 'Prefer official docs.',
      model: 'anthropic/claude-sonnet-4-6',
      score_threshold: 0.3,
      content_aware: true,
      content_chars: 6000,
    });
  });

  it('custom', () => {
    expect(Strategy.custom({ type: 'custom_s', params: { x: 1 } })).toEqual({
      type: 'custom_s',
      params: { x: 1 },
    });
  });
});

// ─── Unit — Synthesizer builders ────────────────────────────────────────

describe('Synthesizer builder', () => {
  it('raw', () => {
    expect(Synthesizer.raw()).toEqual({ type: 'raw', params: {} });
  });

  it('Shape alias === Synthesizer', () => {
    expect(Shape).toBe(Synthesizer);
    expect(Shape.raw()).toEqual({ type: 'raw', params: {} });
  });

  it('markdown — single defaults', () => {
    const out = Synthesizer.markdown();
    expect(out.type).toBe('markdown');
    expect(out.params).toMatchObject({
      mode: 'single',
      instruction: '',
      batch_size: 5,
      max_concurrency: 4,
      include_metadata: true,
      max_chars_per_item: 20000,
    });
  });

  it('markdown — multi with instruction', () => {
    const out = Synthesizer.markdown({ mode: 'multi', instruction: 'Summarise.' });
    expect(out.params.mode).toBe('multi');
    expect(out.params.instruction).toBe('Summarise.');
  });

  it('markdown — bad mode', () => {
    expect(() => Synthesizer.markdown({ mode: 'bogus' as any })).toThrow(/mode must be/);
  });

  it('llm — by example (object → JSON-serialised)', () => {
    const out = Synthesizer.llm({
      instruction: 'extract',
      example: { nodes: [{ id: 1 }] },
    });
    expect(out.type).toBe('llm');
    expect(out.params.instruction).toBe('extract');
    expect(JSON.parse(out.params.output_example as string)).toEqual({ nodes: [{ id: 1 }] });
    expect(out.params.output_schema).toBe('');
  });

  it('llm — by schema dict', () => {
    const out = Synthesizer.llm({
      instruction: 'tabulate',
      schema: { type: 'object', properties: { a: { type: 'string' } } },
    });
    expect(JSON.parse(out.params.output_schema as string).type).toBe('object');
  });

  it('llm — by schema string (passes through)', () => {
    const out = Synthesizer.llm({ instruction: 'x', schema: '{"type":"object"}' });
    expect(out.params.output_schema).toBe('{"type":"object"}');
  });

  it('llm — by description', () => {
    const out = Synthesizer.llm({ instruction: 'x', description: 'an object with a and b' });
    expect(out.params.output_description).toBe('an object with a and b');
  });

  it('llm — instruction required', () => {
    expect(() => Synthesizer.llm({ instruction: '', example: { a: 1 } })).toThrow(
      /instruction is required/,
    );
  });

  it('llm — exactly one of schema/example/description', () => {
    expect(() => Synthesizer.llm({ instruction: 'x' })).toThrow(/exactly one/);
    expect(() => Synthesizer.llm({ instruction: 'x', schema: {}, example: {} })).toThrow(
      /exactly one/,
    );
    expect(() => Synthesizer.llm({ instruction: 'x', schema: {}, description: 'x' })).toThrow(
      /exactly one/,
    );
  });

  it('custom', () => {
    expect(Synthesizer.custom({ type: 'future', params: { k: 'v' } })).toEqual({
      type: 'future',
      params: { k: 'v' },
    });
  });
});

// ─── Unit — Reconciler builders ─────────────────────────────────────────

describe('Reconciler builder', () => {
  it('noop', () => {
    expect(Reconciler.noop()).toEqual({ type: 'noop', params: {} });
  });

  it('custom with schedule', () => {
    const out = Reconciler.custom({ type: 'cron', params: { schedule: '0 6 * * *', tz: 'UTC' } });
    expect(out.type).toBe('cron');
    expect(out.params.schedule).toBe('0 6 * * *');
  });
});

// ─── Unit — Constraints ─────────────────────────────────────────────────

describe('Constraints', () => {
  it('defaults', () => {
    const out = new Constraints().toDict();
    expect(out.max_items).toBe(20);
    expect(out.language).toBe('en');
    expect(out.freshness_days).toBeUndefined();
  });

  it('freshness when set', () => {
    expect(new Constraints({ freshnessDays: 7 }).toDict().freshness_days).toBe(7);
  });
});

// ─── Unit — pipeline body composition ───────────────────────────────────

describe('context() body composition (inline pipeline)', () => {
  const crawler = new AsyncWebCrawler({ apiKey: 'sk_test_dummy', baseUrl: 'http://x' });
  // _buildContextBody is private; cast for tests.
  const build = (opts: any): any => (crawler as any)._buildContextBody(opts);

  it('minimal', () => {
    expect(build({ intent: 'x' })).toEqual({ intent: 'x' });
  });

  it('generator_id', () => {
    expect(build({ intent: 'x', generatorId: 'gen_42' })).toEqual({
      intent: 'x',
      generator_id: 'gen_42',
    });
  });

  it('inline pipeline — basic', () => {
    const body = build({
      intent: 'x',
      sources: [Source.googleWeb()],
    });
    expect(body.pipeline.sources[0].type).toBe('google_web');
    expect(body.pipeline.strategy).toBeUndefined();
    expect(body.pipeline.synthesizer).toBeUndefined();
  });

  it('inline pipeline — full', () => {
    const body = build({
      intent: 'x',
      sources: [
        Source.googleWeb(),
        Source.googleDrive({ mode: 'folder', folderId: 'abc' }),
      ],
      strategy: Strategy.llmRerank({ topN: 5, instruction: 'prefer docs' }),
      synthesizer: Synthesizer.markdown({ mode: 'single' }),
      reconciler: Reconciler.noop(),
    });
    expect(body.pipeline.sources).toHaveLength(2);
    expect(body.pipeline.strategy).toBe('llm_rerank');
    expect(body.pipeline.strategy_params.top_n).toBe(5);
    expect(body.pipeline.synthesizer).toBe('markdown');
    expect(body.pipeline.synthesizer_params.mode).toBe('single');
    expect(body.pipeline.reconciler).toBe('noop');
  });

  it('shape alias kwarg accepted', () => {
    const body = build({
      intent: 'x',
      sources: [Source.googleWeb()],
      shape: Synthesizer.markdown({ mode: 'single' }),
    });
    expect(body.pipeline.synthesizer).toBe('markdown');
  });

  it('synthesizer wins over shape alias', () => {
    const body = build({
      intent: 'x',
      sources: [Source.googleWeb()],
      synthesizer: Synthesizer.markdown({ mode: 'multi' }),
      shape: Synthesizer.raw(),
    });
    expect(body.pipeline.synthesizer).toBe('markdown');
    expect(body.pipeline.synthesizer_params.mode).toBe('multi');
  });

  it('mutual exclusion: generatorId + pillars', () => {
    expect(() =>
      build({ intent: 'x', generatorId: 'gen_x', sources: [Source.googleWeb()] }),
    ).toThrow(/either `generatorId`/);
  });

  it('inline pipeline requires a source', () => {
    expect(() => build({ intent: 'x', synthesizer: Synthesizer.raw() })).toThrow(
      /at least one Source/,
    );
  });
});

// ─── Unit — ContextOutput sugar ─────────────────────────────────────────

describe('ContextOutput sugar', () => {
  it('raw shape', () => {
    const out: ContextOutput = contextOutputFromDict({
      type: 'raw',
      data: { items: [{ url: 'https://x', title: 'T', content: 'C' }] },
    });
    expect(out.shape).toBe('raw');
    expect(out.items).toHaveLength(1);
    expect(out.items[0].url).toBe('https://x');
    expect(out.markdown).toBeUndefined();
    expect(out.files).toBeUndefined();
    expect(out.data).toBeUndefined();
  });

  it('markdown single', () => {
    const out = contextOutputFromDict({
      type: 'markdown',
      data: {
        mode: 'single',
        items: [{ url: 'https://a' }],
        markdown: '# heading\n\nbody',
      },
    });
    expect(out.shape).toBe('markdown');
    expect(out.markdown).toBe('# heading\n\nbody');
    expect(out.files).toBeUndefined();
  });

  it('markdown multi', () => {
    const out = contextOutputFromDict({
      type: 'markdown',
      data: {
        mode: 'multi',
        items: [{ url: 'https://a' }, { url: 'https://b' }],
        files: [
          { filename: 'a.md', markdown: '# A' },
          { filename: 'b.md', markdown: '# B' },
        ],
      },
    });
    expect(out.markdown).toBeUndefined();
    expect(out.files).toHaveLength(2);
    expect(out.files![0].filename).toBe('a.md');
  });

  it('llm', () => {
    const out = contextOutputFromDict({
      type: 'llm',
      data: {
        items: [{ url: 'https://a' }],
        data: { runtimes: [{ name: 'tokio' }] },
        resolved_schema: { type: 'object' },
        notes: ['resolved schema from output_example (walked)'],
      },
    });
    expect(out.shape).toBe('llm');
    expect(out.data).toEqual({ runtimes: [{ name: 'tokio' }] });
    expect(out.resolvedSchema).toEqual({ type: 'object' });
    expect(out.notes).toEqual(['resolved schema from output_example (walked)']);
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
    });
    if (ev === null || ev.type !== 'status') throw new Error('expected status');
    expect(ev.status).toBe('planning');
  });

  it('phase_progress init', () => {
    const ev = parseContextEvent('phase_progress', {
      type: 'phase_progress',
      kind: 'init',
      total: 3,
      items: [{ id: 'a' }],
    });
    if (ev === null || ev.type !== 'phase_progress' || ev.kind !== 'init') {
      throw new Error('expected init');
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
    });
    if (ev === null || ev.type !== 'phase_progress' || ev.kind !== 'item_update') {
      throw new Error('expected item_update');
    }
    expect(ev.id).toBe('abc');
    expect(ev.ms).toBe(1240);
  });

  it('terminal', () => {
    const ev = parseContextEvent('terminal', {
      type: 'terminal',
      status: 'completed',
      total_ms: 21834,
      urls_crawled: 9,
    });
    if (ev === null || ev.type !== 'terminal') throw new Error('expected terminal');
    expect(ev.urlsCrawled).toBe(9);
  });

  it('unknown returns null', () => {
    expect(parseContextEvent('mystery', { type: 'mystery' })).toBeNull();
  });
});

// ─── Unit — ContextResult helpers ───────────────────────────────────────

describe('ContextResult', () => {
  it('isTerminal covers all terminal statuses', () => {
    for (const s of ['completed', 'completed_partial', 'failed', 'cancelled']) {
      expect(new ContextResult({ runId: 'x', status: s, version: 1 }).isTerminal).toBe(true);
      expect(CONTEXT_TERMINAL_STATUSES.has(s)).toBe(true);
    }
  });

  it('isSuccess only on completed / completed_partial', () => {
    expect(new ContextResult({ runId: 'x', status: 'completed', version: 1 }).isSuccess).toBe(true);
    expect(
      new ContextResult({ runId: 'x', status: 'completed_partial', version: 1 }).isSuccess,
    ).toBe(true);
    expect(new ContextResult({ runId: 'x', status: 'failed', version: 1 }).isSuccess).toBe(false);
  });
});

// ─── Live tests ─────────────────────────────────────────────────────────

const RUN_LIVE = !process.env.SKIP_LIVE;
const liveIt = RUN_LIVE ? it : it.skip;
const sleep = (ms: number): Promise<void> => new Promise((r) => setTimeout(r, ms));

describe('Context v2 — live (stage)', () => {
  let crawler: AsyncWebCrawler;
  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
  });
  afterAll(async () => {
    await crawler.close();
  });

  liveIt(
    'default generator — one-shot',
    async () => {
      const result = await crawler.context({
        intent: 'brief overview of LangChain with citations',
        constraints: new Constraints({ maxItems: 5, maxPerSource: 3, maxCrawlTimeS: 60 }),
        wait: true,
        timeoutMs: 180_000,
      });
      expect(result.isTerminal).toBe(true);
      const output = await result.output();
      expect(['raw', 'markdown', 'llm']).toContain(output.shape);
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'inline pipeline raw — succeeds (previously raised NotImplementedError)',
    async () => {
      const result = await crawler.context({
        intent: 'what is HTTP/2',
        sources: [Source.googleWeb({ topKPerBackend: 5 })],
        strategy: Strategy.allItems(),
        synthesizer: Synthesizer.raw(),
        reconciler: Reconciler.noop(),
        constraints: new Constraints({ maxItems: 3, maxCrawlTimeS: 45 }),
        wait: true,
        timeoutMs: 180_000,
      });
      expect(result.isTerminal).toBe(true);
      const output = await result.output();
      expect(output.shape).toBe('raw');
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'inline pipeline markdown single',
    async () => {
      const result = await crawler.context({
        intent: 'one-line overview of WebAssembly',
        sources: [Source.googleWeb({ topKPerBackend: 3 })],
        synthesizer: Synthesizer.markdown({ mode: 'single' }),
        constraints: new Constraints({ maxItems: 2, maxCrawlTimeS: 45 }),
        wait: true,
        timeoutMs: 180_000,
      });
      expect(result.isTerminal).toBe(true);
      const output = await result.output();
      expect(output.shape).toBe('markdown');
    },
    LIVE_TIMEOUT_MS,
  );

  liveIt(
    'catalog — surfaces synthesizers including markdown + llm',
    async () => {
      const catalog: ContextCatalog = await crawler.contextCatalog();
      const sourceNames = catalog.sources.map((s) => s.name);
      const synthNames = catalog.synthesizers.map((s) => s.name);
      expect(sourceNames).toContain('google_web');
      expect(synthNames).toContain('raw');
      expect(synthNames).toContain('markdown');
      expect(synthNames).toContain('llm');
    },
    60_000,
  );
});

// Silence unused-import warnings.
void parseContextEvent;
void ((): ContextEvent | null => null);
