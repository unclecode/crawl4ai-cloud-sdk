// Package crawl4ai provides a Go SDK for Crawl4AI Cloud API.
package crawl4ai

import (
	"encoding/json"
	"fmt"
	"time"
)

// AsyncWebCrawler is the main client for Crawl4AI Cloud API.
type AsyncWebCrawler struct {
	http *HTTPClient
}

// CrawlerOptions are options for creating an AsyncWebCrawler.
type CrawlerOptions struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

// NewAsyncWebCrawler creates a new AsyncWebCrawler.
func NewAsyncWebCrawler(opts CrawlerOptions) (*AsyncWebCrawler, error) {
	httpClient, err := NewHTTPClient(HTTPClientOptions{
		APIKey:     opts.APIKey,
		BaseURL:    opts.BaseURL,
		Timeout:    opts.Timeout,
		MaxRetries: opts.MaxRetries,
	})
	if err != nil {
		return nil, err
	}

	return &AsyncWebCrawler{http: httpClient}, nil
}

// RunOptions are options for the Run method.
type RunOptions struct {
	Config        *CrawlerRunConfig
	BrowserConfig *BrowserConfig
	Strategy      string // "browser" or "http"
	Proxy         interface{}
	BypassCache   bool
}

// Run crawls a single URL.
func (c *AsyncWebCrawler) Run(url string, opts *RunOptions) (*CrawlResult, error) {
	if opts == nil {
		opts = &RunOptions{}
	}

	strategy := opts.Strategy
	if strategy == "" {
		strategy = "browser"
	}

	body := BuildCrawlRequest(map[string]interface{}{
		"url":           url,
		"config":        opts.Config,
		"browserConfig": opts.BrowserConfig,
		"strategy":      strategy,
		"proxy":         opts.Proxy,
		"bypassCache":   opts.BypassCache,
	})

	data, err := c.http.Post("/v1/crawl", body, 120*time.Second)
	if err != nil {
		return nil, err
	}

	return CrawlResultFromMap(data), nil
}

// Arun is an alias for Run (OSS compatibility).
func (c *AsyncWebCrawler) Arun(url string, opts *RunOptions) (*CrawlResult, error) {
	return c.Run(url, opts)
}

// RunManyOptions are options for the RunMany method.
type RunManyOptions struct {
	Config        *CrawlerRunConfig
	BrowserConfig *BrowserConfig
	Strategy      string
	Proxy         interface{}
	BypassCache   bool
	Wait          bool
	PollInterval  time.Duration
	Timeout       time.Duration
	Priority      int
	WebhookURL    string
}

// RunManyResult holds the result of RunMany.
type RunManyResult struct {
	Job     *CrawlJob
	Results []*CrawlResult
}

// RunMany crawls multiple URLs.
// Creates an async job for processing. Use Wait=true to block until
// complete, or poll with GetJob()/WaitJob().
func (c *AsyncWebCrawler) RunMany(urls []string, opts *RunManyOptions) (*RunManyResult, error) {
	if opts == nil {
		opts = &RunManyOptions{}
	}

	// Always use async endpoint for consistent job tracking
	return c.runAsync(urls, opts)
}

// ArunMany is an alias for RunMany (OSS compatibility).
func (c *AsyncWebCrawler) ArunMany(urls []string, opts *RunManyOptions) (*RunManyResult, error) {
	return c.RunMany(urls, opts)
}

func (c *AsyncWebCrawler) runAsync(urls []string, opts *RunManyOptions) (*RunManyResult, error) {
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "browser"
	}

	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	body := BuildCrawlRequest(map[string]interface{}{
		"urls":          urls,
		"config":        opts.Config,
		"browserConfig": opts.BrowserConfig,
		"strategy":      strategy,
		"proxy":         opts.Proxy,
		"bypassCache":   opts.BypassCache,
		"priority":      priority,
		"webhookUrl":    opts.WebhookURL,
	})

	data, err := c.http.Post("/v1/crawl/async", body, 0)
	if err != nil {
		return nil, err
	}

	job := CrawlJobFromMap(data)

	if opts.Wait {
		pollInterval := opts.PollInterval
		if pollInterval == 0 {
			pollInterval = 2 * time.Second
		}

		job, err = c.WaitJob(job.JobID, pollInterval, opts.Timeout)
		if err != nil {
			return nil, err
		}

		// Results are available via DownloadURL() after job completes
		return &RunManyResult{Job: job}, nil
	}

	return &RunManyResult{Job: job}, nil
}

// GetJob gets job status.
// To get results, use DownloadURL() to get a presigned URL for the ZIP file.
func (c *AsyncWebCrawler) GetJob(jobID string) (*CrawlJob, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/crawl/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}

	return CrawlJobFromMap(data), nil
}

// WaitJob polls until job completes.
// To get results after job completes, use DownloadURL() to get a presigned URL for the ZIP file.
func (c *AsyncWebCrawler) WaitJob(jobID string, pollInterval, timeout time.Duration) (*CrawlJob, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	startTime := time.Now()

	for {
		job, err := c.GetJob(jobID)
		if err != nil {
			return nil, err
		}

		if job.IsComplete() {
			return job, nil
		}

		if timeout > 0 && time.Since(startTime) > timeout {
			return nil, NewTimeoutError(fmt.Sprintf(
				"timeout waiting for job %s. Status: %s, Progress: %.1f%%",
				jobID, job.Status, job.Progress.Percent(),
			))
		}

		time.Sleep(pollInterval)
	}
}

// ListJobsOptions are options for ListJobs.
type ListJobsOptions struct {
	Status string
	Limit  int
	Offset int
}

// ListJobs lists jobs with optional filtering.
func (c *AsyncWebCrawler) ListJobs(opts *ListJobsOptions) ([]*CrawlJob, error) {
	if opts == nil {
		opts = &ListJobsOptions{}
	}

	params := make(map[string]string)
	if opts.Status != "" {
		params["status"] = opts.Status
	}
	if opts.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", opts.Limit)
	} else {
		params["limit"] = "20"
	}
	if opts.Offset > 0 {
		params["offset"] = fmt.Sprintf("%d", opts.Offset)
	}

	data, err := c.http.Get("/v1/crawl/jobs", params)
	if err != nil {
		return nil, err
	}

	jobs := make([]*CrawlJob, 0)
	if rawJobs, ok := data["jobs"].([]interface{}); ok {
		for _, j := range rawJobs {
			if m, ok := j.(map[string]interface{}); ok {
				jobs = append(jobs, CrawlJobFromMap(m))
			}
		}
	}

	return jobs, nil
}

// CancelJob cancels a pending or running job.
func (c *AsyncWebCrawler) CancelJob(jobID string) error {
	_, err := c.http.Delete(fmt.Sprintf("/v1/crawl/jobs/%s", jobID))
	return err
}

// DeepCrawlOptions are options for DeepCrawl.
type DeepCrawlOptions struct {
	SourceJob     string
	Strategy      string // "bfs", "dfs", "best_first", "map"
	MaxDepth      int
	MaxURLs       int
	ScanOnly      bool
	Config        *CrawlerRunConfig
	BrowserConfig *BrowserConfig
	CrawlStrategy string // "browser", "http", "auto"
	Proxy         interface{}
	BypassCache   bool
	Wait          bool
	PollInterval  time.Duration
	Timeout       time.Duration
	Filters       map[string]interface{}
	Scorers       map[string]interface{}
	IncludeHTML   bool
	WebhookURL    string
	Priority      int
	// Map strategy options
	Source         string
	Pattern        string
	Query          string
	ScoreThreshold *float64
	// URL filtering shortcuts
	IncludePatterns []string
	ExcludePatterns []string
}

// DeepCrawlResult holds the result of DeepCrawl.
type DeepCrawlResultWrapper struct {
	DeepResult *DeepCrawlResult
	CrawlJob   *CrawlJob
}

// DeepCrawl performs a deep crawl starting from a URL.
func (c *AsyncWebCrawler) DeepCrawl(url string, opts *DeepCrawlOptions) (*DeepCrawlResultWrapper, error) {
	if opts == nil {
		opts = &DeepCrawlOptions{}
	}

	if url == "" && opts.SourceJob == "" {
		return nil, fmt.Errorf("must provide either 'url' or 'SourceJob'")
	}
	if url != "" && opts.SourceJob != "" {
		return nil, fmt.Errorf("provide either 'url' or 'SourceJob', not both")
	}

	strategy := opts.Strategy
	if strategy == "" {
		strategy = "bfs"
	}

	crawlStrategy := opts.CrawlStrategy
	if crawlStrategy == "" {
		crawlStrategy = "auto"
	}

	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = 3
	}

	maxURLs := opts.MaxURLs
	if maxURLs == 0 {
		maxURLs = 100
	}

	body := map[string]interface{}{}

	if opts.SourceJob != "" {
		// Phase 2: extraction from cached HTML — only send source_job_id
		body["source_job_id"] = opts.SourceJob
	} else {
		// Phase 1: URL-based discovery — include scan parameters
		body["url"] = url
		body["strategy"] = strategy
		body["crawl_strategy"] = crawlStrategy
		body["priority"] = priority

		// Tree strategy options
		if strategy == "bfs" || strategy == "dfs" || strategy == "best_first" {
			body["max_depth"] = maxDepth
			body["max_urls"] = maxURLs

			// Build filters from IncludePatterns/ExcludePatterns or use provided filters
			effectiveFilters := make(map[string]interface{})
			if opts.Filters != nil {
				for k, v := range opts.Filters {
					effectiveFilters[k] = v
				}
			}
			if len(opts.IncludePatterns) > 0 {
				effectiveFilters["include_patterns"] = opts.IncludePatterns
			}
			if len(opts.ExcludePatterns) > 0 {
				effectiveFilters["exclude_patterns"] = opts.ExcludePatterns
			}
			if len(effectiveFilters) > 0 {
				body["filters"] = effectiveFilters
			}

			if opts.Scorers != nil {
				body["scorers"] = opts.Scorers
			}
			if opts.ScanOnly {
				body["scan_only"] = true
			}
			if opts.IncludeHTML {
				body["include_html"] = true
			}
		}

		// Map strategy options
		if strategy == "map" {
			seedingConfig := map[string]interface{}{
				"source":  opts.Source,
				"pattern": opts.Pattern,
			}
			if opts.Source == "" {
				seedingConfig["source"] = "sitemap"
			}
			if opts.Pattern == "" {
				seedingConfig["pattern"] = "*"
			}
			if maxURLs > 0 {
				seedingConfig["max_urls"] = maxURLs
			}
			if opts.Query != "" {
				seedingConfig["query"] = opts.Query
			}
			if opts.ScoreThreshold != nil {
				seedingConfig["score_threshold"] = *opts.ScoreThreshold
			}
			body["seeding_config"] = seedingConfig
		}
	}

	// Add configs
	if sanitized := SanitizeCrawlerConfig(opts.Config); sanitized != nil {
		body["crawler_config"] = sanitized
	}
	if sanitized := SanitizeBrowserConfig(opts.BrowserConfig, crawlStrategy); sanitized != nil {
		body["browser_config"] = sanitized
	}

	// Proxy
	if proxyMap, err := NormalizeProxy(opts.Proxy); err == nil && proxyMap != nil {
		body["proxy"] = proxyMap
	}

	if opts.BypassCache {
		body["bypass_cache"] = true
	}
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	data, err := c.http.Post("/v1/crawl/deep", body, 120*time.Second)
	if err != nil {
		return nil, err
	}

	result := DeepCrawlResultFromMap(data)

	if !opts.Wait {
		return &DeepCrawlResultWrapper{DeepResult: result}, nil
	}

	// Wait for scan to complete
	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	result, err = c.waitScanJob(result.JobID, pollInterval, opts.Timeout)
	if err != nil {
		return nil, err
	}

	if opts.ScanOnly {
		return &DeepCrawlResultWrapper{DeepResult: result}, nil
	}

	if result.Status == "no_urls" || result.DiscoveredCount == 0 {
		return &DeepCrawlResultWrapper{DeepResult: result}, nil
	}

	// If crawl job was created, wait for it
	if result.CrawlJobID != "" {
		job, err := c.WaitJob(result.CrawlJobID, pollInterval, opts.Timeout)
		if err != nil {
			return nil, err
		}
		return &DeepCrawlResultWrapper{DeepResult: result, CrawlJob: job}, nil
	}

	return &DeepCrawlResultWrapper{DeepResult: result}, nil
}

func (c *AsyncWebCrawler) waitScanJob(jobID string, pollInterval, timeout time.Duration) (*DeepCrawlResult, error) {
	startTime := time.Now()

	for {
		data, err := c.http.Get(fmt.Sprintf("/v1/crawl/deep/jobs/%s", jobID), nil)
		if err != nil {
			return nil, err
		}

		result := DeepCrawlResultFromMap(data)

		if result.IsComplete() {
			return result, nil
		}

		if timeout > 0 && time.Since(startTime) > timeout {
			return nil, NewTimeoutError(fmt.Sprintf(
				"timeout waiting for scan job %s. Status: %s, Discovered: %d",
				jobID, result.Status, result.DiscoveredCount,
			))
		}

		time.Sleep(pollInterval)
	}
}

// CancelDeepCrawl cancels a running deep crawl job.
// The crawl will stop at the next batch boundary, preserving any
// partial results that have been collected so far.
func (c *AsyncWebCrawler) CancelDeepCrawl(jobID string) (*DeepCrawlResult, error) {
	data, err := c.http.Post(fmt.Sprintf("/v1/crawl/deep/jobs/%s/cancel", jobID), nil, 0)
	if err != nil {
		return nil, err
	}

	return DeepCrawlResultFromMap(data), nil
}

// GetDeepCrawlStatus gets the status of a deep crawl job.
func (c *AsyncWebCrawler) GetDeepCrawlStatus(jobID string) (*DeepCrawlResult, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/crawl/deep/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}

	return DeepCrawlResultFromMap(data), nil
}

// ContextOptions are options for Context.
type ContextOptions struct {
	PAALimit      int
	ResultsPerPAA int
}

// Scan discovers all URLs under a domain without crawling.
//
// Two routing strategies (picked by scan.Mode or inferred from Criteria):
//   - map (sync): DomainMapper — sitemap + CC + wayback. URLs returned inline.
//   - deep (async): best-first tree traversal. Returns a JobID; poll with
//     GetScanJob() or pass opts.Wait = true.
//
// When opts.Criteria is set, the backend LLM generates a unified scan config
// (mode, patterns, filters, scorers, query, threshold). Explicit overrides
// in opts.Scan still win.
func (c *AsyncWebCrawler) Scan(url string, opts *ScanOptions) (*ScanResult, error) {
	body := map[string]interface{}{
		"url": url,
	}
	if opts != nil {
		if opts.Mode != "" {
			body["mode"] = opts.Mode
		}
		if opts.MaxUrls > 0 {
			body["max_urls"] = opts.MaxUrls
		}
		if opts.IncludeSubdomains != nil {
			body["include_subdomains"] = *opts.IncludeSubdomains
		}
		if opts.ExtractHead != nil {
			body["extract_head"] = *opts.ExtractHead
		}
		if opts.Soft404Detection != nil {
			body["soft_404_detection"] = *opts.Soft404Detection
		}
		if opts.Query != "" {
			body["query"] = opts.Query
		}
		if opts.ScoreThreshold != nil {
			body["score_threshold"] = *opts.ScoreThreshold
		}
		if opts.Force {
			body["force"] = true
		}
		if opts.ProbeThreshold != nil {
			body["probe_threshold"] = *opts.ProbeThreshold
		}
		// AI-assisted fields
		if opts.Criteria != "" {
			body["criteria"] = opts.Criteria
		}
		if opts.Scan != nil {
			body["scan"] = opts.Scan.ToMap()
		}
	}

	// Longer HTTP timeout — LLM config gen can take a while.
	data, err := c.http.Post("/v1/scan", body, 180*time.Second)
	if err != nil {
		return nil, err
	}

	result := ScanResultFromMap(data)

	// If the LLM picked deep mode (or the caller forced it) and opts.Wait is set,
	// block until the scan job finishes and merge state back onto the result.
	if opts != nil && opts.Wait && result.IsAsync() {
		final, err := c.waitScanJobV2(result.JobID, opts.PollInterval, opts.Timeout)
		if err != nil {
			return result, err
		}
		result.Status = final.Status
		result.TotalUrls = final.TotalUrls
		result.Urls = final.Urls
		result.DurationMs = final.DurationMs
		if final.Error != "" {
			result.Error = final.Error
		}
	}

	return result, nil
}

// GetScanJob polls a deep scan job started via Scan() with scan.Mode = "deep".
// Returns current discovered URLs, progress, and status. URLs are appended
// as they're discovered.
func (c *AsyncWebCrawler) GetScanJob(jobID string) (*ScanJobStatus, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/scan/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}
	return ScanJobStatusFromMap(data), nil
}

// CancelScanJob cancels a running deep scan. Cancellation happens at the next
// batch boundary — partial results (URLs discovered so far) are preserved.
func (c *AsyncWebCrawler) CancelScanJob(jobID string) (*ScanJobStatus, error) {
	data, err := c.http.Post(fmt.Sprintf("/v1/scan/jobs/%s/cancel", jobID), nil, 0)
	if err != nil {
		return nil, err
	}
	return ScanJobStatusFromMap(data), nil
}

// waitScanJobV2 polls /v1/scan/jobs/{id} until the deep scan finishes.
func (c *AsyncWebCrawler) waitScanJobV2(jobID string, pollInterval, timeout time.Duration) (*ScanJobStatus, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}
	start := time.Now()
	for {
		job, err := c.GetScanJob(jobID)
		if err != nil {
			return nil, err
		}
		if job.IsComplete() {
			return job, nil
		}
		if timeout > 0 && time.Since(start) > timeout {
			return nil, NewTimeoutError(fmt.Sprintf(
				"timeout waiting for scan job %s. Status: %s, found: %d",
				jobID, job.Status, job.TotalUrls,
			))
		}
		time.Sleep(pollInterval)
	}
}

// Context builds context from a search query.
func (c *AsyncWebCrawler) Context(query string, opts *ContextOptions) (*ContextResult, error) {
	if opts == nil {
		opts = &ContextOptions{}
	}

	paaLimit := opts.PAALimit
	if paaLimit == 0 {
		paaLimit = 3
	}

	resultsPerPAA := opts.ResultsPerPAA
	if resultsPerPAA == 0 {
		resultsPerPAA = 5
	}

	body := map[string]interface{}{
		"query":           query,
		"strategy":        "serper_paa",
		"paa_limit":       paaLimit,
		"results_per_paa": resultsPerPAA,
	}

	data, err := c.http.Post("/v1/context", body, 300*time.Second)
	if err != nil {
		return nil, err
	}

	return ContextResultFromMap(data), nil
}

// GenerateSchemaOptions are options for GenerateSchema.
type GenerateSchemaOptions struct {
	Query             string
	SchemaType        string // "CSS" or "XPATH"
	TargetJSONExample map[string]interface{}
	LLMConfig         map[string]interface{}
}

// GenerateSchema generates extraction schema from HTML using LLM.
//
// The html parameter can be:
//   - A single string: One HTML sample
//   - A []string slice: Multiple HTML samples for robust selector generation
//
// Example:
//
//	// Single HTML
//	schema, _ := crawler.GenerateSchema(page.HTML, &GenerateSchemaOptions{Query: "Extract products"})
//
//	// Multiple HTML samples
//	schema, _ := crawler.GenerateSchema([]string{page1.HTML, page2.HTML}, nil)
func (c *AsyncWebCrawler) GenerateSchema(html interface{}, opts *GenerateSchemaOptions) (*GeneratedSchema, error) {
	if opts == nil {
		opts = &GenerateSchemaOptions{}
	}

	schemaType := opts.SchemaType
	if schemaType == "" {
		schemaType = "CSS"
	}

	body := map[string]interface{}{
		"schema_type": schemaType,
	}

	// Handle different html types
	switch v := html.(type) {
	case string:
		body["html"] = v
	case []string:
		body["html"] = v
	default:
		return nil, fmt.Errorf("html must be string or []string, got %T", html)
	}

	if opts.Query != "" {
		body["query"] = opts.Query
	}
	if opts.TargetJSONExample != nil {
		body["target_json_example"] = opts.TargetJSONExample
	}
	if opts.LLMConfig != nil {
		body["llm_config"] = opts.LLMConfig
	}

	data, err := c.http.Post("/v1/schema/generate", body, 60*time.Second)
	if err != nil {
		return nil, err
	}

	return GeneratedSchemaFromMap(data), nil
}

// GenerateSchemaFromURLs generates extraction schema by fetching HTML from URLs.
//
// URLs are fetched in parallel via worker infrastructure (max 3 URLs).
// This is useful when you don't have the HTML content locally.
//
// Example:
//
//	schema, _ := crawler.GenerateSchemaFromURLs(
//	    []string{"https://example.com/p/1", "https://example.com/p/2"},
//	    &GenerateSchemaOptions{Query: "Extract product details"},
//	)
func (c *AsyncWebCrawler) GenerateSchemaFromURLs(urls []string, opts *GenerateSchemaOptions) (*GeneratedSchema, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("at least one URL is required")
	}
	if len(urls) > 3 {
		return nil, fmt.Errorf("maximum 3 URLs allowed, got %d", len(urls))
	}

	if opts == nil {
		opts = &GenerateSchemaOptions{}
	}

	schemaType := opts.SchemaType
	if schemaType == "" {
		schemaType = "CSS"
	}

	body := map[string]interface{}{
		"urls":        urls,
		"schema_type": schemaType,
	}

	if opts.Query != "" {
		body["query"] = opts.Query
	}
	if opts.TargetJSONExample != nil {
		body["target_json_example"] = opts.TargetJSONExample
	}
	if opts.LLMConfig != nil {
		body["llm_config"] = opts.LLMConfig
	}

	data, err := c.http.Post("/v1/schema/generate", body, 60*time.Second)
	if err != nil {
		return nil, err
	}

	return GeneratedSchemaFromMap(data), nil
}

// Storage gets current storage usage.
func (c *AsyncWebCrawler) Storage() (*StorageUsage, error) {
	data, err := c.http.Get("/v1/crawl/storage", nil)
	if err != nil {
		return nil, err
	}

	return StorageUsageFromMap(data), nil
}

// Health checks API health status.
func (c *AsyncWebCrawler) Health() (map[string]interface{}, error) {
	return c.http.Get("/health", nil)
}

// =========================================================================
// Wrapper API -- Simplified endpoints
// =========================================================================

// Markdown gets clean markdown from a URL.
func (c *AsyncWebCrawler) Markdown(url string, opts *MarkdownOptions) (*MarkdownResponse, error) {
	if opts == nil {
		opts = &MarkdownOptions{}
	}
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "browser"
	}
	fit := true
	if opts.Fit != nil {
		fit = *opts.Fit
	}

	body := map[string]interface{}{"url": url, "strategy": strategy, "fit": fit}
	if len(opts.Include) > 0 {
		body["include"] = opts.Include
	}
	if opts.CrawlerConfig != nil {
		body["crawler_config"] = opts.CrawlerConfig
	}
	if opts.BrowserConfig != nil {
		body["browser_config"] = opts.BrowserConfig
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}
	if opts.BypassCache {
		body["bypass_cache"] = true
	}

	data, err := c.http.Post("/v1/markdown", body, 0)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[MarkdownResponse](data)
}

// Screenshot captures a screenshot or PDF of a web page.
func (c *AsyncWebCrawler) Screenshot(url string, opts *ScreenshotOptions) (*ScreenshotResponse, error) {
	if opts == nil {
		opts = &ScreenshotOptions{}
	}
	fullPage := true
	if opts.FullPage != nil {
		fullPage = *opts.FullPage
	}

	body := map[string]interface{}{"url": url, "full_page": fullPage}
	if opts.PDF {
		body["pdf"] = true
	}
	if opts.WaitFor != "" {
		body["wait_for"] = opts.WaitFor
	}
	if opts.CrawlerConfig != nil {
		body["crawler_config"] = opts.CrawlerConfig
	}
	if opts.BrowserConfig != nil {
		body["browser_config"] = opts.BrowserConfig
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}
	if opts.BypassCache {
		body["bypass_cache"] = true
	}

	data, err := c.http.Post("/v1/screenshot", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[ScreenshotResponse](data)
}

// Extract extracts structured data from a web page.
func (c *AsyncWebCrawler) Extract(url string, opts *ExtractOptions) (*ExtractResponse, error) {
	if opts == nil {
		opts = &ExtractOptions{}
	}
	method := opts.Method
	if method == "" {
		method = "auto"
	}
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "http"
	}

	body := map[string]interface{}{"url": url, "method": method, "strategy": strategy}
	if opts.Query != "" {
		body["query"] = opts.Query
	}
	if opts.JSONExample != nil {
		body["json_example"] = opts.JSONExample
	}
	if opts.Schema != nil {
		body["schema"] = opts.Schema
	}
	if opts.CrawlerConfig != nil {
		body["crawler_config"] = opts.CrawlerConfig
	}
	if opts.BrowserConfig != nil {
		body["browser_config"] = opts.BrowserConfig
	}
	if opts.LLMConfig != nil {
		body["llm_config"] = opts.LLMConfig
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}
	if opts.BypassCache {
		body["bypass_cache"] = true
	}

	data, err := c.http.Post("/v1/extract", body, 180*time.Second)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[ExtractResponse](data)
}

// Map discovers all URLs on a domain.
func (c *AsyncWebCrawler) Map(url string, opts *MapOptions) (*MapResponse, error) {
	if opts == nil {
		opts = &MapOptions{}
	}
	mode := opts.Mode
	if mode == "" {
		mode = "default"
	}
	extractHead := true
	if opts.ExtractHead != nil {
		extractHead = *opts.ExtractHead
	}

	body := map[string]interface{}{
		"url": url, "mode": mode,
		"include_subdomains": opts.IncludeSubdomains,
		"extract_head":       extractHead,
	}
	if opts.MaxURLs != nil {
		body["max_urls"] = *opts.MaxURLs
	}
	if opts.Query != "" {
		body["query"] = opts.Query
	}
	if opts.ScoreThreshold != nil {
		body["score_threshold"] = *opts.ScoreThreshold
	}
	if opts.Force {
		body["force"] = true
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}

	data, err := c.http.Post("/v1/map", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[MapResponse](data)
}

// CrawlSite crawls an entire website. Always async.
//
// AI-assisted flagship flow: set opts.Criteria (plain English) and let the
// LLM pick the scan strategy, generate URL filters, and (optionally) build an
// extraction schema from a sample URL via opts.Extract. Poll one unified
// endpoint for both scan and crawl phases with GetSiteCrawlJob(), or pass
// opts.Wait = true to block until completion.
func (c *AsyncWebCrawler) CrawlSite(url string, opts *SiteCrawlOptions) (*SiteCrawlResponse, error) {
	if opts == nil {
		opts = &SiteCrawlOptions{}
	}
	maxPages := opts.MaxPages
	if maxPages == 0 {
		maxPages = 20
	}
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "browser"
	}
	fit := true
	if opts.Fit != nil {
		fit = *opts.Fit
	}
	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	body := map[string]interface{}{
		"url": url, "max_pages": maxPages,
		"strategy": strategy, "fit": fit, "priority": priority,
	}

	// --- AI-assisted fields (new) ---
	if opts.Criteria != "" {
		body["criteria"] = opts.Criteria
	}
	if opts.Scan != nil {
		body["scan"] = opts.Scan.ToMap()
	}
	if opts.Extract != nil {
		body["extract"] = opts.Extract.ToMap()
	}
	if opts.IncludeMarkdown != nil {
		body["include_markdown"] = *opts.IncludeMarkdown
	}

	// --- Legacy / backward-compat fields ---
	if opts.Discovery != "" && opts.Discovery != "map" {
		body["discovery"] = opts.Discovery
	}
	if len(opts.Include) > 0 {
		body["include"] = opts.Include
	}
	if opts.Pattern != "" {
		body["pattern"] = opts.Pattern
	}
	if opts.MaxDepth != nil {
		body["max_depth"] = *opts.MaxDepth
	}
	if opts.CrawlerConfig != nil {
		body["crawler_config"] = opts.CrawlerConfig
	}
	if opts.BrowserConfig != nil {
		body["browser_config"] = opts.BrowserConfig
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	// Longer timeout — extract triggers schema gen (sample fetch + LLM call)
	// which can take 30-120s before the job is even created.
	data, err := c.http.Post("/v1/crawl/site", body, 240*time.Second)
	if err != nil {
		return nil, err
	}
	result, err := unmarshalWrapper[SiteCrawlResponse](data)
	if err != nil {
		return nil, err
	}

	if opts.Wait && result.JobID != "" {
		final, err := c.waitSiteCrawlJob(result.JobID, opts.PollInterval, opts.Timeout)
		if err != nil {
			return result, err
		}
		// Transfer polled state back onto the response
		result.Status = final.Status
		result.DiscoveredURLs = final.Progress.UrlsDiscovered
	}

	return result, nil
}

// GetSiteCrawlJob polls a site crawl job started via CrawlSite(). This is the
// unified polling endpoint — it merges the scan phase (URL discovery) and the
// crawl phase (per-page fetch + extract) into one response. Phase walks
// through "scan" → "crawl" → "done".
func (c *AsyncWebCrawler) GetSiteCrawlJob(jobID string) (*SiteCrawlJobStatus, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/crawl/site/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[SiteCrawlJobStatus](data)
}

// waitSiteCrawlJob polls /v1/crawl/site/jobs/{id} until the crawl finishes.
func (c *AsyncWebCrawler) waitSiteCrawlJob(jobID string, pollInterval, timeout time.Duration) (*SiteCrawlJobStatus, error) {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	start := time.Now()
	for {
		job, err := c.GetSiteCrawlJob(jobID)
		if err != nil {
			return nil, err
		}
		if job.IsComplete() {
			return job, nil
		}
		if timeout > 0 && time.Since(start) > timeout {
			return nil, NewTimeoutError(fmt.Sprintf(
				"timeout waiting for site crawl %s. Phase: %s, crawled: %d/%d",
				jobID, job.Phase, job.Progress.UrlsCrawled, job.Progress.Total,
			))
		}
		time.Sleep(pollInterval)
	}
}

// =========================================================================
// Enrich API
// =========================================================================

// Enrich creates an enrichment job. Each URL becomes a row; the pipeline
// crawls each URL, follows links to find missing fields, and optionally
// searches Google. Always async -- set opts.Wait=true to block until done.
func (c *AsyncWebCrawler) Enrich(urls []string, schema []EnrichFieldSpec, opts *EnrichOptions) (*EnrichJobStatus, error) {
	if opts == nil {
		opts = &EnrichOptions{}
	}

	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = 1
	}
	maxLinks := opts.MaxLinks
	if maxLinks == 0 {
		maxLinks = 5
	}
	retryCount := opts.RetryCount
	if retryCount == 0 {
		retryCount = 1
	}
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "browser"
	}
	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	// Build schema as []map to match the API shape
	schemaList := make([]map[string]interface{}, len(schema))
	for i, f := range schema {
		m := map[string]interface{}{"name": f.Name}
		if f.Description != "" {
			m["description"] = f.Description
		}
		schemaList[i] = m
	}

	body := map[string]interface{}{
		"urls":   urls,
		"schema": schemaList,
		"config": map[string]interface{}{
			"max_depth":     maxDepth,
			"max_links":     maxLinks,
			"enable_search": opts.EnableSearch,
			"retry_count":   retryCount,
		},
		"strategy": strategy,
		"priority": priority,
	}
	if opts.LLMConfig != nil {
		body["llm_config"] = opts.LLMConfig
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	data, err := c.http.Post("/v1/enrich", body, 0)
	if err != nil {
		return nil, err
	}

	resp, err := unmarshalWrapper[EnrichResponse](data)
	if err != nil {
		return nil, err
	}

	if opts.Wait {
		pollInterval := opts.PollInterval
		if pollInterval == 0 {
			pollInterval = 3 * time.Second
		}
		return c.waitEnrichJob(resp.JobID, pollInterval, opts.Timeout)
	}

	// Return minimal status when not waiting
	return &EnrichJobStatus{
		JobID:     resp.JobID,
		Status:    resp.Status,
		CreatedAt: resp.CreatedAt,
	}, nil
}

// GetEnrichJob polls an enrichment job for status and completed rows.
func (c *AsyncWebCrawler) GetEnrichJob(jobID string) (*EnrichJobStatus, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/enrich/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[EnrichJobStatus](data)
}

// ListEnrichJobs lists enrichment jobs for the authenticated user.
func (c *AsyncWebCrawler) ListEnrichJobs(limit, offset int) ([]*EnrichJobStatus, error) {
	if limit == 0 {
		limit = 20
	}
	params := map[string]string{
		"limit":  fmt.Sprintf("%d", limit),
		"offset": fmt.Sprintf("%d", offset),
	}

	data, err := c.http.Get("/v1/enrich/jobs", params)
	if err != nil {
		return nil, err
	}

	jobs := make([]*EnrichJobStatus, 0)
	if rawJobs, ok := data["jobs"].([]interface{}); ok {
		for _, j := range rawJobs {
			if m, ok := j.(map[string]interface{}); ok {
				job, err := unmarshalWrapper[EnrichJobStatus](m)
				if err == nil {
					jobs = append(jobs, job)
				}
			}
		}
	}
	return jobs, nil
}

// CancelEnrichJob cancels an enrichment job.
func (c *AsyncWebCrawler) CancelEnrichJob(jobID string) error {
	_, err := c.http.Delete(fmt.Sprintf("/v1/enrich/jobs/%s", jobID))
	return err
}

// waitEnrichJob polls /v1/enrich/jobs/{id} until the job finishes.
func (c *AsyncWebCrawler) waitEnrichJob(jobID string, pollInterval, timeout time.Duration) (*EnrichJobStatus, error) {
	if pollInterval == 0 {
		pollInterval = 3 * time.Second
	}
	start := time.Now()
	for {
		job, err := c.GetEnrichJob(jobID)
		if err != nil {
			return nil, err
		}
		if job.IsComplete() {
			return job, nil
		}
		if timeout > 0 && time.Since(start) > timeout {
			return nil, NewTimeoutError(fmt.Sprintf(
				"enrich job %s did not complete within %v. Status: %s, progress: %d/%d",
				jobID, timeout, job.Status, job.Progress.Completed, job.Progress.Total,
			))
		}
		time.Sleep(pollInterval)
	}
}

// ---- Wrapper job management ----

// GetMarkdownJob gets a markdown async job status.
func (c *AsyncWebCrawler) GetMarkdownJob(jobID string) (*WrapperJob, error) {
	return c.getWrapperJob(jobID, "markdown")
}

// GetScreenshotJob gets a screenshot async job status.
func (c *AsyncWebCrawler) GetScreenshotJob(jobID string) (*WrapperJob, error) {
	return c.getWrapperJob(jobID, "screenshot")
}

// GetExtractJob gets an extract async job status.
func (c *AsyncWebCrawler) GetExtractJob(jobID string) (*WrapperJob, error) {
	return c.getWrapperJob(jobID, "extract")
}

// CancelMarkdownJob cancels a markdown async job.
func (c *AsyncWebCrawler) CancelMarkdownJob(jobID string) error {
	return c.cancelWrapperJob(jobID, "markdown")
}

// CancelScreenshotJob cancels a screenshot async job.
func (c *AsyncWebCrawler) CancelScreenshotJob(jobID string) error {
	return c.cancelWrapperJob(jobID, "screenshot")
}

// CancelExtractJob cancels an extract async job.
func (c *AsyncWebCrawler) CancelExtractJob(jobID string) error {
	return c.cancelWrapperJob(jobID, "extract")
}

func (c *AsyncWebCrawler) getWrapperJob(jobID, jobType string) (*WrapperJob, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/%s/jobs/%s", jobType, jobID), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[WrapperJob](data)
}

func (c *AsyncWebCrawler) cancelWrapperJob(jobID, jobType string) error {
	_, err := c.http.Delete(fmt.Sprintf("/v1/%s/jobs/%s", jobType, jobID))
	return err
}

// unmarshalWrapper converts a map response to a typed struct via JSON round-trip.
func unmarshalWrapper[T any](data map[string]interface{}) (*T, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	var result T
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &result, nil
}

// Close closes the crawler (no-op in Go, but provided for API compatibility).
func (c *AsyncWebCrawler) Close() error {
	return nil
}
