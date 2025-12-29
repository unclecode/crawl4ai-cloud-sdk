// Basic Single URL Crawl - HTTP Example
//
// This script demonstrates crawling a single URL using direct HTTP requests.
// The endpoint is synchronous and blocks until the crawl completes.
//
// Usage:
//
//	go run 01_basic_crawl_http.go
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
	apiKey = "YOUR_API_KEY" // Replace with your API key
	apiURL = "https://api.crawl4ai.com"
)

func main() {
	fmt.Println("Crawling https://example.com...")

	// Build request body
	payload := map[string]interface{}{
		"url":      "https://example.com",
		"strategy": "browser", // Options: "browser" (JS support) or "http" (faster, no JS)
	}

	body, _ := json.Marshal(payload)

	// Make the crawl request
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
		log.Fatalf("HTTP Error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	// Display results
	fmt.Println("\n=== CRAWL COMPLETE ===")
	fmt.Printf("URL: %v\n", data["url"])

	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		fmt.Printf("Title: %v\n", metadata["title"])
	}

	fmt.Printf("Status: %v\n", data["status_code"])

	if markdown, ok := data["markdown"].(map[string]interface{}); ok {
		if raw, ok := markdown["raw_markdown"].(string); ok && len(raw) > 200 {
			fmt.Printf("\nMarkdown preview (first 200 chars):\n%s...\n", raw[:200])
		}
	}

	if html, ok := data["html"].(string); ok {
		fmt.Printf("\nHTML length: %d characters\n", len(html))
	}
}
