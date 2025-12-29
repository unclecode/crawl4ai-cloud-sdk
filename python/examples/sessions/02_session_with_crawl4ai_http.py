#!/usr/bin/env python3
"""
Sessions - Using Session with Local Crawl4AI (HTTP API)

This script demonstrates how to create a browser session using the HTTP API
and then connect to it using your local Crawl4AI library.

Usage:
    python 02_session_with_crawl4ai_http.py

Requirements:
    pip install httpx crawl4ai
"""

import asyncio
import httpx
from crawl4ai import AsyncWebCrawler, BrowserConfig, CrawlerRunConfig

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key


async def crawl_with_session(url: str):
    """
    Create a cloud session via HTTP and use it with local Crawl4AI.

    Args:
        url: URL to crawl
    """

    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    # Step 1: Create a session using HTTP API
    print("Creating browser session on Crawl4AI Cloud...")

    response = httpx.post(
        f"{API_URL}/v1/sessions",
        headers=headers,
        json={"timeout": 600},
        timeout=30.0
    )

    if response.status_code != 200:
        print(f"Error creating session: {response.status_code} - {response.text}")
        return

    session = response.json()

    print(f"\n=== SESSION CREATED ===")
    print(f"Session ID: {session['session_id']}")
    print(f"WebSocket URL: {session['ws_url']}")

    try:
        # Step 2: Connect to the session using local Crawl4AI
        print(f"\nConnecting to session with local Crawl4AI...")

        # Configure browser to connect to the cloud session
        browser_config = BrowserConfig(
            cdp_url=session['ws_url'],
            headless=True  # Already running in cloud
        )

        # Create crawler with the cloud browser
        async with AsyncWebCrawler(config=browser_config) as crawler:
            print(f"Crawling {url}...")

            # Run the crawl using cloud browser
            result = await crawler.arun(
                url=url,
                config=CrawlerRunConfig(
                    word_count_threshold=10,
                    remove_overlay_elements=True
                )
            )

            # Process results
            print(f"\n=== CRAWL RESULTS ===")
            print(f"URL: {result.url}")
            print(f"Success: {result.success}")
            print(f"Status Code: {result.status_code}")
            print(f"Markdown length: {len(result.markdown_v2.raw_markdown) if result.markdown_v2 else 0} characters")

            # Show first 200 characters
            if result.markdown_v2 and result.markdown_v2.raw_markdown:
                print(f"\nContent preview:")
                print(result.markdown_v2.raw_markdown[:200])
                print("...")

    finally:
        # Step 3: Release the session using HTTP API
        print(f"\nReleasing session...")

        delete_response = httpx.delete(
            f"{API_URL}/v1/sessions/{session['session_id']}",
            headers=headers,
            timeout=30.0
        )

        if delete_response.status_code == 200:
            print("Session released!")
        else:
            print(f"Warning: Failed to release session: {delete_response.status_code}")


async def main():
    """Run the example."""
    await crawl_with_session("https://www.example.com")


if __name__ == "__main__":
    asyncio.run(main())
