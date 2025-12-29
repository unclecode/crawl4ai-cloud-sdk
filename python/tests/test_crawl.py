"""Tests for basic crawl operations."""
import pytest
import pytest_asyncio

from crawl4ai_cloud import (
    AsyncWebCrawler,
    CrawlerRunConfig,
    BrowserConfig,
    CrawlResult,
    CloudError,
    AuthenticationError,
)


class TestAsyncWebCrawler:
    """Tests for AsyncWebCrawler class."""

    def test_init_with_api_key(self, api_key):
        """Test initialization with API key."""
        crawler = AsyncWebCrawler(api_key=api_key)
        assert crawler is not None

    def test_init_missing_api_key(self, monkeypatch):
        """Test initialization fails without API key."""
        monkeypatch.delenv("CRAWL4AI_API_KEY", raising=False)
        with pytest.raises(ValueError, match="API key is required"):
            AsyncWebCrawler(api_key=None)

    def test_init_invalid_api_key(self):
        """Test initialization fails with invalid API key format."""
        with pytest.raises(ValueError, match="Invalid API key format"):
            AsyncWebCrawler(api_key="invalid_key")

    @pytest.mark.asyncio
    async def test_run_single_url(self, api_key, test_url):
        """Test crawling a single URL."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(test_url)

            assert isinstance(result, CrawlResult)
            assert result.success is True
            assert result.url == test_url
            assert result.markdown is not None
            assert result.markdown.raw_markdown is not None
            assert len(result.markdown.raw_markdown) > 0

    @pytest.mark.asyncio
    async def test_arun_alias(self, api_key, test_url):
        """Test arun() alias works same as run()."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.arun(test_url)

            assert isinstance(result, CrawlResult)
            assert result.success is True
            assert result.url == test_url

    @pytest.mark.asyncio
    async def test_run_with_config(self, api_key, test_url):
        """Test crawling with CrawlerRunConfig."""
        config = CrawlerRunConfig(
            word_count_threshold=10,
            exclude_external_links=True,
        )

        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(test_url, config=config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_run_with_browser_config(self, api_key, test_url):
        """Test crawling with BrowserConfig."""
        browser_config = BrowserConfig(
            headless=True,
            viewport_width=1920,
            viewport_height=1080,
        )

        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(test_url, browser_config=browser_config)

            assert result.success is True

    @pytest.mark.asyncio
    async def test_run_http_strategy(self, api_key, test_url):
        """Test crawling with HTTP strategy (no JS)."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(test_url, strategy="http")

            assert result.success is True

    @pytest.mark.asyncio
    async def test_invalid_api_key_returns_auth_error(self, test_url):
        """Test that invalid API key returns AuthenticationError."""
        async with AsyncWebCrawler(api_key="sk_test_invalid_key_12345") as crawler:
            with pytest.raises(AuthenticationError):
                await crawler.run(test_url)


class TestCrawlerRunConfig:
    """Tests for CrawlerRunConfig."""

    def test_default_values(self):
        """Test default configuration values."""
        config = CrawlerRunConfig()
        assert config.word_count_threshold == 200
        assert config.exclude_external_links is False
        assert config.screenshot is False

    def test_custom_values(self):
        """Test custom configuration values."""
        config = CrawlerRunConfig(
            word_count_threshold=50,
            screenshot=True,
            process_iframes=True,
        )
        assert config.word_count_threshold == 50
        assert config.screenshot is True
        assert config.process_iframes is True

    def test_dump(self):
        """Test config serialization."""
        config = CrawlerRunConfig(screenshot=True)
        data = config.dump()

        assert isinstance(data, dict)
        assert data["screenshot"] is True


class TestBrowserConfig:
    """Tests for BrowserConfig."""

    def test_default_values(self):
        """Test default browser configuration."""
        config = BrowserConfig()
        assert config.headless is True
        assert config.browser_type == "chromium"
        assert config.viewport_width == 1080

    def test_custom_values(self):
        """Test custom browser configuration."""
        config = BrowserConfig(
            headless=False,
            viewport_width=1920,
            user_agent="Custom UA",
        )
        assert config.headless is False
        assert config.viewport_width == 1920
        assert config.user_agent == "Custom UA"


class TestConfigSanitization:
    """Tests for config sanitization."""

    def test_sanitize_crawler_config_removes_cache_fields(self):
        """Test that cache-related fields are removed."""
        from crawl4ai_cloud import sanitize_crawler_config

        config = CrawlerRunConfig(
            cache_mode="bypass",
            session_id="test-session",
            screenshot=True,
        )

        sanitized = sanitize_crawler_config(config)

        assert "cache_mode" not in sanitized
        assert "session_id" not in sanitized
        assert sanitized.get("screenshot") is True

    def test_sanitize_browser_config_removes_cdp_fields(self):
        """Test that CDP-related fields are removed."""
        from crawl4ai_cloud import sanitize_browser_config

        config = BrowserConfig(
            cdp_url="ws://localhost:9222",
            use_managed_browser=True,
            headless=False,
        )

        sanitized = sanitize_browser_config(config)

        assert "cdp_url" not in sanitized
        assert "use_managed_browser" not in sanitized
        assert sanitized.get("headless") is False


class TestProxyConfig:
    """Tests for proxy configuration."""

    def test_normalize_proxy_string(self):
        """Test proxy string shorthand."""
        from crawl4ai_cloud import normalize_proxy

        result = normalize_proxy("datacenter")
        assert result == {"mode": "datacenter"}

    def test_normalize_proxy_dict(self):
        """Test proxy dict passthrough."""
        from crawl4ai_cloud import normalize_proxy

        proxy = {"mode": "residential", "country": "US"}
        result = normalize_proxy(proxy)
        assert result == proxy

    def test_normalize_proxy_none(self):
        """Test None proxy."""
        from crawl4ai_cloud import normalize_proxy

        result = normalize_proxy(None)
        assert result is None
