#!/usr/bin/env npx ts-node
/**
 * Sessions - Basic Session Management with HTTP API
 *
 * This script demonstrates how to create and release browser sessions using
 * direct HTTP API calls instead of the SDK.
 *
 * Usage:
 *     npx ts-node 01_basic_session_http.ts
 *
 * Requirements:
 *     Node.js 18+ (for native fetch)
 */

// Configuration
const API_URL = 'https://api.crawl4ai.com';
const API_KEY = 'YOUR_API_KEY'; // Replace with your API key

async function main(): Promise<void> {
  const headers = {
    'X-API-Key': API_KEY,
    'Content-Type': 'application/json',
  };

  // Step 1: Create a browser session
  console.log('Creating browser session...');

  const createResponse = await fetch(`${API_URL}/v1/sessions`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ timeout: 600 }), // 10 minute timeout
  });

  if (!createResponse.ok) {
    console.log(`Error creating session: ${createResponse.status} - ${await createResponse.text()}`);
    return;
  }

  const session = await createResponse.json();

  console.log('\n=== SESSION CREATED ===');
  console.log(`Session ID: ${session.session_id}`);
  console.log(`WebSocket URL: ${session.ws_url}`);
  console.log(`Expires in: ${session.expires_in} seconds`);

  // Step 2: Use the session (see other examples for actual usage)
  console.log('\nYou can now connect to this browser using:');
  console.log(`  - Crawl4AI: BrowserConfig(cdp_url='${session.ws_url}')`);
  console.log(`  - Puppeteer: puppeteer.connect({ browserWSEndpoint: '${session.ws_url}' })`);
  console.log(`  - Playwright: playwright.chromium.connectOverCDP('${session.ws_url}')`);

  // Step 3: Get session status
  console.log('\nChecking session status...');

  const statusResponse = await fetch(`${API_URL}/v1/sessions/${session.session_id}`, {
    headers,
  });

  if (statusResponse.ok) {
    const status = await statusResponse.json();
    console.log(`Session status: ${status.status || 'N/A'}`);
    console.log(`Worker ID: ${status.worker_id || 'N/A'}`);
  }

  // Step 4: Release the session
  console.log('\nReleasing session...');

  const deleteResponse = await fetch(`${API_URL}/v1/sessions/${session.session_id}`, {
    method: 'DELETE',
    headers,
  });

  if (deleteResponse.ok) {
    console.log('Session released successfully!');
  } else {
    console.log(`Failed to release session: ${deleteResponse.status}`);
  }
}

main().catch(console.error);
