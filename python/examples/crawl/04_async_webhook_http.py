#!/usr/bin/env python3
"""
Async Crawl with Webhook - HTTP Example

This script demonstrates async crawling with webhook notification.
Create a job with a webhook URL - the API will POST results when complete.
No polling required!

Usage:
    python 04_async_webhook_http.py

Requirements:
    pip install httpx

Webhook Payload:
    The API will POST to your webhook_url with:
    {
        "job_id": "job_123",
        "status": "completed",
        "progress": {"completed": 4, "failed": 0, "total": 4},
        "results": [...],  # Full crawl results
        "created_at": "2024-01-01T00:00:00Z",
        "completed_at": "2024-01-01T00:01:00Z"
    }
"""

import httpx

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key
API_URL = "https://api.crawl4ai.com"
WEBHOOK_URL = "https://your-webhook-endpoint.com/callback"  # Your webhook URL


def main():
    """Create an async crawl job with webhook notification."""
    # URLs to crawl (can be more than 10 for async)
    urls = [
        "https://example.com",
        "https://httpbin.org/html",
        "https://httpbin.org/json",
        "https://httpbin.org/robots.txt",
    ]

    print(f"Creating async job for {len(urls)} URLs with webhook...")

    with httpx.Client(timeout=120) as client:
        try:
            # Create async job with webhook
            response = client.post(
                f"{API_URL}/v1/crawl/async",
                headers={
                    "X-API-Key": API_KEY,
                    "Content-Type": "application/json"
                },
                json={
                    "urls": urls,
                    "strategy": "http",  # Options: "browser" (JS support) or "http" (faster, no JS)
                    "webhook_url": WEBHOOK_URL,  # API will POST here when complete
                    "priority": 7,  # Higher priority (1-10)
                }
            )

            response.raise_for_status()
            data = response.json()

            # Display job info
            print(f"\n=== JOB CREATED ===")
            print(f"Job ID: {data['job_id']}")
            print(f"Status: {data['status']}")
            print(f"Webhook: {WEBHOOK_URL}")
            print(f"\nThe API will POST results to your webhook when complete.")
            print(f"No polling required!")

            # You can still check status manually if needed
            print(f"\nManual status check: GET {API_URL}/v1/crawl/jobs/{data['job_id']}")

        except httpx.HTTPStatusError as e:
            print(f"HTTP Error {e.response.status_code}: {e.response.text}")
        except Exception as e:
            print(f"Error: {e}")


if __name__ == "__main__":
    main()


# Example webhook handler (Flask):
"""
from flask import Flask, request

app = Flask(__name__)

@app.route('/callback', methods=['POST'])
def webhook():
    data = request.json
    print(f"Job {data['job_id']} completed!")
    print(f"Status: {data['status']}")
    print(f"Results: {len(data['results'])} URLs crawled")
    return {'status': 'received'}, 200

if __name__ == '__main__':
    app.run(port=8000)
"""
