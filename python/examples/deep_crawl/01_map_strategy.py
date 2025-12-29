#!/usr/bin/env python3
"""
Deep Crawl - Map Strategy (Sitemap Discovery)

The "map" strategy discovers URLs from a website's sitemap and crawls them.
This is the fastest way to crawl a site with a well-structured sitemap.

Features:
- Automatic sitemap discovery (sitemap.xml, sitemap_index.xml)
- URL pattern filtering
- Optional Common Crawl fallback

Usage:
    python 01_map_strategy.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

API_KEY = "YOUR_API_KEY"


async def basic_map_crawl():
    """Simplest map strategy - discover from sitemap and crawl."""
    print("=== Basic Map Strategy ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # wait=True blocks until all URLs are crawled
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",  # Default strategy
            max_urls=5,
            wait=True
        )

        print(f"Status: {result.status}")
        print(f"URLs crawled: {result.progress.completed}/{result.progress.total}")

        if result.results:
            print(f"\nResults ({len(result.results)} pages):")
            for r in result.results:
                print(f"  - {r['url']}: {r['success']}")


async def map_with_pattern():
    """Filter URLs by glob pattern."""
    print("\n=== Map Strategy with Pattern Filter ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",
            pattern="*/api/*",  # Only URLs containing /api/
            max_urls=10,
            filter_nonsense_urls=True,  # Skip auth/tracking URLs
            wait=True
        )

        print(f"Matched URLs: {result.progress.total}")
        print(f"Successfully crawled: {result.progress.completed}")


async def map_with_common_crawl():
    """Use Common Crawl as additional/fallback URL source."""
    print("\n=== Map Strategy with Common Crawl ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",
            source="sitemap+cc",  # Try sitemap first, then Common Crawl
            max_urls=10,
            wait=True
        )

        print(f"URLs discovered: {result.progress.total}")
        print(f"Source: sitemap + Common Crawl index")


async def map_no_wait():
    """Start crawl without waiting - poll status manually."""
    print("\n=== Map Strategy (No Wait) ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # wait=False returns immediately with scan job info
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",
            max_urls=5,
            wait=False  # Don't wait
        )

        print(f"Scan Job ID: {result.job_id}")
        print(f"Status: {result.status}")  # Will be "pending" initially
        print(f"Discovered so far: {result.discovered_count}")


async def main():
    await basic_map_crawl()
    # Uncomment to run other examples:
    # await map_with_pattern()
    # await map_with_common_crawl()
    # await map_no_wait()


if __name__ == "__main__":
    asyncio.run(main())
