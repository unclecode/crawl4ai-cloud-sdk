// Screenshots and PDFs - SDK Example
//
// This script demonstrates how to capture screenshots and PDFs using the SDK.
// Screenshots capture the visual state of the page, while PDFs generate print-ready documents.
//
// Usage:
//
//	go run 01_screenshots_sdk.go
package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func captureScreenshot(url string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Printf("Capturing screenshot of %s...\n", url)

	result, err := crawler.Run(url, &crawl4ai.RunOptions{
		Config: &crawl4ai.CrawlerRunConfig{
			Screenshot: true,
			WaitFor:    ".content", // Wait for content to load
		},
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	if result.Screenshot != "" {
		fmt.Printf("Screenshot captured: %d bytes (base64)\n", len(result.Screenshot))

		// Save to file
		data, err := base64.StdEncoding.DecodeString(result.Screenshot)
		if err != nil {
			log.Fatalf("Failed to decode screenshot: %v", err)
		}

		err = os.WriteFile("screenshot.png", data, 0644)
		if err != nil {
			log.Fatalf("Failed to save screenshot: %v", err)
		}
		fmt.Println("Screenshot saved to screenshot.png")
	} else {
		fmt.Println("No screenshot available")
	}
}

func capturePDF(url string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Printf("Generating PDF of %s...\n", url)

	result, err := crawler.Run(url, &crawl4ai.RunOptions{
		Config: &crawl4ai.CrawlerRunConfig{
			PDF: true,
		},
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	if result.PDF != "" {
		fmt.Printf("PDF generated: %d bytes (base64)\n", len(result.PDF))

		// Save to file
		data, err := base64.StdEncoding.DecodeString(result.PDF)
		if err != nil {
			log.Fatalf("Failed to decode PDF: %v", err)
		}

		err = os.WriteFile("page.pdf", data, 0644)
		if err != nil {
			log.Fatalf("Failed to save PDF: %v", err)
		}
		fmt.Println("PDF saved to page.pdf")
	} else {
		fmt.Println("No PDF available")
	}
}

func main() {
	// Example 1: Capture screenshot
	captureScreenshot("https://www.example.com")

	// Example 2: Generate PDF
	capturePDF("https://www.example.com")
}
