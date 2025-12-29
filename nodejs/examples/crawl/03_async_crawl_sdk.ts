#!/usr/bin/env npx ts-node
/**
 * Async Crawl with Wait - SDK Example
 *
 * This script demonstrates async crawling with automatic polling (wait=True).
 * The SDK automatically polls until the job completes and returns the results.
 *
 * Usage:
 *     npx ts-node 03_async_crawl_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler, CrawlResult } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function main(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // URLs to crawl (can be more than 10 for async)
    const urls = [
      'https://example.com',
      'https://httpbin.org/html',
      'https://httpbin.org/json',
      'https://httpbin.org/robots.txt',
    ];

    console.log(`Creating async job for ${urls.length} URLs...`);

    // runMany with wait=true handles polling automatically
    const results = (await crawler.runMany(urls, {
      strategy: 'http', // Options: "browser" (JS support) or "http" (faster, no JS)
      wait: true, // Wait for completion (SDK polls automatically)
    })) as CrawlResult[];

    // Display results
    console.log('\n=== JOB COMPLETE ===');
    console.log(`Total: ${results.length}`);
    const succeeded = results.filter((r) => r.success).length;
    console.log(`Succeeded: ${succeeded}`);

    // Show sample results
    console.log('\nSample Results (first 3):');
    results.slice(0, 3).forEach((result, i) => {
      console.log(`[${i + 1}] ${result.url}`);
      console.log(`    Status: ${result.statusCode}`);
      if (result.success && result.markdown) {
        const preview = result.markdown.rawMarkdown?.slice(0, 80);
        console.log(`    Preview: ${preview}...`);
      }
    });
  } finally {
    await crawler.close();
  }
}

main().catch(console.error);
