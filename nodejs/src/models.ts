/**
 * Response models for Crawl4AI Cloud SDK.
 */

/**
 * Proxy configuration for cloud crawl requests.
 */
export interface ProxyConfig {
  mode: 'none' | 'datacenter' | 'residential' | 'auto';
  country?: string;
  sticky_session?: boolean;
  use_proxy?: boolean;
  skip_direct?: boolean;
}

/**
 * Async job progress.
 */
export interface JobProgress {
  total: number;
  completed: number;
  failed: number;
}

/**
 * Helper to get pending count from progress.
 */
export function getProgressPending(progress: JobProgress): number {
  return progress.total - progress.completed - progress.failed;
}

/**
 * Helper to get completion percentage.
 */
export function getProgressPercent(progress: JobProgress): number {
  if (progress.total === 0) return 0;
  return ((progress.completed + progress.failed) / progress.total) * 100;
}

/**
 * Async crawl job returned by runMany().
 */
export interface CrawlJob {
  jobId: string;
  /** @deprecated Use jobId instead */
  id: string;
  status: string;
  progress: JobProgress;
  urlsCount: number;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
  results?: CrawlResult[];
  error?: string;
  resultSizeBytes?: number;
  /** Resource usage metrics (completed jobs only) */
  usage?: Usage;
}

/**
 * Check if job is in a terminal state.
 */
export function isJobComplete(job: CrawlJob): boolean {
  return ['completed', 'partial', 'failed', 'cancelled'].includes(job.status);
}

/**
 * Check if job completed successfully.
 */
export function isJobSuccessful(job: CrawlJob): boolean {
  return job.status === 'completed';
}

/**
 * Create CrawlJob from API response.
 * Results are automatically converted to CrawlResult objects.
 */
export function crawlJobFromDict(data: Record<string, unknown>, convertResults: boolean = true): CrawlJob {
  const progressData = (data.progress || {}) as Record<string, unknown>;
  const jobId = (data.job_id || '') as string;

  // Convert results to CrawlResult objects if present
  let results: CrawlResult[] | undefined;
  const rawResults = data.results as Record<string, unknown>[] | undefined;
  if (rawResults && convertResults) {
    results = rawResults.map((r) => {
      const result = crawlResultFromDict(r);
      // Set job_id on each result for use with downloadUrl()
      result.id = jobId;
      return result;
    });
  }

  return {
    jobId,
    id: jobId, // backward compatibility alias
    status: (data.status || 'unknown') as string,
    progress: {
      total: (progressData.total || 0) as number,
      completed: (progressData.completed || 0) as number,
      failed: (progressData.failed || 0) as number,
    },
    urlsCount: (data.urls_count || data.url_count || 0) as number,
    createdAt: (data.created_at || '') as string,
    startedAt: data.started_at as string | undefined,
    completedAt: data.completed_at as string | undefined,
    results,
    error: data.error as string | undefined,
    resultSizeBytes: data.result_size_bytes as number | undefined,
    usage: data.usage ? usageFromDict(data.usage as Record<string, unknown>) : undefined,
  };
}

/**
 * Information about a scanned URL.
 */
export interface ScanUrlInfo {
  url: string;
  depth: number;
  score?: number;
  linksFound: number;
  htmlSize: number;
}

/**
 * URL discovered by domain scan (/v1/scan).
 */
export interface DomainScanUrlInfo {
  url: string;
  host: string;
  status: string;
  relevanceScore?: number;
  headData?: Record<string, unknown>;
}

/**
 * Response from domain scan (/v1/scan).
 *
 * For map mode (sync): `urls` is populated inline.
 * For deep mode (async): `jobId` + `status` are set; poll with
 * `getScanJob(jobId)` and the `urls` list will be populated progressively.
 * When `criteria` was supplied in the request, `generatedConfig` carries
 * the LLM output and `modeUsed` tells you which strategy ran.
 */
export interface ScanResult {
  success: boolean;
  domain: string;
  totalUrls: number;
  hostsFound: number;
  mode: string;
  urls: DomainScanUrlInfo[];
  durationMs: number;
  error?: string;
  // AI-assisted / async fields
  modeUsed?: 'map' | 'deep';
  jobId?: string;
  status?: string;
  generatedConfig?: GeneratedConfig;
  message?: string;
}

/**
 * True when the scan response is for an async (deep) scan — poll with getScanJob().
 */
export function isScanResultAsync(result: ScanResult): boolean {
  return result.modeUsed === 'deep' && Boolean(result.jobId);
}

/**
 * Unified scan configuration for AI-assisted URL discovery.
 *
 * Used by scan() and crawlSite(). When `criteria` is set on the parent
 * request, the AI config generator fills unset fields here. Explicit fields
 * always win over LLM output.
 */
export interface SiteScanConfig {
  mode?: 'auto' | 'map' | 'deep';
  patterns?: string[];
  filters?: Record<string, unknown>;
  scorers?: Record<string, unknown>;
  query?: string;
  scoreThreshold?: number;
  includeSubdomains?: boolean;
  maxDepth?: number;
}

/**
 * Convert SiteScanConfig (camelCase) to the snake_case payload the API expects.
 * Only non-undefined fields are included.
 */
export function siteScanConfigToDict(cfg: SiteScanConfig): Record<string, unknown> {
  const d: Record<string, unknown> = {};
  if (cfg.mode !== undefined) d.mode = cfg.mode;
  if (cfg.patterns !== undefined) d.patterns = cfg.patterns;
  if (cfg.filters !== undefined) d.filters = cfg.filters;
  if (cfg.scorers !== undefined) d.scorers = cfg.scorers;
  if (cfg.query !== undefined) d.query = cfg.query;
  if (cfg.scoreThreshold !== undefined) d.score_threshold = cfg.scoreThreshold;
  if (cfg.includeSubdomains !== undefined) d.include_subdomains = cfg.includeSubdomains;
  if (cfg.maxDepth !== undefined) d.max_depth = cfg.maxDepth;
  return d;
}

/**
 * Structured extraction configuration for crawlSite().
 *
 * When set without a pre-built `schema`, the wrapper fetches `sampleUrl`
 * (defaults to the crawl's start URL), generates a schema via LLM, and
 * applies it to every discovered URL.
 */
export interface SiteExtractConfig {
  query?: string;
  jsonExample?: Record<string, unknown>;
  method?: 'auto' | 'llm' | 'schema';
  schema?: Record<string, unknown>;
  sampleUrl?: string;
  urlPattern?: string;
}

/**
 * Convert SiteExtractConfig (camelCase) to the snake_case payload the API expects.
 */
export function siteExtractConfigToDict(cfg: SiteExtractConfig): Record<string, unknown> {
  const d: Record<string, unknown> = {};
  if (cfg.method !== undefined) d.method = cfg.method;
  if (cfg.query !== undefined) d.query = cfg.query;
  if (cfg.jsonExample !== undefined) d.json_example = cfg.jsonExample;
  if (cfg.schema !== undefined) d.schema = cfg.schema;
  if (cfg.sampleUrl !== undefined) d.sample_url = cfg.sampleUrl;
  if (cfg.urlPattern !== undefined) d.url_pattern = cfg.urlPattern;
  return d;
}

/**
 * LLM-generated config echoed back by /v1/scan and /v1/crawl/site when
 * `criteria` was set in the request. Contains the scan config and (for
 * /v1/crawl/site) the extract config, plus LLM reasoning and cache/fallback
 * flags.
 */
export interface GeneratedConfig {
  scan: Record<string, unknown>;
  reasoning: string;
  extract?: Record<string, unknown>;
  fallback: boolean;
  cached: boolean;
}

/**
 * Create GeneratedConfig from API response dict.
 */
export function generatedConfigFromDict(data: Record<string, unknown>): GeneratedConfig {
  return {
    scan: (data.scan as Record<string, unknown>) || {},
    reasoning: (data.reasoning as string) || '',
    extract: data.extract as Record<string, unknown> | undefined,
    fallback: Boolean(data.fallback),
    cached: Boolean(data.cached),
  };
}

/**
 * Polling response for GET /v1/scan/jobs/{job_id} — used with async deep
 * scans. URLs are appended to `urls` as they're discovered.
 */
export interface ScanJobStatus {
  jobId: string;
  status: string;
  modeUsed: string; // "map" | "deep"
  domain?: string;
  totalUrls: number;
  urls: DomainScanUrlInfo[];
  progress?: { completed?: number; total?: number };
  generatedConfig?: GeneratedConfig;
  durationMs: number;
  error?: string;
  createdAt?: string;
  completedAt?: string;
}

/**
 * Create ScanJobStatus from API response.
 */
export function scanJobStatusFromDict(data: Record<string, unknown>): ScanJobStatus {
  const urls: DomainScanUrlInfo[] = [];
  if (data.urls && Array.isArray(data.urls)) {
    for (const u of data.urls as Record<string, unknown>[]) {
      urls.push({
        url: (u.url || '') as string,
        host: (u.host || '') as string,
        status: (u.status || 'valid') as string,
        relevanceScore: u.relevance_score as number | undefined,
        headData: u.head_data as Record<string, unknown> | undefined,
      });
    }
  }
  return {
    jobId: (data.job_id || '') as string,
    status: (data.status || 'pending') as string,
    modeUsed: (data.mode_used || 'deep') as string,
    domain: data.domain as string | undefined,
    totalUrls: (data.total_urls || 0) as number,
    urls,
    progress: data.progress as { completed?: number; total?: number } | undefined,
    generatedConfig: data.generated_config
      ? generatedConfigFromDict(data.generated_config as Record<string, unknown>)
      : undefined,
    durationMs: (data.duration_ms || 0) as number,
    error: data.error as string | undefined,
    createdAt: data.created_at as string | undefined,
    completedAt: data.completed_at as string | undefined,
  };
}

/**
 * True when a scan job has reached a terminal state.
 */
export function isScanJobComplete(job: ScanJobStatus): boolean {
  return ['completed', 'partial', 'failed', 'cancelled'].includes(job.status);
}

/**
 * Options for the scan() method.
 */
export interface ScanOptions {
  mode?: 'default' | 'deep';
  maxUrls?: number;
  includeSubdomains?: boolean;
  extractHead?: boolean;
  soft404Detection?: boolean;
  query?: string;
  scoreThreshold?: number;
  force?: boolean;
  probeThreshold?: number;
  // AI-assisted fields
  criteria?: string;
  scan?: SiteScanConfig | Record<string, unknown>;
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
}

/**
 * Deep crawl response.
 */
export interface DeepCrawlResult {
  jobId: string;
  status: string;
  strategy: string;
  discoveredCount: number;
  queuedUrls: number;
  createdAt: string;
  urls?: ScanUrlInfo[];
  htmlDownloadUrl?: string;
  cacheExpiresAt?: string;
  crawlJobId?: string;
}

/**
 * Create ScanResult from API response.
 */
export function scanResultFromDict(data: Record<string, unknown>): ScanResult {
  const urls: DomainScanUrlInfo[] = [];
  if (data.urls && Array.isArray(data.urls)) {
    for (const u of data.urls as Record<string, unknown>[]) {
      urls.push({
        url: (u.url || '') as string,
        host: (u.host || '') as string,
        status: (u.status || 'valid') as string,
        relevanceScore: u.relevance_score as number | undefined,
        headData: u.head_data as Record<string, unknown> | undefined,
      });
    }
  }

  return {
    success: (data.success || false) as boolean,
    domain: (data.domain || '') as string,
    totalUrls: (data.total_urls || 0) as number,
    hostsFound: (data.hosts_found || 0) as number,
    mode: (data.mode || 'default') as string,
    urls,
    durationMs: (data.duration_ms || 0) as number,
    error: data.error as string | undefined,
    modeUsed: data.mode_used as 'map' | 'deep' | undefined,
    jobId: data.job_id as string | undefined,
    status: data.status as string | undefined,
    generatedConfig: data.generated_config
      ? generatedConfigFromDict(data.generated_config as Record<string, unknown>)
      : undefined,
    message: data.message as string | undefined,
  };
}

/**
 * Create DeepCrawlResult from API response.
 */
export function deepCrawlResultFromDict(data: Record<string, unknown>): DeepCrawlResult {
  let urls: ScanUrlInfo[] | undefined;
  if (data.urls && Array.isArray(data.urls)) {
    urls = (data.urls as Record<string, unknown>[]).map((u) => ({
      url: (u.url || '') as string,
      depth: (u.depth || 0) as number,
      score: u.score as number | undefined,
      linksFound: (u.links_found || 0) as number,
      htmlSize: (u.html_size || 0) as number,
    }));
  }

  return {
    jobId: (data.job_id || '') as string,
    status: (data.status || '') as string,
    strategy: (data.strategy || 'map') as string,
    discoveredCount: (data.discovered_urls || 0) as number,
    queuedUrls: (data.queued_urls || 0) as number,
    createdAt: (data.created_at || '') as string,
    urls,
    htmlDownloadUrl: data.html_download_url as string | undefined,
    cacheExpiresAt: data.cache_expires_at as string | undefined,
    crawlJobId: data.crawl_job_id as string | undefined,
  };
}

/**
 * Check if deep crawl result is complete.
 */
export function isDeepCrawlComplete(result: DeepCrawlResult): boolean {
  return ['completed', 'failed', 'cancelled'].includes(result.status);
}

/**
 * Context API response.
 */
export interface ContextResult {
  jobId: string;
  status: string;
  query: string;
  downloadUrl: string;
  urlsCrawled: number;
  sizeBytes: number;
  durationMs: number;
  cached: boolean;
}

/**
 * Create ContextResult from API response.
 */
export function contextResultFromDict(data: Record<string, unknown>): ContextResult {
  return {
    jobId: data.job_id as string,
    status: data.status as string,
    query: data.query as string,
    downloadUrl: data.download_url as string,
    sizeBytes: (data.storage_size_bytes || 0) as number,
    urlsCrawled: (data.urls_crawled || 0) as number,
    durationMs: (data.duration_ms || 0) as number,
    cached: (data.cached || false) as boolean,
  };
}

/**
 * LLM token usage (per-request).
 */
export interface LLMUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
}

/**
 * Crawl usage metrics returned in API responses.
 */
export interface CrawlUsageMetrics {
  creditsUsed: number;
  creditsRemaining: number;
  durationMs: number;
  cached: boolean | number; // bool for single, number for batch (cache hit count)
  urlsTotal?: number;
  urlsSucceeded?: number;
  urlsFailed?: number;
}

/**
 * LLM usage metrics returned in API responses.
 */
export interface LLMUsageMetrics {
  tokensUsed: number;
  tokensRemaining: number;
  model?: string;
}

/**
 * Storage usage metrics returned in API responses (async jobs only).
 */
export interface StorageUsageMetrics {
  bytesUsed: number;
  bytesRemaining: number;
}

/**
 * Unified usage metrics returned in API responses.
 */
export interface Usage {
  crawl: CrawlUsageMetrics;
  llm?: LLMUsageMetrics;
  storage?: StorageUsageMetrics;
}

/**
 * Create CrawlUsageMetrics from API response.
 */
export function crawlUsageMetricsFromDict(data: Record<string, unknown>): CrawlUsageMetrics {
  return {
    creditsUsed: (data.credits_used || 0) as number,
    creditsRemaining: (data.credits_remaining || 0) as number,
    durationMs: (data.duration_ms || 0) as number,
    cached: data.cached as boolean | number,
    urlsTotal: data.urls_total as number | undefined,
    urlsSucceeded: data.urls_succeeded as number | undefined,
    urlsFailed: data.urls_failed as number | undefined,
  };
}

/**
 * Create LLMUsageMetrics from API response.
 */
export function llmUsageMetricsFromDict(data: Record<string, unknown>): LLMUsageMetrics {
  return {
    tokensUsed: (data.tokens_used || 0) as number,
    tokensRemaining: (data.tokens_remaining || 0) as number,
    model: data.model as string | undefined,
  };
}

/**
 * Create StorageUsageMetrics from API response.
 */
export function storageUsageMetricsFromDict(data: Record<string, unknown>): StorageUsageMetrics {
  return {
    bytesUsed: (data.bytes_used || 0) as number,
    bytesRemaining: (data.bytes_remaining || 0) as number,
  };
}

/**
 * Create Usage from API response.
 */
export function usageFromDict(data: Record<string, unknown>): Usage {
  const crawlData = (data.crawl || {}) as Record<string, unknown>;

  let llm: LLMUsageMetrics | undefined;
  if (data.llm) {
    llm = llmUsageMetricsFromDict(data.llm as Record<string, unknown>);
  }

  let storage: StorageUsageMetrics | undefined;
  if (data.storage) {
    storage = storageUsageMetricsFromDict(data.storage as Record<string, unknown>);
  }

  return {
    crawl: crawlUsageMetricsFromDict(crawlData),
    llm,
    storage,
  };
}

/**
 * Generated extraction schema.
 */
export interface GeneratedSchema {
  success: boolean;
  schema?: Record<string, unknown>;
  error?: string;
  llmUsage?: LLMUsage;
}

/**
 * Create GeneratedSchema from API response.
 */
export function generatedSchemaFromDict(data: Record<string, unknown>): GeneratedSchema {
  let llmUsage: LLMUsage | undefined;
  if (data.llm_usage) {
    const usage = data.llm_usage as Record<string, unknown>;
    llmUsage = {
      promptTokens: (usage.prompt_tokens || 0) as number,
      completionTokens: (usage.completion_tokens || 0) as number,
      totalTokens: (usage.total_tokens || 0) as number,
    };
  }

  return {
    success: (data.success || false) as boolean,
    schema: data.schema as Record<string, unknown> | undefined,
    error: data.error_message as string | undefined,
    llmUsage,
  };
}

/**
 * Storage quota usage.
 */
export interface StorageUsage {
  usedMb: number;
  maxMb: number;
  remainingMb: number;
  percentUsed: number;
}

/**
 * Create StorageUsage from API response.
 */
export function storageUsageFromDict(data: Record<string, unknown>): StorageUsage {
  return {
    usedMb: (data.used_mb || 0) as number,
    maxMb: (data.max_mb || 0) as number,
    remainingMb: (data.remaining_mb || 0) as number,
    percentUsed: (data.percent_used || 0) as number,
  };
}

/**
 * Markdown extraction result.
 */
export interface MarkdownResult {
  rawMarkdown?: string;
  markdownWithCitations?: string;
  referencesMarkdown?: string;
  fitMarkdown?: string;
}

/**
 * Single URL crawl result.
 */
export interface CrawlResult {
  url: string;
  success: boolean;
  html?: string;
  cleanedHtml?: string;
  fitHtml?: string;
  markdown?: MarkdownResult;
  media: Record<string, unknown[]>;
  links: Record<string, unknown[]>;
  metadata?: Record<string, unknown>;
  screenshot?: string;
  pdf?: string;
  extractedContent?: string;
  errorMessage?: string;
  statusCode?: number;
  durationMs: number;
  tables: unknown[];
  networkRequests?: unknown[];
  consoleMessages?: unknown[];
  redirectedUrl?: string;
  llmUsage?: LLMUsage;
  crawlStrategy?: string;
  /** Presigned S3 URLs for file downloads (CSV, PDF, XLSX, etc.) */
  downloadedFiles?: string[];
  /** Job ID for async results (use with downloadUrl()) */
  id?: string;
  /** Resource usage metrics */
  usage?: Usage;
}

/**
 * Create CrawlResult from API response.
 */
export function crawlResultFromDict(data: Record<string, unknown>): CrawlResult {
  let markdown: MarkdownResult | undefined;
  const markdownData = data.markdown;
  if (markdownData) {
    // Handle both string (async results) and object (sync results) formats
    if (typeof markdownData === 'string') {
      markdown = { rawMarkdown: markdownData };
    } else {
      const md = markdownData as Record<string, unknown>;
      markdown = {
        rawMarkdown: md.raw_markdown as string | undefined,
        markdownWithCitations: md.markdown_with_citations as string | undefined,
        referencesMarkdown: md.references_markdown as string | undefined,
        fitMarkdown: md.fit_markdown as string | undefined,
      };
    }
  }

  let llmUsage: LLMUsage | undefined;
  if (data.llm_usage) {
    const usage = data.llm_usage as Record<string, unknown>;
    llmUsage = {
      promptTokens: (usage.prompt_tokens || 0) as number,
      completionTokens: (usage.completion_tokens || 0) as number,
      totalTokens: (usage.total_tokens || 0) as number,
    };
  }

  return {
    url: (data.url || '') as string,
    success: (data.success || false) as boolean,
    html: data.html as string | undefined,
    cleanedHtml: data.cleaned_html as string | undefined,
    fitHtml: data.fit_html as string | undefined,
    markdown,
    media: (data.media || {}) as Record<string, unknown[]>,
    links: (data.links || {}) as Record<string, unknown[]>,
    metadata: data.metadata as Record<string, unknown> | undefined,
    screenshot: data.screenshot as string | undefined,
    pdf: data.pdf as string | undefined,
    extractedContent: data.extracted_content as string | undefined,
    errorMessage: data.error_message as string | undefined,
    statusCode: data.status_code as number | undefined,
    durationMs: (data.duration_ms || 0) as number,
    tables: (data.tables || []) as unknown[],
    networkRequests: data.network_requests as unknown[] | undefined,
    consoleMessages: data.console_messages as unknown[] | undefined,
    redirectedUrl: data.redirected_url as string | undefined,
    llmUsage,
    crawlStrategy: data.crawl_strategy as string | undefined,
    downloadedFiles: data.downloaded_files as string[] | undefined,
    usage: data.usage ? usageFromDict(data.usage as Record<string, unknown>) : undefined,
  };
}


// =============================================================================
// Wrapper API Types
// =============================================================================

export interface WrapperUsage {
  creditsUsed: number;
  creditsRemaining: number;
}

export interface MarkdownResponse {
  success: boolean;
  url: string;
  markdown?: string;
  fitMarkdown?: string;
  fitHtml?: string;
  links?: Record<string, unknown>;
  media?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  tables?: unknown[];
  durationMs: number;
  usage?: WrapperUsage;
  errorMessage?: string;
}

export interface ScreenshotResponse {
  success: boolean;
  url: string;
  screenshot?: string;
  pdf?: string;
  durationMs: number;
  usage?: WrapperUsage;
  errorMessage?: string;
}

export interface ExtractResponse {
  success: boolean;
  url?: string;
  data?: Record<string, unknown>[];
  methodUsed?: string;
  schemaUsed?: Record<string, unknown>;
  queryUsed?: string;
  llmUsage?: LLMUsage;
  durationMs: number;
  errorMessage?: string;
}

export interface MapUrlInfo {
  url: string;
  host: string;
  status: string;
  relevanceScore?: number;
  headData?: Record<string, unknown>;
}

export interface MapResponse {
  success: boolean;
  domain: string;
  totalUrls: number;
  hostsFound: number;
  mode: string;
  urls: MapUrlInfo[];
  durationMs: number;
  errorMessage?: string;
}

/**
 * Response from POST /v1/crawl/site.
 *
 * When `criteria` was in the request, `generatedConfig` carries the
 * LLM-generated scan + extract config. When `extract` was set,
 * `extractionMethodUsed` tells you whether CSS schema generation or LLM
 * extraction was picked, and `schemaUsed` holds the generated CSS schema
 * (if any) so you can reuse it on future crawls.
 *
 * Poll progress with `getSiteCrawlJob(jobId)`.
 */
export interface SiteCrawlResponse {
  jobId: string;
  status: string;
  strategy: string;
  discoveredUrls: number;
  queuedUrls: number;
  createdAt: string;
  generatedConfig?: GeneratedConfig;
  extractionMethodUsed?: 'llm' | 'css_schema';
  schemaUsed?: Record<string, unknown>;
}

/**
 * Progress block inside SiteCrawlJobStatus.
 */
export interface SiteCrawlProgress {
  urlsDiscovered: number;
  urlsCrawled: number;
  urlsFailed: number;
  total: number;
}

/**
 * Create SiteCrawlProgress from API response dict.
 */
export function siteCrawlProgressFromDict(
  data: Record<string, unknown> | undefined | null,
): SiteCrawlProgress {
  const d = data || {};
  return {
    urlsDiscovered: (d.urls_discovered || 0) as number,
    urlsCrawled: (d.urls_crawled || 0) as number,
    urlsFailed: (d.urls_failed || 0) as number,
    total: (d.total || 0) as number,
  };
}

/**
 * Polling response for GET /v1/crawl/site/jobs/{job_id}.
 *
 * This is the unified scan+crawl polling endpoint. `phase` walks through
 * three values: "scan" (URL discovery in progress), "crawl" (pages being
 * fetched + extracted), "done" (everything finished).
 */
export interface SiteCrawlJobStatus {
  jobId: string;
  status: string;
  phase: 'scan' | 'crawl' | 'done';
  progress: SiteCrawlProgress;
  scanJobId?: string;
  crawlJobId?: string;
  downloadUrl?: string;
  createdAt?: string;
  completedAt?: string;
  error?: string;
}

/**
 * Create SiteCrawlJobStatus from API response dict.
 */
export function siteCrawlJobStatusFromDict(data: Record<string, unknown>): SiteCrawlJobStatus {
  return {
    jobId: (data.job_id || '') as string,
    status: (data.status || 'pending') as string,
    phase: (data.phase || 'scan') as 'scan' | 'crawl' | 'done',
    progress: siteCrawlProgressFromDict(data.progress as Record<string, unknown> | undefined),
    scanJobId: data.scan_job_id as string | undefined,
    crawlJobId: data.crawl_job_id as string | undefined,
    downloadUrl: data.download_url as string | undefined,
    createdAt: data.created_at as string | undefined,
    completedAt: data.completed_at as string | undefined,
    error: data.error as string | undefined,
  };
}

/**
 * True when a site crawl job has reached a terminal state.
 */
export function isSiteCrawlJobComplete(job: SiteCrawlJobStatus): boolean {
  return (
    job.phase === 'done' ||
    ['completed', 'partial', 'failed', 'cancelled'].includes(job.status)
  );
}

export interface WrapperJobProgress {
  total: number;
  completed: number;
  failed: number;
}

export interface WrapperJob {
  jobId: string;
  status: string;
  progress?: WrapperJobProgress;
  progressPercent: number;
  urlsCount: number;
  error?: string;
  createdAt?: string;
  startedAt?: string;
  completedAt?: string;
}

export interface MarkdownOptions {
  strategy?: 'browser' | 'http';
  fit?: boolean;
  include?: string[];
  crawlerConfig?: Record<string, unknown>;
  browserConfig?: Record<string, unknown>;
  proxy?: Record<string, unknown>;
  bypassCache?: boolean;
}

export interface MarkdownManyOptions extends MarkdownOptions {
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
  webhookUrl?: string;
  priority?: number;
}

export interface ScreenshotOptions {
  fullPage?: boolean;
  pdf?: boolean;
  waitFor?: string;
  crawlerConfig?: Record<string, unknown>;
  browserConfig?: Record<string, unknown>;
  proxy?: Record<string, unknown>;
  bypassCache?: boolean;
}

export interface ScreenshotManyOptions extends ScreenshotOptions {
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
  webhookUrl?: string;
  priority?: number;
}

export interface ExtractOptions {
  query?: string;
  jsonExample?: Record<string, unknown>;
  schema?: Record<string, unknown>;
  method?: 'auto' | 'llm' | 'schema';
  strategy?: 'browser' | 'http';
  crawlerConfig?: Record<string, unknown>;
  browserConfig?: Record<string, unknown>;
  llmConfig?: Record<string, unknown>;
  proxy?: Record<string, unknown>;
  bypassCache?: boolean;
}

export interface ExtractManyOptions extends Omit<ExtractOptions, 'method'> {
  method: 'llm' | 'schema';
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
  webhookUrl?: string;
  priority?: number;
}

export interface MapOptions {
  mode?: 'default' | 'deep';
  maxUrls?: number;
  includeSubdomains?: boolean;
  extractHead?: boolean;
  query?: string;
  scoreThreshold?: number;
  force?: boolean;
  proxy?: Record<string, unknown>;
}

export interface SiteCrawlOptions {
  maxPages?: number;
  discovery?: 'map' | 'bfs' | 'dfs' | 'best_first';
  strategy?: 'browser' | 'http';
  fit?: boolean;
  include?: string[];
  pattern?: string;
  maxDepth?: number;
  crawlerConfig?: Record<string, unknown>;
  browserConfig?: Record<string, unknown>;
  proxy?: Record<string, unknown>;
  webhookUrl?: string;
  priority?: number;
  wait?: boolean;
  pollInterval?: number;
  timeout?: number;
  // AI-assisted fields
  criteria?: string;
  scan?: SiteScanConfig | Record<string, unknown>;
  extract?: SiteExtractConfig | Record<string, unknown>;
  includeMarkdown?: boolean;
}

// Wrapper fromDict helpers

function wrapperUsageFromDict(data: Record<string, unknown>): WrapperUsage {
  return {
    creditsUsed: (data.credits_used || 0) as number,
    creditsRemaining: (data.credits_remaining || 0) as number,
  };
}

export function markdownResponseFromDict(data: Record<string, unknown>): MarkdownResponse {
  return {
    success: (data.success || false) as boolean,
    url: (data.url || '') as string,
    markdown: data.markdown as string | undefined,
    fitMarkdown: data.fit_markdown as string | undefined,
    fitHtml: data.fit_html as string | undefined,
    links: data.links as Record<string, unknown> | undefined,
    media: data.media as Record<string, unknown> | undefined,
    metadata: data.metadata as Record<string, unknown> | undefined,
    tables: data.tables as unknown[] | undefined,
    durationMs: (data.duration_ms || 0) as number,
    usage: data.usage ? wrapperUsageFromDict(data.usage as Record<string, unknown>) : undefined,
    errorMessage: data.error_message as string | undefined,
  };
}

export function screenshotResponseFromDict(data: Record<string, unknown>): ScreenshotResponse {
  return {
    success: (data.success || false) as boolean,
    url: (data.url || '') as string,
    screenshot: data.screenshot as string | undefined,
    pdf: data.pdf as string | undefined,
    durationMs: (data.duration_ms || 0) as number,
    usage: data.usage ? wrapperUsageFromDict(data.usage as Record<string, unknown>) : undefined,
    errorMessage: data.error_message as string | undefined,
  };
}

export function extractResponseFromDict(data: Record<string, unknown>): ExtractResponse {
  let llmUsage: LLMUsage | undefined;
  if (data.llm_usage) {
    const u = data.llm_usage as Record<string, unknown>;
    llmUsage = {
      promptTokens: (u.prompt_tokens || 0) as number,
      completionTokens: (u.completion_tokens || 0) as number,
      totalTokens: (u.total_tokens || 0) as number,
    };
  }
  return {
    success: (data.success || false) as boolean,
    url: data.url as string | undefined,
    data: data.data as Record<string, unknown>[] | undefined,
    methodUsed: data.method_used as string | undefined,
    schemaUsed: data.schema_used as Record<string, unknown> | undefined,
    queryUsed: data.query_used as string | undefined,
    llmUsage,
    durationMs: (data.duration_ms || 0) as number,
    errorMessage: data.error_message as string | undefined,
  };
}

export function mapResponseFromDict(data: Record<string, unknown>): MapResponse {
  const urls = ((data.urls || []) as Record<string, unknown>[]).map((u) => ({
    url: (u.url || '') as string,
    host: (u.host || '') as string,
    status: (u.status || 'valid') as string,
    relevanceScore: u.relevance_score as number | undefined,
    headData: u.head_data as Record<string, unknown> | undefined,
  }));
  return {
    success: (data.success || false) as boolean,
    domain: (data.domain || '') as string,
    totalUrls: (data.total_urls || 0) as number,
    hostsFound: (data.hosts_found || 0) as number,
    mode: (data.mode || 'default') as string,
    urls,
    durationMs: (data.duration_ms || 0) as number,
    errorMessage: data.error_message as string | undefined,
  };
}

export function siteCrawlResponseFromDict(data: Record<string, unknown>): SiteCrawlResponse {
  return {
    jobId: (data.job_id || '') as string,
    status: (data.status || 'pending') as string,
    strategy: (data.strategy || 'map') as string,
    discoveredUrls: (data.discovered_urls || 0) as number,
    queuedUrls: (data.queued_urls || 0) as number,
    createdAt: (data.created_at || '') as string,
    generatedConfig: data.generated_config
      ? generatedConfigFromDict(data.generated_config as Record<string, unknown>)
      : undefined,
    extractionMethodUsed: data.extraction_method_used as 'llm' | 'css_schema' | undefined,
    schemaUsed: data.schema_used as Record<string, unknown> | undefined,
  };
}

export function wrapperJobFromDict(data: Record<string, unknown>): WrapperJob {
  let progress: WrapperJobProgress | undefined;
  if (data.progress) {
    const p = data.progress as Record<string, unknown>;
    progress = {
      total: (p.total || 0) as number,
      completed: (p.completed || 0) as number,
      failed: (p.failed || 0) as number,
    };
  }
  return {
    jobId: (data.job_id || '') as string,
    status: (data.status || 'pending') as string,
    progress,
    progressPercent: (data.progress_percent || 0) as number,
    urlsCount: (data.urls_count || 0) as number,
    error: data.error as string | undefined,
    createdAt: data.created_at as string | undefined,
    startedAt: data.started_at as string | undefined,
    completedAt: data.completed_at as string | undefined,
  };
}

export function isWrapperJobComplete(job: WrapperJob): boolean {
  return ['completed', 'partial', 'failed', 'cancelled'].includes(job.status);
}
