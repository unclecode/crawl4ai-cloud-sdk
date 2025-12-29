// Deep Crawl - Tree Traversal Strategies (BFS & DFS)
//
// Tree strategies crawl by following links from the start URL:
// - BFS (Breadth-First): Explore all pages at depth N before depth N+1
// - DFS (Depth-First): Follow links deeply before backtracking
//
// Usage:
//
//	go run 02_tree_strategies.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func bfsCrawl() {
	fmt.Println("=== BFS Strategy (Breadth-First) ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 2,
		MaxURLs:  20,
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Status: %s\n", result.CrawlJob.Status)
		fmt.Printf("Pages crawled: %d\n", result.CrawlJob.Progress.Completed)

		if len(result.CrawlJob.Results) > 0 {
			fmt.Println("\nCrawled pages:")
			for i, r := range result.CrawlJob.Results {
				if i >= 5 {
					break
				}
				if m, ok := r.(map[string]interface{}); ok {
					fmt.Printf("  - %v\n", m["url"])
				}
			}
		}
	}
}

func dfsCrawl() {
	fmt.Println("\n=== DFS Strategy (Depth-First) ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "dfs",
		MaxDepth: 3,
		MaxURLs:  15,
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Status: %s\n", result.CrawlJob.Status)
		fmt.Printf("Pages crawled: %d\n", result.CrawlJob.Progress.Completed)
	}
}

func treeWithFilters() {
	fmt.Println("\n=== BFS with URL Filters ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 2,
		MaxURLs:  25,
		Filters: map[string]interface{}{
			"patterns": []string{"/docs/*", "/api/*", "/guide/*"},
			"domains": map[string]interface{}{
				"blocked": []string{"twitter.com", "github.com"},
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Filtered pages crawled: %d\n", result.CrawlJob.Progress.Completed)
	}
}

func main() {
	bfsCrawl()
	// Uncomment to run other examples:
	// dfsCrawl()
	// treeWithFilters()
}
