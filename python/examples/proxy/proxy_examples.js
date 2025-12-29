/**
 * Crawl4AI Cloud - Proxy Usage Examples (JavaScript/Node.js)
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
 */

const API_BASE = "https://api.crawl4ai.com";
const API_KEY = "sk_live_YOUR_API_KEY_HERE";

const headers = {
  "Content-Type": "application/json",
  "X-API-Key": API_KEY,
};

// =============================================================================
// SYNC SINGLE CRAWL - All Proxy Modes
// =============================================================================

/**
 * Crawl without proxy - direct connection.
 * Cost: 1x credits (100 credits per URL)
 */
async function crawlNoProxy(url) {
  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      proxy: { mode: "none" },
      bypass_cache: true,
    }),
  });
  return response.json();
  // Response: {
  //   success: true,
  //   url: "https://example.com",
  //   proxy_used: null,
  //   proxy_mode: "none",
  //   html: "...",
  //   markdown: {...},
  //   cleaned_html: "...",
  //   metadata: {...}
  // }
}

/**
 * Crawl with datacenter proxy - fast and cheap.
 * Providers: NST or Scrapeless (weighted random: 60% NST, 40% Scrapeless)
 * Cost: 2x credits (200 credits per URL)
 *
 * Good for: General scraping, non-protected sites, high volume
 */
async function crawlDatacenterProxy(url, country = null) {
  const proxyConfig = { mode: "datacenter" };
  if (country) {
    proxyConfig.country = country; // ISO code: "US", "GB", "DE", etc.
  }

  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      proxy: proxyConfig,
      bypass_cache: true,
    }),
  });
  return response.json();
  // Response: {
  //   success: true,
  //   url: "https://example.com",
  //   proxy_used: "nst",  // or "scrapeless"
  //   proxy_mode: "datacenter",
  //   html: "...",
  //   ...
  // }
}

/**
 * Crawl with residential proxy - premium IPs for protected sites.
 * Providers: Massive, NST, or Scrapeless (weighted random)
 * Cost: 5x credits (500 credits per URL)
 *
 * Good for: Amazon, LinkedIn, Google, social media, anti-bot sites
 */
async function crawlResidentialProxy(url, country = null) {
  const proxyConfig = { mode: "residential" };
  if (country) {
    proxyConfig.country = country;
  }

  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      proxy: proxyConfig,
      bypass_cache: true,
    }),
  });
  return response.json();
  // Response: {
  //   success: true,
  //   url: "https://amazon.com",
  //   proxy_used: "massive",  // or "nst" or "scrapeless"
  //   proxy_mode: "residential",
  //   html: "...",
  //   ...
  // }
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
async function crawlAutoProxy(url) {
  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      proxy: { mode: "auto" },
      bypass_cache: true,
    }),
  });
  return response.json();
  // Response: {
  //   success: true,
  //   url: "https://amazon.com",
  //   proxy_used: "nst",
  //   proxy_mode: "residential",  // auto selected residential for amazon
  //   html: "...",
  //   ...
  // }
}

/**
 * Force a specific proxy provider (for testing/debugging).
 *
 * @param {string} provider - "massive" (residential only), "nst", or "scrapeless"
 * @param {string} mode - "datacenter" or "residential"
 *
 * Note: Massive only supports residential mode.
 */
async function crawlSpecificProvider(url, provider, mode) {
  const response = await fetch(`${API_BASE}/v1/crawl`, {
    method: "POST",
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
async function batchCrawlWithProxy(urls, mode = "datacenter") {
  const response = await fetch(`${API_BASE}/v1/crawl/batch`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      urls,
      proxy: { mode },
      bypass_cache: true,
    }),
  });
  return response.json();
  // Response: {
  //   results: [
  //     {success: true, url: "...", proxy_used: "nst", ...},
  //     {success: true, url: "...", proxy_used: "scrapeless", ...},
  //     ...
  //   ]
  // }
}

/**
 * Batch crawl with residential proxy from specific country.
 * Useful for geo-restricted content.
 */
async function batchCrawlResidentialGeo(urls, country = "US") {
  const response = await fetch(`${API_BASE}/v1/crawl/batch`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      urls,
      proxy: { mode: "residential", country },
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
async function asyncCrawlWithProxy(urls, mode = "datacenter") {
  const response = await fetch(`${API_BASE}/v1/crawl/async`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      urls,
      proxy: { mode },
      bypass_cache: true,
    }),
  });
  const data = await response.json();
  return data.job_id;
  // Response: {
  //   job_id: "job_abc123...",
  //   status: "pending",
  //   urls_count: 5
  // }
}

/**
 * Check async job status.
 */
async function getJobStatus(jobId) {
  const response = await fetch(`${API_BASE}/v1/crawl/jobs/${jobId}`, {
    method: "GET",
    headers,
  });
  return response.json();
  // Response when pending: {
  //   job_id: "job_abc123...",
  //   status: "processing",
  //   progress: {completed: 2, total: 5}
  // }
  // Response when complete: {
  //   job_id: "job_abc123...",
  //   status: "completed",
  //   results: [...]
  // }
}

/**
 * Submit async job and wait for completion.
 */
async function asyncCrawlWaitComplete(urls, mode = "datacenter", timeout = 300000) {
  const jobId = await asyncCrawlWithProxy(urls, mode);
  console.log(`Job submitted: ${jobId}`);

  const start = Date.now();
  while (Date.now() - start < timeout) {
    const status = await getJobStatus(jobId);
    console.log(`Status: ${status.status}`);

    if (status.status === "completed") {
      return status;
    }
    if (status.status === "failed") {
      throw new Error(`Job failed: ${status.error}`);
    }

    await new Promise((resolve) => setTimeout(resolve, 2000));
  }

  throw new Error(`Job ${jobId} did not complete within ${timeout}ms`);
}

// =============================================================================
// DEEP CRAWL - Multi-page Crawling with Sticky Sessions
// =============================================================================

/**
 * Deep crawl a site with proxy.
 * Discovers and crawls linked pages.
 *
 * Strategies:
 * - bfs: Breadth-first (same level pages first)
 * - dfs: Depth-first (follow links deep)
 * - bestfirst: Prioritize by relevance
 */
async function deepCrawlWithProxy(url, mode = "datacenter", maxDepth = 2, maxUrls = 10) {
  const response = await fetch(`${API_BASE}/v1/crawl/deep`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      strategy: "bfs",
      max_depth: maxDepth,
      max_urls: maxUrls,
      proxy: { mode },
    }),
  });
  return response.json();
  // Response: {
  //   job_id: "deep_abc123...",
  //   status: "processing",
  //   discovered_urls: ["...", "..."]
  // }
}

/**
 * Deep crawl with sticky session - same proxy IP for all URLs.
 *
 * IMPORTANT: Sticky sessions ensure all pages in the crawl use the
 * same proxy IP address. This is crucial for:
 * - Session-based authentication
 * - Rate limiting that tracks by IP
 * - Sites that detect IP changes
 *
 * The proxy IP is cached for the duration of the job and released
 * when the job completes.
 */
async function deepCrawlStickySession(url, mode = "datacenter", maxDepth = 2, maxUrls = 10) {
  const response = await fetch(`${API_BASE}/v1/crawl/deep`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      strategy: "bfs",
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

/**
 * Deep crawl protected site with residential proxy and sticky session.
 * Best for: E-commerce sites, social media, heavily protected targets.
 */
async function deepCrawlResidentialSticky(url, country = "US") {
  const response = await fetch(`${API_BASE}/v1/crawl/deep`, {
    method: "POST",
    headers,
    body: JSON.stringify({
      url,
      strategy: "bfs",
      max_depth: 2,
      max_urls: 20,
      proxy: {
        mode: "residential",
        country,
        sticky_session: true,
      },
    }),
  });
  return response.json();
}

/**
 * Check deep crawl job status.
 */
async function getDeepCrawlStatus(jobId) {
  const response = await fetch(`${API_BASE}/v1/crawl/deep/${jobId}`, {
    method: "GET",
    headers,
  });
  return response.json();
}

// =============================================================================
// RESPONSE TYPE EXAMPLES
// =============================================================================

/*
SYNC CRAWL RESPONSE:
{
    success: true,
    url: "https://example.com",
    proxy_used: "nst",           // Provider used: "massive", "nst", "scrapeless", or null
    proxy_mode: "datacenter",    // Mode used: "none", "datacenter", "residential"
    html: "<!DOCTYPE html>...",
    cleaned_html: "<main>...</main>",
    markdown: {
        raw_markdown: "# Title\n\nContent...",
        markdown_with_citations: "# Title\n\nContent [1]...",
        references_markdown: "[1]: https://...",
        fit_markdown: "Title\nContent..."
    },
    metadata: {
        title: "Page Title",
        description: "Meta description",
        keywords: ["keyword1", "keyword2"],
        author: "Author Name"
    },
    links: {
        internal: ["https://example.com/page1", "..."],
        external: ["https://other.com", "..."]
    },
    media: {
        images: [{src: "...", alt: "..."}],
        videos: [],
        audios: []
    },
    screenshot: "base64...",     // If screenshot: true
    pdf: "base64..."             // If pdf: true
}

BATCH CRAWL RESPONSE:
{
    results: [
        {
            success: true,
            url: "https://example1.com",
            proxy_used: "nst",
            proxy_mode: "datacenter",
            html: "...",
            ...
        },
        {
            success: true,
            url: "https://example2.com",
            proxy_used: "scrapeless",
            proxy_mode: "datacenter",
            html: "...",
            ...
        }
    ]
}

ASYNC JOB RESPONSE (Submit):
{
    job_id: "job_abc123def456...",
    status: "pending",
    urls_count: 5
}

ASYNC JOB RESPONSE (Status - Processing):
{
    job_id: "job_abc123def456...",
    status: "processing",
    progress: {
        completed: 2,
        total: 5,
        failed: 0
    }
}

ASYNC JOB RESPONSE (Status - Complete):
{
    job_id: "job_abc123def456...",
    status: "completed",
    results: [
        {success: true, url: "...", proxy_used: "nst", ...},
        ...
    ]
}

DEEP CRAWL RESPONSE (Submit):
{
    job_id: "deep_abc123...",
    status: "processing",
    seed_url: "https://example.com",
    discovered_urls: [
        "https://example.com/page1",
        "https://example.com/page2"
    ]
}

DEEP CRAWL RESPONSE (Complete):
{
    job_id: "deep_abc123...",
    status: "completed",
    seed_url: "https://example.com",
    crawled_count: 10,
    results: [
        {url: "...", success: true, proxy_used: "nst", ...},
        ...
    ]
}
*/

// =============================================================================
// USAGE EXAMPLES
// =============================================================================

async function main() {
  try {
    // Example 1: Simple crawl without proxy
    console.log("=== No Proxy ===");
    let result = await crawlNoProxy("https://httpbin.org/ip");
    console.log(`Success: ${result.success}`);
    console.log(`Proxy: ${result.proxy_mode}`);

    // Example 2: Datacenter proxy
    console.log("\n=== Datacenter Proxy ===");
    result = await crawlDatacenterProxy("https://httpbin.org/ip");
    console.log(`Success: ${result.success}`);
    console.log(`Provider: ${result.proxy_used}`);
    console.log(`Mode: ${result.proxy_mode}`);

    // Example 3: Residential proxy for protected site
    console.log("\n=== Residential Proxy ===");
    result = await crawlResidentialProxy("https://httpbin.org/ip", "US");
    console.log(`Success: ${result.success}`);
    console.log(`Provider: ${result.proxy_used}`);

    // Example 4: Auto mode
    console.log("\n=== Auto Mode (Amazon) ===");
    result = await crawlAutoProxy("https://amazon.com");
    console.log(`Auto selected: ${result.proxy_mode}`);
    console.log(`Provider: ${result.proxy_used}`);

    // Example 5: Deep crawl with sticky session
    console.log("\n=== Deep Crawl with Sticky Session ===");
    result = await deepCrawlStickySession("https://example.com", "datacenter", 1, 3);
    console.log(`Job ID: ${result.job_id}`);
    console.log(`Status: ${result.status}`);
  } catch (error) {
    console.error("Error:", error.message);
  }
}

// Export functions for use as module
module.exports = {
  crawlNoProxy,
  crawlDatacenterProxy,
  crawlResidentialProxy,
  crawlAutoProxy,
  crawlSpecificProvider,
  batchCrawlWithProxy,
  batchCrawlResidentialGeo,
  asyncCrawlWithProxy,
  getJobStatus,
  asyncCrawlWaitComplete,
  deepCrawlWithProxy,
  deepCrawlStickySession,
  deepCrawlResidentialSticky,
  getDeepCrawlStatus,
};

// Run examples if executed directly
if (require.main === module) {
  main();
}
