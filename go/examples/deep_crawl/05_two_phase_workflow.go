// Deep Crawl - Two-Phase Workflow (Scan -> Extract)
//
// The two-phase workflow separates URL discovery from data extraction:
//
// Phase 1 (Scan):
//   - Discover URLs using BFS/DFS/Best-First
//   - Cache raw HTML in Redis (30-minute TTL)
//   - Return scan_job_id for later use
//
// Phase 2 (Extract):
//   - Use SourceJob to reference cached HTML
//   - Apply extraction strategy (CSS, LLM, etc.)
//   - No re-crawling needed - uses cached HTML
//
// Usage:
//
//	go run 05_two_phase_workflow.go
package main

import (
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func basicTwoPhase() {
	fmt.Println("=== Two-Phase Workflow ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// ========== PHASE 1: SCAN ==========
	fmt.Println("Phase 1: Scanning (URL discovery + HTML caching)...")

	scanResult, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 2,
		MaxURLs:  10,
		ScanOnly: true, // Don't extract yet
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	fmt.Printf("  Scan Job ID: %s\n", scanResult.DeepResult.JobID)
	fmt.Printf("  URLs discovered: %d\n", scanResult.DeepResult.DiscoveredCount)
	fmt.Printf("  Cache expires: %s\n", scanResult.DeepResult.CacheExpiresAt)
	fmt.Println()

	// ========== PHASE 2: EXTRACT ==========
	fmt.Println("Phase 2: Extracting from cached HTML...")

	extractResult, err := crawler.DeepCrawl("", &crawl4ai.DeepCrawlOptions{
		SourceJob: scanResult.DeepResult.JobID, // Use cached HTML
		Config: &crawl4ai.CrawlerRunConfig{
			ExtractionStrategy: map[string]interface{}{
				"type": "json_css",
				"schema": map[string]interface{}{
					"name":         "PageContent",
					"baseSelector": "main, article, .content",
					"fields": []map[string]interface{}{
						{"name": "title", "selector": "h1", "type": "text"},
						{"name": "headings", "selector": "h2, h3", "type": "list"},
					},
				},
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Extract failed: %v", err)
	}

	if extractResult.CrawlJob != nil {
		fmt.Printf("  Job ID: %s\n", extractResult.CrawlJob.ID)
		fmt.Printf("  Pages extracted: %d\n", extractResult.CrawlJob.Progress.Completed)
	}
}

func multipleExtractions() {
	fmt.Println("\n=== Multiple Extractions from Same Scan ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// SCAN ONCE
	fmt.Println("Scanning...")
	scanResult, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 1,
		MaxURLs:  5,
		ScanOnly: true,
		Wait:     true,
	})
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}

	scanJobID := scanResult.DeepResult.JobID
	fmt.Printf("Scan Job ID: %s\n", scanJobID)
	fmt.Printf("URLs cached: %d\n\n", scanResult.DeepResult.DiscoveredCount)

	// EXTRACT #1: Titles only
	fmt.Println("Extraction 1: Titles...")
	job1, _ := crawler.DeepCrawl("", &crawl4ai.DeepCrawlOptions{
		SourceJob: scanJobID,
		Config: &crawl4ai.CrawlerRunConfig{
			ExtractionStrategy: map[string]interface{}{
				"type": "json_css",
				"schema": map[string]interface{}{
					"name":         "Titles",
					"baseSelector": "body",
					"fields": []map[string]interface{}{
						{"name": "title", "selector": "h1", "type": "text"},
					},
				},
			},
		},
		Wait: true,
	})
	if job1.CrawlJob != nil {
		fmt.Printf("  Extracted: %d pages\n", job1.CrawlJob.Progress.Completed)
	}

	// EXTRACT #2: Links
	fmt.Println("Extraction 2: Links...")
	job2, _ := crawler.DeepCrawl("", &crawl4ai.DeepCrawlOptions{
		SourceJob: scanJobID,
		Config: &crawl4ai.CrawlerRunConfig{
			ExtractionStrategy: map[string]interface{}{
				"type": "json_css",
				"schema": map[string]interface{}{
					"name":         "Links",
					"baseSelector": "body",
					"fields": []map[string]interface{}{
						{"name": "links", "selector": "a[href]", "type": "list", "attribute": "href"},
					},
				},
			},
		},
		Wait: true,
	})
	if job2.CrawlJob != nil {
		fmt.Printf("  Extracted: %d pages\n", job2.CrawlJob.Progress.Completed)
	}

	// EXTRACT #3: Full markdown
	fmt.Println("Extraction 3: Markdown...")
	job3, _ := crawler.DeepCrawl("", &crawl4ai.DeepCrawlOptions{
		SourceJob: scanJobID,
		// No extraction strategy = get markdown
		Wait: true,
	})
	if job3.CrawlJob != nil {
		fmt.Printf("  Extracted: %d pages\n", job3.CrawlJob.Progress.Completed)
	}

	fmt.Println("\nAll 3 extractions used the same cached HTML!")
}

func main() {
	basicTwoPhase()
	// Uncomment to run other examples:
	// multipleExtractions()
}
