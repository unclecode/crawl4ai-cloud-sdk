#!/usr/bin/env npx ts-node
/**
 * Sessions - Using Session with Playwright
 *
 * This script demonstrates how to create a browser session on Crawl4AI Cloud
 * and connect to it using Playwright.
 *
 * Usage:
 *     npx ts-node 03_session_with_playwright.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud playwright-core
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Try to import playwright-core
let playwright: any;
try {
  playwright = require('playwright-core');
} catch {
  console.log('Note: playwright-core not installed. Install with: npm install playwright-core');
}

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function crawlWithPlaywright(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Step 1: Create a session on Crawl4AI Cloud
    console.log('Creating browser session on Crawl4AI Cloud...');
    const session = await (crawler as any).createSession({ timeout: 600 });

    console.log('\n=== SESSION CREATED ===');
    console.log(`Session ID: ${session.sessionId}`);
    console.log(`WebSocket URL: ${session.wsUrl}`);

    if (!playwright) {
      console.log('\n=== PLAYWRIGHT EXAMPLE CODE ===');
      console.log('Install playwright-core to run this example:');
      console.log('  npm install playwright-core\n');
      console.log(`Use this code to connect:\n`);
      console.log(`const browser = await chromium.connectOverCDP('${session.wsUrl}');`);

      // Release the session
      await (crawler as any).releaseSession(session.sessionId);
      return;
    }

    // Step 2: Connect to the session using Playwright
    console.log('\nConnecting to session with Playwright...');
    const browser = await playwright.chromium.connectOverCDP(session.wsUrl);

    try {
      const context = browser.contexts()[0] || (await browser.newContext());
      const page = await context.newPage();

      console.log(`Navigating to ${url}...`);
      await page.goto(url);

      // Extract content
      const title = await page.title();
      const content = await page.content();

      console.log('\n=== CRAWL RESULTS ===');
      console.log(`Title: ${title}`);
      console.log(`Content length: ${content.length} characters`);

      // Take a screenshot
      await page.screenshot({ path: 'playwright_screenshot.png' });
      console.log('Screenshot saved to playwright_screenshot.png');

      // Extract data
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
      await browser.close();
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
  await crawlWithPlaywright('https://www.example.com');
}

main().catch(console.error);
