package crawl4ai

import (
	"strings"
	"testing"
)

// =============================================================================
// E2E TESTS — real API calls
// =============================================================================

func TestFileDownload_CSV(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.Run(
		"https://data.gov.au/data/dataset/043f58e0-a188-4458-b61c-04e5b540aea4/resource/f83cdee9-ebcb-4f24-941b-34bb2f0996cf/download/facilities.csv",
		&RunOptions{Strategy: "http", BypassCache: true},
	)
	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Crawl not successful: %s", result.ErrorMessage)
	}
	if len(result.DownloadedFiles) == 0 {
		t.Fatal("Expected downloaded_files to have at least 1 URL")
	}
	if !strings.HasPrefix(result.DownloadedFiles[0], "https://") {
		t.Fatalf("Expected presigned URL, got: %s", result.DownloadedFiles[0])
	}
	// CSV is text-based — html also has content
	if len(result.HTML) < 1000 {
		t.Fatalf("Expected CSV content in html, got %d chars", len(result.HTML))
	}
}

func TestFileDownload_JSON(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.Run(
		"https://jsonplaceholder.typicode.com/posts/1",
		&RunOptions{Strategy: "http", BypassCache: true},
	)
	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Crawl not successful: %s", result.ErrorMessage)
	}
	if len(result.DownloadedFiles) == 0 {
		t.Fatal("Expected downloaded_files for JSON response")
	}
	if !strings.Contains(result.HTML, "userId") {
		t.Fatal("Expected JSON content in html")
	}
}

func TestFileDownload_HTMLNoDownload(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.Run(
		"https://example.com",
		&RunOptions{Strategy: "http", BypassCache: true},
	)
	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Crawl not successful: %s", result.ErrorMessage)
	}
	if len(result.DownloadedFiles) != 0 {
		t.Fatalf("Expected no downloaded_files for HTML page, got %d", len(result.DownloadedFiles))
	}
	if !strings.Contains(result.HTML, "Example Domain") {
		t.Fatal("Expected HTML content")
	}
}

func TestFileDownload_Binary(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}
	defer crawler.Close()

	result, err := crawler.Run(
		"https://httpbin.org/bytes/1024",
		&RunOptions{Strategy: "http", BypassCache: true},
	)
	if err != nil {
		t.Fatalf("Crawl failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Crawl not successful: %s", result.ErrorMessage)
	}
	if len(result.DownloadedFiles) == 0 {
		t.Fatal("Expected downloaded_files for binary response")
	}
	if !strings.HasPrefix(result.DownloadedFiles[0], "https://") {
		t.Fatalf("Expected presigned URL, got: %s", result.DownloadedFiles[0])
	}
}

// =============================================================================
// UNIT TESTS — CrawlResultFromMap parsing
// =============================================================================

func TestCrawlResultFromMap_WithDownloadedFiles(t *testing.T) {
	data := map[string]interface{}{
		"url":     "https://example.com/data.csv",
		"success": true,
		"html":    "a,b,c\n1,2,3",
		"downloaded_files": []interface{}{
			"https://s3.example.com/downloads/abc/data.csv?sig=xyz",
		},
		"status_code": float64(200),
		"duration_ms": float64(500),
	}

	result := CrawlResultFromMap(data)

	if len(result.DownloadedFiles) != 1 {
		t.Fatalf("Expected 1 downloaded file, got %d", len(result.DownloadedFiles))
	}
	if !strings.Contains(result.DownloadedFiles[0], "data.csv") {
		t.Fatalf("Expected URL containing data.csv, got: %s", result.DownloadedFiles[0])
	}
}

func TestCrawlResultFromMap_WithoutDownloadedFiles(t *testing.T) {
	data := map[string]interface{}{
		"url":         "https://example.com",
		"success":     true,
		"html":        "<html>hello</html>",
		"status_code": float64(200),
		"duration_ms": float64(100),
	}

	result := CrawlResultFromMap(data)

	if len(result.DownloadedFiles) != 0 {
		t.Fatalf("Expected no downloaded files, got %d", len(result.DownloadedFiles))
	}
}

func TestCrawlResultFromMap_MultipleFiles(t *testing.T) {
	data := map[string]interface{}{
		"url":     "https://example.com/archive",
		"success": true,
		"downloaded_files": []interface{}{
			"https://s3.example.com/file1.csv",
			"https://s3.example.com/file2.pdf",
		},
		"status_code": float64(200),
		"duration_ms": float64(200),
	}

	result := CrawlResultFromMap(data)

	if len(result.DownloadedFiles) != 2 {
		t.Fatalf("Expected 2 downloaded files, got %d", len(result.DownloadedFiles))
	}
}
