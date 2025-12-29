// Async Crawl with Polling - HTTP Example
//
// This script demonstrates async crawling with manual polling loop.
// Create a job, then poll the status endpoint until completion.
//
// Usage:
//
//	go run 03_async_crawl_http.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	apiKey = "YOUR_API_KEY" // Replace with your API key
	apiURL = "https://api.crawl4ai.com"
)

func main() {
	// URLs to crawl (can be more than 10 for async)
	urls := []string{
		"https://example.com",
		"https://httpbin.org/html",
		"https://httpbin.org/json",
		"https://httpbin.org/robots.txt",
	}

	fmt.Printf("Creating async job for %d URLs...\n", len(urls))

	// Step 1: Create the async job
	payload := map[string]interface{}{
		"urls":     urls,
		"strategy": "http",
		"priority": 5, // Priority 1-10 (default: 5)
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", apiURL+"/v1/crawl/async", bytes.NewBuffer(body))
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

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	jobID := data["job_id"].(string)
	fmt.Printf("Job created: %s\n", jobID)
	fmt.Printf("Status: %v\n", data["status"])

	// Step 2: Poll for completion
	fmt.Println("\nPolling for completion...")
	maxAttempts := 60
	pollInterval := 2 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(pollInterval)

		statusReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/crawl/jobs/%s", apiURL, jobID), nil)
		statusReq.Header.Set("X-API-Key", apiKey)

		statusResp, err := client.Do(statusReq)
		if err != nil {
			log.Fatalf("Status request failed: %v", err)
		}

		var statusData map[string]interface{}
		json.NewDecoder(statusResp.Body).Decode(&statusData)
		statusResp.Body.Close()

		status := statusData["status"].(string)
		progress := statusData["progress"].(map[string]interface{})
		fmt.Printf("  [%d] Status: %s | Progress: %.0f/%.0f\n",
			attempt+1, status, progress["completed"], progress["total"])

		if status == "completed" || status == "partial" || status == "failed" {
			fmt.Println("\n=== JOB COMPLETE ===")
			fmt.Printf("Final status: %s\n", status)
			fmt.Printf("Results available at: /v1/crawl/jobs/%s?include_results=true\n", jobID)
			return
		}
	}

	fmt.Println("\nTimeout: Job did not complete in time")
}
