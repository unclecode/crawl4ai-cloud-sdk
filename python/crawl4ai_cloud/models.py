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

    Examples:
        # Datacenter proxy
        proxy = ProxyConfig(mode="datacenter")

        # Residential with geo-targeting
        proxy = ProxyConfig(mode="residential", country="US")

        # Deep crawl with sticky session (same IP for all URLs)
        proxy = ProxyConfig(mode="datacenter", sticky_session=True)
    """
    mode: str = "none"
    country: Optional[str] = None
    sticky_session: bool = False

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dict for API request."""
        result: Dict[str, Any] = {"mode": self.mode}
        if self.country:
            result["country"] = self.country
        if self.sticky_session:
            result["sticky_session"] = True
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
        progress_data = data.get("progress", {})
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
class DeepCrawlResult:
    """Deep crawl response."""
    job_id: str
    status: str
    strategy: str
    discovered_count: int
    queued_urls: int
    created_at: str
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
        return self.status in ("completed", "failed")

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
            usage=Usage.from_dict(data["usage"]) if data.get("usage") else None,
        )
