// Deep Crawl - URL Filtering and Patterns
//
// Control which URLs are crawled using filters and patterns:
// - Glob patterns: Match URL paths (e.g., "*/docs/*")
// - Domain filters: Allow/block specific domains
//
// Usage:
//
//	go run 06_filters_and_patterns.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func basicPatternFilter() {
	fmt.Println("=== Basic Pattern Filter ===\n")

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
		Pattern:  "*/docs/*", // Only URLs with /docs/ in path
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Matched URLs: %d\n", result.CrawlJob.Progress.Total)
		fmt.Printf("Crawled: %d\n", result.CrawlJob.Progress.Completed)

		if len(result.CrawlJob.Results) > 0 {
			fmt.Println("\nCrawled URLs:")
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

func multiplePatterns() {
	fmt.Println("\n=== Multiple Patterns ===\n")

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
		MaxURLs:  30,
		Filters: map[string]interface{}{
			// Match any of these patterns
			"patterns": []string{
				"/api/*",      // API reference pages
				"/guide/*",    // User guides
				"/tutorial/*", // Tutorials
				"*/example*",  // Example pages
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Matching URLs: %d\n", result.CrawlJob.Progress.Total)
	}
}

func domainFiltering() {
	fmt.Println("\n=== Domain Filtering ===\n")

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
			"domains": map[string]interface{}{
				// Never follow links to these domains
				"blocked": []string{
					"twitter.com",
					"facebook.com",
					"linkedin.com",
					"github.com", // If you want to stay on docs
				},
				// Or whitelist: only follow links to these
				// "allowed": []string{"docs.crawl4ai.com", "crawl4ai.com"},
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Crawled (blocked external): %d\n", result.CrawlJob.Progress.Total)
	}
}

func combinedFilters() {
	fmt.Println("\n=== Combined Filters ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 3,
		MaxURLs:  50,
		Filters: map[string]interface{}{
			// Include patterns (whitelist)
			"patterns": []string{"/docs/*", "/api/*", "/guide/*"},
			// Exclude patterns
			"exclude_patterns": []string{"*changelog*", "*version*"},
			// Domain controls
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
		fmt.Printf("Filtered & crawled: %d\n", result.CrawlJob.Progress.Total)
	}
}

func main() {
	basicPatternFilter()
	// Uncomment to run other examples:
	// multiplePatterns()
	// domainFiltering()
	// combinedFilters()
}
