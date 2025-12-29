// CSS Extraction with SDK - No LLM Cost
//
// This example shows how to extract structured data using CSS selectors.
// CSS extraction is fast, reliable, and has no LLM cost.
//
// Usage:
//
//	go run 01_css_extraction_sdk.go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func main() {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// Define CSS extraction schema
	config := &crawl4ai.CrawlerRunConfig{
		ExtractionStrategy: map[string]interface{}{
			"type": "json_css",
			"schema": map[string]interface{}{
				"name":         "HackerNewsStories",
				"baseSelector": ".athing",
				"fields": []map[string]interface{}{
					{"name": "title", "selector": ".titleline > a", "type": "text"},
					{"name": "url", "selector": ".titleline > a", "type": "attribute", "attribute": "href"},
					{"name": "points", "selector": "+ tr .score", "type": "text"},
					{"name": "author", "selector": "+ tr .hnuser", "type": "text"},
				},
			},
		},
	}

	fmt.Println("Crawling Hacker News with CSS extraction...")
	result, err := crawler.Run("https://news.ycombinator.com", &crawl4ai.RunOptions{
		Strategy: "http", // Fast, no browser needed
		Config:   config,
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	if result.Success && result.ExtractedContent != "" {
		var stories []map[string]interface{}
		if err := json.Unmarshal([]byte(result.ExtractedContent), &stories); err == nil {
			fmt.Printf("\nExtracted %d stories\n", len(stories))
			fmt.Println("\nFirst 3 stories:")
			for i, story := range stories {
				if i >= 3 {
					break
				}
				fmt.Printf("\nTitle: %v\n", story["title"])
				fmt.Printf("URL: %v\n", story["url"])
				fmt.Printf("Points: %v\n", story["points"])
				fmt.Printf("Author: %v\n", story["author"])
			}
		}
	} else {
		fmt.Printf("Error: %s\n", result.ErrorMessage)
	}
}
