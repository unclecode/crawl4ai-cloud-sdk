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
  ContextResult,
  GeneratedSchema,
  StorageUsage,
  ProxyConfig,
  JobProgress,
  crawlResultFromDict,
  crawlJobFromDict,
  deepCrawlResultFromDict,
  scanResultFromDict,
  contextResultFromDict,
  generatedSchemaFromDict,
  storageUsageFromDict,
  isJobComplete,
  isDeepCrawlComplete,
  MarkdownResponse,
  ScreenshotResponse,
  ExtractResponse,
  MapResponse,
  SiteCrawlResponse,
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
  wrapperJobFromDict,
  isWrapperJobComplete,
} from './models';
import {
  CrawlerRunConfig,
  BrowserConfig,
  buildCrawlRequest,
  sanitizeCrawlerConfig,
  sanitizeBrowserConfig,
  normalizeProxy,
} from './configs';

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
  // Deep Crawl
  // -------------------------------------------------------------------------

  /**
   * Deep crawl - discover and crawl URLs from a starting point.
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
   * Discover all URLs under a domain without crawling.
   * Synchronous — results return inline, no polling needed.
   *
   * @example
   * const result = await crawler.scan('https://example.com', { maxUrls: 50 });
   * console.log(`Found ${result.totalUrls} URLs across ${result.hostsFound} hosts`);
   */
  async scan(
    url: string,
    options: ScanOptions = {},
  ): Promise<ScanResult> {
    const {
      mode = 'default',
      maxUrls,
      includeSubdomains = true,
      extractHead = true,
      soft404Detection = true,
      query,
      scoreThreshold,
      force = false,
      probeThreshold,
    } = options;

    const body: Record<string, unknown> = {
      url,
      mode,
      include_subdomains: includeSubdomains,
      extract_head: extractHead,
      soft_404_detection: soft404Detection,
      force,
    };
    if (maxUrls !== undefined) body.max_urls = maxUrls;
    if (query) body.query = query;
    if (scoreThreshold !== undefined) body.score_threshold = scoreThreshold;
    if (probeThreshold !== undefined) body.probe_threshold = probeThreshold;

    const data = await this.http.post('/v1/scan', body);
    return scanResultFromDict(data);
  }

  // -------------------------------------------------------------------------
  // Context API
  // -------------------------------------------------------------------------

  /**
   * Build context from a search query.
   */
  async context(
    query: string,
    options: {
      paaLimit?: number;
      resultsPerPaa?: number;
      wait?: boolean;
    } = {}
  ): Promise<ContextResult> {
    const { paaLimit = 3, resultsPerPaa = 5 } = options;

    const body: Record<string, unknown> = {
      query,
      strategy: 'serper_paa',
      paa_limit: paaLimit,
      results_per_paa: resultsPerPaa,
    };

    const data = await this.http.post('/v1/context', body, 300000);
    return contextResultFromDict(data);
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

  async markdown(url: string, options: MarkdownOptions = {}): Promise<MarkdownResponse> {
    const { strategy = 'browser', fit = true, include, crawlerConfig, browserConfig, proxy, bypassCache } = options;
    const body: Record<string, unknown> = { url, strategy, fit };
    if (include) body.include = include;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    const data = await this.http.post('/v1/markdown', body);
    return markdownResponseFromDict(data);
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

  async map(url: string, options: MapOptions = {}): Promise<MapResponse> {
    const { mode = 'default', maxUrls, includeSubdomains = false, extractHead = true, query, scoreThreshold, force = false, proxy } = options;
    const body: Record<string, unknown> = { url, mode, include_subdomains: includeSubdomains, extract_head: extractHead };
    if (maxUrls !== undefined) body.max_urls = maxUrls;
    if (query) body.query = query;
    if (scoreThreshold !== undefined) body.score_threshold = scoreThreshold;
    if (force) body.force = true;
    if (proxy) body.proxy = proxy;
    const data = await this.http.post('/v1/map', body, 120000);
    return mapResponseFromDict(data);
  }

  // ---- Async batch methods ----

  async markdownMany(urls: string[], options: MarkdownManyOptions = {}): Promise<WrapperJob> {
    const { strategy = 'browser', fit = true, include, crawlerConfig, browserConfig, proxy, bypassCache, wait = false, pollInterval = 2, timeout, webhookUrl, priority = 5 } = options;
    const body: Record<string, unknown> = { urls, strategy, fit, priority };
    if (include) body.include = include;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (bypassCache) body.bypass_cache = true;
    if (webhookUrl) body.webhook_url = webhookUrl;
    const data = await this.http.post('/v1/markdown/async', body);
    let job = wrapperJobFromDict(data);
    if (wait) job = await this.waitWrapperJob(job.jobId, 'markdown', pollInterval, timeout);
    return job;
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

  async extractMany(urls: string[], options: ExtractManyOptions): Promise<WrapperJob> {
    if (options.method === 'auto' as string) {
      throw new Error("AUTO method is not supported for batch extraction. Specify 'llm' or 'schema'.");
    }
    const { method, query, jsonExample, schema, strategy = 'http', crawlerConfig, browserConfig, llmConfig, proxy, bypassCache, wait = false, pollInterval = 2, timeout, webhookUrl, priority = 5 } = options;
    const body: Record<string, unknown> = { urls, method, strategy, priority };
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

  async crawlSite(url: string, options: SiteCrawlOptions = {}): Promise<SiteCrawlResponse> {
    const { maxPages = 20, discovery = 'map', strategy = 'browser', fit = true, include, pattern, maxDepth, crawlerConfig, browserConfig, proxy, webhookUrl, priority = 5 } = options;
    const body: Record<string, unknown> = { url, max_pages: maxPages, discovery, strategy, fit, priority };
    if (include) body.include = include;
    if (pattern) body.pattern = pattern;
    if (maxDepth !== undefined) body.max_depth = maxDepth;
    if (crawlerConfig) body.crawler_config = crawlerConfig;
    if (browserConfig) body.browser_config = browserConfig;
    if (proxy) body.proxy = proxy;
    if (webhookUrl) body.webhook_url = webhookUrl;
    const data = await this.http.post('/v1/crawl/site', body, 120000);
    return siteCrawlResponseFromDict(data);
  }

  // ---- Wrapper job management ----

  private async waitWrapperJob(jobId: string, jobType: string, pollInterval: number = 2, timeout?: number): Promise<WrapperJob> {
    const start = Date.now();
    while (true) {
      const job = await this.getWrapperJob(jobId, jobType);
      if (isWrapperJobComplete(job)) return job;
      if (timeout && (Date.now() - start) > timeout * 1000) {
        throw new TimeoutError(`Job ${jobId} did not complete within ${timeout}s`);
      }
      await this.sleep(pollInterval * 1000);
    }
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

  /**
   * Close the client (no-op for now, but provided for API compatibility).
   */
  async close(): Promise<void> {
    // HTTP client doesn't need explicit closing in Node.js
  }
}
