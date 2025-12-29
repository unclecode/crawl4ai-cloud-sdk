#!/usr/bin/env python3
"""
Async Crawl with Wait - SDK Example

This script demonstrates async crawling with automatic polling (wait=True).
The SDK automatically polls until the job completes and returns the results.

Usage:
    python 03_async_crawl_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def main():
    """Create an async crawl job and wait for completion."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # URLs to crawl (can be more than 10 for async)
        urls = [
            "https://example.com",
            "https://httpbin.org/html",
            "https://httpbin.org/json",
            "https://httpbin.org/robots.txt",
        ]

        print(f"Creating async job for {len(urls)} URLs...")

        # run_many with wait=True handles polling automatically
        results = await crawler.run_many(
            urls=urls,
            strategy="http",  # Options: "browser" (JS support) or "http" (faster, no JS)
            wait=True,  # Wait for completion (SDK polls automatically)
        )

        # Display results
        print(f"\n=== JOB COMPLETE ===")
        print(f"Total: {len(results)}")
        succeeded = sum(1 for r in results if r.success)
        print(f"Succeeded: {succeeded}")

        # Show sample results
        print(f"\nSample Results (first 3):")
        for i, result in enumerate(results[:3], 1):
            print(f"[{i}] {result.url}")
            print(f"    Status: {result.status_code}")
            if result.success and result.markdown:
                preview = result.markdown.raw_markdown[:80]
                print(f"    Preview: {preview}...")


if __name__ == "__main__":
    asyncio.run(main())
