#!/usr/bin/env npx ts-node
/**
 * Custom Browser Configuration - HTTP Example
 *
 * This script demonstrates how to customize browser settings using raw HTTP requests.
 * Covers viewport, headless mode, proxies, and custom headers.
 *
 * Usage:
 *     npx ts-node 02_custom_browser_config_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_URL = 'https://api.crawl4ai.com';
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

const headers = {
  'X-API-Key': API_KEY,
  'Content-Type': 'application/json',
};

async function crawlWithCustomViewport(url: string): Promise<any> {
  console.log(`Crawling ${url} with custom viewport...`);

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      strategy: 'browser',
      browser_config: {
        viewport: { width: 1920, height: 1080 },
      },
    }),
  });

  if (response.ok) {
    const data = await response.json();
    console.log(`Success! Title: ${data.metadata?.title || 'N/A'}`);
    return data;
  } else {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return null;
  }
}

async function crawlWithProxy(url: string, proxyUrl: string): Promise<any> {
  console.log(`Crawling ${url} through proxy...`);

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      strategy: 'browser',
      browser_config: {
        proxy: proxyUrl,
      },
    }),
  });

  if (response.ok) {
    const data = await response.json();
    console.log(`Success! Title: ${data.metadata?.title || 'N/A'}`);
    return data;
  } else {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return null;
  }
}

async function crawlWithCustomHeaders(url: string): Promise<any> {
  console.log(`Crawling ${url} with custom headers...`);

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      strategy: 'browser',
      browser_config: {
        headers: {
          'User-Agent': 'CustomBot/1.0 (Research purposes)',
          'Accept-Language': 'en-US,en;q=0.9',
        },
      },
    }),
  });

  if (response.ok) {
    const data = await response.json();
    console.log(`Success! Title: ${data.metadata?.title || 'N/A'}`);
    return data;
  } else {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return null;
  }
}

async function crawlWithFullConfig(url: string): Promise<any> {
  console.log(`Crawling ${url} with full custom config...`);

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url,
      strategy: 'browser',
      browser_config: {
        headless: true,
        viewport: { width: 1920, height: 1080 },
        headers: {
          'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
        },
      },
      crawler_config: {
        wait_for: 'networkidle',
        page_timeout: 30000,
      },
    }),
  });

  if (response.ok) {
    const data = await response.json();
    console.log(`Success! Title: ${data.metadata?.title || 'N/A'}`);
    console.log(`HTML size: ${data.html?.length || 0} bytes`);
    return data;
  } else {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return null;
  }
}

async function main(): Promise<void> {
  // Example 1: Custom viewport
  await crawlWithCustomViewport('https://www.example.com');

  // Example 2: Custom headers
  await crawlWithCustomHeaders('https://www.example.com');

  // Example 3: Full configuration
  await crawlWithFullConfig('https://www.example.com');

  // Example 4: With proxy (uncomment to use)
  // await crawlWithProxy('https://www.example.com', 'http://proxy.example.com:8080');
}

main().catch(console.error);
