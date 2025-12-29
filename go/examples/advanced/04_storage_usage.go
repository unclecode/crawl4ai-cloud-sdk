// Storage Usage Monitoring - SDK Example
//
// This script demonstrates how to check and monitor your storage quota.
// Storage is used by async job results stored in S3.
//
// Usage:
//
//	go run 04_storage_usage.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY" // Replace with your API key

func checkStorageUsage() {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	fmt.Println("Checking storage usage...")
	usage, err := crawler.Storage()
	if err != nil {
		log.Fatalf("Failed to get storage usage: %v", err)
	}

	fmt.Println("\n=== STORAGE USAGE ===")
	fmt.Printf("Used: %.2f MB\n", usage.UsedMB)
	fmt.Printf("Max: %.2f MB\n", usage.MaxMB)
	fmt.Printf("Remaining: %.2f MB\n", usage.RemainingMB)
	fmt.Printf("Usage: %.1f%%\n", usage.Percentage)

	// Check if storage is getting full
	if usage.Percentage > 90 {
		fmt.Println("\nWARNING: Storage is over 90% full!")
		fmt.Println("Consider deleting old jobs to free up space.")
	} else if usage.Percentage > 75 {
		fmt.Println("\nNOTE: Storage is over 75% full.")
	} else {
		fmt.Println("\nStorage usage is healthy.")
	}
}

func monitorStorageDuringCrawl(urls []string) {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// Check initial storage
	initial, err := crawler.Storage()
	if err != nil {
		log.Fatalf("Failed to get storage usage: %v", err)
	}
	fmt.Printf("Initial storage: %.2f MB / %.2f MB\n", initial.UsedMB, initial.MaxMB)

	// Create async job
	fmt.Printf("\nStarting async crawl for %d URLs...\n", len(urls))
	result, err := crawler.RunMany(urls, &crawl4ai.RunManyOptions{
		Wait: true, // Wait for completion
	})
	if err != nil {
		log.Printf("Crawl failed: %v", err)
		return
	}

	fmt.Printf("Crawl completed: %d results\n", len(result.Results))

	// Check storage after job
	after, err := crawler.Storage()
	if err != nil {
		log.Fatalf("Failed to get storage usage: %v", err)
	}
	fmt.Printf("\nAfter crawl storage: %.2f MB / %.2f MB\n", after.UsedMB, after.MaxMB)
	fmt.Printf("Storage used by this job: %.2f MB\n", after.UsedMB-initial.UsedMB)
}

func cleanupOldJobs() {
	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// Check current storage
	usage, err := crawler.Storage()
	if err != nil {
		log.Fatalf("Failed to get storage usage: %v", err)
	}
	fmt.Printf("Current storage: %.2f MB / %.2f MB\n", usage.UsedMB, usage.MaxMB)

	// List completed jobs
	jobs, err := crawler.ListJobs(&crawl4ai.ListJobsOptions{
		Limit:  20,
		Status: "completed",
	})
	if err != nil {
		log.Fatalf("Failed to list jobs: %v", err)
	}
	fmt.Printf("\nFound %d completed jobs\n", len(jobs))

	if len(jobs) == 0 {
		fmt.Println("No jobs to delete.")
		return
	}

	// Delete oldest jobs (be careful in production!)
	fmt.Println("\nDeleting oldest 3 jobs...")
	deletedCount := 0

	// Get the last 3 (oldest with default sorting)
	start := len(jobs) - 3
	if start < 0 {
		start = 0
	}

	for _, job := range jobs[start:] {
		err := crawler.CancelJob(job.ID)
		if err != nil {
			fmt.Printf("  Failed to delete %s: %v\n", job.ID, err)
		} else {
			fmt.Printf("  Deleted job %s\n", job.ID)
			deletedCount++
		}
	}

	// Check storage after cleanup
	if deletedCount > 0 {
		after, _ := crawler.Storage()
		freed := usage.UsedMB - after.UsedMB
		fmt.Printf("\nFreed %.2f MB of storage\n", freed)
		fmt.Printf("New usage: %.2f MB / %.2f MB\n", after.UsedMB, after.MaxMB)
	}
}

func main() {
	// Example 1: Check storage
	fmt.Println("=== Example 1: Check Storage Usage ===")
	checkStorageUsage()

	// Example 2: Monitor during crawl (uncomment to use)
	// fmt.Println("\n=== Example 2: Monitor During Crawl ===")
	// monitorStorageDuringCrawl([]string{
	//     "https://www.example.com",
	//     "https://www.example.com/about",
	// })

	// Example 3: Cleanup old jobs (uncomment to use)
	// fmt.Println("\n=== Example 3: Cleanup Old Jobs ===")
	// cleanupOldJobs()
}
