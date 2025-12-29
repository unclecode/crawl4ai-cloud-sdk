#!/usr/bin/env npx ts-node
/**
 * LLM Extraction with SDK
 *
 * This example shows how to extract structured data using LLM with natural language instructions.
 * LLM extraction is flexible and can handle complex extraction needs.
 *
 * Usage:
 *     npx ts-node 02_llm_extraction_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function extractWithLlm(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Define LLM extraction strategy
    const config = {
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
    };

    console.log('Crawling Hacker News with LLM extraction...');
    const result = await crawler.run('https://news.ycombinator.com', {
      strategy: 'http',
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
        console.log(`Comments: ${story.comments}`);
      });

      // Show token usage
      if (result.llmUsage) {
        console.log(`\nLLM Tokens Used: ${result.llmUsage.totalTokens}`);
      }
    } else {
      console.log(`Error: ${result.errorMessage}`);
    }
  } finally {
    await crawler.close();
  }
}

extractWithLlm().catch(console.error);
