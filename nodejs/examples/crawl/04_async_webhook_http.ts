#!/usr/bin/env npx ts-node
/**
 * Async Crawl with Webhook - HTTP Example
 *
 * This script demonstrates async crawling with webhook notification.
 * Create a job with a webhook URL - the API will POST results when complete.
 * No polling required!
 *
 * Usage:
 *     npx ts-node 04_async_webhook_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 *
 * Webhook Payload:
 *     The API will POST to your webhook_url with:
 *     {
 *         "job_id": "job_123",
 *         "status": "completed",
 *         "progress": {"completed": 4, "failed": 0, "total": 4},
 *         "results": [...],  // Full crawl results
 *         "created_at": "2024-01-01T00:00:00Z",
 *         "completed_at": "2024-01-01T00:01:00Z"
 *     }
 */

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key
const API_URL = 'https://api.crawl4ai.com';
const WEBHOOK_URL = 'https://your-webhook-endpoint.com/callback'; // Your webhook URL

async function main(): Promise<void> {
  // URLs to crawl (can be more than 10 for async)
  const urls = [
    'https://example.com',
    'https://httpbin.org/html',
    'https://httpbin.org/json',
    'https://httpbin.org/robots.txt',
  ];

  console.log(`Creating async job for ${urls.length} URLs with webhook...`);

  try {
    // Create async job with webhook
    const response = await fetch(`${API_URL}/v1/crawl/async`, {
      method: 'POST',
      headers: {
        'X-API-Key': API_KEY,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        urls,
        strategy: 'http', // Options: "browser" (JS support) or "http" (faster, no JS)
        webhook_url: WEBHOOK_URL, // API will POST here when complete
        priority: 7, // Higher priority (1-10)
      }),
    });

    if (!response.ok) {
      throw new Error(`HTTP Error ${response.status}: ${await response.text()}`);
    }

    const data = await response.json();

    // Display job info
    console.log('\n=== JOB CREATED ===');
    console.log(`Job ID: ${data.job_id}`);
    console.log(`Status: ${data.status}`);
    console.log(`Webhook: ${WEBHOOK_URL}`);
    console.log('\nThe API will POST results to your webhook when complete.');
    console.log('No polling required!');

    // You can still check status manually if needed
    console.log(`\nManual status check: GET ${API_URL}/v1/crawl/jobs/${data.job_id}`);
  } catch (error) {
    console.error('Error:', error);
  }
}

main();

// Example webhook handler (Express.js):
/*
import express from 'express';

const app = express();
app.use(express.json());

app.post('/callback', (req, res) => {
    const data = req.body;
    console.log(`Job ${data.job_id} completed!`);
    console.log(`Status: ${data.status}`);
    console.log(`Results: ${data.results.length} URLs crawled`);
    res.json({ status: 'received' });
});

app.listen(8000, () => {
    console.log('Webhook handler listening on port 8000');
});
*/
