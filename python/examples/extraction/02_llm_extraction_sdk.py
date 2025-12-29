#!/usr/bin/env python3
"""
LLM Extraction with SDK

This example shows how to extract structured data using LLM with natural language instructions.
LLM extraction is flexible and can handle complex extraction needs.

Usage:
    python 02_llm_extraction_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
import json
from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def extract_with_llm():
    """Extract Hacker News stories using LLM."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        # Define LLM extraction strategy
        config = CrawlerRunConfig(
            extraction_strategy={
                "type": "llm",
                "provider": "crawl4ai",
                "model": "openai/gpt-4o-mini",
                "instruction": """Extract all stories from this Hacker News page.
                For each story, extract:
                - title: The story title
                - url: The story URL
                - points: Number of points (if available)
                - author: Username who posted it
                - comments: Number of comments

                Return as a JSON array of story objects."""
            }
        )

        print("Crawling Hacker News with LLM extraction...")
        result = await crawler.run(
            url="https://news.ycombinator.com",
            strategy="http",
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
                print(f"Comments: {story.get('comments')}")

            # Show token usage
            if result.llm_usage:
                print(f"\nLLM Tokens Used: {result.llm_usage.total_tokens}")
        else:
            print(f"Error: {result.error_message}")


if __name__ == "__main__":
    asyncio.run(extract_with_llm())
