#!/usr/bin/env python3
"""
Deep Crawl - Scan-Only Mode (URL Discovery)

Scan-only mode discovers URLs without crawling them. This is useful for:
- Previewing what URLs will be crawled before committing
- Building a URL list for later processing
- Analyzing site structure
- Fast HTML caching for later extraction

The scan phase:
1. Discovers URLs (sitemap, links, Common Crawl)
2. Caches HTML content (30 min TTL)
3. Returns URL list with metadata (depth, links_found, html_size)

Usage:
    python 04_scan_only_discovery.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler

API_KEY = "YOUR_API_KEY"


def basic_scan_only():
    """Discover URLs without crawling."""
    print("=== Scan-Only Mode (Discovery) ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # scan_only=True returns discovered URLs without processing
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=20,
            scan_only=True,  # Just discover, don't crawl
            wait=True
        )

        # Result is a DeepCrawlResult (not a Job)
        print(f"Status: {result.status}")
        print(f"URLs discovered: {result.discovered_count}")
        print(f"Cache expires at: {result.cache_expires_at}")

        # Get list of discovered URLs
        if result.discovered_urls:
            print(f"\nDiscovered URLs:")
            for url in result.discovered_urls[:10]:
                print(f"  - {url}")

    finally:
        


def scan_with_url_details():
    """Get detailed info about each discovered URL."""
    print("\n=== Scan with URL Details ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=15,
            scan_only=True,
            wait=True
        )

        print(f"Discovered {result.discovered_count} URLs\n")

        # urls contains ScanUrlInfo objects with metadata
        if result.urls:
            print("URL Details:")
            print("-" * 70)
            for url_info in result.urls[:10]:
                print(f"URL: {url_info.url}")
                print(f"  Depth: {url_info.depth}")
                print(f"  Links found: {url_info.links_found}")
                print(f"  HTML size: {url_info.html_size:,} bytes")
                if url_info.score is not None:
                    print(f"  Score: {url_info.score:.2f}")
                print()

    finally:
        


def scan_map_strategy():
    """Scan-only with map (sitemap) strategy."""
    print("\n=== Scan-Only with Map Strategy ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="map",
            max_urls=50,
            scan_only=True,
            wait=True
        )

        print(f"Sitemap URLs found: {result.discovered_count}")

        # You can filter/review URLs before deciding to crawl
        if result.discovered_urls:
            api_urls = [u for u in result.discovered_urls if '/api/' in u]
            guide_urls = [u for u in result.discovered_urls if '/guide/' in u]

            print(f"\nAPI pages: {len(api_urls)}")
            print(f"Guide pages: {len(guide_urls)}")

            # Save the scan job ID for later extraction
            print(f"\nScan Job ID: {result.job_id}")
            print("Use this ID with source_job_id to extract later!")

    finally:
        


def scan_with_html_download():
    """Scan and get downloadable HTML archive."""
    print("\n=== Scan with HTML Download ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=1,
            max_urls=10,
            scan_only=True,
            include_html=True,  # Generate downloadable HTML ZIP
            wait=True
        )

        print(f"Discovered: {result.discovered_count} URLs")

        if result.html_download_url:
            print(f"\nHTML Archive URL: {result.html_download_url}")
            print("(Download within 30 minutes before cache expires)")

    finally:
        


def scan_then_decide():
    """Preview URLs, then decide whether to crawl."""
    print("\n=== Scan Then Decide Workflow ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # Step 1: Discover URLs
        print("Step 1: Discovering URLs...")
        scan_result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=30,
            scan_only=True,
            wait=True
        )

        print(f"Found {scan_result.discovered_count} URLs")
        print(f"Scan Job ID: {scan_result.job_id}")

        # Step 2: Review and decide
        total_size = sum(u.html_size for u in scan_result.urls) if scan_result.urls else 0
        print(f"Total HTML size: {total_size:,} bytes")

        # You could prompt user here or apply business logic
        should_crawl = scan_result.discovered_count <= 20

        if should_crawl:
            # Step 3: Extract from cached HTML
            print("\nStep 2: Extracting from cache...")
            job = await crawler.deep_crawl(
                source_job_id=scan_result.job_id,  # Use cached HTML
                wait=True
            )
            print(f"Extracted {job.progress.completed} pages")
        else:
            print("\nToo many URLs, skipping extraction.")
            print("You can still use source_job_id within 30 minutes.")

    finally:
        


if __name__ == "__main__":
    basic_scan_only()
    # Uncomment to run other examples:
    # scan_with_url_details()
    # scan_map_strategy()
    # scan_with_html_download()
    # scan_then_decide()
