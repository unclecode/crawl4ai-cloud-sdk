#!/usr/bin/env npx ts-node
/**
 * Batch Crawl - SDK Example
 *
 * This script demonstrates crawling multiple URLs at once.
 * The runMany() method automatically selects batch (<=10 URLs) or async (>10 URLs).
 *
 * Usage:
 *     npx ts-node 02_batch_crawl_sdk.ts
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
    // URLs to crawl
    const urls = [
      'https://example.com',
      'https://httpbin.org/html',
      'https://httpbin.org/json',
    ];

    console.log(`Crawling ${urls.length} URLs...`);

    // Crawl multiple URLs (auto-selects batch or async based on count)
    const results = (await crawler.runMany(urls, {
      strategy: 'http', // Options: "browser" (JS support) or "http" (faster, no JS)
      wait: true, // Wait for all to complete
    })) as CrawlResult[];

    // Display results
    console.log('\n=== BATCH CRAWL COMPLETE ===');
    console.log(`Total URLs: ${results.length}`);
    const succeeded = results.filter((r) => r.success).length;
    console.log(`Succeeded: ${succeeded}`);
    console.log(`Failed: ${results.length - succeeded}`);

    // Show individual results
    results.forEach((result, i) => {
      console.log(`\n[${i + 1}] ${result.url}`);
      console.log(`    Status: ${result.statusCode}`);
      if (result.success && result.markdown) {
        const preview = result.markdown.rawMarkdown?.slice(0, 100);
        console.log(`    Preview: ${preview}...`);
      } else {
        console.log(`    Error: ${result.errorMessage}`);
      }
    });
  } finally {
    await crawler.close();
  }
}

main().catch(console.error);
