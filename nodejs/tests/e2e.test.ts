/**
 * Comprehensive E2E Tests for Crawl4AI Cloud SDK (Node.js).
 */

import {
  AsyncWebCrawler,
  CrawlResult,
  CrawlJob,
  DeepCrawlResult,
  StorageUsage,
  CloudError,
  AuthenticationError,
  NotFoundError,
  sanitizeCrawlerConfig,
  sanitizeBrowserConfig,
  normalizeProxy,
  CrawlerRunConfig,
  BrowserConfig,
} from '../src';

// Test API key
const API_KEY = process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';

const TEST_URL = 'https://example.com';
const TEST_URL_2 = 'https://httpbin.org/html';

// =============================================================================
// INITIALIZATION TESTS
// =============================================================================

describe('AsyncWebCrawler Initialization', () => {
  test('should initialize with API key', () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY });
    expect(crawler).toBeDefined();
  });

  test('should throw error without API key', () => {
    const originalEnv = process.env.CRAWL4AI_API_KEY;
    delete process.env.CRAWL4AI_API_KEY;

    expect(() => {
      new AsyncWebCrawler({});
    }).toThrow('API key is required');

    process.env.CRAWL4AI_API_KEY = originalEnv;
  });

  test('should throw error with invalid API key format', () => {
    expect(() => {
      new AsyncWebCrawler({ apiKey: 'invalid_key' });
    }).toThrow('Invalid API key format');
  });

  test('should accept sk_test_ prefix', () => {
    const crawler = new AsyncWebCrawler({ apiKey: 'sk_test_dummy_12345' });
    expect(crawler).toBeDefined();
  });

  test('should accept custom base URL', () => {
    const crawler = new AsyncWebCrawler({
      apiKey: API_KEY,
      baseUrl: 'https://api.crawl4ai.com',
    });
    expect(crawler).toBeDefined();
  });
});

// =============================================================================
// SINGLE URL CRAWL TESTS
// =============================================================================

describe('Single URL Crawl', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should crawl a single URL successfully', async () => {
    const result = await crawler.run(TEST_URL);

    expect(result).toBeDefined();
    expect(result.success).toBe(true);
    expect(result.url).toBe(TEST_URL);
  });

  test('should return markdown content', async () => {
    const result = await crawler.run(TEST_URL);

    expect(result.markdown).toBeDefined();
    expect(result.markdown?.rawMarkdown).toBeDefined();
    expect(result.markdown?.rawMarkdown?.length).toBeGreaterThan(0);
    expect(result.markdown?.rawMarkdown).toContain('Example Domain');
  });

  test('should return HTML content', async () => {
    const result = await crawler.run(TEST_URL);

    expect(result.html).toBeDefined();
    expect(result.html?.toLowerCase()).toContain('<html');
  });

  test('should work with browser strategy', async () => {
    const result = await crawler.run(TEST_URL, { strategy: 'browser' });

    expect(result.success).toBe(true);
  });

  test('should work with http strategy', async () => {
    const result = await crawler.run(TEST_URL, { strategy: 'http' });

    expect(result.success).toBe(true);
  });

  test('should work with bypass cache', async () => {
    const result = await crawler.run(TEST_URL, { bypassCache: true });

    expect(result.success).toBe(true);
  });
});

// =============================================================================
// OSS COMPATIBILITY TESTS
// =============================================================================

describe('OSS Compatibility', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('arun() should be alias for run()', async () => {
    const result = await crawler.arun(TEST_URL);

    expect(result).toBeDefined();
    expect(result.success).toBe(true);
    expect(result.url).toBe(TEST_URL);
  });

  test('arunMany() should be alias for runMany()', async () => {
    const urls = [TEST_URL, TEST_URL_2];
    const results = await crawler.arunMany(urls, { wait: true });

    expect(Array.isArray(results)).toBe(true);
    expect((results as CrawlResult[]).length).toBe(2);
  });
});

// =============================================================================
// CONFIGURATION TESTS
// =============================================================================

describe('CrawlerRunConfig', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should work with config options', async () => {
    const config: CrawlerRunConfig = {
      wordCountThreshold: 10,
      excludeExternalLinks: true,
    };

    const result = await crawler.run(TEST_URL, { config });

    expect(result.success).toBe(true);
  });

  test('should sanitize cache fields from config', () => {
    const config: CrawlerRunConfig = {
      cacheMode: 'bypass',
      sessionId: 'test-session',
      screenshot: true,
    };

    const sanitized = sanitizeCrawlerConfig(config);

    expect(sanitized.cache_mode).toBeUndefined();
    expect(sanitized.session_id).toBeUndefined();
    expect(sanitized.screenshot).toBe(true);
  });
});

describe('BrowserConfig', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should work with browser config', async () => {
    const browserConfig: BrowserConfig = {
      viewportWidth: 1920,
      viewportHeight: 1080,
    };

    const result = await crawler.run(TEST_URL, { browserConfig });

    expect(result.success).toBe(true);
  });

  test('should sanitize CDP fields from browser config', () => {
    const config: BrowserConfig = {
      cdpUrl: 'ws://localhost:9222',
      useManagedBrowser: true,
      headless: false,
    };

    const sanitized = sanitizeBrowserConfig(config);

    expect(sanitized.cdp_url).toBeUndefined();
    expect(sanitized.use_managed_browser).toBeUndefined();
    expect(sanitized.headless).toBe(false);
  });
});

// =============================================================================
// PROXY CONFIGURATION TESTS
// =============================================================================

describe('Proxy Configuration', () => {
  test('should normalize string proxy', () => {
    const result = normalizeProxy('datacenter');
    expect(result).toEqual({ mode: 'datacenter' });
  });

  test('should normalize dict proxy', () => {
    const proxy = { mode: 'residential', country: 'US' };
    const result = normalizeProxy(proxy);
    expect(result).toEqual(proxy);
  });

  test('should return undefined for null proxy', () => {
    const result = normalizeProxy(undefined);
    expect(result).toBeUndefined();
  });

  test('should throw for invalid proxy type', () => {
    expect(() => {
      normalizeProxy(12345 as unknown as string);
    }).toThrow('Invalid proxy type');
  });
});

// =============================================================================
// BATCH CRAWL TESTS
// =============================================================================

describe('Batch Crawl', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should crawl multiple URLs with wait=true', async () => {
    const urls = [TEST_URL, TEST_URL_2];
    const results = await crawler.runMany(urls, { wait: true });

    expect(Array.isArray(results)).toBe(true);
    const resultArray = results as CrawlResult[];
    expect(resultArray.length).toBe(2);
    resultArray.forEach((result) => {
      expect(result.success).toBe(true);
    });
  });

  test('should return job with wait=false', async () => {
    const urls = [TEST_URL, TEST_URL_2];
    const job = await crawler.runMany(urls, { wait: false });

    expect((job as CrawlJob).id).toBeDefined();
    expect((job as CrawlJob).status).toBe('completed');
  });
});

// =============================================================================
// JOB MANAGEMENT TESTS
// =============================================================================

describe('Job Management', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should list jobs', async () => {
    const jobs = await crawler.listJobs({ limit: 5 });

    expect(Array.isArray(jobs)).toBe(true);
    jobs.forEach((job) => {
      expect(job.id).toBeDefined();
      expect(job.status).toBeDefined();
    });
  });

  test('should list jobs with status filter', async () => {
    const jobs = await crawler.listJobs({ status: 'completed', limit: 5 });

    expect(Array.isArray(jobs)).toBe(true);
    jobs.forEach((job) => {
      expect(job.status).toBe('completed');
    });
  });
});

// =============================================================================
// STORAGE API TESTS
// =============================================================================

describe('Storage API', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should get storage usage', async () => {
    const usage = await crawler.storage();

    expect(usage).toBeDefined();
    expect(usage.maxMb).toBeGreaterThanOrEqual(0);
    expect(usage.usedMb).toBeGreaterThanOrEqual(0);
    expect(usage.remainingMb).toBeGreaterThanOrEqual(0);
  });
});

// =============================================================================
// HEALTH CHECK TESTS
// =============================================================================

describe('Health Check', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should check API health', async () => {
    const health = await crawler.health();

    expect(health).toBeDefined();
  });
});

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

describe('Error Handling', () => {
  test('should throw AuthenticationError for invalid API key', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: 'sk_test_invalid_12345' });

    await expect(crawler.run(TEST_URL)).rejects.toThrow(AuthenticationError);

    await crawler.close();
  });

  test('should throw NotFoundError for non-existent job', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

    await expect(crawler.getJob('nonexistent-job-12345')).rejects.toThrow(
      NotFoundError
    );

    await crawler.close();
  });
});

// =============================================================================
// DEEP CRAWL TESTS
// =============================================================================

describe('Deep Crawl', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('should require url or sourceJob', async () => {
    await expect(crawler.deepCrawl()).rejects.toThrow(
      "Must provide either 'url' or 'sourceJob'"
    );
  });

  test('should reject both url and sourceJob', async () => {
    await expect(
      crawler.deepCrawl(TEST_URL, { sourceJob: 'some-job' })
    ).rejects.toThrow('not both');
  });

  test('should perform scan_only deep crawl', async () => {
    const result = await crawler.deepCrawl(TEST_URL, {
      strategy: 'bfs',
      maxDepth: 1,
      maxUrls: 5,
      scanOnly: true,
      wait: true,
    });

    expect(result).toBeDefined();
    expect((result as DeepCrawlResult).jobId).toBeDefined();
  });
});

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

describe('Integration', () => {
  test('should complete full crawl workflow', async () => {
    const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

    const config: CrawlerRunConfig = {
      wordCountThreshold: 10,
    };

    const browserConfig: BrowserConfig = {
      viewportWidth: 1280,
      viewportHeight: 720,
    };

    const result = await crawler.run(TEST_URL, {
      config,
      browserConfig,
      strategy: 'browser',
    });

    expect(result.success).toBe(true);
    expect(result.url).toBe(TEST_URL);
    expect(result.markdown?.rawMarkdown).toContain('Example');

    await crawler.close();
  });

  test('should complete OSS migration pattern', async () => {
    // This is how users migrate from OSS to Cloud:
    // 1. Change import
    // 2. Add API key
    // 3. Use same code

    const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

    // OSS users use arun()
    const result = await crawler.arun(TEST_URL);

    expect(result.success).toBe(true);
    expect(result.markdown?.rawMarkdown).toBeDefined();

    await crawler.close();
  });
});
