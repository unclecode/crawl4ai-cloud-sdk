#!/usr/bin/env npx ts-node
/**
 * LLM Extraction with HTTP
 *
 * This example shows how to extract structured data using LLM
 * via direct HTTP API calls (no SDK).
 *
 * Usage:
 *     npx ts-node 02_llm_extraction_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_URL = 'https://api.crawl4ai.com';
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function extractWithLlm(): Promise<void> {
  // Define LLM extraction strategy
  const payload = {
    url: 'https://news.ycombinator.com',
    strategy: 'http',
    crawler_config: {
      extraction_strategy: {
        type: 'llm',
        provider: 'crawl4ai',
        model: 'openai/gpt-4o-mini',
        instruction: `Extract all stories from this Hacker News page.
                For each story, extract:
                - title: The story title
                - url: The story URL
                - points: Number of points (if available)
                - author: Username who posted it
                - comments: Number of comments

                Return as a JSON array of story objects.`,
      },
    },
  };

  console.log('Crawling Hacker News with LLM extraction...');

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
    console.log(`Comments: ${story.comments}`);
  });
}

extractWithLlm().catch(console.error);
