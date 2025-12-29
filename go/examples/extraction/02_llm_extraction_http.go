// LLM Extraction with HTTP
//
// This example shows how to extract structured data using LLM
// via direct HTTP API calls (no SDK).
//
// Usage:
//
//	go run 02_llm_extraction_http.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	apiURL = "https://api.crawl4ai.com"
	apiKey = "YOUR_API_KEY" // Replace with your API key
)

func main() {
	// Define LLM extraction strategy
	payload := map[string]interface{}{
		"url":      "https://news.ycombinator.com",
		"strategy": "http",
		"crawler_config": map[string]interface{}{
			"extraction_strategy": map[string]interface{}{
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
		},
	}

	fmt.Println("Crawling Hacker News with LLM extraction...")

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL+"/v1/crawl", bytes.NewBuffer(body))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("Error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if stories, ok := result["extracted_content"].([]interface{}); ok {
		fmt.Printf("\nExtracted %d stories\n", len(stories))
		fmt.Println("\nFirst 3 stories:")
		for i, s := range stories {
			if i >= 3 {
				break
			}
			story := s.(map[string]interface{})
			fmt.Printf("\nTitle: %v\n", story["title"])
			fmt.Printf("URL: %v\n", story["url"])
			fmt.Printf("Points: %v\n", story["points"])
			fmt.Printf("Author: %v\n", story["author"])
			fmt.Printf("Comments: %v\n", story["comments"])
		}
	}
}
