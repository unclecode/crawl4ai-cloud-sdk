// Batch Crawl - SDK Example
//
// This script demonstrates crawling multiple URLs at once.
// RunMany() automatically selects batch (<=10 URLs) or async (>10 URLs).
//
// Usage:
//
//	go run 02_batch_crawl_sdk.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func main() {
	// Create crawler
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// URLs to crawl
	urls := []string{
		"https://example.com",
		"https://httpbin.org/html",
		"https://httpbin.org/json",
	}

	fmt.Printf("Crawling %d URLs...\n", len(urls))

	// Crawl multiple URLs (auto-selects batch or async based on count)
	result, err := crawler.RunMany(urls, &crawl4ai.RunManyOptions{
		Strategy: "http", // Options: "browser" (JS support) or "http" (faster, no JS)
		Wait:     true,   // Wait for all to complete
	})
	if err != nil {
		log.Fatalf("Batch crawl failed: %v", err)
	}

	// Display results
	fmt.Println("\n=== BATCH CRAWL COMPLETE ===")
	fmt.Printf("Total URLs: %d\n", len(result.Results))

	succeeded := 0
	for _, r := range result.Results {
		if r.Success {
			succeeded++
		}
	}
	fmt.Printf("Succeeded: %d\n", succeeded)
	fmt.Printf("Failed: %d\n", len(result.Results)-succeeded)

	// Show individual results
	for i, r := range result.Results {
		fmt.Printf("\n[%d] %s\n", i+1, r.URL)
		fmt.Printf("    Status: %d\n", r.StatusCode)
		if r.Success && r.Markdown != nil {
			preview := r.Markdown.RawMarkdown
			if len(preview) > 100 {
				preview = preview[:100]
			}
			fmt.Printf("    Preview: %s...\n", preview)
		} else {
			fmt.Printf("    Error: %s\n", r.ErrorMessage)
		}
	}
}
