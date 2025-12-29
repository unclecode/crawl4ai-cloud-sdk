#!/usr/bin/env python3
"""
Custom Browser Configuration - HTTP Example

This script demonstrates how to customize browser settings using raw HTTP requests.
Covers viewport, headless mode, proxies, and custom headers.

Usage:
    python 02_custom_browser_config_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key


def crawl_with_custom_viewport(url: str):
    """Crawl with a custom viewport size."""
    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    print(f"Crawling {url} with custom viewport...")

    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": url,
            "strategy": "browser",
            "browser_config": {
                "viewport": {"width": 1920, "height": 1080}
            }
        },
        timeout=120.0
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Success! Title: {data.get('metadata', {}).get('title', 'N/A')}")
        return data
    else:
        print(f"Error: {response.status_code} - {response.text}")
        return None


def crawl_with_proxy(url: str, proxy_url: str):
    """Crawl through a proxy server."""
    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    print(f"Crawling {url} through proxy...")

    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": url,
            "strategy": "browser",
            "browser_config": {
                "proxy": proxy_url
            }
        },
        timeout=120.0
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Success! Title: {data.get('metadata', {}).get('title', 'N/A')}")
        return data
    else:
        print(f"Error: {response.status_code} - {response.text}")
        return None


def crawl_with_custom_headers(url: str):
    """Crawl with custom User-Agent and headers."""
    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    print(f"Crawling {url} with custom headers...")

    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": url,
            "strategy": "browser",
            "browser_config": {
                "headers": {
                    "User-Agent": "CustomBot/1.0 (Research purposes)",
                    "Accept-Language": "en-US,en;q=0.9"
                }
            }
        },
        timeout=120.0
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Success! Title: {data.get('metadata', {}).get('title', 'N/A')}")
        return data
    else:
        print(f"Error: {response.status_code} - {response.text}")
        return None


def crawl_with_full_config(url: str):
    """Crawl with comprehensive browser configuration."""
    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    print(f"Crawling {url} with full custom config...")

    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": url,
            "strategy": "browser",
            "browser_config": {
                "headless": True,
                "viewport": {"width": 1920, "height": 1080},
                "headers": {
                    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
                }
            },
            "crawler_config": {
                "wait_for": "networkidle",
                "page_timeout": 30000
            }
        },
        timeout=120.0
    )

    if response.status_code == 200:
        data = response.json()
        print(f"Success! Title: {data.get('metadata', {}).get('title', 'N/A')}")
        print(f"HTML size: {len(data.get('html', ''))} bytes")
        return data
    else:
        print(f"Error: {response.status_code} - {response.text}")
        return None


if __name__ == "__main__":
    # Example 1: Custom viewport
    crawl_with_custom_viewport("https://www.example.com")

    # Example 2: Custom headers
    crawl_with_custom_headers("https://www.example.com")

    # Example 3: Full configuration
    crawl_with_full_config("https://www.example.com")

    # Example 4: With proxy (uncomment to use)
    # crawl_with_proxy("https://www.example.com", "http://proxy.example.com:8080")
