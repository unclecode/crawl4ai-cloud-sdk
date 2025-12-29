#!/usr/bin/env npx ts-node
/**
 * Schema Generation with SDK
 *
 * This example shows how to automatically generate CSS extraction schemas
 * from HTML using LLM. The schema can then be reused for fast, no-cost extraction.
 *
 * Usage:
 *     npx ts-node 03_schema_generation_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function generateExtractionSchema(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // First, get the HTML content
    console.log('Fetching Hacker News HTML...');
    const result = await crawler.run('https://news.ycombinator.com', {
      strategy: 'http',
    });

    const html = result.html || '';
    console.log(`Got ${html.length} bytes of HTML`);

    // Generate schema using LLM
    console.log('\nGenerating CSS extraction schema...');
    const schemaResult = await crawler.generateSchema(html, {
      query: 'Extract all stories with their title, URL, points, and author',
    });

    if (schemaResult.error) {
      console.log(`Error: ${schemaResult.error}`);
      return;
    }

    console.log('\nGenerated Schema:');
    console.log(JSON.stringify(schemaResult.schema, null, 2));

    // Now use the generated schema for extraction
    console.log('\n\nTesting generated schema...');
    const extractResult = await crawler.run('https://news.ycombinator.com', {
      strategy: 'http',
      config: {
        extraction_strategy: {
          type: 'json_css',
          schema: schemaResult.schema,
        },
      },
    });

    if (extractResult.success && extractResult.extractedContent) {
      const stories = JSON.parse(extractResult.extractedContent);
      console.log(`\nExtracted ${stories.length} stories`);
      console.log('\nFirst 2 stories:');
      stories.slice(0, 2).forEach((story: any) => {
        console.log(`\n${JSON.stringify(story, null, 2)}`);
      });
    }
  } finally {
    await crawler.close();
  }
}

generateExtractionSchema().catch(console.error);
