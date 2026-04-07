package crawl4ai

// ProxyConfig represents proxy configuration for crawl requests.
type ProxyConfig struct {
	Mode          string `json:"mode"`
	Country       string `json:"country,omitempty"`
	StickySession bool   `json:"sticky_session,omitempty"`
	UseProxy      bool   `json:"use_proxy,omitempty"`
	SkipDirect    bool   `json:"skip_direct,omitempty"`
}

// JobProgress represents async job progress.
type JobProgress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// Pending returns the number of pending items.
func (p *JobProgress) Pending() int {
	return p.Total - p.Completed - p.Failed
}

// Percent returns the completion percentage.
func (p *JobProgress) Percent() float64 {
	if p.Total == 0 {
		return 0
	}
	return float64(p.Completed+p.Failed) / float64(p.Total) * 100
}

// CrawlJob represents an async crawl job.
type CrawlJob struct {
	JobID           string         `json:"job_id"`
	Status          string         `json:"status"`
	Progress        JobProgress    `json:"progress"`
	URLsCount       int            `json:"urls_count"`
	CreatedAt       string         `json:"created_at"`
	StartedAt       string         `json:"started_at,omitempty"`
	CompletedAt     string         `json:"completed_at,omitempty"`
	Results         []*CrawlResult `json:"results,omitempty"`
	Error           string         `json:"error,omitempty"`
	ResultSizeBytes int            `json:"result_size_bytes,omitempty"`
	DownloadURL     string         `json:"download_url,omitempty"`
	// Usage contains resource usage metrics (completed jobs only)
	Usage *Usage `json:"usage,omitempty"`
}

// ID returns the job ID (backward compatibility alias for JobID).
// Deprecated: Use JobID instead.
func (j *CrawlJob) ID() string {
	return j.JobID
}

// IsComplete checks if job is in a terminal state.
func (j *CrawlJob) IsComplete() bool {
	switch j.Status {
	case "completed", "partial", "failed", "cancelled":
		return true
	}
	return false
}

// IsSuccessful checks if job completed successfully.
func (j *CrawlJob) IsSuccessful() bool {
	return j.Status == "completed"
}

// CrawlJobFromMap creates a CrawlJob from API response map.
func CrawlJobFromMap(data map[string]interface{}) *CrawlJob {
	job := &CrawlJob{}

	if v, ok := data["job_id"].(string); ok {
		job.JobID = v
	}
	if v, ok := data["status"].(string); ok {
		job.Status = v
	}
	if v, ok := data["urls_count"].(float64); ok {
		job.URLsCount = int(v)
	} else if v, ok := data["url_count"].(float64); ok {
		job.URLsCount = int(v)
	}
	if v, ok := data["created_at"].(string); ok {
		job.CreatedAt = v
	}
	if v, ok := data["started_at"].(string); ok {
		job.StartedAt = v
	}
	if v, ok := data["completed_at"].(string); ok {
		job.CompletedAt = v
	}
	if v, ok := data["error"].(string); ok {
		job.Error = v
	}
	if v, ok := data["result_size_bytes"].(float64); ok {
		job.ResultSizeBytes = int(v)
	}

	if progress, ok := data["progress"].(map[string]interface{}); ok {
		if v, ok := progress["total"].(float64); ok {
			job.Progress.Total = int(v)
		}
		if v, ok := progress["completed"].(float64); ok {
			job.Progress.Completed = int(v)
		}
		if v, ok := progress["failed"].(float64); ok {
			job.Progress.Failed = int(v)
		}
	}

	// Convert results to CrawlResult objects
	if results, ok := data["results"].([]interface{}); ok {
		job.Results = make([]*CrawlResult, 0, len(results))
		for _, r := range results {
			if m, ok := r.(map[string]interface{}); ok {
				result := CrawlResultFromMap(m)
				// Set job_id on each result for use with DownloadURL()
				result.ID = job.JobID
				job.Results = append(job.Results, result)
			}
		}
	}

	// Parse usage if present
	if usage, ok := data["usage"].(map[string]interface{}); ok {
		job.Usage = UsageFromMap(usage)
	}

	return job
}

// MarkdownResult represents markdown extraction result.
type MarkdownResult struct {
	RawMarkdown           string `json:"raw_markdown,omitempty"`
	MarkdownWithCitations string `json:"markdown_with_citations,omitempty"`
	ReferencesMarkdown    string `json:"references_markdown,omitempty"`
	FitMarkdown           string `json:"fit_markdown,omitempty"`
}

// CrawlResult represents a single URL crawl result.
type CrawlResult struct {
	URL              string                 `json:"url"`
	Success          bool                   `json:"success"`
	HTML             string                 `json:"html,omitempty"`
	CleanedHTML      string                 `json:"cleaned_html,omitempty"`
	FitHTML          string                 `json:"fit_html,omitempty"`
	Markdown         *MarkdownResult        `json:"markdown,omitempty"`
	Media            map[string]interface{} `json:"media,omitempty"`
	Links            map[string]interface{} `json:"links,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Screenshot       string                 `json:"screenshot,omitempty"`
	PDF              string                 `json:"pdf,omitempty"`
	ExtractedContent string                 `json:"extracted_content,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	StatusCode       int                    `json:"status_code,omitempty"`
	DurationMs       int                    `json:"duration_ms,omitempty"`
	Tables           []interface{}          `json:"tables,omitempty"`
	RedirectedURL    string                 `json:"redirected_url,omitempty"`
	CrawlStrategy    string                 `json:"crawl_strategy,omitempty"`
	// DownloadedFiles contains presigned S3 URLs for file downloads (CSV, PDF, XLSX, etc.)
	DownloadedFiles []string `json:"downloaded_files,omitempty"`
	// ID is the job ID for async results (use with DownloadURL())
	ID string `json:"id,omitempty"`
	// Usage contains resource usage metrics
	Usage *Usage `json:"usage,omitempty"`
}

// CrawlResultFromMap creates a CrawlResult from API response map.
func CrawlResultFromMap(data map[string]interface{}) *CrawlResult {
	result := &CrawlResult{}

	if v, ok := data["url"].(string); ok {
		result.URL = v
	}
	if v, ok := data["success"].(bool); ok {
		result.Success = v
	}
	if v, ok := data["html"].(string); ok {
		result.HTML = v
	}
	if v, ok := data["cleaned_html"].(string); ok {
		result.CleanedHTML = v
	}
	if v, ok := data["fit_html"].(string); ok {
		result.FitHTML = v
	}
	if v, ok := data["screenshot"].(string); ok {
		result.Screenshot = v
	}
	if v, ok := data["pdf"].(string); ok {
		result.PDF = v
	}
	if v, ok := data["extracted_content"].(string); ok {
		result.ExtractedContent = v
	}
	if v, ok := data["error_message"].(string); ok {
		result.ErrorMessage = v
	}
	if v, ok := data["status_code"].(float64); ok {
		result.StatusCode = int(v)
	}
	if v, ok := data["duration_ms"].(float64); ok {
		result.DurationMs = int(v)
	}
	if v, ok := data["redirected_url"].(string); ok {
		result.RedirectedURL = v
	}
	if v, ok := data["crawl_strategy"].(string); ok {
		result.CrawlStrategy = v
	}
	if v, ok := data["media"].(map[string]interface{}); ok {
		result.Media = v
	}
	if v, ok := data["links"].(map[string]interface{}); ok {
		result.Links = v
	}
	if v, ok := data["metadata"].(map[string]interface{}); ok {
		result.Metadata = v
	}
	if v, ok := data["tables"].([]interface{}); ok {
		result.Tables = v
	}

	// Parse downloaded_files (presigned S3 URLs for file downloads)
	if files, ok := data["downloaded_files"].([]interface{}); ok {
		result.DownloadedFiles = make([]string, 0, len(files))
		for _, f := range files {
			if s, ok := f.(string); ok {
				result.DownloadedFiles = append(result.DownloadedFiles, s)
			}
		}
	}

	// Handle both string (async results) and object (sync results) formats
	if mdStr, ok := data["markdown"].(string); ok {
		result.Markdown = &MarkdownResult{RawMarkdown: mdStr}
	} else if md, ok := data["markdown"].(map[string]interface{}); ok {
		result.Markdown = &MarkdownResult{}
		if v, ok := md["raw_markdown"].(string); ok {
			result.Markdown.RawMarkdown = v
		}
		if v, ok := md["markdown_with_citations"].(string); ok {
			result.Markdown.MarkdownWithCitations = v
		}
		if v, ok := md["references_markdown"].(string); ok {
			result.Markdown.ReferencesMarkdown = v
		}
		if v, ok := md["fit_markdown"].(string); ok {
			result.Markdown.FitMarkdown = v
		}
	}

	if usage, ok := data["usage"].(map[string]interface{}); ok {
		result.Usage = UsageFromMap(usage)
	}

	return result
}

// DomainScanURLInfo represents a URL discovered by domain scan.
type DomainScanURLInfo struct {
	URL            string                 `json:"url"`
	Host           string                 `json:"host"`
	Status         string                 `json:"status"`
	RelevanceScore *float64               `json:"relevance_score,omitempty"`
	HeadData       map[string]interface{} `json:"head_data,omitempty"`
}

// ScanResult represents a domain scan response (/v1/scan).
type ScanResult struct {
	Success    bool                `json:"success"`
	Domain     string              `json:"domain"`
	TotalUrls  int                 `json:"total_urls"`
	HostsFound int                 `json:"hosts_found"`
	Mode       string              `json:"mode"`
	Urls       []DomainScanURLInfo `json:"urls"`
	DurationMs int                 `json:"duration_ms"`
	Error      string              `json:"error,omitempty"`
}

// ScanOptions configures a domain scan request.
type ScanOptions struct {
	Mode              string   `json:"mode,omitempty"` // "default" or "deep"
	MaxUrls           int      `json:"max_urls,omitempty"`
	IncludeSubdomains *bool    `json:"include_subdomains,omitempty"`
	ExtractHead       *bool    `json:"extract_head,omitempty"`
	Soft404Detection  *bool    `json:"soft_404_detection,omitempty"`
	Query             string   `json:"query,omitempty"`
	ScoreThreshold    *float64 `json:"score_threshold,omitempty"`
	Force             bool     `json:"force,omitempty"`
	ProbeThreshold    *int     `json:"probe_threshold,omitempty"`
}

// ScanResultFromMap creates a ScanResult from API response map.
func ScanResultFromMap(data map[string]interface{}) *ScanResult {
	result := &ScanResult{}

	if v, ok := data["success"].(bool); ok {
		result.Success = v
	}
	if v, ok := data["domain"].(string); ok {
		result.Domain = v
	}
	if v, ok := data["total_urls"].(float64); ok {
		result.TotalUrls = int(v)
	}
	if v, ok := data["hosts_found"].(float64); ok {
		result.HostsFound = int(v)
	}
	if v, ok := data["mode"].(string); ok {
		result.Mode = v
	}
	if v, ok := data["duration_ms"].(float64); ok {
		result.DurationMs = int(v)
	}
	if v, ok := data["error"].(string); ok {
		result.Error = v
	}
	if urls, ok := data["urls"].([]interface{}); ok {
		for _, u := range urls {
			if um, ok := u.(map[string]interface{}); ok {
				info := DomainScanURLInfo{Status: "valid"}
				if v, ok := um["url"].(string); ok {
					info.URL = v
				}
				if v, ok := um["host"].(string); ok {
					info.Host = v
				}
				if v, ok := um["status"].(string); ok {
					info.Status = v
				}
				if score, ok := um["relevance_score"].(float64); ok {
					info.RelevanceScore = &score
				}
				if hd, ok := um["head_data"].(map[string]interface{}); ok {
					info.HeadData = hd
				}
				result.Urls = append(result.Urls, info)
			}
		}
	}

	return result
}

// DeepCrawlResult represents a deep crawl response.
type DeepCrawlResult struct {
	JobID           string `json:"job_id"`
	Status          string `json:"status"`
	Strategy        string `json:"strategy"`
	DiscoveredCount int    `json:"discovered_count"`
	QueuedURLs      int    `json:"queued_urls"`
	CreatedAt       string `json:"created_at"`
	HTMLDownloadURL string `json:"html_download_url,omitempty"`
	CacheExpiresAt  string `json:"cache_expires_at,omitempty"`
	CrawlJobID      string `json:"crawl_job_id,omitempty"`
}

// IsComplete checks if deep crawl is complete.
func (d *DeepCrawlResult) IsComplete() bool {
	return d.Status == "completed" || d.Status == "failed" || d.Status == "cancelled"
}

// DeepCrawlResultFromMap creates a DeepCrawlResult from API response map.
func DeepCrawlResultFromMap(data map[string]interface{}) *DeepCrawlResult {
	result := &DeepCrawlResult{}

	if v, ok := data["job_id"].(string); ok {
		result.JobID = v
	}
	if v, ok := data["status"].(string); ok {
		result.Status = v
	}
	if v, ok := data["strategy"].(string); ok {
		result.Strategy = v
	}
	if v, ok := data["discovered_urls"].(float64); ok {
		result.DiscoveredCount = int(v)
	}
	if v, ok := data["queued_urls"].(float64); ok {
		result.QueuedURLs = int(v)
	}
	if v, ok := data["created_at"].(string); ok {
		result.CreatedAt = v
	}
	if v, ok := data["html_download_url"].(string); ok {
		result.HTMLDownloadURL = v
	}
	if v, ok := data["cache_expires_at"].(string); ok {
		result.CacheExpiresAt = v
	}
	if v, ok := data["crawl_job_id"].(string); ok {
		result.CrawlJobID = v
	}

	return result
}

// StorageUsage represents storage quota usage (from /storage endpoint).
type StorageUsage struct {
	UsedMB      float64 `json:"used_mb"`
	MaxMB       float64 `json:"max_mb"`
	RemainingMB float64 `json:"remaining_mb"`
	PercentUsed float64 `json:"percent_used"`
}

// CrawlUsageMetrics represents crawl usage metrics in API responses.
type CrawlUsageMetrics struct {
	CreditsUsed      float64 `json:"credits_used"`
	CreditsRemaining float64 `json:"credits_remaining"`
	DurationMs       int     `json:"duration_ms"`
	Cached           bool    `json:"cached"` // bool for single crawl, may be int for batch
	URLsTotal        int     `json:"urls_total,omitempty"`
	URLsSucceeded    int     `json:"urls_succeeded,omitempty"`
	URLsFailed       int     `json:"urls_failed,omitempty"`
}

// LLMUsageMetrics represents LLM usage metrics in API responses.
type LLMUsageMetrics struct {
	TokensUsed      int    `json:"tokens_used"`
	TokensRemaining int    `json:"tokens_remaining"`
	Model           string `json:"model,omitempty"`
}

// StorageUsageMetrics represents storage metrics in API responses (async jobs only).
type StorageUsageMetrics struct {
	BytesUsed      int `json:"bytes_used"`
	BytesRemaining int `json:"bytes_remaining"`
}

// Usage represents unified usage metrics returned in API responses.
type Usage struct {
	Crawl   *CrawlUsageMetrics   `json:"crawl"`
	LLM     *LLMUsageMetrics     `json:"llm,omitempty"`
	Storage *StorageUsageMetrics `json:"storage,omitempty"`
}

// UsageFromMap creates a Usage from API response map.
func UsageFromMap(data map[string]interface{}) *Usage {
	usage := &Usage{}

	if crawl, ok := data["crawl"].(map[string]interface{}); ok {
		usage.Crawl = &CrawlUsageMetrics{}
		if v, ok := crawl["credits_used"].(float64); ok {
			usage.Crawl.CreditsUsed = v
		}
		if v, ok := crawl["credits_remaining"].(float64); ok {
			usage.Crawl.CreditsRemaining = v
		}
		if v, ok := crawl["duration_ms"].(float64); ok {
			usage.Crawl.DurationMs = int(v)
		}
		if v, ok := crawl["cached"].(bool); ok {
			usage.Crawl.Cached = v
		}
		if v, ok := crawl["urls_total"].(float64); ok {
			usage.Crawl.URLsTotal = int(v)
		}
		if v, ok := crawl["urls_succeeded"].(float64); ok {
			usage.Crawl.URLsSucceeded = int(v)
		}
		if v, ok := crawl["urls_failed"].(float64); ok {
			usage.Crawl.URLsFailed = int(v)
		}
	}

	if llm, ok := data["llm"].(map[string]interface{}); ok {
		usage.LLM = &LLMUsageMetrics{}
		if v, ok := llm["tokens_used"].(float64); ok {
			usage.LLM.TokensUsed = int(v)
		}
		if v, ok := llm["tokens_remaining"].(float64); ok {
			usage.LLM.TokensRemaining = int(v)
		}
		if v, ok := llm["model"].(string); ok {
			usage.LLM.Model = v
		}
	}

	if storage, ok := data["storage"].(map[string]interface{}); ok {
		usage.Storage = &StorageUsageMetrics{}
		if v, ok := storage["bytes_used"].(float64); ok {
			usage.Storage.BytesUsed = int(v)
		}
		if v, ok := storage["bytes_remaining"].(float64); ok {
			usage.Storage.BytesRemaining = int(v)
		}
	}

	return usage
}

// StorageUsageFromMap creates a StorageUsage from API response map.
func StorageUsageFromMap(data map[string]interface{}) *StorageUsage {
	usage := &StorageUsage{}

	if v, ok := data["used_mb"].(float64); ok {
		usage.UsedMB = v
	}
	if v, ok := data["max_mb"].(float64); ok {
		usage.MaxMB = v
	}
	if v, ok := data["remaining_mb"].(float64); ok {
		usage.RemainingMB = v
	}
	if v, ok := data["percent_used"].(float64); ok {
		usage.PercentUsed = v
	}

	return usage
}

// ContextResult represents a context API response.
type ContextResult struct {
	JobID       string `json:"job_id"`
	Status      string `json:"status"`
	Query       string `json:"query"`
	DownloadURL string `json:"download_url"`
	URLsCrawled int    `json:"urls_crawled"`
	SizeBytes   int    `json:"size_bytes"`
	DurationMs  int    `json:"duration_ms"`
	Cached      bool   `json:"cached"`
}

// ContextResultFromMap creates a ContextResult from API response map.
func ContextResultFromMap(data map[string]interface{}) *ContextResult {
	result := &ContextResult{}

	if v, ok := data["job_id"].(string); ok {
		result.JobID = v
	}
	if v, ok := data["status"].(string); ok {
		result.Status = v
	}
	if v, ok := data["query"].(string); ok {
		result.Query = v
	}
	if v, ok := data["download_url"].(string); ok {
		result.DownloadURL = v
	}
	if v, ok := data["urls_crawled"].(float64); ok {
		result.URLsCrawled = int(v)
	}
	if v, ok := data["storage_size_bytes"].(float64); ok {
		result.SizeBytes = int(v)
	}
	if v, ok := data["duration_ms"].(float64); ok {
		result.DurationMs = int(v)
	}
	if v, ok := data["cached"].(bool); ok {
		result.Cached = v
	}

	return result
}

// GeneratedSchema represents a generated extraction schema.
type GeneratedSchema struct {
	Success bool                   `json:"success"`
	Schema  map[string]interface{} `json:"schema,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// GeneratedSchemaFromMap creates a GeneratedSchema from API response map.
func GeneratedSchemaFromMap(data map[string]interface{}) *GeneratedSchema {
	result := &GeneratedSchema{}

	if v, ok := data["success"].(bool); ok {
		result.Success = v
	}
	if v, ok := data["schema"].(map[string]interface{}); ok {
		result.Schema = v
	}
	if v, ok := data["error_message"].(string); ok {
		result.Error = v
	}

	return result
}

// =============================================================================
// Wrapper API Models
// =============================================================================

// WrapperUsage represents credit usage from wrapper endpoints.
type WrapperUsage struct {
	CreditsUsed      float64 `json:"credits_used"`
	CreditsRemaining float64 `json:"credits_remaining"`
}

// MarkdownResponse represents the response from POST /v1/markdown.
type MarkdownResponse struct {
	Success      bool                     `json:"success"`
	URL          string                   `json:"url"`
	Markdown     string                   `json:"markdown,omitempty"`
	FitMarkdown  string                   `json:"fit_markdown,omitempty"`
	FitHTML      string                   `json:"fit_html,omitempty"`
	Links        map[string]interface{}   `json:"links,omitempty"`
	Media        map[string]interface{}   `json:"media,omitempty"`
	Metadata     map[string]interface{}   `json:"metadata,omitempty"`
	Tables       []map[string]interface{} `json:"tables,omitempty"`
	DurationMs   int                      `json:"duration_ms"`
	Usage        *WrapperUsage            `json:"usage,omitempty"`
	ErrorMessage string                   `json:"error_message,omitempty"`
}

// ScreenshotResponse represents the response from POST /v1/screenshot.
type ScreenshotResponse struct {
	Success      bool          `json:"success"`
	URL          string        `json:"url"`
	Screenshot   string        `json:"screenshot,omitempty"`
	PDF          string        `json:"pdf,omitempty"`
	DurationMs   int           `json:"duration_ms"`
	Usage        *WrapperUsage `json:"usage,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// ExtractResponse represents the response from POST /v1/extract.
type ExtractResponse struct {
	Success      bool                     `json:"success"`
	URL          string                   `json:"url,omitempty"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	MethodUsed   string                   `json:"method_used,omitempty"`
	SchemaUsed   map[string]interface{}   `json:"schema_used,omitempty"`
	QueryUsed    string                   `json:"query_used,omitempty"`
	DurationMs   int                      `json:"duration_ms"`
	ErrorMessage string                   `json:"error_message,omitempty"`
}

// MapUrlInfo represents a discovered URL from POST /v1/map.
type MapUrlInfo struct {
	URL            string                 `json:"url"`
	Host           string                 `json:"host"`
	Status         string                 `json:"status"`
	RelevanceScore *float64               `json:"relevance_score,omitempty"`
	HeadData       map[string]interface{} `json:"head_data,omitempty"`
}

// MapResponse represents the response from POST /v1/map.
type MapResponse struct {
	Success      bool         `json:"success"`
	Domain       string       `json:"domain"`
	TotalUrls    int          `json:"total_urls"`
	HostsFound   int          `json:"hosts_found"`
	Mode         string       `json:"mode"`
	URLs         []MapUrlInfo `json:"urls"`
	DurationMs   int          `json:"duration_ms"`
	ErrorMessage string       `json:"error_message,omitempty"`
}

// SiteCrawlResponse represents the response from POST /v1/crawl/site.
type SiteCrawlResponse struct {
	JobID          string `json:"job_id"`
	Status         string `json:"status"`
	Strategy       string `json:"strategy"`
	DiscoveredURLs int    `json:"discovered_urls"`
	QueuedURLs     int    `json:"queued_urls"`
	CreatedAt      string `json:"created_at"`
}

// WrapperJobProgress represents progress of a wrapper async job.
type WrapperJobProgress struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// Percent returns the completion percentage.
func (p *WrapperJobProgress) Percent() int {
	if p.Total == 0 {
		return 0
	}
	return int(float64(p.Completed+p.Failed) / float64(p.Total) * 100)
}

// WrapperJob represents job status for wrapper async endpoints.
type WrapperJob struct {
	JobID           string              `json:"job_id"`
	Status          string              `json:"status"`
	Progress        *WrapperJobProgress `json:"progress,omitempty"`
	ProgressPercent int                 `json:"progress_percent"`
	URLsCount       int                 `json:"urls_count"`
	Error           string              `json:"error,omitempty"`
	CreatedAt       string              `json:"created_at,omitempty"`
	StartedAt       string              `json:"started_at,omitempty"`
	CompletedAt     string              `json:"completed_at,omitempty"`
}

// IsComplete returns true if the job is in a terminal state.
func (j *WrapperJob) IsComplete() bool {
	switch j.Status {
	case "completed", "partial", "failed", "cancelled":
		return true
	}
	return false
}

// Wrapper option structs

// MarkdownOptions configures the markdown method.
type MarkdownOptions struct {
	Strategy      string                 `json:"strategy,omitempty"`
	Fit           *bool                  `json:"fit,omitempty"`
	Include       []string               `json:"include,omitempty"`
	CrawlerConfig map[string]interface{} `json:"crawler_config,omitempty"`
	BrowserConfig map[string]interface{} `json:"browser_config,omitempty"`
	Proxy         map[string]interface{} `json:"proxy,omitempty"`
	BypassCache   bool                   `json:"bypass_cache,omitempty"`
}

// ScreenshotOptions configures the screenshot method.
type ScreenshotOptions struct {
	FullPage      *bool                  `json:"full_page,omitempty"`
	PDF           bool                   `json:"pdf,omitempty"`
	WaitFor       string                 `json:"wait_for,omitempty"`
	CrawlerConfig map[string]interface{} `json:"crawler_config,omitempty"`
	BrowserConfig map[string]interface{} `json:"browser_config,omitempty"`
	Proxy         map[string]interface{} `json:"proxy,omitempty"`
	BypassCache   bool                   `json:"bypass_cache,omitempty"`
}

// ExtractOptions configures the extract method.
type ExtractOptions struct {
	Query         string                 `json:"query,omitempty"`
	JSONExample   map[string]interface{} `json:"json_example,omitempty"`
	Schema        map[string]interface{} `json:"schema,omitempty"`
	Method        string                 `json:"method,omitempty"`
	Strategy      string                 `json:"strategy,omitempty"`
	CrawlerConfig map[string]interface{} `json:"crawler_config,omitempty"`
	BrowserConfig map[string]interface{} `json:"browser_config,omitempty"`
	LLMConfig     map[string]interface{} `json:"llm_config,omitempty"`
	Proxy         map[string]interface{} `json:"proxy,omitempty"`
	BypassCache   bool                   `json:"bypass_cache,omitempty"`
}

// MapOptions configures the map method.
type MapOptions struct {
	Mode              string                 `json:"mode,omitempty"`
	MaxURLs           *int                   `json:"max_urls,omitempty"`
	IncludeSubdomains bool                   `json:"include_subdomains,omitempty"`
	ExtractHead       *bool                  `json:"extract_head,omitempty"`
	Query             string                 `json:"query,omitempty"`
	ScoreThreshold    *float64               `json:"score_threshold,omitempty"`
	Force             bool                   `json:"force,omitempty"`
	Proxy             map[string]interface{} `json:"proxy,omitempty"`
}

// SiteCrawlOptions configures the crawl_site method.
type SiteCrawlOptions struct {
	MaxPages      int                    `json:"max_pages,omitempty"`
	Discovery     string                 `json:"discovery,omitempty"`
	Strategy      string                 `json:"strategy,omitempty"`
	Fit           *bool                  `json:"fit,omitempty"`
	Include       []string               `json:"include,omitempty"`
	Pattern       string                 `json:"pattern,omitempty"`
	MaxDepth      *int                   `json:"max_depth,omitempty"`
	CrawlerConfig map[string]interface{} `json:"crawler_config,omitempty"`
	BrowserConfig map[string]interface{} `json:"browser_config,omitempty"`
	Proxy         map[string]interface{} `json:"proxy,omitempty"`
	WebhookURL    string                 `json:"webhook_url,omitempty"`
	Priority      int                    `json:"priority,omitempty"`
}
