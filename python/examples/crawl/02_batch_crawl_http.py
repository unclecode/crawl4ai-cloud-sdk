#!/usr/bin/env python3
"""
Batch Crawl - HTTP Example

This script demonstrates crawling multiple URLs (up to 10) using direct HTTP requests.
The batch endpoint processes URLs sequentially and returns all results.

Usage:
    python 02_batch_crawl_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key
API_URL = "https://api.crawl4ai.com"


def main():
    """Crawl multiple URLs using HTTP batch endpoint."""
    # URLs to crawl (max 10)
    urls = [
        "https://example.com",
        "https://httpbin.org/html",
        "https://httpbin.org/json",
    ]

    print(f"Crawling {len(urls)} URLs in batch...")

    with httpx.Client(timeout=120) as client:
        try:
            # Make the batch crawl request
            response = client.post(
                f"{API_URL}/v1/crawl/batch",
                headers={
                    "X-API-Key": API_KEY,
                    "Content-Type": "application/json"
                },
                json={
                    "urls": urls,
                    "strategy": "http",  # Options: "browser" (JS support) or "http" (faster, no JS)
                }
            )

            response.raise_for_status()
            data = response.json()

            # Display results
            print(f"\n=== BATCH CRAWL COMPLETE ===")
            print(f"Total URLs: {len(data['results'])}")
            print(f"Succeeded: {data['succeeded']}")
            print(f"Failed: {data['failed']}")

            # Show individual results
            for i, result in enumerate(data['results'], 1):
                print(f"\n[{i}] {result['url']}")
                print(f"    Status: {result['status_code']}")
                if result.get('markdown'):
                    preview = result['markdown']['raw_markdown'][:100]
                    print(f"    Preview: {preview}...")
                else:
                    print(f"    Error: {result.get('error_message', 'Unknown error')}")

        except httpx.HTTPStatusError as e:
            print(f"HTTP Error {e.response.status_code}: {e.response.text}")
        except Exception as e:
            print(f"Error: {e}")


if __name__ == "__main__":
    main()
