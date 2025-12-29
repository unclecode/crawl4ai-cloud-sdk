// Schema Generation with HTTP
//
// This example shows how to automatically generate CSS extraction schemas
// from HTML using LLM via direct HTTP API calls (no SDK).
//
// Usage:
//
//	go run 03_schema_generation_http.go
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
	client := &http.Client{}

	// First, get the HTML content
	fmt.Println("Fetching Hacker News HTML...")

	crawlPayload := map[string]interface{}{
		"url":      "https://news.ycombinator.com",
		"strategy": "http",
	}
	crawlBody, _ := json.Marshal(crawlPayload)

	crawlReq, _ := http.NewRequest("POST", apiURL+"/v1/crawl", bytes.NewBuffer(crawlBody))
	crawlReq.Header.Set("X-API-Key", apiKey)
	crawlReq.Header.Set("Content-Type", "application/json")

	crawlResp, err := client.Do(crawlReq)
	if err != nil {
		log.Fatalf("Crawl request failed: %v", err)
	}
	defer crawlResp.Body.Close()

	if crawlResp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(crawlResp.Body)
		log.Fatalf("Error: %d - %s", crawlResp.StatusCode, string(bodyBytes))
	}

	var crawlData map[string]interface{}
	json.NewDecoder(crawlResp.Body).Decode(&crawlData)

	html, _ := crawlData["html"].(string)
	fmt.Printf("Got %d bytes of HTML\n", len(html))

	// Generate schema using LLM
	fmt.Println("\nGenerating CSS extraction schema...")

	schemaPayload := map[string]interface{}{
		"html":        html,
		"query":       "Extract all stories with their title, URL, points, and author",
		"schema_type": "CSS",
	}
	schemaBody, _ := json.Marshal(schemaPayload)

	schemaReq, _ := http.NewRequest("POST", apiURL+"/v1/tools/schema", bytes.NewBuffer(schemaBody))
	schemaReq.Header.Set("X-API-Key", apiKey)
	schemaReq.Header.Set("Content-Type", "application/json")

	schemaResp, err := client.Do(schemaReq)
	if err != nil {
		log.Fatalf("Schema request failed: %v", err)
	}
	defer schemaResp.Body.Close()

	if schemaResp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(schemaResp.Body)
		log.Fatalf("Error: %d - %s", schemaResp.StatusCode, string(bodyBytes))
	}

	var schemaData map[string]interface{}
	json.NewDecoder(schemaResp.Body).Decode(&schemaData)

	if errMsg, ok := schemaData["error"].(string); ok && errMsg != "" {
		log.Fatalf("Error: %s", errMsg)
	}

	schema := schemaData["schema"]
	fmt.Println("\nGenerated Schema:")
	schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(schemaJSON))

	// Now use the generated schema for extraction
	fmt.Println("\n\nTesting generated schema...")

	extractPayload := map[string]interface{}{
		"url":      "https://news.ycombinator.com",
		"strategy": "http",
		"crawler_config": map[string]interface{}{
			"extraction_strategy": map[string]interface{}{
				"type":   "json_css",
				"schema": schema,
			},
		},
	}
	extractBody, _ := json.Marshal(extractPayload)

	extractReq, _ := http.NewRequest("POST", apiURL+"/v1/crawl", bytes.NewBuffer(extractBody))
	extractReq.Header.Set("X-API-Key", apiKey)
	extractReq.Header.Set("Content-Type", "application/json")

	extractResp, err := client.Do(extractReq)
	if err != nil {
		log.Fatalf("Extract request failed: %v", err)
	}
	defer extractResp.Body.Close()

	var extractData map[string]interface{}
	json.NewDecoder(extractResp.Body).Decode(&extractData)

	if stories, ok := extractData["extracted_content"].([]interface{}); ok {
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
