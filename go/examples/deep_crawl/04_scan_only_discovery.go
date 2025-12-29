// Deep Crawl - Scan-Only Mode (URL Discovery)
//
// Scan-only mode discovers URLs without crawling them. This is useful for:
// - Previewing what URLs will be crawled before committing
// - Building a URL list for later processing
// - Analyzing site structure
// - Fast HTML caching for later extraction
//
// Usage:
//
//	go run 04_scan_only_discovery.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func basicScanOnly() {
	fmt.Println("=== Scan-Only Mode (Discovery) ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// ScanOnly=true returns discovered URLs without processing
	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 2,
		MaxURLs:  20,
		ScanOnly: true, // Just discover, don't crawl
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	if result.DeepResult != nil {
		fmt.Printf("Status: %s\n", result.DeepResult.Status)
		fmt.Printf("URLs discovered: %d\n", result.DeepResult.DiscoveredCount)
		fmt.Printf("Cache expires at: %s\n", result.DeepResult.CacheExpiresAt)

		// Get list of discovered URLs
		if len(result.DeepResult.DiscoveredURLs) > 0 {
			fmt.Println("\nDiscovered URLs:")
			for i, url := range result.DeepResult.DiscoveredURLs {
				if i >= 10 {
					break
				}
				fmt.Printf("  - %s\n", url)
			}
		}
	}
}

func scanMapStrategy() {
	fmt.Println("\n=== Scan-Only with Map Strategy ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "map",
		MaxURLs:  50,
		ScanOnly: true,
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	if result.DeepResult != nil {
		fmt.Printf("Sitemap URLs found: %d\n", result.DeepResult.DiscoveredCount)

		// You can filter/review URLs before deciding to crawl
		if len(result.DeepResult.DiscoveredURLs) > 0 {
			apiURLs := 0
			guideURLs := 0
			for _, url := range result.DeepResult.DiscoveredURLs {
				if contains(url, "/api/") {
					apiURLs++
				}
				if contains(url, "/guide/") {
					guideURLs++
				}
			}

			fmt.Printf("\nAPI pages: %d\n", apiURLs)
			fmt.Printf("Guide pages: %d\n", guideURLs)

			// Save the scan job ID for later extraction
			fmt.Printf("\nScan Job ID: %s\n", result.DeepResult.JobID)
			fmt.Println("Use this ID with SourceJob to extract later!")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func main() {
	basicScanOnly()
	// Uncomment to run other examples:
	// scanMapStrategy()
}
