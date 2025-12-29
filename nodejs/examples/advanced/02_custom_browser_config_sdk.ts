#!/usr/bin/env npx ts-node
/**
 * Custom Browser Configuration - SDK Example
 *
 * This script demonstrates how to customize browser settings like viewport size,
 * proxies, and custom headers using the Crawl4AI SDK.
 *
 * Usage:
 *     npx ts-node 02_custom_browser_config_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function crawlWithCustomViewport(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Crawling ${url} with custom viewport...`);

    const result = await crawler.run(url, {
      strategy: 'browser',
      browserConfig: {
        viewport: { width: 1920, height: 1080 },
      },
    });

    console.log(`Success! HTML size: ${result.html?.length || 0} bytes`);
  } finally {
    await crawler.close();
  }
}

async function crawlWithProxy(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Crawling ${url} with datacenter proxy...`);

    const result = await crawler.run(url, {
      strategy: 'browser',
      proxy: { mode: 'datacenter' }, // or "residential"
    });

    console.log(`Success! HTML size: ${result.html?.length || 0} bytes`);
  } finally {
    await crawler.close();
  }
}

async function crawlWithCustomHeaders(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Crawling ${url} with custom headers...`);

    const result = await crawler.run(url, {
      strategy: 'browser',
      browserConfig: {
        headers: {
          'User-Agent': 'CustomBot/1.0 (Research purposes)',
          'Accept-Language': 'en-US,en;q=0.9',
        },
      },
    });

    console.log(`Success! HTML size: ${result.html?.length || 0} bytes`);
  } finally {
    await crawler.close();
  }
}

async function crawlWithFullConfig(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Crawling ${url} with full custom config...`);

    const result = await crawler.run(url, {
      strategy: 'browser',
      browserConfig: {
        viewport: { width: 1920, height: 1080 },
        headers: {
          'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        },
      },
      config: {
        wait_until: 'networkidle', // Wait for network to be idle
        page_timeout: 30000, // 30 second timeout
      },
    });

    console.log(`Success! HTML size: ${result.html?.length || 0} bytes`);
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  // Example 1: Custom viewport
  await crawlWithCustomViewport('https://www.example.com');

  // Example 2: Custom headers
  await crawlWithCustomHeaders('https://www.example.com');

  // Example 3: Full configuration
  await crawlWithFullConfig('https://www.example.com');

  // Example 4: With managed proxy (uncomment to use)
  // await crawlWithProxy('https://www.example.com');
}

main().catch(console.error);
