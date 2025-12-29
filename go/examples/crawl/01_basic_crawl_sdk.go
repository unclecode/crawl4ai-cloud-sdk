// Basic Single URL Crawl - SDK Example
//
// This script demonstrates the simplest way to crawl a single URL using the SDK.
// The Run() method blocks until the crawl completes.
//
// Usage:
//
//	go run 01_basic_crawl_sdk.go
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

	fmt.Println("Crawling https://example.com...")

	// Crawl with browser strategy (full JS support)
	result, err := crawler.Run("https://example.com", &crawl4ai.RunOptions{
		Strategy: "browser", // Options: "browser" (JS support) or "http" (faster, no JS)
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	// Display results
	fmt.Println("\n=== CRAWL COMPLETE ===")
	fmt.Printf("URL: %s\n", result.URL)
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Status: %d\n", result.StatusCode)

	if result.Markdown != nil && result.Markdown.RawMarkdown != "" {
		preview := result.Markdown.RawMarkdown
		if len(preview) > 200 {
			preview = preview[:200]
		}
		fmt.Printf("\nMarkdown preview (first 200 chars):\n%s...\n", preview)
	}

	fmt.Printf("\nHTML length: %d characters\n", len(result.HTML))
}
