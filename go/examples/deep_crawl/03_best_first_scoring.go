// Deep Crawl - Best-First Strategy with Scoring
//
// Best-first crawling prioritizes URLs based on relevance scores.
// Uses a priority queue to always crawl the highest-scoring URL next.
//
// Usage:
//
//	go run 03_best_first_scoring.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func bestFirstWithKeywords() {
	fmt.Println("=== Best-First with Keyword Scoring ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "best_first",
		MaxDepth: 3,
		MaxURLs:  15,
		Scorers: map[string]interface{}{
			"keywords": []string{"api", "tutorial", "guide", "example"},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Pages crawled: %d\n", result.CrawlJob.Progress.Completed)

		if len(result.CrawlJob.Results) > 0 {
			fmt.Println("\nTop results (by score):")
			for i, r := range result.CrawlJob.Results {
				if i >= 5 {
					break
				}
				if m, ok := r.(map[string]interface{}); ok {
					fmt.Printf("  %d. %v\n", i+1, m["url"])
				}
			}
		}
	}
}

func bestFirstForDocumentation() {
	fmt.Println("\n=== Best-First for API Docs ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "best_first",
		MaxDepth: 3,
		MaxURLs:  30,
		Scorers: map[string]interface{}{
			"keywords":      []string{"api", "reference", "method", "function", "parameter"},
			"optimal_depth": 2,
			"weights":       map[string]float64{"keywords": 3.0, "depth": 1.0},
		},
		Filters: map[string]interface{}{
			"patterns": []string{"/api/*", "/reference/*", "/docs/*"},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("API docs found: %d\n", result.CrawlJob.Progress.Completed)
	}
}

func main() {
	bestFirstWithKeywords()
	// Uncomment to run other examples:
	// bestFirstForDocumentation()
}
