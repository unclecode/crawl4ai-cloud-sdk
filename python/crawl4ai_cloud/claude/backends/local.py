"""Local backend — wraps OSS crawl4ai.AsyncWebCrawler (dynamic import)."""
from __future__ import annotations

import json
from typing import Any, Dict, List, Optional

from ..config import PluginConfig
from . import BackendError

_INSTALL_MSG = (
    "crawl4ai is not installed. Install with:\n"
    "  pip install crawl4ai && crawl4ai-setup"
)


def _import_crawl4ai():
    """Dynamically import crawl4ai OSS, raising BackendError if missing."""
    try:
        import crawl4ai
        return crawl4ai
    except ImportError:
        raise BackendError(_INSTALL_MSG)


class LocalBackend:
    """Backend using OSS crawl4ai with a local browser."""

    def __init__(self, config: PluginConfig):
        self._config = config
        self._crawler = None

    async def startup(self) -> None:
        c4 = _import_crawl4ai()
        from crawl4ai import AsyncWebCrawler, BrowserConfig
        browser_cfg = BrowserConfig(
            headless=self._config.headless,
            browser_type=self._config.browser_type,
            verbose=self._config.verbose,
        )
        self._crawler = AsyncWebCrawler(config=browser_cfg)
        await self._crawler.__aenter__()

    async def shutdown(self) -> None:
        if self._crawler:
            await self._crawler.__aexit__(None, None, None)
            self._crawler = None

    def _ensure_crawler(self):
        if not self._crawler:
            raise BackendError("Backend not started. Call startup() first.")
        return self._crawler

    # ------------------------------------------------------------------
    # crawl
    # ------------------------------------------------------------------
    async def crawl(
        self,
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
    ) -> dict:
        crawler = self._ensure_crawler()
        try:
            if deep_crawl:
                return await self._deep_crawl(
                    url, strategy=strategy, max_depth=max_depth,
                    max_pages=max_pages, include_patterns=include_patterns,
                    exclude_patterns=exclude_patterns, bypass_cache=bypass_cache,
                )
            else:
                return await self._single_crawl(
                    url, css_selector=css_selector,
                    word_count_threshold=word_count_threshold,
                    bypass_cache=bypass_cache,
                )
        except BackendError:
            raise
        except Exception as e:
            raise BackendError(f"Crawl failed: {e}") from e

    async def _single_crawl(self, url, *, css_selector=None,
                            word_count_threshold=200, bypass_cache=False) -> dict:
        from crawl4ai import CrawlerRunConfig
        config = CrawlerRunConfig(
            css_selector=css_selector,
            word_count_threshold=word_count_threshold,
        )
        result = await self._crawler.arun(url, config=config)
        return self._normalize_result(result)

    async def _deep_crawl(self, url, *, strategy, max_depth, max_pages,
                          include_patterns, exclude_patterns, bypass_cache) -> dict:
        from crawl4ai import CrawlerRunConfig
        from crawl4ai.deep_crawling import BFSDeepCrawlStrategy

        strategy_map = {"bfs": BFSDeepCrawlStrategy}
        # Try to import other strategies if available
        try:
            from crawl4ai.deep_crawling import DFSDeepCrawlStrategy
            strategy_map["dfs"] = DFSDeepCrawlStrategy
        except ImportError:
            pass
        try:
            from crawl4ai.deep_crawling import BestFirstCrawlStrategy
            strategy_map["best_first"] = BestFirstCrawlStrategy
        except ImportError:
            pass

        strategy_cls = strategy_map.get(strategy)
        if not strategy_cls:
            raise BackendError(f"Unknown deep crawl strategy: {strategy}. Available: {list(strategy_map.keys())}")

        deep_strategy = strategy_cls(max_depth=max_depth, max_pages=max_pages)

        # Apply URL filters if the strategy supports them
        if include_patterns and hasattr(deep_strategy, 'include_patterns'):
            deep_strategy.include_patterns = include_patterns
        if exclude_patterns and hasattr(deep_strategy, 'exclude_patterns'):
            deep_strategy.exclude_patterns = exclude_patterns

        config = CrawlerRunConfig(deep_crawl_strategy=deep_strategy)
        results = await self._crawler.arun(url, config=config)

        # arun with deep_crawl_strategy returns a list
        if isinstance(results, list):
            pages = [self._normalize_result(r) for r in results]
        else:
            pages = [self._normalize_result(results)]

        return {
            "pages_crawled": len(pages),
            "results": pages,
        }

    @staticmethod
    def _normalize_result(result) -> dict:
        """Normalize an OSS CrawlResult into the standard dict format."""
        md = result.markdown
        raw = md.raw_markdown if hasattr(md, 'raw_markdown') else str(md) if md else None
        fit = md.fit_markdown if hasattr(md, 'fit_markdown') else None
        return {
            "url": result.url,
            "markdown": raw,
            "fit_markdown": fit,
            "metadata": result.metadata if hasattr(result, 'metadata') else {},
            "links": result.links if hasattr(result, 'links') else {},
            "status_code": result.status_code if hasattr(result, 'status_code') else None,
        }

    # ------------------------------------------------------------------
    # extract
    # ------------------------------------------------------------------
    async def extract(self, url: str, *, schema: dict, schema_type: str = "css") -> dict:
        crawler = self._ensure_crawler()
        try:
            from crawl4ai import CrawlerRunConfig
            from crawl4ai.extraction_strategy import JsonCssExtractionStrategy

            extraction = JsonCssExtractionStrategy(schema=schema)
            config = CrawlerRunConfig(extraction_strategy=extraction)
            result = await crawler.arun(url, config=config)

            extracted = []
            if result.extracted_content:
                try:
                    extracted = json.loads(result.extracted_content)
                except (json.JSONDecodeError, TypeError):
                    extracted = [{"raw": result.extracted_content}]

            return {
                "url": result.url,
                "extracted": extracted if isinstance(extracted, list) else [extracted],
                "items_count": len(extracted) if isinstance(extracted, list) else 1,
            }
        except BackendError:
            raise
        except Exception as e:
            raise BackendError(f"Extract failed: {e}") from e

    # ------------------------------------------------------------------
    # map_urls
    # ------------------------------------------------------------------
    async def map_urls(self, url: str, *, source: str = "sitemap", pattern: str = "*",
                       max_urls: int = 100, query: Optional[str] = None,
                       score_threshold: Optional[float] = None) -> dict:
        crawler = self._ensure_crawler()
        try:
            from crawl4ai import SeedingConfig

            seeding_config = SeedingConfig(
                source=source,
                pattern=pattern,
                max_urls=max_urls,
            )
            if query:
                seeding_config.query = query

            urls = await crawler.aseed_urls(url, config=seeding_config)
            if isinstance(urls, dict):
                url_list = urls.get("urls", [])
            elif isinstance(urls, list):
                url_list = urls
            else:
                url_list = []

            return {
                "domain": url,
                "urls": url_list[:max_urls],
                "total_discovered": len(url_list),
                "returned": min(len(url_list), max_urls),
                "source": source,
            }
        except BackendError:
            raise
        except Exception as e:
            raise BackendError(f"Map failed: {e}") from e

    # ------------------------------------------------------------------
    # screenshot
    # ------------------------------------------------------------------
    async def screenshot(self, url: str, *, wait_for: Optional[str] = None,
                         css_selector: Optional[str] = None,
                         full_page: bool = True) -> dict:
        crawler = self._ensure_crawler()
        try:
            from crawl4ai import CrawlerRunConfig
            config = CrawlerRunConfig(
                screenshot=True,
                screenshot_wait_for=wait_for,
                scan_full_page=full_page,
            )
            result = await crawler.arun(url, config=config)
            if not result.screenshot:
                raise BackendError("Screenshot capture returned empty result")
            return {
                "url": result.url,
                "screenshot_base64": result.screenshot,
                "format": "png",
            }
        except BackendError:
            raise
        except Exception as e:
            raise BackendError(f"Screenshot failed: {e}") from e

    # ------------------------------------------------------------------
    # generate_schema
    # ------------------------------------------------------------------
    async def generate_schema(self, *, url: Optional[str] = None,
                              html: Optional[str] = None,
                              query: str, schema_type: str = "css") -> dict:
        _import_crawl4ai()
        try:
            from crawl4ai.extraction_strategy import JsonCssExtractionStrategy

            kwargs: Dict[str, Any] = {"query": query, "schema_type": schema_type.upper()}
            if url:
                kwargs["url"] = url
            elif html:
                kwargs["html"] = html
            else:
                raise BackendError("Either url or html must be provided")

            schema = JsonCssExtractionStrategy.generate_schema(**kwargs)
            return {
                "schema_type": schema_type.lower(),
                "schema": schema,
            }
        except BackendError:
            raise
        except Exception as e:
            raise BackendError(f"Schema generation failed: {e}") from e

    # ------------------------------------------------------------------
    # profiles
    # ------------------------------------------------------------------
    async def list_profiles(self) -> dict:
        _import_crawl4ai()
        try:
            from crawl4ai.browser_profiler import BrowserProfiler
            profiler = BrowserProfiler()
            profiles = profiler.list_profiles()
            return {
                "mode": "local",
                "profiles": profiles if isinstance(profiles, list) else [],
            }
        except Exception as e:
            raise BackendError(f"Failed to list profiles: {e}") from e

    async def create_profile(self, *, profile_name: Optional[str] = None) -> dict:
        _import_crawl4ai()
        try:
            from crawl4ai.browser_profiler import BrowserProfiler
            profiler = BrowserProfiler()
            path = profiler.create_profile(profile_name=profile_name)
            return {
                "profile_name": profile_name or "default",
                "profile_path": str(path),
                "message": "Profile created.",
            }
        except Exception as e:
            raise BackendError(f"Failed to create profile: {e}") from e
