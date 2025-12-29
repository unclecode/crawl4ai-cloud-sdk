#!/usr/bin/env python3
"""
Sessions - Basic Session Management with SDK

This script demonstrates the simplest way to create, use, and release a browser
session using the Crawl4AI SDK.

A browser session gives you a persistent browser instance with a WebSocket URL
that you can connect to with Crawl4AI, Puppeteer, or Playwright.

Usage:
    python 01_basic_session_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def main():
    """Create a session, print its details, and release it."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # Step 1: Create a browser session
        print("Creating browser session...")
        session = await crawler.create_session(timeout=600)  # 10 minute timeout

        print(f"\n=== SESSION CREATED ===")
        print(f"Session ID: {session.session_id}")
        print(f"WebSocket URL: {session.ws_url}")
        print(f"Expires in: {session.expires_in} seconds")
        print(f"Status: {session.status}")

        # Step 2: Use the session (see other examples for actual usage)
        print(f"\nYou can now connect to this browser using:")
        print(f"  - Crawl4AI: BrowserConfig(cdp_url='{session.ws_url}')")
        print(f"  - Puppeteer: puppeteer.connect({{ browserWSEndpoint: '{session.ws_url}' }})")
        print(f"  - Playwright: playwright.chromium.connectOverCDP('{session.ws_url}')")

        # Step 3: Get session status
        print(f"\nChecking session status...")
        status = await crawler.get_session(session.session_id)
        print(f"Session status: {status.status}")
        print(f"Worker ID: {status.worker_id}")

        # Step 4: Release the session
        print(f"\nReleasing session...")
        released = await crawler.release_session(session.session_id)

        if released:
            print("Session released successfully!")
        else:
            print("Failed to release session")


if __name__ == "__main__":
    asyncio.run(main())
