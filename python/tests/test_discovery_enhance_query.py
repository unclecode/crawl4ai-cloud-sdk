"""
E2E tests for the `enhance_query` opt-in on /v1/discovery/search.

Real HTTP against stage.crawl4ai.com. Covers:
  - Single-backend rewrite happy path
  - Multi-backend fan-out — per-backend rewrites all surfaced
  - Default off → response carries no rewrite fields
  - Async + synth + enhance_query — rewrites land in the polled result
  - Cache opt-in interplay — `use_cache=true` + `enhance_query=true`
    returns the same rewrites on the second hit

Run:
    cd crawl4ai-cloud-sdk/python
    python -m pytest tests/test_discovery_enhance_query.py -v --timeout=180
"""
import os
import pytest
import pytest_asyncio

from crawl4ai_cloud import AsyncWebCrawler


API_KEY = os.environ.get(
    "CRAWL4AI_API_KEY",
    "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ",
)
BASE_URL = os.environ.get("CRAWL4AI_BASE_URL", "https://stage.crawl4ai.com")


@pytest_asyncio.fixture
async def crawler():
    async with AsyncWebCrawler(api_key=API_KEY, base_url=BASE_URL) as c:
        yield c


@pytest.mark.asyncio
async def test_enhance_query_single_backend_returns_rewrite(crawler):
    resp = await crawler.discovery(
        "search",
        query="what are the best nurseries in Toronto for my 2 year old",
        country="ca",
        enhance_query=True,
    )
    assert resp.original_query == \
        "what are the best nurseries in Toronto for my 2 year old"
    assert isinstance(resp.rewritten_queries, dict)
    assert "google" in resp.rewritten_queries
    rewritten = resp.rewritten_queries["google"]
    # Library prompt principles: should strip filler, quote the age,
    # OR-expand "nurseries". We don't pin the exact string (LLM output)
    # but check the operator patterns the prompt teaches.
    assert "2 year old" in rewritten or "toddler" in rewritten.lower()
    assert "Toronto" in rewritten
    # The original conversational phrasing should be gone.
    assert "what are the best" not in rewritten
    # Search itself succeeded.
    assert len(resp.hits) > 0


@pytest.mark.asyncio
async def test_enhance_query_multi_backend_per_backend_rewrites(crawler):
    resp = await crawler.discovery(
        "search",
        query="latest claude news this week",
        country="us",
        enhance_query=True,
        backends=["google", "bing"],
    )
    assert resp.original_query == "latest claude news this week"
    rewrites = resp.rewritten_queries
    assert isinstance(rewrites, dict)
    assert set(rewrites.keys()) == {"google", "bing"}
    # Google has `after:` operator; Bing does not — should use a year token.
    assert "after:" in rewrites["google"]
    assert "2026" in rewrites["bing"]


@pytest.mark.asyncio
async def test_enhance_query_default_off_no_rewrite_fields(crawler):
    resp = await crawler.discovery(
        "search", query="openai latest news", country="us",
    )
    # Default request — no enhance_query, no rewrite fields populated.
    assert resp.original_query is None
    assert resp.rewritten_queries is None
    assert len(resp.hits) > 0


@pytest.mark.asyncio
async def test_enhance_query_with_synth_async_lifecycle(crawler):
    """Async + synth + enhance_query: the rewrites must land in the
    final polled result alongside the synthesized answer."""
    resp = await crawler.discovery(
        "search",
        query="what is warriors next game?",
        country="us",
        backends=["google", "bing"],
        enhance_query=True,
        synthesize=True,
        synth_mode="auto",
        # default wait=True — SDK polls through serp_ready → completed
    )
    # SearchResponse on success
    assert resp.original_query == "what is warriors next game?"
    assert isinstance(resp.rewritten_queries, dict)
    assert "google" in resp.rewritten_queries
    assert "bing" in resp.rewritten_queries
    # Synth still produced an answer.
    assert resp.synthesized_answer is not None
    assert resp.synthesized_answer.text


@pytest.mark.asyncio
async def test_enhance_query_with_use_cache_returns_same_rewrite(crawler):
    """`use_cache=True` + `enhance_query=True`: two back-to-back requests
    with the same query must return the same rewritten string. The
    cloud orchestrator's rewrite cache (deterministic at temp 0) lets
    us assert byte equality."""
    q = "tell me about graph databases for AI agents"
    r1 = await crawler.discovery(
        "search", query=q, country="us",
        enhance_query=True, use_cache=True,
    )
    r2 = await crawler.discovery(
        "search", query=q, country="us",
        enhance_query=True, use_cache=True,
    )
    assert r1.rewritten_queries == r2.rewritten_queries
    assert "google" in r1.rewritten_queries
