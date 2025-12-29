#!/usr/bin/env npx ts-node
/**
 * Screenshots and PDFs - HTTP Example
 *
 * This script demonstrates how to capture screenshots and PDFs using raw HTTP requests.
 * This is useful if you prefer direct API access without the SDK.
 *
 * Usage:
 *     npx ts-node 01_screenshots_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

import * as fs from 'fs';

// Configuration
const API_URL = 'https://api.crawl4ai.com';
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function captureScreenshot(url: string): Promise<string | null> {
  console.log(`Capturing screenshot of ${url}...`);

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      url,
      crawler_config: {
        screenshot: true,
        screenshot_wait_for: '.content',
      },
    }),
  });

  if (!response.ok) {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return null;
  }

  const data = await response.json();

  if (data.screenshot) {
    console.log(`Screenshot captured: ${data.screenshot.length} bytes (base64)`);
    return data.screenshot;
  } else {
    console.log('No screenshot available');
    return null;
  }
}

async function capturePdf(url: string): Promise<string | null> {
  console.log(`Generating PDF of ${url}...`);

  const response = await fetch(`${API_URL}/v1/crawl`, {
    method: 'POST',
    headers: {
      'X-API-Key': API_KEY,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      url,
      crawler_config: {
        pdf: true,
      },
    }),
  });

  if (!response.ok) {
    console.log(`Error: ${response.status} - ${await response.text()}`);
    return null;
  }

  const data = await response.json();

  if (data.pdf) {
    console.log(`PDF generated: ${data.pdf.length} bytes (base64)`);
    return data.pdf;
  } else {
    console.log('No PDF available');
    return null;
  }
}

async function main(): Promise<void> {
  // Example 1: Capture screenshot
  const screenshotB64 = await captureScreenshot('https://www.example.com');

  // Example 2: Generate PDF
  const pdfB64 = await capturePdf('https://www.example.com');

  // Save screenshot to file
  if (screenshotB64) {
    const buffer = Buffer.from(screenshotB64, 'base64');
    fs.writeFileSync('screenshot.png', buffer);
    console.log('\nScreenshot saved to screenshot.png');
  }

  // Save PDF to file
  if (pdfB64) {
    const buffer = Buffer.from(pdfB64, 'base64');
    fs.writeFileSync('page.pdf', buffer);
    console.log('PDF saved to page.pdf');
  }
}

main().catch(console.error);
