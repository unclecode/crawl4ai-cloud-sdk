#!/usr/bin/env npx ts-node
/**
 * Schema Generation with HTTP
 *
 * This example shows how to automatically generate CSS extraction schemas
 * from HTML using LLM via direct HTTP API calls (no SDK).
 *
 * Usage:
 *     npx ts-node 03_schema_generation_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_URL = 'https://api.crawl4ai.com';
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function generateExtractionSchema(): Promise<void> {
  const headers = {
    'X-API-Key': API_KEY,
    'Content-Type': 'application/json',
  };

  // First, get the HTML content
  console.log('Fetching Hacker News HTML...');
  const crawlResponse = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url: 'https://news.ycombinator.com',
      strategy: 'http',
    }),
  });

  if (!crawlResponse.ok) {
    console.log(`Error: ${crawlResponse.status} - ${await crawlResponse.text()}`);
    return;
  }

  const crawlData = await crawlResponse.json();
  const html = crawlData.html || '';
  console.log(`Got ${html.length} bytes of HTML`);

  // Generate schema using LLM
  console.log('\nGenerating CSS extraction schema...');
  const schemaResponse = await fetch(`${API_URL}/v1/tools/schema`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      html,
      query: 'Extract all stories with their title, URL, points, and author',
      schema_type: 'CSS',
    }),
  });

  if (!schemaResponse.ok) {
    console.log(`Error: ${schemaResponse.status} - ${await schemaResponse.text()}`);
    return;
  }

  const schemaData = await schemaResponse.json();

  if (schemaData.error) {
    console.log(`Error: ${schemaData.error}`);
    return;
  }

  const schema = schemaData.schema;
  console.log('\nGenerated Schema:');
  console.log(JSON.stringify(schema, null, 2));

  // Now use the generated schema for extraction
  console.log('\n\nTesting generated schema...');
  const extractResponse = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers,
    body: JSON.stringify({
      url: 'https://news.ycombinator.com',
      strategy: 'http',
      crawler_config: {
        extraction_strategy: {
          type: 'json_css',
          schema,
        },
      },
    }),
  });

  if (!extractResponse.ok) {
    console.log(`Error: ${extractResponse.status} - ${await extractResponse.text()}`);
    return;
  }

  const extractData = await extractResponse.json();
  const stories = extractData.extracted_content || [];
  console.log(`\nExtracted ${stories.length} stories`);
  console.log('\nFirst 2 stories:');
  stories.slice(0, 2).forEach((story: any) => {
    console.log(`\n${JSON.stringify(story, null, 2)}`);
  });
}

generateExtractionSchema().catch(console.error);
