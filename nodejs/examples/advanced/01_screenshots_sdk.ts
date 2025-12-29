#!/usr/bin/env npx ts-node
/**
 * Screenshots and PDFs - SDK Example
 *
 * This script demonstrates how to capture screenshots and PDFs using the Crawl4AI SDK.
 * Screenshots capture the visual state of the page, while PDFs generate a print-ready document.
 *
 * Usage:
 *     npx ts-node 01_screenshots_sdk.ts
 *
 * Requirements:
 *     npm install crawl4ai-cloud
 */

import { AsyncWebCrawler } from 'crawl4ai-cloud';
import * as fs from 'fs';

// Configuration
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function captureScreenshot(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Capturing screenshot of ${url}...`);

    const result = await crawler.run(url, {
      config: {
        screenshot: true,
        wait_for: '.content', // Wait for content to load
      },
    });

    if (result.screenshot) {
      // Screenshot is returned as base64-encoded string
      console.log(`Screenshot captured: ${result.screenshot.length} bytes (base64)`);

      // Save to file
      const buffer = Buffer.from(result.screenshot, 'base64');
      fs.writeFileSync('screenshot.png', buffer);
      console.log('Screenshot saved to screenshot.png');
    } else {
      console.log('No screenshot available');
    }
  } finally {
    await crawler.close();
  }
}

async function capturePdf(url: string): Promise<void> {
  const crawler = new AsyncWebCrawler({ apiKey: API_KEY });

  try {
    console.log(`Generating PDF of ${url}...`);

    const result = await crawler.run(url, {
      config: {
        pdf: true,
      },
    });

    if (result.pdf) {
      // PDF is returned as base64-encoded string
      console.log(`PDF generated: ${result.pdf.length} bytes (base64)`);

      // Save to file
      const buffer = Buffer.from(result.pdf, 'base64');
      fs.writeFileSync('page.pdf', buffer);
      console.log('PDF saved to page.pdf');
    } else {
      console.log('No PDF available');
    }
  } finally {
    await crawler.close();
  }
}

async function main(): Promise<void> {
  // Example 1: Capture screenshot
  await captureScreenshot('https://www.example.com');

  // Example 2: Generate PDF
  await capturePdf('https://www.example.com');
}

main().catch(console.error);
