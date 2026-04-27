package crawl4ai

// SDK 0.7.0 changes — real e2e tests against stage. No mocks.
//
// Covers:
//   - crawler.Scrape() — new canonical name, /v1/scrape
//   - crawler.Markdown() — alias, also routes to /v1/scrape
//   - crawler.ScrapeAsync() / ScreenshotAsync() / ExtractAsync() — new in 0.7
//   - ExtractAsync(url, &ExtractAsyncOptions{ExtraURLs: ...}) shape
//   - Sources field on Map() + Scan() + legacy Mode translation
//   - Composable scan + scrape/extract chain

import (
	"strings"
	"testing"
	"time"
)

func newV07Crawler(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	c, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  wrapperTestKey,
		BaseURL: getBaseURL(),
	})
	if err != nil {
		t.Fatalf("crawler init: %v", err)
	}
	return c
}

// =============================================================================
// Scrape — new canonical name
// =============================================================================

func TestV07_Scrape_Basic(t *testing.T) {
	c := newV07Crawler(t)
	r, err := c.Scrape("https://example.com", &MarkdownOptions{Strategy: "http"})
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	if !r.Success {
		t.Fatalf("scrape returned success=false")
	}
	if !strings.Contains(r.Markdown, "Example") {
		t.Fatalf("expected markdown to contain 'Example', got %q", r.Markdown[:80])
	}
}

func TestV07_Scrape_WithInclude(t *testing.T) {
	c := newV07Crawler(t)
	r, err := c.Scrape("https://example.com", &MarkdownOptions{
		Strategy: "http",
		Include:  []string{"links", "metadata"},
	})
	if err != nil {
		t.Fatalf("scrape: %v", err)
	}
	if r.Links == nil {
		t.Fatalf("expected links to populate when included")
	}
}

func TestV07_Markdown_Alias_StillWorks(t *testing.T) {
	c := newV07Crawler(t)
	r, err := c.Markdown("https://example.com", &MarkdownOptions{Strategy: "http"})
	if err != nil {
		t.Fatalf("markdown alias broke: %v", err)
	}
	if !r.Success {
		t.Fatalf("markdown alias success=false")
	}
}

// =============================================================================
// New async batch wrappers — didn't exist before 0.7
// =============================================================================

func TestV07_ScrapeAsync_Batch(t *testing.T) {
	c := newV07Crawler(t)
	job, err := c.ScrapeAsync(
		[]string{"https://example.com", "https://httpbin.org/html"},
		&ScrapeAsyncOptions{
			MarkdownOptions: MarkdownOptions{Strategy: "http"},
			Wait:            true,
			Timeout:         2 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("scrapeAsync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
}

func TestV07_ScreenshotAsync_Batch(t *testing.T) {
	c := newV07Crawler(t)
	job, err := c.ScreenshotAsync(
		[]string{"https://example.com"},
		&ScreenshotAsyncOptions{
			Wait:    true,
			Timeout: 2 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("screenshotAsync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
}

func TestV07_ExtractAsync_BaseUrlAuto(t *testing.T) {
	c := newV07Crawler(t)
	job, err := c.ExtractAsync("https://example.com", &ExtractAsyncOptions{
		ExtractOptions: ExtractOptions{Method: "auto"},
		Wait:           true,
		Timeout:        2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("extractAsync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
}

func TestV07_ExtractAsync_UrlPlusExtras(t *testing.T) {
	c := newV07Crawler(t)
	job, err := c.ExtractAsync("https://example.com", &ExtractAsyncOptions{
		ExtractOptions: ExtractOptions{
			Method: "llm",
			Query:  "summarize the page",
		},
		ExtraURLs: []string{"https://httpbin.org/html"},
		Wait:      true,
		Timeout:   3 * time.Minute,
	})
	if err != nil {
		t.Fatalf("extractAsync url+extras: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
}

// =============================================================================
// Sources field on Map + Scan
// =============================================================================

func TestV07_Map_Sources_Primary(t *testing.T) {
	c := newV07Crawler(t)
	maxUrls := 5
	r, err := c.Map("https://www.python.org", &MapOptions{
		Sources: "primary",
		MaxURLs: &maxUrls,
	})
	if err != nil {
		t.Fatalf("map: %v", err)
	}
	if !r.Success {
		t.Fatal("map success=false")
	}
	if r.TotalUrls == 0 {
		t.Fatal("expected at least 1 URL")
	}
}

func TestV07_Map_LegacyMode_Translates(t *testing.T) {
	c := newV07Crawler(t)
	maxUrls := 5
	// mode="default" → translates to sources="primary"
	r, err := c.Map("https://www.python.org", &MapOptions{
		Mode:    "default",
		MaxURLs: &maxUrls,
	})
	if err != nil {
		t.Fatalf("map mode=default: %v", err)
	}
	if !r.Success {
		t.Fatal("map success=false")
	}
}

func TestV07_Scan_Sources_Primary(t *testing.T) {
	c := newV07Crawler(t)
	r, err := c.Scan("https://www.python.org", &ScanOptions{
		Sources: "primary",
		MaxUrls: 5,
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if r.TotalUrls == 0 {
		t.Fatal("expected at least 1 URL")
	}
}

// =============================================================================
// Composable chain
// =============================================================================

func TestV07_Chain_ScanThenScrapeAsync(t *testing.T) {
	c := newV07Crawler(t)
	scan, err := c.Scan("https://www.python.org", &ScanOptions{
		Sources: "primary",
		MaxUrls: 3,
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	urls := make([]string, 0, len(scan.Urls))
	for _, u := range scan.Urls {
		urls = append(urls, u.URL)
		if len(urls) >= 3 {
			break
		}
	}
	if len(urls) == 0 {
		t.Fatal("scan returned no URLs")
	}

	job, err := c.ScrapeAsync(urls, &ScrapeAsyncOptions{
		MarkdownOptions: MarkdownOptions{Strategy: "http"},
		Wait:            true,
		Timeout:         2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("scrapeAsync: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
}

func TestV07_Chain_ScanThenExtractAsync(t *testing.T) {
	c := newV07Crawler(t)
	scan, err := c.Scan("https://www.python.org", &ScanOptions{
		Sources: "primary",
		MaxUrls: 3,
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(scan.Urls) == 0 {
		t.Fatal("scan returned no URLs")
	}
	base := scan.Urls[0].URL
	var rest []string
	for _, u := range scan.Urls[1:] {
		rest = append(rest, u.URL)
		if len(rest) >= 2 {
			break
		}
	}

	job, err := c.ExtractAsync(base, &ExtractAsyncOptions{
		ExtractOptions: ExtractOptions{Method: "auto"},
		ExtraURLs:      rest,
		Wait:           true,
		Timeout:        4 * time.Minute,
	})
	if err != nil {
		t.Fatalf("extractAsync chain: %v", err)
	}
	if job.Status != "completed" {
		t.Fatalf("expected completed, got %s", job.Status)
	}
}
