"""AsyncWebCrawler - Main crawler class for Crawl4AI Cloud SDK."""
import asyncio
import time
import warnings
from typing import Optional, Dict, Any, List, Union

from ._client import HTTPClient
from .errors import TimeoutError
from .models import (
    CrawlJob,
    CrawlResult,
    DeepCrawlResult,
    ContextResult,
    GeneratedSchema,
    StorageUsage,
    ProxyConfig,
    JobProgress,
    MarkdownResponse,
    ScreenshotResponse,
    ExtractResponse,
    MapResponse,
    SiteCrawlResponse,
    WrapperJob,
    ScanResult,
)
from .configs import (
    CrawlerRunConfig,
    BrowserConfig,
    sanitize_crawler_config,
    sanitize_browser_config,
    normalize_proxy,
    build_crawl_request,
)


# ─── Enrich vocabulary normalizers (string shortcuts) ────────────────

def _normalize_entity(item: Union[str, Dict[str, Any]]) -> Dict[str, Any]:
    """Accept `"Toronto"` or `{"name": "...", "title": "...", "source_url": "..."}`."""
    if isinstance(item, str):
        return {"name": item}
    return item


def _normalize_criterion(item: Union[str, Dict[str, Any]]) -> Dict[str, Any]:
    """Accept `"Austin TX"` or `{"text": "...", "kind": "location"}`."""
    if isinstance(item, str):
        return {"text": item}
    return item


def _normalize_feature(item: Union[str, Dict[str, Any]]) -> Dict[str, Any]:
    """Accept `"price"` or `{"name": "...", "description": "..."}`."""
    if isinstance(item, str):
        return {"name": item}
    return item


class AsyncWebCrawler:
    """
    Async client for Crawl4AI Cloud API.

    Mirrors the OSS AsyncWebCrawler API for seamless migration.
    Just change your import and add an API key.

    Example:
        ```python
        from crawl4ai_cloud import AsyncWebCrawler

        async with AsyncWebCrawler(api_key="sk_live_xxx") as crawler:
            # Single URL
            result = await crawler.run("https://example.com")
            print(result.markdown.raw_markdown)

            # Multiple URLs
            results = await crawler.run_many(urls, wait=True)

            # OSS compatibility aliases
            result = await crawler.arun(url)  # Same as run()
        ```
    """

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = "https://api.crawl4ai.com",
        timeout: float = 120.0,
        max_retries: int = 3,
        # OSS compatibility - these are ignored but accepted
        verbose: bool = False,
        **kwargs,
    ):
        """
        Initialize AsyncWebCrawler.

        Args:
            api_key: Your Crawl4AI API key (sk_live_* or sk_test_*).
                     If not provided, reads from CRAWL4AI_API_KEY env var.
            base_url: API base URL (default: https://api.crawl4ai.com)
            timeout: Request timeout in seconds (default: 120)
            max_retries: Max retry attempts for transient errors (default: 3)
            verbose: Ignored (OSS compatibility)
            **kwargs: Additional args ignored for OSS compatibility

        Raises:
            ValueError: If API key is missing or has invalid format
        """
        self._http = HTTPClient(
            api_key=api_key,
            base_url=base_url,
            timeout=timeout,
            max_retries=max_retries,
        )

    # -------------------------------------------------------------------------
    # Core Crawl Methods
    # -------------------------------------------------------------------------

    async def run(
        self,
        url: str,
        config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]] = None,
        browser_config: Optional[Union[BrowserConfig, Dict[str, Any]]] = None,
        strategy: str = "browser",
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
        **kwargs,
    ) -> CrawlResult:
        """
        Crawl a single URL.

        Args:
            url: URL to crawl
            config: CrawlerRunConfig instance or dict
            browser_config: BrowserConfig instance or dict
            strategy: "browser" (JS support) or "http" (faster, no JS)
            proxy: Proxy config - "datacenter", "residential", "auto",
                   or dict/ProxyConfig for full control
            bypass_cache: Skip cache and force fresh crawl
            **kwargs: Additional parameters passed to API

        Returns:
            CrawlResult with HTML, markdown, metadata, etc.

        Example:
            ```python
            result = await crawler.run("https://example.com")
            print(result.markdown.raw_markdown)

            # With config
            config = CrawlerRunConfig(screenshot=True)
            result = await crawler.run(url, config=config)
            ```
        """
        body = build_crawl_request(
            url=url,
            config=config,
            browser_config=browser_config,
            strategy=strategy,
            proxy=proxy,
            bypass_cache=bypass_cache,
            **kwargs,
        )

        data = await self._http.request("POST", "/v1/crawl", json=body, timeout=120)
        return CrawlResult.from_dict(data)

    async def arun(
        self,
        url: str,
        config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]] = None,
        **kwargs,
    ) -> CrawlResult:
        """
        Crawl a single URL (OSS compatibility alias for run()).

        This method exists for compatibility with OSS crawl4ai code.
        It simply calls run() with all arguments.
        """
        return await self.run(url, config=config, **kwargs)

    async def run_many(
        self,
        urls: List[str],
        config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]] = None,
        browser_config: Optional[Union[BrowserConfig, Dict[str, Any]]] = None,
        strategy: str = "browser",
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        priority: int = 5,
        webhook_url: Optional[str] = None,
        **kwargs,
    ) -> CrawlJob:
        """
        Crawl multiple URLs.

        Creates an async job for processing. Use wait=True to block until
        complete, or poll with get_job()/wait_job().

        Args:
            urls: List of URLs to crawl
            config: CrawlerRunConfig instance or dict
            browser_config: BrowserConfig instance or dict
            strategy: "browser" (JS support) or "http" (faster, no JS)
            proxy: Proxy configuration
            bypass_cache: Skip cache for all URLs
            wait: If True, poll until job completes
            poll_interval: Seconds between status polls (default: 2.0)
            timeout: Max seconds to wait (None = no timeout)
            priority: Job priority 1-10 (default: 5)
            webhook_url: URL to notify on completion
            **kwargs: Additional parameters passed to API

        Returns:
            CrawlJob with job ID and status

        Note:
            To get results after job completes, use download_url(job.id) to get
            a presigned URL for the ZIP file containing all crawl results.

        Example:
            ```python
            # Fire and forget
            job = await crawler.run_many(urls)
            print(f"Job {job.id} started")

            # Wait for completion, then download results
            job = await crawler.run_many(urls, wait=True)
            if job.is_complete:
                url = await crawler.download_url(job.id)
                print(f"Download results: {url}")
            ```
        """
        # Always use async endpoint for consistent job tracking
        return await self._run_async(
            urls=urls,
            config=config,
            browser_config=browser_config,
            strategy=strategy,
            proxy=proxy,
            bypass_cache=bypass_cache,
            wait=wait,
            poll_interval=poll_interval,
            timeout=timeout,
            priority=priority,
            webhook_url=webhook_url,
            **kwargs,
        )

    async def arun_many(
        self,
        urls: List[str],
        config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]] = None,
        **kwargs,
    ) -> CrawlJob:
        """
        Crawl multiple URLs (OSS compatibility alias for run_many()).

        This method exists for compatibility with OSS crawl4ai code.
        It simply calls run_many() with all arguments.
        """
        return await self.run_many(urls, config=config, **kwargs)

    async def _run_async(
        self,
        urls: List[str],
        config=None,
        browser_config=None,
        strategy: str = "browser",
        proxy=None,
        bypass_cache: bool = False,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        priority: int = 5,
        webhook_url: Optional[str] = None,
        **kwargs,
    ) -> CrawlJob:
        """Internal: Async crawl for >10 URLs."""
        body = build_crawl_request(
            urls=urls,
            config=config,
            browser_config=browser_config,
            strategy=strategy,
            proxy=proxy,
            bypass_cache=bypass_cache,
            priority=priority,
            **kwargs,
        )

        if webhook_url:
            body["webhook_url"] = webhook_url

        data = await self._http.request("POST", "/v1/crawl/async", json=body)
        job = CrawlJob.from_dict(data)

        if wait:
            job = await self.wait_job(
                job.id,
                poll_interval=poll_interval,
                timeout=timeout,
            )
            # Results are available via download_url() after job completes

        return job

    # -------------------------------------------------------------------------
    # Job Management
    # -------------------------------------------------------------------------

    async def get_job(self, job_id: str) -> CrawlJob:
        """
        Get job status.

        Args:
            job_id: Job ID to check

        Returns:
            CrawlJob with current status

        Note:
            To get results, use download_url() to get a presigned URL for the ZIP file.
        """
        data = await self._http.request("GET", f"/v1/crawl/jobs/{job_id}")
        return CrawlJob.from_dict(data)

    async def wait_job(
        self,
        job_id: str,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
    ) -> CrawlJob:
        """
        Poll until job completes.

        Works with both regular crawl jobs (job_xxx) and deep crawl scan jobs (scan_xxx).
        For scan jobs, automatically waits for the scan phase, then for the crawl phase
        if a crawl job was created.

        Args:
            job_id: Job ID to wait for (supports both job_xxx and scan_xxx formats)
            poll_interval: Seconds between polls (default: 2.0)
            timeout: Max seconds to wait (None = no timeout)

        Returns:
            CrawlJob with final status

        Raises:
            TimeoutError: If timeout exceeded

        Note:
            To get results after job completes, use download_url() to get a presigned
            URL for the ZIP file containing all crawl results.
        """
        start_time = time.time()

        # Handle scan jobs (from deep_crawl with wait=False)
        if job_id.startswith("scan_"):
            # Wait for scan phase to complete
            scan_result = await self._wait_scan_job(
                job_id,
                poll_interval=poll_interval,
                timeout=timeout,
            )

            # If scan created a crawl job, wait for that too
            if scan_result.crawl_job_id:
                remaining_timeout = None
                if timeout:
                    elapsed = time.time() - start_time
                    remaining_timeout = max(0, timeout - elapsed)

                return await self.wait_job(
                    scan_result.crawl_job_id,
                    poll_interval=poll_interval,
                    timeout=remaining_timeout,
                )

            # Scan only mode (no crawl job) - return CrawlJob-like response
            return CrawlJob(
                job_id=job_id,
                status=scan_result.status,
                progress=JobProgress(
                    total=scan_result.discovered_count,
                    completed=scan_result.discovered_count,
                    failed=0,
                ),
                urls_count=scan_result.discovered_count,
                created_at=scan_result.created_at,
                results=None,
            )

        # Regular crawl job polling
        while True:
            job = await self.get_job(job_id)

            if job.is_complete:
                return job

            if timeout and (time.time() - start_time) > timeout:
                raise TimeoutError(
                    f"Timeout waiting for job {job_id}. "
                    f"Status: {job.status}, Progress: {job.progress_percent:.1f}%"
                )

            await asyncio.sleep(poll_interval)

    async def list_jobs(
        self,
        status: Optional[str] = None,
        limit: int = 20,
        offset: int = 0,
    ) -> List[CrawlJob]:
        """
        List jobs with optional filtering.

        Args:
            status: Filter by status (pending, running, completed, failed)
            limit: Max jobs to return (default: 20, max: 100)
            offset: Pagination offset

        Returns:
            List of CrawlJob summaries
        """
        params: Dict[str, Any] = {"limit": limit, "offset": offset}
        if status:
            params["status"] = status

        data = await self._http.request("GET", "/v1/crawl/jobs", params=params)

        jobs = []
        for j in data.get("jobs", []):
            jobs.append(CrawlJob.from_dict(j))
        return jobs

    async def cancel_job(self, job_id: str) -> bool:
        """
        Cancel a pending or running job.

        Args:
            job_id: Job ID to cancel

        Returns:
            True if cancelled successfully
        """
        await self._http.request("DELETE", f"/v1/crawl/jobs/{job_id}")
        return True

    async def download_url(
        self,
        job_id: str,
        expires_in: int = 3600,
    ) -> str:
        """
        Get presigned URL for downloading job results.

        Args:
            job_id: Job ID
            expires_in: URL expiration in seconds (default: 3600)

        Returns:
            Presigned S3 download URL
        """
        params = {"expires_in": expires_in}
        data = await self._http.request(
            "GET", f"/v1/crawl/jobs/{job_id}/download", params=params
        )
        return data["download_url"]

    # -------------------------------------------------------------------------
    # Deep Crawl
    # -------------------------------------------------------------------------

    async def deep_crawl(
        self,
        url: Optional[str] = None,
        source_job: Optional[str] = None,
        strategy: str = "bfs",
        max_depth: int = 3,
        max_urls: int = 100,
        scan_only: bool = False,
        config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]] = None,
        browser_config: Optional[Union[BrowserConfig, Dict[str, Any]]] = None,
        crawl_strategy: str = "auto",
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        filters: Optional[Dict[str, Any]] = None,
        scorers: Optional[Dict[str, Any]] = None,
        include_html: bool = False,
        webhook_url: Optional[str] = None,
        priority: int = 5,
        # Map strategy options
        source: str = "sitemap",
        pattern: str = "*",
        query: Optional[str] = None,
        score_threshold: Optional[float] = None,
        # URL filtering shortcuts
        include_patterns: Optional[List[str]] = None,
        exclude_patterns: Optional[List[str]] = None,
    ) -> Union[DeepCrawlResult, CrawlJob]:
        """
        Deep crawl - discover and crawl URLs from a starting point.

        Strategies:
        - "bfs": Breadth-first tree traversal
        - "dfs": Depth-first tree traversal
        - "best_first": Priority-based with scoring
        - "map": Sitemap-based discovery

        Args:
            url: Starting URL for discovery
            source_job: Previous scan job ID (for phase 2)
            strategy: "bfs", "dfs", "best_first", or "map"
            max_depth: Maximum traversal depth (1-10)
            max_urls: Maximum URLs to crawl (plan limit applied)
            scan_only: Return discovered URLs without extraction
            config: CrawlerRunConfig for extraction
            browser_config: BrowserConfig
            crawl_strategy: "browser", "http", or "auto"
            proxy: Proxy configuration (sticky_session recommended)
            bypass_cache: Skip cache for all URLs
            wait: Poll until complete
            poll_interval: Seconds between polls
            timeout: Max seconds to wait
            filters: URL filtering {"patterns": [...], "allowed_domains": [...]}
            scorers: URL scoring {"keywords": [...], "optimal_depth": N}
            include_html: Upload scanned HTML as ZIP to S3
            webhook_url: Notification URL on completion
            priority: Job priority 1-10
            source: URL source for map strategy ("sitemap", "cc", "sitemap+cc")
            pattern: Glob pattern for URL filtering
            query: BM25 relevance query
            score_threshold: Minimum BM25 score
            include_patterns: List of URL patterns to include (shortcut for filters)
            exclude_patterns: List of URL patterns to exclude (shortcut for filters)

        Returns:
            - If wait=False: DeepCrawlResult with job_id and discovered URLs
            - If wait=True and scan_only: DeepCrawlResult with all URLs
            - If wait=True: CrawlJob with extraction results

        Example:
            ```python
            # BFS crawl with extraction
            results = await crawler.deep_crawl(
                "https://docs.example.com",
                strategy="bfs",
                max_depth=2,
                max_urls=50,
                wait=True,
            )

            # With URL filtering
            results = await crawler.deep_crawl(
                "https://docs.example.com",
                include_patterns=["docs", "api"],
                exclude_patterns=["download", "search"],
                wait=True,
            )
            ```
        """
        if not url and not source_job:
            raise ValueError("Must provide either 'url' or 'source_job'")
        if url and source_job:
            raise ValueError("Provide either 'url' or 'source_job', not both")

        warnings.warn(
            "crawler.deep_crawl() targets the deprecated /v1/crawl/deep endpoint. "
            "Migrate to crawler.scan(scan={'mode': 'deep'}) for URL discovery, "
            "then pipe to scrape_many() / extract_many(). Will be removed in 0.8.0.",
            DeprecationWarning, stacklevel=2,
        )

        # Build request body
        body: Dict[str, Any] = {}

        if source_job:
            # Phase 2: extraction from cached HTML — only send source_job_id
            body["source_job_id"] = source_job
        else:
            # Phase 1: URL-based discovery — include scan parameters
            body["url"] = url
            body["strategy"] = strategy
            body["crawl_strategy"] = crawl_strategy
            body["priority"] = priority

            # Tree strategy options
            if strategy in ("bfs", "dfs", "best_first"):
                body["max_depth"] = max_depth
                body["max_urls"] = max_urls

                # Build filters from include_patterns/exclude_patterns or use provided filters
                effective_filters = filters.copy() if filters else {}
                if include_patterns:
                    effective_filters["include_patterns"] = include_patterns
                if exclude_patterns:
                    effective_filters["exclude_patterns"] = exclude_patterns
                if effective_filters:
                    body["filters"] = effective_filters

                if scorers:
                    body["scorers"] = scorers
                if scan_only:
                    body["scan_only"] = True
                if include_html:
                    body["include_html"] = True

            # Map strategy options
            if strategy == "map":
                seeding_config: Dict[str, Any] = {
                    "source": source,
                    "pattern": pattern,
                }
                if max_urls:
                    seeding_config["max_urls"] = max_urls
                if query:
                    seeding_config["query"] = query
                if score_threshold is not None:
                    seeding_config["score_threshold"] = score_threshold
                body["seeding_config"] = seeding_config

        # Shared parameters (apply to both URL crawl and source_job extraction)
        crawler_config = sanitize_crawler_config(config)
        if crawler_config:
            body["crawler_config"] = crawler_config

        browser_cfg = sanitize_browser_config(browser_config, crawl_strategy)
        if browser_cfg:
            body["browser_config"] = browser_cfg

        # Proxy
        proxy_config = normalize_proxy(proxy)
        if proxy_config:
            body["proxy"] = proxy_config

        if bypass_cache:
            body["bypass_cache"] = True
        if webhook_url:
            body["webhook_url"] = webhook_url

        data = await self._http.request("POST", "/v1/crawl/deep", json=body, timeout=120)
        result = DeepCrawlResult.from_dict(data)

        if not wait:
            return result

        # Wait for scan to complete
        scan_result = await self._wait_scan_job(
            result.job_id,
            poll_interval=poll_interval,
            timeout=timeout,
        )

        if scan_only:
            return scan_result

        if not scan_result.has_urls:
            return scan_result

        # If crawl job was created, wait for it
        if scan_result.crawl_job_id:
            return await self.wait_job(
                scan_result.crawl_job_id,
                poll_interval=poll_interval,
                timeout=timeout,
            )

        return scan_result

    async def _wait_scan_job(
        self,
        job_id: str,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
    ) -> DeepCrawlResult:
        """Wait for scan job to complete."""
        start_time = time.time()

        while True:
            data = await self._http.request("GET", f"/v1/crawl/deep/jobs/{job_id}")
            result = DeepCrawlResult.from_dict(data)

            if result.is_complete:
                return result

            if timeout and (time.time() - start_time) > timeout:
                raise TimeoutError(
                    f"Timeout waiting for scan job {job_id}. "
                    f"Status: {result.status}, Discovered: {result.discovered_count}"
                )

            await asyncio.sleep(poll_interval)

    async def cancel_deep_crawl(self, job_id: str) -> DeepCrawlResult:
        """
        Cancel a running deep crawl job.

        The crawl will stop at the next batch boundary, preserving any
        partial results that have been collected so far.

        Args:
            job_id: Deep crawl job ID (scan_xxx format)

        Returns:
            DeepCrawlResult with status "cancelled" and partial results

        Example:
            ```python
            # Start deep crawl without waiting
            result = await crawler.deep_crawl(
                "https://docs.example.com",
                max_urls=500,
                wait=False,
            )

            # Cancel after some time
            await asyncio.sleep(10)
            cancelled = await crawler.cancel_deep_crawl(result.job_id)
            print(f"Cancelled with {cancelled.discovered_count} partial results")
            ```
        """
        data = await self._http.request("POST", f"/v1/crawl/deep/jobs/{job_id}/cancel")
        return DeepCrawlResult.from_dict(data)

    async def get_deep_crawl_status(self, job_id: str) -> DeepCrawlResult:
        """
        Get the status of a deep crawl job.

        Args:
            job_id: Deep crawl job ID (scan_xxx format)

        Returns:
            DeepCrawlResult with current status and discovered URLs
        """
        data = await self._http.request("GET", f"/v1/crawl/deep/jobs/{job_id}")
        return DeepCrawlResult.from_dict(data)

    # -------------------------------------------------------------------------
    # Scan API (URL Discovery)
    # -------------------------------------------------------------------------

    async def scan(
        self,
        url: str,
        criteria: Optional[str] = None,
        scan: Optional[Union[Dict[str, Any], "SiteScanConfig"]] = None,
        sources: str = "primary",
        mode: Optional[str] = None,
        max_urls: Optional[int] = None,
        include_subdomains: bool = True,
        extract_head: bool = True,
        soft_404_detection: bool = True,
        query: Optional[str] = None,
        score_threshold: Optional[float] = None,
        force: bool = False,
        probe_threshold: int = 10,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
    ) -> "ScanResult":
        """
        Discover all URLs under a domain. AI-assisted via `criteria`, with
        optional async deep-mode traversal.

        Two routing strategies (picked by `scan.mode`):
        - **map** (sync): DomainMapper — sitemap + CC + wayback etc. Returns
          URLs inline. 2-60s. Cached 7 days.
        - **deep** (async): best-first tree traversal. Returns a `job_id`;
          poll with `get_scan_job()` or pass `wait=True`.

        Costs 1 credit per scan, flat.

        Args:
            url: Starting URL (e.g., "https://example.com")
            criteria: Plain-English description of what to discover. When set,
                the backend LLM generates a unified scan config (mode,
                patterns, filters, scorers, query, threshold). Explicit
                overrides in `scan=` still win.
            scan: Explicit scan overrides — dict or `SiteScanConfig`. Merged
                on top of LLM output. Fields: `mode` ("auto"|"map"|"deep"),
                `patterns`, `filters`, `scorers`, `query`, `score_threshold`,
                `include_subdomains`, `max_depth`.
            mode: DomainMapper source depth: "default" (fast, adaptive) or
                "deep" (all historical sources). Only applies when the final
                routing is map mode. **This is distinct from `scan.mode`.**
            max_urls: Maximum URLs to return. Plan limit applied server-side.
            include_subdomains: Discover subdomains. Default: True.
            extract_head: Fetch <head> metadata per URL. Default: True.
            soft_404_detection: Filter SPA false positives. Default: True.
            query: Top-level BM25 relevance query.
            score_threshold: Minimum relevance score (0.0-1.0).
            force: Bypass 7-day cache and force fresh scan.
            probe_threshold: In default source mode, auto-probe if fewer
                URLs found. 0 to disable.
            wait: If True and deep mode is picked, poll until completion.
            poll_interval: Seconds between polls when waiting. Default: 2.0.
            timeout: Max seconds to wait. Raises TimeoutError if exceeded.

        Returns:
            ScanResult with discovered URLs + `mode_used` + `generated_config`
            (when criteria was set). When deep mode kicks off, `is_async`
            will be True and `job_id` will be set; the URLs list is empty
            until you poll.

        Example:
            ```python
            # AI-assisted (map mode picked by LLM)
            result = await crawler.scan(
                "https://docs.crawl4ai.com",
                criteria="API reference pages",
                max_urls=50,
            )
            print(f"Mode: {result.mode_used}, found {result.total_urls} URLs")
            if result.generated_config:
                print(f"AI reasoning: {result.generated_config.reasoning}")

            # Explicit deep scan with waiting
            result = await crawler.scan(
                "https://directory.example.com",
                criteria="company profile pages",
                scan={"mode": "deep", "max_depth": 3},
                wait=True,
                poll_interval=3.0,
            )
            ```
        """
        from crawl4ai_cloud.models import ScanResult, SiteScanConfig

        # Legacy mode= → sources= translation. mode is deprecated.
        if mode is not None:
            warnings.warn(
                "scan(mode=…) is deprecated — use sources= ('primary' | 'extended'). "
                "Will be removed in 0.8.0.",
                DeprecationWarning, stacklevel=2,
            )
            sources = "extended" if mode == "deep" else "primary"

        body: Dict[str, Any] = {"url": url, "sources": sources}
        if criteria:
            body["criteria"] = criteria
        if scan is not None:
            if isinstance(scan, SiteScanConfig):
                body["scan"] = scan.to_dict()
            elif isinstance(scan, dict):
                body["scan"] = scan
            else:
                raise TypeError(
                    f"scan must be dict or SiteScanConfig, got {type(scan).__name__}"
                )
        if max_urls is not None:
            body["max_urls"] = max_urls
        body["include_subdomains"] = include_subdomains
        body["extract_head"] = extract_head
        body["soft_404_detection"] = soft_404_detection
        if query:
            body["query"] = query
        if score_threshold is not None:
            body["score_threshold"] = score_threshold
        body["force"] = force
        if probe_threshold != 10:
            body["probe_threshold"] = probe_threshold

        data = await self._http.request("POST", "/v1/scan", json=body, timeout=180)
        result = ScanResult.from_dict(data)

        # If the LLM picked deep mode (or the caller forced it), optionally
        # block until the scan job finishes.
        if wait and result.is_async:
            final = await self._wait_scan_job_v2(
                result.job_id, poll_interval, timeout,
            )
            # Transfer polled state back onto the ScanResult so callers get
            # one object they can inspect.
            result.status = final.status
            result.total_urls = final.total_urls
            result.urls = final.urls
            result.duration_ms = final.duration_ms
            if final.error:
                result.error = final.error

        return result

    async def get_scan_job(self, job_id: str) -> "ScanJobStatus":
        """
        Poll a deep scan job started via `scan(..., scan={"mode": "deep"})`.

        Returns:
            ScanJobStatus with current discovered URLs, progress, and status.
            URLs are appended as they're discovered.

        Example:
            ```python
            scan_result = await crawler.scan(
                "https://directory.example.com",
                criteria="company profile pages",
                scan={"mode": "deep"},
            )

            while True:
                job = await crawler.get_scan_job(scan_result.job_id)
                print(f"Status: {job.status}, found {job.total_urls}")
                if job.is_complete:
                    break
                await asyncio.sleep(3)
            ```
        """
        from crawl4ai_cloud.models import ScanJobStatus
        data = await self._http.request("GET", f"/v1/scan/jobs/{job_id}")
        return ScanJobStatus.from_dict(data)

    async def cancel_scan_job(self, job_id: str) -> "ScanJobStatus":
        """
        Cancel a running deep scan. Cancellation happens at the next batch
        boundary — partial results (URLs discovered so far) are preserved.

        Args:
            job_id: Scan job id returned from a deep-mode `scan()` call.

        Returns:
            ScanJobStatus with status="cancelled" and any partial results.
        """
        from crawl4ai_cloud.models import ScanJobStatus
        data = await self._http.request(
            "POST", f"/v1/scan/jobs/{job_id}/cancel"
        )
        return ScanJobStatus.from_dict(data)

    async def _wait_scan_job_v2(
        self,
        job_id: str,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
    ) -> "ScanJobStatus":
        """Poll /v1/scan/jobs/{id} until the deep scan finishes."""
        from crawl4ai_cloud.models import ScanJobStatus
        start = time.time()
        while True:
            job = await self.get_scan_job(job_id)
            if job.is_complete:
                return job
            if timeout and (time.time() - start) > timeout:
                raise TimeoutError(
                    f"Timeout waiting for scan job {job_id}. "
                    f"Status: {job.status}, found: {job.total_urls}"
                )
            await asyncio.sleep(poll_interval)

    # -------------------------------------------------------------------------
    # Context API
    # -------------------------------------------------------------------------

    async def context(
        self,
        query: str,
        paa_limit: int = 3,
        results_per_paa: int = 5,
        wait: bool = True,
    ) -> ContextResult:
        """
        Build context from a search query.

        Expands query using Google's "People Also Ask" questions and
        crawls results to build comprehensive context stored in S3.

        Args:
            query: Search query to expand
            paa_limit: Number of PAA questions to expand (1-10)
            results_per_paa: Results per PAA search (1-20)
            wait: If True, wait for completion (default: True)

        Returns:
            ContextResult with download_url for retrieving results

        Example:
            ```python
            result = await crawler.context("best database for SaaS")
            print(f"Crawled {result.urls_crawled} URLs")
            print(f"Download: {result.download_url}")
            ```
        """
        body: Dict[str, Any] = {
            "query": query,
            "strategy": "serper_paa",
            "paa_limit": paa_limit,
            "results_per_paa": results_per_paa,
        }

        data = await self._http.request("POST", "/v1/context", json=body, timeout=300)
        return ContextResult.from_dict(data)

    # -------------------------------------------------------------------------
    # Schema Generation
    # -------------------------------------------------------------------------

    async def generate_schema(
        self,
        html: Optional[Union[str, List[str]]] = None,
        urls: Optional[List[str]] = None,
        query: Optional[str] = None,
        schema_type: str = "CSS",
        target_json_example: Optional[Dict[str, Any]] = None,
        llm_config: Optional[Dict[str, Any]] = None,
    ) -> GeneratedSchema:
        """
        Generate extraction schema from HTML using LLM.

        Creates CSS or XPath selectors that can be used with
        JsonCssExtractionStrategy for fast, LLM-free extractions.

        Supports three modes:
        - Single HTML: Pass a single HTML string
        - Multiple HTML: Pass a list of HTML strings for robust selectors
        - From URLs: Pass a list of URLs (max 3) to fetch HTML from

        Args:
            html: HTML content to analyze. Can be a single string or list
                  of strings for multi-sample generation. Required if urls
                  not provided.
            urls: List of URLs to fetch HTML from (max 3). The API fetches
                  these in parallel via workers. Required if html not provided.
            query: Natural language description of what to extract
            schema_type: "CSS" (default) or "XPATH"
            target_json_example: Example of desired output structure
            llm_config: LLM configuration (optional)

        Returns:
            GeneratedSchema with selectors or error

        Raises:
            ValueError: If neither html nor urls provided, or if both are provided

        Example:
            ```python
            # Single HTML
            schema = await crawler.generate_schema(html=page.html, query="Extract products")

            # Multiple HTML samples
            schema = await crawler.generate_schema(
                html=[page1.html, page2.html],
                query="Extract products from these samples"
            )

            # From URLs (max 3)
            schema = await crawler.generate_schema(
                urls=["https://example.com/p/1", "https://example.com/p/2"],
                query="Extract product details"
            )
            ```
        """
        if not html and not urls:
            raise ValueError("Either 'html' or 'urls' must be provided")
        if html and urls:
            raise ValueError("Provide either 'html' or 'urls', not both")

        body: Dict[str, Any] = {"schema_type": schema_type}

        if html is not None:
            body["html"] = html
        if urls is not None:
            if len(urls) > 3:
                raise ValueError("Maximum 3 URLs allowed")
            body["urls"] = urls
        if query:
            body["query"] = query
        if target_json_example:
            body["target_json_example"] = target_json_example
        if llm_config:
            body["llm_config"] = llm_config

        data = await self._http.request("POST", "/v1/schema/generate", json=body, timeout=60)
        return GeneratedSchema.from_dict(data)

    # -------------------------------------------------------------------------
    # Storage & Health
    # -------------------------------------------------------------------------

    async def storage(self) -> StorageUsage:
        """
        Get current storage usage.

        Returns:
            StorageUsage with used/max/remaining MB
        """
        data = await self._http.request("GET", "/v1/crawl/storage")
        return StorageUsage.from_dict(data)

    async def health(self) -> Dict[str, str]:
        """Check API health status."""
        return await self._http.request("GET", "/health")

    # =========================================================================
    # Wrapper API -- Simplified endpoints
    # =========================================================================

    async def scrape(
        self,
        url: str,
        strategy: str = "browser",
        fit: bool = True,
        include: Optional[List[str]] = None,
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
    ) -> MarkdownResponse:
        """Fetch a page, return clean markdown plus optional extras.

        ``POST /v1/scrape`` (sync, single URL). Use :meth:`scrape_many` for
        batch / async / webhooks.

        Args:
            url: URL to fetch.
            strategy: ``"browser"`` (JS-capable, default) or ``"http"`` (3-5× faster, no JS).
            fit: Apply :class:`PruningContentFilter` for nav-stripped markdown (default True).
                When ``True``, response also includes ``fit_markdown`` and ``fit_html``.
            include: Opt back into ``["links", "media", "metadata", "tables"]``.
                Default trimmer drops these to keep responses lean.
            crawler_config: ``CrawlerRunConfig`` overrides (same fields as ``/v1/crawl``).
            browser_config: ``BrowserConfig`` overrides (headers, cookies, ``profile_id``…).
            proxy: ``ProxyConfig``, dict, or alias string. Omit to disable.
            bypass_cache: Skip the page cache and force a fresh fetch.

        Returns:
            :class:`MarkdownResponse` with ``markdown``, ``fit_markdown`` (when
            ``fit=True``), and any opted-in extras.

        Example:
            >>> r = await crawler.scrape("https://example.com", include=["links", "metadata"])
            >>> r.markdown[:80]
        """
        body: Dict[str, Any] = {"url": url, "strategy": strategy, "fit": fit}
        if include:
            body["include"] = include
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if bypass_cache:
            body["bypass_cache"] = True

        data = await self._http.request("POST", "/v1/scrape", json=body)
        return MarkdownResponse.from_dict(data)

    async def markdown(self, *args, **kwargs) -> MarkdownResponse:
        """DEPRECATED — use :meth:`scrape`. Same shape, same response.

        ``/v1/markdown`` was renamed to ``/v1/scrape``. The SDK method
        ``markdown()`` is kept as a back-compat alias for one release and
        will be removed in 0.8.0.
        """
        warnings.warn(
            "crawler.markdown() is deprecated — use crawler.scrape(). "
            "Will be removed in 0.8.0.",
            DeprecationWarning, stacklevel=2,
        )
        return await self.scrape(*args, **kwargs)

    async def screenshot(
        self,
        url: str,
        full_page: bool = True,
        pdf: bool = False,
        wait_for: Optional[str] = None,
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
    ) -> ScreenshotResponse:
        """
        Capture a screenshot or PDF of a web page.

        Args:
            url: URL to screenshot
            full_page: Capture full scrollable page (default True)
            pdf: Generate PDF in addition to screenshot
            wait_for: CSS selector or seconds to wait before capture
            crawler_config: CrawlerRunConfig overrides
            browser_config: BrowserConfig overrides
            proxy: Proxy configuration
            bypass_cache: Skip cache

        Returns:
            ScreenshotResponse with base64 screenshot and/or pdf
        """
        body: Dict[str, Any] = {"url": url, "full_page": full_page}
        if pdf:
            body["pdf"] = True
        if wait_for:
            body["wait_for"] = wait_for
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if bypass_cache:
            body["bypass_cache"] = True

        data = await self._http.request("POST", "/v1/screenshot", json=body, timeout=120)
        return ScreenshotResponse.from_dict(data)

    async def extract(
        self,
        url: str,
        query: Optional[str] = None,
        json_example: Optional[Dict[str, Any]] = None,
        schema: Optional[Dict[str, Any]] = None,
        method: str = "auto",
        strategy: str = "http",
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        llm_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
    ) -> ExtractResponse:
        """
        Extract structured data from a web page.

        Args:
            url: URL to extract from
            query: What to extract (e.g. "get all products with title and price")
            json_example: Example of desired output shape
            schema: Pre-built CSS schema for deterministic extraction
            method: "auto" (default), "llm", or "schema"
            strategy: "http" (default, faster) or "browser" (JS support)
            crawler_config: CrawlerRunConfig overrides
            browser_config: BrowserConfig overrides
            llm_config: LLM provider config for BYOK
            proxy: Proxy configuration
            bypass_cache: Skip cache

        Returns:
            ExtractResponse with extracted data, method_used, schema_used
        """
        body: Dict[str, Any] = {"url": url, "method": method, "strategy": strategy}
        if query:
            body["query"] = query
        if json_example:
            body["json_example"] = json_example
        if schema:
            body["schema"] = schema
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if llm_config:
            body["llm_config"] = llm_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if bypass_cache:
            body["bypass_cache"] = True

        data = await self._http.request("POST", "/v1/extract", json=body, timeout=180)
        return ExtractResponse.from_dict(data)

    async def map(
        self,
        url: str,
        sources: str = "primary",
        mode: Optional[str] = None,
        max_urls: Optional[int] = None,
        include_subdomains: bool = False,
        extract_head: bool = True,
        query: Optional[str] = None,
        score_threshold: Optional[float] = None,
        force: bool = False,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
    ) -> MapResponse:
        """Discover all URLs on a domain via DomainMapper.

        Args:
            url: Domain URL (e.g. ``"https://example.com"``).
            sources: ``"primary"`` (sitemap+homepage+robots+RSS, ~2-15s) or
                ``"extended"`` (adds Wayback+Common Crawl+Cert Transparency, ~30-60s).
                Only flip to extended when primary returns too few URLs.
            mode: DEPRECATED — use ``sources``. ``"default"`` → ``"primary"``,
                ``"deep"`` → ``"extended"``.
            max_urls: Cap on returned URLs.
            include_subdomains: Discover and include subdomains.
            extract_head: Fetch ``<head>`` metadata (title, description, og:*)
                for each URL. Required for BM25 scoring (``query`` + ``score_threshold``).
            query: BM25 relevance query.
            score_threshold: 0.0-1.0; requires ``query``.
            force: Bypass the 7-day archive cache.
            proxy: ``ProxyConfig``, dict, or alias.

        Returns:
            :class:`MapResponse` with ``urls``, ``total_urls``, ``hosts_found``.
        """
        if mode is not None:
            warnings.warn(
                "map(mode=…) is deprecated — use sources= ('primary' | 'extended'). "
                "Will be removed in 0.8.0.",
                DeprecationWarning, stacklevel=2,
            )
            sources = "extended" if mode == "deep" else "primary"

        body: Dict[str, Any] = {"url": url, "sources": sources}
        if max_urls is not None:
            body["max_urls"] = max_urls
        body["include_subdomains"] = include_subdomains
        body["extract_head"] = extract_head
        if query:
            body["query"] = query
        if score_threshold is not None:
            body["score_threshold"] = score_threshold
        if force:
            body["force"] = True
        if proxy:
            body["proxy"] = normalize_proxy(proxy)

        data = await self._http.request("POST", "/v1/map", json=body, timeout=120)
        return MapResponse.from_dict(data)

    # ---- Async batch methods ----

    async def scrape_many(
        self,
        urls: List[str],
        strategy: str = "browser",
        fit: bool = True,
        include: Optional[List[str]] = None,
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        webhook_url: Optional[str] = None,
        priority: int = 5,
    ) -> WrapperJob:
        """Submit an async scrape job over a list of URLs.

        ``POST /v1/scrape/async``. Returns a :class:`WrapperJob` immediately;
        pass ``wait=True`` to poll until terminal. Use :meth:`scrape` for the
        single-URL sync path.

        Args:
            urls: Up to 100 URLs.
            strategy, fit, include, crawler_config, browser_config, proxy, bypass_cache:
                See :meth:`scrape`.
            wait: Block until the job is ``completed`` / ``failed`` / ``cancelled``.
            poll_interval: Seconds between polls when ``wait=True``.
            timeout: Max seconds to wait when ``wait=True`` (raises :class:`TimeoutError`).
            webhook_url: POSTed when the job completes.
            priority: 1-10 (5 default).

        Returns:
            :class:`WrapperJob` with ``job_id``, ``status``, ``urls_count``.
        """
        body: Dict[str, Any] = {"urls": urls, "strategy": strategy, "fit": fit}
        if include:
            body["include"] = include
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if bypass_cache:
            body["bypass_cache"] = True
        if webhook_url:
            body["webhook_url"] = webhook_url
        body["priority"] = priority

        data = await self._http.request("POST", "/v1/scrape/async", json=body)
        job = WrapperJob.from_dict(data)
        if wait:
            job = await self._wait_wrapper_job(job.job_id, "markdown", poll_interval, timeout)
        return job

    async def markdown_many(self, *args, **kwargs) -> WrapperJob:
        """DEPRECATED — use :meth:`scrape_many`. Same shape, same response."""
        warnings.warn(
            "crawler.markdown_many() is deprecated — use crawler.scrape_many(). "
            "Will be removed in 0.8.0.",
            DeprecationWarning, stacklevel=2,
        )
        return await self.scrape_many(*args, **kwargs)

    async def screenshot_many(
        self,
        urls: List[str],
        full_page: bool = True,
        pdf: bool = False,
        wait_for: Optional[str] = None,
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        webhook_url: Optional[str] = None,
        priority: int = 5,
    ) -> WrapperJob:
        """Create an async screenshot job for multiple URLs."""
        body: Dict[str, Any] = {"urls": urls, "full_page": full_page}
        if pdf:
            body["pdf"] = True
        if wait_for:
            body["wait_for"] = wait_for
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if bypass_cache:
            body["bypass_cache"] = True
        if webhook_url:
            body["webhook_url"] = webhook_url
        body["priority"] = priority

        data = await self._http.request("POST", "/v1/screenshot/async", json=body)
        job = WrapperJob.from_dict(data)
        if wait:
            job = await self._wait_wrapper_job(job.job_id, "screenshot", poll_interval, timeout)
        return job

    async def extract_many(
        self,
        url: str,
        extra_urls: Optional[List[str]] = None,
        method: str = "auto",
        query: Optional[str] = None,
        json_example: Optional[Dict[str, Any]] = None,
        schema: Optional[Dict[str, Any]] = None,
        strategy: str = "http",
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        llm_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        bypass_cache: bool = False,
        wait: bool = False,
        poll_interval: float = 2.0,
        timeout: Optional[float] = None,
        webhook_url: Optional[str] = None,
        priority: int = 5,
    ) -> WrapperJob:
        """Submit an async extract job over one base URL plus optional followers.

        ``POST /v1/extract/async``. The base ``url`` is the schema **template**
        in css_schema mode — the server samples it, generates a schema once,
        then re-applies that schema across every entry in ``extra_urls`` for
        free (no extra LLM calls per URL). In ``method="llm"`` mode the base
        has no special role; every URL gets its own LLM call.

        Args:
            url: Base URL (required). Up to 100 URLs total (1 base + 99 extras).
            extra_urls: Follower URLs that share the resolved strategy.
            method: ``"auto"`` (default), ``"schema"``, or ``"llm"``. AUTO works
                for batch as of API v2.2 — the previous "AUTO not allowed for
                batch" restriction was removed.
            query, json_example, schema: shape hints — see :meth:`extract`.
            strategy, crawler_config, browser_config, llm_config, proxy, bypass_cache:
                Standard request fields.
            wait: Block until terminal status.
            poll_interval, timeout, webhook_url, priority: Async controls.

        Returns:
            :class:`WrapperJob`. Poll the job's ``results[]`` for inline per-URL
            extraction records (sync-shaped: ``{url, success, data, method_used,
            schema_used, query_used, duration_ms, error_message}``).
        """
        body: Dict[str, Any] = {"url": url, "method": method, "strategy": strategy}
        if extra_urls:
            body["extra_urls"] = extra_urls
        if query:
            body["query"] = query
        if json_example:
            body["json_example"] = json_example
        if schema:
            body["schema"] = schema
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if llm_config:
            body["llm_config"] = llm_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if bypass_cache:
            body["bypass_cache"] = True
        if webhook_url:
            body["webhook_url"] = webhook_url
        body["priority"] = priority

        data = await self._http.request("POST", "/v1/extract/async", json=body)
        job = WrapperJob.from_dict(data)
        if wait:
            job = await self._wait_wrapper_job(job.job_id, "extract", poll_interval, timeout)
        return job

    # ---- Site crawl (always async) ----

    async def crawl_site(
        self,
        url: str,
        max_pages: int = 20,
        criteria: Optional[str] = None,
        scan: Optional[Union[Dict[str, Any], "SiteScanConfig"]] = None,
        extract: Optional[Union[Dict[str, Any], "SiteExtractConfig"]] = None,
        include: Optional[List[str]] = None,
        include_markdown: Optional[bool] = None,
        strategy: str = "browser",
        fit: bool = True,
        discovery: str = "map",
        pattern: Optional[str] = None,
        max_depth: Optional[int] = None,
        crawler_config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
        webhook_url: Optional[str] = None,
        priority: int = 5,
        wait: bool = False,
        poll_interval: float = 5.0,
        timeout: Optional[float] = None,
    ) -> SiteCrawlResponse:
        """
        Crawl an entire website — AI-assisted discovery + optional extraction.
        Always async.

        The flagship flow: pass a plain-English `criteria` and let the LLM
        pick the scan strategy (mode, patterns, query), generate URL filters,
        and (optionally) build an extraction schema from a sample URL. Poll
        one unified endpoint for both scan and crawl phases.

        Args:
            url: Starting URL.
            max_pages: Maximum pages to crawl (1-1000). Default: 20.
            criteria: Plain-English description — triggers the LLM config
                generator. Example: "all book listing pages".
            scan: Explicit scan overrides (dict or SiteScanConfig). Merged on
                top of LLM output. Set `scan.mode="auto"` to still let the
                LLM pick the routing mode.
            extract: Structured extraction config (dict or SiteExtractConfig).
                Fields: `query`, `json_example`, `method` ("auto"|"llm"|
                "schema"), `schema` (pre-built CSS schema), `sample_url`,
                `url_pattern`. When set, a schema is generated once from the
                sample URL and applied to every discovered page.
            include: Response fields to keep: "markdown", "links", "media",
                "metadata", "tables", "response_headers". Default includes
                markdown. **Drop "markdown" from the list to strip it from
                every result (extract-only mode).**
            include_markdown: Legacy flag. `False` is equivalent to dropping
                "markdown" from `include`.
            strategy: Per-page crawl strategy: "browser" or "http".
            fit: Apply content pruning for cleaner markdown. Default: True.
            discovery: Legacy — prefer `scan.mode`. Keeps backward compat
                with pre-AI clients.
            pattern: Legacy glob filter — prefer `scan.patterns`.
            max_depth: Legacy — prefer `scan.max_depth`.
            crawler_config: Raw CrawlerRunConfig overrides (power user).
            browser_config: Raw BrowserConfig overrides (power user).
            proxy: Proxy config. Omit for no proxy.
            webhook_url: POST target for job-completion callback.
            priority: Job priority (1-10, 1 = highest). Default: 5.
            wait: If True, poll until the crawl finishes.
            poll_interval: Seconds between polls when waiting. Default: 5.0.
            timeout: Max seconds to wait.

        Returns:
            SiteCrawlResponse with `job_id`, `generated_config` (when
            criteria was set), `extraction_method_used`, and `schema_used`
            (when extract was set). Poll progress with
            `get_site_crawl_job(job_id)`.

        Example:
            ```python
            # Flagship AI-assisted flow
            job = await crawler.crawl_site(
                "https://books.toscrape.com",
                criteria="all book listing pages",
                max_pages=50,
                strategy="http",
                extract={
                    "query": "book title, price, rating",
                    "json_example": {"title": "...", "price": "£0.00", "rating": 0},
                    "method": "auto",
                },
                wait=True,
                poll_interval=3.0,
            )
            print(f"AI reasoning: {job.generated_config.reasoning}")
            print(f"Extraction: {job.extraction_method_used}")
            if job.schema_used:
                print(f"Schema fields: {[f['name'] for f in job.schema_used['fields']]}")
            ```
        """
        from crawl4ai_cloud.models import SiteScanConfig, SiteExtractConfig

        warnings.warn(
            "crawler.crawl_site() targets the deprecated /v1/crawl/site endpoint. "
            "Migrate to crawler.scan(criteria=...) for URL discovery, then pipe to "
            "crawler.extract_many(url=first, extra_urls=rest, ...) for structured "
            "fields or crawler.scrape_many(urls=...) for markdown. "
            "Will be removed in 0.8.0.",
            DeprecationWarning, stacklevel=2,
        )

        body: Dict[str, Any] = {
            "url": url,
            "max_pages": max_pages,
            "strategy": strategy,
            "fit": fit,
        }

        # --- AI-assisted fields (new) ---
        if criteria:
            body["criteria"] = criteria
        if scan is not None:
            if isinstance(scan, SiteScanConfig):
                body["scan"] = scan.to_dict()
            elif isinstance(scan, dict):
                body["scan"] = scan
            else:
                raise TypeError(
                    f"scan must be dict or SiteScanConfig, got {type(scan).__name__}"
                )
        if extract is not None:
            if isinstance(extract, SiteExtractConfig):
                body["extract"] = extract.to_dict()
            elif isinstance(extract, dict):
                body["extract"] = extract
            else:
                raise TypeError(
                    f"extract must be dict or SiteExtractConfig, got {type(extract).__name__}"
                )
        if include is not None:
            body["include"] = include
        if include_markdown is not None:
            body["include_markdown"] = include_markdown

        # --- Legacy / backward-compat fields ---
        if discovery != "map":
            body["discovery"] = discovery
        if pattern:
            body["pattern"] = pattern
        if max_depth is not None:
            body["max_depth"] = max_depth
        if crawler_config:
            body["crawler_config"] = crawler_config
        if browser_config:
            body["browser_config"] = browser_config
        if proxy:
            body["proxy"] = normalize_proxy(proxy)
        if webhook_url:
            body["webhook_url"] = webhook_url
        body["priority"] = priority

        # Site crawl can take a while when `extract` triggers schema gen
        # (sample URL fetch + LLM call can take 30-120s by itself).
        data = await self._http.request(
            "POST", "/v1/crawl/site", json=body, timeout=240,
        )
        result = SiteCrawlResponse.from_dict(data)

        if wait:
            final = await self._wait_site_crawl_job(
                result.job_id, poll_interval, timeout,
            )
            # Transfer polled state back onto the response
            result.status = final.status
            result.discovered_urls = final.progress.urls_discovered

        return result

    async def get_site_crawl_job(self, job_id: str) -> "SiteCrawlJobStatus":
        """
        Poll a site crawl job started via `crawl_site()`.

        This is the **unified** polling endpoint — it merges the scan phase
        (URL discovery) and the crawl phase (per-page fetch + extract) into
        one response. `phase` walks through "scan" → "crawl" → "done".

        Returns:
            SiteCrawlJobStatus with current phase, progress (urls_discovered,
            urls_crawled, urls_failed, total), and `download_url` once the
            crawl finishes.

        Example:
            ```python
            job = await crawler.crawl_site(
                "https://books.toscrape.com",
                criteria="book listings",
                extract={"query": "book title, price, rating"},
            )

            while True:
                status = await crawler.get_site_crawl_job(job.job_id)
                print(f"{status.phase}: {status.progress.urls_crawled}/{status.progress.total}")
                if status.is_complete:
                    print(f"Download: {status.download_url}")
                    break
                await asyncio.sleep(3)
            ```
        """
        from crawl4ai_cloud.models import SiteCrawlJobStatus
        data = await self._http.request(
            "GET", f"/v1/crawl/site/jobs/{job_id}"
        )
        return SiteCrawlJobStatus.from_dict(data)

    async def _wait_site_crawl_job(
        self,
        job_id: str,
        poll_interval: float = 5.0,
        timeout: Optional[float] = None,
    ) -> "SiteCrawlJobStatus":
        """Poll /v1/crawl/site/jobs/{id} until the crawl finishes."""
        start = time.time()
        while True:
            job = await self.get_site_crawl_job(job_id)
            if job.is_complete:
                return job
            if timeout and (time.time() - start) > timeout:
                raise TimeoutError(
                    f"Timeout waiting for site crawl {job_id}. "
                    f"Phase: {job.phase}, "
                    f"crawled: {job.progress.urls_crawled}/{job.progress.total}"
                )
            await asyncio.sleep(poll_interval)

    # ---- Wrapper job management (shared implementation) ----

    async def _get_wrapper_job(self, job_id: str, job_type: str) -> WrapperJob:
        data = await self._http.request("GET", f"/v1/{job_type}/jobs/{job_id}")
        return WrapperJob.from_dict(data)

    async def _list_wrapper_jobs(
        self, job_type: str, status: Optional[str] = None, limit: int = 20, offset: int = 0,
    ) -> List[WrapperJob]:
        params: Dict[str, Any] = {"limit": limit, "offset": offset}
        if status:
            params["status"] = status
        data = await self._http.request("GET", f"/v1/{job_type}/jobs", params=params)
        return [WrapperJob.from_dict(j) for j in data.get("jobs", [])]

    async def _cancel_wrapper_job(self, job_id: str, job_type: str) -> bool:
        await self._http.request("DELETE", f"/v1/{job_type}/jobs/{job_id}")
        return True

    async def _wait_wrapper_job(
        self, job_id: str, job_type: str, poll_interval: float = 2.0, timeout: Optional[float] = None,
    ) -> WrapperJob:
        start = time.time()
        while True:
            job = await self._get_wrapper_job(job_id, job_type)
            if job.is_complete:
                return job
            if timeout and (time.time() - start) > timeout:
                raise TimeoutError(f"Job {job_id} did not complete within {timeout}s")
            await asyncio.sleep(poll_interval)

    # Public convenience delegates
    async def get_markdown_job(self, job_id: str) -> WrapperJob:
        return await self._get_wrapper_job(job_id, "markdown")

    async def get_screenshot_job(self, job_id: str) -> WrapperJob:
        return await self._get_wrapper_job(job_id, "screenshot")

    async def get_extract_job(self, job_id: str) -> WrapperJob:
        return await self._get_wrapper_job(job_id, "extract")

    async def list_markdown_jobs(self, status: Optional[str] = None, limit: int = 20, offset: int = 0) -> List[WrapperJob]:
        return await self._list_wrapper_jobs("markdown", status, limit, offset)

    async def list_screenshot_jobs(self, status: Optional[str] = None, limit: int = 20, offset: int = 0) -> List[WrapperJob]:
        return await self._list_wrapper_jobs("screenshot", status, limit, offset)

    async def list_extract_jobs(self, status: Optional[str] = None, limit: int = 20, offset: int = 0) -> List[WrapperJob]:
        return await self._list_wrapper_jobs("extract", status, limit, offset)

    async def cancel_markdown_job(self, job_id: str) -> bool:
        return await self._cancel_wrapper_job(job_id, "markdown")

    async def cancel_screenshot_job(self, job_id: str) -> bool:
        return await self._cancel_wrapper_job(job_id, "screenshot")

    async def cancel_extract_job(self, job_id: str) -> bool:
        return await self._cancel_wrapper_job(job_id, "extract")

    # -------------------------------------------------------------------------
    # Enrich (v2 multi-phase API)
    # -------------------------------------------------------------------------

    async def enrich(
        self,
        *,
        # Inputs (any subset; at least one of query / entities / urls)
        query: Optional[str] = None,
        entities: Optional[List[Union[str, Dict[str, Any]]]] = None,
        criteria: Optional[List[Union[str, Dict[str, Any]]]] = None,
        features: Optional[List[Union[str, Dict[str, Any]]]] = None,
        urls: Optional[List[str]] = None,
        groups: Optional[Dict[str, List[str]]] = None,
        # Phase control
        auto_confirm_plan: bool = True,
        auto_confirm_urls: bool = True,
        # Discover knobs
        top_k_per_entity: int = 3,
        search: bool = True,
        country: Optional[str] = None,
        location_hint: Optional[str] = None,
        # Standard wrapper knobs
        strategy: str = "http",
        config: Optional[Dict[str, Any]] = None,
        browser_config: Optional[Dict[str, Any]] = None,
        crawler_config: Optional[Dict[str, Any]] = None,
        llm_config: Optional[Dict[str, Any]] = None,
        proxy: Optional[Union[str, Dict[str, Any], "ProxyConfig"]] = None,
        webhook_url: Optional[str] = None,
        priority: int = 5,
        # Polling control
        wait: bool = True,
        poll_interval: float = 3.0,
        timeout: Optional[float] = 600.0,
    ) -> "EnrichJobStatus":
        """Create a multi-phase enrichment job.

        The phase machine:
            queued → planning → plan_ready → resolving_urls → urls_ready
                  → extracting → merging → completed

        Defaults (`auto_confirm_plan=True`, `auto_confirm_urls=True`) make
        the worker run the full pipeline in one shot — best for agents and
        scripts. Set either flag to False for human-in-loop review and
        resume via `resume_enrich_job(...)`.

        At least one of `query`, `entities`, or `urls` must be supplied.
        The starting phase is inferred from what you pass:
            query                          → planning
            query + entities + features    → resolving_urls
            urls (+ optional groups)       → extracting

        Args:
            query: Free-form brief to expand into entities/criteria/features.
            entities: Pre-supplied row identifiers. Strings are wrapped as
                `{"name": str}`; dicts pass through as-is.
            criteria: Pre-supplied search-side filters (strings → text-only).
            features: Extraction columns. Required unless `query` is set.
                Strings → `{"name": str}`; dicts pass through.
            urls: Skip URL resolution — extract directly from these.
            groups: Pre-grouped URLs per entity (only valid with `urls`).
            auto_confirm_plan: False → pause at `plan_ready` for review.
            auto_confirm_urls: False → pause at `urls_ready` for review.
            top_k_per_entity: URLs crawled per entity (1-10). Default 3.
            search: Enable Serper grounding during plan expansion.
            country: ISO-2 country code for Serper localization.
            location_hint: City/region string for Serper localization.
            strategy: Per-page crawl strategy: "http" (default) or "browser".
            config: EnrichConfig overrides (max_depth, cross_source_verify, …).
            wait: If True (default), poll until terminal status.
            poll_interval: Seconds between polls when waiting.
            timeout: Max seconds to wait.

        Returns:
            EnrichJobStatus. When `wait=True`, status is terminal and
            `phase_data.rows` is populated. When `wait=False`, returns
            immediately with the initial status — poll with
            `get_enrich_job(...)` or stream with `stream_enrich_job(...)`.

        Examples:
            # Agent one-shot
            result = await crawler.enrich(
                query="licensed nurseries in North York Toronto",
                country="ca",
            )
            for row in result.rows:
                print(row.input_key, row.fields)

            # Pre-resolved URLs
            result = await crawler.enrich(
                urls=["https://example.com/a", "https://example.com/b"],
                features=["price", "hours"],
            )

            # Human review flow
            job = await crawler.enrich(
                query="top US BBQ joints",
                auto_confirm_plan=False,
                auto_confirm_urls=False,
                wait=False,
            )
            job = await crawler.wait_enrich_job(job.job_id, until="plan_ready")
            # ...edit job.plan...
            await crawler.resume_enrich_job(job.job_id, features=edited)
        """
        from crawl4ai_cloud.models import EnrichJobStatus

        body: Dict[str, Any] = {
            "auto_confirm_plan": auto_confirm_plan,
            "auto_confirm_urls": auto_confirm_urls,
            "top_k_per_entity": top_k_per_entity,
            "search": search,
            "strategy": strategy,
            "priority": priority,
        }
        if query is not None:
            body["query"] = query
        if entities is not None:
            body["entities"] = [_normalize_entity(e) for e in entities]
        if criteria is not None:
            body["criteria"] = [_normalize_criterion(c) for c in criteria]
        if features is not None:
            body["features"] = [_normalize_feature(f) for f in features]
        if urls is not None:
            body["urls"] = urls
        if groups is not None:
            body["groups"] = groups
        if country is not None:
            body["country"] = country
        if location_hint is not None:
            body["location_hint"] = location_hint
        if config is not None:
            body["config"] = config
        if browser_config is not None:
            body["browser_config"] = browser_config
        if crawler_config is not None:
            body["crawler_config"] = crawler_config
        if llm_config is not None:
            body["llm_config"] = llm_config
        if proxy is not None:
            body["proxy"] = normalize_proxy(proxy)
        if webhook_url is not None:
            body["webhook_url"] = webhook_url

        data = await self._http.request("POST", "/v1/enrich/async", json=body)
        # POST returns the create envelope (job_id, status, created_at).
        # Re-fetch the full status if we're going to wait, otherwise return
        # the create envelope wrapped in EnrichJobStatus.
        job = EnrichJobStatus.from_dict(data)
        if wait:
            return await self.wait_enrich_job(
                job.job_id, poll_interval=poll_interval, timeout=timeout,
            )
        return job

    async def get_enrich_job(self, job_id: str) -> "EnrichJobStatus":
        """Fetch the current status of an enrichment job — one poll, no wait."""
        from crawl4ai_cloud.models import EnrichJobStatus
        data = await self._http.request("GET", f"/v1/enrich/jobs/{job_id}")
        return EnrichJobStatus.from_dict(data)

    async def wait_enrich_job(
        self,
        job_id: str,
        *,
        until: Optional[str] = None,
        poll_interval: float = 3.0,
        timeout: Optional[float] = 600.0,
    ) -> "EnrichJobStatus":
        """Poll an enrichment job until it reaches `until` or a terminal status.

        Args:
            job_id: Job to poll.
            until: Phase to stop at — one of `plan_ready`, `urls_ready`,
                `extracting`, `merging`, `completed`. Default None → wait
                for any terminal status.
            poll_interval: Seconds between polls. Default 3.
            timeout: Max seconds to wait. Default 600.

        Returns:
            EnrichJobStatus at or past the requested phase.

        Raises:
            TimeoutError: If the deadline elapses before the status is reached.
        """
        from crawl4ai_cloud.models import EnrichJobStatus, ENRICH_TERMINAL_STATUSES

        start = time.time()
        target = until or "completed"
        while True:
            job = await self.get_enrich_job(job_id)
            if job.is_complete:
                return job
            if until is not None and job.status == until:
                return job
            # Treat paused phases as a stop condition when the caller asked
            # for a downstream phase that requires /continue to advance.
            if until is not None and job.status in ("plan_ready", "urls_ready"):
                if until == job.status:
                    return job
                # If user asked for a phase that comes AFTER a pause and
                # auto_confirm is False, surface the pause instead of
                # spinning until timeout.
                if (job.status == "plan_ready" and not job.auto_confirm_plan) or (
                    job.status == "urls_ready" and not job.auto_confirm_urls
                ):
                    return job
            if timeout and (time.time() - start) > timeout:
                raise TimeoutError(
                    f"Enrich job {job_id} did not reach '{target}' within {timeout}s. "
                    f"Current status: {job.status}, progress: "
                    f"{job.progress.completed_urls}/{job.progress.total_urls}"
                )
            await asyncio.sleep(poll_interval)

    async def resume_enrich_job(
        self,
        job_id: str,
        *,
        entities: Optional[List[Union[str, Dict[str, Any]]]] = None,
        criteria: Optional[List[Union[str, Dict[str, Any]]]] = None,
        features: Optional[List[Union[str, Dict[str, Any]]]] = None,
        groups: Optional[Dict[str, List[str]]] = None,
    ) -> "EnrichJobStatus":
        """Advance a paused job (`plan_ready` or `urls_ready`) to the next phase.

        Pass any subset of edits to apply before resuming. An empty body
        ({}) is valid — means "resume with the server's current values".

        At `plan_ready`: edits to `entities` / `criteria` / `features`.
        At `urls_ready`: edits to `groups`.
        """
        from crawl4ai_cloud.models import EnrichJobStatus

        body: Dict[str, Any] = {}
        if entities is not None:
            body["entities"] = [_normalize_entity(e) for e in entities]
        if criteria is not None:
            body["criteria"] = [_normalize_criterion(c) for c in criteria]
        if features is not None:
            body["features"] = [_normalize_feature(f) for f in features]
        if groups is not None:
            body["groups"] = groups

        data = await self._http.request(
            "POST", f"/v1/enrich/jobs/{job_id}/continue", json=body,
        )
        return EnrichJobStatus.from_dict(data)

    async def stream_enrich_job(self, job_id: str):
        """Subscribe to the SSE stream for an enrichment job.

        Yields `EnrichEvent` objects until the stream closes (terminal
        status) or the connection drops.

        Event types:
            "snapshot"  — initial full status (sent once on connect)
            "phase"     — phase transition (event.status set)
            "fragment"  — per-URL extraction completed (event.fragment set)
            "row"       — per-entity merged row completed (event.row set)
            "complete"  — terminal status reached; iterator stops

        Example:
            async for event in crawler.stream_enrich_job(job_id):
                if event.type == "row":
                    print("✓", event.row.input_key)
                elif event.type == "complete":
                    break
        """
        from crawl4ai_cloud.models import EnrichEvent

        async for evt_type, payload in self._http.stream_sse(
            f"/v1/enrich/jobs/{job_id}/stream",
        ):
            yield EnrichEvent.from_dict(evt_type, payload)
            if evt_type == "complete":
                return

    async def cancel_enrich_job(self, job_id: str) -> bool:
        """Cancel a running enrichment job. Returns True on success."""
        await self._http.request("DELETE", f"/v1/enrich/jobs/{job_id}")
        return True

    async def list_enrich_jobs(
        self, limit: int = 20, offset: int = 0,
    ) -> List["EnrichJobListItem"]:
        """List enrichment jobs for the authenticated user."""
        from crawl4ai_cloud.models import EnrichJobListItem
        data = await self._http.request(
            "GET", "/v1/enrich/jobs",
            params={"limit": limit, "offset": offset},
        )
        return [EnrichJobListItem.from_dict(j) for j in data.get("jobs", [])]

    # -------------------------------------------------------------------------
    # Context Manager
    # -------------------------------------------------------------------------

    async def __aenter__(self) -> "AsyncWebCrawler":
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()

    async def close(self):
        """Close the HTTP client."""
        await self._http.close()

    # =========================================================================
    # Discovery — wrapper-services platform: /v1/discovery/<service>
    # =========================================================================
    #
    # One method, dispatches to any registered vertical. `search` is live;
    # `people`, `products`, `posts`, `videos` will land via the same call shape.
    # See `list_discovery_services()` for the registry.

    async def discovery(
        self,
        service: str,
        **params: Any,
    ) -> Any:
        """Run a Discovery vertical and return the typed response.

        ``POST /v1/discovery/<service>`` — the wrapper-services dispatcher.
        New verticals don't add SDK methods; they become a new value for
        ``service``.

        Args:
            service: Vertical name (``"search"`` today; ``"people"`` /
                ``"products"`` / ``"posts"`` / ``"videos"`` to follow).
            **params: Per-vertical request fields. For ``service="search"``:
                ``query`` (required), ``country``, ``language``, ``location``,
                ``num``, ``start``, ``site``, ``mode``, ``time_period``,
                ``bypass_cache``. See the Discovery docs for full schemas.

        Returns:
            ``SearchResponse`` for ``service="search"``. Generic ``dict`` for
            verticals whose typed response classes don't exist yet — callers
            can index it the same way the API returns.

        Example:
            >>> response = await crawler.discovery(
            ...     "search",
            ...     query="best AI code review tools 2026",
            ...     country="us",
            ... )
            >>> for hit in response.hits:
            ...     print(hit.rank, hit.title, hit.url)
        """
        from crawl4ai_cloud.models import SearchResponse

        # Drop None / empty-string optionals so the cache key matches what
        # `runDiscovery()` and the dashboard playground actually send. Wire
        # parity with the playground avoids surprise cache-misses between
        # surfaces that hit the same params.
        body = {k: v for k, v in params.items() if v is not None and v != ""}
        data = await self._http.request(
            "POST", f"/v1/discovery/{service}", json=body,
        )

        # Typed response per vertical. As more verticals ship typed models,
        # extend this dispatch — generic dict otherwise.
        if service == "search":
            return SearchResponse.from_dict(data)
        return data

    async def list_discovery_services(self) -> List["DiscoveryService"]:
        """Fetch the Discovery service registry.

        ``GET /v1/discovery`` — returns every vertical the cloud currently
        ships, plus its request/response JSON schemas. Use this to feature-
        detect new verticals without an SDK update.

        Returns:
            List of :class:`DiscoveryService` entries.

        Example:
            >>> services = await crawler.list_discovery_services()
            >>> for svc in services:
            ...     print(svc.name, "—", svc.description)
        """
        from crawl4ai_cloud.models import DiscoveryService
        data = await self._http.request("GET", "/v1/discovery")
        return [DiscoveryService.from_dict(s) for s in (data.get("services") or [])]
