"""Pytest fixtures for crawl4ai-cloud tests."""
import os
import pytest

# Test API key
TEST_API_KEY = os.getenv(
    "CRAWL4AI_API_KEY",
    "sk_live_cM9VqS3ostZxB0FcjBZScbVnbk_Zni707mxU-uZWJKQ"
)

TEST_URL = "https://example.com"


@pytest.fixture
def api_key():
    return TEST_API_KEY


@pytest.fixture
def test_url():
    return TEST_URL
