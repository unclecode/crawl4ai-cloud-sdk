#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - URL Filtering and Patterns
 *
 * Control which URLs are crawled using filters and patterns:
 * - Glob patterns: Match URL paths (e.g., "* /docs/*")
 * - Domain filters: Allow/block specific domains
 * - Nonsense URL filter: Skip auth, tracking, utility pages
 *
 * Filtering happens during discovery, before crawling,
 * so you don't waste resources on unwanted pages.
 *
 * Usage:
 *     npx ts-node 06_filters_and_patterns.ts
 */

import { AsyncWebCrawler, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function basicPatternFilter(): Promise<void> {
  console.log('=== Basic Pattern Filter ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 20,
      pattern: '*/docs/*', // Only URLs with /docs/ in path
      wait: true,
    })) as CrawlJob;

    console.log(`Matched URLs: ${job.progress.total}`);
    console.log(`Crawled: ${job.progress.completed}`);

    if (job.results) {
      console.log('\nCrawled URLs:');
      job.results.slice(0, 5).forEach((r: any) => {
        console.log(`  - ${r.url}`);
      });
    }
  } finally {
    await crawler.close();
  }
}

async function multiplePatterns(): Promise<void> {
  console.log('\n=== Multiple Patterns ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 30,
      filters: {
        // Match any of these patterns
        patterns: [
          '/api/*', // API reference pages
          '/guide/*', // User guides
          '/tutorial/*', // Tutorials
          '*/example*', // Example pages
        ],
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Matching URLs: ${job.progress.total}`);
  } finally {
    await crawler.close();
  }
}

async function excludePatterns(): Promise<void> {
  console.log('\n=== Exclude Patterns ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 30,
      filters: {
        // Exclude these patterns
        exclude_patterns: [
          '*/changelog/*', // Skip changelogs
          '*/archive/*', // Skip archives
          '*?page=*', // Skip pagination
          '*#*', // Skip anchor links
        ],
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Filtered URLs: ${job.progress.total}`);
  } finally {
    await crawler.close();
  }
}

async function domainFiltering(): Promise<void> {
  console.log('\n=== Domain Filtering ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 25,
      filters: {
        domains: {
          // Never follow links to these domains
          blocked: [
            'twitter.com',
            'facebook.com',
            'linkedin.com',
            'github.com', // If you want to stay on docs
          ],
          // Or whitelist: only follow links to these
          // allowed: ["docs.crawl4ai.com", "crawl4ai.com"]
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Crawled (blocked external): ${job.progress.total}`);
  } finally {
    await crawler.close();
  }
}

async function combinedFilters(): Promise<void> {
  console.log('\n=== Combined Filters ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 3,
      maxUrls: 50,
      filters: {
        // Include patterns (whitelist)
        patterns: ['/docs/*', '/api/*', '/guide/*'],
        // Exclude patterns
        exclude_patterns: ['*changelog*', '*version*'],
        // Domain controls
        domains: {
          blocked: ['twitter.com', 'github.com'],
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Filtered & crawled: ${job.progress.total}`);
  } finally {
    await crawler.close();
  }
}

async function filterForEcommerce(): Promise<void> {
  console.log('\n=== E-commerce Product Filter ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://example-shop.com', {
      strategy: 'bfs',
      maxDepth: 3,
      maxUrls: 100,
      filters: {
        patterns: ['/product/*', '/products/*', '/item/*', '/shop/*'],
        exclude_patterns: [
          '*/cart*',
          '*/checkout*',
          '*/login*',
          '*/account*',
          '*?sort=*', // Skip sorted views
          '*?filter=*', // Skip filtered views
        ],
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Product pages found: ${job.progress.total}`);
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await basicPatternFilter();
  // Uncomment to run other examples:
  // await multiplePatterns();
  // await excludePatterns();
  // await domainFiltering();
  // await combinedFilters();
  // await filterForEcommerce();
}

main().catch(console.error);
