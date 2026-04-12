"""
Enrich API E2E tests -- runs against stage.crawl4ai.com.

Usage:
    pytest tests/test_enrich_e2e.py -v -s
"""

import asyncio
import os
import pytest
import pytest_asyncio

# SDK under test
from crawl4ai_cloud import (
    AsyncWebCrawler,
    EnrichJobStatus,
    EnrichRow,
    EnrichFieldSource,
    EnrichJobProgress,
)

API_KEY = os.environ.get("CRAWL4AI_API_KEY", "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ")
BASE_URL = os.environ.get("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")


@pytest_asyncio.fixture
async def crawler():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as c:
        yield c


class TestEnrichHappyPath:
    """Core enrichment functionality."""

    @pytest.mark.asyncio
    async def test_basic_enrich(self, crawler):
        """Enrich a single URL with 2 fields, depth 0, no search."""
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            schema=[
                {"name": "Company Name"},
                {"name": "Email", "description": "contact email"},
            ],
            max_depth=0,
            enable_search=False,
            strategy="browser",
            timeout=120,
        )

        assert isinstance(result, EnrichJobStatus)
        assert result.is_complete
        assert result.is_successful
        assert result.rows is not None
        assert len(result.rows) == 1

        row = result.rows[0]
        assert isinstance(row, EnrichRow)
        assert row.url == "https://kidocode.com"
        assert row.fields.get("Company Name"), "Company Name should be found"
        assert row.status in ("complete", "partial")
        assert row.depth_used == 0

    @pytest.mark.asyncio
    async def test_enrich_with_depth(self, crawler):
        """Enrich with depth=1, should follow links to find more fields."""
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            schema=[
                {"name": "Company Name"},
                {"name": "Email", "description": "primary contact email"},
                {"name": "Phone", "description": "phone number"},
            ],
            max_depth=1,
            max_links=3,
            enable_search=False,
            timeout=120,
        )

        assert result.is_complete
        assert len(result.rows) == 1
        row = result.rows[0]
        assert row.fields.get("Company Name")
        # With depth 1, should find more fields
        found = sum(1 for v in row.fields.values() if v)
        assert found >= 2, f"Expected at least 2 fields, got {found}"

    @pytest.mark.asyncio
    async def test_enrich_multiple_urls(self, crawler):
        """Enrich 2 URLs, verify both return rows."""
        result = await crawler.enrich(
            urls=["https://kidocode.com", "https://httpbin.org"],
            schema=[{"name": "Title", "description": "page or company title"}],
            max_depth=0,
            enable_search=False,
            timeout=120,
        )

        assert result.is_complete
        assert result.progress.total == 2
        assert result.rows is not None
        assert len(result.rows) == 2
        urls = {r.url for r in result.rows}
        assert "https://kidocode.com" in urls or "https://httpbin.org" in urls


class TestEnrichSourceAttribution:
    """Verify source attribution on enrichment rows."""

    @pytest.mark.asyncio
    async def test_sources_present(self, crawler):
        """Every found field should have a source with method."""
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            schema=[{"name": "Company Name"}, {"name": "Email"}],
            max_depth=0,
            enable_search=False,
            timeout=120,
        )

        row = result.rows[0]
        for field_name, value in row.fields.items():
            if value:
                assert field_name in row.sources, f"Missing source for {field_name}"
                src = row.sources[field_name]
                assert isinstance(src, EnrichFieldSource)
                assert src.method in ("direct", "depth", "search")
                assert src.url, f"Source URL empty for {field_name}"


class TestEnrichJobManagement:
    """Job lifecycle: create, poll, list, cancel."""

    @pytest.mark.asyncio
    async def test_fire_and_forget(self, crawler):
        """Create job without waiting, then poll manually."""
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            schema=[{"name": "Company Name"}],
            max_depth=0,
            enable_search=False,
            wait=False,
        )

        assert isinstance(result, EnrichJobStatus)
        assert result.job_id.startswith("enr_")
        assert result.status == "pending"

        # Poll until done
        for _ in range(30):
            status = await crawler.get_enrich_job(result.job_id)
            if status.is_complete:
                break
            await asyncio.sleep(2)

        assert status.is_complete
        assert status.rows is not None

    @pytest.mark.asyncio
    async def test_list_jobs(self, crawler):
        """List jobs should return recent enrichment jobs."""
        jobs = await crawler.list_enrich_jobs(limit=5)
        assert isinstance(jobs, list)
        # We created jobs in earlier tests, so there should be at least one
        assert len(jobs) >= 1
        assert all(hasattr(j, "job_id") for j in jobs)

    @pytest.mark.asyncio
    async def test_cancel_job(self, crawler):
        """Create a job and cancel it immediately."""
        result = await crawler.enrich(
            urls=["https://example.com", "https://httpbin.org", "https://kidocode.com"],
            schema=[
                {"name": "Title"},
                {"name": "Description", "description": "page description"},
                {"name": "Email"},
            ],
            max_depth=1,
            enable_search=True,
            wait=False,
        )

        assert result.job_id.startswith("enr_")

        # Cancel
        cancelled = await crawler.cancel_enrich_job(result.job_id)
        assert cancelled is True

        # Verify cancelled
        status = await crawler.get_enrich_job(result.job_id)
        assert status.status == "cancelled"


class TestEnrichEdgeCases:
    """Error handling and edge cases."""

    @pytest.mark.asyncio
    async def test_progress_tracking(self, crawler):
        """Verify progress fields are populated."""
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            schema=[{"name": "Company Name"}],
            max_depth=0,
            enable_search=False,
            timeout=120,
        )

        assert isinstance(result.progress, EnrichJobProgress)
        assert result.progress.total == 1
        assert result.progress.completed + result.progress.failed == 1
        assert result.progress_percent == 100

    @pytest.mark.asyncio
    async def test_token_usage_tracked(self, crawler):
        """Verify token usage is reported per row."""
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            schema=[{"name": "Company Name"}, {"name": "Email"}],
            max_depth=0,
            enable_search=False,
            timeout=120,
        )

        row = result.rows[0]
        assert row.token_usage is not None
        assert row.token_usage.get("total_tokens", 0) > 0
