package crawl4ai

import (
	"os"
	"strings"
	"testing"
	"time"
)

var wrapperTestKey = getWrapperTestKey()

func getWrapperTestKey() string {
	if key := os.Getenv("CRAWL4AI_API_KEY"); key != "" {
		return key
	}
	return "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ"
}

func getBaseURL() string {
	if u := os.Getenv("CRAWL4AI_BASE_URL"); u != "" {
		return u
	}
	return "https://stage.crawl4ai.com"
}

func setupWrapper(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	c, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: wrapperTestKey, BaseURL: getBaseURL()})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return c
}

// =============================================================================
// MARKDOWN
// =============================================================================

func TestWrapperMarkdown(t *testing.T) {
	c := setupWrapper(t)

	t.Run("basic", func(t *testing.T) {
		md, err := c.Markdown("https://example.com", &MarkdownOptions{Strategy: "http"})
		if err != nil {
			t.Fatalf("Markdown: %v", err)
		}
		if !md.Success {
			t.Fatalf("expected success, got error: %s", md.ErrorMessage)
		}
		if len(md.Markdown) < 50 {
			t.Fatalf("expected markdown > 50 chars, got %d", len(md.Markdown))
		}
	})

	t.Run("include_fields", func(t *testing.T) {
		md, err := c.Markdown("https://books.toscrape.com", &MarkdownOptions{
			Strategy: "http",
			Include:  []string{"links", "media", "metadata"},
		})
		if err != nil {
			t.Fatalf("Markdown: %v", err)
		}
		if md.Links == nil {
			t.Error("expected links")
		}
		if md.Media == nil {
			t.Error("expected media")
		}
		if md.Metadata == nil {
			t.Error("expected metadata")
		}
	})

	t.Run("credits", func(t *testing.T) {
		md, err := c.Markdown("https://example.com", &MarkdownOptions{Strategy: "http"})
		if err != nil {
			t.Fatalf("Markdown: %v", err)
		}
		if md.Usage == nil || md.Usage.CreditsUsed <= 0 {
			t.Error("expected credits_used > 0")
		}
	})

	t.Run("crawler_config", func(t *testing.T) {
		md, err := c.Markdown("https://books.toscrape.com", &MarkdownOptions{
			Strategy:      "browser",
			CrawlerConfig: map[string]interface{}{"css_selector": "article.product_pod"},
		})
		if err != nil {
			t.Fatalf("Markdown: %v", err)
		}
		if !md.Success {
			t.Fatalf("expected success: %s", md.ErrorMessage)
		}
	})
}

// =============================================================================
// SCREENSHOT
// =============================================================================

func TestWrapperScreenshot(t *testing.T) {
	c := setupWrapper(t)

	t.Run("basic", func(t *testing.T) {
		ss, err := c.Screenshot("https://example.com", nil)
		if err != nil {
			t.Fatalf("Screenshot: %v", err)
		}
		if !ss.Success {
			t.Fatalf("expected success: %s", ss.ErrorMessage)
		}
		if len(ss.Screenshot) < 1000 {
			t.Fatalf("expected screenshot > 1000 chars, got %d", len(ss.Screenshot))
		}
	})

	t.Run("pdf", func(t *testing.T) {
		ss, err := c.Screenshot("https://example.com", &ScreenshotOptions{PDF: true})
		if err != nil {
			t.Fatalf("Screenshot: %v", err)
		}
		if len(ss.PDF) < 1000 {
			t.Fatalf("expected pdf > 1000 chars, got %d", len(ss.PDF))
		}
	})
}

// =============================================================================
// EXTRACT
// =============================================================================

func TestWrapperExtract(t *testing.T) {
	c := setupWrapper(t)

	t.Run("auto", func(t *testing.T) {
		data, err := c.Extract("https://books.toscrape.com", &ExtractOptions{
			Query: "extract all books with title and price",
		})
		if err != nil {
			t.Fatalf("Extract: %v", err)
		}
		if !data.Success {
			t.Fatalf("expected success: %s", data.ErrorMessage)
		}
		if len(data.Data) == 0 {
			t.Fatal("expected data items")
		}
	})

	t.Run("llm", func(t *testing.T) {
		data, err := c.Extract("https://example.com", &ExtractOptions{
			Method: "llm", Query: "what is this page about",
		})
		if err != nil {
			t.Fatalf("Extract: %v", err)
		}
		if data.MethodUsed != "llm" {
			t.Fatalf("expected method_used=llm, got %s", data.MethodUsed)
		}
	})
}

// =============================================================================
// MAP
// =============================================================================

func TestWrapperMap(t *testing.T) {
	c := setupWrapper(t)

	t.Run("basic", func(t *testing.T) {
		maxUrls := 10
		result, err := c.Map("https://crawl4ai.com", &MapOptions{MaxURLs: &maxUrls})
		if err != nil {
			t.Fatalf("Map: %v", err)
		}
		if !result.Success {
			t.Fatalf("expected success: %s", result.ErrorMessage)
		}
		if result.TotalUrls == 0 {
			t.Fatal("expected urls found")
		}
		if result.Domain != "crawl4ai.com" {
			t.Fatalf("expected domain crawl4ai.com, got %s", result.Domain)
		}
	})
}

// =============================================================================
// SITE CRAWL
// =============================================================================

func TestWrapperCrawlSite(t *testing.T) {
	c := setupWrapper(t)

	t.Run("basic", func(t *testing.T) {
		result, err := c.CrawlSite("https://books.toscrape.com", &SiteCrawlOptions{
			MaxPages: 3, Strategy: "http",
		})
		if err != nil {
			t.Fatalf("CrawlSite: %v", err)
		}
		if result.JobID == "" {
			t.Fatal("expected job_id")
		}
	})

	t.Run("with_criteria", func(t *testing.T) {
		// AI-assisted: criteria triggers LLM config generation
		result, err := c.CrawlSite("https://books.toscrape.com", &SiteCrawlOptions{
			Criteria: "book listing pages",
			MaxPages: 3,
			Strategy: "http",
		})
		if err != nil {
			t.Fatalf("CrawlSite with criteria: %v", err)
		}
		if result.JobID == "" {
			t.Fatal("expected job_id")
		}
		if result.GeneratedConfig == nil {
			t.Fatal("expected generated_config when criteria is set")
		}
		if result.GeneratedConfig.Reasoning == "" {
			t.Fatal("expected non-empty reasoning")
		}
		if result.GeneratedConfig.Fallback {
			t.Fatal("expected fallback=false on stage")
		}
	})

	t.Run("with_criteria_and_extract_flagship", func(t *testing.T) {
		// Flagship: criteria + extract -> schema generated from sample URL
		result, err := c.CrawlSite("https://books.toscrape.com", &SiteCrawlOptions{
			Criteria: "book listing pages",
			MaxPages: 3,
			Strategy: "http",
			Extract: &SiteExtractConfig{
				Query:       "book title and price",
				JSONExample: map[string]interface{}{"title": "...", "price": "£0.00"},
				Method:      "auto",
			},
		})
		if err != nil {
			t.Fatalf("CrawlSite flagship: %v", err)
		}
		if result.JobID == "" {
			t.Fatal("expected job_id")
		}
		if result.GeneratedConfig == nil {
			t.Fatal("expected generated_config")
		}
		if result.ExtractionMethodUsed != "css_schema" && result.ExtractionMethodUsed != "llm" {
			t.Fatalf("expected extraction_method_used, got %q", result.ExtractionMethodUsed)
		}
		if result.ExtractionMethodUsed == "css_schema" {
			if result.SchemaUsed == nil {
				t.Fatal("expected schema_used when method is css_schema")
			}
			if _, ok := result.SchemaUsed["fields"]; !ok {
				t.Fatal("expected 'fields' in schema_used")
			}
		}
	})

	t.Run("scan_config_struct", func(t *testing.T) {
		threshold := 0.2
		result, err := c.CrawlSite("https://books.toscrape.com", &SiteCrawlOptions{
			MaxPages: 3,
			Strategy: "http",
			Scan: &SiteScanConfig{
				Mode:           "map",
				Patterns:       []string{"*/catalogue/*"},
				ScoreThreshold: &threshold,
			},
		})
		if err != nil {
			t.Fatalf("CrawlSite with scan config: %v", err)
		}
		if result.JobID == "" {
			t.Fatal("expected job_id")
		}
	})

	t.Run("unified_polling", func(t *testing.T) {
		result, err := c.CrawlSite("https://books.toscrape.com", &SiteCrawlOptions{
			Criteria: "book listings",
			MaxPages: 3,
			Strategy: "http",
		})
		if err != nil {
			t.Fatalf("CrawlSite: %v", err)
		}

		// Poll up to 5 times to see phase transition
		for i := 0; i < 5; i++ {
			status, err := c.GetSiteCrawlJob(result.JobID)
			if err != nil {
				t.Fatalf("GetSiteCrawlJob: %v", err)
			}
			if status.JobID != result.JobID {
				t.Fatalf("job id mismatch: %s != %s", status.JobID, result.JobID)
			}
			phaseOK := status.Phase == "scan" || status.Phase == "crawl" || status.Phase == "done"
			if !phaseOK {
				t.Fatalf("unexpected phase: %q", status.Phase)
			}
			if status.IsComplete() {
				break
			}
			time.Sleep(3 * time.Second)
		}
	})

	t.Run("include_without_markdown", func(t *testing.T) {
		// include=[links] (no markdown) -> backend computes exclude_fields=[markdown]
		result, err := c.CrawlSite("https://books.toscrape.com", &SiteCrawlOptions{
			Criteria: "book listings",
			MaxPages: 3,
			Strategy: "http",
			Include:  []string{"links"},
		})
		if err != nil {
			t.Fatalf("CrawlSite include without markdown: %v", err)
		}
		if result.JobID == "" {
			t.Fatal("expected job_id")
		}
	})
}

// =============================================================================
// SCAN (AI-assisted)
// =============================================================================

func TestWrapperScan(t *testing.T) {
	c := setupWrapper(t)

	t.Run("basic_legacy", func(t *testing.T) {
		// No criteria -> plain DomainMapper sync flow
		result, err := c.Scan("https://crawl4ai.com", &ScanOptions{MaxUrls: 10})
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if !result.Success {
			t.Fatal("expected success")
		}
		if result.TotalUrls == 0 {
			t.Fatal("expected total_urls > 0")
		}
		if result.Domain != "crawl4ai.com" {
			t.Fatalf("unexpected domain: %q", result.Domain)
		}
	})

	t.Run("with_criteria", func(t *testing.T) {
		result, err := c.Scan("https://docs.crawl4ai.com", &ScanOptions{
			Criteria: "API reference and core documentation pages",
			MaxUrls:  20,
		})
		if err != nil {
			t.Fatalf("Scan with criteria: %v", err)
		}
		if !result.Success {
			t.Fatal("expected success")
		}
		if result.ModeUsed != "map" && result.ModeUsed != "deep" {
			t.Fatalf("unexpected mode_used: %q", result.ModeUsed)
		}
		if result.GeneratedConfig == nil {
			t.Fatal("expected generated_config when criteria is set")
		}
		if result.GeneratedConfig.Reasoning == "" {
			t.Fatal("expected reasoning")
		}
	})

	t.Run("with_scan_overrides", func(t *testing.T) {
		result, err := c.Scan("https://docs.crawl4ai.com", &ScanOptions{
			Criteria: "documentation pages",
			Scan:     &SiteScanConfig{Patterns: []string{"*/core/*"}},
			MaxUrls:  10,
		})
		if err != nil {
			t.Fatalf("Scan with overrides: %v", err)
		}
		if !result.Success {
			t.Fatal("expected success")
		}
		if result.ModeUsed != "map" {
			t.Fatalf("expected map mode, got %q", result.ModeUsed)
		}
	})

	t.Run("scan_config_struct", func(t *testing.T) {
		cfg := &SiteScanConfig{Mode: "map", Patterns: []string{"*/docs/*"}}
		result, err := c.Scan("https://docs.crawl4ai.com", &ScanOptions{
			Scan:    cfg,
			MaxUrls: 10,
		})
		if err != nil {
			t.Fatalf("Scan: %v", err)
		}
		if !result.Success {
			t.Fatal("expected success")
		}
	})

	t.Run("deep_mode_async", func(t *testing.T) {
		maxDepth := 1
		result, err := c.Scan("https://httpbin.org", &ScanOptions{
			Scan:    &SiteScanConfig{Mode: "deep", MaxDepth: &maxDepth},
			MaxUrls: 5,
		})
		if err != nil {
			t.Fatalf("Scan deep: %v", err)
		}
		if !result.IsAsync() {
			t.Fatalf("expected async response, got mode_used=%q job_id=%q", result.ModeUsed, result.JobID)
		}
		if result.JobID == "" {
			t.Fatal("expected job_id for deep mode")
		}
		// Poll once
		job, err := c.GetScanJob(result.JobID)
		if err != nil {
			t.Fatalf("GetScanJob: %v", err)
		}
		if job.JobID != result.JobID {
			t.Fatalf("job id mismatch: %s != %s", job.JobID, result.JobID)
		}
		// Cancel so we don't hog worker slots
		cancelled, err := c.CancelScanJob(result.JobID)
		if err != nil {
			t.Fatalf("CancelScanJob: %v", err)
		}
		if cancelled.JobID != result.JobID {
			t.Fatalf("cancel job id mismatch")
		}
	})
}

// =============================================================================
// JOB CANCEL + NAMESPACE
// =============================================================================

func TestWrapperJobCancel(t *testing.T) {
	c := setupWrapper(t)

	// Can't easily do async in Go tests without goroutines, but we can test
	// the cancel flow: the API should return success even if job already done
	t.Run("cancel_nonexistent", func(t *testing.T) {
		err := c.CancelMarkdownJob("job_doesnotexist000000")
		if err == nil {
			t.Fatal("expected error for nonexistent job")
		}
	})
}

func TestWrapperNamespaceIsolation(t *testing.T) {
	c := setupWrapper(t)

	t.Run("cross_namespace_404", func(t *testing.T) {
		// Get a markdown job ID (create via a real call if needed)
		// For now test that a fake ID returns 404 on both
		_, err := c.GetMarkdownJob("job_doesnotexist000000")
		if err == nil {
			t.Fatal("expected 404 error")
		}
		if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "Not Found") {
			t.Fatalf("expected 404-like error, got: %v", err)
		}
	})
}

// =============================================================================
// ADVERSARIAL
// =============================================================================

func TestWrapperAdversarial(t *testing.T) {
	c := setupWrapper(t)

	t.Run("sql_in_query", func(t *testing.T) {
		data, err := c.Extract("https://example.com", &ExtractOptions{
			Method: "llm", Query: "'; DROP TABLE users; --",
		})
		if err != nil {
			t.Fatalf("should not crash: %v", err)
		}
		_ = data // just verify no panic
	})

	t.Run("xss_in_config", func(t *testing.T) {
		md, err := c.Markdown("https://example.com", &MarkdownOptions{
			Strategy:      "http",
			CrawlerConfig: map[string]interface{}{"css_selector": "<script>alert(1)</script>"},
		})
		if err != nil {
			t.Fatalf("should not crash: %v", err)
		}
		_ = md
	})

	t.Run("unicode_url", func(t *testing.T) {
		md, err := c.Markdown("https://example.com/\u00e9\u00e8", &MarkdownOptions{Strategy: "http"})
		if err != nil {
			t.Fatalf("should not crash: %v", err)
		}
		_ = md
	})
}

// Prevent "unused import" for time package
var _ = time.Second
