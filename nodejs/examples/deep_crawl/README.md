# Deep Crawl Examples

This directory contains TypeScript/JavaScript examples for deep crawling with Crawl4AI Cloud.

## Strategies

### Map Strategy (Sitemap Discovery)
- **01_map_strategy.ts** - Discover and crawl URLs from sitemaps

### Tree Strategies (Link Following)
- **02_tree_strategies.ts** - BFS and DFS crawling strategies

### Best-First (Prioritized)
- **03_best_first_scoring.ts** - Crawl high-value pages first using scoring

### Scan-Only (Discovery)
- **04_scan_only_discovery.ts** - Discover URLs without crawling

### Two-Phase Workflow
- **05_two_phase_workflow.ts** - Scan once, extract multiple times

### Filtering
- **06_filters_and_patterns.ts** - URL patterns and domain filtering

### With Extraction
- **07_with_extraction.ts** - Combine deep crawl with data extraction

## Strategy Comparison

| Strategy | Best For | URL Discovery |
|----------|----------|---------------|
| `map` | Sites with sitemaps | Sitemap parsing |
| `bfs` | Broad exploration | Link following |
| `dfs` | Deep navigation | Link following |
| `best_first` | Targeted content | Scored link following |

## Usage

```bash
npx ts-node 01_map_strategy.ts
```

## Key Parameters

- `strategy` - Crawl strategy: "map", "bfs", "dfs", "best_first"
- `maxDepth` - Maximum link depth for tree strategies
- `maxUrls` - Maximum URLs to crawl
- `scanOnly` - Discover URLs without extracting content
- `wait` - Wait for crawl to complete (default: false)
- `filters` - URL pattern and domain filters
- `scorers` - Scoring config for best-first strategy
