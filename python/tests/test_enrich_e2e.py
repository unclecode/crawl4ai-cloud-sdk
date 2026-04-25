"""
Enrich v2 E2E tests — runs against stage.crawl4ai.com.

Covers the seven-method surface:
    enrich(...)          POST /v1/enrich/async
    get_enrich_job       GET  /v1/enrich/jobs/{id}
    wait_enrich_job      poll loop, optional `until=` phase
    resume_enrich_job    POST /v1/enrich/jobs/{id}/continue
    stream_enrich_job    GET  /v1/enrich/jobs/{id}/stream  (SSE)
    cancel_enrich_job    DELETE /v1/enrich/jobs/{id}
    list_enrich_jobs     GET  /v1/enrich/jobs

Run:
    pytest tests/test_enrich_e2e.py -v -s
"""

import asyncio
import os

import pytest
import pytest_asyncio

from crawl4ai_cloud import (
    AsyncWebCrawler,
    EnrichJobStatus,
    EnrichRow,
    EnrichPlan,
    EnrichUrlCandidate,
    EnrichUsage,
    EnrichLlmBucket,
    EnrichJobListItem,
    EnrichEvent,
)

API_KEY = os.environ.get(
    "CRAWL4AI_API_KEY",
    "sk_live_V89kxHtmkxw0jJORu_sWzyuvGw6TKHaJhoNGK8gGdqU",
)
BASE_URL = os.environ.get("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")


@pytest_asyncio.fixture
async def crawler():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as c:
        yield c


# ─── 1. URLs-only mode (fastest path; skips planning + URL resolution) ───

class TestUrlsOnly:
    """Pre-resolved URLs → extraction only. Fast, deterministic."""

    @pytest.mark.asyncio
    async def test_single_url_two_features(self, crawler):
        result = await crawler.enrich(
            urls=["https://kidocode.com"],
            features=[
                {"name": "company_name"},
                {"name": "contact_email", "description": "primary contact email"},
            ],
            strategy="http",
            wait=True,
            timeout=180,
        )
        assert isinstance(result, EnrichJobStatus)
        assert result.is_complete, f"job did not complete: {result.status} ({result.error})"
        assert result.is_successful, f"job ended in {result.status}: {result.error}"
        assert result.rows is not None and len(result.rows) >= 1
        row = result.rows[0]
        assert isinstance(row, EnrichRow)
        # URLs-mode: input_key may be None and url is set
        assert row.url == "https://kidocode.com" or row.group_id == "https://kidocode.com"
        # Should find at least company_name
        assert row.fields, f"no fields extracted: {row.error}"
        assert any(k.lower().startswith("company") for k in row.fields), \
            f"company_name missing from {list(row.fields)}"
        # Usage envelope sanity
        assert isinstance(result.usage, EnrichUsage)
        assert result.usage.crawls >= 1
        assert "extract" in result.usage.llm_tokens_by_purpose
        bucket = result.usage.llm_tokens_by_purpose["extract"]
        assert isinstance(bucket, EnrichLlmBucket)
        assert bucket.input > 0 and bucket.output > 0

    @pytest.mark.asyncio
    async def test_string_features_shorthand(self, crawler):
        """String shortcut for features: ['x', 'y'] instead of [{'name': 'x'}, ...]."""
        result = await crawler.enrich(
            urls=["https://example.com"],
            features=["title", "description"],
            strategy="http",
            wait=True,
            timeout=120,
        )
        assert result.is_complete
        # example.com is tiny; we just verify the request was accepted and completed
        assert result.rows is not None and len(result.rows) >= 1


# ─── 2. Job lifecycle: get / list / cancel ───────────────────────────────

class TestJobLifecycle:

    @pytest.mark.asyncio
    async def test_fire_and_forget_then_get(self, crawler):
        job = await crawler.enrich(
            urls=["https://kidocode.com"],
            features=[{"name": "company_name"}],
            strategy="http",
            wait=False,
        )
        assert job.job_id.startswith("enr_")
        assert job.status in ("queued", "extracting", "merging", "completed")
        # Poll once
        latest = await crawler.get_enrich_job(job.job_id)
        assert latest.job_id == job.job_id

    @pytest.mark.asyncio
    async def test_wait_until_terminal(self, crawler):
        job = await crawler.enrich(
            urls=["https://example.com"],
            features=[{"name": "title"}],
            strategy="http",
            wait=False,
        )
        terminal = await crawler.wait_enrich_job(job.job_id, timeout=120)
        assert terminal.is_complete
        assert terminal.is_successful

    @pytest.mark.asyncio
    async def test_list_jobs(self, crawler):
        jobs = await crawler.list_enrich_jobs(limit=5)
        assert isinstance(jobs, list)
        # We created jobs in earlier tests, so there should be at least one
        assert len(jobs) >= 1
        assert all(isinstance(j, EnrichJobListItem) for j in jobs)
        assert all(j.job_id.startswith("enr_") for j in jobs)

    @pytest.mark.asyncio
    async def test_cancel_running_job(self, crawler):
        # Start a query-based job — slowest path so we have time to cancel
        job = await crawler.enrich(
            query="top BBQ restaurants in Austin Texas with outdoor seating",
            country="us",
            top_k_per_entity=2,
            wait=False,
        )
        assert job.job_id.startswith("enr_")
        # Cancel immediately
        cancelled = await crawler.cancel_enrich_job(job.job_id)
        assert cancelled is True
        # Verify status flipped
        await asyncio.sleep(2)
        latest = await crawler.get_enrich_job(job.job_id)
        assert latest.status == "cancelled"


# ─── 3. Review flow: pause at plan_ready, resume with edits ──────────────

class TestReviewFlow:

    @pytest.mark.asyncio
    async def test_pause_at_plan_ready_then_resume(self, crawler):
        job = await crawler.enrich(
            query="best Italian restaurants in Brooklyn New York",
            country="us",
            top_k_per_entity=1,
            auto_confirm_plan=False,    # pause here
            auto_confirm_urls=True,     # let it run after we resume
            wait=False,
        )
        # Wait for the planning phase to land on plan_ready
        paused = await crawler.wait_enrich_job(
            job.job_id, until="plan_ready", timeout=120,
        )
        assert paused.status == "plan_ready"
        assert paused.is_paused
        assert paused.plan is not None
        plan = paused.plan
        assert isinstance(plan, EnrichPlan)
        assert len(plan.entities) >= 1
        assert len(plan.features) >= 1
        # Plan-phase usage should be populated
        assert "plan_intent" in paused.usage.llm_tokens_by_purpose

        # Trim entities so we don't burn 10 crawls in tests, but keep
        # multiple features so we're not at the mercy of a single page
        # legitimately not having one specific column (data-noise → flaky).
        edited_features = [
            {"name": "name", "description": "restaurant name"},
            {"name": "address"},
            {"name": "phone"},
        ]
        edited_entities = [{"name": plan.entities[0].name}]
        resumed = await crawler.resume_enrich_job(
            job.job_id,
            entities=edited_entities,
            features=edited_features,
        )
        # Server returns updated status — should be past plan_ready
        assert resumed.status not in ("plan_ready",)

        # Wait to terminal
        final = await crawler.wait_enrich_job(job.job_id, timeout=300)
        assert final.is_complete
        # Plan was edited → only one entity → at most one row
        assert final.rows is not None and len(final.rows) <= 1, \
            f"expected ≤1 row, got {final.rows}"
        # AT LEAST ONE row should have populated fields — would have caught
        # the worker P0 (features-not-plumbed) where the job completed but
        # every row had fields={}.
        populated = [r for r in final.rows if r.fields]
        assert populated, (
            f"all {len(final.rows)} rows came back with empty fields — "
            f"likely a regression of the features-not-plumbed bug. "
            f"row[0].error={final.rows[0].error!r}"
        )

        # ── tier + reason on URL candidates (added in 0.6.1) ──
        # The LLM rerank should have populated these on every URL that made
        # it past the resolve phase. Skip the check if no URLs surfaced.
        upe = final.urls_per_entity or {}
        for entity, urls in upe.items():
            for c in urls:
                assert isinstance(c, EnrichUrlCandidate)
                assert c.tier is not None and 0.0 <= c.tier <= 1.0, \
                    f"missing/invalid tier for {entity}/{c.url}: {c.tier!r}"
                assert c.reason and isinstance(c.reason, str), \
                    f"missing reason for {entity}/{c.url}: {c.reason!r}"


# ─── 4. SSE streaming ────────────────────────────────────────────────────

class TestStream:

    @pytest.mark.asyncio
    async def test_stream_yields_snapshot_and_complete(self, crawler):
        # Tiny URLs-mode job so the stream finishes quickly
        job = await crawler.enrich(
            urls=["https://example.com"],
            features=[{"name": "title"}],
            strategy="http",
            wait=False,
        )

        seen_types: list[str] = []
        last_status: str | None = None

        async def collect():
            async for event in crawler.stream_enrich_job(job.job_id):
                assert isinstance(event, EnrichEvent)
                seen_types.append(event.type)
                if event.type == "snapshot":
                    assert event.snapshot is not None
                    assert event.snapshot.job_id == job.job_id
                if event.type in ("phase", "complete"):
                    nonlocal_status = event.status or (
                        event.snapshot.status if event.snapshot else None
                    )
                    if nonlocal_status:
                        # capture in outer scope below
                        pass
                if event.type == "complete":
                    return

        # Hard cap so a stuck stream doesn't hang the test suite
        await asyncio.wait_for(collect(), timeout=120)

        assert "snapshot" in seen_types, f"missing snapshot event: {seen_types}"
        assert "complete" in seen_types, f"stream did not complete: {seen_types}"
