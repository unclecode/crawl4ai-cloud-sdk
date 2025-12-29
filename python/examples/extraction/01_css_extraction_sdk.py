#!/usr/bin/env python3
"""
CSS Extraction with SDK - No LLM Cost

This example shows how to extract structured data using CSS selectors.
CSS extraction is fast, reliable, and has no LLM cost.

Usage:
    python 01_css_extraction_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
import json
from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def extract_with_css():
    """Extract Hacker News stories using CSS selectors."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # Define CSS extraction schema
        config = CrawlerRunConfig(
            extraction_strategy={
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
        )

        print("Crawling Hacker News with CSS extraction...")
        result = await crawler.run(
            url="https://news.ycombinator.com",
            strategy="http",  # Fast, no browser needed
            config=config
        )

        if result.success and result.extracted_content:
            stories = json.loads(result.extracted_content)
            print(f"\nExtracted {len(stories)} stories")
            print("\nFirst 3 stories:")
            for story in stories[:3]:
                print(f"\nTitle: {story.get('title')}")
                print(f"URL: {story.get('url')}")
                print(f"Points: {story.get('points')}")
                print(f"Author: {story.get('author')}")
        else:
            print(f"Error: {result.error_message}")


if __name__ == "__main__":
    asyncio.run(extract_with_css())
