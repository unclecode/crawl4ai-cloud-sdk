#!/usr/bin/env python3
"""
Deep Crawl - Best-First Strategy with Scoring

Best-first crawling prioritizes URLs based on relevance scores.
Uses a priority queue to always crawl the highest-scoring URL next.

Scorers available:
- keywords: Score based on keyword presence in URL/title
- optimal_depth: Prefer URLs at a specific depth level

Usage:
    python 03_best_first_scoring.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

API_KEY = "YOUR_API_KEY"


async def best_first_with_keywords():
    """Score URLs by keyword relevance."""
    print("=== Best-First with Keyword Scoring ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="best_first",
            max_depth=3,
            max_urls=15,
            scorers={
                "keywords": ["api", "tutorial", "guide", "example"],
            },
            wait=True
        )

        print(f"Pages crawled: {result.progress.completed}")

        if result.results:
            print("\nTop results (by score):")
            for i, r in enumerate(result.results[:5], 1):
                print(f"  {i}. {r['url']}")


async def best_first_for_documentation():
    """Find API documentation pages efficiently."""
    print("\n=== Best-First for API Docs ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="best_first",
            max_depth=3,
            max_urls=30,
            scorers={
                "keywords": ["api", "reference", "method", "function", "parameter"],
                "optimal_depth": 2,
                "weights": {"keywords": 3.0, "depth": 1.0}
            },
            filters={"patterns": ["/api/*", "/reference/*", "/docs/*"]},
            wait=True
        )

        print(f"API docs found: {result.progress.completed}")


async def main():
    await best_first_with_keywords()
    # await best_first_for_documentation()


if __name__ == "__main__":
    asyncio.run(main())
