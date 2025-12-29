// Screenshots and PDFs - HTTP Example
//
// This script demonstrates how to capture screenshots and PDFs using raw HTTP requests.
//
// Usage:
//
//	go run 01_screenshots_http.go
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	apiURL = "https://api.crawl4ai.com"
	apiKey = "YOUR_API_KEY" // Replace with your API key
)

func captureScreenshot(url string) string {
	fmt.Printf("Capturing screenshot of %s...\n", url)

	payload := map[string]interface{}{
		"url": url,
		"crawler_config": map[string]interface{}{
			"screenshot":          true,
			"screenshot_wait_for": ".content",
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL+"/v1/crawl", bytes.NewBuffer(body))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error: %d - %s", resp.StatusCode, string(bodyBytes))
		return ""
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if screenshot, ok := data["screenshot"].(string); ok && screenshot != "" {
		fmt.Printf("Screenshot captured: %d bytes (base64)\n", len(screenshot))
		return screenshot
	}

	fmt.Println("No screenshot available")
	return ""
}

func capturePDF(url string) string {
	fmt.Printf("Generating PDF of %s...\n", url)

	payload := map[string]interface{}{
		"url": url,
		"crawler_config": map[string]interface{}{
			"pdf": true,
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiURL+"/v1/crawl", bytes.NewBuffer(body))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error: %d - %s", resp.StatusCode, string(bodyBytes))
		return ""
	}

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	if pdf, ok := data["pdf"].(string); ok && pdf != "" {
		fmt.Printf("PDF generated: %d bytes (base64)\n", len(pdf))
		return pdf
	}

	fmt.Println("No PDF available")
	return ""
}

func main() {
	// Example 1: Capture screenshot
	screenshotB64 := captureScreenshot("https://www.example.com")

	// Example 2: Generate PDF
	pdfB64 := capturePDF("https://www.example.com")

	// Save screenshot to file
	if screenshotB64 != "" {
		data, _ := base64.StdEncoding.DecodeString(screenshotB64)
		os.WriteFile("screenshot.png", data, 0644)
		fmt.Println("\nScreenshot saved to screenshot.png")
	}

	// Save PDF to file
	if pdfB64 != "" {
		data, _ := base64.StdEncoding.DecodeString(pdfB64)
		os.WriteFile("page.pdf", data, 0644)
		fmt.Println("PDF saved to page.pdf")
	}
}
