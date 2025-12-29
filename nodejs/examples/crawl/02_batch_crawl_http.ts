#!/usr/bin/env npx ts-node
/**
 * Batch Crawl - HTTP Example
 *
 * This script demonstrates crawling multiple URLs (up to 10) using direct HTTP requests.
 * The batch endpoint processes URLs sequentially and returns all results.
 *
 * Usage:
 *     npx ts-node 02_batch_crawl_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key
const API_URL = 'https://api.crawl4ai.com';

async function main(): Promise<void> {
  // URLs to crawl (max 10)
  const urls = [
    'https://example.com',
    'https://httpbin.org/html',
    'https://httpbin.org/json',
  ];

  console.log(`Crawling ${urls.length} URLs in batch...`);

  try {
    // Make the batch crawl request
    const response = await fetch(`${API_URL}/v1/crawl/batch`, {
      method: 'POST',
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        urls,
        strategy: 'http', // Options: "browser" (JS support) or "http" (faster, no JS)
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP Error ${response.status}: ${await response.text()}`);
    }

    const data = await response.json();

    // Display results
    console.log('\n=== BATCH CRAWL COMPLETE ===');
    console.log(`Total URLs: ${data.results.length}`);
    console.log(`Succeeded: ${data.succeeded}`);
    console.log(`Failed: ${data.failed}`);

    // Show individual results
    data.results.forEach((result: any, i: number) => {
      console.log(`\n[${i + 1}] ${result.url}`);
      console.log(`    Status: ${result.status_code}`);
      if (result.markdown) {
        const preview = result.markdown.raw_markdown?.slice(0, 100);
        console.log(`    Preview: ${preview}...`);
      } else {
        console.log(`    Error: ${result.error_message || 'Unknown error'}`);
      }
    });
  } catch (error) {
    console.error('Error:', error);
  }
}

main();
