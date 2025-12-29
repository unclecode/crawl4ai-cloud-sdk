// Schema Generation with SDK
//
// This example shows how to automatically generate CSS extraction schemas
// from HTML using LLM. The schema can then be reused for fast, no-cost extraction.
//
// Usage:
//
//	go run 03_schema_generation_sdk.go
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

	// First, get the HTML content
	fmt.Println("Fetching Hacker News HTML...")
	result, err := crawler.Run("https://news.ycombinator.com", &crawl4ai.RunOptions{
		Strategy: "http",
	})
	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	html := result.HTML
	fmt.Printf("Got %d bytes of HTML\n", len(html))

	// Generate schema using LLM
	fmt.Println("\nGenerating CSS extraction schema...")
	schemaResult, err := crawler.GenerateSchema(html, &crawl4ai.GenerateSchemaOptions{
		Query: "Extract all stories with their title, URL, points, and author",
	})
	if err != nil {
		log.Fatalf("Schema generation failed: %v", err)
	}

	if schemaResult.Error != "" {
		fmt.Printf("Error: %s\n", schemaResult.Error)
		return
	}

	fmt.Println("\nGenerated Schema:")
	schemaJSON, _ := json.MarshalIndent(schemaResult.Schema, "", "  ")
	fmt.Println(string(schemaJSON))

	// Now use the generated schema for extraction
	fmt.Println("\n\nTesting generated schema...")
	extractConfig := &crawl4ai.CrawlerRunConfig{
		ExtractionStrategy: map[string]interface{}{
			"type":   "json_css",
			"schema": schemaResult.Schema,
		},
	}

	extractResult, err := crawler.Run("https://news.ycombinator.com", &crawl4ai.RunOptions{
		Strategy: "http",
		Config:   extractConfig,
	})
	if err != nil {
		log.Fatalf("Extraction failed: %v", err)
	}

	if extractResult.Success && extractResult.ExtractedContent != "" {
		var stories []interface{}
		if err := json.Unmarshal([]byte(extractResult.ExtractedContent), &stories); err == nil {
			fmt.Printf("\nExtracted %d stories\n", len(stories))
			fmt.Println("\nFirst 2 stories:")
			for i, story := range stories {
				if i >= 2 {
					break
				}
				storyJSON, _ := json.MarshalIndent(story, "", "  ")
				fmt.Printf("\n%s\n", string(storyJSON))
			}
		}
	}
}
