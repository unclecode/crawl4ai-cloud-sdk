package crawl4ai

import (
	"os"
	"strings"
	"testing"
)

// Test API key
var testAPIKey = getTestAPIKey()

func getTestAPIKey() string {
	if key := os.Getenv("CRAWL4AI_API_KEY"); key != "" {
		return key
	}
	return "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ"
}

const (
	testURL  = "https://example.com"
	testURL2 = "https://httpbin.org/html"
)

// =============================================================================
// INITIALIZATION TESTS
// =============================================================================

func TestNewAsyncWebCrawler_WithAPIKey(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}
	if crawler == nil {
		t.Fatal("Crawler is nil")
	}
}

func TestNewAsyncWebCrawler_MissingAPIKey(t *testing.T) {
	// Temporarily unset env var
	originalKey := os.Getenv("CRAWL4AI_API_KEY")
	os.Unsetenv("CRAWL4AI_API_KEY")
	defer os.Setenv("CRAWL4AI_API_KEY", originalKey)

	_, err := NewAsyncWebCrawler(CrawlerOptions{})
	if err == nil {
		t.Fatal("Expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "API key is required") {
		t.Fatalf("Expected 'API key is required' error, got: %v", err)
	}
}

func TestNewAsyncWebCrawler_InvalidAPIKeyFormat(t *testing.T) {
	_, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: "invalid_key"})
	if err == nil {
		t.Fatal("Expected error for invalid API key format")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid api key format") {
		t.Fatalf("Expected 'invalid API key format' error, got: %v", err)
	}
}

func TestNewAsyncWebCrawler_AcceptsSkTestPrefix(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: "sk_test_dummy_12345"})
	if err != nil {
		t.Fatalf("Failed to accept sk_test_ prefix: %v", err)
	}
	if crawler == nil {
		t.Fatal("Crawler is nil")
	}
}

func TestNewAsyncWebCrawler_CustomBaseURL(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  testAPIKey,
		BaseURL: "https://api.crawl4ai.com",
	})
	if err != nil {
		t.Fatalf("Failed with custom base URL: %v", err)
	}
	if crawler == nil {
		t.Fatal("Crawler is nil")
	}
}

// =============================================================================
// SINGLE URL CRAWL TESTS
// =============================================================================

func TestRun_Basic(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Run(testURL, nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Crawl was not successful: %s", result.ErrorMessage)
	}
	if result.URL != testURL {
		t.Fatalf("Expected URL %s, got %s", testURL, result.URL)
	}
}

func TestRun_ReturnsMarkdown(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Run(testURL, nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Markdown == nil {
		t.Fatal("Markdown is nil")
	}
	if result.Markdown.RawMarkdown == "" {
		t.Fatal("RawMarkdown is empty")
	}
	if !strings.Contains(result.Markdown.RawMarkdown, "Example Domain") {
		t.Fatal("Markdown does not contain expected content")
	}
}

func TestRun_ReturnsHTML(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Run(testURL, nil)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.HTML == "" {
		t.Fatal("HTML is empty")
	}
	if !strings.Contains(strings.ToLower(result.HTML), "<html") {
		t.Fatal("HTML does not contain expected content")
	}
}

func TestRun_BrowserStrategy(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Run(testURL, &RunOptions{Strategy: "browser"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
}

func TestRun_HTTPStrategy(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Run(testURL, &RunOptions{Strategy: "http"})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
}

func TestRun_BypassCache(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Run(testURL, &RunOptions{BypassCache: true})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
}

// =============================================================================
// OSS COMPATIBILITY TESTS
// =============================================================================

func TestArun_Alias(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.Arun(testURL, nil)
	if err != nil {
		t.Fatalf("Arun failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
	if result.URL != testURL {
		t.Fatalf("Expected URL %s, got %s", testURL, result.URL)
	}
}

func TestArunMany_Alias(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	urls := []string{testURL, testURL2}
	result, err := crawler.ArunMany(urls, &RunManyOptions{Wait: true})
	if err != nil {
		t.Fatalf("ArunMany failed: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result.Results))
	}
}

// =============================================================================
// CONFIGURATION TESTS
// =============================================================================

func TestRun_WithConfig(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	config := &CrawlerRunConfig{
		WordCountThreshold:  10,
		ExcludeExternalLinks: true,
	}

	result, err := crawler.Run(testURL, &RunOptions{Config: config})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
}

func TestRun_WithBrowserConfig(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	browserConfig := &BrowserConfig{
		ViewportWidth:  1920,
		ViewportHeight: 1080,
	}

	result, err := crawler.Run(testURL, &RunOptions{BrowserConfig: browserConfig})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
}

func TestSanitizeCrawlerConfig_RemovesCacheFields(t *testing.T) {
	config := &CrawlerRunConfig{
		CacheMode:   "bypass",
		SessionID:   "test-session",
		BypassCache: true,
		Screenshot:  true,
	}

	sanitized := SanitizeCrawlerConfig(config)

	if _, ok := sanitized["cache_mode"]; ok {
		t.Fatal("cache_mode should be removed")
	}
	if _, ok := sanitized["session_id"]; ok {
		t.Fatal("session_id should be removed")
	}
	if _, ok := sanitized["bypass_cache"]; ok {
		t.Fatal("bypass_cache should be removed")
	}
	if sanitized["screenshot"] != true {
		t.Fatal("screenshot should be preserved")
	}
}

func TestSanitizeBrowserConfig_RemovesCDPFields(t *testing.T) {
	config := &BrowserConfig{
		CdpURL:            "ws://localhost:9222",
		UseManagedBrowser: true,
		Headless:          true,
	}

	sanitized := SanitizeBrowserConfig(config, "browser")

	if _, ok := sanitized["cdp_url"]; ok {
		t.Fatal("cdp_url should be removed")
	}
	if _, ok := sanitized["use_managed_browser"]; ok {
		t.Fatal("use_managed_browser should be removed")
	}
	if sanitized["headless"] != true {
		t.Fatal("headless should be preserved")
	}
}

// =============================================================================
// PROXY CONFIGURATION TESTS
// =============================================================================

func TestNormalizeProxy_String(t *testing.T) {
	result, err := NormalizeProxy("datacenter")
	if err != nil {
		t.Fatalf("NormalizeProxy failed: %v", err)
	}
	if result["mode"] != "datacenter" {
		t.Fatalf("Expected mode 'datacenter', got %v", result["mode"])
	}
}

func TestNormalizeProxy_Map(t *testing.T) {
	proxy := map[string]interface{}{"mode": "residential", "country": "US"}
	result, err := NormalizeProxy(proxy)
	if err != nil {
		t.Fatalf("NormalizeProxy failed: %v", err)
	}
	if result["mode"] != "residential" {
		t.Fatalf("Expected mode 'residential', got %v", result["mode"])
	}
	if result["country"] != "US" {
		t.Fatalf("Expected country 'US', got %v", result["country"])
	}
}

func TestNormalizeProxy_ProxyConfig(t *testing.T) {
	proxy := &ProxyConfig{Mode: "auto", Country: "UK"}
	result, err := NormalizeProxy(proxy)
	if err != nil {
		t.Fatalf("NormalizeProxy failed: %v", err)
	}
	if result["mode"] != "auto" {
		t.Fatalf("Expected mode 'auto', got %v", result["mode"])
	}
	if result["country"] != "UK" {
		t.Fatalf("Expected country 'UK', got %v", result["country"])
	}
}

func TestNormalizeProxy_Nil(t *testing.T) {
	result, err := NormalizeProxy(nil)
	if err != nil {
		t.Fatalf("NormalizeProxy failed: %v", err)
	}
	if result != nil {
		t.Fatal("Expected nil result for nil input")
	}
}

func TestNormalizeProxy_InvalidType(t *testing.T) {
	_, err := NormalizeProxy(12345)
	if err == nil {
		t.Fatal("Expected error for invalid type")
	}
}

// =============================================================================
// BATCH CRAWL TESTS
// =============================================================================

func TestRunMany_SmallBatchWait(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	urls := []string{testURL, testURL2}
	result, err := crawler.RunMany(urls, &RunManyOptions{Wait: true})
	if err != nil {
		t.Fatalf("RunMany failed: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result.Results))
	}
	for _, r := range result.Results {
		if !r.Success {
			t.Fatalf("Result was not successful: %s", r.ErrorMessage)
		}
	}
}

func TestRunMany_SmallBatchNoWait(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	urls := []string{testURL, testURL2}
	result, err := crawler.RunMany(urls, &RunManyOptions{Wait: false})
	if err != nil {
		t.Fatalf("RunMany failed: %v", err)
	}

	if result.Job == nil {
		t.Fatal("Job is nil")
	}
	if result.Job.Status != "completed" {
		t.Fatalf("Expected job status 'completed', got '%s'", result.Job.Status)
	}
}

// =============================================================================
// JOB MANAGEMENT TESTS
// =============================================================================

func TestListJobs(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	jobs, err := crawler.ListJobs(&ListJobsOptions{Limit: 5})
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	// Jobs may be empty if none exist
	for _, job := range jobs {
		if job.JobID == "" {
			t.Fatal("Job ID is empty")
		}
		if job.Status == "" {
			t.Fatal("Job status is empty")
		}
	}
}

func TestListJobs_WithStatusFilter(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	jobs, err := crawler.ListJobs(&ListJobsOptions{Status: "completed", Limit: 5})
	if err != nil {
		t.Fatalf("ListJobs failed: %v", err)
	}

	for _, job := range jobs {
		if job.Status != "completed" {
			t.Fatalf("Expected status 'completed', got '%s'", job.Status)
		}
	}
}

// =============================================================================
// STORAGE API TESTS
// =============================================================================

func TestStorage(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	usage, err := crawler.Storage()
	if err != nil {
		t.Fatalf("Storage failed: %v", err)
	}

	if usage.MaxMB < 0 {
		t.Fatal("MaxMB should not be negative")
	}
	if usage.UsedMB < 0 {
		t.Fatal("UsedMB should not be negative")
	}
	if usage.RemainingMB < 0 {
		t.Fatal("RemainingMB should not be negative")
	}
}

// =============================================================================
// HEALTH CHECK TESTS
// =============================================================================

func TestHealth(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	health, err := crawler.Health()
	if err != nil {
		t.Fatalf("Health failed: %v", err)
	}

	if health == nil {
		t.Fatal("Health response is nil")
	}
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestRun_InvalidAPIKey(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: "sk_test_invalid_12345"})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	_, err = crawler.Run(testURL, nil)
	if err == nil {
		t.Fatal("Expected error for invalid API key")
	}

	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("Expected AuthenticationError, got %T: %v", err, err)
	}
}

func TestGetJob_NotFound(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	_, err = crawler.GetJob("nonexistent-job-12345", false)
	if err == nil {
		t.Fatal("Expected error for non-existent job")
	}

	if _, ok := err.(*NotFoundError); !ok {
		t.Fatalf("Expected NotFoundError, got %T: %v", err, err)
	}
}

// =============================================================================
// DEEP CRAWL TESTS
// =============================================================================

func TestDeepCrawl_RequiresURLOrSourceJob(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	_, err = crawler.DeepCrawl("", nil)
	if err == nil {
		t.Fatal("Expected error when neither url nor sourceJob provided")
	}
	if !strings.Contains(err.Error(), "must provide") {
		t.Fatalf("Expected 'must provide' error, got: %v", err)
	}
}

func TestDeepCrawl_RejectsBothURLAndSourceJob(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	_, err = crawler.DeepCrawl(testURL, &DeepCrawlOptions{SourceJob: "some-job"})
	if err == nil {
		t.Fatal("Expected error when both url and sourceJob provided")
	}
	if !strings.Contains(err.Error(), "not both") {
		t.Fatalf("Expected 'not both' error, got: %v", err)
	}
}

func TestDeepCrawl_ScanOnly(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	result, err := crawler.DeepCrawl(testURL, &DeepCrawlOptions{
		Strategy: "bfs",
		MaxDepth: 1,
		MaxURLs:  5,
		ScanOnly: true,
		Wait:     true,
	})
	if err != nil {
		t.Fatalf("DeepCrawl failed: %v", err)
	}

	if result.DeepResult == nil {
		t.Fatal("DeepResult is nil")
	}
	if result.DeepResult.JobID == "" {
		t.Fatal("JobID is empty")
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

func TestFullWorkflow_SingleCrawl(t *testing.T) {
	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	config := &CrawlerRunConfig{
		WordCountThreshold: 10,
	}

	browserConfig := &BrowserConfig{
		ViewportWidth:  1280,
		ViewportHeight: 720,
	}

	result, err := crawler.Run(testURL, &RunOptions{
		Config:        config,
		BrowserConfig: browserConfig,
		Strategy:      "browser",
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
	if result.URL != testURL {
		t.Fatalf("Expected URL %s, got %s", testURL, result.URL)
	}
	if !strings.Contains(result.Markdown.RawMarkdown, "Example") {
		t.Fatal("Markdown does not contain expected content")
	}
}

func TestOSSMigrationPattern(t *testing.T) {
	// This is how users migrate from OSS to Cloud:
	// 1. Change import
	// 2. Add API key
	// 3. Use same code

	crawler, err := NewAsyncWebCrawler(CrawlerOptions{APIKey: testAPIKey})
	if err != nil {
		t.Fatalf("Failed to create crawler: %v", err)
	}

	// OSS users use Arun()
	result, err := crawler.Arun(testURL, nil)
	if err != nil {
		t.Fatalf("Arun failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Crawl was not successful")
	}
	if result.Markdown == nil || result.Markdown.RawMarkdown == "" {
		t.Fatal("Markdown is nil or empty")
	}
}
