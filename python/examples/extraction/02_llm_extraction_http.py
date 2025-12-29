#!/usr/bin/env python3
"""
LLM Extraction with HTTP

This example shows how to extract structured data using LLM
via direct HTTP API calls (no SDK).

Usage:
    python 02_llm_extraction_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key

def extract_with_llm():
    """Extract Hacker News stories using LLM via HTTP API."""

    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    # Define LLM extraction strategy
    payload = {
        "url": "https://news.ycombinator.com",
        "strategy": "http",
        "crawler_config": {
            "extraction_strategy": {
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
        }
    }

    print("Crawling Hacker News with LLM extraction...")
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
        print(f"Comments: {story.get('comments')}")

if __name__ == "__main__":
    extract_with_llm()
