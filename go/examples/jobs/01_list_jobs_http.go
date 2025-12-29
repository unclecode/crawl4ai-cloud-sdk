// Example: List jobs using HTTP API
//
// This example demonstrates:
// - GET /v1/crawl/jobs with query parameters
// - Pagination and filtering
// - Raw JSON response handling
//
// Usage:
//
//	go run 01_list_jobs_http.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	apiKey  = "YOUR_API_KEY"
	baseURL = "https://api.crawl4ai.com"
)

func main() {
	client := &http.Client{}

	// List all jobs
	fmt.Println("=== All Jobs (First 20) ===")
	req, _ := http.NewRequest("GET", baseURL+"/v1/crawl/jobs?limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	fmt.Printf("Total jobs: %v\n", data["total"])
	if jobs, ok := data["jobs"].([]interface{}); ok {
		fmt.Printf("Showing: %d\n", len(jobs))
		for _, j := range jobs {
			job := j.(map[string]interface{})
			urls, _ := job["urls"].([]interface{})
			fmt.Printf("  %v: %v | %d URLs\n", job["job_id"], job["status"], len(urls))
		}
	}

	// Filter by status
	fmt.Println("\n=== Completed Jobs ===")
	req2, _ := http.NewRequest("GET", baseURL+"/v1/crawl/jobs?status=completed&limit=10", nil)
	req2.Header.Set("Authorization", "Bearer "+apiKey)
	resp2, _ := client.Do(req2)
	var completed map[string]interface{}
	json.NewDecoder(resp2.Body).Decode(&completed)
	resp2.Body.Close()

	if jobs, ok := completed["jobs"].([]interface{}); ok {
		for _, j := range jobs {
			job := j.(map[string]interface{})
			urls, _ := job["urls"].([]interface{})
			url := "N/A"
			if len(urls) > 0 {
				url = fmt.Sprintf("%v", urls[0])
			}
			fmt.Printf("  %v: %v\n", job["job_id"], url)
		}
	}

	// Pagination
	fmt.Println("\n=== Pagination (Next 20) ===")
	req3, _ := http.NewRequest("GET", baseURL+"/v1/crawl/jobs?limit=20&offset=20", nil)
	req3.Header.Set("Authorization", "Bearer "+apiKey)
	resp3, _ := client.Do(req3)
	var page2 map[string]interface{}
	json.NewDecoder(resp3.Body).Decode(&page2)
	resp3.Body.Close()

	if jobs, ok := page2["jobs"].([]interface{}); ok {
		fmt.Printf("Page 2: %d jobs\n", len(jobs))
	}

	// Failed jobs
	fmt.Println("\n=== Failed Jobs ===")
	req4, _ := http.NewRequest("GET", baseURL+"/v1/crawl/jobs?status=failed&limit=5", nil)
	req4.Header.Set("Authorization", "Bearer "+apiKey)
	resp4, _ := client.Do(req4)
	var failed map[string]interface{}
	json.NewDecoder(resp4.Body).Decode(&failed)
	resp4.Body.Close()

	if jobs, ok := failed["jobs"].([]interface{}); ok {
		for _, j := range jobs {
			job := j.(map[string]interface{})
			errMsg := job["error"]
			if errMsg == nil {
				errMsg = "Unknown error"
			}
			fmt.Printf("  %v: %v\n", job["job_id"], errMsg)
		}
	}
}
