// Example: Job management using SDK
//
// This example demonstrates:
// - Getting job details
// - Cancelling running jobs
// - Waiting for job completion
//
// Usage:
//
//	go run 02_job_management_sdk.go
package main

import (
	"fmt"
	"log"
	"time"

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

	// Create an async job for testing
	fmt.Println("=== Creating Test Job ===")
	result, err := crawler.RunMany(
		[]string{"https://example.com", "https://example.org"},
		&crawl4ai.RunManyOptions{Wait: false}, // Don't wait, just create the job
	)
	if err != nil {
		log.Fatalf("Failed to create job: %v", err)
	}

	fmt.Printf("Created job: %s\n", result.Job.ID)

	// Get job details
	fmt.Println("\n=== Get Job Details ===")
	job, err := crawler.GetJob(result.Job.ID, false)
	if err != nil {
		log.Fatalf("Failed to get job: %v", err)
	}

	fmt.Printf("Job ID: %s\n", job.ID)
	fmt.Printf("Status: %s\n", job.Status)
	fmt.Printf("URLs: %d\n", job.URLsCount)
	fmt.Printf("Created: %s\n", job.CreatedAt)

	// Wait for job and get results
	fmt.Println("\n=== Wait for Job ===")
	completedJob, err := crawler.WaitJob(
		result.Job.ID,
		2*time.Second,   // Poll interval
		5*time.Minute,   // Timeout
		true,            // Include results
	)
	if err != nil {
		log.Fatalf("Failed to wait for job: %v", err)
	}

	fmt.Printf("Final Status: %s\n", completedJob.Status)
	fmt.Printf("Progress: %d/%d\n", completedJob.Progress.Completed, completedJob.Progress.Total)

	// Cancel a job (create another one first)
	fmt.Println("\n=== Cancel Job ===")
	result2, _ := crawler.RunMany(
		[]string{"https://example.com"},
		&crawl4ai.RunManyOptions{Wait: false},
	)
	fmt.Printf("Created job: %s\n", result2.Job.ID)

	err = crawler.CancelJob(result2.Job.ID)
	if err != nil {
		fmt.Printf("Cancel failed: %v\n", err)
	} else {
		fmt.Println("Job cancelled successfully!")
	}
}
