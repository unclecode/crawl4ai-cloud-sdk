#!/usr/bin/env npx ts-node
/**
 * Deep Crawl - With Extraction Strategies
 *
 * Combine deep crawl with extraction strategies to get structured data
 * from all discovered pages. Extraction runs during the crawl phase.
 *
 * Extraction strategies:
 * - json_css: CSS selectors for structured data
 * - llm: LLM-based extraction (requires LLM config)
 * - cosine: Semantic clustering
 *
 * Usage:
 *     npx ts-node 07_with_extraction.ts
 */

import { AsyncWebCrawler, CrawlJob } from 'crawl4ai-cloud';

const API_KEY = 'YOUR_API_KEY';

async function cssExtraction(): Promise<void> {
  console.log('=== CSS Extraction ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Define CSS extraction schema
    const schema = {
      name: 'Documentation',
      baseSelector: 'main, article, .content',
      fields: [
        {
          name: 'title',
          selector: 'h1',
          type: 'text',
        },
        {
          name: 'description',
          selector: "p.description, .intro, meta[name='description']",
          type: 'text',
        },
        {
          name: 'headings',
          selector: 'h2, h3',
          type: 'list',
        },
        {
          name: 'code_blocks',
          selector: 'pre code, .highlight',
          type: 'list',
        },
      ],
    };

    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 1,
      maxUrls: 5,
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema,
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Pages crawled: ${job.progress.completed}`);

    if (job.results) {
      console.log('\nExtracted content:');
      job.results.slice(0, 3).forEach((r: any) => {
        console.log(`\nURL: ${r.url}`);
        if (r.extracted_content) {
          try {
            const data = JSON.parse(r.extracted_content);
            console.log(`  Title: ${data.title || 'N/A'}`);
            const headings = data.headings || [];
            if (headings.length) {
              console.log(`  Headings: ${headings.length}`);
              headings.slice(0, 3).forEach((h: string) => {
                console.log(`    - ${h}`);
              });
            }
          } catch {
            console.log('  (Parse error)');
          }
        }
      });
    }
  } finally {
    await crawler.close();
  }
}

async function nestedCssExtraction(): Promise<void> {
  console.log('\n=== Nested CSS Extraction ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Schema for extracting repeated items (e.g., products, posts)
    const schema = {
      name: 'APIReference',
      baseSelector: '.method, .function, .endpoint',
      fields: [
        {
          name: 'name',
          selector: 'h3, .method-name',
          type: 'text',
        },
        {
          name: 'signature',
          selector: '.signature, code:first-of-type',
          type: 'text',
        },
        {
          name: 'description',
          selector: 'p, .description',
          type: 'text',
        },
        {
          name: 'parameters',
          selector: '.param, li',
          type: 'nested',
          fields: [
            { name: 'name', selector: '.param-name, code', type: 'text' },
            { name: 'type', selector: '.param-type, em', type: 'text' },
            { name: 'desc', selector: '.param-desc, span', type: 'text' },
          ],
        },
      ],
    };

    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'bfs',
      maxDepth: 1,
      maxUrls: 5,
      pattern: '*/api/*',
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema,
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`API pages: ${job.progress.completed}`);
  } finally {
    await crawler.close();
  }
}

async function extractionWithAttributes(): Promise<void> {
  console.log('\n=== Extract Attributes ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const schema = {
      name: 'PageAssets',
      baseSelector: 'body',
      fields: [
        {
          name: 'links',
          selector: 'a[href]',
          type: 'list',
          attribute: 'href', // Get href attribute
        },
        {
          name: 'images',
          selector: 'img[src]',
          type: 'list',
          attribute: 'src', // Get src attribute
        },
        {
          name: 'meta_tags',
          selector: 'meta[name]',
          type: 'nested',
          fields: [
            { name: 'name', selector: '', type: 'attribute', attribute: 'name' },
            { name: 'content', selector: '', type: 'attribute', attribute: 'content' },
          ],
        },
      ],
    };

    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map',
      maxUrls: 3,
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema,
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`Pages processed: ${job.progress.completed}`);

    if (job.results) {
      job.results.slice(0, 1).forEach((r: any) => {
        if (r.extracted_content) {
          const data = JSON.parse(r.extracted_content);
          const links = data.links || [];
          const images = data.images || [];
          console.log(`\nURL: ${r.url}`);
          console.log(`  Links found: ${links.length}`);
          console.log(`  Images found: ${images.length}`);
        }
      });
    }
  } finally {
    await crawler.close();
  }
}

async function llmExtraction(): Promise<void> {
  console.log('\n=== LLM Extraction ===\n');

  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    const job = (await crawler.deepCrawl('https://docs.crawl4ai.com', {
      strategy: 'map',
      maxUrls: 3,
      config: {
        extraction_strategy: {
          type: 'llm',
          provider: 'openai', // or "anthropic", "ollama"
          model: 'gpt-4o-mini',
          schema: {
            name: 'PageSummary',
            fields: [
              { name: 'title', type: 'string' },
              { name: 'summary', type: 'string', description: '2-3 sentence summary' },
              { name: 'topics', type: 'list', description: 'Main topics covered' },
              { name: 'code_examples', type: 'boolean', description: 'Has code examples?' },
            ],
          },
          instruction: 'Extract the main content and summarize this documentation page.',
        },
      },
      wait: true,
    })) as CrawlJob;

    console.log(`LLM processed: ${job.progress.completed}`);
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await cssExtraction();
  // Uncomment to run other examples:
  // await nestedCssExtraction();
  // await extractionWithAttributes();
  // await llmExtraction();  // Requires LLM provider setup
}

main().catch(console.error);
