#!/usr/bin/env python3
"""
Async Crawl with Polling - HTTP Example

This script demonstrates async crawling with manual polling loop.
Create a job, then poll the status endpoint until completion.

Usage:
    python 03_async_crawl_http.py

Requirements:
    pip install httpx
"""

import httpx
import time

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key
API_URL = "https://api.crawl4ai.com"


def main():
    """Create an async crawl job and poll for completion."""
    # URLs to crawl (can be more than 10 for async)
    urls = [
        "https://example.com",
        "https://httpbin.org/html",
        "https://httpbin.org/json",
        "https://httpbin.org/robots.txt",
    ]

    print(f"Creating async job for {len(urls)} URLs...")

    with httpx.Client(timeout=120) as client:
        try:
            # Step 1: Create the async job
            response = client.post(
                f"{API_URL}/v1/crawl/async",
                headers={
                    "X-API-Key": API_KEY,
                    "Content-Type": "application/json"
                },
                json={
                    "urls": urls,
                    "strategy": "http",  # Options: "browser" (JS support) or "http" (faster, no JS)
                    "priority": 5,  # Priority 1-10 (default: 5)
                }
            )

            response.raise_for_status()
            data = response.json()
            job_id = data["job_id"]

            print(f"Job created: {job_id}")
            print(f"Status: {data['status']}")

            # Step 2: Poll for completion
            print("\nPolling for completion...")
            max_attempts = 60
            poll_interval = 2

            for attempt in range(max_attempts):
                time.sleep(poll_interval)

                status_response = client.get(
                    f"{API_URL}/v1/crawl/jobs/{job_id}",
                    headers={"X-API-Key": API_KEY}
                )

                status_response.raise_for_status()
                status_data = status_response.json()

                print(f"  [{attempt + 1}] Status: {status_data['status']} | "
                      f"Progress: {status_data['progress']['completed']}/{status_data['progress']['total']}")

                if status_data['status'] in ['completed', 'partial', 'failed']:
                    print(f"\n=== JOB COMPLETE ===")
                    print(f"Final status: {status_data['status']}")
                    print(f"Results available at: /v1/crawl/jobs/{job_id}?include_results=true")
                    break
            else:
                print("\nTimeout: Job did not complete in time")

        except httpx.HTTPStatusError as e:
            print(f"HTTP Error {e.response.status_code}: {e.response.text}")
        except Exception as e:
            print(f"Error: {e}")


if __name__ == "__main__":
    main()
