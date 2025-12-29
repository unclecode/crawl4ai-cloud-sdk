#!/usr/bin/env npx ts-node
/**
 * Example: Job management using SDK
 *
 * This example demonstrates:
 * - Getting job details
 * - Cancelling running jobs
 * - Getting presigned download URLs
 *
 * Usage:
 *     npx ts-node 02_job_management_sdk.ts
 */

import { AsyncWebCrawler, CrawlJob } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function main(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Create an async job for testing
    console.log('=== Creating Test Job ===');
    const result = (await crawler.runMany(['https://example.com', 'https://example.org'], {
      wait: false, // Don't wait, just create the job
    })) as CrawlJob;

    console.log(`Created job: ${result.id}`);

    // Get job details
    console.log('\n=== Get Job Details ===');
    const job = await crawler.getJob(result.id, false);
    console.log(`Job ID: ${job.id}`);
    console.log(`Status: ${job.status}`);
    console.log(`URLs: ${job.urlsCount}`);
    console.log(`Created: ${job.createdAt}`);

    // Wait for job and get results
    console.log('\n=== Wait for Job ===');
    const completedJob = await crawler.waitJob(result.id, {
      pollInterval: 2,
      includeResults: true,
    });
    console.log(`Final Status: ${completedJob.status}`);
    console.log(`Progress: ${completedJob.progress.completed}/${completedJob.progress.total}`);

    // Cancel a job (create another one first)
    console.log('\n=== Cancel Job ===');
    const job2 = (await crawler.runMany(['https://example.com'], {
      wait: false,
    })) as CrawlJob;
    console.log(`Created job: ${job2.id}`);

    const cancelled = await crawler.cancelJob(job2.id);
    console.log(`Cancelled: ${cancelled}`);

    // Get download URL for completed job
    console.log('\n=== Get Download URL ===');
    try {
      const completedJobs = await crawler.listJobs({ status: 'completed', limit: 1 });
      if (completedJobs.length > 0) {
        const jobId = completedJobs[0].id;
        const downloadUrl = await crawler.downloadUrl(jobId, 3600);
        console.log(`Download URL: ${downloadUrl.slice(0, 100)}...`);
        console.log('URL expires in 3600 seconds (1 hour)');
      } else {
        console.log('No completed jobs found');
      }
    } catch (e) {
      console.log(`Error: ${e}`);
    }
  } finally {
    await crawler.close();
  }
}

main().catch(console.error);
