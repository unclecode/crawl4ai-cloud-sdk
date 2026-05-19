/**
 * Context v2 — the four-pillar research pipeline.
 *
 * Ships:
 * - **Pillar builders** — typed factory methods (`Source.googleWeb(...)`,
 *   `Strategy.allItems()`, `Shape.raw()`, `Reconciler.noop()`) for pillars
 *   that ship today, plus a dict-passthrough escape hatch
 *   (`Source.custom({ type, params })`) for pillars that ship server-side
 *   before this SDK adds a typed builder.
 * - **Result + event types** — `ContextResult` for the run state (with
 *   lazy `output()` fetch), plus typed `StatusEvent`, `PhaseProgressInit`,
 *   `PhaseProgressItemUpdate`, and `TerminalEvent` for the streaming
 *   iterator.
 * - **Constants** — terminal statuses, phase names.
 *
 * The crawler methods (`context()`, `contextStream()`, `refreshContext()`,
 * etc.) live on `AsyncWebCrawler` in `crawler.ts`. This module is the
 * data layer.
 */

// ─── Constants ──────────────────────────────────────────────────────────

/** Terminal run statuses — stop polling on any of these. */
export const CONTEXT_TERMINAL_STATUSES = new Set<string>([
  'completed',
  'completed_partial',
  'failed',
  'cancelled',
]);

/** Non-terminal statuses — keep polling. */
export const CONTEXT_ACTIVE_STATUSES = new Set<string>(['queued', 'running']);

/** Pipeline phases. */
export const PHASE_PLANNING = 'planning';
export const PHASE_CRAWLING = 'crawling';
export const PHASE_SHAPING = 'shaping';
export const CONTEXT_PHASES = [
  PHASE_PLANNING,
  PHASE_CRAWLING,
  PHASE_SHAPING,
] as const;

// ─── Pillar config types ────────────────────────────────────────────────

/** Pillar config — the wire shape expected by `/v1/context`. */
export interface PillarConfig {
  type: string;
  params: Record<string, unknown>;
  auth_ref?: string;
}

// ─── Pillar builders ────────────────────────────────────────────────────

/**
 * Builder for Context Source configs.
 *
 * Each method returns a plain object in the shape the API expects:
 * `{ type: "<source_name>", params: {...}, auth_ref?: string }`.
 * Use `Source.custom({ type, params })` for pillars that ship server-side
 * before this SDK adds a typed builder.
 */
export class Source {
  /**
   * Google search across multiple SERP backends with RRF merge.
   *
   * @param opts.backends - Subset of `["google", "bing", "duckduckgo", "brave"]`.
   *   Defaults to `["google", "bing"]`.
   * @param opts.topKPerBackend - Per-backend cap before RRF merge (1-50).
   * @param opts.region - 2-letter country code (`"us"`, `"gb"`) biasing results.
   */
  static googleWeb(opts: {
    backends?: string[];
    topKPerBackend?: number;
    region?: string;
  } = {}): PillarConfig {
    const params: Record<string, unknown> = {
      top_k_per_backend: Math.trunc(opts.topKPerBackend ?? 10),
    };
    if (opts.backends !== undefined) {
      params.backends = [...opts.backends];
    }
    if (opts.region !== undefined) {
      params.region = String(opts.region);
    }
    return { type: 'google_web', params };
  }

  /**
   * Recursive site crawl as the corpus.
   */
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
    if (opts.profileId !== undefined) {
      params.profile_id = String(opts.profileId);
    }
    return { type: 'crawl', params };
  }

  /**
   * User-uploaded file as the corpus.
   */
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

  /**
   * Escape hatch for Sources that exist server-side but don't yet have a
   * typed builder in this SDK (e.g. `hackernews`, `github`, `rss` once
   * they ship). Discover available Sources via `crawler.contextCatalog()`.
   */
  static custom(opts: {
    type: string;
    params?: Record<string, unknown>;
    authRef?: string;
  }): PillarConfig {
    const out: PillarConfig = {
      type: String(opts.type),
      params: { ...(opts.params ?? {}) },
    };
    if (opts.authRef !== undefined) {
      out.auth_ref = String(opts.authRef);
    }
    return out;
  }
}

/** Builder for Context Strategy configs. */
export class Strategy {
  /** Passthrough — every candidate item is kept up to
   * `constraints.maxItems`. The default. */
  static allItems(): PillarConfig {
    return { type: 'all_items', params: {} };
  }

  /** Escape hatch for Strategies that ship server-side before this SDK
   * adds a typed builder. */
  static custom(opts: { type: string; params?: Record<string, unknown> }): PillarConfig {
    return {
      type: String(opts.type),
      params: { ...(opts.params ?? {}) },
    };
  }
}

/** Builder for Context Shape configs. */
export class Shape {
  /** Per-item citations with `url` provenance. The default. */
  static raw(): PillarConfig {
    return { type: 'raw', params: {} };
  }

  /** Escape hatch for Shapes that ship server-side before this SDK adds
   * a typed builder. */
  static custom(opts: { type: string; params?: Record<string, unknown> }): PillarConfig {
    return {
      type: String(opts.type),
      params: { ...(opts.params ?? {}) },
    };
  }
}

/** Builder for Context Reconciler configs. */
export class Reconciler {
  /** No auto-refresh. Refreshes are user-initiated via
   * `refreshContext()`. The default. */
  static noop(): PillarConfig {
    return { type: 'noop', params: {} };
  }

  /** Escape hatch for Reconcilers that ship server-side before this SDK
   * adds a typed builder (e.g. `cron`, `event`). */
  static custom(opts: { type: string; params?: Record<string, unknown> }): PillarConfig {
    return {
      type: String(opts.type),
      params: { ...(opts.params ?? {}) },
    };
  }
}

// ─── Constraints ────────────────────────────────────────────────────────

/**
 * Caller-controllable knobs forwarded to the Context pipeline.
 *
 * All fields have sensible defaults that match the API. Pass an instance
 * to `crawler.context({ constraints: ... })` or pass a plain object —
 * both work.
 */
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
  // Allow camelCase too — converted at the boundary.
  maxItems: number;
  maxPerSource: number;
  maxCrawlTimeS: number;
  freshnessDays: number;
}>;

/** Normalize a Constraints instance or a plain object into the wire dict. */
export function constraintsToDict(c: ConstraintsInput): Record<string, unknown> {
  if (c instanceof Constraints) return c.toDict();
  const raw = c as Record<string, unknown>;
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(raw)) {
    if (v === undefined || v === null) continue;
    // Pass snake_case keys through; map camelCase aliases.
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

/**
 * One fetched item — typically one URL the Source query phase surfaced and
 * the fetch phase materialised. For the `raw` Shape, each item is the
 * unit of citation: its `url` + `title` is the provenance, and `content`
 * / `snippet` is what the consumer reads.
 */
export interface ContextItem {
  url?: string;
  title?: string;
  content?: string;
  snippet?: string;
  /** The Source that produced this item, e.g. "google_web". */
  source?: string;
  relevance: number;
  metadata: Record<string, unknown>;
  id?: string;
  fetchedAt?: string;
}

export function contextItemFromDict(data: Record<string, unknown>): ContextItem {
  // The raw Shape returns `source_name`; the catalog calls it `source`.
  // Accept either so we don't break when shapes evolve.
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

/**
 * The shaped output.
 *
 * For the `raw` Shape (today), `items` carries the fetched URLs with
 * content + snippet + source + provenance metadata; each item is the
 * citation unit. For future Shapes (`markdown_digest`, `tabular`,
 * `knowledge_graph`) the top-level structure may differ — `raw` is
 * preserved unmodified at `.raw` so consumers can drop down to it when
 * the typed surface doesn't carry a needed field.
 */
export interface ContextOutput {
  shape: string;
  items: ContextItem[];
  partial: boolean;
  raw: Record<string, unknown>;
}

export function contextOutputFromDict(data: Record<string, unknown>): ContextOutput {
  // Wire shape today is {"type": "raw", "data": {"items": [...]}}.
  // `type` may evolve to `shape`, and `data` may be flattened — accept
  // both for forward compat.
  const shape = (data.shape ?? data.type ?? 'raw') as string;
  const payload =
    typeof data.data === 'object' && data.data !== null
      ? (data.data as Record<string, unknown>)
      : data;
  const itemsData = (payload as Record<string, unknown>).items as
    | Record<string, unknown>[]
    | undefined;
  return {
    shape: String(shape),
    items: (itemsData ?? []).map(contextItemFromDict),
    partial: Boolean(data.partial ?? false),
    raw: data,
  };
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

/** Translate a raw SSE `{event, data}` into a typed event. Returns null
 * for unknown event types (forward-compatible). */
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

/**
 * Diff between two Context versions.
 *
 * Today the diff is item-level (matched by stable URL). Future versions
 * may diff at the claim level once a Shape that emits discrete claims
 * (e.g. `markdown_digest`) is wired through. Until then `added` /
 * `removed` / `unchanged` are lists of `ContextItem`.
 */
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
  shapes: CatalogEntry[];
  reconcilers: CatalogEntry[];
}

// ─── ContextResult — the run state with lazy output ─────────────────────

/**
 * A Context run's state.
 *
 * `output()` is lazy — call `await result.output()` to fetch the shaped
 * output. The first call hits the API; subsequent calls return the
 * cached value.
 *
 * Refresh / cancel / diff / rollback / list-versions can also be called
 * via the matching crawler methods on the same runId.
 */
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

  // Set by the crawler when the result is built so output() can lazy-fetch.
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

  /** Fetch the shaped output for this run. Cached after the first call;
   * safe to call multiple times. */
  async output(): Promise<ContextOutput> {
    if (this._output !== undefined) return this._output;
    if (this._crawler === undefined) {
      throw new Error(
        'ContextResult was built without a crawler reference; ' +
          'cannot fetch output. Use crawler.getContextOutput(runId).',
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
  // GET /{run_id} returns the row with `id` as primary key; POST submit
  // returns `run_id`. Handle both.
  const runId = (data.run_id ?? data.id ?? '') as string;

  // Stats — newer rows surface flat *_ms fields; older API shape used a
  // nested `stats` dict. Fold both into one dict for callers.
  const stats: Record<string, unknown> = {
    ...((data.stats as Record<string, unknown> | undefined) ?? {}),
  };
  for (const k of [
    'planning_ms',
    'crawling_ms',
    'shaping_ms',
    'total_ms',
    'urls_crawled',
    'urls_failed',
    'output_size_bytes',
  ]) {
    if (data[k] !== undefined && data[k] !== null) {
      stats[k] = data[k];
    }
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
