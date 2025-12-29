# Deep Crawl Examples

This directory contains Go examples for deep crawling with Crawl4AI Cloud.

## Strategies

### Map Strategy (Sitemap Discovery)
- **01_map_strategy.go** - Discover and crawl URLs from sitemaps

### Tree Strategies (Link Following)
- **02_tree_strategies.go** - BFS and DFS crawling strategies

### Best-First (Prioritized)
- **03_best_first_scoring.go** - Crawl high-value pages first using scoring

### Scan-Only (Discovery)
- **04_scan_only_discovery.go** - Discover URLs without crawling

### Two-Phase Workflow
- **05_two_phase_workflow.go** - Scan once, extract multiple times

### Filtering
- **06_filters_and_patterns.go** - URL patterns and domain filtering

### With Extraction
- **07_with_extraction.go** - Combine deep crawl with data extraction

## Strategy Comparison

| Strategy | Best For | URL Discovery |
|----------|----------|---------------|
| `map` | Sites with sitemaps | Sitemap parsing |
| `bfs` | Broad exploration | Link following |
| `dfs` | Deep navigation | Link following |
| `best_first` | Targeted content | Scored link following |

## Usage

```bash
cd deep_crawl
go run 01_map_strategy.go
```

## Key Parameters

- `Strategy` - Crawl strategy: "bfs", "dfs", "best_first", "map"
- `MaxDepth` - Maximum link depth for tree strategies
- `MaxURLs` - Maximum URLs to crawl
- `ScanOnly` - Discover URLs without extracting content
- `Wait` - Wait for crawl to complete (default: false)
- `Filters` - URL pattern and domain filters
- `Scorers` - Scoring config for best-first strategy
