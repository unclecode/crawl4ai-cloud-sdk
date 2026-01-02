/**
 * AsyncWebCrawler - Main crawler class for Crawl4AI Cloud SDK.
 */

import { HTTPClient } from './client';
import { TimeoutError } from './errors';
import {
  CrawlResult,
  CrawlJob,
  DeepCrawlResult,
  ContextResult,
  GeneratedSchema,
  StorageUsage,
  ProxyConfig,
  JobProgress,
  crawlResultFromDict,
  crawlJobFromDict,
  deepCrawlResultFromDict,
  contextResultFromDict,
  generatedSchemaFromDict,
  storageUsageFromDict,
  isJobComplete,
  isDeepCrawlComplete,
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
   * For ≤10 URLs, uses sync batch endpoint.
   * For >10 URLs, creates an async job.
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

    // Use batch endpoint for small batches
    if (urls.length <= 10) {
      return this.runBatch(urls, runOptions, wait);
    }

    // Use async endpoint for larger batches
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

  private async runBatch(
    urls: string[],
    options: RunOptions,
    wait: boolean
  ): Promise<CrawlJob | CrawlResult[]> {
    const {
      config,
      browserConfig,
      strategy = 'browser',
      proxy,
      bypassCache = false,
      ...rest
    } = options;

    const body = buildCrawlRequest({
      urls,
      config,
      browserConfig,
      strategy,
      proxy,
      bypassCache,
      ...rest,
    });

    const data = await this.http.post('/v1/crawl/batch', body, 600000);
    const results = ((data.results || []) as Record<string, unknown>[]).map(
      crawlResultFromDict
    );

    if (wait) {
      return results;
    }

    // Wrap in a "completed" job for consistent return type
    const jobId = `batch_${Date.now()}`;
    const job: CrawlJob = {
      jobId,
      id: jobId,
      status: 'completed',
      progress: { total: urls.length, completed: urls.length, failed: 0 },
      urlsCount: urls.length,
      createdAt: '',
      results,  // Already CrawlResult[]
    };
    return job;
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
        includeResults: true,
      });
      // job.results is already CrawlResult[] from crawlJobFromDict
      return job.results || [];
    }

    return job;
  }

  // -------------------------------------------------------------------------
  // Job Management
  // -------------------------------------------------------------------------

  /**
   * Get job status and optionally results.
   */
  async getJob(jobId: string, includeResults = false): Promise<CrawlJob> {
    const params: Record<string, string> = {};
    if (includeResults) {
      params.include_results = 'true';
    }

    const data = await this.http.get(`/v1/crawl/jobs/${jobId}`, params);
    return crawlJobFromDict(data);
  }

  /**
   * Poll until job completes.
   */
  async waitJob(
    jobId: string,
    options: {
      pollInterval?: number;
      timeout?: number;
      includeResults?: boolean;
    } = {}
  ): Promise<CrawlJob> {
    const { pollInterval = 2.0, timeout, includeResults = true } = options;
    const startTime = Date.now();

    while (true) {
      const job = await this.getJob(jobId, false);

      if (isJobComplete(job)) {
        if (includeResults) {
          return this.getJob(jobId, true);
        }
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
    } = options;

    if (!url && !sourceJob) {
      throw new Error("Must provide either 'url' or 'sourceJob'");
    }
    if (url && sourceJob) {
      throw new Error("Provide either 'url' or 'sourceJob', not both");
    }

    // Build request body
    const body: Record<string, unknown> = {
      strategy,
      crawl_strategy: crawlStrategy,
      priority,
    };

    if (url) body.url = url;
    if (sourceJob) body.source_job_id = sourceJob;

    // Tree strategy options
    if (['bfs', 'dfs', 'best_first'].includes(strategy)) {
      body.max_depth = maxDepth;
      body.max_urls = maxUrls;
      if (filters) body.filters = filters;
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
        includeResults: true,
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
   */
  async generateSchema(
    html: string,
    options: {
      query?: string;
      schemaType?: 'CSS' | 'XPATH';
      targetJsonExample?: Record<string, unknown>;
      llmConfig?: Record<string, unknown>;
    } = {}
  ): Promise<GeneratedSchema> {
    const { query, schemaType = 'CSS', targetJsonExample, llmConfig } = options;

    const body: Record<string, unknown> = { html, schema_type: schemaType };
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

  /**
   * Close the client (no-op for now, but provided for API compatibility).
   */
  async close(): Promise<void> {
    // HTTP client doesn't need explicit closing in Node.js
  }
}
