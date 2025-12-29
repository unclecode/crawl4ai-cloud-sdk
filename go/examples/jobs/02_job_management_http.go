// Example: Job management using HTTP API
//
// This example demonstrates:
// - GET /v1/crawl/jobs/{id} - Get job details
// - DELETE /v1/crawl/jobs/{id} - Cancel/delete job
//
// Usage:
//
//	go run 02_job_management_http.go
package main

import (
	"bytes"
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

	// Create a test job
	fmt.Println("=== Creating Test Job ===")
	createPayload := map[string]interface{}{
		"urls":     []string{"https://example.com", "https://example.org"},
		"priority": 5,
	}
	createBody, _ := json.Marshal(createPayload)

	createReq, _ := http.NewRequest("POST", baseURL+"/v1/crawl/async", bytes.NewBuffer(createBody))
	createReq.Header.Set("Authorization", "Bearer "+apiKey)
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := client.Do(createReq)
	if err != nil {
		log.Fatalf("Create job failed: %v", err)
	}

	var job map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&job)
	createResp.Body.Close()

	jobID := job["job_id"].(string)
	fmt.Printf("Created job: %s\n", jobID)

	// Get job details
	fmt.Println("\n=== Get Job Details ===")
	getReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/crawl/jobs/%s", baseURL, jobID), nil)
	getReq.Header.Set("Authorization", "Bearer "+apiKey)

	getResp, _ := client.Do(getReq)
	var jobDetails map[string]interface{}
	json.NewDecoder(getResp.Body).Decode(&jobDetails)
	getResp.Body.Close()

	fmt.Printf("Status: %v\n", jobDetails["status"])
	fmt.Printf("URLs: %v\n", jobDetails["urls"])

	// Cancel job (keep results)
	fmt.Println("\n=== Cancel Job (Keep Results) ===")
	cancelReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/crawl/jobs/%s?delete_results=false", baseURL, jobID), nil)
	cancelReq.Header.Set("Authorization", "Bearer "+apiKey)

	cancelResp, _ := client.Do(cancelReq)
	var cancelled map[string]interface{}
	json.NewDecoder(cancelResp.Body).Decode(&cancelled)
	cancelResp.Body.Close()

	fmt.Printf("Status: %v\n", cancelled["status"])

	// Create and delete completely
	fmt.Println("\n=== Cancel + Delete Results ===")
	create2Payload := map[string]interface{}{
		"urls": []string{"https://example.com"},
	}
	create2Body, _ := json.Marshal(create2Payload)

	create2Req, _ := http.NewRequest("POST", baseURL+"/v1/crawl/async", bytes.NewBuffer(create2Body))
	create2Req.Header.Set("Authorization", "Bearer "+apiKey)
	create2Req.Header.Set("Content-Type", "application/json")

	create2Resp, _ := client.Do(create2Req)
	var job2 map[string]interface{}
	json.NewDecoder(create2Resp.Body).Decode(&job2)
	create2Resp.Body.Close()

	job2ID := job2["job_id"].(string)
	fmt.Printf("Created job: %s\n", job2ID)

	delete2Req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/crawl/jobs/%s?delete_results=true", baseURL, job2ID), nil)
	delete2Req.Header.Set("Authorization", "Bearer "+apiKey)

	delete2Resp, _ := client.Do(delete2Req)
	var deleted map[string]interface{}
	json.NewDecoder(delete2Resp.Body).Decode(&deleted)
	delete2Resp.Body.Close()

	fmt.Printf("Deleted: %v\n", deleted["status"])
}
