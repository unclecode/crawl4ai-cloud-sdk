#!/usr/bin/env python3
"""
Screenshots and PDFs - HTTP Example

This script demonstrates how to capture screenshots and PDFs using raw HTTP requests.
This is useful if you prefer direct API access without the SDK.

Usage:
    python 01_screenshots_http.py

Requirements:
    pip install httpx
"""

import httpx
import base64

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key


def capture_screenshot(url: str):
    """Capture a screenshot of a webpage."""
    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    print(f"Capturing screenshot of {url}...")

    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": url,
            "crawler_config": {
                "screenshot": True,
                "screenshot_wait_for": ".content"
            }
        },
        timeout=120.0
    )

    if response.status_code != 200:
        print(f"Error: {response.status_code} - {response.text}")
        return None

    data = response.json()

    if data.get("screenshot"):
        print(f"Screenshot captured: {len(data['screenshot'])} bytes (base64)")
        return data["screenshot"]
    else:
        print("No screenshot available")
        return None


def capture_pdf(url: str):
    """Generate a PDF of a webpage."""
    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    print(f"Generating PDF of {url}...")

    response = httpx.post(
        f"{API_URL}/v1/crawl",
        headers=headers,
        json={
            "url": url,
            "crawler_config": {
                "pdf": True
            }
        },
        timeout=120.0
    )

    if response.status_code != 200:
        print(f"Error: {response.status_code} - {response.text}")
        return None

    data = response.json()

    if data.get("pdf"):
        print(f"PDF generated: {len(data['pdf'])} bytes (base64)")
        return data["pdf"]
    else:
        print("No PDF available")
        return None


if __name__ == "__main__":
    # Example 1: Capture screenshot
    screenshot_b64 = capture_screenshot("https://www.example.com")

    # Example 2: Generate PDF
    pdf_b64 = capture_pdf("https://www.example.com")

    # Save screenshot to file
    if screenshot_b64:
        with open("screenshot.png", "wb") as f:
            f.write(base64.b64decode(screenshot_b64))
        print("\nScreenshot saved to screenshot.png")

    # Save PDF to file
    if pdf_b64:
        with open("page.pdf", "wb") as f:
            f.write(base64.b64decode(pdf_b64))
        print("PDF saved to page.pdf")
