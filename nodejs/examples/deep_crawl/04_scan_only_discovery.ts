#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - Scan-Only Mode (URL Discovery)
 *
 * Scan-only mode discovers URLs without crawling them. This is useful for:
 * - Previewing what URLs will be crawled before committing
 * - Building a URL list for later processing
 * - Analyzing site structure
 * - Fast HTML caching for later extraction
 *
 * The scan phase:
 * 1. Discovers URLs (sitemap, links, Common Crawl)
 * 2. Caches HTML content (30 min TTL)
 * 3. Returns URL list with metadata (depth, links_found, html_size)
 *
 * Usage:
 *     npx ts-node 04_scan_only_discovery.ts
 */

import { AsyncWebCrawler, DeepCrawlResult, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function basicScanOnly(): Promise<void> {
  console.log('=== Scan-Only Mode (Discovery) ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // scanOnly=true returns discovered URLs without processing
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 20,
      scanOnly: true, // Just discover, don't crawl
      wait: true,
    })) as DeepCrawlResult;

    // Result is a DeepCrawlResult (not a Job)
    console.log(`Status: ${result.status}`);
    console.log(`URLs discovered: ${result.discoveredCount}`);
    console.log(`Cache expires at: ${result.cacheExpiresAt}`);

    // Get list of discovered URLs
    if (result.discoveredUrls) {
      console.log('\nDiscovered URLs:');
      result.discoveredUrls.slice(0, 10).forEach((url) => {
        console.log(`  - ${url}`);
      });
    }
  } finally {
    await crawler.close();
  }
}

async function scanWithUrlDetails(): Promise<void> {
  console.log('\n=== Scan with URL Details ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 15,
      scanOnly: true,
      wait: true,
    })) as DeepCrawlResult;

    console.log(`Discovered ${result.discoveredCount} URLs\n`);

    // urls contains ScanUrlInfo objects with metadata
    if (result.urls) {
      console.log('URL Details:');
      console.log('-'.repeat(70));
      result.urls.slice(0, 10).forEach((urlInfo) => {
        console.log(`URL: ${urlInfo.url}`);
        console.log(`  Depth: ${urlInfo.depth}`);
        console.log(`  Links found: ${urlInfo.linksFound}`);
        console.log(`  HTML size: ${urlInfo.htmlSize?.toLocaleString()} bytes`);
        if (urlInfo.score !== undefined) {
          console.log(`  Score: ${urlInfo.score.toFixed(2)}`);
        }
        console.log();
      });
    }
  } finally {
    await crawler.close();
  }
}

async function scanMapStrategy(): Promise<void> {
  console.log('\n=== Scan-Only with Map Strategy ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map',
      maxUrls: 50,
      scanOnly: true,
      wait: true,
    })) as DeepCrawlResult;

    console.log(`Sitemap URLs found: ${result.discoveredCount}`);

    // You can filter/review URLs before deciding to crawl
    if (result.discoveredUrls) {
      const apiUrls = result.discoveredUrls.filter((u) => u.includes('/api/'));
      const guideUrls = result.discoveredUrls.filter((u) => u.includes('/guide/'));

      console.log(`\nAPI pages: ${apiUrls.length}`);
      console.log(`Guide pages: ${guideUrls.length}`);

      // Save the scan job ID for later extraction
      console.log(`\nScan Job ID: ${result.jobId}`);
      console.log('Use this ID with sourceJob to extract later!');
    }
  } finally {
    await crawler.close();
  }
}

async function scanThenDecide(): Promise<void> {
  console.log('\n=== Scan Then Decide Workflow ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Step 1: Discover URLs
    console.log('Step 1: Discovering URLs...');
    const scanResult = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 30,
      scanOnly: true,
      wait: true,
    })) as DeepCrawlResult;

    console.log(`Found ${scanResult.discoveredCount} URLs`);
    console.log(`Scan Job ID: ${scanResult.jobId}`);

    // Step 2: Review and decide
    const totalSize = scanResult.urls?.reduce((sum, u) => sum + (u.htmlSize || 0), 0) || 0;
    console.log(`Total HTML size: ${totalSize.toLocaleString()} bytes`);

    // You could prompt user here or apply business logic
    const shouldCrawl = scanResult.discoveredCount <= 20;

    if (shouldCrawl) {
      // Step 3: Extract from cached HTML
      console.log('\nStep 2: Extracting from cache...');
      const job = (await crawler.deepCrawl(undefined, {
        sourceJob: scanResult.jobId, // Use cached HTML
        wait: true,
      })) as CrawlJob;
      console.log(`Extracted ${job.progress.completed} pages`);
    } else {
      console.log('\nToo many URLs, skipping extraction.');
      console.log('You can still use sourceJob within 30 minutes.');
    }
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await basicScanOnly();
  // Uncomment to run other examples:
  // await scanWithUrlDetails();
  // await scanMapStrategy();
  // await scanThenDecide();
}

main().catch(console.error);
