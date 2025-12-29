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
  id: string;
  status: string;
  progress: JobProgress;
  urlsCount: number;
  createdAt: string;
  startedAt?: string;
  completedAt?: string;
  results?: Record<string, unknown>[];
  error?: string;
  resultSizeBytes?: number;
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
 */
export function crawlJobFromDict(data: Record<string, unknown>): CrawlJob {
  const progressData = (data.progress || {}) as Record<string, unknown>;
  return {
    id: (data.job_id || '') as string,
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
    results: data.results as Record<string, unknown>[] | undefined,
    error: data.error as string | undefined,
    resultSizeBytes: data.result_size_bytes as number | undefined,
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
  return ['completed', 'failed'].includes(result.status);
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
 * LLM token usage.
 */
export interface LLMUsage {
  promptTokens: number;
  completionTokens: number;
  totalTokens: number;
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
}

/**
 * Create CrawlResult from API response.
 */
export function crawlResultFromDict(data: Record<string, unknown>): CrawlResult {
  let markdown: MarkdownResult | undefined;
  const markdownData = data.markdown as Record<string, unknown> | undefined;
  if (markdownData) {
    markdown = {
      rawMarkdown: markdownData.raw_markdown as string | undefined,
      markdownWithCitations: markdownData.markdown_with_citations as string | undefined,
      referencesMarkdown: markdownData.references_markdown as string | undefined,
      fitMarkdown: markdownData.fit_markdown as string | undefined,
    };
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
  };
}
