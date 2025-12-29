#!/usr/bin/env python3
"""
Screenshots and PDFs - SDK Example

This script demonstrates how to capture screenshots and PDFs using the Crawl4AI SDK.
Screenshots capture the visual state of the page, while PDFs generate a print-ready document.

Usage:
    python 01_screenshots_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
import base64
from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def capture_screenshot(url: str):
    """Capture a screenshot of a webpage."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print(f"Capturing screenshot of {url}...")

        result = await crawler.run(
            url=url,
            config=CrawlerRunConfig(
                screenshot=True,
                wait_for=".content",  # Wait for content to load
            )
        )

        if result.screenshot:
            # Screenshot is returned as base64-encoded string
            print(f"Screenshot captured: {len(result.screenshot)} bytes (base64)")

            # Save to file
            with open("screenshot.png", "wb") as f:
                f.write(base64.b64decode(result.screenshot))
            print("Screenshot saved to screenshot.png")
        else:
            print("No screenshot available")

        return result


async def capture_pdf(url: str):
    """Generate a PDF of a webpage."""
    async with AsyncWebCrawler(api_key=API_KEY) as crawler:
        print(f"Generating PDF of {url}...")

        result = await crawler.run(
            url=url,
            config=CrawlerRunConfig(pdf=True)
        )

        if result.pdf:
            # PDF is returned as base64-encoded string
            print(f"PDF generated: {len(result.pdf)} bytes (base64)")

            # Save to file
            with open("page.pdf", "wb") as f:
                f.write(base64.b64decode(result.pdf))
            print("PDF saved to page.pdf")
        else:
            print("No PDF available")

        return result


async def main():
    # Example 1: Capture screenshot
    await capture_screenshot("https://www.example.com")

    # Example 2: Generate PDF
    await capture_pdf("https://www.example.com")


if __name__ == "__main__":
    asyncio.run(main())
