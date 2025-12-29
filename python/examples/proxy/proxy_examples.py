"""
Crawl4AI Cloud - Proxy Usage Examples

This file demonstrates all proxy configurations and crawl types.

Proxy Modes:
- none: Direct connection (1x credits)
- datacenter: Fast datacenter proxies (2x credits) - NST/Scrapeless
- residential: Premium residential IPs (5x credits) - Massive/NST/Scrapeless
- auto: Smart selection based on target URL

Crawl Types:
- Sync single: POST /v1/crawl
- Sync batch: POST /v1/crawl/batch
- Async job: POST /v1/crawl/async
- Deep crawl: POST /v1/crawl/deep

Provider Pool:
- Massive: Residential ONLY (static credentials)
- NST: Both datacenter AND residential
- Scrapeless: Both datacenter AND residential
"""

import requests
import time
from typing import Optional, Dict, Any

# Configuration
API_BASE = "https://api.crawl4ai.com"
API_KEY = "sk_live_YOUR_API_KEY_HERE"

HEADERS = {
    "Content-Type": "application/json",
    "X-API-Key": API_KEY
}


# =============================================================================
# SYNC SINGLE CRAWL - All Proxy Modes
# =============================================================================

def crawl_no_proxy(url: str) -> Dict[str, Any]:
    """
    Crawl without proxy - direct connection.
    Cost: 1x credits (100 credits per URL)
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl",
        headers=HEADERS,
        json={
            "url": url,
            "proxy": {"mode": "none"},
            "bypass_cache": True
        }
    )
    return response.json()
    # Response: {
    #   "success": true,
    #   "url": "https://example.com",
    #   "proxy_used": null,
    #   "proxy_mode": "none",
    #   "html": "...",
    #   "markdown": {...},
    #   "cleaned_html": "...",
    #   "metadata": {...}
    # }


def crawl_datacenter_proxy(url: str, country: Optional[str] = None) -> Dict[str, Any]:
    """
    Crawl with datacenter proxy - fast and cheap.
    Providers: NST or Scrapeless (weighted random: 60% NST, 40% Scrapeless)
    Cost: 2x credits (200 credits per URL)

    Good for: General scraping, non-protected sites, high volume
    """
    proxy_config = {"mode": "datacenter"}
    if country:
        proxy_config["country"] = country  # ISO code: "US", "GB", "DE", etc.

    response = requests.post(
        f"{API_BASE}/v1/crawl",
        headers=HEADERS,
        json={
            "url": url,
            "proxy": proxy_config,
            "bypass_cache": True
        }
    )
    return response.json()
    # Response: {
    #   "success": true,
    #   "url": "https://example.com",
    #   "proxy_used": "nst",  # or "scrapeless"
    #   "proxy_mode": "datacenter",
    #   "html": "...",
    #   ...
    # }


def crawl_residential_proxy(url: str, country: Optional[str] = None) -> Dict[str, Any]:
    """
    Crawl with residential proxy - premium IPs for protected sites.
    Providers: Massive, NST, or Scrapeless (weighted random)
    Cost: 5x credits (500 credits per URL)

    Good for: Amazon, LinkedIn, Google, social media, anti-bot sites
    """
    proxy_config = {"mode": "residential"}
    if country:
        proxy_config["country"] = country

    response = requests.post(
        f"{API_BASE}/v1/crawl",
        headers=HEADERS,
        json={
            "url": url,
            "proxy": proxy_config,
            "bypass_cache": True
        }
    )
    return response.json()
    # Response: {
    #   "success": true,
    #   "url": "https://amazon.com",
    #   "proxy_used": "massive",  # or "nst" or "scrapeless"
    #   "proxy_mode": "residential",
    #   "html": "...",
    #   ...
    # }


def crawl_auto_proxy(url: str) -> Dict[str, Any]:
    """
    Crawl with auto proxy mode - smart selection based on URL.

    Heuristics:
    - Hard targets (amazon, linkedin, google, etc.) -> residential
    - Easy targets (example.com, httpbin, github) -> no proxy
    - Unknown domains -> datacenter

    Cost: Varies based on selection (1x, 2x, or 5x)
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl",
        headers=HEADERS,
        json={
            "url": url,
            "proxy": {"mode": "auto"},
            "bypass_cache": True
        }
    )
    return response.json()
    # Response: {
    #   "success": true,
    #   "url": "https://amazon.com",
    #   "proxy_used": "nst",
    #   "proxy_mode": "residential",  # auto selected residential for amazon
    #   "html": "...",
    #   ...
    # }


def crawl_specific_provider(url: str, provider: str, mode: str) -> Dict[str, Any]:
    """
    Force a specific proxy provider (for testing/debugging).

    Args:
        provider: "massive" (residential only), "nst", or "scrapeless"
        mode: "datacenter" or "residential"

    Note: Massive only supports residential mode.
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl",
        headers=HEADERS,
        json={
            "url": url,
            "proxy": {
                "mode": mode,
                "provider": provider
            },
            "bypass_cache": True
        }
    )
    return response.json()


# =============================================================================
# SYNC BATCH CRAWL - Multiple URLs with Proxy
# =============================================================================

def batch_crawl_with_proxy(urls: list, mode: str = "datacenter") -> Dict[str, Any]:
    """
    Crawl multiple URLs in a single request with proxy.
    All URLs use the same proxy configuration.

    Cost: (number of URLs) * mode_multiplier credits
    Example: 5 URLs with datacenter = 5 * 200 = 1000 credits
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl/batch",
        headers=HEADERS,
        json={
            "urls": urls,
            "proxy": {"mode": mode},
            "bypass_cache": True
        }
    )
    return response.json()
    # Response: {
    #   "results": [
    #     {"success": true, "url": "...", "proxy_used": "nst", ...},
    #     {"success": true, "url": "...", "proxy_used": "scrapeless", ...},
    #     ...
    #   ]
    # }


def batch_crawl_residential_geo(urls: list, country: str = "US") -> Dict[str, Any]:
    """
    Batch crawl with residential proxy from specific country.
    Useful for geo-restricted content.
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl/batch",
        headers=HEADERS,
        json={
            "urls": urls,
            "proxy": {
                "mode": "residential",
                "country": country
            },
            "bypass_cache": True
        }
    )
    return response.json()


# =============================================================================
# ASYNC JOB CRAWL - Background Processing with Proxy
# =============================================================================

def async_crawl_with_proxy(urls: list, mode: str = "datacenter") -> str:
    """
    Submit async crawl job with proxy.
    Returns job ID for polling.

    Good for: Large batches, long-running crawls, non-blocking operations
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl/async",
        headers=HEADERS,
        json={
            "urls": urls,
            "proxy": {"mode": mode},
            "bypass_cache": True
        }
    )
    data = response.json()
    return data["job_id"]
    # Response: {
    #   "job_id": "job_abc123...",
    #   "status": "pending",
    #   "urls_count": 5
    # }


def get_job_status(job_id: str) -> Dict[str, Any]:
    """
    Check async job status.
    """
    response = requests.get(
        f"{API_BASE}/v1/crawl/jobs/{job_id}",
        headers=HEADERS
    )
    return response.json()
    # Response when pending: {
    #   "job_id": "job_abc123...",
    #   "status": "processing",
    #   "progress": {"completed": 2, "total": 5}
    # }
    # Response when complete: {
    #   "job_id": "job_abc123...",
    #   "status": "completed",
    #   "results": [...]
    # }


def async_crawl_wait_complete(urls: list, mode: str = "datacenter",
                               timeout: int = 300) -> Dict[str, Any]:
    """
    Submit async job and wait for completion.
    """
    job_id = async_crawl_with_proxy(urls, mode)
    print(f"Job submitted: {job_id}")

    start = time.time()
    while time.time() - start < timeout:
        status = get_job_status(job_id)
        print(f"Status: {status.get('status')}")

        if status.get("status") == "completed":
            return status
        if status.get("status") == "failed":
            raise Exception(f"Job failed: {status.get('error')}")

        time.sleep(2)

    raise TimeoutError(f"Job {job_id} did not complete within {timeout}s")


# =============================================================================
# DEEP CRAWL - Multi-page Crawling with Sticky Sessions
# =============================================================================

def deep_crawl_with_proxy(url: str, mode: str = "datacenter",
                          max_depth: int = 2, max_urls: int = 10) -> Dict[str, Any]:
    """
    Deep crawl a site with proxy.
    Discovers and crawls linked pages.

    Strategies:
    - bfs: Breadth-first (same level pages first)
    - dfs: Depth-first (follow links deep)
    - bestfirst: Prioritize by relevance
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl/deep",
        headers=HEADERS,
        json={
            "url": url,
            "strategy": "bfs",
            "max_depth": max_depth,
            "max_urls": max_urls,
            "proxy": {"mode": mode}
        }
    )
    return response.json()
    # Response: {
    #   "job_id": "deep_abc123...",
    #   "status": "processing",
    #   "discovered_urls": ["...", "..."]
    # }


def deep_crawl_sticky_session(url: str, mode: str = "datacenter",
                               max_depth: int = 2, max_urls: int = 10) -> Dict[str, Any]:
    """
    Deep crawl with sticky session - same proxy IP for all URLs.

    IMPORTANT: Sticky sessions ensure all pages in the crawl use the
    same proxy IP address. This is crucial for:
    - Session-based authentication
    - Rate limiting that tracks by IP
    - Sites that detect IP changes

    The proxy IP is cached for the duration of the job and released
    when the job completes.
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl/deep",
        headers=HEADERS,
        json={
            "url": url,
            "strategy": "bfs",
            "max_depth": max_depth,
            "max_urls": max_urls,
            "proxy": {
                "mode": mode,
                "sticky_session": True  # Same IP for all URLs
            }
        }
    )
    return response.json()


def deep_crawl_residential_sticky(url: str, country: str = "US") -> Dict[str, Any]:
    """
    Deep crawl protected site with residential proxy and sticky session.
    Best for: E-commerce sites, social media, heavily protected targets.
    """
    response = requests.post(
        f"{API_BASE}/v1/crawl/deep",
        headers=HEADERS,
        json={
            "url": url,
            "strategy": "bfs",
            "max_depth": 2,
            "max_urls": 20,
            "proxy": {
                "mode": "residential",
                "country": country,
                "sticky_session": True
            }
        }
    )
    return response.json()


def get_deep_crawl_status(job_id: str) -> Dict[str, Any]:
    """
    Check deep crawl job status.
    """
    response = requests.get(
        f"{API_BASE}/v1/crawl/deep/{job_id}",
        headers=HEADERS
    )
    return response.json()


# =============================================================================
# RESPONSE TYPE EXAMPLES
# =============================================================================

"""
SYNC CRAWL RESPONSE:
{
    "success": true,
    "url": "https://example.com",
    "proxy_used": "nst",           # Provider used: "massive", "nst", "scrapeless", or null
    "proxy_mode": "datacenter",    # Mode used: "none", "datacenter", "residential"
    "html": "<!DOCTYPE html>...",
    "cleaned_html": "<main>...</main>",
    "markdown": {
        "raw_markdown": "# Title\n\nContent...",
        "markdown_with_citations": "# Title\n\nContent [1]...",
        "references_markdown": "[1]: https://...",
        "fit_markdown": "Title\nContent..."
    },
    "metadata": {
        "title": "Page Title",
        "description": "Meta description",
        "keywords": ["keyword1", "keyword2"],
        "author": "Author Name"
    },
    "links": {
        "internal": ["https://example.com/page1", "..."],
        "external": ["https://other.com", "..."]
    },
    "media": {
        "images": [{"src": "...", "alt": "..."}],
        "videos": [],
        "audios": []
    },
    "screenshot": "base64...",     # If screenshot=true
    "pdf": "base64..."             # If pdf=true
}

BATCH CRAWL RESPONSE:
{
    "results": [
        {
            "success": true,
            "url": "https://example1.com",
            "proxy_used": "nst",
            "proxy_mode": "datacenter",
            "html": "...",
            ...
        },
        {
            "success": true,
            "url": "https://example2.com",
            "proxy_used": "scrapeless",
            "proxy_mode": "datacenter",
            "html": "...",
            ...
        }
    ]
}

ASYNC JOB RESPONSE (Submit):
{
    "job_id": "job_abc123def456...",
    "status": "pending",
    "urls_count": 5
}

ASYNC JOB RESPONSE (Status - Processing):
{
    "job_id": "job_abc123def456...",
    "status": "processing",
    "progress": {
        "completed": 2,
        "total": 5,
        "failed": 0
    }
}

ASYNC JOB RESPONSE (Status - Complete):
{
    "job_id": "job_abc123def456...",
    "status": "completed",
    "results": [
        {"success": true, "url": "...", "proxy_used": "nst", ...},
        ...
    ]
}

DEEP CRAWL RESPONSE (Submit):
{
    "job_id": "deep_abc123...",
    "status": "processing",
    "seed_url": "https://example.com",
    "discovered_urls": [
        "https://example.com/page1",
        "https://example.com/page2"
    ]
}

DEEP CRAWL RESPONSE (Complete):
{
    "job_id": "deep_abc123...",
    "status": "completed",
    "seed_url": "https://example.com",
    "crawled_count": 10,
    "results": [
        {"url": "...", "success": true, "proxy_used": "nst", ...},
        ...
    ]
}
"""


# =============================================================================
# USAGE EXAMPLES
# =============================================================================

if __name__ == "__main__":
    # Example 1: Simple crawl without proxy
    print("=== No Proxy ===")
    result = crawl_no_proxy("https://httpbin.org/ip")
    print(f"Success: {result.get('success')}")
    print(f"Proxy: {result.get('proxy_mode')}")

    # Example 2: Datacenter proxy
    print("\n=== Datacenter Proxy ===")
    result = crawl_datacenter_proxy("https://httpbin.org/ip")
    print(f"Success: {result.get('success')}")
    print(f"Provider: {result.get('proxy_used')}")
    print(f"Mode: {result.get('proxy_mode')}")

    # Example 3: Residential proxy for protected site
    print("\n=== Residential Proxy ===")
    result = crawl_residential_proxy("https://httpbin.org/ip", country="US")
    print(f"Success: {result.get('success')}")
    print(f"Provider: {result.get('proxy_used')}")

    # Example 4: Auto mode
    print("\n=== Auto Mode (Amazon) ===")
    result = crawl_auto_proxy("https://amazon.com")
    print(f"Auto selected: {result.get('proxy_mode')}")
    print(f"Provider: {result.get('proxy_used')}")

    # Example 5: Deep crawl with sticky session
    print("\n=== Deep Crawl with Sticky Session ===")
    result = deep_crawl_sticky_session(
        "https://example.com",
        mode="datacenter",
        max_depth=1,
        max_urls=3
    )
    print(f"Job ID: {result.get('job_id')}")
    print(f"Status: {result.get('status')}")
