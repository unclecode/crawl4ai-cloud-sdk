// Deep Crawl - With Extraction Strategies
//
// Combine deep crawl with extraction strategies to get structured data
// from all discovered pages. Extraction runs during the crawl phase.
//
// Usage:
//
//	go run 07_with_extraction.go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/unclecode/crawl4ai-cloud-sdk/go/pkg/crawl4ai"
)

const apiKey = "YOUR_API_KEY"

func cssExtraction() {
	fmt.Println("=== CSS Extraction ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	// Define CSS extraction schema
	schema := map[string]interface{}{
		"name":         "Documentation",
		"baseSelector": "main, article, .content",
		"fields": []map[string]interface{}{
			{"name": "title", "selector": "h1", "type": "text"},
			{"name": "description", "selector": "p.description, .intro, meta[name='description']", "type": "text"},
			{"name": "headings", "selector": "h2, h3", "type": "list"},
			{"name": "code_blocks", "selector": "pre code, .highlight", "type": "list"},
		},
	}

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 1,
		MaxURLs:  5,
		Config: &crawl4ai.CrawlerRunConfig{
			ExtractionStrategy: map[string]interface{}{
				"type":   "json_css",
				"schema": schema,
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Pages crawled: %d\n", result.CrawlJob.Progress.Completed)

		if len(result.CrawlJob.Results) > 0 {
			fmt.Println("\nExtracted content:")
			for i, r := range result.CrawlJob.Results {
				if i >= 3 {
					break
				}
				if m, ok := r.(map[string]interface{}); ok {
					fmt.Printf("\nURL: %v\n", m["url"])
					if extractedContent, ok := m["extracted_content"].(string); ok {
						var data map[string]interface{}
						if json.Unmarshal([]byte(extractedContent), &data) == nil {
							fmt.Printf("  Title: %v\n", data["title"])
							if headings, ok := data["headings"].([]interface{}); ok && len(headings) > 0 {
								fmt.Printf("  Headings: %d\n", len(headings))
								for j, h := range headings {
									if j >= 3 {
										break
									}
									fmt.Printf("    - %v\n", h)
								}
							}
						}
					}
				}
			}
		}
	}
}

func extractionWithAttributes() {
	fmt.Println("\n=== Extract Attributes ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	schema := map[string]interface{}{
		"name":         "PageAssets",
		"baseSelector": "body",
		"fields": []map[string]interface{}{
			{"name": "links", "selector": "a[href]", "type": "list", "attribute": "href"},
			{"name": "images", "selector": "img[src]", "type": "list", "attribute": "src"},
		},
	}

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "map",
		MaxURLs:  3,
		Config: &crawl4ai.CrawlerRunConfig{
			ExtractionStrategy: map[string]interface{}{
				"type":   "json_css",
				"schema": schema,
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("Pages processed: %d\n", result.CrawlJob.Progress.Completed)

		if len(result.CrawlJob.Results) > 0 {
			for _, r := range result.CrawlJob.Results[:1] {
				if m, ok := r.(map[string]interface{}); ok {
					if extractedContent, ok := m["extracted_content"].(string); ok {
						var data map[string]interface{}
						if json.Unmarshal([]byte(extractedContent), &data) == nil {
							links, _ := data["links"].([]interface{})
							images, _ := data["images"].([]interface{})
							fmt.Printf("\nURL: %v\n", m["url"])
							fmt.Printf("  Links found: %d\n", len(links))
							fmt.Printf("  Images found: %d\n", len(images))
						}
					}
				}
			}
		}
	}
}

func llmExtraction() {
	fmt.Println("\n=== LLM Extraction ===\n")

	crawler, err := crawl4ai.NewAsyncWebCrawler(crawl4ai.CrawlerOptions{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.DeepCrawl("https://docs.crawl4ai.com", &crawl4ai.DeepCrawlOptions{
		Strategy: "map",
		MaxURLs:  3,
		Config: &crawl4ai.CrawlerRunConfig{
			ExtractionStrategy: map[string]interface{}{
				"type":     "llm",
				"provider": "openai", // or "anthropic", "ollama"
				"model":    "gpt-4o-mini",
				"schema": map[string]interface{}{
					"name": "PageSummary",
					"fields": []map[string]interface{}{
						{"name": "title", "type": "string"},
						{"name": "summary", "type": "string", "description": "2-3 sentence summary"},
						{"name": "topics", "type": "list", "description": "Main topics covered"},
						{"name": "code_examples", "type": "boolean", "description": "Has code examples?"},
					},
				},
				"instruction": "Extract the main content and summarize this documentation page.",
			},
		},
		Wait: true,
	})
	if err != nil {
		log.Fatalf("Deep crawl failed: %v", err)
	}

	if result.CrawlJob != nil {
		fmt.Printf("LLM processed: %d\n", result.CrawlJob.Progress.Completed)
	}
}

func main() {
	cssExtraction()
	// Uncomment to run other examples:
	// extractionWithAttributes()
	// llmExtraction() // Requires LLM provider setup
}
