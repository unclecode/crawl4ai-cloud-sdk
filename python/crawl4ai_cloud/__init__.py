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

__version__ = "0.14.0"

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
    SynthesizedAnswer,
    RubricScore,
    UsageComponent,
    SearchUsage,
    SearchResponse,
    DiscoveryService,
    DiscoveryJobHandle,
    DiscoveryJobStatus,
)

# Context v2 — four-pillar pipeline. Replaces the old PAA-based ContextResult.
from .context import (
    # Pillar builders
    Source,
    Strategy,
    Shape,  # back-compat alias for Synthesizer
    Synthesizer,
    Reconciler,
    # Knobs
    Constraints,
    # Output types
    ContextItem,
    ContextOutput,
    MarkdownFile,
    # Result + events
    ContextResult,
    StatusEvent,
    PhaseProgressInit,
    PhaseProgressItemUpdate,
    TerminalEvent,
    ContextEvent,
    # Versions / diff / catalog
    ContextVersion,
    ContextDiff,
    CatalogEntry,
    ContextCatalog,
    # Constants
    TERMINAL_STATUSES as CONTEXT_TERMINAL_STATUSES,
    PHASE_PLANNING,
    PHASE_CRAWLING,
    PHASE_SHAPING,
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
    # Discovery / search synthesis (v1, since 0.10.0)
    "Sitelink", "SearchHit", "FeaturedSnippet", "PaaItem",
    "KnowledgeGraph", "AiOverview", "ResultStats", "Pagination",
    "SearchMetadata", "SynthesizedAnswer", "RubricScore",
    "UsageComponent", "SearchUsage", "SearchResponse",
    "DiscoveryService", "DiscoveryJobHandle", "DiscoveryJobStatus",
    # Configs
    "CrawlerRunConfig", "BrowserConfig",
    "build_crawl_request", "sanitize_crawler_config", "sanitize_browser_config",
    "normalize_proxy", "normalize_url",
    # Core models
    "CrawlResult", "CrawlJob", "JobProgress", "MarkdownResult",
    "DeepCrawlResult", "ScanUrlInfo", "ScanResult", "ScanJobStatus",
    "DomainScanUrlInfo", "SiteScanConfig", "SiteExtractConfig", "GeneratedConfig",
    "GeneratedSchema", "StorageUsage", "ProxyConfig", "LLMUsage",
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
    # Context v2 — four-pillar pipeline
    "Source", "Strategy", "Shape", "Synthesizer", "Reconciler",
    "Constraints",
    "ContextItem", "ContextOutput", "MarkdownFile",
    "ContextResult",
    "StatusEvent", "PhaseProgressInit", "PhaseProgressItemUpdate", "TerminalEvent",
    "ContextEvent",
    "ContextVersion", "ContextDiff", "CatalogEntry", "ContextCatalog",
    "CONTEXT_TERMINAL_STATUSES",
    "PHASE_PLANNING", "PHASE_CRAWLING", "PHASE_SHAPING",
    # Errors
    "CloudError", "AuthenticationError", "RateLimitError", "QuotaExceededError",
    "NotFoundError", "ValidationError", "TimeoutError", "ServerError",
]
