#!/usr/bin/env npx ts-node
/**
 * Example: Job management using HTTP API
 *
 * This example demonstrates:
 * - GET /v1/crawl/jobs/{id} - Get job details
 * - DELETE /v1/crawl/jobs/{id} - Cancel/delete job
 * - GET /v1/crawl/jobs/{id}/download - Get download URL
 *
 * Usage:
 *     npx ts-node 02_job_management_http.ts
 */

// Configuration
const API_KEY = 'YOUR_API_KEY';
const BASE_URL = 'https://api.crawl4ai.com';

const headers = {
  Authorization: `Bearer ${API_KEY}`,
  'Content-Type': 'application/json',
};

async function main(): Promise<void> {
  // Create a test job
  console.log('=== Creating Test Job ===');
  let response = await fetch(`${BASE_URL}/v1/crawl/async`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      urls: ['https://example.com', 'https://example.org'],
      priority: 5,
    }),
  });

  const job = await response.json();
  const jobId = job.job_id;
  console.log(`Created job: ${jobId}`);

  // Get job details
  console.log('\n=== Get Job Details ===');
  response = await fetch(`${BASE_URL}/v1/crawl/jobs/${jobId}`, { headers });
  const jobDetails = await response.json();
  console.log(`Status: ${jobDetails.status}`);
  console.log(`URLs: ${jobDetails.urls?.join(', ')}`);

  // Cancel job (keep results)
  console.log('\n=== Cancel Job (Keep Results) ===');
  response = await fetch(`${BASE_URL}/v1/crawl/jobs/${jobId}?delete_results=false`, {
    method: 'DELETE',
    headers,
  });
  const cancelled = await response.json();
  console.log(`Status: ${cancelled.status}`);

  // Create and delete completely
  console.log('\n=== Cancel + Delete Results ===');
  response = await fetch(`${BASE_URL}/v1/crawl/async`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ urls: ['https://example.com'] }),
  });
  const job2Id = (await response.json()).job_id;
  console.log(`Created job: ${job2Id}`);

  response = await fetch(`${BASE_URL}/v1/crawl/jobs/${job2Id}?delete_results=true`, {
    method: 'DELETE',
    headers,
  });
  console.log(`Deleted: ${(await response.json()).status}`);

  // Get download URL
  console.log('\n=== Get Download URL ===');
  try {
    // Find a completed job
    response = await fetch(`${BASE_URL}/v1/crawl/jobs?status=completed&limit=1`, { headers });
    const jobs = await response.json();

    if (jobs.jobs?.length) {
      const completedJobId = jobs.jobs[0].job_id;
      response = await fetch(`${BASE_URL}/v1/crawl/jobs/${completedJobId}/download?expires_in=3600`, {
        headers,
      });
      const downloadData = await response.json();
      console.log(`Download URL: ${downloadData.url?.slice(0, 100)}...`);
      console.log('URL expires in 3600 seconds (1 hour)');
    } else {
      console.log('No completed jobs found');
    }
  } catch (e) {
    console.log(`Error: ${e}`);
  }
}

main().catch(console.error);
