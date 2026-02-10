"""Cloud backend — wraps crawl4ai_cloud.AsyncWebCrawler."""
from __future__ import annotations

import io
import json
import zipfile
from typing import Any, Dict, List, Optional

from ..config import PluginConfig
from . import BackendError


class CloudBackend:
    """Backend that uses the Crawl4AI Cloud API via the SDK."""

    def __init__(self, config: PluginConfig):
        self._config = config
        self._crawler = None

    async def startup(self) -> None:
        from crawl4ai_cloud import AsyncWebCrawler
        if not self._config.api_key:
            raise BackendError(
                "API key required. Set CRAWL4AI_API_KEY env var or "
                "add api_key to ~/.crawl4ai/claude_config.json"
            )
        self._crawler = AsyncWebCrawler(
            api_key=self._config.api_key,
            base_url=self._config.api_base_url,
            timeout=self._config.default_timeout,
        )

    async def shutdown(self) -> None:
        if self._crawler:
            await self._crawler.close()
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

    async def _single_crawl(self, url: str, *, css_selector=None,
                            word_count_threshold=200, bypass_cache=False) -> dict:
        from crawl4ai_cloud import CrawlerRunConfig
        config = CrawlerRunConfig(
            css_selector=css_selector,
            word_count_threshold=word_count_threshold,
        )
        result = await self._crawler.run(url, config=config, bypass_cache=bypass_cache)
        return self._normalize_crawl_result(result)

    async def _deep_crawl(self, url: str, *, strategy, max_depth, max_pages,
                          include_patterns, exclude_patterns, bypass_cache) -> dict:
        # Non-blocking: start crawl and return job info immediately
        result = await self._crawler.deep_crawl(
            url,
            strategy=strategy,
            max_depth=max_depth,
            max_urls=max_pages,
            include_patterns=include_patterns,
            exclude_patterns=exclude_patterns,
            bypass_cache=bypass_cache,
            wait=False,
            scan_only=False,
        )
        return {
            "job_id": result.job_id,
            "status": result.status if hasattr(result, "status") else "started",
            "discovered_count": result.discovered_count if hasattr(result, "discovered_count") else 0,
            "discovered_urls": (result.discovered_urls or [])[:20] if hasattr(result, "discovered_urls") else [],
            "message": "Deep crawl started. Use job_status tool to check progress, or run crawl4ai-poll in background.",
        }

    # ------------------------------------------------------------------
    # job_status
    # ------------------------------------------------------------------
    async def job_status(self, job_id: str) -> dict:
        crawler = self._ensure_crawler()
        try:
            if job_id.startswith("scan_"):
                result = await crawler.get_deep_crawl_status(job_id)
                data = {
                    "job_id": job_id,
                    "status": result.status if hasattr(result, "status") else "unknown",
                    "discovered_count": result.discovered_count if hasattr(result, "discovered_count") else 0,
                }
                if hasattr(result, "crawl_job_id") and result.crawl_job_id:
                    data["crawl_job_id"] = result.crawl_job_id
                    data["message"] = f"Scan complete. Crawl job started: {result.crawl_job_id}. Poll that job_id for final results."
                if hasattr(result, "is_complete") and result.is_complete:
                    data["is_complete"] = True
                return data
            else:
                job = await crawler.get_job(job_id)
                data = {
                    "job_id": job_id,
                    "status": job.status if hasattr(job, "status") else "unknown",
                }
                if hasattr(job, "progress") and job.progress:
                    prog = job.progress
                    if hasattr(prog, "total"):
                        data["progress"] = {
                            "total": prog.total,
                            "completed": prog.completed if hasattr(prog, "completed") else 0,
                            "failed": prog.failed if hasattr(prog, "failed") else 0,
                        }
                    else:
                        data["progress"] = str(prog)
                is_complete = (hasattr(job, "is_complete") and job.is_complete) or \
                              (hasattr(job, "status") and job.status in ("completed", "done"))
                if is_complete:
                    data["is_complete"] = True
                    dl_url = await crawler.download_url(job_id)
                    data["download_url"] = dl_url
                    data["message"] = "Job complete. Use fetch tool with the download_url to get results."
                return data
        except Exception as e:
            raise BackendError(f"Job status check failed: {e}") from e

    # ------------------------------------------------------------------
    # fetch_results
    # ------------------------------------------------------------------
    async def fetch_results(self, download_url: str) -> dict:
        try:
            pages = await self._download_from_url(download_url)
            return {
                "pages_crawled": len(pages),
                "results": pages,
            }
        except Exception as e:
            raise BackendError(f"Fetch results failed: {e}") from e

    async def _download_from_url(self, url: str) -> List[dict]:
        """Download results ZIP from a presigned URL and parse each JSON file."""
        import httpx

        async with httpx.AsyncClient() as client:
            resp = await client.get(url, timeout=60.0)
            resp.raise_for_status()

        pages = []
        with zipfile.ZipFile(io.BytesIO(resp.content)) as zf:
            for name in sorted(zf.namelist()):
                if not name.endswith(".json"):
                    continue
                data = json.loads(zf.read(name))
                md = data.get("markdown", {})
                raw = md.get("raw_markdown", "") if isinstance(md, dict) else str(md) if md else ""
                fit = md.get("fit_markdown") if isinstance(md, dict) else None
                pages.append({
                    "url": data.get("url", ""),
                    "markdown": raw,
                    "fit_markdown": fit,
                    "metadata": data.get("metadata", {}),
                    "links": data.get("links", {}),
                    "status_code": data.get("status_code"),
                })
        return pages

    @staticmethod
    def _normalize_crawl_result(result) -> dict:
        """Normalize a CrawlResult into the standard dict format."""
        md = result.markdown
        return {
            "url": result.url,
            "markdown": md.raw_markdown if md else None,
            "fit_markdown": md.fit_markdown if md else None,
            "metadata": result.metadata or {},
            "links": result.links or {},
            "status_code": result.status_code,
        }

    # ------------------------------------------------------------------
    # extract
    # ------------------------------------------------------------------
    async def extract(self, url: str, *, schema: dict, schema_type: str = "css") -> dict:
        crawler = self._ensure_crawler()
        try:
            from crawl4ai_cloud import CrawlerRunConfig
            # Cloud API format: type is "json_css"/"json_xpath", schema at top level
            extraction = {
                "type": "json_css" if schema_type.lower() == "css" else "json_xpath",
                "schema": schema,
            }
            config = CrawlerRunConfig(extraction_strategy=extraction)
            result = await crawler.run(url, config=config)
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
            result = await crawler.deep_crawl(
                url,
                strategy="map",
                source=source,
                pattern=pattern,
                max_urls=max_urls,
                query=query,
                score_threshold=score_threshold,
                scan_only=True,
                wait=True,
            )
            urls = result.discovered_urls if hasattr(result, 'discovered_urls') else []
            return {
                "domain": url,
                "urls": urls,
                "total_discovered": result.discovered_count if hasattr(result, 'discovered_count') else len(urls),
                "returned": len(urls),
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
            from crawl4ai_cloud import CrawlerRunConfig
            config = CrawlerRunConfig(
                screenshot=True,
                screenshot_wait_for=wait_for,
                scan_full_page=full_page,
            )
            result = await crawler.run(url, config=config)
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
        crawler = self._ensure_crawler()
        try:
            kwargs: Dict[str, Any] = {
                "query": query,
                "schema_type": schema_type.upper(),
            }
            if url:
                kwargs["urls"] = [url]
            elif html:
                kwargs["html"] = html
            else:
                raise BackendError("Either url or html must be provided")

            result = await crawler.generate_schema(**kwargs)
            if not result.success:
                raise BackendError(f"Schema generation failed: {result.error}")
            return {
                "schema_type": schema_type.lower(),
                "schema": result.schema,
            }
        except BackendError:
            raise
        except Exception as e:
            raise BackendError(f"Schema generation failed: {e}") from e

    # ------------------------------------------------------------------
    # profiles (cloud: informational only)
    # ------------------------------------------------------------------
    async def list_profiles(self) -> dict:
        return {
            "mode": "cloud",
            "profiles": [],
            "message": "Cloud profiles are managed via the Crawl4AI dashboard or CLI: crwl cloud profiles list",
        }

    async def create_profile(self, *, profile_name: Optional[str] = None) -> dict:
        return {
            "mode": "cloud",
            "message": "Cloud profiles are managed via the Crawl4AI dashboard or CLI: crwl cloud profiles upload",
        }
