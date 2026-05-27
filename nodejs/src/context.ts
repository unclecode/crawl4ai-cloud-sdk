/**
 * Context v2 — the four-pillar research pipeline.
 *
 * Public surface:
 * - **Pillar builders** — typed factories for every pillar registered
 *   server-side:
 *     Source.googleWeb / googleDrive / gmail / crawl / file / custom
 *     Strategy.allItems / llmRerank / custom
 *     Synthesizer.raw / markdown / llm / custom
 *     Reconciler.noop / custom
 *   Each returns the API wire shape `{ type, params }`. `Shape` is a
 *   deprecated alias for `Synthesizer`.
 *
 * - **Result + event types** — `ContextResult` with lazy `output()`,
 *   `ContextOutput` with shape-specific sugar (`.markdown` for
 *   markdown-single, `.files` for markdown-multi, `.data` for llm),
 *   plus typed `StatusEvent` / `PhaseProgressInit` /
 *   `PhaseProgressItemUpdate` / `TerminalEvent` for the streaming
 *   iterator.
 *
 * The crawler methods (`context()`, `contextStream()`,
 * `refreshContext()` …) live on `AsyncWebCrawler` in `crawler.ts`.
 * This module is the pure data layer.
 */

// ─── Constants ──────────────────────────────────────────────────────────

export const CONTEXT_TERMINAL_STATUSES = new Set<string>([
  'completed',
  'completed_partial',
  'failed',
  'cancelled',
]);

export const CONTEXT_ACTIVE_STATUSES = new Set<string>(['queued', 'running']);

export const PHASE_PLANNING = 'planning';
export const PHASE_CRAWLING = 'crawling';
export const PHASE_SHAPING = 'shaping';
export const CONTEXT_PHASES = [
  PHASE_PLANNING,
  PHASE_CRAWLING,
  PHASE_SHAPING,
] as const;

// ─── Pillar config ──────────────────────────────────────────────────────

/** Wire shape expected by `/v1/context` for one Source / Strategy /
 * Synthesizer / Reconciler. */
export interface PillarConfig {
  type: string;
  params: Record<string, unknown>;
  auth_ref?: string;
}

// ─── helpers ────────────────────────────────────────────────────────────

function serialize(value: unknown): string {
  if (value === null || value === undefined) return '';
  if (typeof value === 'string') return value;
  return JSON.stringify(value);
}

// ─── Pillar builders ────────────────────────────────────────────────────

/** Builder for Context Source configs. */
export class Source {
  /**
   * Google search across multiple SERP backends with RRF merge.
   */
  static googleWeb(opts: {
    backends?: string[];
    topKPerBackend?: number;
    region?: string;
  } = {}): PillarConfig {
    const params: Record<string, unknown> = {
      top_k_per_backend: Math.trunc(opts.topKPerBackend ?? 10),
    };
    if (opts.backends !== undefined) params.backends = [...opts.backends];
    if (opts.region !== undefined) params.region = String(opts.region);
    return { type: 'google_web', params };
  }

  /**
   * User's Google Drive. `mode = 'search'` is intent-driven; `mode =
   * 'folder'` lists one folder (requires `folderId`).
   */
  static googleDrive(opts: {
    mode?: 'search' | 'folder';
    folderId?: string;
    authRef?: string;
  } = {}): PillarConfig {
    const mode = opts.mode ?? 'search';
    if (mode !== 'search' && mode !== 'folder') {
      throw new Error(`mode must be 'search' or 'folder', got ${String(mode)}`);
    }
    if (mode === 'folder' && !opts.folderId) {
      throw new Error("folderId is required when mode='folder'");
    }
    const out: PillarConfig = {
      type: 'google_drive',
      params: { mode, folder_id: String(opts.folderId ?? '') },
    };
    if (opts.authRef !== undefined) out.auth_ref = String(opts.authRef);
    return out;
  }

  /**
   * User's Gmail. `mode = 'search'` is intent-driven; `mode = 'label'`
   * lists threads in one label. Dates are `YYYY/MM/DD`.
   */
  static gmail(opts: {
    mode?: 'search' | 'label';
    labelId?: string;
    after?: string;
    before?: string;
    includeSpamTrash?: boolean;
    authRef?: string;
  } = {}): PillarConfig {
    const mode = opts.mode ?? 'search';
    if (mode !== 'search' && mode !== 'label') {
      throw new Error(`mode must be 'search' or 'label', got ${String(mode)}`);
    }
    if (mode === 'label' && !opts.labelId) {
      throw new Error("labelId is required when mode='label'");
    }
    const out: PillarConfig = {
      type: 'gmail',
      params: {
        mode,
        label_id: String(opts.labelId ?? ''),
        after: String(opts.after ?? ''),
        before: String(opts.before ?? ''),
        include_spam_trash: Boolean(opts.includeSpamTrash ?? false),
      },
    };
    if (opts.authRef !== undefined) out.auth_ref = String(opts.authRef);
    return out;
  }

  /** Recursive site crawl as the corpus. */
  static crawl(opts: {
    domain: string;
    maxUrls?: number;
    maxDepth?: number;
    scoreThreshold?: number;
    profileId?: string;
  }): PillarConfig {
    const params: Record<string, unknown> = {
      domain: String(opts.domain),
      max_urls: Math.trunc(opts.maxUrls ?? 50),
      max_depth: Math.trunc(opts.maxDepth ?? 3),
    };
    if (opts.scoreThreshold !== undefined) {
      params.score_threshold = Number(opts.scoreThreshold);
    }
    if (opts.profileId !== undefined) params.profile_id = String(opts.profileId);
    return { type: 'crawl', params };
  }

  /** User-uploaded file as the corpus. */
  static file(opts: {
    fileId: string;
    chunkSize?: number;
    chunkOverlap?: number;
  }): PillarConfig {
    return {
      type: 'file',
      params: {
        file_id: String(opts.fileId),
        chunk_size: Math.trunc(opts.chunkSize ?? 2000),
        chunk_overlap: Math.trunc(opts.chunkOverlap ?? 200),
      },
    };
  }

  /** Escape hatch for Sources without a typed builder yet. */
  static custom(opts: {
    type: string;
    params?: Record<string, unknown>;
    authRef?: string;
  }): PillarConfig {
    const out: PillarConfig = {
      type: String(opts.type),
      params: { ...(opts.params ?? {}) },
    };
    if (opts.authRef !== undefined) out.auth_ref = String(opts.authRef);
    return out;
  }
}

/** Builder for Context Strategy configs. */
export class Strategy {
  /** Passthrough — every candidate kept up to `constraints.maxItems`.
   * The default. */
  static allItems(): PillarConfig {
    return { type: 'all_items', params: {} };
  }

  /**
   * Score every candidate against the intent with an LLM, keep the
   * top N. `topN = 0` means use the request's `maxItems`.
   * `contentAware` scores on the item body (for `owns_content`
   * Sources like Drive / Gmail / HN) instead of just title + snippet.
   */
  static llmRerank(opts: {
    topN?: number;
    instruction?: string;
    model?: string;
    scoreThreshold?: number;
    batchSize?: number;
    maxConcurrency?: number;
    contentAware?: boolean;
    contentChars?: number;
  } = {}): PillarConfig {
    return {
      type: 'llm_rerank',
      params: {
        top_n: Math.trunc(opts.topN ?? 0),
        instruction: String(opts.instruction ?? ''),
        model: String(opts.model ?? 'anthropic/claude-haiku-4-5'),
        score_threshold: Number(opts.scoreThreshold ?? 0.0),
        batch_size: Math.trunc(opts.batchSize ?? 20),
        max_concurrency: Math.trunc(opts.maxConcurrency ?? 4),
        content_aware: Boolean(opts.contentAware ?? false),
        content_chars: Math.trunc(opts.contentChars ?? 4000),
      },
    };
  }

  /** Escape hatch for Strategies without a typed builder yet. */
  static custom(opts: { type: string; params?: Record<string, unknown> }): PillarConfig {
    return { type: String(opts.type), params: { ...(opts.params ?? {}) } };
  }
}

/**
 * Builder for Context Synthesizer configs.
 *
 * Previously named "Shape" — that builder is kept as `Shape` for one
 * release.
 */
export class Synthesizer {
  /** Per-item citations with `url` provenance. The default. */
  static raw(): PillarConfig {
    return { type: 'raw', params: {} };
  }

  /**
   * Render the materialised plan as markdown.
   *
   * `mode = 'single'` — one joined .md body (default).
   * `mode = 'multi'`  — one .md per item (downloadable as zip).
   *
   * When `instruction` is non-empty, each item is rewritten by the
   * LLM before the markdown is built.
   */
  static markdown(opts: {
    mode?: 'single' | 'multi';
    instruction?: string;
    model?: string;
    batchSize?: number;
    maxConcurrency?: number;
    includeMetadata?: boolean;
    maxCharsPerItem?: number;
  } = {}): PillarConfig {
    const mode = opts.mode ?? 'single';
    if (mode !== 'single' && mode !== 'multi') {
      throw new Error(`mode must be 'single' or 'multi', got ${String(mode)}`);
    }
    return {
      type: 'markdown',
      params: {
        mode,
        instruction: String(opts.instruction ?? ''),
        model: String(opts.model ?? 'anthropic/claude-haiku-4-5'),
        batch_size: Math.trunc(opts.batchSize ?? 5),
        max_concurrency: Math.trunc(opts.maxConcurrency ?? 4),
        include_metadata: Boolean(opts.includeMetadata ?? true),
        max_chars_per_item: Math.trunc(opts.maxCharsPerItem ?? 20000),
      },
    };
  }

  /**
   * One LLM call that fills a caller-defined JSON shape.
   *
   * Pass exactly one of:
   *   - `schema`      — full JSON Schema (used as-is)
   *   - `example`     — concrete JSON example (walked into schema)
   *   - `description` — plain-English shape description (LLM drafts schema)
   *
   * Object / array args for `schema` / `example` are JSON-serialised.
   */
  static llm(opts: {
    instruction: string;
    schema?: string | object;
    example?: string | object | unknown[];
    description?: string;
    model?: string;
    temperature?: number;
    maxCorpusChars?: number;
    autoRepair?: boolean;
  }): PillarConfig {
    if (!opts.instruction || !opts.instruction.trim()) {
      throw new Error('instruction is required for Synthesizer.llm');
    }
    // `!= undefined` rather than truthiness — caller passing an explicit
    // empty value is still a value we want to count.
    const provided = [opts.schema, opts.example, opts.description].filter(
      (v) => v !== undefined && v !== null,
    ).length;
    if (provided !== 1) {
      throw new Error('Pass exactly one of: schema, example, description');
    }
    return {
      type: 'llm',
      params: {
        instruction: String(opts.instruction),
        output_schema: serialize(opts.schema),
        output_example: serialize(opts.example),
        output_description: opts.description ? String(opts.description) : '',
        model: String(opts.model ?? 'anthropic/claude-haiku-4-5'),
        temperature: Number(opts.temperature ?? 0.0),
        max_corpus_chars: Math.trunc(opts.maxCorpusChars ?? 40000),
        auto_repair: Boolean(opts.autoRepair ?? true),
      },
    };
  }

  /** Escape hatch for Synthesizers without a typed builder yet. */
  static custom(opts: { type: string; params?: Record<string, unknown> }): PillarConfig {
    return { type: String(opts.type), params: { ...(opts.params ?? {}) } };
  }
}

/** @deprecated Use `Synthesizer`. Kept as alias for one release. */
export const Shape = Synthesizer;

/** Builder for Context Reconciler configs. */
export class Reconciler {
  /** No auto-refresh. Refreshes are user-initiated via
   * `refreshContext()`. The default. */
  static noop(): PillarConfig {
    return { type: 'noop', params: {} };
  }

  /** Escape hatch for Reconcilers without a typed builder yet
   * (e.g. `cron`, `webhook`). */
  static custom(opts: { type: string; params?: Record<string, unknown> }): PillarConfig {
    return { type: String(opts.type), params: { ...(opts.params ?? {}) } };
  }
}

// ─── Constraints ────────────────────────────────────────────────────────

export class Constraints {
  maxItems = 20;
  maxPerSource = 10;
  maxCrawlTimeS = 120.0;
  freshnessDays?: number;
  language = 'en';

  constructor(init: Partial<{
    maxItems: number;
    maxPerSource: number;
    maxCrawlTimeS: number;
    freshnessDays: number;
    language: string;
  }> = {}) {
    if (init.maxItems !== undefined) this.maxItems = init.maxItems;
    if (init.maxPerSource !== undefined) this.maxPerSource = init.maxPerSource;
    if (init.maxCrawlTimeS !== undefined) this.maxCrawlTimeS = init.maxCrawlTimeS;
    if (init.freshnessDays !== undefined) this.freshnessDays = init.freshnessDays;
    if (init.language !== undefined) this.language = init.language;
  }

  toDict(): Record<string, unknown> {
    const out: Record<string, unknown> = {
      max_items: Math.trunc(this.maxItems),
      max_per_source: Math.trunc(this.maxPerSource),
      max_crawl_time_s: Number(this.maxCrawlTimeS),
      language: String(this.language),
    };
    if (this.freshnessDays !== undefined) {
      out.freshness_days = Math.trunc(this.freshnessDays);
    }
    return out;
  }
}

export type ConstraintsInput = Constraints | Partial<{
  max_items: number;
  max_per_source: number;
  max_crawl_time_s: number;
  freshness_days: number;
  language: string;
  maxItems: number;
  maxPerSource: number;
  maxCrawlTimeS: number;
  freshnessDays: number;
}>;

export function constraintsToDict(c: ConstraintsInput): Record<string, unknown> {
  if (c instanceof Constraints) return c.toDict();
  const raw = c as Record<string, unknown>;
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(raw)) {
    if (v === undefined || v === null) continue;
    const key =
      k === 'maxItems' ? 'max_items' :
      k === 'maxPerSource' ? 'max_per_source' :
      k === 'maxCrawlTimeS' ? 'max_crawl_time_s' :
      k === 'freshnessDays' ? 'freshness_days' :
      k;
    out[key] = v;
  }
  return out;
}

// ─── Output types ───────────────────────────────────────────────────────

export interface ContextItem {
  url?: string;
  title?: string;
  content?: string;
  snippet?: string;
  source?: string;
  relevance: number;
  metadata: Record<string, unknown>;
  id?: string;
  fetchedAt?: string;
}

export function contextItemFromDict(data: Record<string, unknown>): ContextItem {
  const src = (data.source ?? data.source_name) as string | undefined;
  return {
    id: data.id as string | undefined,
    source: src,
    url: data.url as string | undefined,
    title: data.title as string | undefined,
    content: data.content as string | undefined,
    snippet: data.snippet as string | undefined,
    relevance: Number(data.relevance ?? 0),
    metadata: (data.metadata as Record<string, unknown>) ?? {},
    fetchedAt: data.fetched_at as string | undefined,
  };
}

/** One per-item markdown file emitted by `Synthesizer.markdown({ mode: 'multi' })`. */
export interface MarkdownFile {
  filename: string;
  markdown: string;
}

/**
 * The synthesized output.
 *
 * Shape-specific accessors:
 *   - `raw`      → `items` is the citation list
 *   - `markdown` → `markdown` (string, single mode) or `files`
 *                  (MarkdownFile[], multi mode)
 *   - `llm`      → `data` (the filled object), plus `resolvedSchema`,
 *                  `notes`, `partialData`
 *
 * For every shape, `rawPayload` is the full wire envelope.
 */
export interface ContextOutput {
  shape: string;
  items: ContextItem[];
  partial: boolean;
  rawPayload: Record<string, unknown>;

  // Markdown
  markdown?: string;
  files?: MarkdownFile[];

  // LLM
  data?: unknown;
  resolvedSchema?: Record<string, unknown>;
  notes?: string[];
  partialData?: unknown;

  /** @deprecated Use `rawPayload`. */
  raw?: Record<string, unknown>;
}

export function contextOutputFromDict(data: Record<string, unknown>): ContextOutput {
  const shape = (data.shape ?? data.type ?? 'raw') as string;
  const payload =
    typeof data.data === 'object' && data.data !== null
      ? (data.data as Record<string, unknown>)
      : {};
  const itemsData = (payload.items as Record<string, unknown>[] | undefined) ?? [];

  const out: ContextOutput = {
    shape: String(shape),
    items: itemsData.map(contextItemFromDict),
    partial: Boolean(data.partial ?? false),
    rawPayload: data,
    raw: data,
  };

  if (shape === 'markdown') {
    const mode = (payload.mode as string | undefined) ?? 'single';
    if (mode === 'multi') {
      const filesRaw = (payload.files as Record<string, unknown>[] | undefined) ?? [];
      out.files = filesRaw.map((f) => ({
        filename: String(f.filename ?? ''),
        markdown: String(f.markdown ?? ''),
      }));
    } else {
      out.markdown = payload.markdown as string | undefined;
    }
  } else if (shape === 'llm') {
    out.data = payload.data;
    out.resolvedSchema = (payload.resolved_schema as Record<string, unknown>) ?? {};
    out.notes = (payload.notes as string[]) ?? [];
    out.partialData = payload.partial;
  }

  return out;
}

// ─── Streaming event types ──────────────────────────────────────────────

export interface StatusEvent {
  type: 'status';
  status: string;
  phase?: string;
  version: number;
  planningMs: number;
  crawlingMs: number;
  shapingMs: number;
  ts?: string;
}

export interface PhaseProgressInit {
  type: 'phase_progress';
  kind: 'init';
  phase: string;
  total: number;
  items: Record<string, unknown>[];
}

export interface PhaseProgressItemUpdate {
  type: 'phase_progress';
  kind: 'item_update';
  phase: string;
  id: string;
  status: string;
  ms: number;
  size?: number;
  reason?: string;
}

export interface TerminalEvent {
  type: 'terminal';
  status: string;
  totalMs: number;
  urlsCrawled: number;
  urlsFailed: number;
  outputS3Key?: string;
  errorMessage?: string;
}

export type ContextEvent =
  | StatusEvent
  | PhaseProgressInit
  | PhaseProgressItemUpdate
  | TerminalEvent;

export function parseContextEvent(
  eventType: string,
  data: Record<string, unknown>,
): ContextEvent | null {
  const t = (data.type as string | undefined) ?? eventType;
  if (t === 'status') {
    return {
      type: 'status',
      status: String(data.status ?? ''),
      phase: data.phase as string | undefined,
      version: Number(data.version ?? 1),
      planningMs: Number(data.planning_ms ?? 0),
      crawlingMs: Number(data.crawling_ms ?? 0),
      shapingMs: Number(data.shaping_ms ?? 0),
      ts: data.ts as string | undefined,
    };
  }
  if (t === 'terminal') {
    return {
      type: 'terminal',
      status: String(data.status ?? ''),
      totalMs: Number(data.total_ms ?? 0),
      urlsCrawled: Number(data.urls_crawled ?? 0),
      urlsFailed: Number(data.urls_failed ?? 0),
      outputS3Key: data.output_s3_key as string | undefined,
      errorMessage: data.error_message as string | undefined,
    };
  }
  if (t === 'phase_progress') {
    const kind = (data.kind as string | undefined) ?? 'init';
    if (kind === 'init') {
      return {
        type: 'phase_progress',
        kind: 'init',
        phase: String(data.phase ?? 'fetch'),
        total: Number(data.total ?? 0),
        items: (data.items as Record<string, unknown>[] | undefined) ?? [],
      };
    }
    if (kind === 'item_update') {
      return {
        type: 'phase_progress',
        kind: 'item_update',
        phase: String(data.phase ?? 'fetch'),
        id: String(data.id ?? ''),
        status: String(data.status ?? ''),
        ms: Number(data.ms ?? 0),
        size: data.size as number | undefined,
        reason: data.reason as string | undefined,
      };
    }
  }
  return null;
}

// ─── Diff / Version / Catalog ───────────────────────────────────────────

export interface ContextVersion {
  version: number;
  status: string;
  submittedAt?: string;
  completedAt?: string;
  urlsCrawled: number;
  triggeredBy: string;
  outputS3Key?: string;
}

export function contextVersionFromDict(data: Record<string, unknown>): ContextVersion {
  return {
    version: Number(data.version ?? 1),
    status: String(data.status ?? ''),
    submittedAt: data.submitted_at as string | undefined,
    completedAt: data.completed_at as string | undefined,
    urlsCrawled: Number(data.urls_crawled ?? 0),
    triggeredBy: String(data.triggered_by ?? 'user'),
    outputS3Key: data.output_s3_key as string | undefined,
  };
}

export interface ContextDiff {
  added: ContextItem[];
  removed: ContextItem[];
  unchanged: ContextItem[];
  sourcesAdded: string[];
  sourcesRemoved: string[];
  raw: Record<string, unknown>;
}

export function contextDiffFromDict(data: Record<string, unknown>): ContextDiff {
  const arr = (k: string): Record<string, unknown>[] =>
    (data[k] as Record<string, unknown>[] | undefined) ?? [];
  return {
    added: arr('added').map(contextItemFromDict),
    removed: arr('removed').map(contextItemFromDict),
    unchanged: arr('unchanged').map(contextItemFromDict),
    sourcesAdded: (data.sources_added as string[] | undefined) ?? [],
    sourcesRemoved: (data.sources_removed as string[] | undefined) ?? [],
    raw: data,
  };
}

export interface CatalogEntry {
  name: string;
  displayName: string;
  summary: string;
  helpMd: string;
  paramsSchema: Record<string, unknown>;
}

export function catalogEntryFromDict(data: Record<string, unknown>): CatalogEntry {
  return {
    name: String(data.name ?? ''),
    displayName: String(data.display_name ?? data.name ?? ''),
    summary: String(data.summary ?? ''),
    helpMd: String(data.help_md ?? ''),
    paramsSchema:
      (data.params_schema as Record<string, unknown> | undefined) ??
      (data.query_params_schema as Record<string, unknown> | undefined) ??
      {},
  };
}

export interface ContextCatalog {
  sources: CatalogEntry[];
  strategies: CatalogEntry[];
  synthesizers: CatalogEntry[];
  reconcilers: CatalogEntry[];
  /** @deprecated Use `synthesizers`. */
  shapes?: CatalogEntry[];
}

// ─── ContextResult — the run state with lazy output ─────────────────────

export class ContextResult {
  runId: string;
  status: string;
  version: number;
  phase?: string;
  generatorId?: string;
  intent?: string;
  constraints: Record<string, unknown>;
  stats: Record<string, unknown>;
  errorMessage?: string;
  submittedAt?: string;
  completedAt?: string;

  private _crawler?: {
    getContextOutput: (runId: string) => Promise<ContextOutput>;
  };
  private _output?: ContextOutput;

  constructor(init: {
    runId: string;
    status: string;
    version: number;
    phase?: string;
    generatorId?: string;
    intent?: string;
    constraints?: Record<string, unknown>;
    stats?: Record<string, unknown>;
    errorMessage?: string;
    submittedAt?: string;
    completedAt?: string;
    crawler?: { getContextOutput: (runId: string) => Promise<ContextOutput> };
  }) {
    this.runId = init.runId;
    this.status = init.status;
    this.version = init.version;
    this.phase = init.phase;
    this.generatorId = init.generatorId;
    this.intent = init.intent;
    this.constraints = init.constraints ?? {};
    this.stats = init.stats ?? {};
    this.errorMessage = init.errorMessage;
    this.submittedAt = init.submittedAt;
    this.completedAt = init.completedAt;
    this._crawler = init.crawler;
  }

  get isTerminal(): boolean {
    return CONTEXT_TERMINAL_STATUSES.has(this.status);
  }

  get isSuccess(): boolean {
    return this.status === 'completed' || this.status === 'completed_partial';
  }

  /** Fetch the synthesized output. Cached after the first call. */
  async output(): Promise<ContextOutput> {
    if (this._output !== undefined) return this._output;
    if (this._crawler === undefined) {
      throw new Error(
        'ContextResult was built without a crawler reference; ' +
          'use crawler.getContextOutput(runId).',
      );
    }
    this._output = await this._crawler.getContextOutput(this.runId);
    return this._output;
  }
}

export function contextResultFromDict(
  data: Record<string, unknown>,
  crawler?: { getContextOutput: (runId: string) => Promise<ContextOutput> },
): ContextResult {
  const runId = (data.run_id ?? data.id ?? '') as string;
  const stats: Record<string, unknown> = {
    ...((data.stats as Record<string, unknown> | undefined) ?? {}),
  };
  for (const k of [
    'planning_ms', 'crawling_ms', 'shaping_ms', 'total_ms',
    'urls_crawled', 'urls_failed', 'output_size_bytes',
  ]) {
    if (data[k] !== undefined && data[k] !== null) stats[k] = data[k];
  }
  return new ContextResult({
    runId: String(runId),
    status: String(data.status ?? ''),
    version: Number(data.version ?? 1),
    phase: data.phase as string | undefined,
    generatorId: data.generator_id as string | undefined,
    intent: data.intent as string | undefined,
    constraints: (data.constraints as Record<string, unknown> | undefined) ?? {},
    stats,
    errorMessage: data.error_message as string | undefined,
    submittedAt:
      (data.submitted_at as string | undefined) ??
      (data.created_at as string | undefined),
    completedAt: data.completed_at as string | undefined,
    crawler,
  });
}
