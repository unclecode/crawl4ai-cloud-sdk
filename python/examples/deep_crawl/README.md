# Deep Crawl V2 Examples

Examples demonstrating the Deep Crawl V2 API features for multi-page crawling.

## Overview

Deep Crawl V2 enables intelligent URL discovery and two-phase scan→extract workflows.

### Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `map` | Sitemap-based discovery | Sites with good sitemaps |
| `bfs` | Breadth-first traversal | Explore all pages at each depth level |
| `dfs` | Depth-first traversal | Follow paths deeply before backtracking |
| `best_first` | Priority-based with scoring | Find most relevant content first |

### Two-Phase Workflow

```
Phase 1: SCAN (scan_only=True)
  - Discover URLs (sitemap, links, Common Crawl)
  - Cache HTML content (30 min TTL)
  - Return scan_job_id

Phase 2: EXTRACT (source_job_id=scan_job_id)
  - Use cached HTML from scan phase
  - Apply extraction strategy
  - No re-crawling needed
```

## Examples

| # | File | Description |
|---|------|-------------|
| 01 | `01_map_strategy.py` | Sitemap-based discovery (map strategy) |
| 02 | `02_tree_strategies.py` | BFS and DFS traversal with max_depth |
| 03 | `03_best_first_scoring.py` | Priority crawling with keyword scorers |
| 04 | `04_scan_only_discovery.py` | URL discovery without crawling |
| 05 | `05_two_phase_workflow.py` | Full scan → extract workflow |
| 06 | `06_filters_and_patterns.py` | URL filtering and domain controls |
| 07 | `07_with_extraction.py` | Deep crawl with CSS/LLM extraction |

## Quick Start

```python
from crawl4ai_cloud import Crawl4AI

client = Crawl4AI(api_key="your_api_key")

# Simple deep crawl with BFS
job = client.deep_crawl(
    url="https://docs.crawl4ai.com",
    strategy="bfs",
    max_depth=2,
    max_urls=20,
    wait=True  # Block until complete
)

print(f"Crawled {job.progress.completed} pages")

# Results are dicts (not objects)
for result in job.results:
    print(f"  - {result['url']}")
    if result.get('markdown'):
        print(f"    {result['markdown']['raw_markdown'][:100]}...")

client.close()
```

## Scan-Only Example

```python
# Discover URLs without crawling
scan = client.deep_crawl(
    url="https://docs.crawl4ai.com",
    strategy="bfs",
    max_depth=2,
    max_urls=50,
    scan_only=True,  # Just discover
    wait=True
)

print(f"Found {scan.discovered_count} URLs")
print(f"Cache expires: {scan.cache_expires_at}")

# Later: extract from cached HTML
job = client.deep_crawl(
    source_job_id=scan.job_id,  # Use cached HTML
    crawler_config={
        "extraction_strategy": {
            "type": "json_css",
            "schema": {...}
        }
    },
    wait=True
)
```

## Best-First with Scoring

```python
job = client.deep_crawl(
    url="https://docs.crawl4ai.com",
    strategy="best_first",
    max_depth=3,
    max_urls=30,
    scorers={
        "keywords": ["api", "tutorial", "guide"],
        "optimal_depth": 2,
        "weights": {"keywords": 2.0, "depth": 1.0}
    },
    wait=True
)
```

## Requirements

```bash
pip install crawl4ai-cloud
```

## Configuration

Replace `"your_api_key_here"` in examples with your actual API key:
```python
API_KEY = "sk_live_xxxxx"
```

## Notes

- **Results are dicts**: Access with `result['url']`, not `result.url`
- **Cache TTL**: HTML cache expires after 30 minutes
- **scan_only=True**: Returns `DeepCrawlResult` (not `Job`)
- **wait=True**: Returns `Job` with full results
- **wait=False**: Returns immediately, poll status manually

## API Documentation

- SDK: https://docs.crawl4ai.com/cloud/sdk
- API Reference: https://docs.crawl4ai.com/cloud/api

## Support

- Issues: https://github.com/unclecode/crawl4ai/issues
- Discord: https://discord.gg/crawl4ai
