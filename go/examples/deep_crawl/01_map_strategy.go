// Deep Crawl - Map Strategy (Sitemap Discovery)
//
// The "map" strategy discovers URLs from a website's sitemap and crawls them.
// This is the fastest way to crawl a site with a well-structured sitemap.
//
// Usage:
//
//	go run 01_map_strategy.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func basicMapCrawl() {
	fmt.Println("=== Basic Map Strategy ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// wait=true blocks until all URLs are crawled
	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "map", // Default strategy
		MaxURLs:  5,
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Status: %s\n", result.CrawlJob.Status)
		fmt.Printf("URLs crawled: %d/%d\n", result.CrawlJob.Progress.Completed, result.CrawlJob.Progress.Total)

		if len(result.CrawlJob.Results) > 0 {
			fmt.Printf("\nResults (%d pages):\n", len(result.CrawlJob.Results))
			for i, r := range result.CrawlJob.Results {
				if i >= 5 {
					break
				}
				if m, ok := r.(map[string]interface{}); ok {
					fmt.Printf("  - %v: %v\n", m["url"], m["success"])
				}
			}
		}
	}
}

func mapWithPattern() {
	fmt.Println("\n=== Map Strategy with Pattern Filter ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "map",
		Pattern:  "*/api/*", // Only URLs containing /api/
		MaxURLs:  10,
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Matched URLs: %d\n", result.CrawlJob.Progress.Total)
		fmt.Printf("Successfully crawled: %d\n", result.CrawlJob.Progress.Completed)
	}
}

func main() {
	basicMapCrawl()
	// Uncomment to run other examples:
	// mapWithPattern()
}
