// CSS Extraction with HTTP - No LLM Cost
//
// This example shows how to extract structured data using CSS selectors
// via direct HTTP API calls (no SDK).
//
// Usage:
//
//	go run 01_css_extraction_http.go
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
	// Define CSS extraction schema
	payload := map[string]interface{}{
		"url":      "https://news.ycombinator.com",
		"strategy": "http", // Fast, no browser needed
		"crawler_config": map[string]interface{}{
			"extraction_strategy": map[string]interface{}{
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
		},
	}

	fmt.Println("Crawling Hacker News with CSS extraction...")

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
		}
	}
}
