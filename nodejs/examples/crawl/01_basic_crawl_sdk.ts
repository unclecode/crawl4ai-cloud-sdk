#!/usr/bin/env npx ts-node
/**
 * Basic Single URL Crawl - SDK Example
 *
 * This script demonstrates the simplest way to crawl a single URL using the SDK.
 * The run() method is async and returns when the crawl completes.
 *
 * Usage:
 *     npx ts-node 01_basic_crawl_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function main(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log('Crawling https://example.com...');

    // Crawl with browser strategy (full JS support)
    const result = await crawler.run('https://example.com', {
      strategy: 'browser', // Options: "browser" (JS support) or "http" (faster, no JS)
    });

    // Display results
    console.log('\n=== CRAWL COMPLETE ===');
    console.log(`URL: ${result.url}`);
    console.log(`Success: ${result.success}`);
    console.log(`Status: ${result.statusCode}`);
    console.log('\nMarkdown preview (first 200 chars):');
    console.log(result.markdown?.rawMarkdown?.slice(0, 200) + '...');
    console.log(`\nHTML length: ${result.html?.length || 0} characters`);
  } finally {
    await crawler.close();
  }
}

main().catch(console.error);
