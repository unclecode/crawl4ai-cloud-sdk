#!/usr/bin/env python3
"""
CSS Extraction with HTTP - No LLM Cost

This example shows how to extract structured data using CSS selectors
via direct HTTP API calls (no SDK).

Usage:
    python 01_css_extraction_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key

def extract_with_css():
    """Extract Hacker News stories using CSS selectors via HTTP API."""

    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    # Define CSS extraction schema
    payload = {
        "url": "https://news.ycombinator.com",
        "strategy": "http",  # Fast, no browser needed
        "crawler_config": {
            "extraction_strategy": {
                "type": "json_css",
                "schema": {
                    "name": "HackerNewsStories",
                    "baseSelector": ".athing",
                    "fields": [
                        {"name": "title", "selector": ".titleline > a", "type": "text"},
                        {"name": "url", "selector": ".titleline > a", "type": "attribute", "attribute": "href"},
                        {"name": "points", "selector": "+ tr .score", "type": "text"},
                        {"name": "author", "selector": "+ tr .hnuser", "type": "text"}
                    ]
                }
            }
        }
    }

    print("Crawling Hacker News with CSS extraction...")
    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json=payload,
        timeout=60.0
    )

    if response.status_code != 200:
        print(f"Error: {response.status_code} - {response.text}")
        return

    result = response.json()
    stories = result.get("extracted_content", [])

    print(f"\nExtracted {len(stories)} stories")
    print("\nFirst 3 stories:")
    for story in stories[:3]:
        print(f"\nTitle: {story.get('title')}")
        print(f"URL: {story.get('url')}")
        print(f"Points: {story.get('points')}")
        print(f"Author: {story.get('author')}")

if __name__ == "__main__":
    extract_with_css()
