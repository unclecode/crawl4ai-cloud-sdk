#!/usr/bin/env npx ts-node
/**
 * CSS Extraction with SDK - No LLM Cost
 *
 * This example shows how to extract structured data using CSS selectors.
 * CSS extraction is fast, reliable, and has no LLM cost.
 *
 * Usage:
 *     npx ts-node 01_css_extraction_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function extractWithCss(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Define CSS extraction schema
    const config = {
      extraction_strategy: {
        type: 'json_css',
        schema: {
          name: 'HackerNewsStories',
          baseSelector: '.athing',
          fields: [
            { name: 'title', selector: '.titleline > a', type: 'text' },
            { name: 'url', selector: '.titleline > a', type: 'attribute', attribute: 'href' },
            { name: 'points', selector: '+ tr .score', type: 'text' },
            { name: 'author', selector: '+ tr .hnuser', type: 'text' },
          ],
        },
      },
    };

    console.log('Crawling Hacker News with CSS extraction...');
    const result = await crawler.run('https://news.ycombinator.com', {
      strategy: 'http', // Fast, no browser needed
      config,
    });

    if (result.success && result.extractedContent) {
      const stories = JSON.parse(result.extractedContent);
      console.log(`\nExtracted ${stories.length} stories`);
      console.log('\nFirst 3 stories:');
      stories.slice(0, 3).forEach((story: any) => {
        console.log(`\nTitle: ${story.title}`);
        console.log(`URL: ${story.url}`);
        console.log(`Points: ${story.points}`);
        console.log(`Author: ${story.author}`);
      });
    } else {
      console.log(`Error: ${result.errorMessage}`);
    }
  } finally {
    await crawler.close();
  }
}

extractWithCss().catch(console.error);
