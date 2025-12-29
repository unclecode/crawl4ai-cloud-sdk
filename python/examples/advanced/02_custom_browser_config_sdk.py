#!/usr/bin/env python3
"""
Custom Browser Configuration - SDK Example

This script demonstrates how to customize browser settings like viewport size,
proxies, and custom headers using the Crawl4AI SDK.

Usage:
    python 02_custom_browser_config_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig, BrowserConfig, ProxyConfig

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def crawl_with_custom_viewport(url: str):
    """Crawl with a custom viewport size (useful for responsive testing)."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print(f"Crawling {url} with custom viewport...")

        result = await crawler.run(
            url=url,
            strategy="browser",
            browser_config=BrowserConfig(
                viewport={"width": 1920, "height": 1080}
            )
        )

        print(f"Success! HTML size: {len(result.html)} bytes")
        return result


async def crawl_with_proxy(url: str):
    """Crawl using managed proxy service."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print(f"Crawling {url} with datacenter proxy...")

        result = await crawler.run(
            url=url,
            strategy="browser",
            proxy=ProxyConfig(mode="datacenter")  # or "residential"
        )

        print(f"Success! HTML size: {len(result.html)} bytes")
        return result


async def crawl_with_custom_headers(url: str):
    """Crawl with custom User-Agent and headers."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print(f"Crawling {url} with custom headers...")

        result = await crawler.run(
            url=url,
            strategy="browser",
            browser_config=BrowserConfig(
                headers={
                    "User-Agent": "CustomBot/1.0 (Research purposes)",
                    "Accept-Language": "en-US,en;q=0.9"
                }
            )
        )

        print(f"Success! HTML size: {len(result.html)} bytes")
        return result


async def crawl_with_full_config(url: str):
    """Crawl with comprehensive browser configuration."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print(f"Crawling {url} with full custom config...")

        result = await crawler.run(
            url=url,
            strategy="browser",
            browser_config=BrowserConfig(
                viewport={"width": 1920, "height": 1080},
                headers={
                    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
                }
            ),
            config=CrawlerRunConfig(
                wait_until="networkidle",  # Wait for network to be idle
                page_timeout=30000  # 30 second timeout
            )
        )

        print(f"Success! HTML size: {len(result.html)} bytes")
        return result


async def main():
    # Example 1: Custom viewport
    await crawl_with_custom_viewport("https://www.example.com")

    # Example 2: Custom headers
    await crawl_with_custom_headers("https://www.example.com")

    # Example 3: Full configuration
    await crawl_with_full_config("https://www.example.com")

    # Example 4: With managed proxy
    # await crawl_with_proxy("https://www.example.com")


if __name__ == "__main__":
    asyncio.run(main())
