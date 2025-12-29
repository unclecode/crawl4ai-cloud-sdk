#!/usr/bin/env python3
"""
Error Handling - SDK Example

This script demonstrates how to properly handle exceptions when using the Crawl4AI SDK.
Covers rate limits, quota errors, authentication errors, and other common exceptions.

Usage:
    python 03_error_handling_sdk.py

Requirements:
    pip install crawl4ai-cloud
"""

import asyncio
import time
from crawl4ai_cloud import AsyncWebCrawler
from crawl4ai_cloud.exceptions import (
    RateLimitError,
    QuotaExceededError,
    AuthenticationError,
    ValidationError,
    NotFoundError,
    ServerError,
    TimeoutError,
    Crawl4AIError
)

# Configuration
API_KEY = "YOUR_API_KEY"  # Replace with your API key


async def crawl_with_error_handling(url: str):
    """Crawl a URL with comprehensive error handling."""
    try:
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            print(f"Crawling {url}...")
            result = await crawler.run(url=url)
            print(f"Success! HTML size: {len(result.html)} bytes")
            return result

    except AuthenticationError as e:
        print(f"Authentication failed: {e}")
        print("Check your API key and make sure it's valid")

    except RateLimitError as e:
        print(f"Rate limit exceeded: {e}")
        print(f"Limit: {e.limit} requests per minute")
        print(f"Remaining: {e.remaining}")
        print(f"Retry after: {e.retry_after} seconds")

    except QuotaExceededError as e:
        print(f"Quota exceeded: {e}")
        print(f"Quota type: {e.quota_type}")  # 'daily', 'concurrent', or 'storage'

        if e.quota_type == "storage":
            print("Your storage is full. Delete old jobs to free up space.")
        elif e.quota_type == "daily":
            print("Daily crawl limit reached. Wait until tomorrow or upgrade plan.")
        elif e.quota_type == "concurrent":
            print("Too many concurrent requests. Wait for some to complete.")

    except ValidationError as e:
        print(f"Invalid request: {e}")
        print("Check your URL and configuration parameters")

    except NotFoundError as e:
        print(f"Resource not found: {e}")
        print("The job or session ID may be invalid or expired")

    except TimeoutError as e:
        print(f"Request timed out: {e}")
        print("The page took too long to load. Try increasing timeout.")

    except ServerError as e:
        print(f"Server error: {e}")
        print("The API is experiencing issues. Try again later.")

    except Crawl4AIError as e:
        # Catch-all for other SDK errors
        print(f"Crawl4AI error: {e}")
        print(f"Status code: {e.status_code}")

    except Exception as e:
        # Catch-all for unexpected errors
        print(f"Unexpected error: {e}")

    return None


async def crawl_with_retry_logic(url: str, max_retries: int = 3):
    """Crawl with automatic retry on transient errors."""
    for attempt in range(max_retries):
        try:
            async with AsyncWebCrawler(api_key=API_KEY) as crawler:
                print(f"Attempt {attempt + 1}/{max_retries}: Crawling {url}...")
                result = await crawler.run(url=url)
                print(f"Success! HTML size: {len(result.html)} bytes")
                return result

        except RateLimitError as e:
            if e.retry_after > 0:
                print(f"Rate limited. Waiting {e.retry_after}s...")
                await asyncio.sleep(e.retry_after)
                continue
            else:
                break

        except (ServerError, TimeoutError) as e:
            if attempt < max_retries - 1:
                wait_time = 2 ** attempt  # Exponential backoff
                print(f"Transient error: {e}. Retrying in {wait_time}s...")
                await asyncio.sleep(wait_time)
                continue
            else:
                print(f"Max retries reached. Last error: {e}")
                break

        except (AuthenticationError, QuotaExceededError, ValidationError) as e:
            # Don't retry on these errors
            print(f"Non-retryable error: {e}")
            break

    return None


async def main():
    # Example 1: Basic error handling
    print("=== Example 1: Basic Error Handling ===")
    await crawl_with_error_handling("https://www.example.com")

    # Example 2: Retry logic
    print("\n=== Example 2: With Retry Logic ===")
    await crawl_with_retry_logic("https://www.example.com")


if __name__ == "__main__":
    asyncio.run(main())
