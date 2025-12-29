#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - Two-Phase Workflow (Scan -> Extract)
 *
 * The two-phase workflow separates URL discovery from data extraction:
 *
 * Phase 1 (Scan):
 *   - Discover URLs using BFS/DFS/Best-First
 *   - Cache raw HTML in Redis (30-minute TTL)
 *   - Return scan_job_id for later use
 *
 * Phase 2 (Extract):
 *   - Use sourceJob to reference cached HTML
 *   - Apply extraction strategy (CSS, LLM, etc.)
 *   - No re-crawling needed - uses cached HTML
 *
 * Benefits:
 *   - Scan once, extract multiple times
 *   - Apply different extraction strategies to same data
 *   - Preview URLs before committing to full extraction
 *   - Faster extraction (HTML already cached)
 *
 * Usage:
 *     npx ts-node 05_two_phase_workflow.ts
 */

import { AsyncWebCrawler, DeepCrawlResult, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function basicTwoPhase(): Promise<void> {
  console.log('=== Two-Phase Workflow ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // ========== PHASE 1: SCAN ==========
    console.log('Phase 1: Scanning (URL discovery + HTML caching)...');

    const scanResult = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 2,
      maxUrls: 10,
      scanOnly: true, // Don't extract yet
      wait: true,
    })) as DeepCrawlResult;

    console.log(`  Scan Job ID: ${scanResult.jobId}`);
    console.log(`  URLs discovered: ${scanResult.discoveredCount}`);
    console.log(`  Cache expires: ${scanResult.cacheExpiresAt}`);
    console.log();

    // ========== PHASE 2: EXTRACT ==========
    console.log('Phase 2: Extracting from cached HTML...');

    const job = (await crawler.deepCrawl(undefined, {
      sourceJob: scanResult.jobId, // Use cached HTML
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema: {
            name: 'PageContent',
            baseSelector: 'main, article, .content',
            fields: [
              { name: 'title', selector: 'h1', type: 'text' },
              { name: 'headings', selector: 'h2, h3', type: 'list' },
            ],
          },
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`  Job ID: ${job.id}`);
    console.log(`  Pages extracted: ${job.progress.completed}`);

    if (job.results) {
      console.log('\nExtracted content:');
      job.results.slice(0, 3).forEach((r: any) => {
        console.log(`  URL: ${r.url}`);
        if (r.extracted_content) {
          try {
            const data = JSON.parse(r.extracted_content);
            console.log(`    Title: ${data.title || 'N/A'}`);
          } catch {
            // Parse error
          }
        }
      });
    }
  } finally {
    await crawler.close();
  }
}

async function multipleExtractions(): Promise<void> {
  console.log('\n=== Multiple Extractions from Same Scan ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // SCAN ONCE
    console.log('Scanning...');
    const scanResult = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 1,
      maxUrls: 5,
      scanOnly: true,
      wait: true,
    })) as DeepCrawlResult;

    const scanJobId = scanResult.jobId;
    console.log(`Scan Job ID: ${scanJobId}`);
    console.log(`URLs cached: ${scanResult.discoveredCount}\n`);

    // EXTRACT #1: Titles only
    console.log('Extraction 1: Titles...');
    const job1 = (await crawler.deepCrawl(undefined, {
      sourceJob: scanJobId,
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema: {
            name: 'Titles',
            baseSelector: 'body',
            fields: [{ name: 'title', selector: 'h1', type: 'text' }],
          },
        },
      },
      wait: true,
    })) as CrawlJob;
    console.log(`  Extracted: ${job1.progress.completed} pages`);

    // EXTRACT #2: Links
    console.log('Extraction 2: Links...');
    const job2 = (await crawler.deepCrawl(undefined, {
      sourceJob: scanJobId,
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema: {
            name: 'Links',
            baseSelector: 'body',
            fields: [{ name: 'links', selector: 'a[href]', type: 'list', attribute: 'href' }],
          },
        },
      },
      wait: true,
    })) as CrawlJob;
    console.log(`  Extracted: ${job2.progress.completed} pages`);

    // EXTRACT #3: Full markdown
    console.log('Extraction 3: Markdown...');
    const job3 = (await crawler.deepCrawl(undefined, {
      sourceJob: scanJobId,
      // No extraction strategy = get markdown
      wait: true,
    })) as CrawlJob;
    console.log(`  Extracted: ${job3.progress.completed} pages`);

    console.log('\nAll 3 extractions used the same cached HTML!');
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await basicTwoPhase();
  // Uncomment to run other examples:
  // await multipleExtractions();
}

main().catch(console.error);
