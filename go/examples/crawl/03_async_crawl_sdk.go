// Async Crawl with Wait - SDK Example
//
// This script demonstrates async crawling with automatic polling (Wait=true).
// The SDK automatically polls until the job completes and returns the results.
//
// Usage:
//
//	go run 03_async_crawl_sdk.go
package main

import (
	"fmt"
	"log"
	"time"

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

	// URLs to crawl (can be more than 10 for async)
	urls := []string{
		"https://example.com",
		"https://httpbin.org/html",
		"https://httpbin.org/json",
		"https://httpbin.org/robots.txt",
	}

	fmt.Printf("Creating async job for %d URLs...\n", len(urls))

	// RunMany with Wait=true handles polling automatically
	result, err := crawler.RunMany(urls, &crawl4ai.RunManyOptions{
		Strategy:     "http",           // Options: "browser" (JS support) or "http" (faster, no JS)
		Wait:         true,             // Wait for completion (SDK polls automatically)
		PollInterval: 2 * time.Second,  // How often to check status
		Timeout:      5 * time.Minute,  // Maximum wait time
	})
	if err != nil {
		log.Fatalf("Async crawl failed: %v", err)
	}

	// Display results
	fmt.Println("\n=== JOB COMPLETE ===")
	fmt.Printf("Total: %d\n", len(result.Results))

	succeeded := 0
	for _, r := range result.Results {
		if r.Success {
			succeeded++
		}
	}
	fmt.Printf("Succeeded: %d\n", succeeded)

	// Show sample results
	fmt.Println("\nSample Results (first 3):")
	for i, r := range result.Results {
		if i >= 3 {
			break
		}
		fmt.Printf("[%d] %s\n", i+1, r.URL)
		fmt.Printf("    Status: %d\n", r.StatusCode)
		if r.Success && r.Markdown != nil {
			preview := r.Markdown.RawMarkdown
			if len(preview) > 80 {
				preview = preview[:80]
			}
			fmt.Printf("    Preview: %s...\n", preview)
		}
	}
}
