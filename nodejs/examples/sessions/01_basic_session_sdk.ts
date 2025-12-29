#!/usr/bin/env npx ts-node
/**
 * Sessions - Basic Session Management with SDK
 *
 * This script demonstrates the simplest way to create, use, and release a browser
 * session using the Crawl4AI SDK.
 *
 * A browser session gives you a persistent browser instance with a WebSocket URL
 * that you can connect to with Crawl4AI, Puppeteer, or Playwright.
 *
 * Usage:
 *     npx ts-node 01_basic_session_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function main(): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    // Step 1: Create a browser session
    console.log('Creating browser session...');
    const session = await (crawler as any).createSession({ timeout: 600 }); // 10 minute timeout

    console.log('\n=== SESSION CREATED ===');
    console.log(`Session ID: ${session.sessionId}`);
    console.log(`WebSocket URL: ${session.wsUrl}`);
    console.log(`Expires in: ${session.expiresIn} seconds`);
    console.log(`Status: ${session.status}`);

    // Step 2: Use the session (see other examples for actual usage)
    console.log('\nYou can now connect to this browser using:');
    console.log(`  - Crawl4AI: BrowserConfig(cdp_url='${session.wsUrl}')`);
    console.log(`  - Puppeteer: puppeteer.connect({ browserWSEndpoint: '${session.wsUrl}' })`);
    console.log(`  - Playwright: playwright.chromium.connectOverCDP('${session.wsUrl}')`);

    // Step 3: Get session status
    console.log('\nChecking session status...');
    const status = await (crawler as any).getSession(session.sessionId);
    console.log(`Session status: ${status.status}`);
    console.log(`Worker ID: ${status.workerId}`);

    // Step 4: Release the session
    console.log('\nReleasing session...');
    const released = await (crawler as any).releaseSession(session.sessionId);

    if (released) {
      console.log('Session released successfully!');
    } else {
      console.log('Failed to release session');
    }
  } finally {
    await crawler.close();
  }
}

main().catch(console.error);
