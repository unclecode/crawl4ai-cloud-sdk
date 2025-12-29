#!/usr/bin/env npx ts-node
/**
 * Storage Usage Monitoring - SDK Example
 *
 * This script demonstrates how to check and monitor your storage quota.
 * Storage is used by async job results stored in S3.
 *
 * Usage:
 *     npx ts-node 04_storage_usage_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler, CrawlResult, CrawlJob } from 'crawl4ai-cloud';
import { QuotaExceededError } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function checkStorageUsage(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log('Checking storage usage...');
    const usage = await (crawler as any).getStorageUsage();

    console.log('\n=== STORAGE USAGE ===');
    console.log(`Used: ${usage.usedMb.toFixed(2)} MB`);
    console.log(`Max: ${usage.maxMb.toFixed(2)} MB`);
    console.log(`Remaining: ${usage.remainingMb.toFixed(2)} MB`);
    console.log(`Usage: ${usage.percentage.toFixed(1)}%`);

    // Check if storage is getting full
    if (usage.percentage > 90) {
      console.log('\nWARNING: Storage is over 90% full!');
      console.log('Consider deleting old jobs to free up space.');
    } else if (usage.percentage > 75) {
      console.log('\nNOTE: Storage is over 75% full.');
    } else {
      console.log('\nStorage usage is healthy.');
    }
  } finally {
    await crawler.close();
  }
}

async function monitorStorageDuringCrawl(urls: string[]): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Check initial storage
    const initial = await (crawler as any).getStorageUsage();
    console.log(`Initial storage: ${initial.usedMb.toFixed(2)} MB / ${initial.maxMb.toFixed(2)} MB`);

    try {
      // Create async job
      console.log(`\nStarting async crawl for ${urls.length} URLs...`);
      const results = (await crawler.runMany(urls, {
        wait: true, // Wait for completion
      })) as CrawlResult[];

      console.log(`Crawl completed: ${results.length} results`);

      // Check storage after job
      const after = await (crawler as any).getStorageUsage();
      console.log(
        `\nAfter crawl storage: ${after.usedMb.toFixed(2)} MB / ${after.maxMb.toFixed(2)} MB`
      );
      console.log(`Storage used by this job: ${(after.usedMb - initial.usedMb).toFixed(2)} MB`);
    } catch (e) {
      if (e instanceof QuotaExceededError && e.quotaType === 'storage') {
        console.log('\nStorage quota exceeded!');
        console.log('Delete old jobs to free up space.');

        // List jobs to find candidates for deletion
        const jobs = await crawler.listJobs({ limit: 10, status: 'completed' });
        console.log(`\nRecent completed jobs (${jobs.length}):`);
        jobs.slice(0, 5).forEach((job) => {
          console.log(`  - ${job.id} (created: ${job.createdAt})`);
        });
      } else {
        throw e;
      }
    }
  } finally {
    await crawler.close();
  }
}

async function cleanupOldJobs(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Check current storage
    const usage = await (crawler as any).getStorageUsage();
    console.log(`Current storage: ${usage.usedMb.toFixed(2)} MB / ${usage.maxMb.toFixed(2)} MB`);

    // List completed jobs
    const jobs = await crawler.listJobs({ limit: 20, status: 'completed' });
    console.log(`\nFound ${jobs.length} completed jobs`);

    if (!jobs.length) {
      console.log('No jobs to delete.');
      return;
    }

    // Delete oldest jobs (be careful in production!)
    console.log('\nDeleting oldest 3 jobs...');
    let deletedCount = 0;

    const oldestJobs = jobs.slice(-3); // Last 3 (oldest with default sorting)
    for (const job of oldestJobs) {
      try {
        await crawler.cancelJob(job.id);
        console.log(`  Deleted job ${job.id}`);
        deletedCount++;
      } catch (e) {
        console.log(`  Failed to delete ${job.id}: ${e}`);
      }
    }

    // Check storage after cleanup
    if (deletedCount > 0) {
      const after = await (crawler as any).getStorageUsage();
      const freed = usage.usedMb - after.usedMb;
      console.log(`\nFreed ${freed.toFixed(2)} MB of storage`);
      console.log(`New usage: ${after.usedMb.toFixed(2)} MB / ${after.maxMb.toFixed(2)} MB`);
    }
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  // Example 1: Check storage
  console.log('=== Example 1: Check Storage Usage ===');
  await checkStorageUsage();

  // Example 2: Monitor during crawl (uncomment to use)
  // console.log('\n=== Example 2: Monitor During Crawl ===');
  // await monitorStorageDuringCrawl([
  //     'https://www.example.com',
  //     'https://www.example.com/about',
  // ]);

  // Example 3: Cleanup old jobs (uncomment to use)
  // console.log('\n=== Example 3: Cleanup Old Jobs ===');
  // await cleanupOldJobs();
}

main().catch(console.error);
