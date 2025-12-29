/**
 * Configuration classes and sanitization for Crawl4AI Cloud SDK.
 */

import { ProxyConfig } from './models';

// Fields that cloud controls - removed from CrawlerRunConfig
const CRAWLER_CONFIG_SANITIZE_FIELDS = [
  'cache_mode',
  'cacheMode',
  'session_id',
  'sessionId',
  'bypass_cache',
  'bypassCache',
  'no_cache_read',
  'noCacheRead',
  'no_cache_write',
  'noCacheWrite',
  'disable_cache',
  'disableCache',
];

// Fields that cloud controls - removed from BrowserConfig
const BROWSER_CONFIG_SANITIZE_FIELDS = [
  'cdp_url',
  'cdpUrl',
  'create_isolated_context',
  'createIsolatedContext',
  'cdp_cleanup_on_close',
  'cdpCleanupOnClose',
  'browser_context_id',
  'browserContextId',
  'target_id',
  'targetId',
  'use_managed_browser',
  'useManagedBrowser',
  'browser_mode',
  'browserMode',
  'user_data_dir',
  'userDataDir',
  'chrome_channel',
  'chromeChannel',
];

/**
 * Configuration for crawl requests. Mirrors OSS CrawlerRunConfig.
 */
export interface CrawlerRunConfig {
  // Content processing
  wordCountThreshold?: number;
  excludeExternalLinks?: boolean;
  excludeSocialMediaLinks?: boolean;
  excludeExternalImages?: boolean;
  excludeDomains?: string[];

  // HTML processing
  processIframes?: boolean;
  removeForms?: boolean;
  keepDataAttributes?: boolean;

  // Output options
  onlyText?: boolean;
  prettiify?: boolean;

  // Screenshot/PDF
  screenshot?: boolean;
  screenshotWaitFor?: string;
  pdf?: boolean;

  // Extraction
  extractionStrategy?: unknown;
  chunkingStrategy?: unknown;
  contentFilter?: unknown;

  // Markdown generation
  markdownGenerator?: unknown;

  // Wait conditions
  waitFor?: string;
  delayBeforeReturnHtml?: number;

  // Page interaction
  jsCode?: string | string[];
  jsOnly?: boolean;
  ignoreBodyVisibility?: boolean;
  scanFullPage?: boolean;
  scrollDelay?: number;

  // Network
  waitForImages?: boolean;
  adjustViewportToContent?: boolean;
  pageTimeout?: number;

  // Cache (cloud-controlled, will be stripped)
  cacheMode?: string;
  sessionId?: string;
  bypassCache?: boolean;
  noCacheRead?: boolean;
  noCacheWrite?: boolean;
  disableCache?: boolean;

  // Magic mode
  magic?: boolean;

  // Simulate user
  simulateUser?: boolean;
  overrideNavigator?: boolean;

  // Allow additional fields
  [key: string]: unknown;
}

/**
 * Browser configuration for crawl requests. Mirrors OSS BrowserConfig.
 */
export interface BrowserConfig {
  // Browser settings
  headless?: boolean;
  browserType?: string;
  verbose?: boolean;

  // Viewport
  viewportWidth?: number;
  viewportHeight?: number;

  // User agent
  userAgent?: string;
  userAgentMode?: string;
  userAgentGeneratorConfig?: Record<string, unknown>;

  // Headers & cookies
  headers?: Record<string, string>;
  cookies?: Record<string, unknown>[];

  // Storage state
  storageState?: string;

  // Proxy
  proxy?: string;
  proxyConfig?: Record<string, unknown>;

  // Browser args
  extraArgs?: string[];
  chromeChannel?: string;
  acceptDownloads?: boolean;
  downloadsPath?: string;

  // HTTPS errors
  ignoreHttpsErrors?: boolean;
  javaScriptEnabled?: boolean;

  // Cloud-controlled fields (will be stripped)
  cdpUrl?: string;
  useManagedBrowser?: boolean;
  browserMode?: string;
  userDataDir?: string;

  // Text mode
  textMode?: boolean;
  lightMode?: boolean;

  // Allow additional fields
  [key: string]: unknown;
}

/**
 * Convert camelCase to snake_case.
 */
function toSnakeCase(str: string): string {
  return str.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`);
}

/**
 * Convert object keys from camelCase to snake_case.
 */
function keysToSnakeCase(obj: Record<string, unknown>): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    const snakeKey = toSnakeCase(key);
    result[snakeKey] = value;
  }
  return result;
}

/**
 * Sanitize CrawlerRunConfig for cloud API.
 */
export function sanitizeCrawlerConfig(
  config?: CrawlerRunConfig | Record<string, unknown>
): Record<string, unknown> {
  if (!config) return {};

  // Convert to snake_case
  const data = keysToSnakeCase(config as Record<string, unknown>);

  // Remove cloud-controlled fields
  for (const field of CRAWLER_CONFIG_SANITIZE_FIELDS) {
    delete data[field];
    delete data[toSnakeCase(field)];
  }

  // Remove undefined/null values
  for (const key of Object.keys(data)) {
    if (data[key] === undefined || data[key] === null) {
      delete data[key];
    }
  }

  return data;
}

/**
 * Sanitize BrowserConfig for cloud API.
 */
export function sanitizeBrowserConfig(
  config?: BrowserConfig | Record<string, unknown>,
  strategy: string = 'browser'
): Record<string, unknown> {
  if (!config) return {};

  // Warn if browser config with HTTP strategy
  if (strategy === 'http') {
    console.warn(
      "browser_config is ignored when using strategy='http'. " +
        'Browser configuration only applies to browser-based crawling.'
    );
    return {};
  }

  // Convert to snake_case
  const data = keysToSnakeCase(config as Record<string, unknown>);

  // Remove cloud-controlled fields
  for (const field of BROWSER_CONFIG_SANITIZE_FIELDS) {
    delete data[field];
    delete data[toSnakeCase(field)];
  }

  // Remove undefined/null values
  for (const key of Object.keys(data)) {
    if (data[key] === undefined || data[key] === null) {
      delete data[key];
    }
  }

  return data;
}

/**
 * Normalize proxy configuration to dict format.
 */
export function normalizeProxy(
  proxy?: string | ProxyConfig | Record<string, unknown>
): Record<string, unknown> | undefined {
  if (!proxy) return undefined;

  if (typeof proxy === 'string') {
    return { mode: proxy };
  }

  if (typeof proxy === 'object') {
    return proxy as Record<string, unknown>;
  }

  throw new Error(
    `Invalid proxy type: ${typeof proxy}. Expected string or object.`
  );
}

/**
 * Build a crawl request body for the cloud API.
 */
export function buildCrawlRequest(options: {
  url?: string;
  urls?: string[];
  config?: CrawlerRunConfig | Record<string, unknown>;
  browserConfig?: BrowserConfig | Record<string, unknown>;
  strategy?: string;
  proxy?: string | ProxyConfig | Record<string, unknown>;
  bypassCache?: boolean;
  [key: string]: unknown;
}): Record<string, unknown> {
  const {
    url,
    urls,
    config,
    browserConfig,
    strategy = 'browser',
    proxy,
    bypassCache,
    ...rest
  } = options;

  const body: Record<string, unknown> = { strategy };

  if (url) body.url = url;
  if (urls) body.urls = urls;

  // Sanitize and add configs
  const crawlerConfig = sanitizeCrawlerConfig(config);
  if (Object.keys(crawlerConfig).length > 0) {
    body.crawler_config = crawlerConfig;
  }

  const browserCfg = sanitizeBrowserConfig(browserConfig, strategy);
  if (Object.keys(browserCfg).length > 0) {
    body.browser_config = browserCfg;
  }

  // Normalize and add proxy
  const proxyConfig = normalizeProxy(proxy);
  if (proxyConfig) {
    body.proxy = proxyConfig;
  }

  if (bypassCache) {
    body.bypass_cache = true;
  }

  // Add any additional options
  Object.assign(body, rest);

  return body;
}
