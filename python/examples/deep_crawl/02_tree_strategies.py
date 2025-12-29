#!/usr/bin/env python3
"""
Deep Crawl - Tree Traversal Strategies (BFS & DFS)

Tree strategies crawl by following links from the start URL:
- BFS (Breadth-First): Explore all pages at depth N before depth N+1
- DFS (Depth-First): Follow links deeply before backtracking

Use these when:
- Site has no sitemap
- You want to explore link structure
- You need depth-limited crawling

Usage:
    python 02_tree_strategies.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

API_KEY = "YOUR_API_KEY"


async def bfs_crawl():
    """BFS - Breadth-First Search crawl."""
    print("=== BFS Strategy (Breadth-First) ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=20,
            wait=True
        )

        print(f"Status: {result.status}")
        print(f"Pages crawled: {result.progress.completed}")

        if result.results:
            print(f"\nCrawled pages:")
            for r in result.results[:5]:
                print(f"  - {r['url']}")


async def dfs_crawl():
    """DFS - Depth-First Search crawl."""
    print("\n=== DFS Strategy (Depth-First) ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="dfs",
            max_depth=3,
            max_urls=15,
            wait=True
        )

        print(f"Status: {result.status}")
        print(f"Pages crawled: {result.progress.completed}")


async def tree_with_filters():
    """Tree crawl with URL filters."""
    print("\n=== BFS with URL Filters ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=25,
            filters={
                "patterns": ["/docs/*", "/api/*", "/guide/*"],
                "domains": {"blocked": ["twitter.com", "github.com"]}
            },
            wait=True
        )

        print(f"Filtered pages crawled: {result.progress.completed}")


async def main():
    await bfs_crawl()
    # Uncomment to run other examples:
    # await dfs_crawl()
    # await tree_with_filters()


if __name__ == "__main__":
    asyncio.run(main())
