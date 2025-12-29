"""
Crawl4AI Cloud SDK - Lightweight cloud client for Crawl4AI API.

Example:
    ```python
    from crawl4ai_cloud import AsyncWebCrawler, CrawlerRunConfig

    async with AsyncWebCrawler(api_key="sk_live_xxx") as crawler:
        result = await crawler.run("https://example.com")
        print(result.markdown.raw_markdown)
    ```
"""

__version__ = "0.1.0"

# Main crawler class
from .crawler import AsyncWebCrawler

# Configuration classes
from .configs import (
    CrawlerRunConfig,
    BrowserConfig,
    build_crawl_request,
    sanitize_crawler_config,
    sanitize_browser_config,
    normalize_proxy,
)

# Response models
from .models import (
    CrawlResult,
    CrawlJob,
    JobProgress,
    MarkdownResult,
    DeepCrawlResult,
    ScanUrlInfo,
    ContextResult,
    GeneratedSchema,
    StorageUsage,
    ProxyConfig,
    LLMUsage,
)

# Errors
from .errors import (
    CloudError,
    AuthenticationError,
    RateLimitError,
    QuotaExceededError,
    NotFoundError,
    ValidationError,
    TimeoutError,
    ServerError,
)

__all__ = [
    # Version
    "__version__",
    # Main class
    "AsyncWebCrawler",
    # Configs
    "CrawlerRunConfig",
    "BrowserConfig",
    "build_crawl_request",
    "sanitize_crawler_config",
    "sanitize_browser_config",
    "normalize_proxy",
    # Models
    "CrawlResult",
    "CrawlJob",
    "JobProgress",
    "MarkdownResult",
    "DeepCrawlResult",
    "ScanUrlInfo",
    "ContextResult",
    "GeneratedSchema",
    "StorageUsage",
    "ProxyConfig",
    "LLMUsage",
    # Errors
    "CloudError",
    "AuthenticationError",
    "RateLimitError",
    "QuotaExceededError",
    "NotFoundError",
    "ValidationError",
    "TimeoutError",
    "ServerError",
]
