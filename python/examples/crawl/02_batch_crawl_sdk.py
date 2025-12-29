#!/usr/bin/env python3
"""
Batch Crawl - SDK Example

This script demonstrates crawling multiple URLs at once.
The run_many() method automatically selects batch (<=10 URLs) or async (>10 URLs).

Usage:
    python 02_batch_crawl_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def main():
    """Crawl multiple URLs using the SDK."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # URLs to crawl
        urls = [
            "https://example.com",
            "https://httpbin.org/html",
            "https://httpbin.org/json",
        ]

        print(f"Crawling {len(urls)} URLs...")

        # Crawl multiple URLs (auto-selects batch or async based on count)
        results = await crawler.run_many(
            urls=urls,
            strategy="http",  # Options: "browser" (JS support) or "http" (faster, no JS)
            wait=True,  # Wait for all to complete
        )

        # Display results
        print(f"\n=== BATCH CRAWL COMPLETE ===")
        print(f"Total URLs: {len(results)}")
        succeeded = sum(1 for r in results if r.success)
        print(f"Succeeded: {succeeded}")
        print(f"Failed: {len(results) - succeeded}")

        # Show individual results
        for i, result in enumerate(results, 1):
            print(f"\n[{i}] {result.url}")
            print(f"    Status: {result.status_code}")
            if result.success and result.markdown:
                preview = result.markdown.raw_markdown[:100]
                print(f"    Preview: {preview}...")
            else:
                print(f"    Error: {result.error_message}")


if __name__ == "__main__":
    asyncio.run(main())
