"""Backend protocol and factory for Crawl4AI Claude plugin."""
from __future__ import annotations

from typing import Any, Dict, List, Optional, Protocol, runtime_checkable

from ..config import PluginConfig


class BackendError(Exception):
    """Error raised by backends, caught by core.py."""

    def __init__(self, message: str, details: Optional[Dict[str, Any]] = None):
        super().__init__(message)
        self.message = message
        self.details = details or {}


@runtime_checkable
class CrawlBackend(Protocol):
    async def startup(self) -> None: ...
    async def shutdown(self) -> None: ...
    async def crawl(self, url: str, *, deep_crawl: bool = False, strategy: str = "bfs",
                    max_depth: int = 2, max_pages: int = 50,
                    include_patterns: Optional[List[str]] = None,
                    exclude_patterns: Optional[List[str]] = None,
                    css_selector: Optional[str] = None,
                    word_count_threshold: int = 200,
                    bypass_cache: bool = False) -> dict: ...
    async def extract(self, url: str, *, schema: dict, schema_type: str = "css") -> dict: ...
    async def map_urls(self, url: str, *, source: str = "sitemap", pattern: str = "*",
                       max_urls: int = 100, query: Optional[str] = None,
                       score_threshold: Optional[float] = None) -> dict: ...
    async def screenshot(self, url: str, *, wait_for: Optional[str] = None,
                         css_selector: Optional[str] = None, full_page: bool = True) -> dict: ...
    async def generate_schema(self, *, url: Optional[str] = None, html: Optional[str] = None,
                              query: str, schema_type: str = "css") -> dict: ...
    async def list_profiles(self) -> dict: ...
    async def create_profile(self, *, profile_name: Optional[str] = None) -> dict: ...


def get_backend(config: PluginConfig) -> CrawlBackend:
    """Factory: return the appropriate backend based on config.mode."""
    if config.mode == "local":
        from .local import LocalBackend
        return LocalBackend(config)
    else:
        from .cloud import CloudBackend
        return CloudBackend(config)
