#!/usr/bin/env python3
"""
Sessions - Basic Session Management with HTTP API

This script demonstrates how to create and release browser sessions using
direct HTTP API calls instead of the SDK.

Usage:
    python 01_basic_session_http.py

Requirements:
    pip install httpx
"""

import httpx

# Configuration
API_URL = "https://api.crawl4ai.com"
API_KEY = "your_api_key_here"  # Replace with your API key


def main():
    """Create a session using HTTP API, print its details, and release it."""

    headers = {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json"
    }

    # Step 1: Create a browser session
    print("Creating browser session...")

    response = httpx.post(
        f"{API_URL}/v1/sessions",
        headers=headers,
        json={"timeout": 600},  # 10 minute timeout
        timeout=30.0
    )

    if response.status_code != 200:
        print(f"Error creating session: {response.status_code} - {response.text}")
        return

    session = response.json()

    print(f"\n=== SESSION CREATED ===")
    print(f"Session ID: {session['session_id']}")
    print(f"WebSocket URL: {session['ws_url']}")
    print(f"Expires in: {session['expires_in']} seconds")

    # Step 2: Use the session (see other examples for actual usage)
    print(f"\nYou can now connect to this browser using:")
    print(f"  - Crawl4AI: BrowserConfig(cdp_url='{session['ws_url']}')")
    print(f"  - Puppeteer: puppeteer.connect({{ browserWSEndpoint: '{session['ws_url']}' }})")
    print(f"  - Playwright: playwright.chromium.connectOverCDP('{session['ws_url']}')")

    # Step 3: Get session status
    print(f"\nChecking session status...")

    status_response = httpx.get(
        f"{API_URL}/v1/sessions/{session['session_id']}",
        headers=headers,
        timeout=30.0
    )

    if status_response.status_code == 200:
        status = status_response.json()
        print(f"Session status: {status.get('status', 'N/A')}")
        print(f"Worker ID: {status.get('worker_id', 'N/A')}")

    # Step 4: Release the session
    print(f"\nReleasing session...")

    delete_response = httpx.delete(
        f"{API_URL}/v1/sessions/{session['session_id']}",
        headers=headers,
        timeout=30.0
    )

    if delete_response.status_code == 200:
        print("Session released successfully!")
    else:
        print(f"Failed to release session: {delete_response.status_code}")


if __name__ == "__main__":
    main()
