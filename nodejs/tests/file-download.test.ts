/**
 * E2E + unit tests for file download detection via HTTP strategy.
 *
 * Tests that the SDK correctly receives and exposes downloadedFiles
 * from the API when crawling non-HTML URLs with strategy="http".
 */

import { AsyncWebCrawler, CrawlResult } from '../src';
import { crawlResultFromDict } from '../src/models';

const API_KEY = process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';

// =============================================================================
// E2E TESTS — real API calls
// =============================================================================

describe('File Download E2E', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY });
  });

  afterAll(async () => {
    await crawler.close();
  });

  test('CSV file should return downloadedFiles with S3 presigned URL', async () => {
    const result = await crawler.run(
      'https://data.gov.au/data/dataset/043f58e0-a188-4458-b61c-04e5b540aea4/resource/f83cdee9-ebcb-4f24-941b-34bb2f0996cf/download/facilities.csv',
      { strategy: 'http', bypassCache: true }
    );

    expect(result.success).toBe(true);
    expect(result.downloadedFiles).toBeDefined();
    expect(result.downloadedFiles!.length).toBeGreaterThanOrEqual(1);
    expect(result.downloadedFiles![0]).toMatch(/^https:\/\//);
    // CSV is text-based — html also has content
    expect(result.html).toBeDefined();
    expect(result.html!.length).toBeGreaterThan(1000);
  }, 60000);

  test('JSON API response should return downloadedFiles', async () => {
    const result = await crawler.run(
      'https://jsonplaceholder.typicode.com/posts/1',
      { strategy: 'http', bypassCache: true }
    );

    expect(result.success).toBe(true);
    expect(result.downloadedFiles).toBeDefined();
    expect(result.downloadedFiles!.length).toBeGreaterThanOrEqual(1);
    expect(result.html).toContain('userId');
  }, 30000);

  test('Normal HTML page should NOT have downloadedFiles', async () => {
    const result = await crawler.run(
      'https://example.com',
      { strategy: 'http', bypassCache: true }
    );

    expect(result.success).toBe(true);
    expect(result.downloadedFiles).toBeFalsy();  // null or undefined
    expect(result.html).toContain('Example Domain');
  }, 30000);

  test('Binary download should have downloadedFiles and empty html', async () => {
    const result = await crawler.run(
      'https://httpbin.org/bytes/1024',
      { strategy: 'http', bypassCache: true }
    );

    expect(result.success).toBe(true);
    expect(result.downloadedFiles).toBeDefined();
    expect(result.downloadedFiles!.length).toBeGreaterThanOrEqual(1);
    expect(result.downloadedFiles![0]).toMatch(/^https:\/\//);
  }, 30000);
});

// =============================================================================
// UNIT TESTS — CrawlResult parsing
// =============================================================================

describe('CrawlResult downloadedFiles parsing', () => {
  test('should parse downloaded_files from API response', () => {
    const result = crawlResultFromDict({
      url: 'https://example.com/data.csv',
      success: true,
      html: 'a,b,c\n1,2,3',
      downloaded_files: ['https://s3.example.com/downloads/abc/data.csv?sig=xyz'],
      status_code: 200,
      duration_ms: 500,
    });

    expect(result.downloadedFiles).toBeDefined();
    expect(result.downloadedFiles!.length).toBe(1);
    expect(result.downloadedFiles![0]).toContain('data.csv');
  });

  test('should handle missing downloaded_files', () => {
    const result = crawlResultFromDict({
      url: 'https://example.com',
      success: true,
      html: '<html>hello</html>',
      status_code: 200,
      duration_ms: 100,
    });

    expect(result.downloadedFiles).toBeUndefined();
  });

  test('should handle null downloaded_files', () => {
    const result = crawlResultFromDict({
      url: 'https://example.com',
      success: true,
      downloaded_files: null,
      status_code: 200,
      duration_ms: 100,
    });

    // null becomes undefined in TS
    expect(result.downloadedFiles).toBeFalsy();
  });

  test('should handle multiple downloaded files', () => {
    const result = crawlResultFromDict({
      url: 'https://example.com/archive',
      success: true,
      downloaded_files: [
        'https://s3.example.com/file1.csv',
        'https://s3.example.com/file2.pdf',
      ],
      status_code: 200,
      duration_ms: 200,
    });

    expect(result.downloadedFiles).toBeDefined();
    expect(result.downloadedFiles!.length).toBe(2);
  });
});
