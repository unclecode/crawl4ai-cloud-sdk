#!/usr/bin/env python3
"""
Schema Generation with HTTP

This example shows how to automatically generate CSS extraction schemas
from HTML using LLM via direct HTTP API calls (no SDK).

Usage:
    python 03_schema_generation_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key

def generate_extraction_schema():
    """Generate CSS schema for Hacker News stories via HTTP API."""

    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    # First, get the HTML content
    print("Fetching Hacker News HTML...")
    crawl_response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": "https://news.ycombinator.com",
            "strategy": "http"
        },
        timeout=60.0
    )

    if crawl_response.status_code != 200:
        print(f"Error: {crawl_response.status_code} - {crawl_response.text}")
        return

    html = crawl_response.json().get("html", "")
    print(f"Got {len(html)} bytes of HTML")

    # Generate schema using LLM
    print("\nGenerating CSS extraction schema...")
    schema_response = httpx.post(
        f"{API_URL}/v1/tools/schema",
        headers=headers,
        json={
            "html": html,
            "query": "Extract all stories with their title, URL, points, and author",
            "schema_type": "CSS"
        },
        timeout=60.0
    )

    if schema_response.status_code != 200:
        print(f"Error: {schema_response.status_code} - {schema_response.text}")
        return

    schema_data = schema_response.json()

    if schema_data.get("error"):
        print(f"Error: {schema_data['error']}")
        return

    schema = schema_data["schema"]
    print("\nGenerated Schema:")
    print(schema)

    # Now use the generated schema for extraction
    print("\n\nTesting generated schema...")
    extract_response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": "https://news.ycombinator.com",
            "strategy": "http",
            "crawler_config": {
                "extraction_strategy": {
                    "type": "json_css",
                    "schema": schema
                }
            }
        },
        timeout=60.0
    )

    if extract_response.status_code != 200:
        print(f"Error: {extract_response.status_code} - {extract_response.text}")
        return

    stories = extract_response.json().get("extracted_content", [])
    print(f"\nExtracted {len(stories)} stories")
    print("\nFirst 2 stories:")
    for story in stories[:2]:
        print(f"\n{story}")

if __name__ == "__main__":
    generate_extraction_schema()
