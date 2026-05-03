// E2E tests for the 0.8.1 multi-URL fan-out + url_statuses[] flow (Go).
//
// Covers:
//   - Wrapper async GET responses parse URLStatuses and DownloadURL.
//   - GetPerUrlResult fetches one URL's CrawlResult (recipe-agnostic).
//   - Wait=true hydrates job.Results from per-URL S3 (auto-hydrate path).
//   - Single-URL submits leave URLStatuses nil.
package crawl4ai

import (
	"encoding/json"
	"testing"
	"time"
)

// ─── Pure unit tests (no network) ────────────────────────────────────────

func TestWrapperJob_UrlStatusesParsed(t *testing.T) {
	raw := `{
		"job_id":"job_abc",
		"status":"completed",
		"progress":{"total":2,"completed":2,"failed":0},
		"progress_percent":100,
		"url_statuses":[
			{"index":0,"url":"https://a.com","status":"done","duration_ms":100,"error":null},
			{"index":1,"url":"https://b.com","status":"failed","duration_ms":500,"error":"timeout"}
		],
		"download_url":"https://...zip",
		"created_at":"2026-05-03T00:00:00Z"
	}`
	var job WrapperJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(job.URLStatuses) != 2 {
		t.Fatalf("expected 2 url_statuses, got %d", len(job.URLStatuses))
	}
	if job.URLStatuses[0].Status != "done" {
		t.Errorf("expected status=done, got %s", job.URLStatuses[0].Status)
	}
	if job.URLStatuses[0].DurationMs == nil || *job.URLStatuses[0].DurationMs != 100 {
		t.Errorf("expected duration_ms=100, got %v", job.URLStatuses[0].DurationMs)
	}
	if job.URLStatuses[1].Error != "timeout" {
		t.Errorf("expected error=timeout, got %q", job.URLStatuses[1].Error)
	}
	if job.DownloadURL != "https://...zip" {
		t.Errorf("expected download_url, got %q", job.DownloadURL)
	}
	if job.Results != nil {
		t.Errorf("expected Results=nil before hydration")
	}
}

func TestWrapperJob_SingleUrlNoStatuses(t *testing.T) {
	raw := `{
		"job_id":"job_single",
		"status":"completed",
		"progress":{"total":1,"completed":1,"failed":0},
		"created_at":"2026-05-03T00:00:00Z"
	}`
	var job WrapperJob
	if err := json.Unmarshal([]byte(raw), &job); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if job.URLStatuses != nil {
		t.Errorf("expected URLStatuses=nil, got %v", job.URLStatuses)
	}
	if job.Results != nil {
		t.Errorf("expected Results=nil")
	}
}

// ─── E2E tests (hit stage) ──────────────────────────────────────────────

func TestE2E_ScrapeAsync_WaitHydratesResults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E in -short mode")
	}
	c, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: wrapperTestKey, BaseURL: getBaseURL()})
	if err != nil {
		t.Fatalf("NewAsyncWebCrawler: %v", err)
	}
	job, err := c.ScrapeAsync(
		[]string{"https://example.com", "https://example.org"},
		&ScrapeAsyncOptions{
			MarkdownOptions: MarkdownOptions{Strategy: "http"},
			Wait:            true,
			Timeout:         60 * time.Second,
			PollInterval:    2 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("ScrapeAsync: %v", err)
	}
	if !job.IsComplete() {
		t.Fatalf("expected complete, got %s", job.Status)
	}
	if len(job.URLStatuses) != 2 {
		t.Fatalf("expected 2 url_statuses, got %d", len(job.URLStatuses))
	}
	if len(job.Results) != 2 {
		t.Fatalf("expected 2 hydrated results, got %d", len(job.Results))
	}
}

func TestE2E_ExtractAsync_WaitHydratesExtractedContent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E in -short mode")
	}
	c, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: wrapperTestKey, BaseURL: getBaseURL()})
	if err != nil {
		t.Fatalf("NewAsyncWebCrawler: %v", err)
	}
	job, err := c.ExtractAsync("https://example.com", &ExtractAsyncOptions{
		ExtractOptions: ExtractOptions{
			Method:   "auto",
			Strategy: "http",
			Query:    "page title",
		},
		ExtraURLs:    []string{"https://example.org"},
		Wait:         true,
		Timeout:      120 * time.Second,
		PollInterval: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("ExtractAsync: %v", err)
	}
	if !job.IsComplete() {
		t.Fatalf("expected complete, got %s", job.Status)
	}
	if len(job.URLStatuses) != 2 {
		t.Fatalf("expected 2 url_statuses, got %d", len(job.URLStatuses))
	}
	if len(job.Results) != 2 {
		t.Fatalf("expected 2 hydrated results, got %d", len(job.Results))
	}
}

func TestE2E_GetPerUrlResult_RecipeAgnostic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E in -short mode")
	}
	c, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: wrapperTestKey, BaseURL: getBaseURL()})
	if err != nil {
		t.Fatalf("NewAsyncWebCrawler: %v", err)
	}
	job, err := c.ScrapeAsync(
		[]string{"https://example.com", "https://example.org"},
		&ScrapeAsyncOptions{
			MarkdownOptions: MarkdownOptions{Strategy: "http"},
			Wait:            true,
			Timeout:         60 * time.Second,
			PollInterval:    2 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("ScrapeAsync: %v", err)
	}
	r, err := c.GetPerUrlResult(job.JobID, 0)
	if err != nil {
		t.Fatalf("GetPerUrlResult: %v", err)
	}
	if r.URL == "" {
		t.Errorf("expected URL on result")
	}
}

func TestE2E_SingleUrlNoFanOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E in -short mode")
	}
	c, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: wrapperTestKey, BaseURL: getBaseURL()})
	if err != nil {
		t.Fatalf("NewAsyncWebCrawler: %v", err)
	}
	job, err := c.ScrapeAsync(
		[]string{"https://example.com"},
		&ScrapeAsyncOptions{
			MarkdownOptions: MarkdownOptions{Strategy: "http"},
			Wait:            true,
			Timeout:         60 * time.Second,
			PollInterval:    2 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("ScrapeAsync: %v", err)
	}
	if job.URLStatuses != nil {
		t.Errorf("expected URLStatuses=nil for single-URL, got %v", job.URLStatuses)
	}
	if job.Results != nil {
		t.Errorf("expected Results=nil for single-URL")
	}
}
