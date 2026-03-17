"""E2E tests for file download detection via HTTP strategy.

Tests that the SDK correctly receives and exposes downloaded_files
from the API when crawling non-HTML URLs with strategy="http".
"""
import pytest

from crawl4ai_cloud import AsyncWebCrawler, CrawlResult


class TestFileDownloadE2E:
    """End-to-end tests hitting the real API."""

    @pytest.mark.asyncio
    async def test_csv_download_has_presigned_url(self, api_key):
        """CSV file should return downloaded_files with S3 presigned URL."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(
                "https://data.gov.au/data/dataset/043f58e0-a188-4458-b61c-04e5b540aea4/resource/f83cdee9-ebcb-4f24-941b-34bb2f0996cf/download/facilities.csv",
                strategy="http",
                bypass_cache=True,
            )

            assert isinstance(result, CrawlResult)
            assert result.success is True
            assert result.downloaded_files is not None
            assert len(result.downloaded_files) >= 1
            assert result.downloaded_files[0].startswith("https://")
            # CSV is text-based — html also has content (backward compat)
            assert result.html is not None
            assert len(result.html) > 1000

    @pytest.mark.asyncio
    async def test_json_download_has_presigned_url(self, api_key):
        """JSON API response should return downloaded_files."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(
                "https://jsonplaceholder.typicode.com/posts/1",
                strategy="http",
                bypass_cache=True,
            )

            assert result.success is True
            assert result.downloaded_files is not None
            assert len(result.downloaded_files) >= 1
            # JSON is text-based — html has content
            assert result.html is not None
            assert "userId" in result.html

    @pytest.mark.asyncio
    async def test_html_page_no_downloaded_files(self, api_key):
        """Normal HTML page should NOT have downloaded_files."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(
                "https://example.com",
                strategy="http",
                bypass_cache=True,
            )

            assert result.success is True
            assert result.downloaded_files is None
            assert result.html is not None
            assert "Example Domain" in result.html

    @pytest.mark.asyncio
    async def test_binary_download_empty_html(self, api_key):
        """Binary file (octet-stream) should have downloaded_files and empty html."""
        async with AsyncWebCrawler(api_key=api_key) as crawler:
            result = await crawler.run(
                "https://httpbin.org/bytes/1024",
                strategy="http",
                bypass_cache=True,
            )

            assert result.success is True
            assert result.downloaded_files is not None
            assert len(result.downloaded_files) >= 1
            assert result.downloaded_files[0].startswith("https://")


class TestCrawlResultParsing:
    """Unit tests for CrawlResult.from_dict parsing."""

    def test_from_dict_with_downloaded_files(self):
        """CrawlResult.from_dict should parse downloaded_files field."""
        data = {
            "url": "https://example.com/data.csv",
            "success": True,
            "html": "a,b,c\n1,2,3",
            "downloaded_files": [
                "https://s3.example.com/downloads/abc/data.csv?signature=xyz"
            ],
            "status_code": 200,
            "duration_ms": 500,
        }
        result = CrawlResult.from_dict(data)

        assert result.downloaded_files is not None
        assert len(result.downloaded_files) == 1
        assert "data.csv" in result.downloaded_files[0]

    def test_from_dict_without_downloaded_files(self):
        """CrawlResult.from_dict should handle missing downloaded_files."""
        data = {
            "url": "https://example.com",
            "success": True,
            "html": "<html>hello</html>",
            "status_code": 200,
            "duration_ms": 100,
        }
        result = CrawlResult.from_dict(data)

        assert result.downloaded_files is None

    def test_from_dict_null_downloaded_files(self):
        """CrawlResult.from_dict should handle null downloaded_files."""
        data = {
            "url": "https://example.com",
            "success": True,
            "downloaded_files": None,
            "status_code": 200,
            "duration_ms": 100,
        }
        result = CrawlResult.from_dict(data)

        assert result.downloaded_files is None

    def test_from_dict_multiple_downloaded_files(self):
        """CrawlResult.from_dict should handle multiple files."""
        data = {
            "url": "https://example.com/archive",
            "success": True,
            "downloaded_files": [
                "https://s3.example.com/file1.csv",
                "https://s3.example.com/file2.pdf",
            ],
            "status_code": 200,
            "duration_ms": 200,
        }
        result = CrawlResult.from_dict(data)

        assert result.downloaded_files is not None
        assert len(result.downloaded_files) == 2
