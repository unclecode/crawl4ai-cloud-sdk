package crawl4ai

import "time"

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

// SiteScanConfig is the unified scan configuration for AI-assisted URL
// discovery. Used by /v1/scan and /v1/crawl/site. When `criteria` is set on
// the parent request, the AI config generator fills unset fields here;
// explicit fields always win over LLM output.
type SiteScanConfig struct {
	Mode              string                 `json:"mode,omitempty"` // "auto" | "map" | "deep"
	Patterns          []string               `json:"patterns,omitempty"`
	Filters           map[string]interface{} `json:"filters,omitempty"`
	Scorers           map[string]interface{} `json:"scorers,omitempty"`
	Query             string                 `json:"query,omitempty"`
	ScoreThreshold    *float64               `json:"score_threshold,omitempty"`
	IncludeSubdomains bool                   `json:"include_subdomains,omitempty"`
	MaxDepth          *int                   `json:"max_depth,omitempty"`
}

// ToMap converts SiteScanConfig to a request body dict, omitting zero values.
func (s *SiteScanConfig) ToMap() map[string]interface{} {
	if s == nil {
		return nil
	}
	d := map[string]interface{}{}
	if s.Mode != "" {
		d["mode"] = s.Mode
	} else {
		d["mode"] = "auto"
	}
	if s.Patterns != nil {
		d["patterns"] = s.Patterns
	}
	if s.Filters != nil {
		d["filters"] = s.Filters
	}
	if s.Scorers != nil {
		d["scorers"] = s.Scorers
	}
	if s.Query != "" {
		d["query"] = s.Query
	}
	if s.ScoreThreshold != nil {
		d["score_threshold"] = *s.ScoreThreshold
	}
	if s.IncludeSubdomains {
		d["include_subdomains"] = true
	}
	if s.MaxDepth != nil {
		d["max_depth"] = *s.MaxDepth
	}
	return d
}

// SiteExtractConfig is the structured extraction configuration for
// /v1/crawl/site. Mirrors /v1/extract's shape. When set without a pre-built
// schema, the backend fetches `SampleURL` (defaults to the crawl's start URL),
// generates a schema via LLM, and applies it to every discovered URL.
type SiteExtractConfig struct {
	Query       string                 `json:"query,omitempty"`
	JSONExample map[string]interface{} `json:"json_example,omitempty"`
	Method      string                 `json:"method,omitempty"` // "auto" | "llm" | "schema"
	Schema      map[string]interface{} `json:"schema,omitempty"`
	SampleURL   string                 `json:"sample_url,omitempty"`
	URLPattern  string                 `json:"url_pattern,omitempty"`
}

// ToMap converts SiteExtractConfig to a request body dict.
func (e *SiteExtractConfig) ToMap() map[string]interface{} {
	if e == nil {
		return nil
	}
	d := map[string]interface{}{}
	if e.Method != "" {
		d["method"] = e.Method
	} else {
		d["method"] = "auto"
	}
	if e.Query != "" {
		d["query"] = e.Query
	}
	if e.JSONExample != nil {
		d["json_example"] = e.JSONExample
	}
	if e.Schema != nil {
		d["schema"] = e.Schema
	}
	if e.SampleURL != "" {
		d["sample_url"] = e.SampleURL
	}
	if e.URLPattern != "" {
		d["url_pattern"] = e.URLPattern
	}
	return d
}

// GeneratedConfig is the LLM-generated config echoed back by /v1/scan and
// /v1/crawl/site when `criteria` was set. Contains the scan config and
// (for /v1/crawl/site) the extract config, plus LLM reasoning and
// cache/fallback flags.
type GeneratedConfig struct {
	Scan      map[string]interface{} `json:"scan"`
	Reasoning string                 `json:"reasoning"`
	Extract   map[string]interface{} `json:"extract,omitempty"`
	Fallback  bool                   `json:"fallback"`
	Cached    bool                   `json:"cached"`
}

// ScanResult represents a domain scan response (/v1/scan).
//
// For map mode (sync): Urls is populated inline.
// For deep mode (async): JobID + Status are set; poll with GetScanJob().
// When `criteria` was supplied in the request, GeneratedConfig carries the
// LLM output and ModeUsed tells you which strategy ran.
type ScanResult struct {
	Success    bool                `json:"success"`
	Domain     string              `json:"domain"`
	TotalUrls  int                 `json:"total_urls"`
	HostsFound int                 `json:"hosts_found"`
	Mode       string              `json:"mode"`
	Urls       []DomainScanURLInfo `json:"urls"`
	DurationMs int                 `json:"duration_ms"`
	Error      string              `json:"error,omitempty"`
	// AI-assisted / async fields
	ModeUsed        string           `json:"mode_used,omitempty"` // "map" | "deep"
	JobID           string           `json:"job_id,omitempty"`
	Status          string           `json:"status,omitempty"`
	GeneratedConfig *GeneratedConfig `json:"generated_config,omitempty"`
	Message         string           `json:"message,omitempty"`
}

// IsAsync returns true when the response is for an async (deep) scan and the
// caller should poll via GetScanJob().
func (r *ScanResult) IsAsync() bool {
	return r.ModeUsed == "deep" && r.JobID != ""
}

// ScanOptions configures a domain scan request.
type ScanOptions struct {
	Mode              string   `json:"mode,omitempty"` // "default" or "deep" — DomainMapper source depth
	MaxUrls           int      `json:"max_urls,omitempty"`
	IncludeSubdomains *bool    `json:"include_subdomains,omitempty"`
	ExtractHead       *bool    `json:"extract_head,omitempty"`
	Soft404Detection  *bool    `json:"soft_404_detection,omitempty"`
	Query             string   `json:"query,omitempty"`
	ScoreThreshold    *float64 `json:"score_threshold,omitempty"`
	Force             bool     `json:"force,omitempty"`
	ProbeThreshold    *int     `json:"probe_threshold,omitempty"`

	// AI-assisted fields (new in 0.4.0)
	Criteria string          `json:"-"` // plain-English — triggers LLM config gen
	Scan     *SiteScanConfig `json:"-"` // explicit scan overrides

	// Async polling (only used when scan.Mode = "deep")
	Wait         bool          `json:"-"`
	PollInterval time.Duration `json:"-"`
	Timeout      time.Duration `json:"-"`
}

// ScanJobStatus is the polling response for GET /v1/scan/jobs/{job_id} — used
// with async deep scans. URLs are appended to Urls as they're discovered.
// Progress carries {completed, total} once the backend starts tracking.
type ScanJobStatus struct {
	JobID           string              `json:"job_id"`
	Status          string              `json:"status"`
	ModeUsed        string              `json:"mode_used"`
	Domain          string              `json:"domain,omitempty"`
	TotalUrls       int                 `json:"total_urls"`
	Urls            []DomainScanURLInfo `json:"urls,omitempty"`
	Progress        map[string]int      `json:"progress,omitempty"`
	GeneratedConfig *GeneratedConfig    `json:"generated_config,omitempty"`
	DurationMs      int                 `json:"duration_ms"`
	Error           string              `json:"error,omitempty"`
	CreatedAt       string              `json:"created_at,omitempty"`
	CompletedAt     string              `json:"completed_at,omitempty"`
}

// IsComplete returns true when the scan job has finished (terminal state).
func (j *ScanJobStatus) IsComplete() bool {
	switch j.Status {
	case "completed", "partial", "failed", "cancelled":
		return true
	}
	return false
}

// IsSuccessful returns true when the scan job completed with usable results.
func (j *ScanJobStatus) IsSuccessful() bool {
	return j.Status == "completed" || j.Status == "partial"
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
				result.Urls = append(result.Urls, domainScanURLInfoFromMap(um))
			}
		}
	}

	// AI-assisted / async fields
	if v, ok := data["mode_used"].(string); ok {
		result.ModeUsed = v
	}
	if v, ok := data["job_id"].(string); ok {
		result.JobID = v
	}
	if v, ok := data["status"].(string); ok {
		result.Status = v
	}
	if v, ok := data["message"].(string); ok {
		result.Message = v
	}
	if gc, ok := data["generated_config"].(map[string]interface{}); ok {
		result.GeneratedConfig = generatedConfigFromMap(gc)
	}

	return result
}

// domainScanURLInfoFromMap is a shared helper used by ScanResult + ScanJobStatus.
func domainScanURLInfoFromMap(um map[string]interface{}) DomainScanURLInfo {
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
	return info
}

// generatedConfigFromMap decodes the generated_config block.
func generatedConfigFromMap(data map[string]interface{}) *GeneratedConfig {
	gc := &GeneratedConfig{}
	if v, ok := data["scan"].(map[string]interface{}); ok {
		gc.Scan = v
	}
	if v, ok := data["reasoning"].(string); ok {
		gc.Reasoning = v
	}
	if v, ok := data["extract"].(map[string]interface{}); ok {
		gc.Extract = v
	}
	if v, ok := data["fallback"].(bool); ok {
		gc.Fallback = v
	}
	if v, ok := data["cached"].(bool); ok {
		gc.Cached = v
	}
	return gc
}

// ScanJobStatusFromMap creates a ScanJobStatus from API response map.
func ScanJobStatusFromMap(data map[string]interface{}) *ScanJobStatus {
	js := &ScanJobStatus{ModeUsed: "deep"}
	if v, ok := data["job_id"].(string); ok {
		js.JobID = v
	}
	if v, ok := data["status"].(string); ok {
		js.Status = v
	}
	if v, ok := data["mode_used"].(string); ok {
		js.ModeUsed = v
	}
	if v, ok := data["domain"].(string); ok {
		js.Domain = v
	}
	if v, ok := data["total_urls"].(float64); ok {
		js.TotalUrls = int(v)
	}
	if v, ok := data["duration_ms"].(float64); ok {
		js.DurationMs = int(v)
	}
	if v, ok := data["error"].(string); ok {
		js.Error = v
	}
	if v, ok := data["created_at"].(string); ok {
		js.CreatedAt = v
	}
	if v, ok := data["completed_at"].(string); ok {
		js.CompletedAt = v
	}
	if urls, ok := data["urls"].([]interface{}); ok {
		for _, u := range urls {
			if um, ok := u.(map[string]interface{}); ok {
				js.Urls = append(js.Urls, domainScanURLInfoFromMap(um))
			}
		}
	}
	if p, ok := data["progress"].(map[string]interface{}); ok {
		js.Progress = map[string]int{}
		for k, val := range p {
			if f, ok := val.(float64); ok {
				js.Progress[k] = int(f)
			}
		}
	}
	if gc, ok := data["generated_config"].(map[string]interface{}); ok {
		js.GeneratedConfig = generatedConfigFromMap(gc)
	}
	return js
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
// Enrich API Models
// =============================================================================

// =============================================================================
// Enrich v2 — multi-phase API
// =============================================================================
//
// Phase machine:
//
//   queued → planning → plan_ready → resolving_urls → urls_ready
//          → extracting → merging → completed | partial | failed | cancelled
//
// Defaults AutoConfirmPlan=true and AutoConfirmUrls=true make jobs run
// straight through. Set either to false to pause for review and resume via
// crawler.ResumeEnrichJob(...).

// EnrichStatus is the union of phase + terminal statuses for a job.
type EnrichStatus = string

// Status constants — use these instead of bare strings.
const (
	EnrichStatusQueued        EnrichStatus = "queued"
	EnrichStatusPlanning      EnrichStatus = "planning"
	EnrichStatusPlanReady     EnrichStatus = "plan_ready"
	EnrichStatusResolvingURLs EnrichStatus = "resolving_urls"
	EnrichStatusURLsReady     EnrichStatus = "urls_ready"
	EnrichStatusExtracting    EnrichStatus = "extracting"
	EnrichStatusMerging       EnrichStatus = "merging"
	EnrichStatusCompleted     EnrichStatus = "completed"
	EnrichStatusPartial       EnrichStatus = "partial"
	EnrichStatusFailed        EnrichStatus = "failed"
	EnrichStatusCancelled     EnrichStatus = "cancelled"
)

// EnrichTerminalStatuses lists statuses where the job has stopped advancing.
var EnrichTerminalStatuses = []EnrichStatus{
	EnrichStatusCompleted, EnrichStatusPartial,
	EnrichStatusFailed, EnrichStatusCancelled,
}

// EnrichPausedStatuses lists statuses that require /continue to advance.
var EnrichPausedStatuses = []EnrichStatus{
	EnrichStatusPlanReady, EnrichStatusURLsReady,
}

// EnrichEntity is one row identifier (specific proper noun).
type EnrichEntity struct {
	Name      string `json:"name"`
	Title     string `json:"title,omitempty"`
	SourceURL string `json:"source_url,omitempty"`
}

// EnrichCriterion is a search-side filter used when finding URLs per entity.
type EnrichCriterion struct {
	Text string `json:"text"`
	Kind string `json:"kind,omitempty"` // "location" | "filter" | "other"
}

// EnrichFeature is one extraction column — a field pulled off each crawled page.
type EnrichFeature struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// EnrichConfig configures enrichment crawl behavior.
type EnrichConfig struct {
	MaxDepth          int  `json:"max_depth,omitempty"`
	MaxLinks          int  `json:"max_links,omitempty"`
	EnableSearch      bool `json:"enable_search,omitempty"`
	RetryCount        int  `json:"retry_count,omitempty"`
	KeysTopK          int  `json:"keys_top_k,omitempty"`
	CrossSourceVerify bool `json:"cross_source_verify,omitempty"`
}

// EnrichPlan is the LLM-expanded plan for a query.
type EnrichPlan struct {
	Entities         []EnrichEntity    `json:"entities"`
	Criteria         []EnrichCriterion `json:"criteria"`
	Features         []EnrichFeature   `json:"features"`
	AssistantMessage string            `json:"assistant_message,omitempty"`
	QueriesUsed      []string          `json:"queries_used,omitempty"`
}

// EnrichURLCandidate is one URL found for an entity by Serper grounding.
type EnrichURLCandidate struct {
	URL          string  `json:"url"`
	Rank         int     `json:"rank"`
	DomainTier   float64 `json:"domain_tier"`
	Title        string  `json:"title,omitempty"`
	QueryUsed    string  `json:"query_used,omitempty"`
	RequiresAuth bool    `json:"requires_auth,omitempty"`
}

// EnrichRow is one merged row in the enrichment table.
type EnrichRow struct {
	GroupID       string                            `json:"group_id"`
	InputKey      *string                           `json:"input_key,omitempty"`
	URL           *string                           `json:"url,omitempty"`
	Fields        map[string]interface{}            `json:"fields"`
	Sources       map[string]map[string]interface{} `json:"sources,omitempty"`
	Certainty     map[string]float64                `json:"certainty,omitempty"`
	Disputed      []string                          `json:"disputed,omitempty"`
	FragmentsUsed int                               `json:"fragments_used,omitempty"`
	Status        string                            `json:"status"` // "complete" | "partial" | "failed"
	Error         string                            `json:"error,omitempty"`
}

// EnrichPhaseData holds the per-phase payload — fields appear as their phase completes.
type EnrichPhaseData struct {
	Plan          *EnrichPlan                      `json:"plan,omitempty"`
	URLsPerEntity map[string][]EnrichURLCandidate  `json:"urls_per_entity,omitempty"`
	Fragments     []map[string]interface{}         `json:"fragments,omitempty"`
	Rows          []EnrichRow                      `json:"rows,omitempty"`
}

// EnrichProgress is URL- and group-level progress during extraction + merge.
type EnrichProgress struct {
	TotalURLs       int `json:"total_urls"`
	CompletedURLs   int `json:"completed_urls"`
	FailedURLs      int `json:"failed_urls"`
	TotalGroups     int `json:"total_groups"`
	CompletedGroups int `json:"completed_groups"`
}

// Percent returns 0–100 based on URL completion.
func (p *EnrichProgress) Percent() int {
	if p.TotalURLs == 0 {
		return 0
	}
	return int(float64(p.CompletedURLs+p.FailedURLs) / float64(p.TotalURLs) * 100)
}

// EnrichLlmBucket is LLM token usage for one purpose.
type EnrichLlmBucket struct {
	Input  int    `json:"input"`
	Output int    `json:"output"`
	Model  string `json:"model,omitempty"`
}

// EnrichUsage is the per-purpose usage envelope.
type EnrichUsage struct {
	Crawls             int                        `json:"crawls"`
	Searches           int                        `json:"searches"`
	LlmTokensByPurpose map[string]EnrichLlmBucket `json:"llm_tokens_by_purpose"`
	LlmTotals          map[string]int             `json:"llm_totals"`
}

// EnrichJobStatus is returned from POST /v1/enrich/async and GET /v1/enrich/jobs/{id}.
type EnrichJobStatus struct {
	JobID            string          `json:"job_id"`
	Status           EnrichStatus    `json:"status"`
	PhaseData        EnrichPhaseData `json:"phase_data"`
	Progress         EnrichProgress  `json:"progress"`
	Usage            EnrichUsage     `json:"usage"`
	AutoConfirmPlan  bool            `json:"auto_confirm_plan"`
	AutoConfirmURLs  bool            `json:"auto_confirm_urls"`
	CreatedAt        string          `json:"created_at,omitempty"`
	StartedAt        string          `json:"started_at,omitempty"`
	PausedAt         string          `json:"paused_at,omitempty"`
	CompletedAt      string          `json:"completed_at,omitempty"`
	Error            string          `json:"error,omitempty"`
}

// IsComplete returns true when the enrichment job is in a terminal state.
func (j *EnrichJobStatus) IsComplete() bool {
	for _, s := range EnrichTerminalStatuses {
		if j.Status == s {
			return true
		}
	}
	return false
}

// IsPaused returns true when the job is at plan_ready or urls_ready.
func (j *EnrichJobStatus) IsPaused() bool {
	for _, s := range EnrichPausedStatuses {
		if j.Status == s {
			return true
		}
	}
	return false
}

// IsSuccessful returns true when the enrichment job completed with usable results.
func (j *EnrichJobStatus) IsSuccessful() bool {
	return j.Status == EnrichStatusCompleted || j.Status == EnrichStatusPartial
}

// EnrichJobListItem is one row in the GET /v1/enrich/jobs list response.
type EnrichJobListItem struct {
	JobID         string       `json:"job_id"`
	Status        EnrichStatus `json:"status"`
	QueryPreview  string       `json:"query_preview,omitempty"`
	CreatedAt     string       `json:"created_at,omitempty"`
	CompletedAt   string       `json:"completed_at,omitempty"`
}

// EnrichOptions configures POST /v1/enrich/async.
//
// At least one of Query, Entities, or URLs must be set.
type EnrichOptions struct {
	// Inputs
	Query    string                 `json:"-"`
	Entities []EnrichEntity         `json:"-"`
	Criteria []EnrichCriterion      `json:"-"`
	Features []EnrichFeature        `json:"-"`
	URLs     []string               `json:"-"`
	Groups   map[string][]string    `json:"-"`

	// Phase control — both default true (one-shot mode).
	AutoConfirmPlan *bool `json:"-"`
	AutoConfirmURLs *bool `json:"-"`

	// Discover knobs
	TopKPerEntity int    `json:"-"` // default 3
	Search        *bool  `json:"-"` // default true
	Country       string `json:"-"`
	LocationHint  string `json:"-"`

	// Standard wrapper knobs
	Strategy      string                 `json:"-"` // "http" (default) or "browser"
	Config        map[string]interface{} `json:"-"`
	BrowserConfig map[string]interface{} `json:"-"`
	CrawlerConfig map[string]interface{} `json:"-"`
	LLMConfig     map[string]interface{} `json:"-"`
	Proxy         map[string]interface{} `json:"-"`
	WebhookURL    string                 `json:"-"`
	Priority      int                    `json:"-"`

	// Polling
	Wait         bool          `json:"-"`
	PollInterval time.Duration `json:"-"`
	Timeout      time.Duration `json:"-"`
}

// ResumeEnrichOptions are edits applied on POST /v1/enrich/jobs/{id}/continue.
//
// Pass nil/empty to resume with the server's current values.
type ResumeEnrichOptions struct {
	Entities []EnrichEntity         `json:"-"`
	Criteria []EnrichCriterion      `json:"-"`
	Features []EnrichFeature        `json:"-"`
	Groups   map[string][]string    `json:"-"`
}

// WaitEnrichOptions controls WaitEnrichJob.
type WaitEnrichOptions struct {
	Until        EnrichStatus  // empty → wait for any terminal status
	PollInterval time.Duration // default 3s
	Timeout      time.Duration // default 10m
}

// EnrichEvent is one Server-Sent Event from StreamEnrichJob.
//
// Type is one of "snapshot", "phase", "fragment", "row", "complete".
type EnrichEvent struct {
	Type     string                 // event name from the server
	Status   EnrichStatus           // populated on phase/complete events
	Snapshot *EnrichJobStatus       // populated on snapshot events
	Fragment map[string]interface{} // populated on fragment events
	Row      *EnrichRow             // populated on row events
	Raw      map[string]interface{} // unparsed payload, always set
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
//
// When `criteria` was in the request, GeneratedConfig carries the LLM-generated
// scan + extract config. When `extract` was set, ExtractionMethodUsed tells you
// whether CSS schema generation or LLM extraction was picked, and SchemaUsed
// holds the generated CSS schema (if any).
//
// Poll progress with GetSiteCrawlJob(job_id).
type SiteCrawlResponse struct {
	JobID                string                 `json:"job_id"`
	Status               string                 `json:"status"`
	Strategy             string                 `json:"strategy"`
	DiscoveredURLs       int                    `json:"discovered_urls"`
	QueuedURLs           int                    `json:"queued_urls"`
	CreatedAt            string                 `json:"created_at"`
	GeneratedConfig      *GeneratedConfig       `json:"generated_config,omitempty"`
	ExtractionMethodUsed string                 `json:"extraction_method_used,omitempty"` // "llm" | "css_schema"
	SchemaUsed           map[string]interface{} `json:"schema_used,omitempty"`
}

// SiteCrawlProgress is the progress block inside SiteCrawlJobStatus.
type SiteCrawlProgress struct {
	UrlsDiscovered int `json:"urls_discovered"`
	UrlsCrawled    int `json:"urls_crawled"`
	UrlsFailed     int `json:"urls_failed"`
	Total          int `json:"total"`
}

// SiteCrawlJobStatus is the polling response for GET /v1/crawl/site/jobs/{job_id}.
//
// Unified scan+crawl polling endpoint. Phase walks through three values:
// "scan" (URL discovery in progress), "crawl" (pages being fetched + extracted),
// "done" (everything finished). When Phase is "done" and Status is "completed",
// DownloadURL is a fresh S3 presigned URL (1-hour expiry) for the result ZIP.
type SiteCrawlJobStatus struct {
	JobID       string            `json:"job_id"`
	Status      string            `json:"status"`
	Phase       string            `json:"phase"` // "scan" | "crawl" | "done"
	Progress    SiteCrawlProgress `json:"progress"`
	ScanJobID   string            `json:"scan_job_id,omitempty"`
	CrawlJobID  string            `json:"crawl_job_id,omitempty"`
	DownloadURL string            `json:"download_url,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	CompletedAt string            `json:"completed_at,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// IsComplete returns true when the site crawl is in a terminal state.
func (j *SiteCrawlJobStatus) IsComplete() bool {
	if j.Phase == "done" {
		return true
	}
	switch j.Status {
	case "completed", "partial", "failed", "cancelled":
		return true
	}
	return false
}

// IsSuccessful returns true when the crawl finished with usable results.
func (j *SiteCrawlJobStatus) IsSuccessful() bool {
	return j.Status == "completed" || j.Status == "partial"
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

	// AI-assisted fields (new in 0.4.0)
	Criteria        string             `json:"-"` // plain-English — triggers LLM config gen
	Scan            *SiteScanConfig    `json:"-"` // explicit scan overrides
	Extract         *SiteExtractConfig `json:"-"` // structured extraction config
	IncludeMarkdown *bool              `json:"-"` // legacy flag; nil = unset

	// Async polling
	Wait         bool          `json:"-"`
	PollInterval time.Duration `json:"-"`
	Timeout      time.Duration `json:"-"`
}
