#!/usr/bin/env npx ts-node
/**
 * Example: List jobs using SDK
 *
 * This example demonstrates:
 * - Listing all jobs with pagination
 * - Filtering jobs by status
 * - Accessing job metadata (timestamps, URLs, status)
 *
 * Usage:
 *     npx ts-node 01_list_jobs_sdk.ts
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function main(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // List all jobs (default: 20 per page)
    console.log('=== All Jobs (First 20) ===');
    const jobs = await crawler.listJobs({ limit: 20 });
    console.log(`Total jobs: ${jobs.length}`);

    jobs.forEach((job) => {
      console.log(`  ${job.id}: ${job.status} | ${job.urlsCount} URLs | Created: ${job.createdAt}`);
    });

    // Filter by status
    console.log('\n=== Completed Jobs ===');
    const completed = await crawler.listJobs({ status: 'completed', limit: 10 });
    completed.forEach((job) => {
      console.log(`  ${job.id}`);
    });

    console.log('\n=== Running Jobs ===');
    const running = await crawler.listJobs({ status: 'running' });
    console.log(`Found ${running.length} running jobs`);

    // Pagination example
    console.log('\n=== Pagination (Next 20) ===');
    const page2 = await crawler.listJobs({ limit: 20, offset: 20 });
    console.log(`Page 2: ${page2.length} jobs`);

    // Available statuses: pending, running, completed, failed, cancelled
    console.log('\n=== Failed Jobs ===');
    const failed = await crawler.listJobs({ status: 'failed', limit: 5 });
    failed.forEach((job) => {
      console.log(`  ${job.id}: ${job.error || 'Unknown error'}`);
    });
  } finally {
    await crawler.close();
  }
}

main().catch(console.error);
