#!/usr/bin/env npx ts-node
/**
 * CSS Extraction with HTTP - No LLM Cost
 *
 * This example shows how to extract structured data using CSS selectors
 * via direct HTTP API calls (no SDK).
 *
 * Usage:
 *     npx ts-node 01_css_extraction_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_URL = 'https://api.crawl4ai.com';
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function extractWithCss(): Promise<void> {
  // Define CSS extraction schema
  const payload = {
    url: 'https://news.ycombinator.com',
    strategy: 'http', // Fast, no browser needed
    crawler_config: {
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
    },
  };

  console.log('Crawling Hacker News with CSS extraction...');

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return;
  }

  const result = await response.json();
  const stories = result.extracted_content || [];

  console.log(`\nExtracted ${stories.length} stories`);
  console.log('\nFirst 3 stories:');
  stories.slice(0, 3).forEach((story: any) => {
    console.log(`\nTitle: ${story.title}`);
    console.log(`URL: ${story.url}`);
    console.log(`Points: ${story.points}`);
    console.log(`Author: ${story.author}`);
  });
}

extractWithCss().catch(console.error);
