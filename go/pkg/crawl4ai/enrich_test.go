package crawl4ai

import (
	"os"
	"testing"
	"time"
)

var enrichTestKey = getEnrichTestKey()

func getEnrichTestKey() string {
	if key := os.Getenv("CRAWL4AI_API_KEY"); key != "" {
		return key
	}
	return "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ"
}

func getEnrichBaseURL() string {
	if u := os.Getenv("CRAWL4AI_BASE_URL"); u != "" {
		return u
	}
	return "https://stage.crawl4ai.com"
}

func setupEnrich(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	c, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  enrichTestKey,
		BaseURL: getEnrichBaseURL(),
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return c
}

// =============================================================================
// Happy path
// =============================================================================

func TestEnrichBasic(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com"},
		[]EnrichFieldSpec{
			{Name: "Company Name"},
			{Name: "Email", Description: "contact email"},
		},
		&EnrichOptions{
			MaxDepth: 0,
			Wait:     true,
			Strategy: "browser",
			Timeout:  120 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if !result.IsComplete() {
		t.Fatalf("expected complete, got status=%s", result.Status)
	}
	if !result.IsSuccessful() {
		t.Fatalf("expected successful, got status=%s, error=%s", result.Status, result.Error)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	row := result.Rows[0]
	if row.URL != "https://kidocode.com" {
		t.Fatalf("expected url=https://kidocode.com, got %s", row.URL)
	}
	if row.Fields["Company Name"] == nil || row.Fields["Company Name"] == "" {
		t.Fatal("Company Name should be found")
	}
	if row.Status != "complete" && row.Status != "partial" {
		t.Fatalf("expected row status complete|partial, got %s", row.Status)
	}
	if row.DepthUsed != 0 {
		t.Fatalf("expected depth_used=0, got %d", row.DepthUsed)
	}
}

func TestEnrichWithDepth(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com"},
		[]EnrichFieldSpec{
			{Name: "Company Name"},
			{Name: "Email", Description: "primary contact email"},
			{Name: "Phone", Description: "phone number"},
		},
		&EnrichOptions{
			MaxDepth: 1,
			MaxLinks: 3,
			Wait:     true,
			Timeout:  120 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Enrich with depth: %v", err)
	}
	if !result.IsComplete() {
		t.Fatalf("expected complete, got status=%s", result.Status)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}

	row := result.Rows[0]
	if row.Fields["Company Name"] == nil || row.Fields["Company Name"] == "" {
		t.Fatal("Company Name should be found")
	}
	// With depth 1, should find more fields
	found := 0
	for _, v := range row.Fields {
		if v != nil && v != "" {
			found++
		}
	}
	if found < 2 {
		t.Fatalf("expected at least 2 fields found, got %d", found)
	}
}

func TestEnrichMultipleUrls(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com", "https://httpbin.org"},
		[]EnrichFieldSpec{
			{Name: "Title", Description: "page or company title"},
		},
		&EnrichOptions{
			MaxDepth: 0,
			Wait:     true,
			Timeout:  120 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Enrich multiple: %v", err)
	}
	if !result.IsComplete() {
		t.Fatalf("expected complete, got status=%s", result.Status)
	}
	if result.Progress.Total != 2 {
		t.Fatalf("expected progress.total=2, got %d", result.Progress.Total)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}

	urls := make(map[string]bool)
	for _, r := range result.Rows {
		urls[r.URL] = true
	}
	if !urls["https://kidocode.com"] && !urls["https://httpbin.org"] {
		t.Fatal("expected at least one known URL in rows")
	}
}

// =============================================================================
// Source attribution
// =============================================================================

func TestEnrichSourceAttribution(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com"},
		[]EnrichFieldSpec{
			{Name: "Company Name"},
			{Name: "Email"},
		},
		&EnrichOptions{
			MaxDepth: 0,
			Wait:     true,
			Timeout:  120 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}

	row := result.Rows[0]
	for fieldName, value := range row.Fields {
		if value == nil || value == "" {
			continue
		}
		src, ok := row.Sources[fieldName]
		if !ok {
			t.Fatalf("missing source for field %q", fieldName)
		}
		if src.Method != "direct" && src.Method != "depth" && src.Method != "search" {
			t.Fatalf("unexpected source method %q for field %q", src.Method, fieldName)
		}
		if src.URL == "" {
			t.Fatalf("source URL empty for field %q", fieldName)
		}
	}
}

// =============================================================================
// Job management
// =============================================================================

func TestEnrichFireAndForget(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com"},
		[]EnrichFieldSpec{{Name: "Company Name"}},
		&EnrichOptions{
			MaxDepth: 0,
			Wait:     false,
		},
	)
	if err != nil {
		t.Fatalf("Enrich fire-and-forget: %v", err)
	}
	if result.JobID == "" {
		t.Fatal("expected job_id")
	}
	if result.Status != "pending" {
		t.Fatalf("expected status=pending, got %s", result.Status)
	}

	// Poll until done
	var status *EnrichJobStatus
	for i := 0; i < 30; i++ {
		status, err = c.GetEnrichJob(result.JobID)
		if err != nil {
			t.Fatalf("GetEnrichJob: %v", err)
		}
		if status.IsComplete() {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if status == nil || !status.IsComplete() {
		t.Fatal("job did not complete within polling window")
	}
	if status.Rows == nil {
		t.Fatal("expected rows on completed job")
	}
}

func TestEnrichListJobs(t *testing.T) {
	c := setupEnrich(t)

	jobs, err := c.ListEnrichJobs(5, 0)
	if err != nil {
		t.Fatalf("ListEnrichJobs: %v", err)
	}
	if len(jobs) < 1 {
		t.Fatal("expected at least 1 enrich job in list")
	}
	for _, j := range jobs {
		if j.JobID == "" {
			t.Fatal("expected job_id on each listed job")
		}
	}
}

func TestEnrichCancelJob(t *testing.T) {
	c := setupEnrich(t)

	// Create a job with multiple URLs and search to keep it running
	result, err := c.Enrich(
		[]string{"https://example.com", "https://httpbin.org", "https://kidocode.com"},
		[]EnrichFieldSpec{
			{Name: "Title"},
			{Name: "Description", Description: "page description"},
			{Name: "Email"},
		},
		&EnrichOptions{
			MaxDepth:     1,
			EnableSearch: true,
			Wait:         false,
		},
	)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if result.JobID == "" {
		t.Fatal("expected job_id")
	}

	// Cancel
	if err := c.CancelEnrichJob(result.JobID); err != nil {
		t.Fatalf("CancelEnrichJob: %v", err)
	}

	// Verify cancelled
	status, err := c.GetEnrichJob(result.JobID)
	if err != nil {
		t.Fatalf("GetEnrichJob after cancel: %v", err)
	}
	if status.Status != "cancelled" {
		t.Fatalf("expected status=cancelled, got %s", status.Status)
	}
}

// =============================================================================
// Progress and token usage
// =============================================================================

func TestEnrichProgressTracking(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com"},
		[]EnrichFieldSpec{{Name: "Company Name"}},
		&EnrichOptions{
			MaxDepth: 0,
			Wait:     true,
			Timeout:  120 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}

	if result.Progress.Total != 1 {
		t.Fatalf("expected progress.total=1, got %d", result.Progress.Total)
	}
	if result.Progress.Completed+result.Progress.Failed != 1 {
		t.Fatalf("expected completed+failed=1, got %d+%d",
			result.Progress.Completed, result.Progress.Failed)
	}
	if result.ProgressPercent != 100 {
		t.Fatalf("expected progress_percent=100, got %d", result.ProgressPercent)
	}
}

func TestEnrichTokenUsage(t *testing.T) {
	c := setupEnrich(t)

	result, err := c.Enrich(
		[]string{"https://kidocode.com"},
		[]EnrichFieldSpec{
			{Name: "Company Name"},
			{Name: "Email"},
		},
		&EnrichOptions{
			MaxDepth: 0,
			Wait:     true,
			Timeout:  120 * time.Second,
		},
	)
	if err != nil {
		t.Fatalf("Enrich: %v", err)
	}
	if len(result.Rows) == 0 {
		t.Fatal("expected at least 1 row")
	}

	row := result.Rows[0]
	if row.TokenUsage == nil {
		t.Fatal("expected token_usage to be populated")
	}
	if row.TokenUsage["total_tokens"] <= 0 {
		t.Fatalf("expected total_tokens > 0, got %d", row.TokenUsage["total_tokens"])
	}
}
