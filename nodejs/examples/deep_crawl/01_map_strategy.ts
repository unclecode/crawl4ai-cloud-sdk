#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - Map Strategy (Sitemap Discovery)
 *
 * The "map" strategy discovers URLs from a website's sitemap and crawls them.
 * This is the fastest way to crawl a site with a well-structured sitemap.
 *
 * Features:
 * - Automatic sitemap discovery (sitemap.xml, sitemap_index.xml)
 * - URL pattern filtering
 * - Optional Common Crawl fallback
 *
 * Usage:
 *     npx ts-node 01_map_strategy.ts
 */

import { AsyncWebCrawler, DeepCrawlResult, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function basicMapCrawl(): Promise<void> {
  console.log('=== Basic Map Strategy ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // wait=true blocks until all URLs are crawled
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map', // Default strategy
      maxUrls: 5,
      wait: true,
    })) as CrawlJob;

    console.log(`Status: ${result.status}`);
    console.log(`URLs crawled: ${result.progress.completed}/${result.progress.total}`);

    if (result.results) {
      console.log(`\nResults (${result.results.length} pages):`);
      result.results.slice(0, 5).forEach((r: any) => {
        console.log(`  - ${r.url}: ${r.success}`);
      });
    }
  } finally {
    await crawler.close();
  }
}

async function mapWithPattern(): Promise<void> {
  console.log('\n=== Map Strategy with Pattern Filter ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map',
      pattern: '*/api/*', // Only URLs containing /api/
      maxUrls: 10,
      wait: true,
    })) as CrawlJob;

    console.log(`Matched URLs: ${result.progress.total}`);
    console.log(`Successfully crawled: ${result.progress.completed}`);
  } finally {
    await crawler.close();
  }
}

async function mapWithCommonCrawl(): Promise<void> {
  console.log('\n=== Map Strategy with Common Crawl ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map',
      source: 'sitemap+cc', // Try sitemap first, then Common Crawl
      maxUrls: 10,
      wait: true,
    })) as CrawlJob;

    console.log(`URLs discovered: ${result.progress.total}`);
    console.log('Source: sitemap + Common Crawl index');
  } finally {
    await crawler.close();
  }
}

async function mapNoWait(): Promise<void> {
  console.log('\n=== Map Strategy (No Wait) ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // wait=false returns immediately with scan job info
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map',
      maxUrls: 5,
      wait: false, // Don't wait
    })) as DeepCrawlResult;

    console.log(`Scan Job ID: ${result.jobId}`);
    console.log(`Status: ${result.status}`); // Will be "pending" initially
    console.log(`Discovered so far: ${result.discoveredCount}`);
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await basicMapCrawl();
  // Uncomment to run other examples:
  // await mapWithPattern();
  // await mapWithCommonCrawl();
  // await mapNoWait();
}

main().catch(console.error);
