"""Exception classes for Crawl4AI Cloud SDK."""
from typing import Optional, Dict, Any


class CloudError(Exception):
    """Base exception for Crawl4AI Cloud SDK."""

    def __init__(
        self,
        message: str,
        status_code: Optional[int] = None,
        response: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None,
    ):
        super().__init__(message)
        self.message = message
        self.status_code = status_code
        self.response = response or {}
        self.headers = headers or {}

    def __str__(self) -> str:
        if self.status_code:
            return f"[{self.status_code}] {self.message}"
        return self.message


class AuthenticationError(CloudError):
    """401 - Invalid or missing API key."""
    pass


class RateLimitError(CloudError):
    """429 - Rate limit exceeded."""

    @property
    def retry_after(self) -> int:
        """Seconds until rate limit resets."""
        try:
            return int(self.headers.get("x-ratelimit-reset", 0))
        except (ValueError, TypeError):
            return 0

    @property
    def limit(self) -> int:
        """Rate limit per minute."""
        try:
            return int(self.headers.get("x-ratelimit-limit", 0))
        except (ValueError, TypeError):
            return 0

    @property
    def remaining(self) -> int:
        """Remaining requests in current window."""
        try:
            return int(self.headers.get("x-ratelimit-remaining", 0))
        except (ValueError, TypeError):
            return 0


class QuotaExceededError(CloudError):
    """429 - Daily/concurrent/storage quota exceeded."""

    @property
    def quota_type(self) -> str:
        """Returns 'daily', 'concurrent', or 'storage'."""
        msg = self.message.lower()
        if "storage" in msg:
            return "storage"
        elif "concurrent" in msg:
            return "concurrent"
        return "daily"


class NotFoundError(CloudError):
    """404 - Resource not found (job, session, etc.)."""
    pass


class ValidationError(CloudError):
    """400 - Invalid request parameters."""
    pass


class TimeoutError(CloudError):
    """504 or client timeout - Request/crawl timed out."""
    pass


class ServerError(CloudError):
    """500/503 - Server or worker error."""
    pass
