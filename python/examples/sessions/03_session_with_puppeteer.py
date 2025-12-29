#!/usr/bin/env python3
"""
Sessions - Using Session with Puppeteer

This script demonstrates how to create a browser session on Crawl4AI Cloud
and connect to it using Puppeteer (Node.js).

The Python script manages the session lifecycle, while Puppeteer connects to
the browser for advanced automation.

Usage:
    1. Run this Python script to create a session
    2. Use the WebSocket URL in your Puppeteer script (see example below)

Requirements:
    pip install crawl4ai-cloud
"""

from crawl4ai_cloud import Crawl4AI
import time

# Configuration
API_KEY = "your_api_key_here"  # Replace with your API key


def main():
    """
    Create a session and show how to use it with Puppeteer.
    """

    client = Crawl4AI(API_KEY)

    try:
        # Step 1: Create a browser session
        print("Creating browser session on Crawl4AI Cloud...")
        session = client.create_session(timeout=3600)  # 1 hour timeout

        print(f"\n=== SESSION CREATED ===")
        print(f"Session ID: {session.session_id}")
        print(f"WebSocket URL: {session.ws_url}")

        # Step 2: Show Puppeteer connection code
        print(f"\n=== PUPPETEER EXAMPLE ===")
        print("Use this JavaScript/TypeScript code to connect with Puppeteer:\n")

        puppeteer_code = f"""
// Puppeteer Example (Node.js)
// npm install puppeteer-core

const puppeteer = require('puppeteer-core');

async function crawlWithSession() {{
    // Connect to the cloud browser
    const browser = await puppeteer.connect({{
        browserWSEndpoint: '{session.ws_url}'
    }});

    try {{
        const page = await browser.newPage();

        // Navigate and interact
        await page.goto('https://www.example.com', {{
            waitUntil: 'networkidle0'
        }});

        // Extract content
        const title = await page.title();
        const content = await page.content();

        console.log('Title:', title);
        console.log('Content length:', content.length);

        // Take screenshot
        await page.screenshot({{ path: 'screenshot.png' }});

        // Extract data
        const data = await page.evaluate(() => {{
            return {{
                title: document.title,
                headings: Array.from(document.querySelectorAll('h1, h2'))
                    .map(h => h.textContent)
            }};
        }});

        console.log('Extracted data:', data);

    }} finally {{
        // Don't call browser.close() - the session is managed by Crawl4AI Cloud
        await browser.disconnect();
    }}
}}

crawlWithSession().catch(console.error);
"""

        print(puppeteer_code)

        # Step 3: Keep session alive for testing
        print("\n=== SESSION ACTIVE ===")
        print("Session will remain active for 60 seconds for testing.")
        print("Press Ctrl+C to release immediately.\n")

        try:
            # Check session status every 10 seconds
            for i in range(6):
                status = client.get_session(session.session_id)
                print(f"[{i*10}s] Session status: {status.status} | Worker: {status.worker_id}")
                time.sleep(10)

        except KeyboardInterrupt:
            print("\nInterrupted by user")

        # Step 4: Release the session
        print(f"\nReleasing session...")
        client.release_session(session.session_id)
        print("Session released!")

    finally:
        client.close()


if __name__ == "__main__":
    main()


# Alternative: Playwright Example
"""
// Playwright Example (Node.js)
// npm install playwright-core

const { chromium } = require('playwright-core');

async function crawlWithSession() {
    const browser = await chromium.connectOverCDP('{WS_URL}');

    try {
        const context = browser.contexts()[0];
        const page = await context.newPage();

        await page.goto('https://www.example.com');

        const title = await page.title();
        console.log('Title:', title);

        await page.screenshot({ path: 'screenshot.png' });

    } finally {
        await browser.close();
    }
}

crawlWithSession().catch(console.error);
"""
