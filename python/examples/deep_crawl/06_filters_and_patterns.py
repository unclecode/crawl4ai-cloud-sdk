#!/usr/bin/env python3
"""
Deep Crawl - URL Filtering and Patterns

Control which URLs are crawled using filters and patterns:
- Glob patterns: Match URL paths (e.g., "*/docs/*")
- Domain filters: Allow/block specific domains
- Nonsense URL filter: Skip auth, tracking, utility pages

Filtering happens during discovery, before crawling,
so you don't waste resources on unwanted pages.

Usage:
    python 06_filters_and_patterns.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

API_KEY = "YOUR_API_KEY"


def basic_pattern_filter():
    """Filter URLs by glob pattern."""
    print("=== Basic Pattern Filter ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=20,
            pattern="*/docs/*",  # Only URLs with /docs/ in path
            wait=True
        )

        print(f"Matched URLs: {job.progress.total}")
        print(f"Crawled: {job.progress.completed}")

        if job.results:
            print("\nCrawled URLs:")
            for r in job.results[:5]:
                print(f"  - {r['url']}")

    finally:
        


def multiple_patterns():
    """Match multiple URL patterns."""
    print("\n=== Multiple Patterns ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=30,
            filters={
                # Match any of these patterns
                "patterns": [
                    "/api/*",       # API reference pages
                    "/guide/*",     # User guides
                    "/tutorial/*",  # Tutorials
                    "*/example*",   # Example pages
                ]
            },
            wait=True
        )

        print(f"Matching URLs: {job.progress.total}")

    finally:
        


def exclude_patterns():
    """Exclude URLs matching certain patterns."""
    print("\n=== Exclude Patterns ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=30,
            filters={
                # Exclude these patterns
                "exclude_patterns": [
                    "*/changelog/*",   # Skip changelogs
                    "*/archive/*",     # Skip archives
                    "*?page=*",        # Skip pagination
                    "*#*",             # Skip anchor links
                ]
            },
            wait=True
        )

        print(f"Filtered URLs: {job.progress.total}")

    finally:
        


def domain_filtering():
    """Control cross-domain link following."""
    print("\n=== Domain Filtering ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=25,
            filters={
                "domains": {
                    # Never follow links to these domains
                    "blocked": [
                        "twitter.com",
                        "facebook.com",
                        "linkedin.com",
                        "github.com",  # If you want to stay on docs
                    ],
                    # Or whitelist: only follow links to these
                    # "allowed": ["docs.crawl4ai.com", "crawl4ai.com"]
                }
            },
            wait=True
        )

        print(f"Crawled (blocked external): {job.progress.total}")

    finally:
        


def filter_nonsense_urls():
    """Filter common non-content URLs."""
    print("\n=== Filter Nonsense URLs ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # filter_nonsense_urls removes:
        # - Login/auth pages
        # - Tracking URLs
        # - Print/email versions
        # - RSS/Atom feeds
        # - Search result pages
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=30,
            filter_nonsense_urls=True,  # Enable automatic filtering
            wait=True
        )

        print(f"Content URLs: {job.progress.total}")

    finally:
        


def combined_filters():
    """Combine multiple filter types."""
    print("\n=== Combined Filters ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=3,
            max_urls=50,
            filter_nonsense_urls=True,
            filters={
                # Include patterns (whitelist)
                "patterns": [
                    "/docs/*",
                    "/api/*",
                    "/guide/*"
                ],
                # Exclude patterns
                "exclude_patterns": [
                    "*changelog*",
                    "*version*"
                ],
                # Domain controls
                "domains": {
                    "blocked": ["twitter.com", "github.com"]
                }
            },
            wait=True
        )

        print(f"Filtered & crawled: {job.progress.total}")

    finally:
        


def filter_for_ecommerce():
    """Example: Filter for e-commerce product pages."""
    print("\n=== E-commerce Product Filter ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://example-shop.com",
            strategy="bfs",
            max_depth=3,
            max_urls=100,
            filter_nonsense_urls=True,
            filters={
                "patterns": [
                    "/product/*",
                    "/products/*",
                    "/item/*",
                    "/shop/*"
                ],
                "exclude_patterns": [
                    "*/cart*",
                    "*/checkout*",
                    "*/login*",
                    "*/account*",
                    "*?sort=*",      # Skip sorted views
                    "*?filter=*",    # Skip filtered views
                ]
            },
            wait=True
        )

        print(f"Product pages found: {job.progress.total}")

    finally:
        


def filter_for_blog():
    """Example: Filter for blog posts only."""
    print("\n=== Blog Post Filter ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        job = await crawler.deep_crawl(
            url="https://example-blog.com",
            strategy="bfs",
            max_depth=2,
            max_urls=50,
            filter_nonsense_urls=True,
            filters={
                "patterns": [
                    "/blog/*",
                    "/posts/*",
                    "/articles/*",
                    "*/20[0-9][0-9]/*"  # Year-based URLs
                ],
                "exclude_patterns": [
                    "*/tag/*",
                    "*/category/*",
                    "*/author/*",
                    "*/page/*"
                ]
            },
            wait=True
        )

        print(f"Blog posts found: {job.progress.total}")

    finally:
        


if __name__ == "__main__":
    basic_pattern_filter()
    # Uncomment to run other examples:
    # multiple_patterns()
    # exclude_patterns()
    # domain_filtering()
    # filter_nonsense_urls()
    # combined_filters()
    # filter_for_ecommerce()
    # filter_for_blog()
