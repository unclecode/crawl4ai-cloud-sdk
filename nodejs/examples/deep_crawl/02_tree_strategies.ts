#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - Tree Traversal Strategies (BFS & DFS)
 *
 * Tree strategies crawl by following links from the start URL:
 * - BFS (Breadth-First): Explore all pages at depth N before depth N+1
 * - DFS (Depth-First): Follow links deeply before backtracking
 *
 * Use these when:
 * - Site has no sitemap
 * - You want to explore link structure
 * - You need depth-limited crawling
 *
 * Usage:
 *     npx ts-node 02_tree_strategies.ts
 */

import { AsyncWebCrawler, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function bfsCrawl(): Promise<void> {
  console.log('=== BFS Strategy (Breadth-First) ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 20,
      wait: true,
    })) as CrawlJob;

    console.log(`Status: ${result.status}`);
    console.log(`Pages crawled: ${result.progress.completed}`);

    if (result.results) {
      console.log('\nCrawled pages:');
      result.results.slice(0, 5).forEach((r: any) => {
        console.log(`  - ${r.url}`);
      });
    }
  } finally {
    await crawler.close();
  }
}

async function dfsCrawl(): Promise<void> {
  console.log('\n=== DFS Strategy (Depth-First) ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'dfs',
      maxDepth: 3,
      maxUrls: 15,
      wait: true,
    })) as CrawlJob;

    console.log(`Status: ${result.status}`);
    console.log(`Pages crawled: ${result.progress.completed}`);
  } finally {
    await crawler.close();
  }
}

async function treeWithFilters(): Promise<void> {
  console.log('\n=== BFS with URL Filters ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 25,
      filters: {
        patterns: ['/docs/*', '/api/*', '/guide/*'],
        domains: { blocked: ['twitter.com', 'github.com'] },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Filtered pages crawled: ${result.progress.completed}`);
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await bfsCrawl();
  // Uncomment to run other examples:
  // await dfsCrawl();
  // await treeWithFilters();
}

main().catch(console.error);
