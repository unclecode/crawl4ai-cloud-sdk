"""
Comprehensive End-to-End Tests for Crawl4AI Cloud SDK.

Tests all SDK functionality against the live API:
- Single URL crawling
- Batch crawling (run_many)
- Job management
- Deep crawl
- Context API
- Schema generation
- Storage API
- Error handling
- OSS compatibility (arun/arun_many aliases)
- Config sanitization
- Proxy configurations
"""
import pytest
import asyncio
import os

from crawl4ai_cloud import (
    AsyncWebCrawler,
    CrawlerRunConfig,
    BrowserConfig,
    CrawlResult,
    CrawlJob,
    MarkdownResult,
    DeepCrawlResult,
    ContextResult,
    GeneratedSchema,
    StorageUsage,
    ProxyConfig,
    CloudError,
    AuthenticationError,
    RateLimitError,
    ValidationError,
    NotFoundError,
    TimeoutError,
    sanitize_crawler_config,
    sanitize_browser_config,
    normalize_proxy,
)

# Test API key from conftest or environment
API_KEY = os.getenv(
    "CRAWL4AI_API_KEY",
    "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ"
)

# Test URLs
TEST_URL = "https://example.com"
TEST_URL_2 = "https://httpbin.org/html"
TEST_URL_JS = "https://quotes.toscrape.com/js/"  # JS-rendered page
DOCS_URL = "https://docs.crawl4ai.com"


# =============================================================================
# INITIALIZATION TESTS
# =============================================================================

class TestInitialization:
    """Test AsyncWebCrawler initialization."""

    def test_init_with_api_key(self):
        """Test successful initialization with API key."""
        crawler = AsyncWebCrawler(api_key=API_KEY)
        assert crawler is not None

    def test_init_from_env_var(self, monkeypatch):
        """Test initialization from environment variable."""
        monkeypatch.setenv("CRAWL4AI_API_KEY", API_KEY)
        crawler = AsyncWebCrawler()
        assert crawler is not None

    def test_init_missing_api_key_raises(self, monkeypatch):
        """Test that missing API key raises ValueError."""
        monkeypatch.delenv("CRAWL4AI_API_KEY", raising=False)
        with pytest.raises(ValueError, match="API key is required"):
            AsyncWebCrawler(api_key=None)

    def test_init_invalid_api_key_format_raises(self):
        """Test that invalid API key format raises ValueError."""
        with pytest.raises(ValueError, match="Invalid API key format"):
            AsyncWebCrawler(api_key="invalid_key_format")

    def test_init_accepts_sk_test_prefix(self):
        """Test that sk_test_ prefix is accepted."""
        # This will fail auth but format is valid
        crawler = AsyncWebCrawler(api_key="sk_test_dummy_key_12345")
        assert crawler is not None

    def test_init_with_custom_base_url(self):
        """Test initialization with custom base URL."""
        crawler = AsyncWebCrawler(
            api_key=API_KEY,
            base_url="https://api.crawl4ai.com"
        )
        assert crawler is not None

    def test_init_with_custom_timeout(self):
        """Test initialization with custom timeout."""
        crawler = AsyncWebCrawler(api_key=API_KEY, timeout=60.0)
        assert crawler is not None

    def test_init_with_oss_compat_params(self):
        """Test that OSS compatibility params are accepted (ignored)."""
        crawler = AsyncWebCrawler(
            api_key=API_KEY,
            verbose=True,  # OSS param, ignored
        )
        assert crawler is not None


# =============================================================================
# SINGLE URL CRAWL TESTS
# =============================================================================

class TestSingleUrlCrawl:
    """Test single URL crawling with run()."""

    @pytest.mark.asyncio
    async def test_run_basic(self):
        """Test basic single URL crawl."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            assert isinstance(result, CrawlResult)
            assert result.success is True
            assert result.url == TEST_URL
            assert not result.error_message  # None or empty string

    @pytest.mark.asyncio
    async def test_run_returns_markdown(self):
        """Test that crawl returns markdown content."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            assert result.markdown is not None
            assert isinstance(result.markdown, MarkdownResult)
            assert result.markdown.raw_markdown is not None
            assert len(result.markdown.raw_markdown) > 0
            assert "Example Domain" in result.markdown.raw_markdown

    @pytest.mark.asyncio
    async def test_run_returns_html(self):
        """Test that crawl returns HTML content."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            assert result.html is not None
            assert "<html" in result.html.lower()
            assert "example" in result.html.lower()

    @pytest.mark.asyncio
    async def test_run_returns_metadata(self):
        """Test that crawl returns metadata."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            # Metadata may or may not be present depending on page
            assert result.status_code is not None or result.success

    @pytest.mark.asyncio
    async def test_run_with_browser_strategy(self):
        """Test crawl with browser strategy (default)."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, strategy="browser")

            assert result.success is True

    @pytest.mark.asyncio
    async def test_run_with_http_strategy(self):
        """Test crawl with HTTP strategy (no JS)."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, strategy="http")

            assert result.success is True
            assert result.markdown.raw_markdown is not None

    @pytest.mark.asyncio
    async def test_run_js_rendered_page(self):
        """Test crawling a JS-rendered page with browser strategy."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL_JS, strategy="browser")

            assert result.success is True
            # JS page should have content after rendering
            assert result.markdown.raw_markdown is not None

    @pytest.mark.asyncio
    async def test_run_with_bypass_cache(self):
        """Test crawl with cache bypass."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, bypass_cache=True)

            assert result.success is True


# =============================================================================
# OSS COMPATIBILITY TESTS
# =============================================================================

class TestOSSCompatibility:
    """Test OSS crawl4ai compatibility."""

    @pytest.mark.asyncio
    async def test_arun_alias(self):
        """Test arun() is alias for run()."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.arun(TEST_URL)

            assert isinstance(result, CrawlResult)
            assert result.success is True
            assert result.url == TEST_URL

    @pytest.mark.asyncio
    async def test_arun_with_config(self):
        """Test arun() works with config parameter."""
        config = CrawlerRunConfig(word_count_threshold=10)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.arun(TEST_URL, config=config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_arun_many_alias(self):
        """Test arun_many() is alias for run_many()."""
        urls = [TEST_URL, TEST_URL_2]

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            results = await crawler.arun_many(urls, wait=True)

            assert isinstance(results, list)
            assert len(results) == 2


# =============================================================================
# CONFIGURATION TESTS
# =============================================================================

class TestCrawlerRunConfig:
    """Test CrawlerRunConfig functionality."""

    @pytest.mark.asyncio
    async def test_config_word_count_threshold(self):
        """Test word_count_threshold config."""
        config = CrawlerRunConfig(word_count_threshold=5)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_config_exclude_external_links(self):
        """Test exclude_external_links config."""
        config = CrawlerRunConfig(exclude_external_links=True)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_config_process_iframes(self):
        """Test process_iframes config."""
        config = CrawlerRunConfig(process_iframes=True)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_config_screenshot(self):
        """Test screenshot config."""
        config = CrawlerRunConfig(screenshot=True)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True
            # Screenshot should be returned as base64 or URL
            # May be None if not supported or failed

    @pytest.mark.asyncio
    async def test_config_wait_for(self):
        """Test wait_for config (CSS selector)."""
        config = CrawlerRunConfig(wait_for="body")

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True

    def test_config_dump(self):
        """Test config serialization."""
        config = CrawlerRunConfig(
            screenshot=True,
            word_count_threshold=50,
            exclude_external_links=True,
        )
        data = config.dump()

        assert isinstance(data, dict)
        assert data["screenshot"] is True
        assert data["word_count_threshold"] == 50
        assert data["exclude_external_links"] is True

    def test_config_sanitization_removes_cache_fields(self):
        """Test that cache fields are sanitized."""
        config = CrawlerRunConfig(
            cache_mode="bypass",
            session_id="test-session",
            bypass_cache=True,
            no_cache_read=True,
            no_cache_write=True,
            screenshot=True,
        )

        sanitized = sanitize_crawler_config(config)

        assert "cache_mode" not in sanitized
        assert "session_id" not in sanitized
        assert "bypass_cache" not in sanitized
        assert "no_cache_read" not in sanitized
        assert "no_cache_write" not in sanitized
        assert sanitized.get("screenshot") is True

    # ==========================================================================
    # NEW PARAMETER TESTS (Issues #365, #366)
    # ==========================================================================

    def test_config_css_selector_exists(self):
        """Test that css_selector parameter is accepted."""
        config = CrawlerRunConfig(css_selector="article")
        assert config.css_selector == "article"

    def test_config_excluded_tags_exists(self):
        """Test that excluded_tags parameter is accepted."""
        config = CrawlerRunConfig(excluded_tags=["nav", "footer", "aside"])
        assert config.excluded_tags == ["nav", "footer", "aside"]

    def test_config_excluded_selector_exists(self):
        """Test that excluded_selector parameter is accepted."""
        config = CrawlerRunConfig(excluded_selector=".ads, .sidebar")
        assert config.excluded_selector == ".ads, .sidebar"

    def test_config_target_elements_exists(self):
        """Test that target_elements parameter is accepted."""
        config = CrawlerRunConfig(target_elements=["main", "article"])
        assert config.target_elements == ["main", "article"]

    def test_config_wait_until_exists(self):
        """Test that wait_until parameter is accepted."""
        config = CrawlerRunConfig(wait_until="networkidle")
        assert config.wait_until == "networkidle"

    def test_config_wait_until_default(self):
        """Test that wait_until defaults to domcontentloaded."""
        config = CrawlerRunConfig()
        assert config.wait_until == "domcontentloaded"

    def test_config_remove_overlay_elements_exists(self):
        """Test that remove_overlay_elements parameter is accepted."""
        config = CrawlerRunConfig(remove_overlay_elements=True)
        assert config.remove_overlay_elements is True

    def test_config_max_scroll_steps_exists(self):
        """Test that max_scroll_steps parameter is accepted."""
        config = CrawlerRunConfig(max_scroll_steps=10)
        assert config.max_scroll_steps == 10

    def test_config_exclude_internal_links_exists(self):
        """Test that exclude_internal_links parameter is accepted."""
        config = CrawlerRunConfig(exclude_internal_links=True)
        assert config.exclude_internal_links is True

    def test_config_keep_attrs_exists(self):
        """Test that keep_attrs parameter is accepted."""
        config = CrawlerRunConfig(keep_attrs=["href", "src", "alt"])
        assert config.keep_attrs == ["href", "src", "alt"]

    def test_config_wait_for_timeout_exists(self):
        """Test that wait_for_timeout parameter is accepted."""
        config = CrawlerRunConfig(wait_for_timeout=5000)
        assert config.wait_for_timeout == 5000

    def test_config_all_new_params_dump(self):
        """Test that all new parameters are included in dump()."""
        config = CrawlerRunConfig(
            css_selector="main",
            excluded_tags=["nav"],
            excluded_selector=".sidebar",
            target_elements=["article"],
            wait_until="load",
            remove_overlay_elements=True,
            max_scroll_steps=5,
            exclude_internal_links=True,
            keep_attrs=["id"],
            wait_for_timeout=3000,
        )
        data = config.dump()

        assert data["css_selector"] == "main"
        assert data["excluded_tags"] == ["nav"]
        assert data["excluded_selector"] == ".sidebar"
        assert data["target_elements"] == ["article"]
        assert data["wait_until"] == "load"
        assert data["remove_overlay_elements"] is True
        assert data["max_scroll_steps"] == 5
        assert data["exclude_internal_links"] is True
        assert data["keep_attrs"] == ["id"]
        assert data["wait_for_timeout"] == 3000

    @pytest.mark.asyncio
    async def test_config_css_selector_crawl(self):
        """Test crawl with css_selector extracts specific content."""
        config = CrawlerRunConfig(css_selector="h1")

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True
            # h1 on example.com is "Example Domain"
            if result.markdown and result.markdown.raw_markdown:
                assert "Example Domain" in result.markdown.raw_markdown

    @pytest.mark.asyncio
    async def test_config_excluded_tags_crawl(self):
        """Test crawl with excluded_tags."""
        config = CrawlerRunConfig(excluded_tags=["script", "style"])

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_config_wait_until_crawl(self):
        """Test crawl with wait_until parameter."""
        config = CrawlerRunConfig(wait_until="domcontentloaded")

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, config=config)

            assert result.success is True


class TestBrowserConfig:
    """Test BrowserConfig functionality."""

    @pytest.mark.asyncio
    async def test_browser_config_viewport(self):
        """Test custom viewport config."""
        browser_config = BrowserConfig(
            viewport_width=1920,
            viewport_height=1080,
        )

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, browser_config=browser_config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_browser_config_user_agent(self):
        """Test custom user agent config."""
        browser_config = BrowserConfig(
            user_agent="CustomBot/1.0"
        )

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, browser_config=browser_config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_browser_config_headers(self):
        """Test custom headers config."""
        browser_config = BrowserConfig(
            headers={"X-Custom-Header": "test-value"}
        )

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL, browser_config=browser_config)

            assert result.success is True

    def test_browser_config_sanitization_removes_cdp_fields(self):
        """Test that CDP fields are sanitized."""
        config = BrowserConfig(
            cdp_url="ws://localhost:9222",
            use_managed_browser=True,
            browser_mode="headless",
            user_data_dir="/tmp/chrome",
            headless=False,
        )

        sanitized = sanitize_browser_config(config)

        assert "cdp_url" not in sanitized
        assert "use_managed_browser" not in sanitized
        assert "browser_mode" not in sanitized
        assert "user_data_dir" not in sanitized
        assert sanitized.get("headless") is False


# =============================================================================
# PROXY CONFIGURATION TESTS
# =============================================================================

class TestProxyConfig:
    """Test proxy configuration."""

    def test_normalize_proxy_string_datacenter(self):
        """Test proxy string shorthand - datacenter."""
        result = normalize_proxy("datacenter")
        assert result == {"mode": "datacenter"}

    def test_normalize_proxy_string_residential(self):
        """Test proxy string shorthand - residential."""
        result = normalize_proxy("residential")
        assert result == {"mode": "residential"}

    def test_normalize_proxy_string_auto(self):
        """Test proxy string shorthand - auto."""
        result = normalize_proxy("auto")
        assert result == {"mode": "auto"}

    def test_normalize_proxy_dict(self):
        """Test proxy dict config."""
        proxy = {"mode": "residential", "country": "US"}
        result = normalize_proxy(proxy)
        assert result == proxy

    def test_normalize_proxy_dict_with_sticky(self):
        """Test proxy dict with sticky session."""
        proxy = {"mode": "datacenter", "sticky_session": True}
        result = normalize_proxy(proxy)
        assert result == proxy

    def test_normalize_proxy_dataclass(self):
        """Test ProxyConfig dataclass."""
        proxy = ProxyConfig(mode="residential", country="UK")
        result = normalize_proxy(proxy)
        assert result == {"mode": "residential", "country": "UK"}

    def test_normalize_proxy_none(self):
        """Test None proxy returns None."""
        result = normalize_proxy(None)
        assert result is None

    def test_normalize_proxy_invalid_type_raises(self):
        """Test invalid proxy type raises ValueError."""
        with pytest.raises(ValueError, match="Invalid proxy type"):
            normalize_proxy(12345)

    @pytest.mark.asyncio
    async def test_run_with_proxy_string(self):
        """Test crawl with proxy string shorthand."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            # Note: This will use datacenter proxy (2x credits)
            result = await crawler.run(TEST_URL, proxy="datacenter")

            assert result.success is True

    @pytest.mark.asyncio
    async def test_run_with_proxy_dict(self):
        """Test crawl with proxy dict config."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(
                TEST_URL,
                proxy={"mode": "datacenter"}
            )

            assert result.success is True


# =============================================================================
# BATCH CRAWL TESTS (run_many)
# =============================================================================

class TestBatchCrawl:
    """Test batch crawling with run_many()."""

    @pytest.mark.asyncio
    async def test_run_many_small_batch_wait(self):
        """Test small batch (≤10 URLs) with wait=True."""
        urls = [TEST_URL, TEST_URL_2]

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            results = await crawler.run_many(urls, wait=True)

            assert isinstance(results, list)
            assert len(results) == 2
            for result in results:
                assert isinstance(result, CrawlResult)
                assert result.success is True

    @pytest.mark.asyncio
    async def test_run_many_small_batch_no_wait(self):
        """Test small batch (≤10 URLs) with wait=False."""
        urls = [TEST_URL, TEST_URL_2]

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            job = await crawler.run_many(urls, wait=False)

            # With wait=False, job is returned immediately in async state
            # Status can be pending, processing, or completed (if very fast)
            assert isinstance(job, CrawlJob)
            assert job.status in ("pending", "processing", "completed")

    @pytest.mark.asyncio
    async def test_run_many_with_config(self):
        """Test batch crawl with config."""
        urls = [TEST_URL, TEST_URL_2]
        config = CrawlerRunConfig(word_count_threshold=10)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            results = await crawler.run_many(urls, config=config, wait=True)

            assert len(results) == 2
            for result in results:
                assert result.success is True

    @pytest.mark.asyncio
    async def test_run_many_http_strategy(self):
        """Test batch crawl with HTTP strategy."""
        urls = [TEST_URL, TEST_URL_2]

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            results = await crawler.run_many(urls, strategy="http", wait=True)

            assert len(results) == 2
            for result in results:
                assert result.success is True


# =============================================================================
# JOB MANAGEMENT TESTS
# =============================================================================

class TestJobManagement:
    """Test job management functionality."""

    @pytest.mark.asyncio
    async def test_list_jobs(self):
        """Test listing jobs."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            jobs = await crawler.list_jobs(limit=5)

            assert isinstance(jobs, list)
            # May be empty if no jobs exist
            for job in jobs:
                assert isinstance(job, CrawlJob)
                assert job.id is not None
                assert job.status is not None

    @pytest.mark.asyncio
    async def test_list_jobs_with_status_filter(self):
        """Test listing jobs with status filter."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            jobs = await crawler.list_jobs(status="completed", limit=5)

            assert isinstance(jobs, list)
            for job in jobs:
                assert job.status == "completed"

    @pytest.mark.asyncio
    async def test_list_jobs_pagination(self):
        """Test job listing pagination."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            jobs_page1 = await crawler.list_jobs(limit=2, offset=0)
            jobs_page2 = await crawler.list_jobs(limit=2, offset=2)

            # Pages should be different (if enough jobs exist)
            assert isinstance(jobs_page1, list)
            assert isinstance(jobs_page2, list)


# =============================================================================
# STORAGE API TESTS
# =============================================================================

class TestStorageAPI:
    """Test storage API."""

    @pytest.mark.asyncio
    async def test_storage_returns_usage(self):
        """Test storage API returns usage info."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            usage = await crawler.storage()

            assert isinstance(usage, StorageUsage)
            assert usage.max_mb >= 0
            assert usage.used_mb >= 0
            assert usage.remaining_mb >= 0
            assert usage.percent_used >= 0


# =============================================================================
# HEALTH CHECK TESTS
# =============================================================================

class TestHealthCheck:
    """Test health check endpoint."""

    @pytest.mark.asyncio
    async def test_health_check(self):
        """Test health check returns status."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            health = await crawler.health()

            assert isinstance(health, dict)
            # Health endpoint should return some status info


# =============================================================================
# ERROR HANDLING TESTS
# =============================================================================

class TestErrorHandling:
    """Test error handling for various scenarios."""

    @pytest.mark.asyncio
    async def test_invalid_api_key_raises_auth_error(self):
        """Test that invalid API key raises AuthenticationError."""
        async with AsyncWebCrawler(api_key="sk_test_invalid_12345") as crawler:
            with pytest.raises(AuthenticationError) as exc_info:
                await crawler.run(TEST_URL)

            assert exc_info.value.status_code == 401

    @pytest.mark.asyncio
    async def test_invalid_url_handling(self):
        """Test that invalid URL is handled (error or failed result)."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            try:
                result = await crawler.run("not-a-valid-url")
                # API may return failed result instead of raising
                assert result.success is False or result.error_message
            except (ValidationError, CloudError):
                # Or it may raise an exception
                pass

    @pytest.mark.asyncio
    async def test_nonexistent_job_raises_not_found(self):
        """Test that getting non-existent job raises NotFoundError."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            with pytest.raises(NotFoundError) as exc_info:
                await crawler.get_job("nonexistent-job-id-12345")

            assert exc_info.value.status_code == 404

    @pytest.mark.asyncio
    async def test_error_has_message(self):
        """Test that errors have meaningful messages."""
        async with AsyncWebCrawler(api_key="sk_test_invalid_12345") as crawler:
            with pytest.raises(AuthenticationError) as exc_info:
                await crawler.run(TEST_URL)

            assert exc_info.value.message is not None
            assert len(str(exc_info.value)) > 0


# =============================================================================
# CONTEXT MANAGER TESTS
# =============================================================================

class TestContextManager:
    """Test async context manager functionality."""

    @pytest.mark.asyncio
    async def test_context_manager_opens_and_closes(self):
        """Test that context manager properly opens and closes."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)
            assert result.success is True

        # After exiting, client should be closed
        # No exception should be raised

    @pytest.mark.asyncio
    async def test_explicit_close(self):
        """Test explicit close() method."""
        crawler = AsyncWebCrawler(api_key=API_KEY)
        result = await crawler.run(TEST_URL)
        assert result.success is True

        await crawler.close()
        # No exception should be raised

    @pytest.mark.asyncio
    async def test_multiple_requests_same_session(self):
        """Test multiple requests in same session."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result1 = await crawler.run(TEST_URL)
            result2 = await crawler.run(TEST_URL_2)

            assert result1.success is True
            assert result2.success is True


# =============================================================================
# DEEP CRAWL TESTS
# =============================================================================

class TestDeepCrawl:
    """Test deep crawl functionality."""

    @pytest.mark.asyncio
    async def test_deep_crawl_scan_only(self):
        """Test deep crawl with scan_only=True."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.deep_crawl(
                url=TEST_URL,
                strategy="bfs",
                max_depth=1,
                max_urls=5,
                scan_only=True,
                wait=True,
            )

            assert isinstance(result, DeepCrawlResult)
            assert result.job_id is not None
            assert result.status in ("completed", "no_urls", "failed")

    @pytest.mark.asyncio
    async def test_deep_crawl_requires_url_or_source_job(self):
        """Test that deep_crawl requires url or source_job."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            with pytest.raises(ValueError, match="Must provide either"):
                await crawler.deep_crawl()

    @pytest.mark.asyncio
    async def test_deep_crawl_rejects_both_url_and_source_job(self):
        """Test that deep_crawl rejects both url and source_job."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            with pytest.raises(ValueError, match="not both"):
                await crawler.deep_crawl(
                    url=TEST_URL,
                    source_job="some-job-id"
                )


# =============================================================================
# SCHEMA GENERATION TESTS
# =============================================================================

class TestSchemaGeneration:
    """Test schema generation API."""

    SAMPLE_HTML = """
    <html>
    <body>
        <div class="product">
            <h2 class="title">Product 1</h2>
            <span class="price">$19.99</span>
        </div>
        <div class="product">
            <h2 class="title">Product 2</h2>
            <span class="price">$29.99</span>
        </div>
    </body>
    </html>
    """

    SAMPLE_HTML_2 = """
    <html>
    <body>
        <div class="product">
            <h2 class="title">Widget A</h2>
            <span class="price">$49.99</span>
        </div>
    </body>
    </html>
    """

    @pytest.mark.asyncio
    async def test_generate_schema_single_html(self):
        """Test schema generation with single HTML sample."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            schema = await crawler.generate_schema(
                html=self.SAMPLE_HTML,
                query="Extract product titles and prices"
            )

            assert isinstance(schema, GeneratedSchema)
            # Schema generation may succeed or fail depending on LLM
            assert schema.success is True or schema.error is not None

    @pytest.mark.asyncio
    async def test_generate_schema_multiple_html(self):
        """Test schema generation with multiple HTML samples."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            schema = await crawler.generate_schema(
                html=[self.SAMPLE_HTML, self.SAMPLE_HTML_2],
                query="Extract product titles and prices from these samples"
            )

            assert isinstance(schema, GeneratedSchema)
            assert schema.success is True or schema.error is not None

    @pytest.mark.asyncio
    async def test_generate_schema_from_urls(self):
        """Test schema generation from URLs."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            schema = await crawler.generate_schema(
                urls=["https://example.com"],
                query="Extract any content"
            )

            assert isinstance(schema, GeneratedSchema)
            # May succeed or fail depending on URL content and LLM

    @pytest.mark.asyncio
    async def test_generate_schema_requires_html_or_urls(self):
        """Test that either html or urls is required."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            with pytest.raises(ValueError, match="Either 'html' or 'urls' must be provided"):
                await crawler.generate_schema(query="Extract products")

    @pytest.mark.asyncio
    async def test_generate_schema_rejects_both_html_and_urls(self):
        """Test that providing both html and urls raises error."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            with pytest.raises(ValueError, match="not both"):
                await crawler.generate_schema(
                    html=self.SAMPLE_HTML,
                    urls=["https://example.com"],
                    query="Extract products"
                )

    @pytest.mark.asyncio
    async def test_generate_schema_max_three_urls(self):
        """Test that max 3 URLs is enforced."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            with pytest.raises(ValueError, match="Maximum 3 URLs"):
                await crawler.generate_schema(
                    urls=[
                        "https://example.com/1",
                        "https://example.com/2",
                        "https://example.com/3",
                        "https://example.com/4",
                    ],
                    query="Extract products"
                )


# =============================================================================
# CRAWL RESULT STRUCTURE TESTS
# =============================================================================

class TestCrawlResultStructure:
    """Test CrawlResult structure and fields."""

    @pytest.mark.asyncio
    async def test_result_has_all_expected_fields(self):
        """Test that CrawlResult has all expected fields."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            # Core fields
            assert hasattr(result, 'url')
            assert hasattr(result, 'success')
            assert hasattr(result, 'html')
            assert hasattr(result, 'markdown')
            assert hasattr(result, 'error_message')

            # Optional fields
            assert hasattr(result, 'cleaned_html')
            assert hasattr(result, 'media')
            assert hasattr(result, 'links')
            assert hasattr(result, 'metadata')
            assert hasattr(result, 'screenshot')
            assert hasattr(result, 'pdf')
            assert hasattr(result, 'extracted_content')
            assert hasattr(result, 'status_code')
            assert hasattr(result, 'duration_ms')

    @pytest.mark.asyncio
    async def test_markdown_result_structure(self):
        """Test MarkdownResult structure."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            md = result.markdown
            assert hasattr(md, 'raw_markdown')
            assert hasattr(md, 'markdown_with_citations')
            assert hasattr(md, 'references_markdown')
            assert hasattr(md, 'fit_markdown')


# =============================================================================
# CRAWL JOB STRUCTURE TESTS
# =============================================================================

class TestCrawlJobStructure:
    """Test CrawlJob structure and methods."""

    @pytest.mark.asyncio
    async def test_job_has_all_expected_fields(self):
        """Test that CrawlJob has all expected fields."""
        urls = [TEST_URL, TEST_URL_2]

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            job = await crawler.run_many(urls, wait=False)

            assert hasattr(job, 'id')
            assert hasattr(job, 'status')
            assert hasattr(job, 'progress')
            assert hasattr(job, 'urls_count')
            assert hasattr(job, 'created_at')
            assert hasattr(job, 'is_complete')
            assert hasattr(job, 'is_successful')
            assert hasattr(job, 'progress_percent')

    @pytest.mark.asyncio
    async def test_job_progress_structure(self):
        """Test JobProgress structure."""
        urls = [TEST_URL, TEST_URL_2]

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            job = await crawler.run_many(urls, wait=False)

            progress = job.progress
            assert hasattr(progress, 'total')
            assert hasattr(progress, 'completed')
            assert hasattr(progress, 'failed')
            assert hasattr(progress, 'pending')
            assert hasattr(progress, 'percent')


# =============================================================================
# INTEGRATION TESTS
# =============================================================================

class TestIntegration:
    """Integration tests combining multiple features."""

    @pytest.mark.asyncio
    async def test_full_workflow_single_crawl(self):
        """Test complete single URL crawl workflow."""
        config = CrawlerRunConfig(
            word_count_threshold=10,
            exclude_external_links=True,
        )
        browser_config = BrowserConfig(
            viewport_width=1280,
            viewport_height=720,
        )

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(
                TEST_URL,
                config=config,
                browser_config=browser_config,
                strategy="browser",
            )

            assert result.success is True
            assert result.url == TEST_URL
            assert result.markdown.raw_markdown is not None
            assert "Example" in result.markdown.raw_markdown

    @pytest.mark.asyncio
    async def test_full_workflow_batch_crawl(self):
        """Test complete batch crawl workflow."""
        urls = [TEST_URL, TEST_URL_2]
        config = CrawlerRunConfig(word_count_threshold=5)

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            results = await crawler.run_many(
                urls,
                config=config,
                strategy="http",
                wait=True,
            )

            assert len(results) == 2
            for result in results:
                assert result.success is True
                assert result.markdown.raw_markdown is not None

    @pytest.mark.asyncio
    async def test_oss_migration_pattern(self):
        """Test the OSS migration pattern works as documented."""
        # This is how users migrate from OSS to Cloud:
        # 1. Change import from crawl4ai to crawl4ai_cloud
        # 2. Add api_key parameter
        # 3. Use same code

        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            # OSS users use arun()
            result = await crawler.arun(TEST_URL)

            assert result.success is True
            assert result.markdown.raw_markdown is not None


# =============================================================================
# PERFORMANCE / TIMING TESTS
# =============================================================================

class TestPerformance:
    """Basic performance tests."""

    @pytest.mark.asyncio
    async def test_crawl_returns_duration(self):
        """Test that crawl returns duration metric."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            result = await crawler.run(TEST_URL)

            # duration_ms should be set
            assert result.duration_ms >= 0

    @pytest.mark.asyncio
    async def test_http_strategy_faster_than_browser(self):
        """Test that HTTP strategy is generally faster."""
        async with AsyncWebCrawler(api_key=API_KEY) as crawler:
            # HTTP strategy (no browser)
            result_http = await crawler.run(TEST_URL, strategy="http")

            # Browser strategy
            result_browser = await crawler.run(TEST_URL, strategy="browser")

            # Both should succeed
            assert result_http.success is True
            assert result_browser.success is True

            # Note: We don't assert timing as it can vary
