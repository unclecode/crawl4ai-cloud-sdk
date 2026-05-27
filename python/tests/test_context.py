#!/usr/bin/env python3
"""
Context v2 SDK tests — unit + live.

Unit tests cover:
  - All pillar builders produce the documented wire dict shape
  - Synthesizer.markdown / llm validation
  - Strategy.llm_rerank
  - Inline pipeline body composition
  - ContextOutput shape-specific sugar (.markdown / .files / .data)
  - SSE event parsing
  - ContextResult helpers

Live tests (gated on CRAWL4AI_API_KEY) submit real runs against stage:
  - default generator one-shot
  - inline pipeline (google_web + raw)
  - markdown synthesizer (single)
  - llm synthesizer (by example)
  - catalog
  - streaming
  - refresh / versions / diff / cancel

Run:
    pytest tests/test_context.py -k unit -v                                # unit only
    CRAWL4AI_API_KEY=sk_live_... pytest tests/test_context.py -v           # all
"""
import asyncio
import os

import pytest

from crawl4ai_cloud import (
    AsyncWebCrawler,
    Source,
    Strategy,
    Synthesizer,
    Shape,           # deprecated alias
    Reconciler,
    Constraints,
    ContextResult,
    ContextOutput,
    ContextItem,
    MarkdownFile,
    StatusEvent,
    TerminalEvent,
    PhaseProgressInit,
    PhaseProgressItemUpdate,
    ContextDiff,
    ContextVersion,
    ContextCatalog,
)
from crawl4ai_cloud.context import _parse_event


BASE_URL = os.getenv("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")
API_KEY = os.getenv("CRAWL4AI_API_KEY")

livetest = pytest.mark.skipif(
    not API_KEY,
    reason="CRAWL4AI_API_KEY not set — skipping live tests",
)


# ─── Unit — Source builders ─────────────────────────────────────────────


@pytest.mark.unit
class TestSourceBuilders:
    def test_google_web_defaults(self):
        assert Source.google_web() == {
            "type": "google_web",
            "params": {"top_k_per_backend": 10},
        }

    def test_google_web_full(self):
        out = Source.google_web(backends=["google", "bing"], top_k_per_backend=8, region="us")
        assert out == {
            "type": "google_web",
            "params": {"backends": ["google", "bing"], "top_k_per_backend": 8, "region": "us"},
        }

    def test_google_drive_search_default(self):
        out = Source.google_drive()
        assert out["type"] == "google_drive"
        assert out["params"]["mode"] == "search"
        assert out["params"]["folder_id"] == ""

    def test_google_drive_folder(self):
        out = Source.google_drive(mode="folder", folder_id="abc123")
        assert out["params"] == {"mode": "folder", "folder_id": "abc123"}

    def test_google_drive_folder_requires_id(self):
        with pytest.raises(ValueError, match="folder_id is required"):
            Source.google_drive(mode="folder")

    def test_google_drive_bad_mode(self):
        with pytest.raises(ValueError, match="mode must be"):
            Source.google_drive(mode="bogus")

    def test_google_drive_with_auth_ref(self):
        out = Source.google_drive(auth_ref="link_abc")
        assert out["auth_ref"] == "link_abc"

    def test_gmail_search_default(self):
        out = Source.gmail()
        assert out["type"] == "gmail"
        assert out["params"]["mode"] == "search"
        assert out["params"]["include_spam_trash"] is False

    def test_gmail_label(self):
        out = Source.gmail(mode="label", label_id="Label_42",
                           after="2026/01/01", before="2026/05/01",
                           include_spam_trash=True)
        assert out["params"] == {
            "mode": "label", "label_id": "Label_42",
            "after": "2026/01/01", "before": "2026/05/01",
            "include_spam_trash": True,
        }

    def test_gmail_label_requires_id(self):
        with pytest.raises(ValueError, match="label_id is required"):
            Source.gmail(mode="label")

    def test_gmail_bad_mode(self):
        with pytest.raises(ValueError, match="mode must be"):
            Source.gmail(mode="bogus")

    def test_crawl(self):
        out = Source.crawl(domain="https://example.com", max_urls=30, max_depth=2,
                           score_threshold=0.5, profile_id="my-profile")
        assert out["params"]["domain"] == "https://example.com"
        assert out["params"]["max_urls"] == 30
        assert out["params"]["score_threshold"] == 0.5
        assert out["params"]["profile_id"] == "my-profile"

    def test_file(self):
        out = Source.file(file_id="file_abc")
        assert out == {
            "type": "file",
            "params": {"file_id": "file_abc", "chunk_size": 2000, "chunk_overlap": 200},
        }

    def test_custom(self):
        out = Source.custom(type="hackernews", params={"tag": "ai"}, auth_ref="link_x")
        assert out == {"type": "hackernews", "params": {"tag": "ai"}, "auth_ref": "link_x"}


# ─── Unit — Strategy builders ───────────────────────────────────────────


@pytest.mark.unit
class TestStrategyBuilders:
    def test_all_items(self):
        assert Strategy.all_items() == {"type": "all_items", "params": {}}

    def test_llm_rerank_defaults(self):
        out = Strategy.llm_rerank()
        assert out["type"] == "llm_rerank"
        p = out["params"]
        assert p["top_n"] == 0
        assert p["instruction"] == ""
        assert p["score_threshold"] == 0.0
        assert p["batch_size"] == 20
        assert p["max_concurrency"] == 4
        assert p["content_aware"] is False
        assert p["content_chars"] == 4000

    def test_llm_rerank_full(self):
        out = Strategy.llm_rerank(
            top_n=5, instruction="Prefer official docs.",
            model="anthropic/claude-sonnet-4-6",
            score_threshold=0.3,
            content_aware=True, content_chars=6000,
        )
        p = out["params"]
        assert p["top_n"] == 5
        assert p["instruction"] == "Prefer official docs."
        assert p["model"] == "anthropic/claude-sonnet-4-6"
        assert p["score_threshold"] == 0.3
        assert p["content_aware"] is True
        assert p["content_chars"] == 6000

    def test_custom(self):
        out = Strategy.custom(type="custom_strat", params={"x": 1})
        assert out == {"type": "custom_strat", "params": {"x": 1}}


# ─── Unit — Synthesizer builders ────────────────────────────────────────


@pytest.mark.unit
class TestSynthesizerBuilders:
    def test_raw(self):
        assert Synthesizer.raw() == {"type": "raw", "params": {}}

    def test_shape_alias_is_synthesizer(self):
        # Deprecated alias kept for one release.
        assert Shape is Synthesizer
        assert Shape.raw() == {"type": "raw", "params": {}}

    def test_markdown_single_defaults(self):
        out = Synthesizer.markdown()
        assert out["type"] == "markdown"
        p = out["params"]
        assert p["mode"] == "single"
        assert p["instruction"] == ""
        assert p["batch_size"] == 5
        assert p["max_concurrency"] == 4
        assert p["include_metadata"] is True
        assert p["max_chars_per_item"] == 20000

    def test_markdown_multi(self):
        out = Synthesizer.markdown(mode="multi", instruction="Summarise to 200 words.")
        assert out["params"]["mode"] == "multi"
        assert out["params"]["instruction"] == "Summarise to 200 words."

    def test_markdown_bad_mode(self):
        with pytest.raises(ValueError, match="mode must be"):
            Synthesizer.markdown(mode="bogus")

    def test_llm_by_example(self):
        out = Synthesizer.llm(
            instruction="extract a knowledge graph",
            example={"nodes": [{"id": 1, "label": "x"}]},
        )
        assert out["type"] == "llm"
        p = out["params"]
        assert p["instruction"] == "extract a knowledge graph"
        # dict example is JSON-serialised
        import json
        assert json.loads(p["output_example"]) == {"nodes": [{"id": 1, "label": "x"}]}
        assert p["output_schema"] == ""
        assert p["output_description"] == ""

    def test_llm_by_schema_dict(self):
        out = Synthesizer.llm(
            instruction="tabulate",
            schema={"type": "object", "properties": {"a": {"type": "string"}}},
        )
        import json
        assert json.loads(out["params"]["output_schema"])["type"] == "object"

    def test_llm_by_schema_string(self):
        # Strings pass through unchanged.
        out = Synthesizer.llm(
            instruction="tabulate",
            schema='{"type":"object"}',
        )
        assert out["params"]["output_schema"] == '{"type":"object"}'

    def test_llm_by_description(self):
        out = Synthesizer.llm(
            instruction="summarise",
            description="An object with title and body.",
        )
        assert out["params"]["output_description"] == "An object with title and body."

    def test_llm_requires_instruction(self):
        with pytest.raises(ValueError, match="instruction is required"):
            Synthesizer.llm(instruction="", example={"a": 1})

    def test_llm_requires_exactly_one_input(self):
        with pytest.raises(ValueError, match="exactly one"):
            Synthesizer.llm(instruction="x")
        with pytest.raises(ValueError, match="exactly one"):
            Synthesizer.llm(instruction="x", schema={}, example={})
        with pytest.raises(ValueError, match="exactly one"):
            Synthesizer.llm(instruction="x", schema={}, description="x")

    def test_custom(self):
        out = Synthesizer.custom(type="future_synth", params={"k": "v"})
        assert out == {"type": "future_synth", "params": {"k": "v"}}


# ─── Unit — Reconciler builders ─────────────────────────────────────────


@pytest.mark.unit
class TestReconcilerBuilders:
    def test_noop(self):
        assert Reconciler.noop() == {"type": "noop", "params": {}}

    def test_custom(self):
        out = Reconciler.custom(type="cron", params={"schedule": "0 6 * * *"})
        assert out == {"type": "cron", "params": {"schedule": "0 6 * * *"}}


# ─── Unit — Constraints ─────────────────────────────────────────────────


@pytest.mark.unit
class TestConstraints:
    def test_defaults(self):
        out = Constraints().to_dict()
        assert out["max_items"] == 20
        assert out["language"] == "en"
        assert "freshness_days" not in out

    def test_freshness_when_set(self):
        assert Constraints(freshness_days=7).to_dict()["freshness_days"] == 7


# ─── Unit — body composition ────────────────────────────────────────────


@pytest.mark.unit
class TestBuildBody:
    @pytest.fixture
    def crawler(self):
        return AsyncWebCrawler(api_key="sk_test_dummy", base_url="http://x")

    def test_minimal(self, crawler):
        body = crawler._build_context_body(intent="x")
        assert body == {"intent": "x"}

    def test_with_generator_id(self, crawler):
        body = crawler._build_context_body(intent="x", generator_id="gen_42")
        assert body == {"intent": "x", "generator_id": "gen_42"}

    def test_inline_pipeline_basic(self, crawler):
        body = crawler._build_context_body(
            intent="x",
            sources=[Source.google_web()],
        )
        assert body["pipeline"]["sources"][0]["type"] == "google_web"
        # No strategy / synthesizer / reconciler passed → server applies defaults.
        assert "strategy" not in body["pipeline"]
        assert "synthesizer" not in body["pipeline"]

    def test_inline_pipeline_full(self, crawler):
        body = crawler._build_context_body(
            intent="x",
            sources=[Source.google_web(), Source.google_drive(folder_id="abc", mode="folder")],
            strategy=Strategy.llm_rerank(top_n=5, instruction="prefer docs"),
            synthesizer=Synthesizer.markdown(mode="single"),
            reconciler=Reconciler.noop(),
        )
        p = body["pipeline"]
        assert len(p["sources"]) == 2
        assert p["strategy"] == "llm_rerank"
        assert p["strategy_params"]["top_n"] == 5
        assert p["synthesizer"] == "markdown"
        assert p["synthesizer_params"]["mode"] == "single"
        assert p["reconciler"] == "noop"
        assert p["reconciler_params"] == {}

    def test_shape_alias_kwarg(self, crawler):
        # Old `shape=` kwarg still works (deprecated).
        body = crawler._build_context_body(
            intent="x",
            sources=[Source.google_web()],
            shape=Synthesizer.markdown(mode="single"),
        )
        assert body["pipeline"]["synthesizer"] == "markdown"

    def test_synthesizer_wins_over_shape_alias(self, crawler):
        # When both are passed, the canonical `synthesizer=` kwarg wins.
        body = crawler._build_context_body(
            intent="x",
            sources=[Source.google_web()],
            synthesizer=Synthesizer.markdown(mode="multi"),
            shape=Synthesizer.raw(),
        )
        assert body["pipeline"]["synthesizer"] == "markdown"
        assert body["pipeline"]["synthesizer_params"]["mode"] == "multi"

    def test_mutual_exclusion(self, crawler):
        with pytest.raises(ValueError, match="either `generator_id`"):
            crawler._build_context_body(
                intent="x",
                generator_id="gen_x",
                sources=[Source.google_web()],
            )

    def test_inline_pipeline_requires_sources(self, crawler):
        with pytest.raises(ValueError, match="at least one Source"):
            crawler._build_context_body(
                intent="x",
                strategy=Strategy.all_items(),
                synthesizer=Synthesizer.raw(),
            )

    def test_constraints_instance(self, crawler):
        body = crawler._build_context_body(intent="x", constraints=Constraints(max_items=5))
        assert body["constraints"]["max_items"] == 5

    def test_constraints_dict(self, crawler):
        body = crawler._build_context_body(intent="x", constraints={"max_items": 5})
        assert body["constraints"]["max_items"] == 5

    def test_mission_and_webhook(self, crawler):
        body = crawler._build_context_body(
            intent="x", mission="background",
            webhook_url="https://hooks.example.com/cb",
        )
        assert body["mission"] == "background"
        assert body["webhook_url"] == "https://hooks.example.com/cb"


# ─── Unit — ContextOutput sugar ─────────────────────────────────────────


@pytest.mark.unit
class TestOutputSugar:
    def test_raw_shape(self):
        out = ContextOutput.from_api({
            "type": "raw",
            "data": {"items": [{"url": "https://x", "title": "T", "content": "C"}]},
        })
        assert out.shape == "raw"
        assert len(out.items) == 1
        assert out.items[0].url == "https://x"
        assert out.markdown is None
        assert out.files is None
        assert out.data is None

    def test_markdown_single(self):
        out = ContextOutput.from_api({
            "type": "markdown",
            "data": {
                "mode": "single",
                "items": [{"url": "https://a", "title": "A"}],
                "markdown": "# heading\n\nbody",
            },
        })
        assert out.shape == "markdown"
        assert out.markdown == "# heading\n\nbody"
        assert out.files is None

    def test_markdown_multi(self):
        out = ContextOutput.from_api({
            "type": "markdown",
            "data": {
                "mode": "multi",
                "items": [{"url": "https://a"}, {"url": "https://b"}],
                "files": [
                    {"filename": "a.md", "markdown": "# A"},
                    {"filename": "b.md", "markdown": "# B"},
                ],
            },
        })
        assert out.markdown is None
        assert len(out.files) == 2
        assert out.files[0].filename == "a.md"
        assert out.files[0].markdown == "# A"
        assert isinstance(out.files[0], MarkdownFile)

    def test_llm(self):
        out = ContextOutput.from_api({
            "type": "llm",
            "data": {
                "items": [{"url": "https://a"}],
                "data": {"runtimes": [{"name": "tokio"}]},
                "resolved_schema": {"type": "object"},
                "notes": ["resolved schema from output_example (walked)"],
            },
        })
        assert out.shape == "llm"
        assert out.data == {"runtimes": [{"name": "tokio"}]}
        assert out.resolved_schema == {"type": "object"}
        assert out.notes == ["resolved schema from output_example (walked)"]

    def test_raw_payload_back_compat_alias(self):
        out = ContextOutput.from_api({"type": "raw", "data": {"items": []}})
        # `.raw` is a deprecated alias for `.raw_payload`.
        assert out.raw is out.raw_payload


# ─── Unit — SSE event parsing ───────────────────────────────────────────


@pytest.mark.unit
class TestParseEvent:
    def test_status(self):
        ev = _parse_event("status", {
            "type": "status", "status": "planning", "phase": "planning",
            "version": 1, "ts": "2026-05-26T12:00:00Z",
        })
        assert isinstance(ev, StatusEvent)
        assert ev.status == "planning"
        assert ev.phase == "planning"

    def test_phase_progress_init(self):
        ev = _parse_event("phase_progress", {
            "type": "phase_progress", "kind": "init",
            "phase": "fetch", "total": 3,
            "items": [{"id": "a", "url": "https://x"}],
        })
        assert isinstance(ev, PhaseProgressInit)
        assert ev.total == 3

    def test_phase_progress_item_update(self):
        ev = _parse_event("phase_progress", {
            "type": "phase_progress", "kind": "item_update",
            "id": "abc", "status": "done", "ms": 1240, "size": 18432,
        })
        assert isinstance(ev, PhaseProgressItemUpdate)
        assert ev.id == "abc" and ev.status == "done"

    def test_terminal(self):
        ev = _parse_event("terminal", {
            "type": "terminal", "status": "completed",
            "total_ms": 21834, "urls_crawled": 9, "urls_failed": 0,
        })
        assert isinstance(ev, TerminalEvent)
        assert ev.urls_crawled == 9

    def test_unknown_returns_none(self):
        assert _parse_event("mystery", {"type": "mystery"}) is None


# ─── Unit — ContextResult ───────────────────────────────────────────────


@pytest.mark.unit
class TestContextResult:
    def test_from_api(self):
        r = ContextResult.from_api({"run_id": "ctx-run_abc", "status": "queued", "version": 1})
        assert r.run_id == "ctx-run_abc"
        assert r.is_terminal is False
        assert r.is_success is False

    def test_terminal_flags(self):
        for s in ("completed", "completed_partial", "failed", "cancelled"):
            r = ContextResult.from_api({"run_id": "x", "status": s, "version": 1})
            assert r.is_terminal
        assert ContextResult.from_api({"run_id": "x", "status": "completed", "version": 1}).is_success
        assert not ContextResult.from_api({"run_id": "x", "status": "failed", "version": 1}).is_success


# ─── Live — submit, stream, output, refresh, cancel ─────────────────────


@livetest
@pytest.mark.asyncio
async def test_live_default_generator_one_shot():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        result = await crawler.context(
            intent="brief overview of what LangChain is, with citations",
            constraints=Constraints(max_items=5, max_per_source=3, max_crawl_time_s=60),
            wait=True, timeout=180.0,
        )
        assert result.is_terminal
        out = await result.output()
        assert isinstance(out, ContextOutput)
        assert out.shape in ("raw", "markdown", "llm")


@livetest
@pytest.mark.asyncio
async def test_live_inline_pipeline_raw():
    """Inline pipeline — should now succeed (previously raised NotImplementedError)."""
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        result = await crawler.context(
            intent="what is HTTP/2",
            sources=[Source.google_web(top_k_per_backend=5)],
            strategy=Strategy.all_items(),
            synthesizer=Synthesizer.raw(),
            reconciler=Reconciler.noop(),
            constraints=Constraints(max_items=3, max_crawl_time_s=45),
            wait=True, timeout=180.0,
        )
        assert result.is_terminal
        out = await result.output()
        assert out.shape == "raw"


@livetest
@pytest.mark.asyncio
async def test_live_inline_pipeline_markdown_single():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        result = await crawler.context(
            intent="one-line overview of WebAssembly",
            sources=[Source.google_web(top_k_per_backend=3)],
            synthesizer=Synthesizer.markdown(mode="single"),
            constraints=Constraints(max_items=2, max_crawl_time_s=45),
            wait=True, timeout=180.0,
        )
        assert result.is_terminal
        out = await result.output()
        assert out.shape == "markdown"
        assert out.markdown is None or isinstance(out.markdown, str)


@livetest
@pytest.mark.asyncio
async def test_live_streaming_events():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        seen_terminal = False
        async for event in crawler.context_stream(
            intent="one-line answer: what is RAG",
            constraints=Constraints(max_items=2, max_per_source=2, max_crawl_time_s=30),
        ):
            if isinstance(event, TerminalEvent):
                seen_terminal = True
        assert seen_terminal


@livetest
@pytest.mark.asyncio
async def test_live_catalog():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as crawler:
        catalog = await crawler.context_catalog()
        assert isinstance(catalog, ContextCatalog)
        source_names = {s.name for s in catalog.sources}
        synth_names = {s.name for s in catalog.synthesizers}
        assert "google_web" in source_names
        assert "raw" in synth_names
        assert "markdown" in synth_names
        assert "llm" in synth_names


if __name__ == "__main__":
    import sys
    sys.exit(pytest.main([__file__, "-v"]))
