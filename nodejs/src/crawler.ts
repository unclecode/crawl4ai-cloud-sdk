/**
 * AsyncWebCrawler - Main crawler class for Crawl4AI Cloud SDK.
 */

import { HTTPClient } from './client';
import { TimeoutError } from './errors';
import {
  CrawlResult,
  CrawlJob,
  DeepCrawlResult,
  ScanResult,
  ScanOptions,
  ScanJobStatus,
  SiteScanConfig,
  SiteExtractConfig,
  GeneratedSchema,
  StorageUsage,
  ProxyConfig,
  JobProgress,
  crawlResultFromDict,
  crawlJobFromDict,
  deepCrawlResultFromDict,
  scanResultFromDict,
  scanJobStatusFromDict,
  siteScanConfigToDict,
  siteExtractConfigToDict,
  isScanResultAsync,
  isScanJobComplete,
  generatedSchemaFromDict,
  storageUsageFromDict,
  isJobComplete,
  isDeepCrawlComplete,
  MarkdownResponse,
  ScreenshotResponse,
  ExtractResponse,
  MapResponse,
  SiteCrawlResponse,
  SiteCrawlJobStatus,
  WrapperJob,
  MarkdownOptions,
  MarkdownManyOptions,
  ScreenshotOptions,
  ScreenshotManyOptions,
  ExtractOptions,
  ExtractManyOptions,
  MapOptions,
  SiteCrawlOptions,
  markdownResponseFromDict,
  screenshotResponseFromDict,
  extractResponseFromDict,
  mapResponseFromDict,
  siteCrawlResponseFromDict,
  siteCrawlJobStatusFromDict,
  isSiteCrawlJobComplete,
  wrapperJobFromDict,
  isWrapperJobComplete,
  EnrichJobStatus,
  EnrichJobListItem,
  EnrichEntity,
  EnrichCriterion,
  EnrichFeature,
  EnrichEvent,
  EnrichStatus,
  enrichJobStatusFromDict,
  enrichJobListItemFromDict,
  enrichEventFromDict,
  isEnrichJobComplete,
  ENRICH_PAUSED_STATUSES,
  // Discovery / Search
  SearchResponse,
  DiscoveryService,
  DiscoveryJobHandle,
  DiscoveryJobStatus,
  discoveryJobHandleFromDict,
  discoveryJobStatusFromDict,
  searchResponseFromDict,
  discoveryServiceFromDict,
} from './models';
import {
  CrawlerRunConfig,
  BrowserConfig,
  buildCrawlRequest,
  sanitizeCrawlerConfig,
  sanitizeBrowserConfig,
  normalizeProxy,
} from './configs';
import {
  PillarConfig,
  ConstraintsInput,
  Constraints,
  constraintsToDict,
  ContextResult,
  contextResultFromDict,
  ContextOutput,
  contextOutputFromDict,
  ContextEvent,
  parseContextEvent,
  ContextVersion,
  contextVersionFromDict,
  ContextDiff,
  contextDiffFromDict,
  ContextCatalog,
  CatalogEntry,
  catalogEntryFromDict,
} from './context';


// ─── Context pipeline body builder ───────────────────────────────────────

function splitPillar(cfg: PillarConfig): { name: string; params: Record<string, unknown> } {
  if (cfg === null || typeof cfg !== 'object' || typeof cfg.type !== 'string') {
    throw new Error(
      'pillar must be a builder dict produced by Strategy/Synthesizer/Reconciler ' +
        `(got ${JSON.stringify(cfg)})`,
    );
  }
  return { name: cfg.type, params: { ...(cfg.params ?? {}) } };
}

function buildPipeline(args: {
  sources: PillarConfig[];
  strategy?: PillarConfig;
  synthesizer?: PillarConfig;
  reconciler?: PillarConfig;
}): Record<string, unknown> {
  if (!args.sources || args.sources.length === 0) {
    throw new Error('inline pipeline requires at least one source');
  }
  const out: Record<string, unknown> = {
    sources: args.sources.map((s) => ({ ...s })),
  };
  if (args.strategy !== undefined) {
    const { name, params } = splitPillar(args.strategy);
    out.strategy = name;
    out.strategy_params = params;
  }
  if (args.synthesizer !== undefined) {
    const { name, params } = splitPillar(args.synthesizer);
    out.synthesizer = name;
    out.synthesizer_params = params;
  }
  if (args.reconciler !== undefined) {
    const { name, params } = splitPillar(args.reconciler);
    out.reconciler = name;
    out.reconciler_params = params;
  }
  return out;
}


export interface AsyncWebCrawlerOptions {
  apiKey?: string;
  baseUrl?: string;
  timeout?: number;
  maxRetries?: number;
  // OSS compatibility - ignored
  verbose?: boolean;
}

export interface RunOptions {
  config?: CrawlerRunConfig;
  browserConfig?: BrowserConfig;
  strategy?: 'browser' | 'http';
  proxy?: string | ProxyConfig | Record<string, unknown>;
  bypassCache?: boolean;
  [key: string]: unknown;
}

export interface RunManyOptions extends RunOptions {
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
  priority?: number;
  webhookUrl?: string;
}

export interface SiteOptions {
  /** 'map' = sync sitemap discovery, 'traverse' = async recursive crawl across the worker pool. Default: 'traverse'. */
  mode?: 'map' | 'traverse';
  maxUrls?: number;
  maxDepth?: number;
  scanOnly?: boolean;
  patterns?: string[];
  filters?: Record<string, unknown>;
  scorers?: Record<string, unknown>;
  query?: string;
  scoreThreshold?: number;
  includeSubdomains?: boolean;
  crawlerConfig?: CrawlerRunConfig | Record<string, unknown>;
  browserConfig?: BrowserConfig | Record<string, unknown>;
  proxy?: string | ProxyConfig | Record<string, unknown>;
  webhookUrl?: string;
  priority?: number;
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
}

export interface DeepCrawlOptions {
  sourceJob?: string;
  strategy?: 'bfs' | 'dfs' | 'best_first' | 'map';
  maxDepth?: number;
  maxUrls?: number;
  scanOnly?: boolean;
  config?: CrawlerRunConfig;
  browserConfig?: BrowserConfig;
  crawlStrategy?: 'browser' | 'http' | 'auto';
  proxy?: string | ProxyConfig | Record<string, unknown>;
  bypassCache?: boolean;
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
  filters?: Record<string, unknown>;
  scorers?: Record<string, unknown>;
  includeHtml?: boolean;
  webhookUrl?: string;
  priority?: number;
  // Map strategy options
  source?: string;
  pattern?: string;
  query?: string;
  scoreThreshold?: number;
  // URL filtering shortcuts
  includePatterns?: string[];
  excludePatterns?: string[];
}

/**
 * Options for `crawler.enrich(...)` — multi-phase enrichment job.
 *
 * Pass any subset of `query`, `entities`, or `urls` (at least one required).
 * Strings in `entities` / `criteria` / `features` are auto-wrapped as
 * `{name: str}` / `{text: str}`.
 */
export interface EnrichOptions {
  // Inputs
  query?: string;
  entities?: Array<string | EnrichEntity>;
  criteria?: Array<string | EnrichCriterion>;
  features?: Array<string | EnrichFeature>;
  urls?: string[];
  groups?: Record<string, string[]>;
  // Phase control
  autoConfirmPlan?: boolean;
  autoConfirmUrls?: boolean;
  // Discover knobs
  topKPerEntity?: number;
  search?: boolean;
  country?: string;
  locationHint?: string;
  // Standard wrapper knobs
  strategy?: 'browser' | 'http';
  config?: Record<string, unknown>;
  browserConfig?: Record<string, unknown>;
  crawlerConfig?: Record<string, unknown>;
  llmConfig?: Record<string, unknown>;
  proxy?: string | ProxyConfig | Record<string, unknown>;
  webhookUrl?: string;
  priority?: number;
  // Polling
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
}

/** Edits to apply when resuming a paused enrich job. */
export interface ResumeEnrichOptions {
  entities?: Array<string | EnrichEntity>;
  criteria?: Array<string | EnrichCriterion>;
  features?: Array<string | EnrichFeature>;
  groups?: Record<string, string[]>;
}

/** Options for `crawler.waitEnrichJob(jobId, options)`. */
export interface WaitEnrichOptions {
  until?: EnrichStatus;
  pollInterval?: number;
  timeout?: number;
}

// ─── Enrich vocabulary normalizers (string shortcuts) ────────────────

function normalizeEntity(item: string | EnrichEntity): Record<string, unknown> {
  if (typeof item === 'string') return { name: item };
  const out: Record<string, unknown> = { name: item.name };
  if (item.title !== undefined) out.title = item.title;
  if (item.sourceUrl !== undefined) out.source_url = item.sourceUrl;
  return out;
}

function normalizeCriterion(item: string | EnrichCriterion): Record<string, unknown> {
  if (typeof item === 'string') return { text: item };
  const out: Record<string, unknown> = { text: item.text };
  if (item.kind !== undefined) out.kind = item.kind;
  return out;
}

function normalizeFeature(item: string | EnrichFeature): Record<string, unknown> {
  if (typeof item === 'string') return { name: item };
  const out: Record<string, unknown> = { name: item.name };
  if (item.description !== undefined) out.description = item.description;
  return out;
}

/**
 * Async client for Crawl4AI Cloud API.
 *
 * Mirrors the OSS AsyncWebCrawler API for seamless migration.
 * Just change your import and add an API key.
 *
 * @example
 * ```typescript
 * import { AsyncWebCrawler } from 'crawl4ai-cloud';
 *
 * const crawler = new AsyncWebCrawler({ apiKey: 'sk_live_xxx' });
 * const result = await crawler.run('https://example.com');
 * console.log(result.markdown?.rawMarkdown);
 * await crawler.close();
 * ```
 */
export class AsyncWebCrawler {
  private http: HTTPClient;

  constructor(options: AsyncWebCrawlerOptions = {}) {
    this.http = new HTTPClient({
      apiKey: options.apiKey,
      baseUrl: options.baseUrl,
      timeout: options.timeout,
      maxRetries: options.maxRetries,
    });
  }

  // -------------------------------------------------------------------------
  // Core Crawl Methods
  // -------------------------------------------------------------------------

  /**
   * Crawl a single URL.
   */
  async run(url: string, options: RunOptions = {}): Promise<CrawlResult> {
    const {
      config,
      browserConfig,
      strategy = 'browser',
      proxy,
      bypassCache = false,
      ...rest
    } = options;

    const body = buildCrawlRequest({
      url,
      config,
      browserConfig,
      strategy,
      proxy,
      bypassCache,
      ...rest,
    });

    const data = await this.http.post('/v1/crawl', body, 120000);
    return crawlResultFromDict(data);
  }

  /**
   * Crawl a single URL (OSS compatibility alias for run()).
   */
  async arun(url: string, options: RunOptions = {}): Promise<CrawlResult> {
    return this.run(url, options);
  }

  /**
   * Crawl multiple URLs.
   *
   * Creates an async job for processing. Use wait=true to block until
   * complete, or poll with getJob()/waitJob().
   */
  async runMany(
    urls: string[],
    options: RunManyOptions = {}
  ): Promise<CrawlJob | CrawlResult[]> {
    const {
      wait = false,
      pollInterval = 2.0,
      timeout,
      priority = 5,
      webhookUrl,
      ...runOptions
    } = options;

    // Always use async endpoint for consistent job tracking
    return this.runAsync(urls, {
      ...runOptions,
      wait,
      pollInterval,
      timeout,
      priority,
      webhookUrl,
    });
  }

  /**
   * Crawl multiple URLs (OSS compatibility alias for runMany()).
   */
  async arunMany(
    urls: string[],
    options: RunManyOptions = {}
  ): Promise<CrawlJob | CrawlResult[]> {
    return this.runMany(urls, options);
  }

  private async runAsync(
    urls: string[],
    options: RunManyOptions
  ): Promise<CrawlJob | CrawlResult[]> {
    const {
      config,
      browserConfig,
      strategy = 'browser',
      proxy,
      bypassCache = false,
      wait = false,
      pollInterval = 2.0,
      timeout,
      priority = 5,
      webhookUrl,
      ...rest
    } = options;

    const body = buildCrawlRequest({
      urls,
      config,
      browserConfig,
      strategy,
      proxy,
      bypassCache,
      priority,
      ...rest,
    });

    if (webhookUrl) {
      body.webhook_url = webhookUrl;
    }

    const data = await this.http.post('/v1/crawl/async', body);
    let job = crawlJobFromDict(data);

    if (wait) {
      job = await this.waitJob(job.jobId, {
        pollInterval,
        timeout,
      });
      // Results are available via downloadUrl() after job completes
    }

    return job;
  }

  // -------------------------------------------------------------------------
  // Job Management
  // -------------------------------------------------------------------------

  /**
   * Get job status.
   * To get results, use downloadUrl() to get a presigned URL for the ZIP file.
   */
  async getJob(jobId: string): Promise<CrawlJob> {
    const data = await this.http.get(`/v1/crawl/jobs/${jobId}`);
    return crawlJobFromDict(data);
  }

  /**
   * Poll until job completes.
   * To get results after job completes, use downloadUrl() to get a presigned URL for the ZIP file.
   */
  async waitJob(
    jobId: string,
    options: {
      pollInterval?: number;
      timeout?: number;
    } = {}
  ): Promise<CrawlJob> {
    const { pollInterval = 2.0, timeout } = options;
    const startTime = Date.now();

    while (true) {
      const job = await this.getJob(jobId);

      if (isJobComplete(job)) {
        return job;
      }

      if (timeout && Date.now() - startTime > timeout * 1000) {
        throw new TimeoutError(
          `Timeout waiting for job ${jobId}. ` +
            `Status: ${job.status}, Progress: ${job.progress.completed}/${job.progress.total}`
        );
      }

      await this.sleep(pollInterval * 1000);
    }
  }

  /**
   * List jobs with optional filtering.
   */
  async listJobs(options: {
    status?: string;
    limit?: number;
    offset?: number;
  } = {}): Promise<CrawlJob[]> {
    const { status, limit = 20, offset = 0 } = options;
    const params: Record<string, string | number> = { limit, offset };
    if (status) {
      params.status = status;
    }

    const data = await this.http.get('/v1/crawl/jobs', params);
    return ((data.jobs || []) as Record<string, unknown>[]).map((job) => crawlJobFromDict(job));
  }

  /**
   * Cancel a pending or running job.
   */
  async cancelJob(jobId: string): Promise<boolean> {
    await this.http.delete(`/v1/crawl/jobs/${jobId}`);
    return true;
  }

  /**
   * Get presigned URL for downloading job results.
   */
  async downloadUrl(jobId: string, expiresIn = 3600): Promise<string> {
    const data = await this.http.get(`/v1/crawl/jobs/${jobId}/download`, {
      expires_in: expiresIn,
    });
    return data.download_url as string;
  }

  // -------------------------------------------------------------------------
  // Site (canonical) + Deep Crawl (back-compat alias)
  // -------------------------------------------------------------------------

  /**
   * Discover or crawl an entire site (canonical, /v1/site).
   *
   * `mode='map'` runs sync sitemap-based URL discovery. `mode='traverse'`
   * (default) returns a `jobId` immediately and runs best-first recursive
   * crawl across the worker pool via the cloud's WorkerPoolDispatcher.
   */
  async site(url: string, options: SiteOptions = {}): Promise<DeepCrawlResult> {
    const {
      mode = 'traverse',
      maxUrls,
      maxDepth,
      scanOnly = false,
      patterns,
      filters,
      scorers,
      query,
      scoreThreshold,
      includeSubdomains = false,
      crawlerConfig,
      browserConfig,
      proxy,
      webhookUrl,
      priority = 5,
      wait = false,
      pollInterval = 2.0,
      timeout,
    } = options;

    const body: Record<string, unknown> = {
      url,
      mode,
      scan_only: scanOnly,
      include_subdomains: includeSubdomains,
      priority,
    };
    if (maxUrls !== undefined) body.max_urls = maxUrls;
    if (maxDepth !== undefined) body.max_depth = maxDepth;
    if (patterns) body.patterns = patterns;
    if (filters) body.filters = filters;
    if (scorers) body.scorers = scorers;
    if (query) body.query = query;
    if (scoreThreshold !== undefined) body.score_threshold = scoreThreshold;
    if (crawlerConfig) {
      const cc = sanitizeCrawlerConfig(crawlerConfig as CrawlerRunConfig);
      if (Object.keys(cc).length > 0) body.crawler_config = cc;
    }
    if (browserConfig) {
      const bc = sanitizeBrowserConfig(browserConfig as BrowserConfig);
      if (Object.keys(bc).length > 0) body.browser_config = bc;
    }
    const proxyConfig = normalizeProxy(proxy);
    if (proxyConfig) body.proxy = proxyConfig;
    if (webhookUrl) body.webhook_url = webhookUrl;

    const data = await this.http.post('/v1/site', body, 120000);
    let result = deepCrawlResultFromDict(data);

    if (!wait || result.status === 'completed' || result.status === 'cancelled' || result.status === 'failed') {
      return result;
    }

    // Traverse mode — wait for scan, then crawl phase if any
    result = await this.waitSiteJob(result.jobId, pollInterval, timeout);
    if (scanOnly || !result.crawlJobId) return result;
    return await this.waitJob(result.crawlJobId, { pollInterval, timeout }) as unknown as DeepCrawlResult;
  }

  private async waitSiteJob(
    jobId: string,
    pollInterval: number = 2.0,
    timeout?: number,
  ): Promise<DeepCrawlResult> {
    const start = Date.now();
    while (true) {
      const data = await this.http.get(`/v1/site/jobs/${jobId}`);
      const result = deepCrawlResultFromDict(data);
      if (['completed', 'cancelled', 'failed'].includes(result.status)) return result;
      if (timeout && (Date.now() - start) / 1000 > timeout) {
        throw new Error(
          `Timeout waiting for site job ${jobId}. Status: ${result.status}, Discovered: ${result.discoveredCount}`
        );
      }
      await new Promise(r => setTimeout(r, pollInterval * 1000));
    }
  }

  /** Cancel a running site (traverse) job. */
  async cancelSite(jobId: string): Promise<DeepCrawlResult> {
    const data = await this.http.post(`/v1/site/jobs/${jobId}/cancel`, {});
    return deepCrawlResultFromDict(data);
  }

  /** Get the status of a site (traverse) job. */
  async getSiteStatus(jobId: string): Promise<DeepCrawlResult> {
    const data = await this.http.get(`/v1/site/jobs/${jobId}`);
    return deepCrawlResultFromDict(data);
  }

  /**
   * Deep crawl - discover and crawl URLs from a starting point.
   *
   * `/v1/crawl/deep` is now a server-side alias for `/v1/site` (Phase 4).
   * `crawler.deepCrawl()` is kept as a back-compat alias — no warning.
   * New code should call `crawler.site()` directly.
   */
  async deepCrawl(
    url?: string,
    options: DeepCrawlOptions = {}
  ): Promise<DeepCrawlResult | CrawlJob> {
    const {
      sourceJob,
      strategy = 'bfs',
      maxDepth = 3,
      maxUrls = 100,
      scanOnly = false,
      config,
      browserConfig,
      crawlStrategy = 'auto',
      proxy,
      bypassCache = false,
      wait = false,
      pollInterval = 2.0,
      timeout,
      filters,
      scorers,
      includeHtml = false,
      webhookUrl,
      priority = 5,
      source = 'sitemap',
      pattern = '*',
      query,
      scoreThreshold,
      includePatterns,
      excludePatterns,
    } = options;

    if (!url && !sourceJob) {
      throw new Error("Must provide either 'url' or 'sourceJob'");
    }
    if (url && sourceJob) {
      throw new Error("Provide either 'url' or 'sourceJob', not both");
    }

    // Build request body
    const body: Record<string, unknown> = {};

    if (sourceJob) {
      // Phase 2: extraction from cached HTML — only send source_job_id
      body.source_job_id = sourceJob;
    } else {
      // Phase 1: URL-based discovery — include scan parameters
      body.url = url;
      body.strategy = strategy;
      body.crawl_strategy = crawlStrategy;
      body.priority = priority;

      // Tree strategy options
      if (['bfs', 'dfs', 'best_first'].includes(strategy)) {
        body.max_depth = maxDepth;
        body.max_urls = maxUrls;

        // Build filters from includePatterns/excludePatterns or use provided filters
        const effectiveFilters: Record<string, unknown> = filters ? { ...filters } : {};
        if (includePatterns) effectiveFilters.include_patterns = includePatterns;
        if (excludePatterns) effectiveFilters.exclude_patterns = excludePatterns;
        if (Object.keys(effectiveFilters).length > 0) body.filters = effectiveFilters;

        if (scorers) body.scorers = scorers;
        if (scanOnly) body.scan_only = true;
        if (includeHtml) body.include_html = true;
      }

      // Map strategy options
      if (strategy === 'map') {
        const seedingConfig: Record<string, unknown> = {
          source,
          pattern,
        };
        if (maxUrls) seedingConfig.max_urls = maxUrls;
        if (query) seedingConfig.query = query;
        if (scoreThreshold !== undefined) seedingConfig.score_threshold = scoreThreshold;
        body.seeding_config = seedingConfig;
      }
    }

    // Add configs
    const crawlerConfig = sanitizeCrawlerConfig(config);
    if (Object.keys(crawlerConfig).length > 0) {
      body.crawler_config = crawlerConfig;
    }

    const browserCfg = sanitizeBrowserConfig(browserConfig, crawlStrategy);
    if (Object.keys(browserCfg).length > 0) {
      body.browser_config = browserCfg;
    }

    // Proxy
    const proxyConfig = normalizeProxy(proxy);
    if (proxyConfig) body.proxy = proxyConfig;

    if (bypassCache) body.bypass_cache = true;
    if (webhookUrl) body.webhook_url = webhookUrl;

    const data = await this.http.post('/v1/crawl/deep', body, 120000);
    let result = deepCrawlResultFromDict(data);

    if (!wait) {
      return result;
    }

    // Wait for scan to complete
    result = await this.waitScanJob(result.jobId, pollInterval, timeout);

    if (scanOnly) {
      return result;
    }

    if (result.status === 'no_urls' || result.discoveredCount === 0) {
      return result;
    }

    // If crawl job was created, wait for it
    if (result.crawlJobId) {
      return this.waitJob(result.crawlJobId, {
        pollInterval,
        timeout,
      });
    }

    return result;
  }

  private async waitScanJob(
    jobId: string,
    pollInterval = 2.0,
    timeout?: number
  ): Promise<DeepCrawlResult> {
    const startTime = Date.now();

    while (true) {
      const data = await this.http.get(`/v1/crawl/deep/jobs/${jobId}`);
      const result = deepCrawlResultFromDict(data);

      if (isDeepCrawlComplete(result)) {
        return result;
      }

      if (timeout && Date.now() - startTime > timeout * 1000) {
        throw new TimeoutError(
          `Timeout waiting for scan job ${jobId}. ` +
            `Status: ${result.status}, Discovered: ${result.discoveredCount}`
        );
      }

      await this.sleep(pollInterval * 1000);
    }
  }

  /**
   * Cancel a running deep crawl job.
   *
   * The crawl will stop at the next batch boundary, preserving any
   * partial results that have been collected so far.
   *
   * @param jobId - Deep crawl job ID (scan_xxx format)
   * @returns DeepCrawlResult with status "cancelled" and partial results
   *
   * @example
   * ```typescript
   * // Start deep crawl without waiting
   * const result = await crawler.deepCrawl('https://docs.example.com', {
   *   maxUrls: 500,
   *   wait: false,
   * });
   *
   * // Cancel after some time
   * await new Promise(r => setTimeout(r, 10000));
   * const cancelled = await crawler.cancelDeepCrawl(result.jobId);
   * console.log(`Cancelled with ${cancelled.discoveredCount} partial results`);
   * ```
   */
  async cancelDeepCrawl(jobId: string): Promise<DeepCrawlResult> {
    const data = await this.http.post(`/v1/crawl/deep/jobs/${jobId}/cancel`, {});
    return deepCrawlResultFromDict(data);
  }

  /**
   * Get the status of a deep crawl job.
   *
   * @param jobId - Deep crawl job ID (scan_xxx format)
   * @returns DeepCrawlResult with current status and discovered URLs
   */
  async getDeepCrawlStatus(jobId: string): Promise<DeepCrawlResult> {
    const data = await this.http.get(`/v1/crawl/deep/jobs/${jobId}`);
    return deepCrawlResultFromDict(data);
  }

  // -------------------------------------------------------------------------
  // Scan API (URL Discovery)
  // -------------------------------------------------------------------------

  /**
   * Discover all URLs under a domain. AI-assisted via `criteria`, with
   * optional async deep-mode traversal.
   *
   * Two routing strategies (picked by `scan.mode`):
   * - **map** (sync): DomainMapper — sitemap + CC + wayback etc. Returns
   *   URLs inline. 2-60s. Cached 7 days.
   * - **deep** (async): best-first tree traversal. Returns a `jobId`;
   *   poll with `getScanJob()` or pass `wait=true`.
   *
   * @example
   * ```typescript
   * // AI-assisted (map mode picked by LLM)
   * const result = await crawler.scan('https://docs.crawl4ai.com', {
   *   criteria: 'API reference pages',
   *   maxUrls: 50,
   * });
   * console.log(`Mode: ${result.modeUsed}, found ${result.totalUrls} URLs`);
   *
   * // Explicit deep scan with waiting
   * const result = await crawler.scan('https://directory.example.com', {
   *   criteria: 'company profile pages',
   *   scan: { mode: 'deep', maxDepth: 3 },
   *   wait: true,
   *   pollInterval: 3,
   * });
   * ```
   */
  async scan(
    url: string,
    options: ScanOptions = {},
  ): Promise<ScanResult> {
    let { sources = 'primary' } = options;
    const {
      mode,
      maxUrls,
      includeSubdomains = true,
      extractHead = true,
      soft404Detection = true,
      query,
      scoreThreshold,
      force = false,
      probeThreshold,
      criteria,
      scan,
      wait = false,
      pollInterval = 2.0,
      timeout,
    } = options;

    if (mode === 'deep') {
      throw new Error(
        "scan(mode='deep') is no longer supported — recursive URL discovery " +
        "moved to crawler.site(url, { mode: 'traverse', scanOnly: true, ... }). " +
        "For one-shot sitemap discovery, drop the mode option and call scan(url) directly."
      );
    }
    if (mode !== undefined) {
      if (typeof process !== 'undefined' && process.emitWarning) {
        process.emitWarning('scan(mode) is deprecated — use sources ("primary" | "extended"). Will be removed in 0.10.0.', 'DeprecationWarning');
      }
      sources = 'primary';
    }

    const body: Record<string, unknown> = {
      url,
      sources,
      include_subdomains: includeSubdomains,
      extract_head: extractHead,
      soft_404_detection: soft404Detection,
      force,
    };
    if (criteria) body.criteria = criteria;
    if (scan !== undefined) {
      // Accept either a typed SiteScanConfig or a raw snake_case dict
      body.scan = this.isSiteScanConfig(scan) ? siteScanConfigToDict(scan) : scan;
    }
    if (maxUrls !== undefined) body.max_urls = maxUrls;
    if (query) body.query = query;
    if (scoreThreshold !== undefined) body.score_threshold = scoreThreshold;
    if (probeThreshold !== undefined) body.probe_threshold = probeThreshold;

    // LLM config generation can take a while on the initial POST.
    const data = await this.http.post('/v1/scan', body, 180000);
    const result = scanResultFromDict(data);

    // If the LLM picked deep mode (or caller forced it), optionally block
    // until the scan job finishes, merging final state back onto the result.
    if (wait && isScanResultAsync(result) && result.jobId) {
      const final = await this.waitScanJobV2(result.jobId, pollInterval, timeout);
      result.status = final.status;
      result.totalUrls = final.totalUrls;
      result.urls = final.urls;
      result.durationMs = final.durationMs;
      if (final.error) result.error = final.error;
    }

    return result;
  }

  /**
   * Poll a deep scan job started via `scan(url, { scan: { mode: 'deep' } })`.
   * Returns current discovered URLs, progress, and status.
   */
  async getScanJob(jobId: string): Promise<ScanJobStatus> {
    const data = await this.http.get(`/v1/scan/jobs/${jobId}`);
    return scanJobStatusFromDict(data);
  }

  /**
   * Cancel a running deep scan. Cancellation happens at the next batch
   * boundary — partial results (URLs discovered so far) are preserved.
   */
  async cancelScanJob(jobId: string): Promise<ScanJobStatus> {
    const data = await this.http.post(`/v1/scan/jobs/${jobId}/cancel`, {});
    return scanJobStatusFromDict(data);
  }

  private async waitScanJobV2(
    jobId: string,
    pollInterval: number = 2.0,
    timeout?: number,
  ): Promise<ScanJobStatus> {
    const start = Date.now();
    while (true) {
      const job = await this.getScanJob(jobId);
      if (isScanJobComplete(job)) return job;
      if (timeout && Date.now() - start > timeout * 1000) {
        throw new TimeoutError(
          `Timeout waiting for scan job ${jobId}. ` +
            `Status: ${job.status}, found: ${job.totalUrls}`,
        );
      }
      await this.sleep(pollInterval * 1000);
    }
  }

  private isSiteScanConfig(
    value: SiteScanConfig | Record<string, unknown>,
  ): value is SiteScanConfig {
    // Heuristic: camelCase keys unique to SiteScanConfig indicate the typed variant.
    // Raw dicts use snake_case (score_threshold, include_subdomains, max_depth).
    if (!value || typeof value !== 'object') return false;
    return (
      'scoreThreshold' in value ||
      'includeSubdomains' in value ||
      'maxDepth' in value ||
      // Otherwise fall back to "no snake_case keys present" means typed.
      !(
        'score_threshold' in value ||
        'include_subdomains' in value ||
        'max_depth' in value
      )
    );
  }

  private isSiteExtractConfig(
    value: SiteExtractConfig | Record<string, unknown>,
  ): value is SiteExtractConfig {
    if (!value || typeof value !== 'object') return false;
    return (
      'jsonExample' in value ||
      'sampleUrl' in value ||
      'urlPattern' in value ||
      !(
        'json_example' in value ||
        'sample_url' in value ||
        'url_pattern' in value
      )
    );
  }

  // =========================================================================
  // Context v2 — four-pillar research pipeline
  // =========================================================================
  // See `crawl4ai-cloud/src/context.ts` for pillar builders (Source / Strategy
  // / Shape / Reconciler) and the typed event/result types. See the public
  // walkthrough at /docs/guides/context-walkthrough for the worked example.

  /**
   * Compose the POST /v1/context body.
   *
   * Two pipeline modes:
   *   - `generatorId` — reference a saved generator
   *   - inline pillars — pass `sources` / `strategy` / `synthesizer` /
   *     `reconciler` directly; the API runs them as an un-persisted snapshot
   *
   * The two are mutually exclusive. If neither is set, the user's
   * default generator is used.
   */
  private _buildContextBody(args: {
    intent: string;
    mission?: string;
    generatorId?: string;
    sources?: PillarConfig[];
    strategy?: PillarConfig;
    synthesizer?: PillarConfig;
    shape?: PillarConfig;
    reconciler?: PillarConfig;
    constraints?: ConstraintsInput;
    webhookUrl?: string;
  }): Record<string, unknown> {
    // `shape` is the deprecated alias for `synthesizer`. Prefer the new key.
    const synthesizer = args.synthesizer ?? args.shape;

    const hasPillars =
      args.sources !== undefined ||
      args.strategy !== undefined ||
      synthesizer !== undefined ||
      args.reconciler !== undefined;

    if (args.generatorId !== undefined && hasPillars) {
      throw new Error(
        'Pass either `generatorId` OR pillar params ' +
          '(sources/strategy/synthesizer/reconciler), not both.',
      );
    }
    if (hasPillars && (!args.sources || args.sources.length === 0)) {
      throw new Error(
        'Inline pipelines must include at least one Source — ' +
          'pass `sources: [Source.googleWeb(), ...]`.',
      );
    }

    const body: Record<string, unknown> = { intent: String(args.intent) };
    if (args.mission) body.mission = String(args.mission);
    if (args.generatorId) body.generator_id = String(args.generatorId);
    if (hasPillars) {
      body.pipeline = buildPipeline({
        sources: args.sources!,
        strategy: args.strategy,
        synthesizer,
        reconciler: args.reconciler,
      });
    }
    if (args.constraints !== undefined) {
      body.constraints = constraintsToDict(args.constraints);
    }
    if (args.webhookUrl) body.webhook_url = String(args.webhookUrl);
    return body;
  }

  /**
   * Submit a Context run.
   *
   * One-liner — uses the user's default generator (`google_web` +
   * `all_items` + `raw` + `noop`):
   *
   * ```ts
   * const result = await crawler.context({
   *   intent: 'compare LangChain and AutoGen',
   * });
   * for (const item of (await result.output()).items) {
   *   console.log(item.title, '—', item.url);
   * }
   * ```
   *
   * Pillar params are reserved for the day public generator CRUD ships on
   * the API-key surface. Until then, custom pillars must be wrapped in a
   * saved generator (created via the dashboard); pass its `generatorId`.
   */
  async context(opts: {
    intent: string;
    mission?: string;
    generatorId?: string;
    sources?: PillarConfig[];
    strategy?: PillarConfig;
    synthesizer?: PillarConfig;
    /** @deprecated Use `synthesizer`. */
    shape?: PillarConfig;
    reconciler?: PillarConfig;
    constraints?: ConstraintsInput;
    webhookUrl?: string;
    idempotencyKey?: string;
    wait?: boolean;
    pollIntervalMs?: number;
    timeoutMs?: number;
  }): Promise<ContextResult> {
    if (!opts.intent || !opts.intent.trim()) {
      throw new Error('`intent` is required and must be non-empty');
    }
    const wait = opts.wait !== false;
    const pollIntervalMs = opts.pollIntervalMs ?? 3000;
    const timeoutMs = opts.timeoutMs ?? 600_000;

    const body = this._buildContextBody(opts);
    const headers: Record<string, string> = {};
    if (opts.idempotencyKey) headers['Idempotency-Key'] = String(opts.idempotencyKey);

    const data = await this.http.post(
      '/v1/context',
      body,
      30000,
      Object.keys(headers).length ? headers : undefined,
    );
    const runId = String(data.run_id ?? '');
    const result = await this.getContextRun(runId);

    if (!wait || result.isTerminal) return result;

    return this._waitContextRun(runId, { pollIntervalMs, timeoutMs });
  }

  /**
   * Submit (or attach to) a Context run and yield typed events.
   *
   * Two modes:
   * 1. Submit + stream — pass `intent` (and optional pillar params).
   *    Submits the run, opens the SSE stream, yields typed events until
   *    terminal.
   * 2. Attach — pass `runId` only. Opens the SSE stream on an existing run.
   */
  async *contextStream(opts: {
    intent?: string;
    runId?: string;
    mission?: string;
    generatorId?: string;
    sources?: PillarConfig[];
    strategy?: PillarConfig;
    synthesizer?: PillarConfig;
    /** @deprecated Use `synthesizer`. */
    shape?: PillarConfig;
    reconciler?: PillarConfig;
    constraints?: ConstraintsInput;
    webhookUrl?: string;
    idempotencyKey?: string;
  }): AsyncGenerator<ContextEvent> {
    let runId = opts.runId;
    if (runId === undefined) {
      if (!opts.intent) {
        throw new Error(
          'Pass `intent` to submit + stream, or `runId` to attach to an existing run.',
        );
      }
      const body = this._buildContextBody({
        intent: opts.intent,
        mission: opts.mission,
        generatorId: opts.generatorId,
        sources: opts.sources,
        strategy: opts.strategy,
        synthesizer: opts.synthesizer,
        shape: opts.shape,
        reconciler: opts.reconciler,
        constraints: opts.constraints,
        webhookUrl: opts.webhookUrl,
      });
      const headers: Record<string, string> = {};
      if (opts.idempotencyKey) headers['Idempotency-Key'] = String(opts.idempotencyKey);
      const submit = await this.http.post(
        '/v1/context',
        body,
        30000,
        Object.keys(headers).length ? headers : undefined,
      );
      runId = String(submit.run_id ?? '');
    }

    for await (const [eventType, payload] of this.http.streamSse(
      `/v1/context/${runId}/stream`,
    )) {
      const typed = parseContextEvent(eventType, payload);
      if (typed !== null) {
        yield typed;
        if (typed.type === 'terminal') return;
      }
    }
  }

  /**
   * Stream a Context run to terminal. Falls back to polling if the stream
   * dies before terminal.
   */
  private async _waitContextRun(
    runId: string,
    opts: { pollIntervalMs: number; timeoutMs: number },
  ): Promise<ContextResult> {
    const deadline = Date.now() + opts.timeoutMs;
    try {
      for await (const event of this.contextStream({ runId })) {
        if (event.type === 'terminal') break;
        if (Date.now() > deadline) {
          throw new TimeoutError(
            `Context run ${runId} did not terminate within ${opts.timeoutMs}ms`,
            408,
            {},
            {},
          );
        }
      }
    } catch (err) {
      if (err instanceof TimeoutError) throw err;
      // Stream died — fall through to polling for the final state.
    }

    while (true) {
      const result = await this.getContextRun(runId);
      if (result.isTerminal) return result;
      if (Date.now() > deadline) {
        throw new TimeoutError(
          `Context run ${runId} did not terminate within ${opts.timeoutMs}ms`,
          408,
          {},
          {},
        );
      }
      await new Promise((r) => setTimeout(r, opts.pollIntervalMs));
    }
  }

  /** Fetch the current state of a Context run. */
  async getContextRun(runId: string): Promise<ContextResult> {
    const data = await this.http.get(`/v1/context/${runId}`);
    return contextResultFromDict(data, this);
  }

  /** Fetch the ShapedOutput for a Context run. */
  async getContextOutput(runId: string): Promise<ContextOutput> {
    const data = await this.http.get(`/v1/context/${runId}/output`);
    return contextOutputFromDict(data);
  }

  /** Cancel an in-flight Context run (asynchronous on the server side). */
  async cancelContextRun(runId: string): Promise<void> {
    await this.http.delete(`/v1/context/${runId}`);
  }

  /**
   * Create a new version on the same chain. Re-runs the entire pipeline
   * against the same generator config. Returns the new version's
   * ContextResult.
   *
   * Set `wait: false` to return immediately after submit.
   */
  async refreshContext(
    runId: string,
    opts: { wait?: boolean; pollIntervalMs?: number; timeoutMs?: number } = {},
  ): Promise<ContextResult> {
    const wait = opts.wait !== false;
    const pollIntervalMs = opts.pollIntervalMs ?? 3000;
    const timeoutMs = opts.timeoutMs ?? 600_000;

    const data = await this.http.post(`/v1/context/${runId}/refresh`);
    const newRunId = String(data.run_id ?? runId);
    const result = await this.getContextRun(newRunId);
    if (!wait || result.isTerminal) return result;
    return this._waitContextRun(newRunId, { pollIntervalMs, timeoutMs });
  }

  /** List all versions on a Context run's chain (newest last). */
  async listContextVersions(runId: string): Promise<ContextVersion[]> {
    const data = await this.http.get(`/v1/context/${runId}/versions`);
    const versions = (data.versions as Record<string, unknown>[] | undefined) ?? [];
    return versions.map(contextVersionFromDict);
  }

  /**
   * Diff two Context versions. When both ids are on the same chain, diffs
   * the two most recent versions; when they're on different chains, diffs
   * cross-chain.
   */
  async diffContext(runId: string, otherRunId: string): Promise<ContextDiff> {
    const data = await this.http.get(`/v1/context/${runId}/diff/${otherRunId}`);
    return contextDiffFromDict(data);
  }

  /**
   * Move the chain's current pointer back to an earlier version. Pointer
   * move, not delete — the newer version stays on the chain and you can
   * roll forward again.
   */
  async rollbackContext(runId: string, version: number): Promise<ContextResult> {
    await this.http.post(
      `/v1/context/${runId}/rollback/${Math.trunc(version)}`,
    );
    return this.getContextRun(runId);
  }

  /**
   * Discover what Sources / Strategies / Synthesizers / Reconcilers are
   * available. Useful for building a generator-creation UI.
   */
  async contextCatalog(): Promise<ContextCatalog> {
    const fetchOne = async (path: string): Promise<CatalogEntry[]> => {
      try {
        const data = await this.http.get(path);
        const items =
          (data.items as Record<string, unknown>[] | undefined) ??
          (Array.isArray(data) ? (data as unknown as Record<string, unknown>[]) : []);
        return items.map(catalogEntryFromDict);
      } catch {
        return [];
      }
    };
    const [sources, strategies, synthesizers, reconcilers] = await Promise.all([
      fetchOne('/v1/context/sources'),
      fetchOne('/v1/context/strategies'),
      fetchOne('/v1/context/synthesizers'),
      fetchOne('/v1/context/reconcilers'),
    ]);
    return { sources, strategies, synthesizers, reconcilers, shapes: synthesizers };
  }

  // -------------------------------------------------------------------------
  // Schema Generation
  // -------------------------------------------------------------------------

  /**
   * Generate extraction schema from HTML using LLM.
   *
   * Supports three modes:
   * - Single HTML: Pass a single HTML string
   * - Multiple HTML: Pass an array of HTML strings for robust selectors
   * - From URLs: Pass an object with `urls` array (max 3) to fetch HTML from
   *
   * @param htmlOrUrls - HTML string, array of HTML strings, or object with urls array
   * @param options - Generation options
   * @returns GeneratedSchema with selectors or error
   *
   * @example
   * ```typescript
   * // Single HTML
   * const schema = await crawler.generateSchema(page.html, { query: 'Extract products' });
   *
   * // Multiple HTML samples
   * const schema = await crawler.generateSchema([page1.html, page2.html], {
   *   query: 'Extract products from these samples'
   * });
   *
   * // From URLs (max 3)
   * const schema = await crawler.generateSchema(
   *   { urls: ['https://example.com/p/1', 'https://example.com/p/2'] },
   *   { query: 'Extract product details' }
   * );
   * ```
   */
  async generateSchema(
    htmlOrUrls: string | string[] | { urls: string[] },
    options: {
      query?: string;
      schemaType?: 'CSS' | 'XPATH';
      targetJsonExample?: Record<string, unknown>;
      llmConfig?: Record<string, unknown>;
    } = {}
  ): Promise<GeneratedSchema> {
    const { query, schemaType = 'CSS', targetJsonExample, llmConfig } = options;

    const body: Record<string, unknown> = { schema_type: schemaType };

    if (typeof htmlOrUrls === 'string') {
      // Single HTML string
      body.html = htmlOrUrls;
    } else if (Array.isArray(htmlOrUrls)) {
      // Array of HTML strings
      body.html = htmlOrUrls;
    } else if (htmlOrUrls && 'urls' in htmlOrUrls) {
      // URLs object
      if (htmlOrUrls.urls.length > 3) {
        throw new Error('Maximum 3 URLs allowed');
      }
      body.urls = htmlOrUrls.urls;
    } else {
      throw new Error("Either 'html' or 'urls' must be provided");
    }

    if (query) body.query = query;
    if (targetJsonExample) body.target_json_example = targetJsonExample;
    if (llmConfig) body.llm_config = llmConfig;

    const data = await this.http.post('/v1/schema/generate', body, 60000);
    return generatedSchemaFromDict(data);
  }

  // -------------------------------------------------------------------------
  // Storage & Health
  // -------------------------------------------------------------------------

  /**
   * Get current storage usage.
   */
  async storage(): Promise<StorageUsage> {
    const data = await this.http.get('/v1/crawl/storage');
    return storageUsageFromDict(data);
  }

  /**
   * Check API health status.
   */
  async health(): Promise<Record<string, unknown>> {
    return this.http.get('/health');
  }

  // -------------------------------------------------------------------------
  // Utilities
  // -------------------------------------------------------------------------

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  // =========================================================================
  // Wrapper API -- Simplified endpoints
  // =========================================================================

  /**
   * Fetch a page and return clean markdown plus optional extras.
   * `POST /v1/scrape` (sync, single URL). Use {@link scrapeMany} for batch / async / webhooks.
   *
   * @example
   * const r = await crawler.scrape('https://example.com', { include: ['links', 'metadata'] });
   * console.log(r.markdown);
   */
  async scrape(url: string, options: MarkdownOptions = {}): Promise<MarkdownResponse> {
    const { strategy = 'browser', fit = true, include, crawlerConfig, browserConfig, proxy, bypassCache } = options;
    const body: Record<string, unknown> = { url, strategy, fit };
    if (include) body.include = include;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    const data = await this.http.post('/v1/scrape', body);
    return markdownResponseFromDict(data);
  }

  /**
   * @deprecated Use {@link scrape}. `/v1/markdown` was renamed to `/v1/scrape`.
   * Will be removed in 0.8.0.
   */
  async markdown(url: string, options: MarkdownOptions = {}): Promise<MarkdownResponse> {
    if (typeof process !== 'undefined' && process.emitWarning) {
      process.emitWarning('crawler.markdown() is deprecated — use crawler.scrape().', 'DeprecationWarning');
    }
    return this.scrape(url, options);
  }

  async screenshot(url: string, options: ScreenshotOptions = {}): Promise<ScreenshotResponse> {
    const { fullPage = true, pdf = false, waitFor, crawlerConfig, browserConfig, proxy, bypassCache } = options;
    const body: Record<string, unknown> = { url, full_page: fullPage };
    if (pdf) body.pdf = true;
    if (waitFor) body.wait_for = waitFor;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    const data = await this.http.post('/v1/screenshot', body, 120000);
    return screenshotResponseFromDict(data);
  }

  async extract(url: string, options: ExtractOptions = {}): Promise<ExtractResponse> {
    const { query, jsonExample, schema, method = 'auto', strategy = 'http', crawlerConfig, browserConfig, llmConfig, proxy, bypassCache } = options;
    const body: Record<string, unknown> = { url, method, strategy };
    if (query) body.query = query;
    if (jsonExample) body.json_example = jsonExample;
    if (schema) body.schema = schema;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (llmConfig) body.llm_config = llmConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    const data = await this.http.post('/v1/extract', body, 180000);
    return extractResponseFromDict(data);
  }

  /**
   * Discover all URLs on a domain via DomainMapper.
   *
   * @param options.sources `'primary'` (sitemap+homepage+robots+RSS, ~2-15s)
   *   or `'extended'` (adds Wayback+CC+CRT, ~30-60s). Default `'primary'`.
   *   Only flip to `'extended'` when primary returns too few URLs.
   * @param options.mode DEPRECATED — use `sources`. `'default'` → `'primary'`,
   *   `'deep'` → `'extended'`. Will be removed in 0.8.0.
   */
  async map(url: string, options: MapOptions = {}): Promise<MapResponse> {
    let { sources = 'primary' } = options;
    const { mode, maxUrls, includeSubdomains = false, extractHead = true, query, scoreThreshold, force = false, proxy } = options;
    if (mode !== undefined) {
      if (typeof process !== 'undefined' && process.emitWarning) {
        process.emitWarning('map(mode) is deprecated — use sources ("primary" | "extended").', 'DeprecationWarning');
      }
      sources = mode === 'deep' ? 'extended' : 'primary';
    }
    const body: Record<string, unknown> = { url, sources, include_subdomains: includeSubdomains, extract_head: extractHead };
    if (maxUrls !== undefined) body.max_urls = maxUrls;
    if (query) body.query = query;
    if (scoreThreshold !== undefined) body.score_threshold = scoreThreshold;
    if (force) body.force = true;
    if (proxy) body.proxy = proxy;
    const data = await this.http.post('/v1/map', body, 120000);
    return mapResponseFromDict(data);
  }

  // ---- Async batch methods ----

  /**
   * Submit an async scrape job over a list of URLs.
   * `POST /v1/scrape/async`. Returns a job; pass `wait: true` to poll until terminal.
   */
  async scrapeMany(urls: string[], options: MarkdownManyOptions = {}): Promise<WrapperJob> {
    const { strategy = 'browser', fit = true, include, crawlerConfig, browserConfig, proxy, bypassCache, wait = false, pollInterval = 2, timeout, webhookUrl, priority = 5 } = options;
    const body: Record<string, unknown> = { urls, strategy, fit, priority };
    if (include) body.include = include;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    if (webhookUrl) body.webhook_url = webhookUrl;
    const data = await this.http.post('/v1/scrape/async', body);
    let job = wrapperJobFromDict(data);
    if (wait) job = await this.waitWrapperJob(job.jobId, 'markdown', pollInterval, timeout);
    return job;
  }

  /**
   * @deprecated Use {@link scrapeMany}. Will be removed in 0.8.0.
   */
  async markdownMany(urls: string[], options: MarkdownManyOptions = {}): Promise<WrapperJob> {
    if (typeof process !== 'undefined' && process.emitWarning) {
      process.emitWarning('crawler.markdownMany() is deprecated — use crawler.scrapeMany().', 'DeprecationWarning');
    }
    return this.scrapeMany(urls, options);
  }

  async screenshotMany(urls: string[], options: ScreenshotManyOptions = {}): Promise<WrapperJob> {
    const { fullPage = true, pdf = false, waitFor, crawlerConfig, browserConfig, proxy, bypassCache, wait = false, pollInterval = 2, timeout, webhookUrl, priority = 5 } = options;
    const body: Record<string, unknown> = { urls, full_page: fullPage, priority };
    if (pdf) body.pdf = true;
    if (waitFor) body.wait_for = waitFor;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    if (webhookUrl) body.webhook_url = webhookUrl;
    const data = await this.http.post('/v1/screenshot/async', body);
    let job = wrapperJobFromDict(data);
    if (wait) job = await this.waitWrapperJob(job.jobId, 'screenshot', pollInterval, timeout);
    return job;
  }

  /**
   * Submit an async extract job over one base URL plus optional followers.
   * `POST /v1/extract/async`.
   *
   * The base `url` is the schema **template** in css_schema mode — the server
   * samples it, generates a schema once, then re-applies that schema across
   * every entry in `extraUrls` for free (no extra LLM calls per URL). In
   * `method: 'llm'` mode the base has no special role; every URL gets its
   * own LLM call.
   *
   * @param url Base URL (required). Up to 100 URLs total (1 base + 99 extras).
   * @param options.extraUrls Follower URLs that share the resolved strategy.
   * @param options.method `'auto'` (default), `'schema'`, or `'llm'`. AUTO works
   *   for batch as of API v2.2 — the previous "AUTO not allowed for batch"
   *   restriction was removed.
   */
  async extractMany(url: string, options: ExtractManyOptions): Promise<WrapperJob> {
    const { extraUrls, method = 'auto', query, jsonExample, schema, strategy = 'http', crawlerConfig, browserConfig, llmConfig, proxy, bypassCache, wait = false, pollInterval = 2, timeout, webhookUrl, priority = 5 } = options;
    const body: Record<string, unknown> = { url, method, strategy, priority };
    if (extraUrls && extraUrls.length > 0) body.extra_urls = extraUrls;
    if (query) body.query = query;
    if (jsonExample) body.json_example = jsonExample;
    if (schema) body.schema = schema;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (llmConfig) body.llm_config = llmConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    if (webhookUrl) body.webhook_url = webhookUrl;
    const data = await this.http.post('/v1/extract/async', body);
    let job = wrapperJobFromDict(data);
    if (wait) job = await this.waitWrapperJob(job.jobId, 'extract', pollInterval, timeout);
    return job;
  }

  // ---- Site crawl (always async) ----

  /**
   * Crawl an entire website — AI-assisted discovery + optional extraction.
   * Always async.
   *
   * The flagship flow: pass a plain-English `criteria` and let the LLM pick
   * the scan strategy (mode, patterns, query), generate URL filters, and
   * (optionally) build an extraction schema from a sample URL. Poll one
   * unified endpoint for both scan and crawl phases via getSiteCrawlJob().
   *
   * @example
   * ```typescript
   * const job = await crawler.crawlSite('https://books.toscrape.com', {
   *   criteria: 'all book listing pages',
   *   maxPages: 50,
   *   strategy: 'http',
   *   extract: {
   *     query: 'book title, price, rating',
   *     jsonExample: { title: '...', price: '£0.00', rating: 0 },
   *     method: 'auto',
   *   },
   *   wait: true,
   *   pollInterval: 3,
   * });
   * console.log('AI reasoning:', job.generatedConfig?.reasoning);
   * console.log('Extraction:', job.extractionMethodUsed);
   * ```
   */
  /**
   * /v1/crawl/site was removed (zero traffic for 14 days, deletion approved
   * 2026-05). Refuse client-side rather than 404 on the wire.
   *
   * For recursive site crawling, use `crawler.site(url, { mode: 'traverse' })`.
   * For AI-assisted discovery + extraction, pipeline `crawler.scan({ criteria })`
   * → `crawler.extractMany(first, { extraUrls: rest, ... })`.
   */
  async crawlSite(_url: string, _options: SiteCrawlOptions = {}): Promise<SiteCrawlResponse> {
    throw new Error(
      "crawler.crawlSite() — the /v1/crawl/site endpoint was removed. " +
      "For recursive site crawling, use crawler.site(url, { mode: 'traverse', maxUrls }). " +
      "For AI-assisted discovery + extraction, pipeline crawler.scan({ criteria }) → " +
      "crawler.extractMany(first, { extraUrls: rest })."
    );
  }

  /**
   * Poll a site crawl job started via `crawlSite()`.
   *
   * This is the unified polling endpoint — it merges the scan phase (URL
   * discovery) and the crawl phase (per-page fetch + extract) into one
   * response. `phase` walks through "scan" → "crawl" → "done".
   */
  async getSiteCrawlJob(jobId: string): Promise<SiteCrawlJobStatus> {
    const data = await this.http.get(`/v1/crawl/site/jobs/${jobId}`);
    return siteCrawlJobStatusFromDict(data);
  }

  private async waitSiteCrawlJob(
    jobId: string,
    pollInterval: number = 5.0,
    timeout?: number,
  ): Promise<SiteCrawlJobStatus> {
    const start = Date.now();
    while (true) {
      const job = await this.getSiteCrawlJob(jobId);
      if (isSiteCrawlJobComplete(job)) return job;
      if (timeout && Date.now() - start > timeout * 1000) {
        throw new TimeoutError(
          `Timeout waiting for site crawl ${jobId}. ` +
            `Phase: ${job.phase}, crawled: ${job.progress.urlsCrawled}/${job.progress.total}`,
        );
      }
      await this.sleep(pollInterval * 1000);
    }
  }

  // ---- Wrapper job management ----

  private async waitWrapperJob(jobId: string, jobType: string, pollInterval: number = 2, timeout?: number): Promise<WrapperJob> {
    const start = Date.now();
    while (true) {
      const job = await this.getWrapperJob(jobId, jobType);
      if (isWrapperJobComplete(job)) {
        // The cloud GET endpoint returns urlStatuses[] for fan-out parents
        // but never inlines per-URL data — that lives in S3 and is fetched
        // separately. wait=true callers expect job.results populated, so we
        // hydrate here. Failed URLs become CrawlResult stubs (success=false
        // + errorMessage) so results.length always equals urlStatuses.length.
        if (job.urlStatuses && job.urlStatuses.length > 0) {
          job.results = await this.hydrateResults(job);
        }
        return job;
      }
      if (timeout && (Date.now() - start) > timeout * 1000) {
        throw new TimeoutError(`Job ${jobId} did not complete within ${timeout}s`);
      }
      await this.sleep(pollInterval * 1000);
    }
  }

  private async hydrateResults(job: WrapperJob): Promise<CrawlResult[]> {
    if (!job.urlStatuses) return [];
    const fetches = job.urlStatuses.map(async (entry): Promise<CrawlResult> => {
      if (entry.status === 'failed') {
        return {
          url: entry.url,
          success: false,
          errorMessage: entry.error || 'URL failed',
          durationMs: entry.durationMs || 0,
        } as CrawlResult;
      }
      try {
        return await this.getPerUrlResult(job.jobId, entry.index);
      } catch (e) {
        return {
          url: entry.url,
          success: false,
          errorMessage: `per-URL fetch failed: ${(e as Error).message}`,
          durationMs: entry.durationMs || 0,
        } as CrawlResult;
      }
    });
    return Promise.all(fetches);
  }

  /**
   * Fetch one URL's full result from a multi-URL fan-out parent.
   *
   * Recipe-agnostic — works for any wrapper async parent (scrape /
   * screenshot / extract / crawl). Children all write to a unified S3
   * prefix keyed on (jobId, urlIndex), so the path is the same regardless
   * of which wrapper created the parent.
   *
   * @param jobId Parent job ID (from any *Many / *Async call).
   * @param urlIndex 0-based index into the parent's submitted URL list.
   *                 Match this against entries in `job.urlStatuses`.
   * @returns CrawlResult. `markdown` is populated for scrape jobs,
   *          `screenshot` (base64) for screenshot jobs, `extractedContent`
   *          for extract jobs.
   */
  async getPerUrlResult(jobId: string, urlIndex: number): Promise<CrawlResult> {
    const data = await this.http.get(`/v1/crawl/jobs/${jobId}/result/${urlIndex}`);
    return crawlResultFromDict(data as Record<string, unknown>);
  }

  private async getWrapperJob(jobId: string, jobType: string): Promise<WrapperJob> {
    const data = await this.http.get(`/v1/${jobType}/jobs/${jobId}`);
    return wrapperJobFromDict(data);
  }

  async getMarkdownJob(jobId: string): Promise<WrapperJob> { return this.getWrapperJob(jobId, 'markdown'); }
  async getScreenshotJob(jobId: string): Promise<WrapperJob> { return this.getWrapperJob(jobId, 'screenshot'); }
  async getExtractJob(jobId: string): Promise<WrapperJob> { return this.getWrapperJob(jobId, 'extract'); }

  async listMarkdownJobs(options: { status?: string; limit?: number; offset?: number } = {}): Promise<WrapperJob[]> {
    const params: Record<string, string | number | boolean> = { limit: options.limit || 20, offset: options.offset || 0 };
    if (options.status) params.status = options.status;
    const data = await this.http.get('/v1/markdown/jobs', params);
    return ((data as Record<string, unknown>).jobs as Record<string, unknown>[] || []).map(wrapperJobFromDict);
  }

  async listScreenshotJobs(options: { status?: string; limit?: number; offset?: number } = {}): Promise<WrapperJob[]> {
    const params: Record<string, string | number | boolean> = { limit: options.limit || 20, offset: options.offset || 0 };
    if (options.status) params.status = options.status;
    const data = await this.http.get('/v1/screenshot/jobs', params);
    return ((data as Record<string, unknown>).jobs as Record<string, unknown>[] || []).map(wrapperJobFromDict);
  }

  async listExtractJobs(options: { status?: string; limit?: number; offset?: number } = {}): Promise<WrapperJob[]> {
    const params: Record<string, string | number | boolean> = { limit: options.limit || 20, offset: options.offset || 0 };
    if (options.status) params.status = options.status;
    const data = await this.http.get('/v1/extract/jobs', params);
    return ((data as Record<string, unknown>).jobs as Record<string, unknown>[] || []).map(wrapperJobFromDict);
  }

  async cancelMarkdownJob(jobId: string): Promise<boolean> { await this.http.delete(`/v1/markdown/jobs/${jobId}`); return true; }
  async cancelScreenshotJob(jobId: string): Promise<boolean> { await this.http.delete(`/v1/screenshot/jobs/${jobId}`); return true; }
  async cancelExtractJob(jobId: string): Promise<boolean> { await this.http.delete(`/v1/extract/jobs/${jobId}`); return true; }

  // =========================================================================
  // Enrich v2 API (multi-phase)
  // =========================================================================

  /**
   * Create a multi-phase enrichment job.
   *
   * Phase machine:
   *   queued → planning → plan_ready → resolving_urls → urls_ready
   *         → extracting → merging → completed
   *
   * Defaults `autoConfirmPlan=true, autoConfirmUrls=true` make the job run
   * straight through. Set either to false for a human-in-loop review flow
   * and resume via `resumeEnrichJob(...)`.
   *
   * @example
   * // Agent one-shot
   * const result = await crawler.enrich({
   *   query: 'licensed nurseries in North York Toronto',
   *   country: 'ca',
   * });
   * for (const row of result.phaseData.rows ?? []) console.log(row.fields);
   *
   * @example
   * // Pre-resolved URLs
   * const result = await crawler.enrich({
   *   urls: ['https://example.com/a', 'https://example.com/b'],
   *   features: ['price', 'hours'],
   * });
   */
  async enrich(options: EnrichOptions = {}): Promise<EnrichJobStatus> {
    const {
      query, entities, criteria, features, urls, groups,
      autoConfirmPlan = true, autoConfirmUrls = true,
      topKPerEntity = 3, search = true, country, locationHint,
      strategy = 'http',
      config, browserConfig, crawlerConfig, llmConfig,
      proxy, webhookUrl, priority = 5,
      wait = true, pollInterval = 3.0, timeout = 600,
    } = options;

    const body: Record<string, unknown> = {
      auto_confirm_plan: autoConfirmPlan,
      auto_confirm_urls: autoConfirmUrls,
      top_k_per_entity: topKPerEntity,
      search,
      strategy,
      priority,
    };
    if (query !== undefined) body.query = query;
    if (entities !== undefined) body.entities = entities.map(normalizeEntity);
    if (criteria !== undefined) body.criteria = criteria.map(normalizeCriterion);
    if (features !== undefined) body.features = features.map(normalizeFeature);
    if (urls !== undefined) body.urls = urls;
    if (groups !== undefined) body.groups = groups;
    if (country !== undefined) body.country = country;
    if (locationHint !== undefined) body.location_hint = locationHint;
    if (config !== undefined) body.config = config;
    if (browserConfig !== undefined) body.browser_config = browserConfig;
    if (crawlerConfig !== undefined) body.crawler_config = crawlerConfig;
    if (llmConfig !== undefined) body.llm_config = llmConfig;
    if (proxy !== undefined) {
      body.proxy = typeof proxy === 'string' ? { mode: proxy } : proxy;
    }
    if (webhookUrl !== undefined) body.webhook_url = webhookUrl;

    const data = await this.http.post('/v1/enrich/async', body);
    const job = enrichJobStatusFromDict(data);

    if (wait) {
      return this.waitEnrichJob(job.jobId, { pollInterval, timeout });
    }
    return job;
  }

  /** Fetch the current status of an enrichment job — one poll, no wait. */
  async getEnrichJob(jobId: string): Promise<EnrichJobStatus> {
    const data = await this.http.get(`/v1/enrich/jobs/${jobId}`);
    return enrichJobStatusFromDict(data);
  }

  /**
   * Poll an enrichment job until it reaches `until` or a terminal status.
   *
   * If `until` is set and the job pauses at a paused phase (`plan_ready` /
   * `urls_ready`) without auto-confirm, returns the paused state immediately
   * rather than spinning until timeout.
   */
  async waitEnrichJob(
    jobId: string,
    options: WaitEnrichOptions = {},
  ): Promise<EnrichJobStatus> {
    const { until, pollInterval = 3.0, timeout = 600 } = options;
    const start = Date.now();
    while (true) {
      const job = await this.getEnrichJob(jobId);
      if (isEnrichJobComplete(job)) return job;
      if (until !== undefined && job.status === until) return job;
      if (
        until !== undefined && ENRICH_PAUSED_STATUSES.includes(job.status) &&
        ((job.status === 'plan_ready' && !job.autoConfirmPlan) ||
         (job.status === 'urls_ready' && !job.autoConfirmUrls))
      ) {
        return job;
      }
      if (timeout && Date.now() - start > timeout * 1000) {
        throw new TimeoutError(
          `Enrich job ${jobId} did not reach '${until ?? 'completed'}' within ${timeout}s. ` +
            `Status: ${job.status}, progress: ${job.progress.completedUrls}/${job.progress.totalUrls}`,
        );
      }
      await this.sleep(pollInterval * 1000);
    }
  }

  /**
   * Advance a paused job (`plan_ready` or `urls_ready`) to the next phase.
   *
   * Pass any subset of edits to apply before resuming. An empty body is valid
   * — means "resume with the server's current values".
   */
  async resumeEnrichJob(
    jobId: string,
    options: ResumeEnrichOptions = {},
  ): Promise<EnrichJobStatus> {
    const { entities, criteria, features, groups } = options;
    const body: Record<string, unknown> = {};
    if (entities !== undefined) body.entities = entities.map(normalizeEntity);
    if (criteria !== undefined) body.criteria = criteria.map(normalizeCriterion);
    if (features !== undefined) body.features = features.map(normalizeFeature);
    if (groups !== undefined) body.groups = groups;
    const data = await this.http.post(`/v1/enrich/jobs/${jobId}/continue`, body);
    return enrichJobStatusFromDict(data);
  }

  /**
   * Subscribe to the SSE stream for an enrichment job.
   *
   * Returns an async iterable of `EnrichEvent` objects. Iteration ends when
   * the server sends a `complete` event or the connection drops.
   */
  async *streamEnrichJob(jobId: string): AsyncGenerator<EnrichEvent> {
    for await (const [evtType, payload] of this.http.streamSse(`/v1/enrich/jobs/${jobId}/stream`)) {
      yield enrichEventFromDict(evtType, payload);
      if (evtType === 'complete') return;
    }
  }

  /** Cancel a running enrichment job. */
  async cancelEnrichJob(jobId: string): Promise<boolean> {
    await this.http.delete(`/v1/enrich/jobs/${jobId}`);
    return true;
  }

  /** List enrichment jobs for the authenticated user. */
  async listEnrichJobs(options: {
    limit?: number;
    offset?: number;
  } = {}): Promise<EnrichJobListItem[]> {
    const { limit = 20, offset = 0 } = options;
    const params: Record<string, string | number> = { limit, offset };
    const data = await this.http.get('/v1/enrich/jobs', params);
    return ((data.jobs || []) as Record<string, unknown>[]).map(enrichJobListItemFromDict);
  }

  // ───────────────────────────────────────────────────────────────────
  // Discovery — wrapper-services platform: /v1/discovery/<service>
  // ───────────────────────────────────────────────────────────────────
  //
  // One method, dispatches to any registered vertical. `search` is live;
  // `people` / `products` / `posts` / `videos` will land via the same
  // call shape — your code never updates when a new vertical ships.

  /**
   * Run a Discovery vertical and return the typed response.
   *
   * `POST /v1/discovery/<service>` — the wrapper-services dispatcher.
   * New verticals don't add SDK methods; they become a new value for
   * the `service` argument.
   *
   * @param service - Vertical name (`"search"` today; `"people"` /
   *   `"products"` / `"posts"` / `"videos"` to follow).
   * @param params - Per-vertical request fields. For `service="search"`:
   *   `query` (required), `country`, `language`, `location`, `num`,
   *   `start`, `site`, `mode`, `time_period`, `backends` (string[],
   *   1-4 names from `google` / `bing` / `duckduckgo` / `brave`; omit
   *   for the server's default of `["google"]`; >1 fans out + merges
   *   via RRF + URL dedup), `use_cache` (boolean, default false —
   *   opt into the SERP cache; legacy `bypass_cache` still wins when
   *   true), `enhance_query` (boolean, default false — LLM rewrites the
   *   query into a backend-specific operator-rich query before the SERP
   *   fetch. Response carries `originalQuery` + `rewrittenQueries`.
   *   Never fails the search — provider errors fall back to original.),
   *   plus the synth knobs (`synthesize`, `synth_mode`,
   *   `synth_adaptive`, `synth_prompt`).
   *
   * @returns `SearchResponse` for `service="search"`. Generic object for
   *   verticals whose typed response shapes don't exist yet.
   *
   * Async lifecycle (synth requests):
   *   `queued → running → serp_ready → completed | failed`. The SDK's
   *   `wait: true` mode polls through `serp_ready` transparently.
   *   `wait: false` callers polling via `getDiscoveryJob` can read
   *   the SERP-only response at `serp_ready` for progressive
   *   rendering before synth lands.
   *
   * @example
   * // Default — single Google fetch, fresh.
   * const response = await crawler.discovery("search", {
   *   query: "best AI code review tools 2026",
   *   country: "us",
   * });
   * for (const hit of response.hits) {
   *   console.log(hit.rank, hit.title, hit.url);
   * }
   *
   * @example
   * // Multi-backend fan-out + merge, with synth.
   * const response = await crawler.discovery("search", {
   *   query: "what is warrior's next game?",
   *   country: "us",
   *   backends: ["google", "bing", "brave"],
   *   synthesize: true,
   *   synth_mode: "auto",
   * });
   * console.log(response.synthesizedAnswer?.text);
   *
   * @example
   * // LLM query rewrite per backend.
   * const response = await crawler.discovery("search", {
   *   query: "best nurseries in Toronto for my 2 year old",
   *   country: "ca",
   *   enhance_query: true,
   *   backends: ["google", "bing"],
   * });
   * console.log(response.originalQuery);     // what you typed
   * console.log(response.rewrittenQueries);  // { google: "...", bing: "..." }
   */
  async discovery(
    service: string,
    params: Record<string, unknown> = {},
    options: { wait?: boolean; pollIntervalMs?: number; pollTimeoutMs?: number } = {},
  ): Promise<SearchResponse | DiscoveryJobHandle | Record<string, unknown>> {
    const wait = options.wait !== false;            // default true
    const pollIntervalMs = options.pollIntervalMs ?? 1500;
    const pollTimeoutMs = options.pollTimeoutMs ?? 5 * 60 * 1000;

    // Drop null / empty-string optionals so the cache key matches the
    // dashboard playground exactly. Wire parity avoids surprise misses
    // between surfaces hitting the same params.
    const body: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(params)) {
      if (v !== null && v !== undefined && v !== '') body[k] = v;
    }

    // Synth-aware routing: the sync endpoint 422s when synthesize=true,
    // so the SDK targets /async automatically. wait=true polls; wait=false
    // returns a handle for the caller to drive.
    const wantsSynth = Boolean(body.synthesize);
    if (wantsSynth) {
      const handleData = await this.http.post(
        `/v1/discovery/${service}/async`, body,
      ) as Record<string, unknown>;
      const handle = discoveryJobHandleFromDict(handleData);
      if (!wait) return handle;

      const status = await this._waitForDiscoveryJob(
        handle.jobId, { pollIntervalMs, pollTimeoutMs },
      );
      if (status.status === 'failed') {
        throw new Error(
          `discovery(${service}) async job failed: ${status.error ?? 'unknown'}`,
        );
      }
      if (!status.result) {
        throw new Error(
          `discovery(${service}) async job completed but result missing`,
        );
      }
      if (service === 'search') {
        return searchResponseFromDict(status.result);
      }
      return status.result;
    }

    // Sync path — body returned inline.
    const data = await this.http.post(
      `/v1/discovery/${service}`, body,
    ) as Record<string, unknown>;
    if (service === 'search') {
      return searchResponseFromDict(data);
    }
    return data;
  }

  /**
   * Fetch the Discovery service registry.
   *
   * `GET /v1/discovery` — returns every vertical the cloud currently
   * ships, plus its request/response JSON schemas. Use this to
   * feature-detect new verticals without an SDK update.
   */
  async listDiscoveryServices(): Promise<DiscoveryService[]> {
    const data = await this.http.get('/v1/discovery') as Record<string, unknown>;
    const services = (data.services as Record<string, unknown>[]) || [];
    return services.map(discoveryServiceFromDict);
  }

  /**
   * Poll a Discovery async job by id.
   *
   * `GET /v1/discovery/jobs/{jobId}` — used by callers that opted out of
   * `wait: true` on `discovery()` and drive polling themselves.
   */
  async getDiscoveryJob(jobId: string): Promise<DiscoveryJobStatus> {
    const data = await this.http.get(
      `/v1/discovery/jobs/${jobId}`,
    ) as Record<string, unknown>;
    return discoveryJobStatusFromDict(data);
  }

  /** Internal: poll a job until terminal status or timeout. Mild backoff
   *  (×1.15 per poll, capped at 4s) so a slow job doesn't hammer the proxy. */
  private async _waitForDiscoveryJob(
    jobId: string,
    opts: { pollIntervalMs: number; pollTimeoutMs: number },
  ): Promise<DiscoveryJobStatus> {
    const deadline = Date.now() + opts.pollTimeoutMs;
    let interval = opts.pollIntervalMs;
    // eslint-disable-next-line no-constant-condition
    while (true) {
      if (Date.now() > deadline) {
        throw new Error(`discovery job ${jobId} did not complete within ${opts.pollTimeoutMs}ms`);
      }
      await new Promise((r) => setTimeout(r, interval));
      const status = await this.getDiscoveryJob(jobId);
      if (status.status === 'completed' || status.status === 'failed') {
        return status;
      }
      interval = Math.min(4000, Math.round(interval * 1.15));
    }
  }

  /**
   * Close the client (no-op for now, but provided for API compatibility).
   */
  async close(): Promise<void> {
    // HTTP client doesn't need explicit closing in Node.js
  }
}
