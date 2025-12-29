#!/usr/bin/env python3
"""
Sessions - Using Session with Local Crawl4AI (SDK)

This script demonstrates how to create a browser session on Crawl4AI Cloud
and then connect to it using your local Crawl4AI library.

This is useful when you want to use cloud-managed browsers but run your own
crawling logic locally with full Crawl4AI features.

Usage:
    python 02_session_with_crawl4ai_sdk.py

Requirements:
    pip install crawl4ai-cloud crawl4ai
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler as CloudCrawler
from crawl4ai import AsyncWebCrawler, BrowserConfig, CrawlerRunConfig

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def crawl_with_session(url: str):
    """
    Create a cloud session and use it with local Crawl4AI.

    Args:
        url: URL to crawl
    """
    # Step 1: Create a session on Crawl4AI Cloud
    async with CloudCrawler(api_key=API_KEY) as cloud:
        print("Creating browser session on Crawl4AI Cloud...")
        session = await cloud.create_session(timeout=600)

        print(f"\n=== SESSION CREATED ===")
        print(f"Session ID: {session.session_id}")
        print(f"WebSocket URL: {session.ws_url}")

        # Step 2: Connect to the session using local Crawl4AI
        print(f"\nConnecting to session with local Crawl4AI...")

        # Configure browser to connect to the cloud session
        browser_config = BrowserConfig(
            cdp_url=session.ws_url,
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

        # Step 3: Release the session
        print(f"\nReleasing session...")
        await cloud.release_session(session.session_id)
        print("Session released!")


async def main():
    """Run the example."""
    await crawl_with_session("https://www.example.com")


if __name__ == "__main__":
    asyncio.run(main())
