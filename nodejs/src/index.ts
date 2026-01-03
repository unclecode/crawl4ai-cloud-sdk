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

export const VERSION = '0.1.1';

// Main crawler class
export { AsyncWebCrawler } from './crawler';
export type {
  AsyncWebCrawlerOptions,
  RunOptions,
  RunManyOptions,
  DeepCrawlOptions,
} from './crawler';

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
