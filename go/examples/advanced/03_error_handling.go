// Error Handling - SDK Example
//
// This script demonstrates how to properly handle errors when using the Crawl4AI SDK.
// Covers rate limits, quota errors, authentication errors, and retry logic.
//
// Usage:
//
//	go run 03_error_handling.go
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func crawlWithErrorHandling(url string) *crawl4ai.CrawlResult {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Printf("Failed to create crawler: %v", err)
		return nil
	}
	defer crawler.Close()

	fmt.Printf("Crawling %s...\n", url)
	result, err := crawler.Run(url, nil)

	if err != nil {
		// Check error type using type assertions or error wrapping
		var rateLimitErr *crawl4ai.RateLimitError
		var quotaErr *crawl4ai.QuotaExceededError
		var authErr *crawl4ai.AuthenticationError
		var validationErr *crawl4ai.ValidationError
		var notFoundErr *crawl4ai.NotFoundError
		var timeoutErr *crawl4ai.TimeoutError
		var serverErr *crawl4ai.ServerError

		switch {
		case errors.As(err, &authErr):
			fmt.Printf("Authentication failed: %v\n", err)
			fmt.Println("Check your API key and make sure it's valid")

		case errors.As(err, &rateLimitErr):
			fmt.Printf("Rate limit exceeded: %v\n", err)
			fmt.Printf("Retry after: %d seconds\n", rateLimitErr.RetryAfter)

		case errors.As(err, &quotaErr):
			fmt.Printf("Quota exceeded: %v\n", err)
			fmt.Printf("Quota type: %s\n", quotaErr.QuotaType) // 'daily', 'concurrent', or 'storage'

			switch quotaErr.QuotaType {
			case "storage":
				fmt.Println("Your storage is full. Delete old jobs to free up space.")
			case "daily":
				fmt.Println("Daily crawl limit reached. Wait until tomorrow or upgrade plan.")
			case "concurrent":
				fmt.Println("Too many concurrent requests. Wait for some to complete.")
			}

		case errors.As(err, &validationErr):
			fmt.Printf("Invalid request: %v\n", err)
			fmt.Println("Check your URL and configuration parameters")

		case errors.As(err, &notFoundErr):
			fmt.Printf("Resource not found: %v\n", err)
			fmt.Println("The job or session ID may be invalid or expired")

		case errors.As(err, &timeoutErr):
			fmt.Printf("Request timed out: %v\n", err)
			fmt.Println("The page took too long to load. Try increasing timeout.")

		case errors.As(err, &serverErr):
			fmt.Printf("Server error: %v\n", err)
			fmt.Println("The API is experiencing issues. Try again later.")

		default:
			fmt.Printf("Unexpected error: %v\n", err)
		}

		return nil
	}

	fmt.Printf("Success! HTML size: %d bytes\n", len(result.HTML))
	return result
}

func crawlWithRetryLogic(url string, maxRetries int) *crawl4ai.CrawlResult {
	for attempt := 0; attempt < maxRetries; attempt++ {
		crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
			APIKey: apiKey,
		})
		if err != nil {
			log.Printf("Failed to create crawler: %v", err)
			continue
		}

		fmt.Printf("Attempt %d/%d: Crawling %s...\n", attempt+1, maxRetries, url)
		result, err := crawler.Run(url, nil)
		crawler.Close()

		if err == nil {
			fmt.Printf("Success! HTML size: %d bytes\n", len(result.HTML))
			return result
		}

		var rateLimitErr *crawl4ai.RateLimitError
		var serverErr *crawl4ai.ServerError
		var timeoutErr *crawl4ai.TimeoutError

		switch {
		case errors.As(err, &rateLimitErr):
			if rateLimitErr.RetryAfter > 0 {
				fmt.Printf("Rate limited. Waiting %d seconds...\n", rateLimitErr.RetryAfter)
				time.Sleep(time.Duration(rateLimitErr.RetryAfter) * time.Second)
				continue
			}
			return nil

		case errors.As(err, &serverErr), errors.As(err, &timeoutErr):
			if attempt < maxRetries-1 {
				waitTime := 1 << attempt // Exponential backoff: 1, 2, 4 seconds
				fmt.Printf("Transient error: %v. Retrying in %d seconds...\n", err, waitTime)
				time.Sleep(time.Duration(waitTime) * time.Second)
				continue
			}
			fmt.Printf("Max retries reached. Last error: %v\n", err)
			return nil

		default:
			// Don't retry on other errors (auth, quota, validation)
			fmt.Printf("Non-retryable error: %v\n", err)
			return nil
		}
	}

	return nil
}

func main() {
	// Example 1: Basic error handling
	fmt.Println("=== Example 1: Basic Error Handling ===")
	crawlWithErrorHandling("https://www.example.com")

	// Example 2: Retry logic
	fmt.Println("\n=== Example 2: With Retry Logic ===")
	crawlWithRetryLogic("https://www.example.com", 3)
}
