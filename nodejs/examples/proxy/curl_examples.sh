#!/bin/bash
#
# Crawl4AI Cloud - Proxy Usage Examples (cURL)
#
# Quick reference for all proxy configurations and crawl types.
#
# Proxy Modes:
# - none: Direct connection (1x credits)
# - datacenter: Fast datacenter proxies (2x credits)
# - residential: Premium residential IPs (5x credits)
# - auto: Smart selection based on target URL
#
# Provider Pool:
# - Massive: Residential ONLY
# - NST: Both datacenter AND residential
# - Scrapeless: Both datacenter AND residential

API_KEY="sk_live_YOUR_API_KEY_HERE"
BASE_URL="https://api.crawl4ai.com"

# =============================================================================
# SYNC SINGLE CRAWL
# =============================================================================

# No proxy (1x credits)
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://httpbin.org/ip",
    "proxy": {"mode": "none"},
    "bypass_cache": true
  }' | jq '{success, proxy_mode, proxy_used}'

# Datacenter proxy (2x credits)
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://httpbin.org/ip",
    "proxy": {"mode": "datacenter"},
    "bypass_cache": true
  }' | jq '{success, proxy_mode, proxy_used}'

# Datacenter with country
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://httpbin.org/ip",
    "proxy": {"mode": "datacenter", "country": "US"},
    "bypass_cache": true
  }' | jq

# Residential proxy (5x credits)
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://httpbin.org/ip",
    "proxy": {"mode": "residential"},
    "bypass_cache": true
  }' | jq '{success, proxy_mode, proxy_used}'

# Residential with country
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://httpbin.org/ip",
    "proxy": {"mode": "residential", "country": "GB"},
    "bypass_cache": true
  }' | jq

# Auto mode (smart selection)
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://amazon.com",
    "proxy": {"mode": "auto"},
    "bypass_cache": true
  }' | jq '{success, proxy_mode, proxy_used}'

# Force specific provider (for testing)
curl -s -X POST "$BASE_URL/v1/crawl" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://httpbin.org/ip",
    "proxy": {"mode": "datacenter", "provider": "nst"},
    "bypass_cache": true
  }' | jq

# =============================================================================
# BATCH CRAWL
# =============================================================================

# Batch with datacenter proxy
curl -s -X POST "$BASE_URL/v1/crawl/batch" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "urls": [
      "https://httpbin.org/ip",
      "https://httpbin.org/headers",
      "https://httpbin.org/user-agent"
    ],
    "proxy": {"mode": "datacenter"},
    "bypass_cache": true
  }' | jq '.results[] | {url, success, proxy_used}'

# Batch with residential + geo
curl -s -X POST "$BASE_URL/v1/crawl/batch" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "urls": [
      "https://httpbin.org/ip",
      "https://httpbin.org/headers"
    ],
    "proxy": {"mode": "residential", "country": "US"},
    "bypass_cache": true
  }' | jq

# =============================================================================
# ASYNC JOB
# =============================================================================

# Submit async job
curl -s -X POST "$BASE_URL/v1/crawl/async" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "urls": [
      "https://example.com",
      "https://httpbin.org/ip"
    ],
    "proxy": {"mode": "datacenter"},
    "bypass_cache": true
  }' | jq

# Check job status (replace JOB_ID)
JOB_ID="job_xxx"
curl -s -X GET "$BASE_URL/v1/crawl/jobs/$JOB_ID" \
  -H "X-API-Key: $API_KEY" | jq

# =============================================================================
# DEEP CRAWL
# =============================================================================

# Deep crawl with datacenter
curl -s -X POST "$BASE_URL/v1/crawl/deep" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://example.com",
    "strategy": "bfs",
    "max_depth": 2,
    "max_urls": 10,
    "proxy": {"mode": "datacenter"}
  }' | jq

# Deep crawl with STICKY SESSION (same IP for all URLs)
curl -s -X POST "$BASE_URL/v1/crawl/deep" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://example.com",
    "strategy": "bfs",
    "max_depth": 2,
    "max_urls": 10,
    "proxy": {
      "mode": "datacenter",
      "sticky_session": true
    }
  }' | jq

# Deep crawl residential + sticky (for protected sites)
curl -s -X POST "$BASE_URL/v1/crawl/deep" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "url": "https://docs.example.com",
    "strategy": "bfs",
    "max_depth": 2,
    "max_urls": 20,
    "proxy": {
      "mode": "residential",
      "country": "US",
      "sticky_session": true
    }
  }' | jq

# Check deep crawl status (replace JOB_ID)
DEEP_JOB_ID="deep_xxx"
curl -s -X GET "$BASE_URL/v1/crawl/deep/$DEEP_JOB_ID" \
  -H "X-API-Key: $API_KEY" | jq

# =============================================================================
# CREDIT COSTS
# =============================================================================
#
# Mode          | Multiplier | Cost per URL
# ------------- | ---------- | ------------
# none          | 1x         | 100 credits
# datacenter    | 2x         | 200 credits
# residential   | 5x         | 500 credits
# auto          | varies     | 100-500 credits
#
# Example: Batch of 10 URLs with residential = 10 * 500 = 5000 credits

# =============================================================================
# PROVIDER INFO
# =============================================================================
#
# Provider   | Supports        | Notes
# ---------- | --------------- | -----
# Massive    | Residential     | Static credentials, weight: 50
# NST        | DC + Residential| API-based, weight: 60
# Scrapeless | DC + Residential| API-based, weight: 40
#
# Selection is weighted random within the eligible providers for each mode.
