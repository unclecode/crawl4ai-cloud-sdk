#!/usr/bin/env python3
"""
Schema Generation with SDK

This example shows how to automatically generate CSS extraction schemas
from HTML using LLM. The schema can then be reused for fast, no-cost extraction.

Usage:
    python 03_schema_generation_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
import json
from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def generate_extraction_schema():
    """Generate CSS schema for Hacker News stories."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # First, get the HTML content
        print("Fetching Hacker News HTML...")
        result = await crawler.run(
            url="https://news.ycombinator.com",
            strategy="http"
        )

        html = result.html
        print(f"Got {len(html)} bytes of HTML")

        # Generate schema using LLM
        print("\nGenerating CSS extraction schema...")
        schema_result = await crawler.generate_schema(
            html=html,
            query="Extract all stories with their title, URL, points, and author"
        )

        if schema_result.error:
            print(f"Error: {schema_result.error}")
            return

        print("\nGenerated Schema:")
        print(json.dumps(schema_result.schema, indent=2))

        # Now use the generated schema for extraction
        print("\n\nTesting generated schema...")
        config = CrawlerRunConfig(
            extraction_strategy={
                "type": "json_css",
                "schema": schema_result.schema
            }
        )

        extract_result = await crawler.run(
            url="https://news.ycombinator.com",
            strategy="http",
            config=config
        )

        if extract_result.success and extract_result.extracted_content:
            stories = json.loads(extract_result.extracted_content)
            print(f"\nExtracted {len(stories)} stories")
            print("\nFirst 2 stories:")
            for story in stories[:2]:
                print(f"\n{story}")


if __name__ == "__main__":
    asyncio.run(generate_extraction_schema())
