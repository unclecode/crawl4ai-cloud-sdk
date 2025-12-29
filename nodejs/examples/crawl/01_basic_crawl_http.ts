#!/usr/bin/env npx ts-node
/**
 * Basic Single URL Crawl - HTTP Example
 *
 * This script demonstrates crawling a single URL using direct HTTP requests with fetch.
 * The endpoint is synchronous and blocks until the crawl completes.
 *
 * Usage:
 *     npx ts-node 01_basic_crawl_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key
const API_URL = 'https://api.crawl4ai.com';

async function main(): Promise<void> {
  console.log('Crawling https://example.com...');

  try {
    // Make the crawl request
    const response = await fetch(`${API_URL}/v1/crawl`, {
      method: 'POST',
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        url: 'https://example.com',
        strategy: 'browser', // Options: "browser" (JS support) or "http" (faster, no JS)
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP Error ${response.status}: ${await response.text()}`);
    }

    const data = await response.json();

    // Display results
    console.log('\n=== CRAWL COMPLETE ===');
    console.log(`URL: ${data.url}`);
    console.log(`Title: ${data.metadata?.title || 'N/A'}`);
    console.log(`Status: ${data.status_code}`);
    console.log('\nMarkdown preview (first 200 chars):');
    console.log(data.markdown?.raw_markdown?.slice(0, 200) + '...');
    console.log(`\nHTML length: ${data.html?.length || 0} characters`);
  } catch (error) {
    console.error('Error:', error);
  }
}

main();
