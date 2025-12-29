// Batch Crawl - HTTP Example
//
// This script demonstrates crawling multiple URLs (up to 10) using direct HTTP requests.
// The batch endpoint processes URLs sequentially and returns all results.
//
// Usage:
//
//	go run 02_batch_crawl_http.go
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
	// URLs to crawl (max 10)
	urls := []string{
		"https://example.com",
		"https://httpbin.org/html",
		"https://httpbin.org/json",
	}

	fmt.Printf("Crawling %d URLs in batch...\n", len(urls))

	// Build request body
	payload := map[string]interface{}{
		"urls":     urls,
		"strategy": "http", // Options: "browser" (JS support) or "http" (faster, no JS)
	}

	body, _ := json.Marshal(payload)

	// Make the batch crawl request
	req, _ := http.NewRequest("POST", apiURL+"/v1/crawl/batch", bytes.NewBuffer(body))
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
	fmt.Println("\n=== BATCH CRAWL COMPLETE ===")

	results, _ := data["results"].([]interface{})
	fmt.Printf("Total URLs: %d\n", len(results))

	succeeded := 0
	for _, r := range results {
		if result, ok := r.(map[string]interface{}); ok {
			if success, _ := result["success"].(bool); success {
				succeeded++
			}
		}
	}
	fmt.Printf("Succeeded: %d\n", succeeded)
	fmt.Printf("Failed: %d\n", len(results)-succeeded)

	// Show individual results
	for i, r := range results {
		result := r.(map[string]interface{})
		fmt.Printf("\n[%d] %v\n", i+1, result["url"])
		fmt.Printf("    Status: %v\n", result["status_code"])

		if markdown, ok := result["markdown"].(map[string]interface{}); ok {
			if raw, ok := markdown["raw_markdown"].(string); ok {
				preview := raw
				if len(preview) > 100 {
					preview = preview[:100]
				}
				fmt.Printf("    Preview: %s...\n", preview)
			}
		}
	}
}
