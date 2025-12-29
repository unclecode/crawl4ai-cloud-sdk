// Crawl4AI Cloud - Proxy Usage Examples (Go)
//
// This file demonstrates all proxy configurations and crawl types.
//
// Proxy Modes:
// - none: Direct connection (1x credits)
// - datacenter: Fast datacenter proxies (2x credits) - NST/Scrapeless
// - residential: Premium residential IPs (5x credits) - Massive/NST/Scrapeless
// - auto: Smart selection based on target URL
//
// Usage:
//
//	go run proxy_examples.go
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
	apiBase = "https://api.crawl4ai.com"
	apiKey  = "sk_live_YOUR_API_KEY_HERE"
)

// =============================================================================
// SYNC SINGLE CRAWL - All Proxy Modes
// =============================================================================

// crawlNoProxy crawls without proxy - direct connection.
// Cost: 1x credits (100 credits per URL)
func crawlNoProxy(url string) map[string]interface{} {
	return makeRequest(url, map[string]interface{}{
		"mode": "none",
	})
}

// crawlDatacenterProxy crawls with datacenter proxy - fast and cheap.
// Cost: 2x credits (200 credits per URL)
func crawlDatacenterProxy(url string, country string) map[string]interface{} {
	proxy := map[string]interface{}{"mode": "datacenter"}
	if country != "" {
		proxy["country"] = country // ISO code: "US", "GB", "DE", etc.
	}
	return makeRequest(url, proxy)
}

// crawlResidentialProxy crawls with residential proxy - premium IPs.
// Cost: 5x credits (500 credits per URL)
func crawlResidentialProxy(url string, country string) map[string]interface{} {
	proxy := map[string]interface{}{"mode": "residential"}
	if country != "" {
		proxy["country"] = country
	}
	return makeRequest(url, proxy)
}

// crawlAutoProxy crawls with auto proxy mode - smart selection.
// Cost: Varies based on selection (1x, 2x, or 5x)
func crawlAutoProxy(url string) map[string]interface{} {
	return makeRequest(url, map[string]interface{}{
		"mode": "auto",
	})
}

// crawlSpecificProvider forces a specific proxy provider.
func crawlSpecificProvider(url, provider, mode string) map[string]interface{} {
	return makeRequest(url, map[string]interface{}{
		"mode":     mode,
		"provider": provider,
	})
}

// =============================================================================
// BATCH CRAWL - Multiple URLs with Proxy
// =============================================================================

// batchCrawlWithProxy crawls multiple URLs with proxy.
func batchCrawlWithProxy(urls []string, mode string) map[string]interface{} {
	payload := map[string]interface{}{
		"urls":         urls,
		"proxy":        map[string]interface{}{"mode": mode},
		"bypass_cache": true,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiBase+"/v1/crawl/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// =============================================================================
// DEEP CRAWL - Multi-page Crawling with Sticky Sessions
// =============================================================================

// deepCrawlWithProxy performs a deep crawl with proxy.
func deepCrawlWithProxy(url, mode string, maxDepth, maxURLs int) map[string]interface{} {
	payload := map[string]interface{}{
		"url":       url,
		"strategy":  "bfs",
		"max_depth": maxDepth,
		"max_urls":  maxURLs,
		"proxy":     map[string]interface{}{"mode": mode},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiBase+"/v1/crawl/deep", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// deepCrawlStickySession performs a deep crawl with sticky session.
// Ensures all pages use the same proxy IP address.
func deepCrawlStickySession(url, mode string, maxDepth, maxURLs int) map[string]interface{} {
	payload := map[string]interface{}{
		"url":       url,
		"strategy":  "bfs",
		"max_depth": maxDepth,
		"max_urls":  maxURLs,
		"proxy": map[string]interface{}{
			"mode":           mode,
			"sticky_session": true, // Same IP for all URLs
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiBase+"/v1/crawl/deep", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func makeRequest(url string, proxy map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{
		"url":          url,
		"proxy":        proxy,
		"bypass_cache": true,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", apiBase+"/v1/crawl", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error: %d - %s", resp.StatusCode, string(bodyBytes))
		return nil
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

// =============================================================================
// USAGE EXAMPLES
// =============================================================================

func main() {
	// Example 1: Simple crawl without proxy
	fmt.Println("=== No Proxy ===")
	result := crawlNoProxy("https://httpbin.org/ip")
	if result != nil {
		fmt.Printf("Success: %v\n", result["success"])
		fmt.Printf("Proxy: %v\n", result["proxy_mode"])
	}

	// Example 2: Datacenter proxy
	fmt.Println("\n=== Datacenter Proxy ===")
	result = crawlDatacenterProxy("https://httpbin.org/ip", "")
	if result != nil {
		fmt.Printf("Success: %v\n", result["success"])
		fmt.Printf("Provider: %v\n", result["proxy_used"])
		fmt.Printf("Mode: %v\n", result["proxy_mode"])
	}

	// Example 3: Residential proxy for protected site
	fmt.Println("\n=== Residential Proxy ===")
	result = crawlResidentialProxy("https://httpbin.org/ip", "US")
	if result != nil {
		fmt.Printf("Success: %v\n", result["success"])
		fmt.Printf("Provider: %v\n", result["proxy_used"])
	}

	// Example 4: Auto mode
	fmt.Println("\n=== Auto Mode (Amazon) ===")
	result = crawlAutoProxy("https://amazon.com")
	if result != nil {
		fmt.Printf("Auto selected: %v\n", result["proxy_mode"])
		fmt.Printf("Provider: %v\n", result["proxy_used"])
	}

	// Example 5: Deep crawl with sticky session
	fmt.Println("\n=== Deep Crawl with Sticky Session ===")
	result = deepCrawlStickySession("https://example.com", "datacenter", 1, 3)
	if result != nil {
		fmt.Printf("Job ID: %v\n", result["job_id"])
		fmt.Printf("Status: %v\n", result["status"])
	}
}
