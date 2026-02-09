"""Core functions for Crawl4AI Claude plugin — backend-agnostic."""
from __future__ import annotations

import asyncio
from typing import Any, Dict, List, Optional

from .config import load_config, PluginConfig
from .backends import BackendError, CrawlBackend, get_backend

# Singleton state
_backend: Optional[CrawlBackend] = None
_current_mode: Optional[str] = None


async def _get_backend() -> CrawlBackend:
    """Get or create the backend singleton, auto-resetting on mode change."""
    global _backend, _current_mode
    config = load_config()

    if _backend is not None and _current_mode == config.mode:
        return _backend

    # Mode changed or first call — (re)initialize
    if _backend is not None:
        try:
            await _backend.shutdown()
        except Exception:
            pass

    _backend = get_backend(config)
    _current_mode = config.mode
    await _backend.startup()
    return _backend


async def reset_backend() -> Dict[str, Any]:
    """Shutdown current backend (next call will re-create from config)."""
    global _backend, _current_mode
    if _backend is not None:
        try:
            await _backend.shutdown()
        except Exception:
            pass
        _backend = None
        _current_mode = None
    return {"success": True, "data": {"message": "Backend reset. Next call will reinitialize."}}


# ======================================================================
# 7 tool functions — each returns {"success": bool, "data"|"error": ...}
# ======================================================================

async def crawl(
    url: str,
    *,
    deep_crawl: bool = False,
    strategy: str = "bfs",
    max_depth: int = 2,
    max_pages: int = 50,
    include_patterns: Optional[List[str]] = None,
    exclude_patterns: Optional[List[str]] = None,
    css_selector: Optional[str] = None,
    word_count_threshold: int = 200,
    bypass_cache: bool = False,
) -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.crawl(
            url, deep_crawl=deep_crawl, strategy=strategy,
            max_depth=max_depth, max_pages=max_pages,
            include_patterns=include_patterns, exclude_patterns=exclude_patterns,
            css_selector=css_selector, word_count_threshold=word_count_threshold,
            bypass_cache=bypass_cache,
        )
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


async def extract(
    url: str,
    *,
    schema: dict,
    schema_type: str = "css",
) -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.extract(url, schema=schema, schema_type=schema_type)
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


async def map_urls(
    url: str,
    *,
    source: str = "sitemap",
    pattern: str = "*",
    max_urls: int = 100,
    query: Optional[str] = None,
    score_threshold: Optional[float] = None,
) -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.map_urls(
            url, source=source, pattern=pattern,
            max_urls=max_urls, query=query, score_threshold=score_threshold,
        )
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


async def screenshot(
    url: str,
    *,
    wait_for: Optional[str] = None,
    css_selector: Optional[str] = None,
    full_page: bool = True,
) -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.screenshot(
            url, wait_for=wait_for, css_selector=css_selector, full_page=full_page,
        )
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


async def schema(
    *,
    query: str,
    url: Optional[str] = None,
    html: Optional[str] = None,
    schema_type: str = "css",
) -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.generate_schema(
            url=url, html=html, query=query, schema_type=schema_type,
        )
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


async def profile_list() -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.list_profiles()
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


async def profile_create(
    *,
    profile_name: Optional[str] = None,
) -> Dict[str, Any]:
    try:
        backend = await _get_backend()
        data = await backend.create_profile(profile_name=profile_name)
        return {"success": True, "data": data}
    except BackendError as e:
        return {"success": False, "error": e.message}
    except Exception as e:
        return {"success": False, "error": str(e)}


# ======================================================================
# Standalone test
# ======================================================================

async def _test():
    """Quick smoke test — crawl example.com."""
    config = load_config()
    print(f"Mode: {config.mode}")
    print(f"API key: {'set' if config.api_key else 'NOT SET'}")
    print()

    print("--- crawl ---")
    result = await crawl("https://example.com")
    if result["success"]:
        md = result["data"].get("markdown", "")
        print(f"OK: {len(md)} chars of markdown")
    else:
        print(f"FAIL: {result['error']}")

    print()
    print("--- map_urls ---")
    result = await map_urls("https://example.com")
    if result["success"]:
        print(f"OK: {result['data'].get('total_discovered', 0)} URLs found")
    else:
        print(f"FAIL: {result['error']}")

    print()
    print("--- profile_list ---")
    result = await profile_list()
    if result["success"]:
        print(f"OK: {result['data']}")
    else:
        print(f"FAIL: {result['error']}")

    # Cleanup
    await reset_backend()
    print("\nDone.")


if __name__ == "__main__":
    asyncio.run(_test())
