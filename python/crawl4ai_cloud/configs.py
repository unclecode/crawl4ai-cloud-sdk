"""Configuration classes and sanitization for Crawl4AI Cloud SDK."""
import warnings
from dataclasses import dataclass, field, asdict
from typing import Dict, Any, Optional, Union, List

from .models import ProxyConfig


# Fields that cloud controls - removed from CrawlerRunConfig
CRAWLER_CONFIG_SANITIZE_FIELDS = [
    "cache_mode",
    "session_id",
    "bypass_cache",
    "no_cache_read",
    "no_cache_write",
    "disable_cache",
]

# Fields that cloud controls - removed from BrowserConfig
BROWSER_CONFIG_SANITIZE_FIELDS = [
    "cdp_url",
    "create_isolated_context",
    "cdp_cleanup_on_close",
    "browser_context_id",
    "target_id",
    "use_managed_browser",
    "browser_mode",
    "user_data_dir",
    "chrome_channel",
]


@dataclass
class CrawlerRunConfig:
    """
    Configuration for crawl requests. Mirrors OSS CrawlerRunConfig.

    Cloud-unsupported fields are accepted but silently stripped when sent to API.
    This ensures OSS code works without modification.

    Example:
        config = CrawlerRunConfig(
            word_count_threshold=10,
            exclude_external_links=True,
            process_iframes=True,
        )
        result = await crawler.run(url, config=config)
    """
    # Content processing
    word_count_threshold: int = 200
    exclude_external_links: bool = False
    exclude_social_media_links: bool = False
    exclude_external_images: bool = False
    exclude_domains: List[str] = field(default_factory=list)

    # HTML processing
    process_iframes: bool = False
    remove_forms: bool = False
    keep_data_attributes: bool = False

    # Output options
    only_text: bool = False
    prettiify: bool = False  # Note: typo matches OSS

    # Screenshot/PDF
    screenshot: bool = False
    screenshot_wait_for: Optional[str] = None
    pdf: bool = False

    # Extraction
    extraction_strategy: Optional[Any] = None
    chunking_strategy: Optional[Any] = None
    content_filter: Optional[Any] = None

    # Markdown generation
    markdown_generator: Optional[Any] = None

    # Wait conditions
    wait_for: Optional[str] = None
    delay_before_return_html: float = 0.0

    # Page interaction
    js_code: Optional[Union[str, List[str]]] = None
    js_only: bool = False
    ignore_body_visibility: bool = True
    scan_full_page: bool = False
    scroll_delay: float = 0.2

    # Network
    wait_for_images: bool = False
    adjust_viewport_to_content: bool = False
    page_timeout: int = 60000

    # Cache (cloud-controlled, will be stripped)
    cache_mode: Optional[str] = None
    session_id: Optional[str] = None
    bypass_cache: bool = False
    no_cache_read: bool = False
    no_cache_write: bool = False
    disable_cache: bool = False

    # Magic mode
    magic: bool = False

    # Simulate user
    simulate_user: bool = False
    override_navigator: bool = False

    # Additional fields stored as extras
    _extras: Dict[str, Any] = field(default_factory=dict, repr=False)

    def __post_init__(self):
        """Handle any post-initialization processing."""
        pass

    def dump(self) -> Dict[str, Any]:
        """Serialize config to dict format expected by API."""
        data = asdict(self)
        # Remove private fields
        data.pop("_extras", None)
        # Remove None values
        return {k: v for k, v in data.items() if v is not None}


@dataclass
class BrowserConfig:
    """
    Browser configuration for crawl requests. Mirrors OSS BrowserConfig.

    Cloud-unsupported fields are accepted but silently stripped when sent to API.

    Example:
        browser_config = BrowserConfig(
            headless=True,
            viewport_width=1920,
            viewport_height=1080,
        )
        result = await crawler.run(url, browser_config=browser_config)
    """
    # Browser settings
    headless: bool = True
    browser_type: str = "chromium"
    verbose: bool = False

    # Viewport
    viewport_width: int = 1080
    viewport_height: int = 600

    # User agent
    user_agent: Optional[str] = None
    user_agent_mode: Optional[str] = None
    user_agent_generator_config: Optional[Dict[str, Any]] = None

    # Headers & cookies
    headers: Dict[str, str] = field(default_factory=dict)
    cookies: List[Dict[str, Any]] = field(default_factory=list)

    # Storage state
    storage_state: Optional[str] = None

    # Proxy (handled separately in cloud)
    proxy: Optional[str] = None
    proxy_config: Optional[Dict[str, Any]] = None

    # Browser args
    extra_args: List[str] = field(default_factory=list)
    chrome_channel: Optional[str] = None
    accept_downloads: bool = False
    downloads_path: Optional[str] = None

    # Ignore HTTPS errors
    ignore_https_errors: bool = True
    java_script_enabled: bool = True

    # Cloud-controlled fields (will be stripped)
    cdp_url: Optional[str] = None
    use_managed_browser: bool = False
    browser_mode: Optional[str] = None
    user_data_dir: Optional[str] = None

    # Text mode
    text_mode: bool = False
    light_mode: bool = False

    def dump(self) -> Dict[str, Any]:
        """Serialize config to dict format expected by API."""
        data = asdict(self)
        # Remove None values
        return {k: v for k, v in data.items() if v is not None}


def sanitize_crawler_config(config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]]) -> Dict[str, Any]:
    """
    Sanitize CrawlerRunConfig for cloud API.

    Removes cloud-controlled fields and returns a dict suitable for the API.
    """
    if config is None:
        return {}

    # Get dict representation
    if hasattr(config, "dump"):
        data = config.dump()
        # Handle {type, params} structure from OSS dump()
        if isinstance(data, dict) and "params" in data:
            data = data.get("params", {})
    elif isinstance(config, dict):
        data = config.copy()
    else:
        return {}

    # Remove cloud-controlled fields
    for fld in CRAWLER_CONFIG_SANITIZE_FIELDS:
        data.pop(fld, None)

    # Flatten serialized nested objects
    data = _flatten_serialized_objects(data)

    return data


def sanitize_browser_config(
    config: Optional[Union[BrowserConfig, Dict[str, Any]]],
    strategy: str = "browser"
) -> Dict[str, Any]:
    """
    Sanitize BrowserConfig for cloud API.

    Removes cloud-controlled fields. Warns if browser config provided with HTTP strategy.
    """
    if config is None:
        return {}

    # Warn if browser config with HTTP strategy
    if strategy == "http" and config:
        warnings.warn(
            "browser_config is ignored when using strategy='http'. "
            "Browser configuration only applies to browser-based crawling.",
            UserWarning,
        )
        return {}

    # Get dict representation
    if hasattr(config, "dump"):
        data = config.dump()
        if isinstance(data, dict) and "params" in data:
            data = data.get("params", {})
    elif isinstance(config, dict):
        data = config.copy()
    else:
        return {}

    # Remove cloud-controlled fields
    for fld in BROWSER_CONFIG_SANITIZE_FIELDS:
        data.pop(fld, None)

    data = _flatten_serialized_objects(data)

    return data


def _flatten_serialized_objects(data: Dict[str, Any]) -> Dict[str, Any]:
    """Flatten serialized objects with {type, params} structure."""
    # Cloud API understands this format, return as-is
    return data


def normalize_proxy(
    proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]]
) -> Optional[Dict[str, Any]]:
    """
    Normalize proxy configuration to dict format.

    Accepts:
    - None: No proxy
    - str: Shorthand mode ("datacenter", "residential", "auto")
    - dict: Full config {"mode": "...", "country": "...", "sticky_session": ...}
    - ProxyConfig: Dataclass instance

    Examples:
        >>> normalize_proxy("datacenter")
        {"mode": "datacenter"}

        >>> normalize_proxy({"mode": "residential", "country": "US"})
        {"mode": "residential", "country": "US"}
    """
    if proxy is None:
        return None

    if isinstance(proxy, str):
        return {"mode": proxy}

    if isinstance(proxy, ProxyConfig):
        return proxy.to_dict()

    if isinstance(proxy, dict):
        return proxy

    raise ValueError(
        f"Invalid proxy type: {type(proxy)}. "
        "Expected str, dict, or ProxyConfig."
    )


def build_crawl_request(
    url: Optional[str] = None,
    urls: Optional[List[str]] = None,
    config: Optional[Union[CrawlerRunConfig, Dict[str, Any]]] = None,
    browser_config: Optional[Union[BrowserConfig, Dict[str, Any]]] = None,
    strategy: str = "browser",
    proxy: Optional[Union[str, Dict[str, Any], ProxyConfig]] = None,
    bypass_cache: bool = False,
    **kwargs,
) -> Dict[str, Any]:
    """
    Build a crawl request body for the cloud API.

    Args:
        url: Single URL to crawl
        urls: List of URLs to crawl
        config: CrawlerRunConfig instance or dict
        browser_config: BrowserConfig instance or dict
        strategy: Crawl strategy ("browser" or "http")
        proxy: Proxy configuration
        bypass_cache: Skip cache lookup
        **kwargs: Additional API parameters

    Returns:
        Dict request body for cloud API
    """
    body: Dict[str, Any] = {"strategy": strategy}

    if url:
        body["url"] = url
    if urls:
        body["urls"] = urls

    # Sanitize and add configs
    crawler_config = sanitize_crawler_config(config)
    if crawler_config:
        body["crawler_config"] = crawler_config

    browser_cfg = sanitize_browser_config(browser_config, strategy)
    if browser_cfg:
        body["browser_config"] = browser_cfg

    # Normalize and add proxy
    proxy_config = normalize_proxy(proxy)
    if proxy_config:
        body["proxy"] = proxy_config

    if bypass_cache:
        body["bypass_cache"] = True

    # Add any additional kwargs
    body.update(kwargs)

    return body
