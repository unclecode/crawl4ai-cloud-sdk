#!/usr/bin/env npx ts-node
/**
 * Async Crawl with Polling - HTTP Example
 *
 * This script demonstrates async crawling with manual polling loop.
 * Create a job, then poll the status endpoint until completion.
 *
 * Usage:
 *     npx ts-node 03_async_crawl_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key
const API_URL = 'https://api.crawl4ai.com';

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function main(): Promise<void> {
  // URLs to crawl (can be more than 10 for async)
  const urls = [
    'https://example.com',
    'https://httpbin.org/html',
    'https://httpbin.org/json',
    'https://httpbin.org/robots.txt',
  ];

  console.log(`Creating async job for ${urls.length} URLs...`);

  try {
    // Step 1: Create the async job
    const createResponse = await fetch(`${API_URL}/v1/crawl/async`, {
      method: 'POST',
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        urls,
        strategy: 'http', // Options: "browser" (JS support) or "http" (faster, no JS)
        priority: 5, // Priority 1-10 (default: 5)
      }),
    });

    if (!createResponse.ok) {
      throw new Error(`HTTP Error ${createResponse.status}: ${await createResponse.text()}`);
    }

    const data = await createResponse.json();
    const jobId = data.job_id;

    console.log(`Job created: ${jobId}`);
    console.log(`Status: ${data.status}`);

    // Step 2: Poll for completion
    console.log('\nPolling for completion...');
    const maxAttempts = 60;
    const pollInterval = 2000;

    for (let attempt = 0; attempt < maxAttempts; attempt++) {
      await sleep(pollInterval);

      const statusResponse = await fetch(`${API_URL}/v1/crawl/jobs/${jobId}`, {
        headers: { 'X-API-Key': API_KEY },
      });

      if (!statusResponse.ok) {
        throw new Error(`HTTP Error ${statusResponse.status}: ${await statusResponse.text()}`);
      }

      const statusData = await statusResponse.json();

      console.log(
        `  [${attempt + 1}] Status: ${statusData.status} | ` +
          `Progress: ${statusData.progress.completed}/${statusData.progress.total}`
      );

      if (['completed', 'partial', 'failed'].includes(statusData.status)) {
        console.log('\n=== JOB COMPLETE ===');
        console.log(`Final status: ${statusData.status}`);
        console.log(`Results available at: /v1/crawl/jobs/${jobId}?include_results=true`);
        return;
      }
    }

    console.log('\nTimeout: Job did not complete in time');
  } catch (error) {
    console.error('Error:', error);
  }
}

main();
