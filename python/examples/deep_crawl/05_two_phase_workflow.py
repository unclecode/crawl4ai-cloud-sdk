#!/usr/bin/env python3
"""
Deep Crawl - Two-Phase Workflow (Scan → Extract)

The two-phase workflow separates URL discovery from data extraction:

Phase 1 (Scan):
  - Discover URLs using BFS/DFS/Best-First
  - Cache raw HTML in Redis (30-minute TTL)
  - Return scan_job_id for later use

Phase 2 (Extract):
  - Use source_job_id to reference cached HTML
  - Apply extraction strategy (CSS, LLM, etc.)
  - No re-crawling needed - uses cached HTML

Benefits:
  - Scan once, extract multiple times
  - Apply different extraction strategies to same data
  - Preview URLs before committing to full extraction
  - Faster extraction (HTML already cached)

Usage:
    python 05_two_phase_workflow.py
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler
import json

API_KEY = "YOUR_API_KEY"


def basic_two_phase():
    """Simple scan then extract workflow."""
    print("=== Two-Phase Workflow ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # ========== PHASE 1: SCAN ==========
        print("Phase 1: Scanning (URL discovery + HTML caching)...")

        scan_result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=10,
            scan_only=True,  # Don't extract yet
            wait=True
        )

        print(f"  Scan Job ID: {scan_result.job_id}")
        print(f"  URLs discovered: {scan_result.discovered_count}")
        print(f"  Cache expires: {scan_result.cache_expires_at}")
        print()

        # ========== PHASE 2: EXTRACT ==========
        print("Phase 2: Extracting from cached HTML...")

        job = await crawler.deep_crawl(
            source_job_id=scan_result.job_id,  # Use cached HTML
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": {
                        "name": "PageContent",
                        "baseSelector": "main, article, .content",
                        "fields": [
                            {"name": "title", "selector": "h1", "type": "text"},
                            {"name": "headings", "selector": "h2, h3", "type": "list"}
                        ]
                    }
                }
            },
            wait=True
        )

        print(f"  Job ID: {job.job_id}")
        print(f"  Pages extracted: {job.progress.completed}")

        if job.results:
            print("\nExtracted content:")
            for r in job.results[:3]:
                print(f"  URL: {r['url']}")
                if r.get('extracted_content'):
                    try:
                        data = json.loads(r['extracted_content'])
                        print(f"    Title: {data.get('title', 'N/A')}")
                    except:
                        pass

    finally:
        


def multiple_extractions():
    """Extract same content with different strategies."""
    print("\n=== Multiple Extractions from Same Scan ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # SCAN ONCE
        print("Scanning...")
        scan_result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=1,
            max_urls=5,
            scan_only=True,
            wait=True
        )

        scan_job_id = scan_result.job_id
        print(f"Scan Job ID: {scan_job_id}")
        print(f"URLs cached: {scan_result.discovered_count}\n")

        # EXTRACT #1: Titles only
        print("Extraction 1: Titles...")
        job1 = await crawler.deep_crawl(
            source_job_id=scan_job_id,
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": {
                        "name": "Titles",
                        "baseSelector": "body",
                        "fields": [
                            {"name": "title", "selector": "h1", "type": "text"}
                        ]
                    }
                }
            },
            wait=True
        )
        print(f"  Extracted: {job1.progress.completed} pages")

        # EXTRACT #2: Links
        print("Extraction 2: Links...")
        job2 = await crawler.deep_crawl(
            source_job_id=scan_job_id,
            crawler_config={
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": {
                        "name": "Links",
                        "baseSelector": "body",
                        "fields": [
                            {"name": "links", "selector": "a[href]", "type": "list", "attribute": "href"}
                        ]
                    }
                }
            },
            wait=True
        )
        print(f"  Extracted: {job2.progress.completed} pages")

        # EXTRACT #3: Full markdown
        print("Extraction 3: Markdown...")
        job3 = await crawler.deep_crawl(
            source_job_id=scan_job_id,
            # No extraction strategy = get markdown
            wait=True
        )
        print(f"  Extracted: {job3.progress.completed} pages")

        print("\nAll 3 extractions used the same cached HTML!")

    finally:
        


def partial_extraction():
    """Extract only specific URLs from a scan."""
    print("\n=== Partial Extraction (Select URLs) ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # Scan a larger set
        print("Scanning 20 URLs...")
        scan_result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=2,
            max_urls=20,
            scan_only=True,
            wait=True
        )

        print(f"Found {scan_result.discovered_count} URLs")

        # Filter to only API-related URLs
        if scan_result.urls:
            api_urls = [
                u.url for u in scan_result.urls
                if '/api/' in u.url or 'reference' in u.url
            ]
            print(f"API URLs: {len(api_urls)}")

            if api_urls:
                # Extract only the filtered URLs
                # Note: The extract phase will use cached HTML
                # for URLs that were in the original scan
                print("\nExtracting API pages only...")
                job = await crawler.deep_crawl(
                    source_job_id=scan_result.job_id,
                    # Future: url_filter parameter for selective extraction
                    wait=True
                )
                print(f"Extracted: {job.progress.completed} pages")

    finally:
        


def handle_cache_expiry():
    """Handle cache expiration gracefully."""
    print("\n=== Handling Cache Expiry ===\n")

    async with AsyncWebCrawler(api_key=API_KEY) as crawler:

    try:
        # Get scan result
        scan_result = await crawler.deep_crawl(
            url="https://docs.crawl4ai.com",
            strategy="bfs",
            max_depth=1,
            max_urls=5,
            scan_only=True,
            wait=True
        )

        print(f"Scan Job ID: {scan_result.job_id}")
        print(f"Cache expires: {scan_result.cache_expires_at}")
        print()

        # Check if cache is still valid before extracting
        # In production, parse cache_expires_at and compare to now

        # If cache expired, you'll get a 410 Gone error
        # Handle it by re-running the scan
        try:
            job = await crawler.deep_crawl(
                source_job_id=scan_result.job_id,
                wait=True
            )
            print(f"Extracted: {job.progress.completed} pages")
        except Exception as e:
            if "expired" in str(e).lower() or "410" in str(e):
                print("Cache expired! Re-running scan...")
                # Re-scan and extract in one go (scan_only=False)
                job = await crawler.deep_crawl(
                    url="https://docs.crawl4ai.com",
                    strategy="bfs",
                    max_depth=1,
                    max_urls=5,
                    wait=True
                )
                print(f"Re-crawled: {job.progress.completed} pages")
            else:
                raise

    finally:
        


if __name__ == "__main__":
    basic_two_phase()
    # Uncomment to run other examples:
    # multiple_extractions()
    # partial_extraction()
    # handle_cache_expiry()
