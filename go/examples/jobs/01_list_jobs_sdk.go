// Example: List jobs using SDK
//
// This example demonstrates:
// - Listing all jobs with pagination
// - Filtering jobs by status
// - Accessing job metadata (timestamps, URLs, status)
//
// Usage:
//
//	go run 01_list_jobs_sdk.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func main() {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// List all jobs (default: 20 per page)
	fmt.Println("=== All Jobs (First 20) ===")
	jobs, err := crawler.ListJobs(&crawl4ai.ListJobsOptions{Limit: 20})
	if err != nil {
		log.Fatalf("Failed to list jobs: %v", err)
	}

	fmt.Printf("Total jobs: %d\n", len(jobs))
	for _, job := range jobs {
		fmt.Printf("  %s: %s | %d URLs | Created: %s\n",
			job.ID, job.Status, job.URLsCount, job.CreatedAt)
	}

	// Filter by status
	fmt.Println("\n=== Completed Jobs ===")
	completed, _ := crawler.ListJobs(&crawl4ai.ListJobsOptions{
		Status: "completed",
		Limit:  10,
	})
	for _, job := range completed {
		fmt.Printf("  %s\n", job.ID)
	}

	fmt.Println("\n=== Running Jobs ===")
	running, _ := crawler.ListJobs(&crawl4ai.ListJobsOptions{Status: "running"})
	fmt.Printf("Found %d running jobs\n", len(running))

	// Pagination example
	fmt.Println("\n=== Pagination (Next 20) ===")
	page2, _ := crawler.ListJobs(&crawl4ai.ListJobsOptions{Limit: 20, Offset: 20})
	fmt.Printf("Page 2: %d jobs\n", len(page2))

	// Available statuses: pending, running, completed, failed, cancelled
	fmt.Println("\n=== Failed Jobs ===")
	failed, _ := crawler.ListJobs(&crawl4ai.ListJobsOptions{
		Status: "failed",
		Limit:  5,
	})
	for _, job := range failed {
		fmt.Printf("  %s: %s\n", job.ID, job.Error)
	}
}
