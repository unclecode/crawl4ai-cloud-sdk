"""
SDK 0.7.0 changes — real e2e tests against stage. No mocks.

Covers:
  - crawler.scrape() / scrape_many() — new canonical names
  - crawler.markdown() / markdown_many() — deprecated aliases (still work, warn)
  - crawler.extract_many() — new shape (url + extra_urls, AUTO allowed)
  - sources= kwarg on scan() + map() (legacy mode= still works + warns)
  - Composable scan + extract chain (the recommended whole-site pattern)
  - crawler.crawl_site() / deep_crawl() — deprecated, still respond, warn

Run:
    cd crawl4ai-cloud-sdk/python
    python -m pytest tests/test_v07_changes.py -v --timeout=180
"""

import os
import warnings

import pytest
import pytest_asyncio

from crawl4ai_cloud import (
    AsyncWebCrawler,
    MarkdownResponse,
    WrapperJob,
    MapResponse,
    ScanResult,
)

API_KEY = os.environ.get(
    "CRAWL4AI_API_KEY",
    "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ",
)
BASE_URL = os.environ.get("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")


@pytest_asyncio.fixture
async def crawler():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as c:
        yield c


# =============================================================================
# Scrape — new canonical name
# =============================================================================


class TestScrape:
    @pytest.mark.asyncio
    async def test_scrape_basic(self, crawler):
        r = await crawler.scrape("https://example.com", strategy="http")
        assert isinstance(r, MarkdownResponse)
        assert r.success
        assert r.markdown
        assert "Example" in r.markdown

    @pytest.mark.asyncio
    async def test_scrape_with_include(self, crawler):
        r = await crawler.scrape(
            "https://example.com",
            strategy="http",
            include=["links", "metadata"],
        )
        assert r.success
        assert r.links is not None
        assert r.metadata is not None

    @pytest.mark.asyncio
    async def test_scrape_many_async(self, crawler):
        job = await crawler.scrape_many(
            urls=["https://example.com", "https://httpbin.org/html"],
            strategy="http",
            wait=True,
            timeout=120,
        )
        assert isinstance(job, WrapperJob)
        assert job.is_complete


# =============================================================================
# Markdown alias — deprecated but still works
# =============================================================================


class TestMarkdownAlias:
    @pytest.mark.asyncio
    async def test_markdown_still_works(self, crawler):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            r = await crawler.markdown("https://example.com", strategy="http")
            assert r.success
            assert r.markdown
            # Verify a DeprecationWarning was emitted
            depr = [w for w in caught if issubclass(w.category, DeprecationWarning)]
            assert depr, "markdown() should emit DeprecationWarning"
            assert "scrape" in str(depr[0].message)

    @pytest.mark.asyncio
    async def test_markdown_many_still_works(self, crawler):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            job = await crawler.markdown_many(
                ["https://example.com"], strategy="http",
                wait=True, timeout=60,
            )
            assert job.is_complete
            depr = [w for w in caught if issubclass(w.category, DeprecationWarning)]
            assert depr


# =============================================================================
# Extract — new shape (url + extra_urls), AUTO allowed
# =============================================================================


class TestExtractMany:
    @pytest.mark.asyncio
    async def test_extract_single_url_auto(self, crawler):
        # Just a base URL, AUTO mode. Used to be rejected for batch-shaped calls.
        job = await crawler.extract_many(
            url="https://example.com",
            method="auto",
            wait=True,
            timeout=120,
        )
        assert job.is_complete

    @pytest.mark.asyncio
    async def test_extract_url_plus_extra_urls(self, crawler):
        # The new canonical shape: base + followers, schema generated once.
        job = await crawler.extract_many(
            url="https://example.com",
            extra_urls=["https://httpbin.org/html"],
            method="llm",
            query="summarize the page",
            wait=True,
            timeout=180,
        )
        assert job.is_complete

    @pytest.mark.asyncio
    async def test_extract_auto_was_previously_blocked(self, crawler):
        # Before 0.7.0, method="auto" raised ValueError on extract_many.
        # Confirm the block is gone.
        try:
            await crawler.extract_many(
                url="https://example.com",
                method="auto",
                wait=False,  # don't actually wait — just confirm submission
            )
        except ValueError as e:
            pytest.fail(f"AUTO method should no longer raise ValueError: {e}")


# =============================================================================
# sources= kwarg (replaces mode= which is now deprecated)
# =============================================================================


class TestSourcesKwarg:
    @pytest.mark.asyncio
    async def test_map_sources_primary(self, crawler):
        r = await crawler.map(
            "https://www.python.org",
            sources="primary",
            max_urls=5,
        )
        assert isinstance(r, MapResponse)
        assert r.success
        assert r.total_urls > 0

    @pytest.mark.asyncio
    async def test_map_legacy_mode_default(self, crawler):
        # mode="default" should still work (translated to sources="primary")
        # and emit a DeprecationWarning.
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            r = await crawler.map(
                "https://www.python.org",
                mode="default",
                max_urls=5,
            )
            assert r.success
            depr = [w for w in caught if issubclass(w.category, DeprecationWarning)]
            assert depr
            assert "sources" in str(depr[0].message)

    @pytest.mark.asyncio
    async def test_scan_sources_primary(self, crawler):
        r = await crawler.scan(
            "https://www.python.org",
            sources="primary",
            max_urls=5,
        )
        assert isinstance(r, ScanResult)
        assert r.total_urls > 0

    @pytest.mark.asyncio
    async def test_scan_legacy_mode_deep_translates(self, crawler):
        # mode="deep" should translate to sources="extended" + emit warning.
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            # Don't actually run — sources="extended" triggers slow Wayback+CC.
            # Just verify the kwarg is accepted + warning emitted.
            try:
                await crawler.scan(
                    "https://example.com",
                    mode="deep",
                    max_urls=3,
                    soft_404_detection=False,
                )
            finally:
                depr = [w for w in caught if issubclass(w.category, DeprecationWarning)]
                assert depr
                assert "sources" in str(depr[0].message)


# =============================================================================
# Composable chain — the recommended whole-site pattern
# =============================================================================


class TestChainPattern:
    @pytest.mark.asyncio
    async def test_scan_then_scrape_chain(self, crawler):
        # Step 1: scan to get URLs
        scan = await crawler.scan(
            "https://www.python.org",
            sources="primary",
            max_urls=3,
        )
        assert scan.total_urls > 0
        urls = [u.url for u in scan.urls[:3]]
        assert len(urls) >= 1

        # Step 2: pipe URLs to scrape_many
        job = await crawler.scrape_many(
            urls=urls,
            strategy="http",
            wait=True,
            timeout=120,
        )
        assert job.is_complete

    @pytest.mark.asyncio
    async def test_scan_then_extract_chain(self, crawler):
        # Step 1: scan
        scan = await crawler.scan(
            "https://www.python.org",
            sources="primary",
            max_urls=3,
        )
        urls = [u.url for u in scan.urls[:3]]
        assert urls

        # Step 2: extract with url + extra_urls (schema-once-applied-to-many)
        base, *rest = urls
        job = await crawler.extract_many(
            url=base,
            extra_urls=rest,
            method="auto",
            wait=True,
            timeout=180,
        )
        assert job.is_complete


# =============================================================================
# Deprecated endpoints — still respond, SDK warns
# =============================================================================


class TestDeprecatedEndpoints:
    @pytest.mark.asyncio
    async def test_crawl_site_warns(self, crawler):
        # Just confirm the deprecation warning fires; don't wait for the
        # full crawl to complete (it's a deprecated endpoint, not core path).
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            try:
                await crawler.crawl_site(
                    "https://example.com", max_pages=2,
                )
            except Exception:
                pass  # endpoint may behave oddly; we only care about the warning
            depr = [w for w in caught if issubclass(w.category, DeprecationWarning)]
            assert depr, "crawl_site() should emit DeprecationWarning"
            assert "scan" in str(depr[0].message).lower()

    @pytest.mark.asyncio
    async def test_deep_crawl_warns(self, crawler):
        with warnings.catch_warnings(record=True) as caught:
            warnings.simplefilter("always")
            try:
                await crawler.deep_crawl(
                    "https://example.com", strategy="bfs",
                    max_depth=1, max_urls=2,
                )
            except Exception:
                pass
            depr = [w for w in caught if issubclass(w.category, DeprecationWarning)]
            assert depr, "deep_crawl() should emit DeprecationWarning"
