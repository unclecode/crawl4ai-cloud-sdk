package crawl4ai

import (
	"os"
	"strings"
	"testing"
	"time"
)

// E2E tests for the enhance_query opt-in. Real HTTP against
// stage.crawl4ai.com.

func enhanceTestCrawler(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	apiKey := os.Getenv("CRAWL4AI_API_KEY")
	if apiKey == "" {
		apiKey = "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ"
	}
	baseURL := os.Getenv("CRAWL4AI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://stage.crawl4ai.com"
	}
	c, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Timeout: 180 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewAsyncWebCrawler: %v", err)
	}
	return c
}

func TestDiscoverySearch_EnhanceQuery_SingleBackend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in -short")
	}
	c := enhanceTestCrawler(t)
	resp, err := c.DiscoverySearch(map[string]interface{}{
		"query":         "what are the best nurseries in Toronto for my 2 year old",
		"country":       "ca",
		"enhance_query": true,
	})
	if err != nil {
		t.Fatalf("DiscoverySearch: %v", err)
	}
	if resp.OriginalQuery == nil {
		t.Fatal("OriginalQuery is nil — expected the echoed input")
	}
	if !strings.Contains(*resp.OriginalQuery, "best nurseries") {
		t.Errorf("OriginalQuery mismatch: %q", *resp.OriginalQuery)
	}
	if len(resp.RewrittenQueries) == 0 {
		t.Fatal("RewrittenQueries is empty — expected at least google")
	}
	google, ok := resp.RewrittenQueries["google"]
	if !ok {
		t.Fatalf("RewrittenQueries missing google key: got %v", resp.RewrittenQueries)
	}
	if !strings.Contains(google, "Toronto") {
		t.Errorf("Google rewrite missing Toronto: %q", google)
	}
	if strings.Contains(google, "what are the best") {
		t.Errorf("Google rewrite still has filler phrasing: %q", google)
	}
	time.Sleep(3 * time.Second)
}

func TestDiscoverySearch_EnhanceQuery_MultiBackend(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in -short")
	}
	c := enhanceTestCrawler(t)
	resp, err := c.DiscoverySearch(map[string]interface{}{
		"query":         "latest claude news this week",
		"country":       "us",
		"enhance_query": true,
		"backends":      []string{"google", "bing"},
	})
	if err != nil {
		t.Fatalf("DiscoverySearch: %v", err)
	}
	if resp.OriginalQuery == nil || *resp.OriginalQuery != "latest claude news this week" {
		t.Fatalf("OriginalQuery mismatch: %v", resp.OriginalQuery)
	}
	if len(resp.RewrittenQueries) == 0 {
		t.Fatal("expected at least one backend rewrite, got empty map")
	}
	// Multi-backend failure isolation: if a backend's SERP fetch times
	// out it's dropped from the merge AND its rewrite from the dict.
	// Check operator vocabulary per backend when present.
	if g, ok := resp.RewrittenQueries["google"]; ok {
		if !strings.Contains(g, "after:") {
			t.Errorf("Google rewrite missing after: operator: %q", g)
		}
	}
	if b, ok := resp.RewrittenQueries["bing"]; ok {
		if !strings.Contains(b, "2026") {
			t.Errorf("Bing rewrite missing quoted-year fallback: %q", b)
		}
	}
	// Throttle a touch before the next test to dodge captcha walls.
	time.Sleep(3 * time.Second)
}

func TestDiscoverySearch_EnhanceQuery_DefaultOff(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in -short")
	}
	c := enhanceTestCrawler(t)
	resp, err := c.DiscoverySearch(map[string]interface{}{
		"query":   "openai latest news",
		"country": "us",
	})
	if err != nil {
		t.Fatalf("DiscoverySearch: %v", err)
	}
	if resp.OriginalQuery != nil {
		t.Errorf("OriginalQuery should be nil when enhance_query=false, got %v",
			*resp.OriginalQuery)
	}
	if len(resp.RewrittenQueries) > 0 {
		t.Errorf("RewrittenQueries should be empty when enhance_query=false, got %v",
			resp.RewrittenQueries)
	}
}
