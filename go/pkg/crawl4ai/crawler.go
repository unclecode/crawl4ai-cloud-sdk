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
func (c *AsyncWebCrawler) RunMany(urls []string, opts *RunManyOptions) (*RunManyResult, error) {
	if opts == nil {
		opts = &RunManyOptions{}
	}

	if len(urls) <= 10 {
		return c.runBatch(urls, opts)
	}
	return c.runAsync(urls, opts)
}

// ArunMany is an alias for RunMany (OSS compatibility).
func (c *AsyncWebCrawler) ArunMany(urls []string, opts *RunManyOptions) (*RunManyResult, error) {
	return c.RunMany(urls, opts)
}

func (c *AsyncWebCrawler) runBatch(urls []string, opts *RunManyOptions) (*RunManyResult, error) {
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "browser"
	}

	body := BuildCrawlRequest(map[string]interface{}{
		"urls":          urls,
		"config":        opts.Config,
		"browserConfig": opts.BrowserConfig,
		"strategy":      strategy,
		"proxy":         opts.Proxy,
		"bypassCache":   opts.BypassCache,
	})

	data, err := c.http.Post("/v1/crawl/batch", body, 600*time.Second)
	if err != nil {
		return nil, err
	}

	results := make([]*CrawlResult, 0)
	if rawResults, ok := data["results"].([]interface{}); ok {
		for _, r := range rawResults {
			if m, ok := r.(map[string]interface{}); ok {
				results = append(results, CrawlResultFromMap(m))
			}
		}
	}

	if opts.Wait {
		return &RunManyResult{Results: results}, nil
	}

	// Wrap in completed job
	job := &CrawlJob{
		ID:        fmt.Sprintf("batch_%d", time.Now().Unix()),
		Status:    "completed",
		URLsCount: len(urls),
		Progress: JobProgress{
			Total:     len(urls),
			Completed: len(urls),
			Failed:    0,
		},
	}
	return &RunManyResult{Job: job, Results: results}, nil
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

		job, err = c.WaitJob(job.ID, pollInterval, opts.Timeout, true)
		if err != nil {
			return nil, err
		}

		if job.Results != nil {
			results := make([]*CrawlResult, len(job.Results))
			for i, r := range job.Results {
				results[i] = CrawlResultFromMap(r)
			}
			return &RunManyResult{Job: job, Results: results}, nil
		}
		return &RunManyResult{Job: job}, nil
	}

	return &RunManyResult{Job: job}, nil
}

// GetJob gets job status and optionally results.
func (c *AsyncWebCrawler) GetJob(jobID string, includeResults bool) (*CrawlJob, error) {
	params := make(map[string]string)
	if includeResults {
		params["include_results"] = "true"
	}

	data, err := c.http.Get(fmt.Sprintf("/v1/crawl/jobs/%s", jobID), params)
	if err != nil {
		return nil, err
	}

	return CrawlJobFromMap(data), nil
}

// WaitJob polls until job completes.
func (c *AsyncWebCrawler) WaitJob(jobID string, pollInterval, timeout time.Duration, includeResults bool) (*CrawlJob, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}

	startTime := time.Now()

	for {
		job, err := c.GetJob(jobID, false)
		if err != nil {
			return nil, err
		}

		if job.IsComplete() {
			if includeResults {
				return c.GetJob(jobID, true)
			}
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
		if opts.Filters != nil {
			body["filters"] = opts.Filters
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
		job, err := c.WaitJob(result.CrawlJobID, pollInterval, opts.Timeout, true)
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
func (c *AsyncWebCrawler) GenerateSchema(html string, opts *GenerateSchemaOptions) (*GeneratedSchema, error) {
	if opts == nil {
		opts = &GenerateSchemaOptions{}
	}

	schemaType := opts.SchemaType
	if schemaType == "" {
		schemaType = "CSS"
	}

	body := map[string]interface{}{
		"html":        html,
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
