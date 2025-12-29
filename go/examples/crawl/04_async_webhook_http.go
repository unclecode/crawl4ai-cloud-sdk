// Async Crawl with Webhook - HTTP Example
//
// This script demonstrates async crawling with webhook notification.
// Create a job with a webhook URL - the API will POST results when complete.
// No polling required!
//
// Usage:
//
//	go run 04_async_webhook_http.go
//
// Webhook Payload:
//
//	The API will POST to your webhook_url with:
//	{
//	    "job_id": "job_123",
//	    "status": "completed",
//	    "progress": {"completed": 4, "failed": 0, "total": 4},
//	    "results": [...],  // Full crawl results
//	    "created_at": "2024-01-01T00:00:00Z",
//	    "completed_at": "2024-01-01T00:01:00Z"
//	}
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
	apiKey     = "YOUR_API_KEY"                              // Replace with your API key
	apiURL     = "https://api.crawl4ai.com"
	webhookURL = "https://your-webhook-endpoint.com/callback" // Your webhook URL
)

func main() {
	// URLs to crawl (can be more than 10 for async)
	urls := []string{
		"https://example.com",
		"https://httpbin.org/html",
		"https://httpbin.org/json",
		"https://httpbin.org/robots.txt",
	}

	fmt.Printf("Creating async job for %d URLs with webhook...\n", len(urls))

	// Create async job with webhook
	payload := map[string]interface{}{
		"urls":        urls,
		"strategy":    "http",
		"webhook_url": webhookURL, // API will POST here when complete
		"priority":    7,          // Higher priority (1-10)
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

	// Display job info
	fmt.Println("\n=== JOB CREATED ===")
	fmt.Printf("Job ID: %v\n", data["job_id"])
	fmt.Printf("Status: %v\n", data["status"])
	fmt.Printf("Webhook: %s\n", webhookURL)
	fmt.Println("\nThe API will POST results to your webhook when complete.")
	fmt.Println("No polling required!")

	// You can still check status manually if needed
	fmt.Printf("\nManual status check: GET %s/v1/crawl/jobs/%v\n", apiURL, data["job_id"])
}

// Example webhook handler (Go):
/*
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	json.NewDecoder(r.Body).Decode(&data)

	fmt.Printf("Job %v completed!\n", data["job_id"])
	fmt.Printf("Status: %v\n", data["status"])

	if results, ok := data["results"].([]interface{}); ok {
		fmt.Printf("Results: %d URLs crawled\n", len(results))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

func main() {
	http.HandleFunc("/callback", webhookHandler)
	fmt.Println("Webhook handler listening on port 8000")
	http.ListenAndServe(":8000", nil)
}
*/
