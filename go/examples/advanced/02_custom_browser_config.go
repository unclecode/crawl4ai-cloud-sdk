// Custom Browser Configuration - SDK Example
//
// This script demonstrates how to customize browser settings like viewport size,
// proxies, and custom headers using the Crawl4AI SDK.
//
// Usage:
//
//	go run 02_custom_browser_config.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func crawlWithCustomViewport(url string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Printf("Crawling %s with custom viewport...\n", url)

	result, err := crawler.Run(url, &crawl4ai.RunOptions{
		Strategy: "browser",
		BrowserConfig: &crawl4ai.BrowserConfig{
			Viewport: map[string]int{"width": 1920, "height": 1080},
		},
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	fmt.Printf("Success! HTML size: %d bytes\n", len(result.HTML))
}

func crawlWithProxy(url string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Printf("Crawling %s with datacenter proxy...\n", url)

	result, err := crawler.Run(url, &crawl4ai.RunOptions{
		Strategy: "browser",
		Proxy:    map[string]interface{}{"mode": "datacenter"}, // or "residential"
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	fmt.Printf("Success! HTML size: %d bytes\n", len(result.HTML))
}

func crawlWithCustomHeaders(url string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Printf("Crawling %s with custom headers...\n", url)

	result, err := crawler.Run(url, &crawl4ai.RunOptions{
		Strategy: "browser",
		BrowserConfig: &crawl4ai.BrowserConfig{
			Headers: map[string]string{
				"User-Agent":      "CustomBot/1.0 (Research purposes)",
				"Accept-Language": "en-US,en;q=0.9",
			},
		},
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	fmt.Printf("Success! HTML size: %d bytes\n", len(result.HTML))
}

func crawlWithFullConfig(url string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Printf("Crawling %s with full custom config...\n", url)

	result, err := crawler.Run(url, &crawl4ai.RunOptions{
		Strategy: "browser",
		BrowserConfig: &crawl4ai.BrowserConfig{
			Viewport: map[string]int{"width": 1920, "height": 1080},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
		},
		Config: &crawl4ai.CrawlerRunConfig{
			WaitUntil:   "networkidle", // Wait for network to be idle
			PageTimeout: 30000,         // 30 second timeout
		},
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	fmt.Printf("Success! HTML size: %d bytes\n", len(result.HTML))
}

func main() {
	// Example 1: Custom viewport
	crawlWithCustomViewport("https://www.example.com")

	// Example 2: Custom headers
	crawlWithCustomHeaders("https://www.example.com")

	// Example 3: Full configuration
	crawlWithFullConfig("https://www.example.com")

	// Example 4: With managed proxy (uncomment to use)
	// crawlWithProxy("https://www.example.com")
}
