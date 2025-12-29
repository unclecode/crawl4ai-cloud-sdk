#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - Best-First Strategy with Scoring
 *
 * Best-first crawling prioritizes URLs based on relevance scores.
 * Uses a priority queue to always crawl the highest-scoring URL next.
 *
 * Scorers available:
 * - keywords: Score based on keyword presence in URL/title
 * - optimal_depth: Prefer URLs at a specific depth level
 *
 * Usage:
 *     npx ts-node 03_best_first_scoring.ts
 */

import { AsyncWebCrawler, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function bestFirstWithKeywords(): Promise<void> {
  console.log('=== Best-First with Keyword Scoring ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'best_first',
      maxDepth: 3,
      maxUrls: 15,
      scorers: {
        keywords: ['api', 'tutorial', 'guide', 'example'],
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Pages crawled: ${result.progress.completed}`);

    if (result.results) {
      console.log('\nTop results (by score):');
      result.results.slice(0, 5).forEach((r: any, i: number) => {
        console.log(`  ${i + 1}. ${r.url}`);
      });
    }
  } finally {
    await crawler.close();
  }
}

async function bestFirstForDocumentation(): Promise<void> {
  console.log('\n=== Best-First for API Docs ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const result = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'best_first',
      maxDepth: 3,
      maxUrls: 30,
      scorers: {
        keywords: ['api', 'reference', 'method', 'function', 'parameter'],
        optimal_depth: 2,
        weights: { keywords: 3.0, depth: 1.0 },
      },
      filters: { patterns: ['/api/*', '/reference/*', '/docs/*'] },
      wait: true,
    })) as CrawlJob;

    console.log(`API docs found: ${result.progress.completed}`);
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await bestFirstWithKeywords();
  // await bestFirstForDocumentation();
}

main().catch(console.error);
