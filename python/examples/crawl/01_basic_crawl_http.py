#!/usr/bin/env python3
"""
Basic Single URL Crawl - HTTP Example

This script demonstrates crawling a single URL using direct HTTP requests with httpx.
The endpoint is synchronous and blocks until the crawl completes.

Usage:
    python 01_basic_crawl_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key
API_URL = "https://api.crawl4ai.com"


def main():
    """Crawl a single URL using HTTP requests."""
    print("Crawling https://example.com...")

    with httpx.Client(timeout=120) as client:
        try:
            # Make the crawl request
            response = client.post(
                f"{API_URL}/v1/crawl",
                headers={
                    "X-API-Key": API_KEY,
                    "Content-Type": "application/json"
                },
                json={
                    "url": "https://example.com",
                    "strategy": "browser",  # Options: "browser" (JS support) or "http" (faster, no JS)
                }
            )

            response.raise_for_status()
            data = response.json()

            # Display results
            print(f"\n=== CRAWL COMPLETE ===")
            print(f"URL: {data['url']}")
            print(f"Title: {data['metadata'].get('title', 'N/A')}")
            print(f"Status: {data['status_code']}")
            print(f"\nMarkdown preview (first 200 chars):")
            print(data['markdown']['raw_markdown'][:200] + "...")
            print(f"\nHTML length: {len(data['html'])} characters")

        except httpx.HTTPStatusError as e:
            print(f"HTTP Error {e.response.status_code}: {e.response.text}")
        except Exception as e:
            print(f"Error: {e}")


if __name__ == "__main__":
    main()
