// LLM Extraction with SDK
//
// This example shows how to extract structured data using LLM with natural language instructions.
// LLM extraction is flexible and can handle complex extraction needs.
//
// Usage:
//
//	go run 02_llm_extraction_sdk.go
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

	// Define LLM extraction strategy
	config := &crawl4ai.CrawlerRunConfig{
		ExtractionStrategy: map[string]interface{}{
			"type":     "llm",
			"provider": "crawl4ai",
			"model":    "openai/gpt-4o-mini",
			"instruction": `Extract all stories from this Hacker News page.
                For each story, extract:
                - title: The story title
                - url: The story URL
                - points: Number of points (if available)
                - author: Username who posted it
                - comments: Number of comments

                Return as a JSON array of story objects.`,
		},
	}

	fmt.Println("Crawling Hacker News with LLM extraction...")
	result, err := crawler.Run("https://news.ycombinator.com", &crawl4ai.RunOptions{
		Strategy: "http",
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
				fmt.Printf("Comments: %v\n", story["comments"])
			}

			// Show token usage
			if result.LLMUsage != nil {
				fmt.Printf("\nLLM Tokens Used: %d\n", result.LLMUsage.TotalTokens)
			}
		}
	} else {
		fmt.Printf("Error: %s\n", result.ErrorMessage)
	}
}
