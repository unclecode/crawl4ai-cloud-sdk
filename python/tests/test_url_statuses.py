"""End-to-end tests for the 0.8.1 multi-URL fan-out + url_statuses[] flow.

Covers:
- Wrapper async GET responses parse `url_statuses` and `download_url`.
- get_per_url_result fetches one URL's CrawlResult (recipe-agnostic).
- wait=True hydrates job.results from per-URL S3 (auto-hydrate path).
- Single-URL submits leave url_statuses as None.

Tests run against the configured base URL — defaults to stage. Skip with:
    pytest -k "not url_statuses"
"""
import os
import pytest

from crawl4ai_cloud import AsyncWebCrawler, UrlStatus, WrapperJob
from crawl4ai_cloud.models import CrawlResult


API_KEY = os.getenv(
    "CRAWL4AI_API_KEY",
    "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ",
)
BASE_URL = os.getenv("CRAWL4AI_API_URL", "https://stage.crawl4ai.com")


# ─── Pure unit tests (no network) ────────────────────────────────────────


class TestWrapperJobShape:
    """Parse the new fields off a synthetic API response."""

    def test_url_statuses_parsed(self):
        data = {
            "job_id": "job_abc",
            "status": "completed",
            "progress": {"total": 2, "completed": 2, "failed": 0},
            "progress_percent": 100,
            "url_statuses": [
                {"index": 0, "url": "https://a.com", "status": "done", "duration_ms": 100, "error": None},
                {"index": 1, "url": "https://b.com", "status": "failed", "duration_ms": 500, "error": "timeout"},
            ],
            "download_url": "https://...zip",
            "created_at": "2026-05-03T00:00:00Z",
        }
        job = WrapperJob.from_dict(data)
        assert job.url_statuses is not None
        assert len(job.url_statuses) == 2
        assert isinstance(job.url_statuses[0], UrlStatus)
        assert job.url_statuses[0].status == "done"
        assert job.url_statuses[0].duration_ms == 100
        assert job.url_statuses[1].error == "timeout"
        assert job.download_url == "https://...zip"
        # results stays None until SDK hydrates it (e.g. wait=True)
        assert job.results is None

    def test_single_url_leaves_statuses_none(self):
        data = {
            "job_id": "job_single",
            "status": "completed",
            "progress": {"total": 1, "completed": 1, "failed": 0},
            "created_at": "2026-05-03T00:00:00Z",
        }
        job = WrapperJob.from_dict(data)
        assert job.url_statuses is None
        assert job.results is None


# ─── E2E tests (hit stage) ───────────────────────────────────────────────


@pytest.mark.asyncio
class TestUrlStatusesE2E:

    async def test_scrape_many_wait_true_hydrates_results(self):
        """wait=True must populate job.results from per-URL S3."""
        urls = ["https://example.com", "https://example.org"]
        async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
            job = await crawler.scrape_many(
                urls, strategy="http", wait=True, timeout=60,
            )
            assert job.is_complete
            # url_statuses populated by GET response
            assert job.url_statuses is not None
            assert len(job.url_statuses) == 2
            # results auto-hydrated by _wait_wrapper_job
            assert job.results is not None
            assert len(job.results) == 2
            assert all(isinstance(r, CrawlResult) for r in job.results)
            # At least one must have markdown if scrape succeeded
            successful = [r for r in job.results if r.success]
            if successful:
                assert successful[0].markdown is not None

    async def test_extract_many_wait_true_hydrates_extracted_content(self):
        """Extract should hydrate results with extracted_content per URL."""
        async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
            job = await crawler.extract_many(
                url="https://example.com",
                extra_urls=["https://example.org"],
                strategy="http",
                query="page title",
                method="auto",
                wait=True,
                timeout=120,
            )
            assert job.is_complete
            assert job.url_statuses is not None
            assert len(job.url_statuses) == 2
            assert job.results is not None
            assert len(job.results) == 2
            # Extract jobs land extracted_content (not markdown) on the
            # per-URL CrawlResult
            successful = [r for r in job.results if r.success]
            assert successful, "expected at least one successful URL"

    async def test_get_per_url_result_recipe_agnostic(self):
        """Submit a scrape_many, then directly call get_per_url_result."""
        urls = ["https://example.com", "https://example.org"]
        async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
            job = await crawler.scrape_many(
                urls, strategy="http", wait=True, timeout=60,
            )
            # Direct per-URL fetch — same path users would call
            result = await crawler.get_per_url_result(job.job_id, 0)
            assert isinstance(result, CrawlResult)
            assert result.url
            # Either success=True with markdown, or success=False with error
            if result.success:
                assert result.markdown is not None

    async def test_single_url_submit_no_url_statuses(self):
        """Single-URL submit shouldn't have url_statuses or hydrated results."""
        async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
            job = await crawler.scrape_many(
                ["https://example.com"], strategy="http",
                wait=True, timeout=60,
            )
            assert job.is_complete
            # Single-URL: API returns url_statuses=None
            assert job.url_statuses is None
            # And without url_statuses, hydrate is a no-op
            assert job.results is None
