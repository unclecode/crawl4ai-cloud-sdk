#!/usr/bin/env python3
"""
Basic Single URL Crawl - SDK Example

This script demonstrates the simplest way to crawl a single URL using the SDK.
The run() method is async and returns when the crawl completes.

Usage:
    python 01_basic_crawl_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def main():
    """Crawl a single URL using the SDK."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print("Crawling https://example.com...")

        # Crawl with browser strategy (full JS support)
        result = await crawler.run(
            url="https://example.com",
            strategy="browser",  # Options: "browser" (JS support) or "http" (faster, no JS)
        )

        # Display results
        print(f"\n=== CRAWL COMPLETE ===")
        print(f"URL: {result.url}")
        print(f"Success: {result.success}")
        print(f"Status: {result.status_code}")
        print(f"\nMarkdown preview (first 200 chars):")
        print(result.markdown.raw_markdown[:200] + "...")
        print(f"\nHTML length: {len(result.html)} characters")


if __name__ == "__main__":
    asyncio.run(main())
