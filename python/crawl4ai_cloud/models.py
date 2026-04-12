"""Response models for Crawl4AI Cloud SDK."""
from dataclasses import dataclass, field
from typing import Optional, Dict, List, Any


@dataclass
class ProxyConfig:
    """
    Proxy configuration for cloud crawl requests.

    Proxy Modes:
    - "none": Direct connection (1x credits)
    - "datacenter": Fast datacenter proxies (2x credits)
    - "residential": Premium residential IPs (5x credits)
    - "auto": Smart selection based on target URL

    Attributes:
        mode: Proxy mode (see above).
        country: ISO country code for geo-targeting (e.g. "US", "DE").
        sticky_session: Use the same IP for all URLs in a batch request.
        use_proxy: Whether to route through a proxy at all. Set to False
            to force a direct connection regardless of server defaults.
        skip_direct: Skip the initial direct-connection attempt and go
            straight to the proxy. Useful for sites known to block
            datacenter IPs.

    Examples:
        # Datacenter proxy
        proxy = ProxyConfig(mode="datacenter")

        # Residential with geo-targeting
        proxy = ProxyConfig(mode="residential", country="US")

        # Deep crawl with sticky session (same IP for all URLs)
        proxy = ProxyConfig(mode="datacenter", sticky_session=True)

        # Force direct connection (no proxy)
        proxy = ProxyConfig(use_proxy=False)

        # Skip the direct attempt, go straight to proxy
        proxy = ProxyConfig(mode="datacenter", skip_direct=True)
    """
    mode: str = "none"
    country: Optional[str] = None
    sticky_session: bool = False
    use_proxy: bool = True
    skip_direct: bool = False

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dict for API request."""
        result: Dict[str, Any] = {
            "mode": self.mode,
            "use_proxy": self.use_proxy,
        }
        if self.country:
            result["country"] = self.country
        if self.sticky_session:
            result["sticky_session"] = True
        if self.skip_direct:
            result["skip_direct"] = True
        return result


@dataclass
class JobProgress:
    """Async job progress."""
    total: int
    completed: int
    failed: int

    @property
    def pending(self) -> int:
        """Get pending count."""
        return self.total - self.completed - self.failed

    @property
    def percent(self) -> float:
        """Get completion percentage."""
        if self.total == 0:
            return 0.0
        return ((self.completed + self.failed) / self.total) * 100


@dataclass
class CrawlJob:
    """Async crawl job returned by run_many()."""
    job_id: str
    status: str
    progress: JobProgress
    urls_count: int
    created_at: str
    started_at: Optional[str] = None
    completed_at: Optional[str] = None
    results: Optional[List["CrawlResult"]] = None
    error: Optional[str] = None
    result_size_bytes: Optional[int] = None
    usage: Optional["Usage"] = None  # Resource usage metrics (completed jobs only)

    @property
    def id(self) -> str:
        """Alias for job_id (backward compatibility)."""
        return self.job_id

    @property
    def is_complete(self) -> bool:
        """Check if job is in a terminal state."""
        return self.status in ("completed", "partial", "failed", "cancelled")

    @property
    def is_successful(self) -> bool:
        """Check if job completed successfully."""
        return self.status == "completed"

    @property
    def progress_percent(self) -> float:
        """Get completion percentage."""
        return self.progress.percent

    @classmethod
    def from_dict(cls, data: Dict[str, Any], convert_results: bool = True) -> "CrawlJob":
        """Create CrawlJob from API response dict.

        Args:
            data: API response dictionary
            convert_results: If True, convert results to CrawlResult objects
        """
        progress_data = data.get("progress") or {}
        progress = JobProgress(
            total=progress_data.get("total", 0),
            completed=progress_data.get("completed", 0),
            failed=progress_data.get("failed", 0),
        )

        job_id = data.get("job_id", "")

        # Convert results to CrawlResult objects if present
        results = None
        raw_results = data.get("results")
        if raw_results and convert_results:
            # Import here to avoid circular import at module level
            results = [CrawlResult.from_dict(r) for r in raw_results]
            # Set job_id on each result for use with download_url()
            for r in results:
                r.id = job_id
        elif raw_results:
            results = raw_results

        # Parse usage if present
        usage = None
        if data.get("usage"):
            usage = Usage.from_dict(data["usage"])

        return cls(
            job_id=job_id,
            status=data.get("status", "unknown"),
            progress=progress,
            urls_count=data.get("urls_count", data.get("url_count", 0)),
            created_at=data.get("created_at", ""),
            started_at=data.get("started_at"),
            completed_at=data.get("completed_at"),
            results=results,
            error=data.get("error"),
            result_size_bytes=data.get("result_size_bytes"),
            usage=usage,
        )


@dataclass
class ScanUrlInfo:
    """Information about a scanned URL (for scan_only mode)."""
    url: str
    depth: int
    score: Optional[float] = None
    links_found: int = 0
    html_size: int = 0


@dataclass
class DomainScanUrlInfo:
    """URL discovered by domain scan (/v1/scan)."""
    url: str
    host: str = ""
    status: str = "valid"
    relevance_score: Optional[float] = None
    head_data: Optional[Dict[str, Any]] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "DomainScanUrlInfo":
        return cls(
            url=data.get("url", ""),
            host=data.get("host", ""),
            status=data.get("status", "valid"),
            relevance_score=data.get("relevance_score"),
            head_data=data.get("head_data"),
        )


@dataclass
class SiteScanConfig:
    """
    Unified scan configuration for AI-assisted URL discovery.

    Used by /v1/scan and /v1/crawl/site. When `criteria` is set on the parent
    request, the AI config generator fills unset fields here. Explicit fields
    always win over LLM output.

    Fields match the SiteScanConfig Pydantic model in the backend.
    """
    mode: str = "auto"  # "auto" | "map" | "deep"
    patterns: Optional[List[str]] = None
    filters: Optional[Dict[str, Any]] = None
    scorers: Optional[Dict[str, Any]] = None
    query: Optional[str] = None
    score_threshold: Optional[float] = None
    include_subdomains: bool = False
    max_depth: Optional[int] = None

    def to_dict(self) -> Dict[str, Any]:
        d: Dict[str, Any] = {"mode": self.mode}
        if self.patterns is not None:
            d["patterns"] = self.patterns
        if self.filters is not None:
            d["filters"] = self.filters
        if self.scorers is not None:
            d["scorers"] = self.scorers
        if self.query is not None:
            d["query"] = self.query
        if self.score_threshold is not None:
            d["score_threshold"] = self.score_threshold
        if self.include_subdomains:
            d["include_subdomains"] = True
        if self.max_depth is not None:
            d["max_depth"] = self.max_depth
        return d


@dataclass
class SiteExtractConfig:
    """
    Structured extraction configuration for /v1/crawl/site.

    Mirrors /v1/extract's shape. When set without a pre-built `schema`, the
    wrapper fetches `sample_url` (defaults to the crawl's start URL), generates
    a schema via LLM, and injects it into crawler_config.extraction_strategy
    for all discovered URLs.
    """
    query: Optional[str] = None
    json_example: Optional[Dict[str, Any]] = None
    method: str = "auto"  # "auto" | "llm" | "schema"
    schema: Optional[Dict[str, Any]] = None
    sample_url: Optional[str] = None
    url_pattern: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        d: Dict[str, Any] = {"method": self.method}
        if self.query is not None:
            d["query"] = self.query
        if self.json_example is not None:
            d["json_example"] = self.json_example
        if self.schema is not None:
            d["schema"] = self.schema
        if self.sample_url is not None:
            d["sample_url"] = self.sample_url
        if self.url_pattern is not None:
            d["url_pattern"] = self.url_pattern
        return d


@dataclass
class GeneratedConfig:
    """
    LLM-generated config echoed back by /v1/scan and /v1/crawl/site when
    `criteria` was set in the request. Contains the scan config and (for
    /v1/crawl/site) the extract config, plus LLM reasoning and cache/fallback
    flags.
    """
    scan: Dict[str, Any]
    reasoning: str = ""
    extract: Optional[Dict[str, Any]] = None
    fallback: bool = False
    cached: bool = False

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "GeneratedConfig":
        return cls(
            scan=data.get("scan") or {},
            reasoning=data.get("reasoning", ""),
            extract=data.get("extract"),
            fallback=bool(data.get("fallback", False)),
            cached=bool(data.get("cached", False)),
        )


@dataclass
class ScanResult:
    """
    Response from /v1/scan.

    For map mode (sync): `urls` is populated inline.
    For deep mode (async): `job_id` + `status` are set; poll with
    `get_scan_job(job_id)` and the `urls` list will be populated progressively.

    When `criteria` was supplied in the request, `generated_config` carries
    the LLM output and `mode_used` tells you which strategy ran.
    """
    success: bool
    domain: str
    total_urls: int
    hosts_found: int
    mode: str
    urls: List[DomainScanUrlInfo]
    duration_ms: int
    error: Optional[str] = None
    # AI-assisted / async fields
    mode_used: Optional[str] = None  # "map" | "deep"
    job_id: Optional[str] = None     # set when async deep mode kicked off
    status: Optional[str] = None     # "pending" | "running" | "completed" | "failed"
    generated_config: Optional[GeneratedConfig] = None
    message: Optional[str] = None

    @property
    def is_async(self) -> bool:
        """True when the response is for an async (deep) scan — poll with get_scan_job()."""
        return self.mode_used == "deep" and bool(self.job_id)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ScanResult":
        urls = [DomainScanUrlInfo.from_dict(u) for u in data.get("urls", [])]
        generated_config = None
        if data.get("generated_config"):
            generated_config = GeneratedConfig.from_dict(data["generated_config"])
        return cls(
            success=data.get("success", False),
            domain=data.get("domain", ""),
            total_urls=data.get("total_urls", 0),
            hosts_found=data.get("hosts_found", 0),
            mode=data.get("mode", "default"),
            urls=urls,
            duration_ms=data.get("duration_ms", 0),
            error=data.get("error"),
            mode_used=data.get("mode_used"),
            job_id=data.get("job_id"),
            status=data.get("status"),
            generated_config=generated_config,
            message=data.get("message"),
        )


@dataclass
class ScanJobStatus:
    """
    Polling response for GET /v1/scan/jobs/{job_id} — used with async deep
    scans. URLs are appended to `urls` as they're discovered. `progress`
    carries `{completed, total}` once the backend starts tracking.
    """
    job_id: str
    status: str
    mode_used: str = "deep"
    domain: Optional[str] = None
    total_urls: int = 0
    urls: List[DomainScanUrlInfo] = field(default_factory=list)
    progress: Optional[Dict[str, int]] = None
    generated_config: Optional[GeneratedConfig] = None
    duration_ms: int = 0
    error: Optional[str] = None
    created_at: Optional[str] = None
    completed_at: Optional[str] = None

    @property
    def is_complete(self) -> bool:
        return self.status in ("completed", "partial", "failed", "cancelled")

    @property
    def is_successful(self) -> bool:
        return self.status in ("completed", "partial")

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ScanJobStatus":
        urls = [DomainScanUrlInfo.from_dict(u) for u in data.get("urls") or []]
        generated_config = None
        if data.get("generated_config"):
            generated_config = GeneratedConfig.from_dict(data["generated_config"])
        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", "pending"),
            mode_used=data.get("mode_used", "deep"),
            domain=data.get("domain"),
            total_urls=data.get("total_urls", 0),
            urls=urls,
            progress=data.get("progress"),
            generated_config=generated_config,
            duration_ms=data.get("duration_ms", 0),
            error=data.get("error"),
            created_at=data.get("created_at"),
            completed_at=data.get("completed_at"),
        )


@dataclass
class DeepCrawlResult:
    """Deep crawl response."""
    job_id: str
    status: str
    strategy: str
    discovered_count: int
    queued_urls: int = 0
    created_at: str = ""
    urls: Optional[List[ScanUrlInfo]] = None
    html_download_url: Optional[str] = None
    cache_expires_at: Optional[str] = None
    crawl_job_id: Optional[str] = None

    @property
    def discovered_urls(self) -> List[str]:
        """Get list of discovered URL strings."""
        if self.urls:
            return [u.url for u in self.urls]
        return []

    @property
    def has_urls(self) -> bool:
        """Check if URLs were discovered."""
        return self.status != "no_urls" and self.discovered_count > 0

    @property
    def is_complete(self) -> bool:
        """Check if the scan job has finished."""
        return self.status in ("completed", "failed", "cancelled")

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "DeepCrawlResult":
        """Create DeepCrawlResult from API response dict."""
        urls = None
        if data.get("urls"):
            urls = [
                ScanUrlInfo(
                    url=u.get("url", ""),
                    depth=u.get("depth", 0),
                    score=u.get("score"),
                    links_found=u.get("links_found", 0),
                    html_size=u.get("html_size", 0),
                )
                for u in data["urls"]
            ]

        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", ""),
            strategy=data.get("strategy", "map"),
            discovered_count=data.get("discovered_urls", 0),
            queued_urls=data.get("queued_urls", 0),
            created_at=data.get("created_at", ""),
            urls=urls,
            html_download_url=data.get("html_download_url"),
            cache_expires_at=data.get("cache_expires_at"),
            crawl_job_id=data.get("crawl_job_id"),
        )


@dataclass
class ContextResult:
    """Context API response."""
    job_id: str
    status: str
    query: str
    download_url: str
    urls_crawled: int
    size_bytes: int
    duration_ms: int
    cached: bool = False

    @property
    def size_mb(self) -> float:
        """Size in megabytes."""
        return self.size_bytes / (1024 * 1024)

    @property
    def duration_seconds(self) -> float:
        """Duration in seconds."""
        return self.duration_ms / 1000

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ContextResult":
        """Create ContextResult from API response dict."""
        return cls(
            job_id=data["job_id"],
            status=data["status"],
            query=data["query"],
            download_url=data["download_url"],
            size_bytes=data.get("storage_size_bytes", 0),
            urls_crawled=data.get("urls_crawled", 0),
            duration_ms=data.get("duration_ms", 0),
            cached=data.get("cached", False),
        )


@dataclass
class LLMUsage:
    """LLM token usage for managed service (per-request)."""
    prompt_tokens: int = 0
    completion_tokens: int = 0
    total_tokens: int = 0


@dataclass
class CrawlUsageMetrics:
    """Crawl resource usage metrics returned in API responses."""
    credits_used: float = 0.0
    credits_remaining: float = 0.0
    duration_ms: int = 0
    cached: bool = False  # bool for single, but API may return int for batch
    urls_total: Optional[int] = None
    urls_succeeded: Optional[int] = None
    urls_failed: Optional[int] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "CrawlUsageMetrics":
        """Create from API response dict."""
        return cls(
            credits_used=data.get("credits_used", 0.0),
            credits_remaining=data.get("credits_remaining", 0.0),
            duration_ms=data.get("duration_ms", 0),
            cached=bool(data.get("cached", False)),
            urls_total=data.get("urls_total"),
            urls_succeeded=data.get("urls_succeeded"),
            urls_failed=data.get("urls_failed"),
        )


@dataclass
class LLMUsageMetrics:
    """LLM usage metrics returned in API responses."""
    tokens_used: int = 0
    tokens_remaining: int = 0
    model: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "LLMUsageMetrics":
        """Create from API response dict."""
        return cls(
            tokens_used=data.get("tokens_used", 0),
            tokens_remaining=data.get("tokens_remaining", 0),
            model=data.get("model"),
        )


@dataclass
class StorageUsageMetrics:
    """Storage usage metrics returned in API responses (async jobs only)."""
    bytes_used: int = 0
    bytes_remaining: int = 0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "StorageUsageMetrics":
        """Create from API response dict."""
        return cls(
            bytes_used=data.get("bytes_used", 0),
            bytes_remaining=data.get("bytes_remaining", 0),
        )


@dataclass
class Usage:
    """
    Unified usage metrics returned in API responses.

    Shows resource consumption and remaining quotas after each request.

    Attributes:
        crawl: Crawl credits used and remaining
        llm: LLM tokens used and remaining (only if LLM extraction was used)
        storage: Storage bytes used and remaining (async jobs only)
    """
    crawl: CrawlUsageMetrics
    llm: Optional[LLMUsageMetrics] = None
    storage: Optional[StorageUsageMetrics] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Usage":
        """Create from API response dict."""
        crawl = CrawlUsageMetrics.from_dict(data.get("crawl", {}))

        llm = None
        if data.get("llm"):
            llm = LLMUsageMetrics.from_dict(data["llm"])

        storage = None
        if data.get("storage"):
            storage = StorageUsageMetrics.from_dict(data["storage"])

        return cls(crawl=crawl, llm=llm, storage=storage)


@dataclass
class GeneratedSchema:
    """Generated extraction schema from LLM."""
    success: bool
    schema: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
    llm_usage: Optional[LLMUsage] = None

    @property
    def fields(self) -> List[Dict[str, Any]]:
        """Get the fields/selectors from the generated schema."""
        if self.schema:
            if isinstance(self.schema, list):
                return self.schema
            return self.schema.get("fields", [])
        return []

    @property
    def name(self) -> Optional[str]:
        """Get the schema name."""
        if self.schema and isinstance(self.schema, dict):
            return self.schema.get("name")
        return None

    @property
    def base_selector(self) -> Optional[str]:
        """Get the base CSS/XPath selector."""
        if self.schema and isinstance(self.schema, dict):
            return self.schema.get("base_selector") or self.schema.get("baseSelector")
        return None

    def to_dict(self) -> Dict[str, Any]:
        """Convert the generated schema to a dict for use in extraction strategies."""
        return self.schema or {}

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "GeneratedSchema":
        """Create GeneratedSchema from API response dict."""
        llm_usage = None
        if data.get("llm_usage"):
            llm_usage = LLMUsage(
                prompt_tokens=data["llm_usage"].get("prompt_tokens", 0),
                completion_tokens=data["llm_usage"].get("completion_tokens", 0),
                total_tokens=data["llm_usage"].get("total_tokens", 0),
            )
        return cls(
            success=data.get("success", False),
            schema=data.get("schema"),
            error=data.get("error_message"),
            llm_usage=llm_usage,
        )


@dataclass
class StorageUsage:
    """Storage quota usage."""
    used_mb: float
    max_mb: float
    remaining_mb: float
    percent_used: float = 0.0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "StorageUsage":
        """Create StorageUsage from API response dict."""
        return cls(
            used_mb=data.get("used_mb", 0.0),
            max_mb=data.get("max_mb", 0.0),
            remaining_mb=data.get("remaining_mb", 0.0),
            percent_used=data.get("percent_used", 0.0),
        )


@dataclass
class MarkdownResult:
    """Markdown extraction result."""
    raw_markdown: Optional[str] = None
    markdown_with_citations: Optional[str] = None
    references_markdown: Optional[str] = None
    fit_markdown: Optional[str] = None


@dataclass
class CrawlResult:
    """Single URL crawl result from cloud API."""
    url: str
    success: bool
    html: Optional[str] = None
    cleaned_html: Optional[str] = None
    fit_html: Optional[str] = None
    markdown: Optional[MarkdownResult] = None
    media: Dict[str, List[Any]] = field(default_factory=dict)
    links: Dict[str, List[Any]] = field(default_factory=dict)
    metadata: Optional[Dict[str, Any]] = None
    screenshot: Optional[str] = None
    pdf: Optional[str] = None
    extracted_content: Optional[str] = None
    error_message: Optional[str] = None
    status_code: Optional[int] = None
    duration_ms: int = 0
    tables: List[Any] = field(default_factory=list)
    network_requests: Optional[List[Any]] = None
    console_messages: Optional[List[Any]] = None
    redirected_url: Optional[str] = None
    llm_usage: Optional[LLMUsage] = None
    crawl_strategy: Optional[str] = None
    proxy_mode: Optional[str] = None  # Effective proxy mode: "none", "datacenter", "residential"
    proxy_used: Optional[str] = None  # Proxy provider identifier
    downloaded_files: Optional[List[str]] = None  # Presigned S3 URLs for file downloads (CSV, PDF, XLSX, etc.)
    id: Optional[str] = None  # Job ID for async results (use with download_url())
    usage: Optional[Usage] = None  # Resource usage metrics

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "CrawlResult":
        """Create CrawlResult from API response dict."""
        markdown_data = data.get("markdown")
        markdown = None
        if markdown_data:
            # Handle both string (async results) and dict (sync results) formats
            if isinstance(markdown_data, str):
                markdown = MarkdownResult(raw_markdown=markdown_data)
            else:
                markdown = MarkdownResult(
                    raw_markdown=markdown_data.get("raw_markdown"),
                    markdown_with_citations=markdown_data.get("markdown_with_citations"),
                    references_markdown=markdown_data.get("references_markdown"),
                    fit_markdown=markdown_data.get("fit_markdown"),
                )

        llm_usage = None
        if data.get("llm_usage"):
            llm_usage = LLMUsage(
                prompt_tokens=data["llm_usage"].get("prompt_tokens", 0),
                completion_tokens=data["llm_usage"].get("completion_tokens", 0),
                total_tokens=data["llm_usage"].get("total_tokens", 0),
            )

        return cls(
            url=data.get("url", ""),
            success=data.get("success", False),
            html=data.get("html"),
            cleaned_html=data.get("cleaned_html"),
            fit_html=data.get("fit_html"),
            markdown=markdown,
            media=data.get("media", {}),
            links=data.get("links", {}),
            metadata=data.get("metadata"),
            screenshot=data.get("screenshot"),
            pdf=data.get("pdf"),
            extracted_content=data.get("extracted_content"),
            error_message=data.get("error_message"),
            status_code=data.get("status_code"),
            duration_ms=data.get("duration_ms", 0),
            tables=data.get("tables", []),
            network_requests=data.get("network_requests"),
            console_messages=data.get("console_messages"),
            redirected_url=data.get("redirected_url"),
            llm_usage=llm_usage,
            crawl_strategy=data.get("crawl_strategy"),
            proxy_mode=data.get("proxy_mode"),
            proxy_used=data.get("proxy_used"),
            downloaded_files=data.get("downloaded_files"),
            usage=Usage.from_dict(data["usage"]) if data.get("usage") else None,
        )


# =============================================================================
# Wrapper API Response Models
# =============================================================================


@dataclass
class WrapperUsage:
    """Credit usage from wrapper endpoints."""
    credits_used: float = 0
    credits_remaining: float = 0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WrapperUsage":
        return cls(
            credits_used=data.get("credits_used", 0),
            credits_remaining=data.get("credits_remaining", 0),
        )


@dataclass
class MarkdownResponse:
    """Response from POST /v1/markdown."""
    success: bool
    url: str
    markdown: Optional[str] = None
    fit_markdown: Optional[str] = None
    fit_html: Optional[str] = None
    links: Optional[Dict[str, Any]] = None
    media: Optional[Dict[str, Any]] = None
    metadata: Optional[Dict[str, Any]] = None
    tables: Optional[List[Dict[str, Any]]] = None
    duration_ms: int = 0
    usage: Optional[WrapperUsage] = None
    error_message: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "MarkdownResponse":
        return cls(
            success=data.get("success", False),
            url=data.get("url", ""),
            markdown=data.get("markdown"),
            fit_markdown=data.get("fit_markdown"),
            fit_html=data.get("fit_html"),
            links=data.get("links"),
            media=data.get("media"),
            metadata=data.get("metadata"),
            tables=data.get("tables"),
            duration_ms=data.get("duration_ms", 0),
            usage=WrapperUsage.from_dict(data["usage"]) if data.get("usage") else None,
            error_message=data.get("error_message"),
        )


@dataclass
class ScreenshotResponse:
    """Response from POST /v1/screenshot."""
    success: bool
    url: str
    screenshot: Optional[str] = None
    pdf: Optional[str] = None
    duration_ms: int = 0
    usage: Optional[WrapperUsage] = None
    error_message: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "ScreenshotResponse":
        return cls(
            success=data.get("success", False),
            url=data.get("url", ""),
            screenshot=data.get("screenshot"),
            pdf=data.get("pdf"),
            duration_ms=data.get("duration_ms", 0),
            usage=WrapperUsage.from_dict(data["usage"]) if data.get("usage") else None,
            error_message=data.get("error_message"),
        )


@dataclass
class ExtractResponse:
    """Response from POST /v1/extract."""
    success: bool
    url: Optional[str] = None
    data: Optional[List[Dict[str, Any]]] = None
    method_used: Optional[str] = None
    schema_used: Optional[Dict[str, Any]] = None
    query_used: Optional[str] = None
    llm_usage: Optional[LLMUsage] = None
    duration_ms: int = 0
    error_message: Optional[str] = None

    @classmethod
    def from_dict(cls, d: Dict[str, Any]) -> "ExtractResponse":
        llm_usage = None
        if d.get("llm_usage"):
            llm_usage = LLMUsage(
                prompt_tokens=d["llm_usage"].get("prompt_tokens", 0),
                completion_tokens=d["llm_usage"].get("completion_tokens", 0),
                total_tokens=d["llm_usage"].get("total_tokens", 0),
            )
        return cls(
            success=d.get("success", False),
            url=d.get("url"),
            data=d.get("data"),
            method_used=d.get("method_used"),
            schema_used=d.get("schema_used"),
            query_used=d.get("query_used"),
            llm_usage=llm_usage,
            duration_ms=d.get("duration_ms", 0),
            error_message=d.get("error_message"),
        )


@dataclass
class MapUrlInfo:
    """A discovered URL from POST /v1/map."""
    url: str
    host: str = ""
    status: str = "valid"
    relevance_score: Optional[float] = None
    head_data: Optional[Dict[str, Any]] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "MapUrlInfo":
        return cls(
            url=data.get("url", ""),
            host=data.get("host", ""),
            status=data.get("status", "valid"),
            relevance_score=data.get("relevance_score"),
            head_data=data.get("head_data"),
        )


@dataclass
class MapResponse:
    """Response from POST /v1/map."""
    success: bool
    domain: str = ""
    total_urls: int = 0
    hosts_found: int = 0
    mode: str = "default"
    urls: List[MapUrlInfo] = field(default_factory=list)
    duration_ms: int = 0
    error_message: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "MapResponse":
        return cls(
            success=data.get("success", False),
            domain=data.get("domain", ""),
            total_urls=data.get("total_urls", 0),
            hosts_found=data.get("hosts_found", 0),
            mode=data.get("mode", "default"),
            urls=[MapUrlInfo.from_dict(u) for u in data.get("urls", [])],
            duration_ms=data.get("duration_ms", 0),
            error_message=data.get("error_message"),
        )


@dataclass
class SiteCrawlResponse:
    """
    Response from POST /v1/crawl/site.

    When `criteria` was in the request, `generated_config` carries the
    LLM-generated scan + extract config. When `extract` was set,
    `extraction_method_used` tells you whether CSS schema generation or LLM
    extraction was picked, and `schema_used` holds the generated CSS schema
    (if any) so you can reuse it on future crawls.

    Poll progress with `get_site_crawl_job(job_id)`.
    """
    job_id: str
    status: str = "pending"
    strategy: str = "map"
    discovered_urls: int = 0
    queued_urls: int = 0
    created_at: str = ""
    generated_config: Optional[GeneratedConfig] = None
    extraction_method_used: Optional[str] = None  # "llm" | "css_schema"
    schema_used: Optional[Dict[str, Any]] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SiteCrawlResponse":
        generated_config = None
        if data.get("generated_config"):
            generated_config = GeneratedConfig.from_dict(data["generated_config"])
        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", "pending"),
            strategy=data.get("strategy", "map"),
            discovered_urls=data.get("discovered_urls", 0),
            queued_urls=data.get("queued_urls", 0),
            created_at=data.get("created_at", ""),
            generated_config=generated_config,
            extraction_method_used=data.get("extraction_method_used"),
            schema_used=data.get("schema_used"),
        )


@dataclass
class SiteCrawlProgress:
    """Progress block inside SiteCrawlJobStatus."""
    urls_discovered: int = 0
    urls_crawled: int = 0
    urls_failed: int = 0
    total: int = 0

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SiteCrawlProgress":
        return cls(
            urls_discovered=data.get("urls_discovered", 0),
            urls_crawled=data.get("urls_crawled", 0),
            urls_failed=data.get("urls_failed", 0),
            total=data.get("total", 0),
        )


@dataclass
class SiteCrawlJobStatus:
    """
    Polling response for GET /v1/crawl/site/jobs/{job_id}.

    This is the unified scan+crawl polling endpoint. `phase` walks through
    three values: "scan" (URL discovery in progress), "crawl" (pages being
    fetched + extracted), "done" (everything finished).

    When phase is "done" and status is "completed", `download_url` is a fresh
    S3 presigned URL (1-hour expiry) for the result ZIP.
    """
    job_id: str
    status: str = "pending"
    phase: str = "scan"  # "scan" | "crawl" | "done"
    progress: SiteCrawlProgress = field(default_factory=SiteCrawlProgress)
    scan_job_id: Optional[str] = None
    crawl_job_id: Optional[str] = None
    download_url: Optional[str] = None
    created_at: Optional[str] = None
    completed_at: Optional[str] = None
    error: Optional[str] = None

    @property
    def is_complete(self) -> bool:
        return self.phase == "done" or self.status in (
            "completed", "partial", "failed", "cancelled"
        )

    @property
    def is_successful(self) -> bool:
        return self.status in ("completed", "partial")

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "SiteCrawlJobStatus":
        progress = SiteCrawlProgress.from_dict(data.get("progress") or {})
        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", "pending"),
            phase=data.get("phase", "scan"),
            progress=progress,
            scan_job_id=data.get("scan_job_id"),
            crawl_job_id=data.get("crawl_job_id"),
            download_url=data.get("download_url"),
            created_at=data.get("created_at"),
            completed_at=data.get("completed_at"),
            error=data.get("error"),
        )


@dataclass
class EnrichFieldSource:
    """Source attribution for a single enriched field."""
    url: str = ""
    method: str = ""  # "direct", "depth", "search"

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EnrichFieldSource":
        return cls(url=data.get("url", ""), method=data.get("method", ""))


@dataclass
class EnrichSearchCitation:
    """Citation for a field found via search fallback."""
    field: str = ""
    source_url: str = ""
    source_title: str = ""
    query_used: str = ""

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EnrichSearchCitation":
        return cls(
            field=data.get("field", ""),
            source_url=data.get("source_url", ""),
            source_title=data.get("source_title", ""),
            query_used=data.get("query_used", ""),
        )


@dataclass
class EnrichRow:
    """Result for a single URL in an enrichment job."""
    url: str = ""
    fields: Dict[str, Any] = field(default_factory=dict)
    missing: List[str] = field(default_factory=list)
    sources: Dict[str, EnrichFieldSource] = field(default_factory=dict)
    search_citations: List[EnrichSearchCitation] = field(default_factory=list)
    status: str = "pending"  # "complete", "partial", "failed"
    depth_used: int = 0
    search_used: bool = False
    token_usage: Optional[Dict[str, int]] = None
    duration_ms: int = 0
    error: Optional[str] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EnrichRow":
        sources = {}
        for fname, src in (data.get("sources") or {}).items():
            sources[fname] = EnrichFieldSource.from_dict(src) if isinstance(src, dict) else EnrichFieldSource()
        citations = [
            EnrichSearchCitation.from_dict(c) for c in (data.get("search_citations") or [])
        ]
        return cls(
            url=data.get("url", ""),
            fields=data.get("fields") or {},
            missing=data.get("missing") or [],
            sources=sources,
            search_citations=citations,
            status=data.get("status", "pending"),
            depth_used=data.get("depth_used", 0),
            search_used=data.get("search_used", False),
            token_usage=data.get("token_usage"),
            duration_ms=data.get("duration_ms", 0),
            error=data.get("error"),
        )


@dataclass
class EnrichJobProgress:
    """Progress for an enrichment job."""
    total: int = 0
    completed: int = 0
    failed: int = 0

    @property
    def percent(self) -> int:
        if self.total == 0:
            return 0
        return int((self.completed + self.failed) / self.total * 100)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EnrichJobProgress":
        return cls(
            total=data.get("total", 0),
            completed=data.get("completed", 0),
            failed=data.get("failed", 0),
        )


@dataclass
class EnrichResponse:
    """Response from POST /v1/enrich."""
    job_id: str = ""
    status: str = "pending"
    urls_count: int = 0
    schema_fields: int = 0
    created_at: str = ""

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EnrichResponse":
        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", "pending"),
            urls_count=data.get("urls_count", 0),
            schema_fields=data.get("schema_fields", 0),
            created_at=data.get("created_at", ""),
        )


@dataclass
class EnrichJobStatus:
    """Polling response for GET /v1/enrich/jobs/{job_id}."""
    job_id: str = ""
    status: str = "pending"
    progress: EnrichJobProgress = field(default_factory=EnrichJobProgress)
    progress_percent: int = 0
    rows: Optional[List[EnrichRow]] = None
    created_at: Optional[str] = None
    started_at: Optional[str] = None
    completed_at: Optional[str] = None
    error: Optional[str] = None

    @property
    def is_complete(self) -> bool:
        return self.status in ("completed", "partial", "failed", "cancelled")

    @property
    def is_successful(self) -> bool:
        return self.status in ("completed", "partial")

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "EnrichJobStatus":
        progress = EnrichJobProgress.from_dict(data.get("progress") or {})
        rows = None
        if data.get("rows") is not None:
            rows = [EnrichRow.from_dict(r) for r in data["rows"]]
        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", "pending"),
            progress=progress,
            progress_percent=data.get("progress_percent", 0),
            rows=rows,
            created_at=data.get("created_at"),
            started_at=data.get("started_at"),
            completed_at=data.get("completed_at"),
            error=data.get("error"),
        )


@dataclass
class WrapperJobProgress:
    """Progress for a wrapper async job."""
    total: int = 0
    completed: int = 0
    failed: int = 0

    @property
    def percent(self) -> int:
        if self.total == 0:
            return 0
        return int((self.completed + self.failed) / self.total * 100)

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WrapperJobProgress":
        return cls(
            total=data.get("total", 0),
            completed=data.get("completed", 0),
            failed=data.get("failed", 0),
        )


@dataclass
class WrapperJob:
    """Job status for wrapper async endpoints (/v1/markdown/async, etc.)."""
    job_id: str
    status: str = "pending"
    progress: Optional[WrapperJobProgress] = None
    progress_percent: int = 0
    urls_count: int = 0
    error: Optional[str] = None
    created_at: Optional[str] = None
    started_at: Optional[str] = None
    completed_at: Optional[str] = None

    @property
    def is_complete(self) -> bool:
        return self.status in ("completed", "partial", "failed", "cancelled")

    @property
    def is_successful(self) -> bool:
        return self.status in ("completed", "partial")

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "WrapperJob":
        progress = None
        if data.get("progress"):
            progress = WrapperJobProgress.from_dict(data["progress"])
        return cls(
            job_id=data.get("job_id", ""),
            status=data.get("status", "pending"),
            progress=progress,
            progress_percent=data.get("progress_percent", 0),
            urls_count=data.get("urls_count", 0),
            error=data.get("error"),
            created_at=data.get("created_at"),
            started_at=data.get("started_at"),
            completed_at=data.get("completed_at"),
        )
