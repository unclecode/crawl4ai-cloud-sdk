#!/usr/bin/env npx ts-node
/**
 * Example: List jobs using HTTP API
 *
 * This example demonstrates:
 * - GET /v1/crawl/jobs with query parameters
 * - Pagination and filtering
 * - Raw JSON response handling
 *
 * Usage:
 *     npx ts-node 01_list_jobs_http.ts
 */

// Configuration
const API_KEY = 'YOUR_API_KEY';
const BASE_URL = 'https://api.crawl4ai.com';

const headers = {
  Authorization: `Bearer ${API_KEY}`,
  'Content-Type': 'application/json',
};

async function main(): Promise<void> {
  // List all jobs
  console.log('=== All Jobs (First 20) ===');
  let response = await fetch(`${BASE_URL}/v1/crawl/jobs?limit=20`, { headers });
  let data = await response.json();

  console.log(`Total jobs: ${data.total}`);
  console.log(`Showing: ${data.jobs.length}`);

  data.jobs.forEach((job: any) => {
    console.log(`  ${job.job_id}: ${job.status} | ${job.urls?.length || 0} URLs`);
  });

  // Filter by status
  console.log('\n=== Completed Jobs ===');
  response = await fetch(`${BASE_URL}/v1/crawl/jobs?status=completed&limit=10`, { headers });
  const completed = await response.json();
  completed.jobs.forEach((job: any) => {
    console.log(`  ${job.job_id}: ${job.urls?.[0] || 'N/A'}`);
  });

  // Pagination
  console.log('\n=== Pagination (Next 20) ===');
  response = await fetch(`${BASE_URL}/v1/crawl/jobs?limit=20&offset=20`, { headers });
  const page2 = await response.json();
  console.log(`Page 2: ${page2.jobs.length} jobs`);

  // Failed jobs
  console.log('\n=== Failed Jobs ===');
  response = await fetch(`${BASE_URL}/v1/crawl/jobs?status=failed&limit=5`, { headers });
  const failed = await response.json();
  failed.jobs.forEach((job: any) => {
    console.log(`  ${job.job_id}: ${job.error || 'Unknown error'}`);
  });
}

main().catch(console.error);
