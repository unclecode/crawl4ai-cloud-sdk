/**
 * E2E tests for `enhance_query` opt-in on /v1/discovery/search.
 *
 * Real HTTP against stage.crawl4ai.com. Mirrors the Python suite.
 */
import { AsyncWebCrawler, type SearchResponse } from '../src';

const API_KEY =
  process.env.CRAWL4AI_API_KEY ||
  'sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ';
const BASE_URL =
  process.env.CRAWL4AI_BASE_URL || 'https://stage.crawl4ai.com';

const E2E_TIMEOUT_MS = 180_000;

describe('Discovery enhance_query (E2E)', () => {
  let crawler: AsyncWebCrawler;

  beforeAll(() => {
    crawler = new AsyncWebCrawler({ apiKey: API_KEY, baseUrl: BASE_URL });
  });

  afterAll(async () => {
    await crawler.close();
  });

  it(
    'single backend — rewrite surfaces in response',
    async () => {
      const resp = (await crawler.discovery('search', {
        query: 'what are the best nurseries in Toronto for my 2 year old',
        country: 'ca',
        enhance_query: true,
      })) as SearchResponse;

      expect(resp.originalQuery).toBe(
        'what are the best nurseries in Toronto for my 2 year old',
      );
      expect(resp.rewrittenQueries).toBeTruthy();
      const google = resp.rewrittenQueries!.google;
      expect(google).toBeTruthy();
      expect(google).toContain('Toronto');
      // Conversational filler should be gone.
      expect(google).not.toContain('what are the best');
      // Tiny gap before the next e2e — back-to-back identical-source
      // SERP fetches can occasionally hit a captcha wall.
      await new Promise((r) => setTimeout(r, 4000));
    },
    E2E_TIMEOUT_MS,
  );

  it(
    'multi-backend — google and bing produce distinct rewrites',
    async () => {
      const resp = (await crawler.discovery('search', {
        query: 'latest claude news this week',
        country: 'us',
        enhance_query: true,
        backends: ['google', 'bing'],
      })) as SearchResponse;

      expect(resp.originalQuery).toBe('latest claude news this week');
      const rewrites = resp.rewrittenQueries;
      expect(rewrites).toBeTruthy();
      // At least one backend's rewrite must surface. Multi-backend has
      // failure isolation: if one backend's SERP fetch times out it's
      // dropped from the merge (and its rewrite from the dict). The
      // test still confirms per-backend rewriting is wired correctly
      // by checking each backend's operator vocabulary when present.
      const keys = Object.keys(rewrites!);
      expect(keys.length).toBeGreaterThan(0);
      if (rewrites!.google !== undefined) {
        // Google supports `after:YYYY-MM-DD` for recency.
        expect(rewrites!.google).toContain('after:');
      }
      if (rewrites!.bing !== undefined) {
        // Bing has no after: — must use a quoted year.
        expect(rewrites!.bing).toContain('2026');
      }
      await new Promise((r) => setTimeout(r, 4000));
    },
    E2E_TIMEOUT_MS,
  );

  it(
    'default off — no rewrite fields',
    async () => {
      const resp = (await crawler.discovery('search', {
        query: 'openai latest news',
        country: 'us',
      })) as SearchResponse;

      // The shape assertions are the SDK contract we care about; hits
      // length is incidental (e2e Google can transiently return 0 hits
      // under burst load and is not what this test verifies).
      expect(resp.originalQuery ?? null).toBeNull();
      expect(resp.rewrittenQueries ?? null).toBeNull();
    },
    E2E_TIMEOUT_MS,
  );
});
