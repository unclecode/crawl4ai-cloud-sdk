#!/usr/bin/env npx ts-node
/**
 * Crawl4AI Cloud - Proxy Usage Examples (TypeScript)
 *
 * This file demonstrates all proxy configurations and crawl types.
 *
 * Proxy Modes:
 * - none: Direct connection (1x credits)
 * - datacenter: Fast datacenter proxies (2x credits) - NST/Scrapeless
 * - residential: Premium residential IPs (5x credits) - Massive/NST/Scrapeless
 * - auto: Smart selection based on target URL
 *
 * Crawl Types:
 * - Sync single: POST /v1/crawl
 * - Sync batch: POST /v1/crawl/batch
 * - Async job: POST /v1/crawl/async
 * - Deep crawl: POST /v1/crawl/deep
 *
 * Provider Pool:
 * - Massive: Residential ONLY (static credentials)
 * - NST: Both datacenter AND residential
 * - Scrapeless: Both datacenter AND residential
 *
 * Usage:
 *     npx ts-node proxy_examples.ts
 */

const API_BASE = 'https://api.crawl4ai.com';
const API_KEY = 'sk_live_YOUR_API_KEY_HERE';

const headers = {
  'Content-Type': 'application/json',
  'X-API-Key': API_KEY,
};

// =============================================================================
// SYNC SINGLE CRAWL - All Proxy Modes
// =============================================================================

/**
 * Crawl without proxy - direct connection.
 * Cost: 1x credits (100 credits per URL)
 */
async function crawlNoProxy(url: string): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      proxy: { mode: 'none' },
      bypass_cache: true,
    }),
  });
  return response.json();
}

/**
 * Crawl with datacenter proxy - fast and cheap.
 * Providers: NST or Scrapeless (weighted random: 60% NST, 40% Scrapeless)
 * Cost: 2x credits (200 credits per URL)
 *
 * Good for: General scraping, non-protected sites, high volume
 */
async function crawlDatacenterProxy(url: string, country?: string): Promise<any> {
  const proxyConfig: any = { mode: 'datacenter' };
  if (country) {
    proxyConfig.country = country; // ISO code: "US", "GB", "DE", etc.
  }

  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      proxy: proxyConfig,
      bypass_cache: true,
    }),
  });
  return response.json();
}

/**
 * Crawl with residential proxy - premium IPs for protected sites.
 * Providers: Massive, NST, or Scrapeless (weighted random)
 * Cost: 5x credits (500 credits per URL)
 *
 * Good for: Amazon, LinkedIn, Google, social media, anti-bot sites
 */
async function crawlResidentialProxy(url: string, country?: string): Promise<any> {
  const proxyConfig: any = { mode: 'residential' };
  if (country) {
    proxyConfig.country = country;
  }

  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      proxy: proxyConfig,
      bypass_cache: true,
    }),
  });
  return response.json();
}

/**
 * Crawl with auto proxy mode - smart selection based on URL.
 *
 * Heuristics:
 * - Hard targets (amazon, linkedin, google, etc.) -> residential
 * - Easy targets (example.com, httpbin, github) -> no proxy
 * - Unknown domains -> datacenter
 *
 * Cost: Varies based on selection (1x, 2x, or 5x)
 */
async function crawlAutoProxy(url: string): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      proxy: { mode: 'auto' },
      bypass_cache: true,
    }),
  });
  return response.json();
}

/**
 * Force a specific proxy provider (for testing/debugging).
 *
 * @param provider - "massive" (residential only), "nst", or "scrapeless"
 * @param mode - "datacenter" or "residential"
 *
 * Note: Massive only supports residential mode.
 */
async function crawlSpecificProvider(url: string, provider: string, mode: string): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      proxy: { mode, provider },
      bypass_cache: true,
    }),
  });
  return response.json();
}

// =============================================================================
// SYNC BATCH CRAWL - Multiple URLs with Proxy
// =============================================================================

/**
 * Crawl multiple URLs in a single request with proxy.
 * All URLs use the same proxy configuration.
 *
 * Cost: (number of URLs) * mode_multiplier credits
 * Example: 5 URLs with datacenter = 5 * 200 = 1000 credits
 */
async function batchCrawlWithProxy(urls: string[], mode = 'datacenter'): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl/batch`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      urls,
      proxy: { mode },
      bypass_cache: true,
    }),
  });
  return response.json();
}

// =============================================================================
// ASYNC JOB CRAWL - Background Processing with Proxy
// =============================================================================

/**
 * Submit async crawl job with proxy.
 * Returns job ID for polling.
 *
 * Good for: Large batches, long-running crawls, non-blocking operations
 */
async function asyncCrawlWithProxy(urls: string[], mode = 'datacenter'): Promise<string> {
  const response = await fetch(`${API_BASE}/v1/crawl/async`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      urls,
      proxy: { mode },
      bypass_cache: true,
    }),
  });
  const data = await response.json();
  return data.job_id;
}

/**
 * Check async job status.
 */
async function getJobStatus(jobId: string): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl/jobs/${jobId}`, {
    method: 'GET',
    headers,
  });
  return response.json();
}

// =============================================================================
// DEEP CRAWL - Multi-page Crawling with Sticky Sessions
// =============================================================================

/**
 * Deep crawl a site with proxy.
 * Discovers and crawls linked pages.
 */
async function deepCrawlWithProxy(
  url: string,
  mode = 'datacenter',
  maxDepth = 2,
  maxUrls = 10
): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl/deep`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      strategy: 'bfs',
      max_depth: maxDepth,
      max_urls: maxUrls,
      proxy: { mode },
    }),
  });
  return response.json();
}

/**
 * Deep crawl with sticky session - same proxy IP for all URLs.
 *
 * IMPORTANT: Sticky sessions ensure all pages in the crawl use the
 * same proxy IP address. This is crucial for:
 * - Session-based authentication
 * - Rate limiting that tracks by IP
 * - Sites that detect IP changes
 */
async function deepCrawlStickySession(
  url: string,
  mode = 'datacenter',
  maxDepth = 2,
  maxUrls = 10
): Promise<any> {
  const response = await fetch(`${API_BASE}/v1/crawl/deep`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      strategy: 'bfs',
      max_depth: maxDepth,
      max_urls: maxUrls,
      proxy: {
        mode,
        sticky_session: true, // Same IP for all URLs
      },
    }),
  });
  return response.json();
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================

async function main(): Promise<void> {
  try {
    // Example 1: Simple crawl without proxy
    console.log('=== No Proxy ===');
    let result = await crawlNoProxy('https://httpbin.org/ip');
    console.log(`Success: ${result.success}`);
    console.log(`Proxy: ${result.proxy_mode}`);

    // Example 2: Datacenter proxy
    console.log('\n=== Datacenter Proxy ===');
    result = await crawlDatacenterProxy('https://httpbin.org/ip');
    console.log(`Success: ${result.success}`);
    console.log(`Provider: ${result.proxy_used}`);
    console.log(`Mode: ${result.proxy_mode}`);

    // Example 3: Residential proxy for protected site
    console.log('\n=== Residential Proxy ===');
    result = await crawlResidentialProxy('https://httpbin.org/ip', 'US');
    console.log(`Success: ${result.success}`);
    console.log(`Provider: ${result.proxy_used}`);

    // Example 4: Auto mode
    console.log('\n=== Auto Mode (Amazon) ===');
    result = await crawlAutoProxy('https://amazon.com');
    console.log(`Auto selected: ${result.proxy_mode}`);
    console.log(`Provider: ${result.proxy_used}`);

    // Example 5: Deep crawl with sticky session
    console.log('\n=== Deep Crawl with Sticky Session ===');
    result = await deepCrawlStickySession('https://example.com', 'datacenter', 1, 3);
    console.log(`Job ID: ${result.job_id}`);
    console.log(`Status: ${result.status}`);
  } catch (error) {
    console.error('Error:', error);
  }
}

// Export functions for use as module
export {
  crawlNoProxy,
  crawlDatacenterProxy,
  crawlResidentialProxy,
  crawlAutoProxy,
  crawlSpecificProvider,
  batchCrawlWithProxy,
  asyncCrawlWithProxy,
  getJobStatus,
  deepCrawlWithProxy,
  deepCrawlStickySession,
};

// Run examples if executed directly
main().catch(console.error);
