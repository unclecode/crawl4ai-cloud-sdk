// Sessions - Basic Session Management with HTTP API
//
// This script demonstrates how to create and release browser sessions using
// direct HTTP API calls. Sessions provide persistent browser instances.
//
// Usage:
//
//	go run 01_basic_session_http.go
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

	// Step 1: Create a browser session
	fmt.Println("Creating browser session...")

	createPayload := map[string]interface{}{
		"timeout": 600, // 10 minute timeout
	}
	createBody, _ := json.Marshal(createPayload)

	createReq, _ := http.NewRequest("POST", apiURL+"/v1/sessions", bytes.NewBuffer(createBody))
	createReq.Header.Set("X-API-Key", apiKey)
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := client.Do(createReq)
	if err != nil {
		log.Fatalf("Create session failed: %v", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(createResp.Body)
		log.Fatalf("Error creating session: %d - %s", createResp.StatusCode, string(bodyBytes))
	}

	var session map[string]interface{}
	json.NewDecoder(createResp.Body).Decode(&session)

	fmt.Println("\n=== SESSION CREATED ===")
	fmt.Printf("Session ID: %v\n", session["session_id"])
	fmt.Printf("WebSocket URL: %v\n", session["ws_url"])
	fmt.Printf("Expires in: %v seconds\n", session["expires_in"])

	// Step 2: Use the session (see other examples for actual usage)
	fmt.Println("\nYou can now connect to this browser using:")
	fmt.Printf("  - Crawl4AI: BrowserConfig(cdp_url='%v')\n", session["ws_url"])
	fmt.Printf("  - Puppeteer: puppeteer.connect({ browserWSEndpoint: '%v' })\n", session["ws_url"])
	fmt.Printf("  - Playwright: playwright.chromium.connectOverCDP('%v')\n", session["ws_url"])

	sessionID := session["session_id"].(string)

	// Step 3: Get session status
	fmt.Println("\nChecking session status...")

	statusReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/sessions/%s", apiURL, sessionID), nil)
	statusReq.Header.Set("X-API-Key", apiKey)

	statusResp, err := client.Do(statusReq)
	if err == nil && statusResp.StatusCode == 200 {
		var status map[string]interface{}
		json.NewDecoder(statusResp.Body).Decode(&status)
		statusResp.Body.Close()
		fmt.Printf("Session status: %v\n", status["status"])
		fmt.Printf("Worker ID: %v\n", status["worker_id"])
	}

	// Step 4: Release the session
	fmt.Println("\nReleasing session...")

	deleteReq, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/v1/sessions/%s", apiURL, sessionID), nil)
	deleteReq.Header.Set("X-API-Key", apiKey)

	deleteResp, err := client.Do(deleteReq)
	if err == nil && deleteResp.StatusCode == 200 {
		fmt.Println("Session released successfully!")
	} else {
		fmt.Printf("Failed to release session: %d\n", deleteResp.StatusCode)
	}
	if deleteResp != nil {
		deleteResp.Body.Close()
	}
}
