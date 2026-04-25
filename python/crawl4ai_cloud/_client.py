"""Internal HTTP client for Crawl4AI Cloud SDK."""
import asyncio
import json as json_module
import os
from typing import AsyncIterator, Optional, Dict, Any, Tuple

import httpx

from .errors import (
    CloudError,
    AuthenticationError,
    RateLimitError,
    QuotaExceededError,
    NotFoundError,
    ValidationError,
    ServerError,
    TimeoutError,
)

__version__ = "0.6.1"

DEFAULT_BASE_URL = "https://api.crawl4ai.com"
DEFAULT_TIMEOUT = 120.0
DEFAULT_MAX_RETRIES = 3


class HTTPClient:
    """Internal async HTTP client with retries and error mapping."""

    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = DEFAULT_BASE_URL,
        timeout: float = DEFAULT_TIMEOUT,
        max_retries: int = DEFAULT_MAX_RETRIES,
    ):
        """
        Initialize the HTTP client.

        Args:
            api_key: Your Crawl4AI API key (sk_live_* or sk_test_*).
                     If not provided, reads from CRAWL4AI_API_KEY env var.
            base_url: API base URL (default: https://api.crawl4ai.com)
            timeout: Request timeout in seconds (default: 120)
            max_retries: Max retry attempts for transient errors (default: 3)

        Raises:
            ValueError: If API key is missing or has invalid format
        """
        self._api_key = api_key or os.getenv("CRAWL4AI_API_KEY")

        if not self._api_key:
            raise ValueError(
                "API key is required. Provide it as an argument or set "
                "the CRAWL4AI_API_KEY environment variable."
            )

        if not self._api_key.startswith(("sk_live_", "sk_test_")):
            raise ValueError(
                "Invalid API key format. Expected sk_live_* or sk_test_*"
            )

        self._base_url = base_url.rstrip("/")
        self._timeout = timeout
        self._max_retries = max_retries
        self._client: Optional[httpx.AsyncClient] = None

    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create the HTTP client."""
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(
                base_url=self._base_url,
                headers={
                    "X-API-Key": self._api_key,
                    "Content-Type": "application/json",
                    "User-Agent": f"crawl4ai-cloud/{__version__}",
                },
                timeout=httpx.Timeout(self._timeout),
            )
        return self._client

    async def request(
        self,
        method: str,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        json: Optional[Dict[str, Any]] = None,
        timeout: Optional[float] = None,
    ) -> Dict[str, Any]:
        """
        Make HTTP request with error handling and retries.

        Args:
            method: HTTP method (GET, POST, DELETE, etc.)
            path: API endpoint path
            params: Query parameters
            json: JSON body
            timeout: Request timeout override

        Returns:
            Parsed JSON response

        Raises:
            AuthenticationError: 401 - Invalid API key
            NotFoundError: 404 - Resource not found
            RateLimitError: 429 - Rate limit exceeded
            QuotaExceededError: 429 - Quota exceeded
            ValidationError: 400 - Invalid request
            TimeoutError: 504 or client timeout
            ServerError: 500/503 - Server error
            CloudError: Other errors
        """
        client = await self._get_client()

        for attempt in range(self._max_retries):
            try:
                response = await client.request(
                    method,
                    path,
                    params=params,
                    json=json,
                    timeout=timeout or self._timeout,
                )

                # Success
                if response.status_code < 400:
                    if response.content:
                        return response.json()
                    return {}

                # Parse error response
                try:
                    error_data = response.json()
                    detail = error_data.get("detail", str(error_data))
                except Exception:
                    detail = response.text or f"HTTP {response.status_code}"
                    error_data = {}

                headers = {k.lower(): v for k, v in response.headers.items()}

                # Map status codes to exceptions
                if response.status_code == 401:
                    raise AuthenticationError(detail, 401, error_data, headers)
                elif response.status_code == 404:
                    raise NotFoundError(detail, 404, error_data, headers)
                elif response.status_code == 429:
                    if "rate limit" in detail.lower():
                        raise RateLimitError(detail, 429, error_data, headers)
                    else:
                        raise QuotaExceededError(detail, 429, error_data, headers)
                elif response.status_code == 400:
                    raise ValidationError(detail, 400, error_data, headers)
                elif response.status_code == 504:
                    raise TimeoutError(detail, 504, error_data, headers)
                elif response.status_code >= 500:
                    if attempt < self._max_retries - 1:
                        await asyncio.sleep(2 ** attempt)
                        continue
                    raise ServerError(
                        detail, response.status_code, error_data, headers
                    )
                else:
                    raise CloudError(
                        detail, response.status_code, error_data, headers
                    )

            except httpx.TimeoutException as e:
                if attempt < self._max_retries - 1:
                    await asyncio.sleep(2 ** attempt)
                    continue
                raise TimeoutError(f"Request timed out: {e}")

            except httpx.RequestError as e:
                if attempt < self._max_retries - 1:
                    await asyncio.sleep(2 ** attempt)
                    continue
                raise CloudError(f"Request failed: {e}")

        raise CloudError("Max retries exceeded")

    async def stream_sse(
        self,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        timeout: Optional[float] = None,
    ) -> AsyncIterator[Tuple[str, Dict[str, Any]]]:
        """Open an SSE connection and yield (event_type, parsed_data) pairs.

        Skips heartbeat comments (lines starting with ':'). Stops when the
        server closes the connection or sends an event with `event: complete`.
        """
        client = await self._get_client()
        # SSE needs an open-ended timeout — let the server decide when to close.
        request_timeout = httpx.Timeout(
            connect=10.0, read=None, write=10.0, pool=10.0,
        ) if timeout is None else httpx.Timeout(timeout)

        async with client.stream(
            "GET", path, params=params, timeout=request_timeout,
            headers={"Accept": "text/event-stream"},
        ) as response:
            if response.status_code >= 400:
                body = await response.aread()
                detail = body.decode("utf-8", errors="replace") or f"HTTP {response.status_code}"
                if response.status_code == 401:
                    raise AuthenticationError(detail, 401, {}, {})
                if response.status_code == 404:
                    raise NotFoundError(detail, 404, {}, {})
                raise CloudError(detail, response.status_code, {}, {})

            event_name: Optional[str] = None
            data_buf: list[str] = []
            async for line in response.aiter_lines():
                if line == "":
                    # End of one event — dispatch
                    if data_buf:
                        raw = "\n".join(data_buf)
                        try:
                            parsed = json_module.loads(raw)
                        except Exception:
                            parsed = {"raw": raw}
                        yield (event_name or "message", parsed)
                    event_name = None
                    data_buf = []
                    continue
                if line.startswith(":"):
                    # Heartbeat comment
                    continue
                if line.startswith("event:"):
                    event_name = line[6:].strip()
                elif line.startswith("data:"):
                    data_buf.append(line[5:].lstrip())
                # Ignore other SSE field types (id, retry) — we don't reconnect.

    async def close(self):
        """Close the HTTP client."""
        if self._client and not self._client.is_closed:
            await self._client.aclose()
            self._client = None

    async def __aenter__(self) -> "HTTPClient":
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
