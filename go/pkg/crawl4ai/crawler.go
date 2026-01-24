// Package crawl4ai provides a Go SDK for Crawl4AI Cloud API.
package crawl4ai

import (
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

	body := map[string]interface{}{
		"strategy":       strategy,
		"crawl_strategy": crawlStrategy,
		"priority":       priority,
	}

	if url != "" {
		body["url"] = url
	}
	if opts.SourceJob != "" {
		body["source_job_id"] = opts.SourceJob
	}

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

	discoveredCount := 0
	if v, ok := data["discovered_urls"].(float64); ok {
		discoveredCount = int(v)
	}

	strategy := ""
	if v, ok := data["strategy"].(string); ok {
		strategy = v
	}

	return &DeepCrawlResult{
		JobID:           jobID,
		Status:          "cancelled",
		Strategy:        strategy,
		DiscoveredCount: discoveredCount,
	}, nil
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

// Close closes the crawler (no-op in Go, but provided for API compatibility).
func (c *AsyncWebCrawler) Close() error {
	return nil
}
