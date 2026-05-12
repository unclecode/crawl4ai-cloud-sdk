// Package crawl4ai provides a Go SDK for Crawl4AI Cloud API.
package crawl4ai

import (
	"context"
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

// SiteOptions are options for Site (the canonical /v1/site endpoint).
type SiteOptions struct {
	Mode              string // "map" (sync sitemap discovery) | "traverse" (async, default)
	MaxURLs           int
	MaxDepth          int
	ScanOnly          bool
	Patterns          []string
	Filters           map[string]interface{}
	Scorers           map[string]interface{}
	Query             string
	ScoreThreshold    *float64
	IncludeSubdomains bool
	CrawlerConfig     *CrawlerRunConfig
	BrowserConfig     *BrowserConfig
	Proxy             interface{}
	WebhookURL        string
	Priority          int
	Wait              bool
	PollInterval      time.Duration
	Timeout           time.Duration
}

// Site discovers or crawls an entire site (canonical, /v1/site).
//
// Mode "map" runs sync sitemap-based URL discovery. Mode "traverse" (default)
// returns a job_id immediately and runs best-first recursive crawl across the
// worker pool via the cloud's WorkerPoolDispatcher.
func (c *AsyncWebCrawler) Site(url string, opts *SiteOptions) (*DeepCrawlResult, error) {
	if opts == nil {
		opts = &SiteOptions{}
	}
	mode := opts.Mode
	if mode == "" {
		mode = "traverse"
	}
	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	body := map[string]interface{}{
		"url":                url,
		"mode":               mode,
		"scan_only":          opts.ScanOnly,
		"include_subdomains": opts.IncludeSubdomains,
		"priority":           priority,
	}
	if opts.MaxURLs > 0 {
		body["max_urls"] = opts.MaxURLs
	}
	if opts.MaxDepth > 0 {
		body["max_depth"] = opts.MaxDepth
	}
	if len(opts.Patterns) > 0 {
		body["patterns"] = opts.Patterns
	}
	if opts.Filters != nil {
		body["filters"] = opts.Filters
	}
	if opts.Scorers != nil {
		body["scorers"] = opts.Scorers
	}
	if opts.Query != "" {
		body["query"] = opts.Query
	}
	if opts.ScoreThreshold != nil {
		body["score_threshold"] = *opts.ScoreThreshold
	}
	if opts.CrawlerConfig != nil {
		if cc := SanitizeCrawlerConfig(opts.CrawlerConfig); len(cc) > 0 {
			body["crawler_config"] = cc
		}
	}
	if opts.BrowserConfig != nil {
		if bc := SanitizeBrowserConfig(opts.BrowserConfig, ""); len(bc) > 0 {
			body["browser_config"] = bc
		}
	}
	if opts.Proxy != nil {
		body["proxy"] = opts.Proxy
	}
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	data, err := c.http.Post("/v1/site", body, 120*time.Second)
	if err != nil {
		return nil, err
	}
	result := DeepCrawlResultFromMap(data)

	terminal := result.Status == "completed" || result.Status == "cancelled" || result.Status == "failed"
	if !opts.Wait || terminal {
		return result, nil
	}

	// Traverse mode — wait for scan to finish, then crawl phase if any
	scan, err := c.waitSiteJob(result.JobID, opts.PollInterval, opts.Timeout)
	if err != nil {
		return result, err
	}
	return scan, nil
}

// waitSiteJob polls /v1/site/jobs/{id} until the scan completes.
func (c *AsyncWebCrawler) waitSiteJob(jobID string, pollInterval, timeout time.Duration) (*DeepCrawlResult, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}
	start := time.Now()
	for {
		data, err := c.http.Get(fmt.Sprintf("/v1/site/jobs/%s", jobID), nil)
		if err != nil {
			return nil, err
		}
		result := DeepCrawlResultFromMap(data)
		if result.Status == "completed" || result.Status == "cancelled" || result.Status == "failed" {
			return result, nil
		}
		if timeout > 0 && time.Since(start) > timeout {
			return result, fmt.Errorf("timeout waiting for site job %s (status: %s)", jobID, result.Status)
		}
		time.Sleep(pollInterval)
	}
}

// CancelSite cancels a running site (traverse) job.
func (c *AsyncWebCrawler) CancelSite(jobID string) (*DeepCrawlResult, error) {
	data, err := c.http.Post(fmt.Sprintf("/v1/site/jobs/%s/cancel", jobID), nil, 0)
	if err != nil {
		return nil, err
	}
	return DeepCrawlResultFromMap(data), nil
}

// GetSiteStatus returns the status of a site (traverse) job.
func (c *AsyncWebCrawler) GetSiteStatus(jobID string) (*DeepCrawlResult, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/site/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}
	return DeepCrawlResultFromMap(data), nil
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
//
// /v1/crawl/deep is now a server-side alias for /v1/site (Phase 4).
// DeepCrawl() is kept as a back-compat alias — no warning.
// New code should call Site() directly.
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
		// mode='deep' was the recursive-scan path that has moved to /v1/site.
		// Hard guard: refuse client-side rather than 422 on the wire.
		if opts.Mode == "deep" {
			return nil, fmt.Errorf(
				"Scan(mode=\"deep\") is no longer supported — recursive URL " +
					"discovery moved to Site(url, &SiteOptions{Mode: \"traverse\", ScanOnly: true, ...}). " +
					"For one-shot sitemap discovery, leave Mode empty and call Scan(url, ...) directly.",
			)
		}
		// Sources (canonical) — Mode (other than 'deep') maps to 'primary' for back-compat.
		sources := opts.Sources
		if sources == "" {
			sources = "primary"
		}
		body["sources"] = sources

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

// waitWrapperJob polls /v1/{type}/jobs/{id} until the async wrapper job
// finishes, then auto-hydrates job.Results from the per-URL S3 endpoint.
// jobType is one of "markdown" / "screenshot" / "extract".
//
// The cloud GET endpoint returns URLStatuses for fan-out parents but never
// inlines per-URL data — that lives in S3 and is fetched separately. Wait=true
// callers expect job.Results populated, so we hydrate here. Failed URLs become
// CrawlResult stubs (Success=false + ErrorMessage) so len(Results) always
// equals len(URLStatuses).
func (c *AsyncWebCrawler) waitWrapperJob(jobID, jobType string, pollInterval, timeout time.Duration) (*WrapperJob, error) {
	if pollInterval == 0 {
		pollInterval = 2 * time.Second
	}
	// /v1/markdown/jobs path stays for back-compat, but new code should hit /v1/scrape/jobs.
	pathPrefix := "/v1/" + jobType
	if jobType == "markdown" {
		pathPrefix = "/v1/scrape"
	}
	start := time.Now()
	for {
		data, err := c.http.Get(pathPrefix+"/jobs/"+jobID, nil)
		if err != nil {
			return nil, err
		}
		job, err := unmarshalWrapper[WrapperJob](data)
		if err != nil {
			return nil, err
		}
		if job.IsComplete() {
			if len(job.URLStatuses) > 0 {
				if hydrated, herr := c.hydrateResults(job); herr == nil {
					job.Results = hydrated
				}
				// Hydration failure is non-fatal — the parent job already
				// terminalized successfully; caller can still drive
				// GetPerUrlResult themselves with URLStatuses.
			}
			return job, nil
		}
		if timeout > 0 && time.Since(start) > timeout {
			return nil, NewTimeoutError(fmt.Sprintf(
				"timeout waiting for %s job %s (status: %s)",
				jobType, jobID, job.Status,
			))
		}
		time.Sleep(pollInterval)
	}
}

// hydrateResults fetches each URL's CrawlResult in parallel via the
// recipe-agnostic per-URL endpoint. Failed URLs get a stub so the slice
// stays aligned with URLStatuses.
func (c *AsyncWebCrawler) hydrateResults(job *WrapperJob) ([]CrawlResult, error) {
	results := make([]CrawlResult, len(job.URLStatuses))
	type fetchResult struct {
		idx int
		res CrawlResult
	}
	ch := make(chan fetchResult, len(job.URLStatuses))
	for i, entry := range job.URLStatuses {
		go func(i int, entry UrlStatus) {
			if entry.Status == "failed" {
				ms := 0
				if entry.DurationMs != nil {
					ms = *entry.DurationMs
				}
				errMsg := entry.Error
				if errMsg == "" {
					errMsg = "URL failed"
				}
				ch <- fetchResult{idx: i, res: CrawlResult{
					URL: entry.URL, Success: false, ErrorMessage: errMsg, DurationMs: ms,
				}}
				return
			}
			r, err := c.GetPerUrlResult(job.JobID, entry.Index)
			if err != nil {
				ms := 0
				if entry.DurationMs != nil {
					ms = *entry.DurationMs
				}
				ch <- fetchResult{idx: i, res: CrawlResult{
					URL: entry.URL, Success: false,
					ErrorMessage: fmt.Sprintf("per-URL fetch failed: %v", err),
					DurationMs:   ms,
				}}
				return
			}
			ch <- fetchResult{idx: i, res: *r}
		}(i, entry)
	}
	for range job.URLStatuses {
		fr := <-ch
		results[fr.idx] = fr.res
	}
	return results, nil
}

// GetPerUrlResult fetches one URL's full result from a multi-URL fan-out
// parent. Recipe-agnostic — works for any wrapper async parent (scrape /
// screenshot / extract / crawl). Children all write to a unified S3 prefix
// keyed on (jobID, urlIndex), so the path is the same regardless of which
// wrapper created the parent.
//
// urlIndex is the 0-based index into the parent's submitted URL list; match
// against entries in job.URLStatuses.
//
// Returns CrawlResult. Markdown is populated for scrape jobs, Screenshot
// (base64) for screenshot jobs, ExtractedContent for extract jobs.
func (c *AsyncWebCrawler) GetPerUrlResult(jobID string, urlIndex int) (*CrawlResult, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/crawl/jobs/%s/result/%d", jobID, urlIndex), nil)
	if err != nil {
		return nil, err
	}
	// CrawlResultFromMap handles the string-or-object polymorphism on the
	// `markdown` field (sync returns a struct, async returns a raw string).
	// Plain json.Unmarshal would choke on the string form.
	return CrawlResultFromMap(data), nil
}

// ScreenshotAsync submits an async screenshot job over a list of URLs.
// POST /v1/screenshot/async.
func (c *AsyncWebCrawler) ScreenshotAsync(urls []string, opts *ScreenshotAsyncOptions) (*WrapperJob, error) {
	if opts == nil {
		opts = &ScreenshotAsyncOptions{}
	}
	fullPage := true
	if opts.FullPage != nil {
		fullPage = *opts.FullPage
	}
	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	body := map[string]interface{}{"urls": urls, "full_page": fullPage, "priority": priority}
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
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	data, err := c.http.Post("/v1/screenshot/async", body, 0)
	if err != nil {
		return nil, err
	}
	job, err := unmarshalWrapper[WrapperJob](data)
	if err != nil {
		return nil, err
	}
	if opts.Wait {
		return c.waitWrapperJob(job.JobID, "screenshot", opts.PollInterval, opts.Timeout)
	}
	return job, nil
}

// ExtractAsync submits an async extract job over one base URL plus optional followers.
// POST /v1/extract/async.
//
// The base url is the schema TEMPLATE in css_schema mode — sampled once, schema generated,
// then re-applied across opts.ExtraURLs for free. In method="llm" mode the base has no
// special role; every URL gets its own LLM call. Up to 100 URLs total (1 base + 99 extras).
func (c *AsyncWebCrawler) ExtractAsync(url string, opts *ExtractAsyncOptions) (*WrapperJob, error) {
	if opts == nil {
		opts = &ExtractAsyncOptions{}
	}
	method := opts.Method
	if method == "" {
		method = "auto"
	}
	strategy := opts.Strategy
	if strategy == "" {
		strategy = "http"
	}
	priority := opts.Priority
	if priority == 0 {
		priority = 5
	}

	body := map[string]interface{}{"url": url, "method": method, "strategy": strategy, "priority": priority}
	if len(opts.ExtraURLs) > 0 {
		body["extra_urls"] = opts.ExtraURLs
	}
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
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	data, err := c.http.Post("/v1/extract/async", body, 0)
	if err != nil {
		return nil, err
	}
	job, err := unmarshalWrapper[WrapperJob](data)
	if err != nil {
		return nil, err
	}
	if opts.Wait {
		return c.waitWrapperJob(job.JobID, "extract", opts.PollInterval, opts.Timeout)
	}
	return job, nil
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
// Scrape fetches a page and returns clean markdown plus optional extras.
// POST /v1/scrape (sync, single URL). Use ScrapeAsync for batch / webhooks.
func (c *AsyncWebCrawler) Scrape(url string, opts *MarkdownOptions) (*MarkdownResponse, error) {
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

	data, err := c.http.Post("/v1/scrape", body, 0)
	if err != nil {
		return nil, err
	}
	return unmarshalWrapper[MarkdownResponse](data)
}

// Markdown is the legacy alias for Scrape — same shape, same response.
//
// Deprecated: Use Scrape. /v1/markdown was renamed to /v1/scrape. Will be removed in 0.8.0.
func (c *AsyncWebCrawler) Markdown(url string, opts *MarkdownOptions) (*MarkdownResponse, error) {
	return c.Scrape(url, opts)
}

// ScrapeAsync submits an async scrape job over a list of URLs.
// POST /v1/scrape/async. Returns a job; set opts.Wait = true to block until terminal.
func (c *AsyncWebCrawler) ScrapeAsync(urls []string, opts *ScrapeAsyncOptions) (*WrapperJob, error) {
	if opts == nil {
		opts = &ScrapeAsyncOptions{}
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

	body := map[string]interface{}{"urls": urls, "strategy": strategy, "fit": fit, "priority": priority}
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
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}

	data, err := c.http.Post("/v1/scrape/async", body, 0)
	if err != nil {
		return nil, err
	}
	job, err := unmarshalWrapper[WrapperJob](data)
	if err != nil {
		return nil, err
	}
	if opts.Wait {
		return c.waitWrapperJob(job.JobID, "markdown", opts.PollInterval, opts.Timeout)
	}
	return job, nil
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
	sources := opts.Sources
	if sources == "" {
		sources = "primary"
	}
	// Legacy Mode → Sources translation. Mode is deprecated.
	if opts.Mode != "" {
		if opts.Mode == "deep" {
			sources = "extended"
		} else {
			sources = "primary"
		}
	}
	extractHead := true
	if opts.ExtractHead != nil {
		extractHead = *opts.ExtractHead
	}

	body := map[string]interface{}{
		"url": url, "sources": sources,
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

// CrawlSite is no longer supported.
//
// /v1/crawl/site was removed (zero traffic for 14 days, deletion approved
// 2026-05). Returns an error pointing at the canonical alternatives:
//   - Site(url, &SiteOptions{Mode: "traverse", MaxURLs: ...}) for recursive crawling
//   - Scan(url, &ScanOptions{Criteria: ...}) → ExtractAsync(first, &ExtractAsyncOptions{ExtraURLs: rest})
//     for AI-assisted discovery + extraction
func (c *AsyncWebCrawler) CrawlSite(_ string, _ *SiteCrawlOptions) (*SiteCrawlResponse, error) {
	return nil, fmt.Errorf(
		"CrawlSite() — the /v1/crawl/site endpoint was removed. " +
			"For recursive site crawling, use Site(url, &SiteOptions{Mode: \"traverse\", MaxURLs: N}). " +
			"For AI-assisted discovery + extraction, pipeline Scan(url, &ScanOptions{Criteria: ...}) → " +
			"ExtractAsync(first, &ExtractAsyncOptions{ExtraURLs: rest}).",
	)
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
// Enrich v2 API (multi-phase)
// =========================================================================

// Enrich creates a multi-phase enrichment job.
//
// Defaults AutoConfirmPlan and AutoConfirmURLs to true (set the pointer
// fields to override) — the worker runs the full pipeline in one shot.
// Set either to false for human-in-loop review and resume via
// ResumeEnrichJob.
//
// Examples:
//
//	// Agent one-shot
//	result, _ := crawler.Enrich(&crawl4ai.EnrichOptions{
//	    Query:   "licensed nurseries in North York Toronto",
//	    Country: "ca",
//	    Wait:    true,
//	})
//	for _, row := range result.PhaseData.Rows {
//	    fmt.Println(row.InputKey, row.Fields)
//	}
//
//	// Pre-resolved URLs
//	result, _ := crawler.Enrich(&crawl4ai.EnrichOptions{
//	    URLs:     []string{"https://example.com/a", "https://example.com/b"},
//	    Features: []crawl4ai.EnrichFeature{{Name: "price"}, {Name: "hours"}},
//	    Wait:     true,
//	})
func (c *AsyncWebCrawler) Enrich(opts *EnrichOptions) (*EnrichJobStatus, error) {
	if opts == nil {
		opts = &EnrichOptions{}
	}
	body := buildEnrichRequest(opts)

	data, err := c.http.Post("/v1/enrich/async", body, 0)
	if err != nil {
		return nil, err
	}
	job, err := unmarshalEnrichJobStatus(data)
	if err != nil {
		return nil, err
	}

	if opts.Wait {
		waitOpts := WaitEnrichOptions{
			PollInterval: opts.PollInterval,
			Timeout:      opts.Timeout,
		}
		return c.WaitEnrichJob(job.JobID, waitOpts)
	}
	return job, nil
}

// GetEnrichJob fetches the current status of an enrichment job — one poll, no wait.
func (c *AsyncWebCrawler) GetEnrichJob(jobID string) (*EnrichJobStatus, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/enrich/jobs/%s", jobID), nil)
	if err != nil {
		return nil, err
	}
	return unmarshalEnrichJobStatus(data)
}

// WaitEnrichJob polls an enrichment job until it reaches opts.Until or a
// terminal status. Defaults: PollInterval=3s, Timeout=10m, Until="" (any
// terminal status).
//
// If Until is set and the job pauses at plan_ready or urls_ready without
// auto-confirm, the paused state is returned immediately rather than
// spinning until timeout.
func (c *AsyncWebCrawler) WaitEnrichJob(jobID string, opts WaitEnrichOptions) (*EnrichJobStatus, error) {
	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 3 * time.Second
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
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
		if opts.Until != "" && job.Status == opts.Until {
			return job, nil
		}
		if opts.Until != "" && job.IsPaused() {
			if (job.Status == EnrichStatusPlanReady && !job.AutoConfirmPlan) ||
				(job.Status == EnrichStatusURLsReady && !job.AutoConfirmURLs) {
				return job, nil
			}
		}
		if time.Since(start) > timeout {
			until := opts.Until
			if until == "" {
				until = "completed"
			}
			return nil, NewTimeoutError(fmt.Sprintf(
				"enrich job %s did not reach %q within %v. Status: %s, progress: %d/%d",
				jobID, until, timeout, job.Status, job.Progress.CompletedURLs, job.Progress.TotalURLs,
			))
		}
		time.Sleep(pollInterval)
	}
}

// ResumeEnrichJob advances a paused job (plan_ready or urls_ready) to the next phase.
//
// Pass nil to resume with the server's current values. At plan_ready: edit
// Entities/Criteria/Features. At urls_ready: edit Groups.
func (c *AsyncWebCrawler) ResumeEnrichJob(jobID string, opts *ResumeEnrichOptions) (*EnrichJobStatus, error) {
	body := map[string]interface{}{}
	if opts != nil {
		if opts.Entities != nil {
			body["entities"] = opts.Entities
		}
		if opts.Criteria != nil {
			body["criteria"] = opts.Criteria
		}
		if opts.Features != nil {
			body["features"] = opts.Features
		}
		if opts.Groups != nil {
			body["groups"] = opts.Groups
		}
	}
	data, err := c.http.Post(fmt.Sprintf("/v1/enrich/jobs/%s/continue", jobID), body, 0)
	if err != nil {
		return nil, err
	}
	return unmarshalEnrichJobStatus(data)
}

// StreamEnrichJob opens an SSE stream for an enrichment job. Events arrive
// on the returned channel until the server sends "complete" or the context
// is cancelled.
func (c *AsyncWebCrawler) StreamEnrichJob(ctx context.Context, jobID string) (<-chan EnrichEvent, error) {
	raw, err := c.http.StreamSse(ctx, fmt.Sprintf("/v1/enrich/jobs/%s/stream", jobID), nil)
	if err != nil {
		return nil, err
	}
	out := make(chan EnrichEvent, 16)
	go func() {
		defer close(out)
		for sse := range raw {
			if sse.Err != nil {
				// Stream broke — surface as a final event with empty type
				select {
				case out <- EnrichEvent{Type: "", Raw: map[string]interface{}{"error": sse.Err.Error()}}:
				case <-ctx.Done():
				}
				return
			}
			evt := decodeEnrichEvent(sse.Event, sse.Data)
			select {
			case out <- evt:
			case <-ctx.Done():
				return
			}
			if evt.Type == "complete" {
				return
			}
		}
	}()
	return out, nil
}

// CancelEnrichJob cancels a running enrichment job.
func (c *AsyncWebCrawler) CancelEnrichJob(jobID string) error {
	_, err := c.http.Delete(fmt.Sprintf("/v1/enrich/jobs/%s", jobID))
	return err
}

// ListEnrichJobs lists enrichment jobs for the authenticated user.
func (c *AsyncWebCrawler) ListEnrichJobs(limit, offset int) ([]*EnrichJobListItem, error) {
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
	jobs := make([]*EnrichJobListItem, 0)
	if rawJobs, ok := data["jobs"].([]interface{}); ok {
		for _, j := range rawJobs {
			if m, ok := j.(map[string]interface{}); ok {
				item, err := unmarshalEnrichJobListItem(m)
				if err == nil {
					jobs = append(jobs, item)
				}
			}
		}
	}
	return jobs, nil
}

// ─── Internal helpers (kept here to match how other services serialise) ───

func buildEnrichRequest(o *EnrichOptions) map[string]interface{} {
	autoPlan := true
	if o.AutoConfirmPlan != nil {
		autoPlan = *o.AutoConfirmPlan
	}
	autoURLs := true
	if o.AutoConfirmURLs != nil {
		autoURLs = *o.AutoConfirmURLs
	}
	topK := o.TopKPerEntity
	if topK == 0 {
		topK = 3
	}
	search := true
	if o.Search != nil {
		search = *o.Search
	}
	strategy := o.Strategy
	if strategy == "" {
		strategy = "http"
	}
	priority := o.Priority
	if priority == 0 {
		priority = 5
	}

	body := map[string]interface{}{
		"auto_confirm_plan": autoPlan,
		"auto_confirm_urls": autoURLs,
		"top_k_per_entity":  topK,
		"search":            search,
		"strategy":          strategy,
		"priority":          priority,
	}
	if o.Query != "" {
		body["query"] = o.Query
	}
	if o.Entities != nil {
		body["entities"] = o.Entities
	}
	if o.Criteria != nil {
		body["criteria"] = o.Criteria
	}
	if o.Features != nil {
		body["features"] = o.Features
	}
	if o.URLs != nil {
		body["urls"] = o.URLs
	}
	if o.Groups != nil {
		body["groups"] = o.Groups
	}
	if o.Country != "" {
		body["country"] = o.Country
	}
	if o.LocationHint != "" {
		body["location_hint"] = o.LocationHint
	}
	if o.Config != nil {
		body["config"] = o.Config
	}
	if o.BrowserConfig != nil {
		body["browser_config"] = o.BrowserConfig
	}
	if o.CrawlerConfig != nil {
		body["crawler_config"] = o.CrawlerConfig
	}
	if o.LLMConfig != nil {
		body["llm_config"] = o.LLMConfig
	}
	if o.Proxy != nil {
		body["proxy"] = o.Proxy
	}
	if o.WebhookURL != "" {
		body["webhook_url"] = o.WebhookURL
	}
	return body
}

// unmarshalEnrichJobStatus round-trips through JSON so nested heterogeneous
// fields (like sources / certainty maps) preserve their typed shapes.
func unmarshalEnrichJobStatus(m map[string]interface{}) (*EnrichJobStatus, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, NewCloudError(fmt.Sprintf("marshal enrich status: %v", err), 0, nil, nil)
	}
	out := &EnrichJobStatus{}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, NewCloudError(fmt.Sprintf("unmarshal enrich status: %v", err), 0, nil, nil)
	}
	// Sane defaults when the server omits the auto_confirm fields entirely
	if _, ok := m["auto_confirm_plan"]; !ok {
		out.AutoConfirmPlan = true
	}
	if _, ok := m["auto_confirm_urls"]; !ok {
		out.AutoConfirmURLs = true
	}
	return out, nil
}

func unmarshalEnrichJobListItem(m map[string]interface{}) (*EnrichJobListItem, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, NewCloudError(fmt.Sprintf("marshal list item: %v", err), 0, nil, nil)
	}
	out := &EnrichJobListItem{}
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, NewCloudError(fmt.Sprintf("unmarshal list item: %v", err), 0, nil, nil)
	}
	return out, nil
}

func decodeEnrichEvent(name string, payload map[string]interface{}) EnrichEvent {
	out := EnrichEvent{Type: name, Raw: payload}
	if name == "snapshot" {
		if snap, err := unmarshalEnrichJobStatus(payload); err == nil {
			out.Snapshot = snap
		}
	}
	if s, ok := payload["status"].(string); ok {
		out.Status = s
	}
	if rowMap, ok := payload["row"].(map[string]interface{}); ok {
		raw, err := json.Marshal(rowMap)
		if err == nil {
			row := &EnrichRow{}
			if err := json.Unmarshal(raw, row); err == nil {
				out.Row = row
			}
		}
	}
	if frag, ok := payload["fragment"].(map[string]interface{}); ok {
		out.Fragment = frag
	}
	return out
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

// ─── Discovery — wrapper-services platform ──────────────────────────
//
// One method, dispatches to any registered vertical via /v1/discovery/<service>.
// `search` is live; `people` / `products` / `posts` / `videos` will land
// via the same call shape — Discovery() never needs an SDK update when
// a new vertical ships.

// Discovery runs a vertical and returns the wire-shape response as a
// generic map. For service="search", call DiscoverySearch instead to
// get a typed *SearchResponse.
func (c *AsyncWebCrawler) Discovery(service string, params map[string]interface{}) (map[string]interface{}, error) {
	body := filterDiscoveryParams(params)
	return c.http.Post("/v1/discovery/"+service, body, 60*time.Second)
}

// DiscoverySearch is the typed convenience wrapper for service="search".
// Returns *SearchResponse parsed from /v1/discovery/search. For any
// other vertical, use Discovery() and read the generic map.
//
// Per-vertical request fields for service="search":
//   - "query" (string, required, 1-500 chars)
//   - "country" / "language" / "location"
//   - "num" / "start" / "site" / "mode" / "time_period"
//   - "backends" ([]string, 1-4 names from "google" / "bing" /
//     "duckduckgo" / "brave"; omit for the server's default of
//     ["google"]; >1 fans out + merges via RRF + URL dedup)
//   - "use_cache" (bool, default false — opt into the SERP cache;
//     legacy "bypass_cache" still wins when true)
//   - "enhance_query" (bool, default false — LLM rewrites the query
//     into a backend-specific operator-rich query before the SERP
//     fetch. Response carries OriginalQuery + RewrittenQueries so
//     callers see what hit the SERP. Never fails the search —
//     provider errors fall back to the original query.)
//   - synth knobs: "synthesize", "synth_mode", "synth_adaptive",
//     "synth_prompt"
//
// Synth requests (params["synthesize"]=true) are auto-routed to the
// async surface — sync 422s. The SDK posts to /v1/discovery/search/async
// and polls /v1/discovery/jobs/{id} until completion. The async
// lifecycle for synth is queued → running → serp_ready → completed |
// failed; the SDK polls through serp_ready transparently. For
// caller-driven polling (e.g. progressive UI rendering), use
// DiscoverySearchAsync + GetDiscoveryJob and watch for
// DiscoveryStatusSerpReady.
//
// Example — sync, no synth:
//
//	resp, err := crawler.DiscoverySearch(map[string]interface{}{
//	    "query":   "best AI code review tools 2026",
//	    "country": "us",
//	})
//
// Example — multi-backend fan-out + merge:
//
//	resp, err := crawler.DiscoverySearch(map[string]interface{}{
//	    "query":    "...",
//	    "country":  "us",
//	    "backends": []string{"google", "bing", "brave"},
//	})
//	// resp.Hits has the merged + RRF-ranked union (~25-35 hits)
//
// Example — synth, SDK polls transparently:
//
//	resp, err := crawler.DiscoverySearch(map[string]interface{}{
//	    "query":       "what is warrior's next game?",
//	    "country":     "us",
//	    "backends":    []string{"google", "bing", "brave"},
//	    "synthesize":  true,
//	    "synth_mode":  "auto",
//	})
//	fmt.Println(resp.SynthesizedAnswer.Text)
//
// Example — LLM query rewrite per backend:
//
//	resp, err := crawler.DiscoverySearch(map[string]interface{}{
//	    "query":         "best nurseries in Toronto for my 2 year old",
//	    "country":       "ca",
//	    "enhance_query": true,
//	    "backends":      []string{"google", "bing"},
//	})
//	fmt.Println(*resp.OriginalQuery)
//	for backend, rewritten := range resp.RewrittenQueries {
//	    fmt.Printf("  %-8s %s\n", backend, rewritten)
//	}
func (c *AsyncWebCrawler) DiscoverySearch(params map[string]interface{}) (*SearchResponse, error) {
	body := filterDiscoveryParams(params)

	// Synth-aware routing: post to /async + poll, or sync /v1/discovery/search.
	if wantsSynth, _ := body["synthesize"].(bool); wantsSynth {
		handle, err := c.DiscoverySearchAsync(body)
		if err != nil {
			return nil, err
		}
		status, err := c.waitForDiscoveryJob(
			handle.JobID, 1500*time.Millisecond, 5*time.Minute,
		)
		if err != nil {
			return nil, err
		}
		if status.Status == "failed" {
			msg := "unknown"
			if status.Error != nil {
				msg = *status.Error
			}
			return nil, fmt.Errorf("discovery search async job failed: %s", msg)
		}
		if status.Result == nil {
			return nil, fmt.Errorf("discovery search async job completed but result missing")
		}
		raw, err := json.Marshal(status.Result)
		if err != nil {
			return nil, fmt.Errorf("marshal async result: %w", err)
		}
		var sr SearchResponse
		if err := json.Unmarshal(raw, &sr); err != nil {
			return nil, fmt.Errorf("decode SearchResponse: %w", err)
		}
		return &sr, nil
	}

	// Sync path.
	data, err := c.http.Post("/v1/discovery/search", body, 60*time.Second)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal discovery response: %w", err)
	}
	var sr SearchResponse
	if err := json.Unmarshal(raw, &sr); err != nil {
		return nil, fmt.Errorf("decode SearchResponse: %w", err)
	}
	return &sr, nil
}

// DiscoverySearchAsync kicks off an async search job and returns a
// handle. Use it with GetDiscoveryJob when you want to drive polling
// yourself; for the common "fire and wait" case, just call DiscoverySearch
// with synthesize=true (the SDK handles the async + poll for you).
func (c *AsyncWebCrawler) DiscoverySearchAsync(params map[string]interface{}) (*DiscoveryJobHandle, error) {
	body := filterDiscoveryParams(params)
	data, err := c.http.Post("/v1/discovery/search/async", body, 60*time.Second)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal async start response: %w", err)
	}
	var handle DiscoveryJobHandle
	if err := json.Unmarshal(raw, &handle); err != nil {
		return nil, fmt.Errorf("decode DiscoveryJobHandle: %w", err)
	}
	return &handle, nil
}

// GetDiscoveryJob polls a discovery async job by id.
// status.Result is populated when status.Status == "completed".
func (c *AsyncWebCrawler) GetDiscoveryJob(jobID string) (*DiscoveryJobStatus, error) {
	data, err := c.http.Get("/v1/discovery/jobs/"+jobID, nil)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal job status: %w", err)
	}
	var status DiscoveryJobStatus
	if err := json.Unmarshal(raw, &status); err != nil {
		return nil, fmt.Errorf("decode DiscoveryJobStatus: %w", err)
	}
	return &status, nil
}

// waitForDiscoveryJob polls until the job hits a terminal status or the
// deadline elapses. Mild backoff (×1.15 per poll, capped at 4s) so a
// stalled job doesn't hammer the proxy.
func (c *AsyncWebCrawler) waitForDiscoveryJob(
	jobID string, interval time.Duration, timeout time.Duration,
) (*DiscoveryJobStatus, error) {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf(
				"discovery job %s did not complete within %s", jobID, timeout,
			)
		}
		time.Sleep(interval)
		status, err := c.GetDiscoveryJob(jobID)
		if err != nil {
			return nil, err
		}
		if status.Status == "completed" || status.Status == "failed" {
			return status, nil
		}
		// Backoff
		next := time.Duration(float64(interval) * 1.15)
		if next > 4*time.Second {
			next = 4 * time.Second
		}
		interval = next
	}
}

// ListDiscoveryServices fetches the GET /v1/discovery registry.
// Use this to feature-detect new verticals without an SDK update.
func (c *AsyncWebCrawler) ListDiscoveryServices() ([]DiscoveryService, error) {
	data, err := c.http.Get("/v1/discovery", nil)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal discovery registry: %w", err)
	}
	var wire struct {
		Services []DiscoveryService `json:"services"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		return nil, fmt.Errorf("decode discovery registry: %w", err)
	}
	return wire.Services, nil
}

// filterDiscoveryParams drops nil + empty-string optionals so the cache
// key matches the dashboard playground exactly. Wire parity avoids
// surprise misses between surfaces hitting the same params.
func filterDiscoveryParams(params map[string]interface{}) map[string]interface{} {
	body := map[string]interface{}{}
	for k, v := range params {
		if v == nil {
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			continue
		}
		body[k] = v
	}
	return body
}
