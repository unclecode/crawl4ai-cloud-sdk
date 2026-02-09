"""MCP server for Crawl4AI Claude Code plugin."""
from __future__ import annotations

import json
from typing import List, Optional

from mcp.server.fastmcp import FastMCP

from . import core

mcp = FastMCP("crawl4ai")


def _json(data: dict) -> str:
    return json.dumps(data, indent=2, default=str)


@mcp.tool()
async def crawl(
    url: str,
    deep_crawl: bool = False,
    strategy: str = "bfs",
    max_depth: int = 2,
    max_pages: int = 50,
    include_patterns: Optional[List[str]] = None,
    exclude_patterns: Optional[List[str]] = None,
    css_selector: Optional[str] = None,
    word_count_threshold: int = 200,
    bypass_cache: bool = False,
) -> str:
    """Crawl a web page and return its content as markdown.

    For a single page, returns markdown text, metadata, and links.
    Set deep_crawl=True for multi-page crawling with configurable strategy/depth.
    """
    result = await core.crawl(
        url, deep_crawl=deep_crawl, strategy=strategy,
        max_depth=max_depth, max_pages=max_pages,
        include_patterns=include_patterns, exclude_patterns=exclude_patterns,
        css_selector=css_selector, word_count_threshold=word_count_threshold,
        bypass_cache=bypass_cache,
    )
    return _json(result)


@mcp.tool()
async def extract(
    url: str,
    schema: dict,
    schema_type: str = "css",
) -> str:
    """Extract structured data from a web page using a CSS/XPath schema.

    The schema should have: name, baseSelector, and fields[].
    Each field has: name, selector, type (text/attribute/html/list).
    """
    result = await core.extract(url, schema=schema, schema_type=schema_type)
    return _json(result)


@mcp.tool()
async def map(
    url: str,
    source: str = "sitemap",
    pattern: str = "*",
    max_urls: int = 100,
    query: Optional[str] = None,
    score_threshold: Optional[float] = None,
) -> str:
    """Discover URLs on a domain via sitemap, Common Crawl, or both.

    Sources: "sitemap", "cc" (Common Crawl), "sitemap+cc".
    Use query for BM25 relevance filtering.
    """
    result = await core.map_urls(
        url, source=source, pattern=pattern,
        max_urls=max_urls, query=query, score_threshold=score_threshold,
    )
    return _json(result)


@mcp.tool()
async def screenshot(
    url: str,
    wait_for: Optional[str] = None,
    css_selector: Optional[str] = None,
    full_page: bool = True,
) -> str:
    """Take a screenshot of a web page. Returns base64-encoded PNG.

    Use wait_for to wait for a CSS selector or JS expression before capture.
    Use css_selector to screenshot a specific element.
    """
    result = await core.screenshot(
        url, wait_for=wait_for, css_selector=css_selector, full_page=full_page,
    )
    return _json(result)


@mcp.tool()
async def schema(
    query: str,
    url: Optional[str] = None,
    html: Optional[str] = None,
    schema_type: str = "css",
) -> str:
    """Generate a CSS/XPath extraction schema from a page using AI.

    Provide either a url or raw html, plus a natural language query
    describing what data to extract. Returns a reusable schema for the extract tool.
    """
    result = await core.schema(query=query, url=url, html=html, schema_type=schema_type)
    return _json(result)


@mcp.tool()
async def profile_list() -> str:
    """List available browser profiles.

    In cloud mode, profiles are managed via the dashboard/CLI.
    In local mode, lists profiles from ~/.crawl4ai/profiles/.
    """
    result = await core.profile_list()
    return _json(result)


@mcp.tool()
async def profile_create(
    profile_name: Optional[str] = None,
) -> str:
    """Create a new browser profile for authenticated crawling.

    Opens a browser window to log in to sites. The session is saved
    and can be reused in future crawls for auth-gated content.
    """
    result = await core.profile_create(profile_name=profile_name)
    return _json(result)


def main():
    mcp.run(transport="stdio")


if __name__ == "__main__":
    main()
