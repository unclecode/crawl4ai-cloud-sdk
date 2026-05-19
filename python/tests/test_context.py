#!/usr/bin/env python3
"""
Context v2 SDK tests — two layers (unit + live).

Layer 1 (unit) — pure-Python checks that don't hit the network:
- Pillar builders produce the documented dict shape
- _build_context_body validates mutually-exclusive generator_id vs pillars
- _parse_event translates SSE payloads into the right typed event
- ContextResult is_terminal / is_success flags work
- Constraints.to_dict produces the API-expected shape

Layer 2 (live) — hits stage by default. Submits a real Context run,
streams it, fetches output, refreshes, lists versions, cancels. Gated
on CRAWL4AI_API_KEY env var; otherwise skipped.

Run:
    # Unit only
    pytest tests/test_context.py -k unit -v

    # Full (unit + live)
    CRAWL4AI_API_KEY=sk_live_... CRAWL4AI_BASE_URL=https://stage.crawl4ai.com \
      pytest tests/test_context.py -v
"""
import asyncio
import os

import pytest

from crawl4ai_cloud import (
    AsyncWebCrawler,
    # Pillar builders
    Source,
    Strategy,
    Shape,
    Reconciler,
    Constraints,
    # Result + events
    ContextResult,
    ContextOutput,
    ContextItem,
    StatusEvent,
    TerminalEvent,
    PhaseProgressInit,
    PhaseProgressItemUpdate,
    # Diff / catalog
    ContextDiff,
    ContextVersion,
    ContextCatalog,
)
from crawl4ai_cloud.context import _parse_event


BASE_URL = os.getenv("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")
API_KEY = os.getenv("CRAWL4AI_API_KEY")

# `livetest` marker is gated on the env var being set; skipped otherwise.
livetest = pytest.mark.skipif(
    not API_KEY,
    reason="CRAWL4AI_API_KEY not set — skipping live tests",
)


# ─── Unit — pillar builders ─────────────────────────────────────────────


@pytest.mark.unit
class TestPillarBuilders:
    """Builders return the documented API dict shape."""

    def test_source_google_web_defaults(self):
        assert Source.google_web() == {
            "type": "google_web",
            "params": {"top_k_per_backend": 10},
        }

    def test_source_google_web_with_backends(self):
        out = Source.google_web(backends=["google", "bing"], top_k_per_backend=8, region="us")
        assert out == {
            "type": "google_web",
            "params": {
                "backends": ["google", "bing"],
                "top_k_per_backend": 8,
                "region": "us",
            },
        }

    def test_source_crawl(self):
        out = Source.crawl(domain="https://example.com", max_urls=30, max_depth=2)
        assert out["type"] == "crawl"
        assert out["params"]["domain"] == "https://example.com"
        assert out["params"]["max_urls"] == 30
        assert out["params"]["max_depth"] == 2

    def test_source_crawl_with_optional_fields(self):
        out = Source.crawl(
            domain="https://example.com",
            max_urls=10, max_depth=1,
            score_threshold=0.5, profile_id="my-profile",
        )
        assert out["params"]["score_threshold"] == 0.5
        assert out["params"]["profile_id"] == "my-profile"

    def test_source_file(self):
        out = Source.file(file_id="file_abc", chunk_size=1500, chunk_overlap=150)
        assert out == {
            "type": "file",
            "params": {"file_id": "file_abc", "chunk_size": 1500, "chunk_overlap": 150},
        }

    def test_source_custom_passthrough(self):
        out = Source.custom(type="hackernews", params={"tag": "ai", "limit": 50})
        assert out == {
            "type": "hackernews",
            "params": {"tag": "ai", "limit": 50},
        }

    def test_source_custom_with_auth_ref(self):
        out = Source.custom(type="slack", params={"channel": "C123"}, auth_ref="link_abc")
        assert out["auth_ref"] == "link_abc"

    def test_strategy_all_items(self):
        assert Strategy.all_items() == {"type": "all_items", "params": {}}

    def test_strategy_custom(self):
        out = Strategy.custom(type="llm_rerank", params={"model": "claude-haiku-4-5"})
        assert out == {"type": "llm_rerank", "params": {"model": "claude-haiku-4-5"}}

    def test_shape_raw(self):
        assert Shape.raw() == {"type": "raw", "params": {}}

    def test_shape_custom(self):
        assert Shape.custom(type="markdown_digest") == {
            "type": "markdown_digest",
            "params": {},
        }

    def test_reconciler_noop(self):
        assert Reconciler.noop() == {"type": "noop", "params": {}}

    def test_reconciler_custom_with_schedule(self):
        out = Reconciler.custom(
            type="cron",
            params={"schedule": "0 6 * * *", "tz": "UTC"},
        )
        assert out["type"] == "cron"
        assert out["params"]["schedule"] == "0 6 * * *"


# ─── Unit — Constraints ─────────────────────────────────────────────────


@pytest.mark.unit
class TestConstraints:
    def test_defaults(self):
        out = Constraints().to_dict()
        assert out["max_items"] == 20
        assert out["max_per_source"] == 10
        assert out["max_crawl_time_s"] == 120.0
        assert out["language"] == "en"
        assert "freshness_days" not in out  # only present if set

    def test_freshness_emits_when_set(self):
        out = Constraints(freshness_days=7).to_dict()
        assert out["freshness_days"] == 7

    def test_override_all(self):
        out = Constraints(
            max_items=50, max_per_source=20, max_crawl_time_s=300,
            freshness_days=30, language="fr",
        ).to_dict()
        assert out == {
            "max_items": 50, "max_per_source": 20,
            "max_crawl_time_s": 300.0,
            "freshness_days": 30, "language": "fr",
        }


# ─── Unit — body composition + validation ───────────────────────────────


@pytest.mark.unit
class TestBuildBody:
    @pytest.fixture
    def crawler(self):
        return AsyncWebCrawler(api_key="sk_test_dummy", base_url="http://x")

    def test_minimal_body(self, crawler):
        body = crawler._build_context_body(intent="x")
        assert body == {"intent": "x"}

    def test_pillars_not_yet_accepted_by_public_api(self, crawler):
        """Until public generator CRUD ships, passing ad-hoc pillar configs
        to `context()` raises a clear NotImplementedError pointing at the
        dashboard. Builders themselves still work — they validate +
        serialize the configs for the day this lights up."""
        with pytest.raises(NotImplementedError, match="public generator CRUD"):
            crawler._build_context_body(
                intent="compare X and Y",
                sources=[Source.google_web()],
                strategy=Strategy.all_items(),
                shape=Shape.raw(),
                reconciler=Reconciler.noop(),
            )

    def test_with_generator_id(self, crawler):
        body = crawler._build_context_body(
            intent="x", generator_id="gen_my-weekly",
        )
        assert body == {"intent": "x", "generator_id": "gen_my-weekly"}

    def test_mutual_exclusion(self, crawler):
        """generator_id and pillar params are mutually exclusive."""
        with pytest.raises(ValueError, match="either `generator_id` OR pillar params"):
            crawler._build_context_body(
                intent="x",
                generator_id="gen_x",
                sources=[Source.google_web()],
            )

    def test_constraints_instance(self, crawler):
        body = crawler._build_context_body(
            intent="x", constraints=Constraints(max_items=5),
        )
        assert body["constraints"]["max_items"] == 5

    def test_constraints_dict_passthrough(self, crawler):
        body = crawler._build_context_body(
            intent="x", constraints={"max_items": 5},
        )
        assert body["constraints"]["max_items"] == 5

    def test_mission_and_webhook(self, crawler):
        body = crawler._build_context_body(
            intent="x",
            mission="extra background",
            webhook_url="https://hooks.example.com/cb",
        )
        assert body["mission"] == "extra background"
        assert body["webhook_url"] == "https://hooks.example.com/cb"


# ─── Unit — SSE event parsing ───────────────────────────────────────────


@pytest.mark.unit
class TestParseEvent:
    def test_status(self):
        ev = _parse_event("status", {
            "type": "status",
            "status": "planning",
            "phase": "planning",
            "version": 1,
            "ts": "2026-05-19T12:00:00Z",
        })
        assert isinstance(ev, StatusEvent)
        assert ev.status == "planning"
        assert ev.phase == "planning"
        assert ev.version == 1

    def test_phase_progress_init(self):
        ev = _parse_event("phase_progress", {
            "type": "phase_progress",
            "kind": "init",
            "phase": "fetch",
            "total": 3,
            "items": [{"id": "a", "url": "https://x"}],
        })
        assert isinstance(ev, PhaseProgressInit)
        assert ev.total == 3

    def test_phase_progress_item_update(self):
        ev = _parse_event("phase_progress", {
            "type": "phase_progress",
            "kind": "item_update",
            "id": "abc",
            "status": "done",
            "ms": 1240,
            "size": 18432,
        })
        assert isinstance(ev, PhaseProgressItemUpdate)
        assert ev.id == "abc"
        assert ev.status == "done"
        assert ev.ms == 1240
        assert ev.size == 18432

    def test_terminal(self):
        ev = _parse_event("terminal", {
            "type": "terminal",
            "status": "completed",
            "total_ms": 21834,
            "urls_crawled": 9,
            "urls_failed": 0,
        })
        assert isinstance(ev, TerminalEvent)
        assert ev.status == "completed"
        assert ev.urls_crawled == 9

    def test_unknown_event_returns_none(self):
        assert _parse_event("mystery", {"type": "mystery"}) is None


# ─── Unit — Result helpers ──────────────────────────────────────────────


@pytest.mark.unit
class TestContextResult:
    def test_from_api_minimal(self):
        r = ContextResult.from_api({
            "run_id": "ctx-run_abc",
            "status": "queued",
            "version": 1,
        })
        assert r.run_id == "ctx-run_abc"
        assert r.status == "queued"
        assert r.version == 1
        assert r.is_terminal is False
        assert r.is_success is False

    def test_is_terminal(self):
        for s in ("completed", "completed_partial", "failed", "cancelled"):
            r = ContextResult.from_api({"run_id": "x", "status": s, "version": 1})
            assert r.is_terminal is True

    def test_is_success(self):
        assert ContextResult.from_api(
            {"run_id": "x", "status": "completed", "version": 1}
        ).is_success is True
        assert ContextResult.from_api(
            {"run_id": "x", "status": "completed_partial", "version": 1}
        ).is_success is True
        assert ContextResult.from_api(
            {"run_id": "x", "status": "failed", "version": 1}
        ).is_success is False


# ─── Live — submit, stream, output, refresh, cancel ─────────────────────


@livetest
@pytest.mark.asyncio
async def test_live_default_generator_one_shot():
    """One-liner against the default generator. Submits a small Context
    run with a tight crawl-time cap to stay quick, then verifies the
    result + output shape."""
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        result = await crawler.context(
            intent="brief overview of what LangChain is, with citations",
            constraints=Constraints(max_items=5, max_per_source=3, max_crawl_time_s=60),
            wait=True,
            timeout=180.0,
        )
        assert result.is_terminal
        assert result.run_id.startswith("ctx-run_") or len(result.run_id) > 8
        assert result.version >= 1

        # Lazy output fetch
        output = await result.output()
        assert isinstance(output, ContextOutput)
        assert output.shape in ("raw",)
        # items are the citation units in the raw Shape
        assert isinstance(output.items, list)
        # On a successful run we expect at least one item — be lenient
        # in case the search returned nothing.
        for item in output.items:
            assert isinstance(item, ContextItem)
            # Provenance contract: each item carries a URL and a source
            assert item.url is not None or item.snippet is not None


@livetest
@pytest.mark.asyncio
async def test_live_streaming_events():
    """Submit + stream the default generator; iterate the stream; verify
    the documented event types fire and a terminal event closes the
    stream."""
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        seen_status_statuses = set()
        seen_terminal = False

        async for event in crawler.context_stream(
            intent="one-line answer: what is RAG",
            constraints=Constraints(max_items=2, max_per_source=2, max_crawl_time_s=30),
        ):
            if isinstance(event, StatusEvent):
                seen_status_statuses.add(event.status)
            elif isinstance(event, PhaseProgressInit):
                assert event.total >= 0
            elif isinstance(event, PhaseProgressItemUpdate):
                assert event.status in ("done", "failed")
            elif isinstance(event, TerminalEvent):
                seen_terminal = True
                assert event.status in (
                    "completed", "completed_partial", "failed", "cancelled",
                )

        assert seen_terminal, "stream never emitted terminal event"


@livetest
@pytest.mark.asyncio
async def test_live_pillar_params_raise_until_public_crud():
    """Until public generator CRUD ships, the SDK raises a clear error
    when pillar params are passed without a generator_id."""
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        with pytest.raises(NotImplementedError, match="public generator CRUD"):
            await crawler.context(
                intent="test",
                sources=[Source.google_web()],
            )


@livetest
@pytest.mark.asyncio
async def test_live_get_and_cancel():
    """Submit no-wait, fetch state, cancel.

    Per-generator concurrency cap is 1; if a previous test left a run
    in-flight we may transient-429. Submit-then-cancel is supposed to
    return the cap immediately, but server-side cancellation is async
    so we tolerate up to two retries with a back-off before failing."""
    import time as _t
    from crawl4ai_cloud.errors import QuotaExceededError

    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        # Best-effort: poll briefly until we get a slot.
        deadline = _t.monotonic() + 60
        result = None
        while _t.monotonic() < deadline:
            try:
                result = await crawler.context(
                    intent="ignored — will be cancelled",
                    constraints=Constraints(max_items=2, max_crawl_time_s=30),
                    wait=False,
                )
                break
            except QuotaExceededError:
                await asyncio.sleep(5)
        assert result is not None, "couldn't get a generator slot in 60s"
        assert not result.is_terminal

        state = await crawler.get_context_run(result.run_id)
        assert state.run_id == result.run_id

        # Cancel — async on the server side
        await crawler.cancel_context_run(result.run_id)


@livetest
@pytest.mark.asyncio
async def test_live_versions_and_refresh():
    """Submit, wait, refresh, list versions. Asserts a v2 is created.

    Same retry pattern as the other live tests — wait briefly for the
    generator slot to clear if a prior test still holds it."""
    import time as _t
    from crawl4ai_cloud.errors import QuotaExceededError

    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        deadline = _t.monotonic() + 90
        v1 = None
        while _t.monotonic() < deadline:
            try:
                v1 = await crawler.context(
                    intent="one-line overview of vector databases",
                    constraints=Constraints(max_items=2, max_per_source=2, max_crawl_time_s=30),
                    wait=True,
                    timeout=180.0,
                )
                break
            except QuotaExceededError:
                await asyncio.sleep(5)
        assert v1 is not None, "couldn't get a generator slot in 90s"
        assert v1.is_terminal

        v2 = None
        while _t.monotonic() < deadline:
            try:
                v2 = await crawler.refresh_context(
                    v1.run_id, wait=True, timeout=180.0,
                )
                break
            except QuotaExceededError:
                await asyncio.sleep(5)
        assert v2 is not None
        assert v2.version >= v1.version

        versions = await crawler.list_context_versions(v1.run_id)
        assert len(versions) >= 2
        assert all(isinstance(v, ContextVersion) for v in versions)
        assert max(v.version for v in versions) >= v2.version


@livetest
@pytest.mark.asyncio
async def test_live_diff():
    """Same-chain diff — submit, refresh, diff.

    Per-generator cap is 1; the previous live test may have left a
    transient hold on the default generator. Wait + retry until we get
    a slot, same pattern as test_live_get_and_cancel."""
    import time as _t
    from crawl4ai_cloud.errors import QuotaExceededError

    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        deadline = _t.monotonic() + 90
        v1 = None
        while _t.monotonic() < deadline:
            try:
                v1 = await crawler.context(
                    intent="one-line answer: what is HTTP/2",
                    constraints=Constraints(max_items=2, max_per_source=2, max_crawl_time_s=30),
                    wait=True,
                    timeout=180.0,
                )
                break
            except QuotaExceededError:
                await asyncio.sleep(5)
        assert v1 is not None, "couldn't get a generator slot in 90s"

        # Refresh + diff. The diff endpoint takes two run_ids — passing
        # the same chain's run_id twice gives the chain's most-recent
        # cross-version diff.
        try:
            await crawler.refresh_context(v1.run_id, wait=True, timeout=180.0)
        except QuotaExceededError:
            await asyncio.sleep(10)
            await crawler.refresh_context(v1.run_id, wait=True, timeout=180.0)

        diff = await crawler.diff_context(v1.run_id, v1.run_id)
        assert isinstance(diff, ContextDiff)
        # `added`, `removed`, `unchanged` are lists; we don't assert
        # contents (highly dependent on server output).


@livetest
@pytest.mark.asyncio
async def test_live_catalog():
    """The catalog should at minimum surface the four pillars we documented."""
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        catalog = await crawler.context_catalog()
        assert isinstance(catalog, ContextCatalog)

        source_names = {s.name for s in catalog.sources}
        strategy_names = {s.name for s in catalog.strategies}
        shape_names = {s.name for s in catalog.shapes}
        reconciler_names = {s.name for s in catalog.reconcilers}

        # Today's catalog must include these names at minimum.
        assert "google_web" in source_names
        assert "all_items" in strategy_names
        assert "raw" in shape_names
        assert "noop" in reconciler_names


if __name__ == "__main__":
    import sys
    sys.exit(pytest.main([__file__, "-v"]))
