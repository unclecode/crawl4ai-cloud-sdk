"""
E2E Adversarial Test Suite for Wrapper API SDK Methods.

Real HTTP tests against stage.crawl4ai.com. No mocks.
Tests: happy path, config passthrough, async lifecycle, errors,
injection, namespace isolation, batch AUTO rejection, cancel.

Run:
    cd crawl4ai-cloud-sdk/python
    python -m pytest tests/test_wrapper_e2e.py -v --timeout=120
"""

import asyncio
import os
import pytest
import pytest_asyncio
import time

from crawl4ai_cloud import (
    AsyncWebCrawler,
    MarkdownResponse,
    ScreenshotResponse,
    ExtractResponse,
    MapResponse,
    SiteCrawlResponse,
    SiteCrawlJobStatus,
    ScanResult,
    ScanJobStatus,
    SiteScanConfig,
    SiteExtractConfig,
    GeneratedConfig,
    WrapperJob,
    WrapperUsage,
)
from crawl4ai_cloud.errors import AuthenticationError, NotFoundError

API_KEY = os.environ.get("CRAWL4AI_API_KEY", "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ")
FREE_KEY = os.environ.get("CRAWL4AI_FREE_KEY", "sk_live_oA8wznhyeNNCjhoSwVrEPAJjkQodQSwPq_stsg-gL3c")
BASE_URL = os.environ.get("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")


@pytest_asyncio.fixture
async def crawler():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as c:
        yield c


@pytest_asyncio.fixture
async def free_crawler():
    async with AsyncWebCrawler(api_key=FREE_KEY, base_url=BASE_URL) as c:
        yield c


# =============================================================================
# MARKDOWN
# =============================================================================


class TestMarkdown:
    @pytest.mark.asyncio
    async def test_basic(self, crawler):
        md = await crawler.markdown("https://example.com", strategy="http")
        assert isinstance(md, MarkdownResponse)
        assert md.success is True
        assert md.url == "https://example.com"
        assert md.markdown and len(md.markdown) > 50

    @pytest.mark.asyncio
    async def test_fit_markdown(self, crawler):
        md = await crawler.markdown("https://httpbin.org/html", strategy="http", fit=True)
        assert md.success is True
        assert md.fit_markdown is not None

    @pytest.mark.asyncio
    async def test_include_fields(self, crawler):
        md = await crawler.markdown(
            "https://books.toscrape.com", strategy="http",
            include=["links", "media", "metadata"],
        )
        assert md.success is True
        assert isinstance(md.links, dict)
        assert isinstance(md.media, dict)
        assert isinstance(md.metadata, dict)

    @pytest.mark.asyncio
    async def test_credits_returned(self, crawler):
        md = await crawler.markdown("https://example.com", strategy="http")
        assert md.usage is not None
        assert isinstance(md.usage, WrapperUsage)
        assert md.usage.credits_used > 0

    @pytest.mark.asyncio
    async def test_crawler_config_passthrough(self, crawler):
        md = await crawler.markdown(
            "https://books.toscrape.com", strategy="browser",
            crawler_config={"css_selector": "article.product_pod", "wait_until": "domcontentloaded"},
        )
        assert md.success is True
        assert md.markdown and len(md.markdown) > 0

    @pytest.mark.asyncio
    async def test_browser_config_passthrough(self, crawler):
        md = await crawler.markdown(
            "https://httpbin.org/headers", strategy="browser",
            browser_config={"headers": {"X-SDK-Test": "wrapper-test"}},
        )
        assert md.success is True

    @pytest.mark.asyncio
    async def test_http_strategy(self, crawler):
        md = await crawler.markdown("https://example.com", strategy="http")
        assert md.success is True
        assert md.duration_ms > 0

    @pytest.mark.asyncio
    async def test_bypass_cache(self, crawler):
        md = await crawler.markdown("https://example.com", strategy="http", bypass_cache=True)
        assert md.success is True


# =============================================================================
# SCREENSHOT
# =============================================================================


class TestScreenshot:
    @pytest.mark.asyncio
    async def test_basic(self, crawler):
        ss = await crawler.screenshot("https://example.com")
        assert isinstance(ss, ScreenshotResponse)
        assert ss.success is True
        assert ss.screenshot and len(ss.screenshot) > 1000

    @pytest.mark.asyncio
    async def test_pdf(self, crawler):
        ss = await crawler.screenshot("https://example.com", pdf=True)
        assert ss.success is True
        assert ss.pdf and len(ss.pdf) > 1000

    @pytest.mark.asyncio
    async def test_viewport_only(self, crawler):
        ss = await crawler.screenshot("https://example.com", full_page=False)
        assert ss.success is True
        assert ss.screenshot and len(ss.screenshot) > 1000

    @pytest.mark.asyncio
    async def test_with_wait_for(self, crawler):
        ss = await crawler.screenshot("https://books.toscrape.com", wait_for=".product_pod")
        assert ss.success is True

    @pytest.mark.asyncio
    async def test_crawler_config(self, crawler):
        ss = await crawler.screenshot(
            "https://example.com",
            crawler_config={"page_timeout": 15000},
        )
        assert ss.success is True


# =============================================================================
# EXTRACT
# =============================================================================


class TestExtract:
    @pytest.mark.asyncio
    async def test_auto(self, crawler):
        data = await crawler.extract(
            "https://books.toscrape.com",
            query="extract all books with title and price",
        )
        assert isinstance(data, ExtractResponse)
        assert data.success is True
        assert data.data and len(data.data) > 0
        assert data.method_used in ("css_schema", "llm")

    @pytest.mark.asyncio
    async def test_llm_method(self, crawler):
        data = await crawler.extract(
            "https://example.com", method="llm",
            query="what is this page about",
        )
        assert data.success is True
        assert data.method_used == "llm"

    @pytest.mark.asyncio
    async def test_with_json_example(self, crawler):
        data = await crawler.extract(
            "https://books.toscrape.com",
            json_example={"title": "...", "price": "$0.00"},
            query="extract books",
        )
        assert data.success is True


# =============================================================================
# MAP
# =============================================================================


class TestMap:
    @pytest.mark.asyncio
    async def test_basic(self, crawler):
        result = await crawler.map("https://crawl4ai.com", max_urls=10)
        assert isinstance(result, MapResponse)
        assert result.success is True
        assert result.total_urls > 0
        assert result.domain == "crawl4ai.com"

    @pytest.mark.asyncio
    async def test_with_query(self, crawler):
        result = await crawler.map(
            "https://docs.crawl4ai.com",
            query="extraction", max_urls=5, score_threshold=0.1,
        )
        assert result.success is True
        if result.urls:
            for u in result.urls:
                if u.relevance_score is not None:
                    assert u.relevance_score >= 0.1


# =============================================================================
# SITE CRAWL
# =============================================================================


class TestSiteCrawl:
    @pytest.mark.asyncio
    async def test_basic(self, crawler):
        result = await crawler.crawl_site(
            "https://books.toscrape.com", max_pages=3, strategy="http",
        )
        assert isinstance(result, SiteCrawlResponse)
        assert result.job_id
        assert result.strategy == "map"

    @pytest.mark.asyncio
    async def test_bfs(self, crawler):
        result = await crawler.crawl_site(
            "https://httpbin.org", max_pages=3, discovery="bfs",
            max_depth=1, strategy="http",
        )
        assert result.job_id

    @pytest.mark.asyncio
    async def test_with_criteria(self, crawler):
        """AI-assisted: criteria triggers LLM config generation."""
        result = await crawler.crawl_site(
            "https://books.toscrape.com",
            criteria="book listing pages",
            max_pages=3,
            strategy="http",
        )
        assert result.job_id
        assert result.generated_config is not None
        assert isinstance(result.generated_config, GeneratedConfig)
        assert result.generated_config.scan  # has a scan config
        assert result.generated_config.reasoning  # non-empty reasoning
        # fallback should be False on stage unless LLM is unreachable
        assert result.generated_config.fallback is False

    @pytest.mark.asyncio
    async def test_with_criteria_and_extract(self, crawler):
        """Flagship flow: criteria + extract → schema generated + echoed back."""
        result = await crawler.crawl_site(
            "https://books.toscrape.com",
            criteria="book listing pages",
            max_pages=3,
            strategy="http",
            extract={
                "query": "book title and price",
                "json_example": {"title": "...", "price": "£0.00"},
                "method": "auto",
            },
        )
        assert result.job_id
        assert result.generated_config is not None
        assert result.extraction_method_used in ("css_schema", "llm")
        # When CSS path is picked, schema_used should be populated
        if result.extraction_method_used == "css_schema":
            assert result.schema_used is not None
            assert "fields" in result.schema_used

    @pytest.mark.asyncio
    async def test_scan_config_dataclass(self, crawler):
        """SiteScanConfig dataclass accepted alongside plain dict."""
        scan_cfg = SiteScanConfig(
            mode="map",
            patterns=["*/catalogue/*"],
            score_threshold=0.2,
        )
        result = await crawler.crawl_site(
            "https://books.toscrape.com",
            max_pages=3,
            strategy="http",
            scan=scan_cfg,
        )
        assert result.job_id

    @pytest.mark.asyncio
    async def test_unified_polling(self, crawler):
        """get_site_crawl_job returns unified scan+crawl status with phase."""
        result = await crawler.crawl_site(
            "https://books.toscrape.com",
            criteria="book listings",
            max_pages=3,
            strategy="http",
        )
        # Poll a few times to verify the shape
        for _ in range(5):
            status = await crawler.get_site_crawl_job(result.job_id)
            assert isinstance(status, SiteCrawlJobStatus)
            assert status.job_id == result.job_id
            assert status.phase in ("scan", "crawl", "done")
            assert status.progress is not None
            if status.is_complete:
                break
            await asyncio.sleep(3)

    @pytest.mark.asyncio
    async def test_include_without_markdown(self, crawler):
        """include=['links'] should trigger markdown stripping at trim time."""
        result = await crawler.crawl_site(
            "https://books.toscrape.com",
            criteria="book listings",
            max_pages=3,
            strategy="http",
            include=["links"],
        )
        assert result.job_id


# =============================================================================
# SCAN (AI-assisted)
# =============================================================================


class TestScan:
    @pytest.mark.asyncio
    async def test_basic(self, crawler):
        """Legacy: no criteria → plain DomainMapper sync flow."""
        result = await crawler.scan("https://crawl4ai.com", max_urls=10)
        assert isinstance(result, ScanResult)
        assert result.success is True
        assert result.total_urls > 0
        assert result.domain == "crawl4ai.com"

    @pytest.mark.asyncio
    async def test_with_criteria(self, crawler):
        """AI-assisted: criteria triggers LLM config generator."""
        result = await crawler.scan(
            "https://docs.crawl4ai.com",
            criteria="API reference and core documentation pages",
            max_urls=20,
        )
        assert result.success is True
        assert result.mode_used in ("map", "deep")
        assert result.generated_config is not None
        assert isinstance(result.generated_config, GeneratedConfig)
        assert result.generated_config.reasoning

    @pytest.mark.asyncio
    async def test_with_scan_overrides(self, crawler):
        """Explicit scan overrides merged on top of LLM output."""
        result = await crawler.scan(
            "https://docs.crawl4ai.com",
            criteria="documentation pages",
            scan={"patterns": ["*/core/*"]},
            max_urls=10,
        )
        assert result.success is True
        assert result.mode_used == "map"

    @pytest.mark.asyncio
    async def test_scan_config_dataclass(self, crawler):
        """SiteScanConfig dataclass accepted as `scan` parameter."""
        cfg = SiteScanConfig(mode="map", patterns=["*/docs/*"])
        result = await crawler.scan(
            "https://docs.crawl4ai.com",
            scan=cfg,
            max_urls=10,
        )
        assert result.success is True

    @pytest.mark.asyncio
    async def test_deep_mode_async(self, crawler):
        """Deep mode returns job_id + is_async=True."""
        result = await crawler.scan(
            "https://httpbin.org",
            scan={"mode": "deep", "max_depth": 1},
            max_urls=5,
        )
        # Deep mode should kick off an async job
        assert result.is_async is True
        assert result.job_id is not None
        assert result.mode_used == "deep"
        # And we should be able to poll it
        job = await crawler.get_scan_job(result.job_id)
        assert isinstance(job, ScanJobStatus)
        assert job.job_id == result.job_id
        # Cancel it so we don't tie up worker slots
        cancelled = await crawler.cancel_scan_job(result.job_id)
        assert cancelled.job_id == result.job_id


# =============================================================================
# ASYNC LIFECYCLE
# =============================================================================


class TestAsyncLifecycle:
    @pytest.mark.asyncio
    async def test_markdown_many_no_wait(self, crawler):
        job = await crawler.markdown_many(
            ["https://example.com", "https://httpbin.org/html"],
            strategy="http",
        )
        assert isinstance(job, WrapperJob)
        assert job.job_id
        assert job.urls_count == 2

    @pytest.mark.asyncio
    async def test_markdown_many_with_wait(self, crawler):
        job = await crawler.markdown_many(
            ["https://example.com", "https://httpbin.org/html"],
            strategy="http", wait=True, timeout=60,
        )
        assert job.is_complete
        assert job.progress.completed == 2

    @pytest.mark.asyncio
    async def test_poll_and_list(self, crawler):
        job = await crawler.markdown_many(
            ["https://example.com"], strategy="http",
        )
        # Wait a bit then poll
        await asyncio.sleep(5)
        status = await crawler.get_markdown_job(job.job_id)
        assert status.job_id == job.job_id

        # List
        jobs = await crawler.list_markdown_jobs(limit=5)
        assert len(jobs) > 0

    @pytest.mark.asyncio
    async def test_screenshot_many(self, crawler):
        job = await crawler.screenshot_many(
            ["https://example.com"], wait=True, timeout=60,
        )
        assert job.is_complete

    @pytest.mark.asyncio
    async def test_extract_many(self, crawler):
        # 0.7.0: extract_many takes a base url + optional extra_urls
        # (was urls=[...]). The base URL acts as the schema template in
        # css_schema mode.
        job = await crawler.extract_many(
            url="https://example.com", method="llm",
            query="summarize", wait=True, timeout=120,
        )
        assert job.is_complete


# =============================================================================
# JOB CANCEL
# =============================================================================


class TestJobCancel:
    @pytest.mark.asyncio
    async def test_cancel_markdown_job(self, crawler):
        job = await crawler.markdown_many(
            ["https://httpbin.org/delay/10"] * 3, strategy="http",
        )
        await asyncio.sleep(1)
        cancelled = await crawler.cancel_markdown_job(job.job_id)
        assert cancelled is True

        status = await crawler.get_markdown_job(job.job_id)
        assert status.status == "cancelled"


# =============================================================================
# NAMESPACE ISOLATION
# =============================================================================


class TestNamespaceIsolation:
    @pytest.mark.asyncio
    async def test_cross_namespace_404(self, crawler):
        # Create a markdown job
        job = await crawler.markdown_many(["https://example.com"], strategy="http")

        await asyncio.sleep(2)

        # Try to access it via screenshot namespace -- should 404
        with pytest.raises(NotFoundError):
            await crawler.get_screenshot_job(job.job_id)

    @pytest.mark.asyncio
    async def test_own_namespace_works(self, crawler):
        job = await crawler.markdown_many(["https://example.com"], strategy="http")
        await asyncio.sleep(2)
        status = await crawler.get_markdown_job(job.job_id)
        assert status.job_id == job.job_id


# =============================================================================
# ERROR CASES
# =============================================================================


class TestErrors:
    @pytest.mark.asyncio
    async def test_bad_auth(self):
        async with AsyncWebCrawler(api_key="sk_live_bad_key_000", base_url=BASE_URL) as c:
            with pytest.raises(AuthenticationError):
                await c.markdown("https://example.com")

    @pytest.mark.asyncio
    async def test_nonexistent_job(self, crawler):
        with pytest.raises(NotFoundError):
            await crawler.get_markdown_job("job_doesnotexist000000")

    @pytest.mark.asyncio
    async def test_extract_auto_batch_rejected(self, crawler):
        with pytest.raises(ValueError, match="AUTO"):
            await crawler.extract_many(["https://example.com"], method="auto")


# =============================================================================
# ADVERSARIAL / INJECTION
# =============================================================================


class TestAdversarial:
    @pytest.mark.asyncio
    async def test_sql_in_query(self, crawler):
        """SQL injection in query should not crash."""
        data = await crawler.extract(
            "https://example.com", method="llm",
            query="'; DROP TABLE users; --",
        )
        # Should complete without server error (may return no data)
        assert isinstance(data, ExtractResponse)

    @pytest.mark.asyncio
    async def test_xss_in_config(self, crawler):
        """XSS in crawler_config should not crash."""
        md = await crawler.markdown(
            "https://example.com", strategy="http",
            crawler_config={"css_selector": "<script>alert(1)</script>"},
        )
        assert isinstance(md, MarkdownResponse)

    @pytest.mark.asyncio
    async def test_huge_payload(self, crawler):
        """Huge json_example should not crash."""
        huge = {f"field_{i}": f"value_{i}" for i in range(200)}
        data = await crawler.extract(
            "https://example.com", method="llm",
            json_example=huge, query="extract everything",
        )
        assert isinstance(data, ExtractResponse)

    @pytest.mark.asyncio
    async def test_unicode_url(self, crawler):
        """Unicode in URL should not crash."""
        md = await crawler.markdown("https://example.com/\u00e9\u00e8\u00ea", strategy="http")
        assert isinstance(md, MarkdownResponse)


# =============================================================================
# FREE TIER
# =============================================================================


class TestFreeTier:
    @pytest.mark.asyncio
    async def test_markdown_free_tier(self, free_crawler):
        md = await free_crawler.markdown("https://example.com", strategy="http")
        assert md.success is True
        assert md.usage and md.usage.credits_used > 0

    @pytest.mark.asyncio
    async def test_screenshot_free_tier(self, free_crawler):
        ss = await free_crawler.screenshot("https://example.com")
        assert ss.success is True
