"""AsyncWebCrawler - Main crawler class for Crawl4AI Cloud SDK."""
import asyncio
import time
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
)
from .configs import (
    CrawlerRunConfig,
    BrowserConfig,
    sanitize_crawler_config,
    sanitize_browser_config,
    normalize_proxy,
    build_crawl_request,
)


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

        # Build request body
        body: Dict[str, Any] = {
            "strategy": strategy,
            "crawl_strategy": crawl_strategy,
            "priority": priority,
        }

        if url:
            body["url"] = url
        if source_job:
            body["source_job_id"] = source_job

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

        # Add configs
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
        return DeepCrawlResult(
            job_id=job_id,
            status="cancelled",
            strategy=data.get("strategy"),
            discovered_count=data.get("discovered_urls", 0),
        )

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
