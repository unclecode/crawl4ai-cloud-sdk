/**
 * Crawl4AI Cloud SDK - Lightweight cloud client for Crawl4AI API.
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

export const VERSION = '0.4.0';

// Main crawler class
export { AsyncWebCrawler } from './crawler';
export type {
  AsyncWebCrawlerOptions,
  RunOptions,
  RunManyOptions,
  DeepCrawlOptions,
} from './crawler';

// Wrapper API types
export type {
  WrapperUsage,
  MarkdownResponse,
  ScreenshotResponse,
  ExtractResponse,
  MapUrlInfo,
  MapResponse,
  SiteCrawlResponse,
  SiteCrawlJobStatus,
  SiteCrawlProgress,
  SiteScanConfig,
  SiteExtractConfig,
  GeneratedConfig,
  ScanJobStatus,
  WrapperJob,
  WrapperJobProgress,
  MarkdownOptions,
  MarkdownManyOptions,
  ScreenshotOptions,
  ScreenshotManyOptions,
  ExtractOptions,
  ExtractManyOptions,
  MapOptions,
  SiteCrawlOptions,
  // Enrich API types
  EnrichFieldSource,
  EnrichSearchCitation,
  EnrichRow,
  EnrichJobProgress,
  EnrichResponse,
  EnrichJobStatus,
} from './models';

// Configuration types and helpers
export type { CrawlerRunConfig, BrowserConfig } from './configs';
export {
  sanitizeCrawlerConfig,
  sanitizeBrowserConfig,
  normalizeProxy,
  buildCrawlRequest,
} from './configs';

// Response models
export type {
  CrawlResult,
  CrawlJob,
  JobProgress,
  MarkdownResult,
  DeepCrawlResult,
  ScanUrlInfo,
  ScanResult,
  ScanOptions,
  DomainScanUrlInfo,
  ContextResult,
  GeneratedSchema,
  StorageUsage,
  ProxyConfig,
  LLMUsage,
  // Usage metrics
  Usage,
  CrawlUsageMetrics,
  LLMUsageMetrics,
  StorageUsageMetrics,
} from './models';

export {
  crawlResultFromDict,
  crawlJobFromDict,
  deepCrawlResultFromDict,
  scanResultFromDict,
  contextResultFromDict,
  generatedSchemaFromDict,
  storageUsageFromDict,
  isJobComplete,
  isJobSuccessful,
  isDeepCrawlComplete,
  getProgressPending,
  getProgressPercent,
  // Usage helpers
  usageFromDict,
  crawlUsageMetricsFromDict,
  llmUsageMetricsFromDict,
  storageUsageMetricsFromDict,
  // Wrapper helpers
  markdownResponseFromDict,
  screenshotResponseFromDict,
  extractResponseFromDict,
  mapResponseFromDict,
  siteCrawlResponseFromDict,
  wrapperJobFromDict,
  isWrapperJobComplete,
  // AI-assisted scan + site crawl helpers
  generatedConfigFromDict,
  scanJobStatusFromDict,
  siteCrawlJobStatusFromDict,
  siteCrawlProgressFromDict,
  isScanJobComplete,
  isScanResultAsync,
  isSiteCrawlJobComplete,
  siteScanConfigToDict,
  siteExtractConfigToDict,
  // Enrich helpers
  enrichResponseFromDict,
  enrichJobStatusFromDict,
  isEnrichJobComplete,
  isEnrichJobSuccessful,
} from './models';

// Errors
export {
  CloudError,
  AuthenticationError,
  RateLimitError,
  QuotaExceededError,
  NotFoundError,
  ValidationError,
  TimeoutError,
  ServerError,
} from './errors';
export type { ErrorResponse, ErrorHeaders } from './errors';
