"""
Crawl4AI Cloud SDK - Lightweight cloud client for Crawl4AI API.

Example:
    ```python
    from crawl4ai_cloud import AsyncWebCrawler

    async with AsyncWebCrawler(api_key="sk_live_xxx") as crawler:
        # Simple wrapper endpoints
        md = await crawler.markdown("https://example.com")
        ss = await crawler.screenshot("https://example.com")
        data = await crawler.extract("https://example.com", query="get products")
        urls = await crawler.map("https://example.com")

        # Full power endpoint (unchanged)
        result = await crawler.run("https://example.com")
    ```
"""

__version__ = "0.8.0"

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
    normalize_url,
)

# Response models
from .models import (
    CrawlResult,
    CrawlJob,
    JobProgress,
    MarkdownResult,
    DeepCrawlResult,
    ScanUrlInfo,
    ScanResult,
    ScanJobStatus,
    DomainScanUrlInfo,
    SiteScanConfig,
    SiteExtractConfig,
    GeneratedConfig,
    ContextResult,
    GeneratedSchema,
    StorageUsage,
    ProxyConfig,
    LLMUsage,
    # Usage metrics
    Usage,
    CrawlUsageMetrics,
    LLMUsageMetrics,
    StorageUsageMetrics,
    # Wrapper API models
    WrapperUsage,
    MarkdownResponse,
    ScreenshotResponse,
    ExtractResponse,
    MapUrlInfo,
    MapResponse,
    SiteCrawlResponse,
    SiteCrawlJobStatus,
    SiteCrawlProgress,
    WrapperJob,
    WrapperJobProgress,
    UrlStatus,
    # Enrich v2 API models
    EnrichEntity,
    EnrichCriterion,
    EnrichFeature,
    EnrichPlan,
    EnrichUrlCandidate,
    EnrichRow,
    EnrichPhaseData,
    EnrichProgress,
    EnrichLlmBucket,
    EnrichUsage,
    EnrichJobStatus,
    EnrichJobListItem,
    EnrichEvent,
    ENRICH_TERMINAL_STATUSES,
    ENRICH_PAUSED_STATUSES,
    # Discovery / Search
    Sitelink,
    SearchHit,
    FeaturedSnippet,
    PaaItem,
    KnowledgeGraph,
    AiOverview,
    ResultStats,
    Pagination,
    SearchMetadata,
    SearchResponse,
    DiscoveryService,
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
    "__version__",
    "AsyncWebCrawler",
    # Configs
    "CrawlerRunConfig", "BrowserConfig",
    "build_crawl_request", "sanitize_crawler_config", "sanitize_browser_config",
    "normalize_proxy", "normalize_url",
    # Core models
    "CrawlResult", "CrawlJob", "JobProgress", "MarkdownResult",
    "DeepCrawlResult", "ScanUrlInfo", "ScanResult", "ScanJobStatus",
    "DomainScanUrlInfo", "SiteScanConfig", "SiteExtractConfig", "GeneratedConfig",
    "ContextResult", "GeneratedSchema", "StorageUsage", "ProxyConfig", "LLMUsage",
    "Usage", "CrawlUsageMetrics", "LLMUsageMetrics", "StorageUsageMetrics",
    # Wrapper API models
    "WrapperUsage", "MarkdownResponse", "ScreenshotResponse", "ExtractResponse",
    "MapUrlInfo", "MapResponse", "SiteCrawlResponse", "SiteCrawlJobStatus",
    "SiteCrawlProgress", "WrapperJob", "WrapperJobProgress", "UrlStatus",
    # Enrich v2
    "EnrichEntity", "EnrichCriterion", "EnrichFeature", "EnrichPlan",
    "EnrichUrlCandidate", "EnrichRow", "EnrichPhaseData", "EnrichProgress",
    "EnrichLlmBucket", "EnrichUsage", "EnrichJobStatus", "EnrichJobListItem",
    "EnrichEvent",
    "ENRICH_TERMINAL_STATUSES", "ENRICH_PAUSED_STATUSES",
    # Errors
    "CloudError", "AuthenticationError", "RateLimitError", "QuotaExceededError",
    "NotFoundError", "ValidationError", "TimeoutError", "ServerError",
]
