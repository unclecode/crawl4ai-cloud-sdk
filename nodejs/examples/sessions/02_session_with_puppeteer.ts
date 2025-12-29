#!/usr/bin/env npx ts-node
/**
 * Sessions - Using Session with Puppeteer
 *
 * This script demonstrates how to create a browser session on Crawl4AI Cloud
 * and connect to it using Puppeteer.
 *
 * Usage:
 *     npx ts-node 02_session_with_puppeteer.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud puppeteer-core
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Try to import puppeteer-core
let puppeteer: any;
try {
  puppeteer = require('puppeteer-core');
} catch {
  console.log('Note: puppeteer-core not installed. Install with: npm install puppeteer-core');
}

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function crawlWithPuppeteer(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Step 1: Create a session on Crawl4AI Cloud
    console.log('Creating browser session on Crawl4AI Cloud...');
    const session = await (crawler as any).createSession({ timeout: 600 });

    console.log('\n=== SESSION CREATED ===');
    console.log(`Session ID: ${session.sessionId}`);
    console.log(`WebSocket URL: ${session.wsUrl}`);

    if (!puppeteer) {
      console.log('\n=== PUPPETEER EXAMPLE CODE ===');
      console.log('Install puppeteer-core to run this example:');
      console.log('  npm install puppeteer-core\n');
      console.log(`Use this code to connect:\n`);
      console.log(`const browser = await puppeteer.connect({`);
      console.log(`  browserWSEndpoint: '${session.wsUrl}'`);
      console.log(`});`);

      // Release the session
      await (crawler as any).releaseSession(session.sessionId);
      return;
    }

    // Step 2: Connect to the session using Puppeteer
    console.log('\nConnecting to session with Puppeteer...');
    const browser = await puppeteer.connect({
      browserWSEndpoint: session.wsUrl,
    });

    try {
      const page = await browser.newPage();

      console.log(`Navigating to ${url}...`);
      await page.goto(url, { waitUntil: 'networkidle0' });

      // Extract content
      const title = await page.title();
      const content = await page.content();

      console.log('\n=== CRAWL RESULTS ===');
      console.log(`Title: ${title}`);
      console.log(`Content length: ${content.length} characters`);

      // Extract data using page.evaluate
      const data = await page.evaluate(() => {
        return {
          title: document.title,
          headings: Array.from(document.querySelectorAll('h1, h2'))
            .map((h) => (h as HTMLElement).textContent)
            .slice(0, 5),
        };
      });

      console.log('\nExtracted data:');
      console.log(`  Title: ${data.title}`);
      console.log(`  Headings: ${data.headings.join(', ')}`);
    } finally {
      // Don't call browser.close() - the session is managed by Crawl4AI Cloud
      await browser.disconnect();
    }

    // Step 3: Release the session
    console.log('\nReleasing session...');
    await (crawler as any).releaseSession(session.sessionId);
    console.log('Session released!');
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  await crawlWithPuppeteer('https://www.example.com');
}

main().catch(console.error);
