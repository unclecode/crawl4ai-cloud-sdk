"""
Context v2 — the four-pillar research pipeline.

This module ships:

- **Pillar builders** — typed factory methods (`Source.google_web(...)`,
  `Strategy.all_items()`, `Shape.raw()`, `Reconciler.noop()`) for the
  pillars that ship today, plus a dict-passthrough escape hatch
  (`Source.custom(type="...", params={...})`) for pillars that ship
  server-side before this SDK adds a typed builder.

- **Result + event types** — `ContextResult` for the run state (with
  lazy `output()` fetch), plus typed `StatusEvent`,
  `PhaseProgressInit`, `PhaseProgressItemUpdate`, and `TerminalEvent`
  for the streaming iterator.

- **Constants** — terminal statuses, phase names.

The crawler methods (`context()`, `context_stream()`,
`refresh_context()`, etc.) live on `AsyncWebCrawler` in `crawler.py`.
This module is the data layer.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, AsyncIterator, Callable, Dict, List, Literal, Optional, Union

# ─── Constants ──────────────────────────────────────────────────────────

# Terminal run statuses — stop polling on any of these.
TERMINAL_STATUSES = frozenset({
    "completed",
    "completed_partial",
    "failed",
    "cancelled",
})

# Non-terminal statuses — keep polling.
ACTIVE_STATUSES = frozenset({"queued", "running"})

# Pipeline phases.
PHASE_PLANNING = "planning"
PHASE_CRAWLING = "crawling"
PHASE_SHAPING = "shaping"
PHASES = (PHASE_PLANNING, PHASE_CRAWLING, PHASE_SHAPING)


# ─── Pillar builders ────────────────────────────────────────────────────


class Source:
    """Builder for Context Source configs.

    Each method returns a plain dict in the shape the API expects:
    ``{"type": "<source_name>", "params": {...}, "auth_ref": Optional[str]}``.
    Use `Source.custom(...)` for pillars that ship server-side before
    this SDK adds a typed builder.
    """

    @staticmethod
    def google_web(
        *,
        backends: Optional[List[str]] = None,
        top_k_per_backend: int = 10,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Google search across multiple SERP backends with RRF merge.

        Args:
            backends: Subset of ``["google", "bing", "duckduckgo", "brave"]``.
                Defaults to ``["google", "bing"]``.
            top_k_per_backend: Per-backend cap before RRF merge (1-50).
            region: 2-letter country code (e.g. ``"us"``, ``"gb"``) to
                bias results.
        """
        params: Dict[str, Any] = {"top_k_per_backend": int(top_k_per_backend)}
        if backends is not None:
            params["backends"] = list(backends)
        if region is not None:
            params["region"] = str(region)
        return {"type": "google_web", "params": params}

    @staticmethod
    def crawl(
        *,
        domain: str,
        max_urls: int = 50,
        max_depth: int = 3,
        score_threshold: Optional[float] = None,
        profile_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Recursive site crawl as the corpus.

        Args:
            domain: Root URL or domain to crawl.
            max_urls: Hard cap on pages indexed (1-500).
            max_depth: Recursion depth from the root.
            score_threshold: BM25 relevance gate against the intent (0-1).
                Pages below are dropped during the query phase.
            profile_id: Saved browser-profile to reuse for auth.
        """
        params: Dict[str, Any] = {
            "domain": str(domain),
            "max_urls": int(max_urls),
            "max_depth": int(max_depth),
        }
        if score_threshold is not None:
            params["score_threshold"] = float(score_threshold)
        if profile_id is not None:
            params["profile_id"] = str(profile_id)
        return {"type": "crawl", "params": params}

    @staticmethod
    def file(
        *,
        file_id: str,
        chunk_size: int = 2000,
        chunk_overlap: int = 200,
    ) -> Dict[str, Any]:
        """User-uploaded file as the corpus.

        Args:
            file_id: Reference to a file uploaded via /v1/files/upload.
            chunk_size: Characters per chunk; each chunk becomes one item.
            chunk_overlap: Character overlap between adjacent chunks.
        """
        return {
            "type": "file",
            "params": {
                "file_id": str(file_id),
                "chunk_size": int(chunk_size),
                "chunk_overlap": int(chunk_overlap),
            },
        }

    @staticmethod
    def custom(
        *,
        type: str,
        params: Optional[Dict[str, Any]] = None,
        auth_ref: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Escape hatch for Sources that exist server-side but don't yet
        have a typed builder in this SDK (e.g. ``hackernews``, ``github``,
        ``rss`` once they ship).

        The dict is forwarded to the API as-is. Discover available Sources
        via ``crawler.context_catalog()``.
        """
        out: Dict[str, Any] = {"type": str(type), "params": dict(params or {})}
        if auth_ref is not None:
            out["auth_ref"] = str(auth_ref)
        return out


class Strategy:
    """Builder for Context Strategy configs."""

    @staticmethod
    def all_items() -> Dict[str, Any]:
        """Passthrough — every candidate item is kept up to
        ``constraints.max_items``. The default."""
        return {"type": "all_items", "params": {}}

    @staticmethod
    def custom(*, type: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Escape hatch for Strategies that ship server-side before this
        SDK adds a typed builder."""
        return {"type": str(type), "params": dict(params or {})}


class Shape:
    """Builder for Context Shape configs."""

    @staticmethod
    def raw() -> Dict[str, Any]:
        """Per-claim items with ``source_url`` provenance. The default."""
        return {"type": "raw", "params": {}}

    @staticmethod
    def custom(*, type: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Escape hatch for Shapes that ship server-side before this SDK
        adds a typed builder."""
        return {"type": str(type), "params": dict(params or {})}


class Reconciler:
    """Builder for Context Reconciler configs."""

    @staticmethod
    def noop() -> Dict[str, Any]:
        """No auto-refresh. Refreshes are user-initiated via
        ``refresh_context()``. The default."""
        return {"type": "noop", "params": {}}

    @staticmethod
    def custom(*, type: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Escape hatch for Reconcilers that ship server-side before
        this SDK adds a typed builder (e.g. ``cron``, ``event``)."""
        return {"type": str(type), "params": dict(params or {})}


# ─── Constraints ────────────────────────────────────────────────────────


@dataclass
class Constraints:
    """Caller-controllable knobs forwarded to the Context pipeline.

    All fields have sensible defaults that match the API. Pass an
    instance to ``crawler.context(constraints=...)`` or pass a plain
    dict — both work."""

    max_items: int = 20
    max_per_source: int = 10
    max_crawl_time_s: float = 120.0
    freshness_days: Optional[int] = None
    language: str = "en"

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {
            "max_items": int(self.max_items),
            "max_per_source": int(self.max_per_source),
            "max_crawl_time_s": float(self.max_crawl_time_s),
            "language": str(self.language),
        }
        if self.freshness_days is not None:
            out["freshness_days"] = int(self.freshness_days)
        return out


# ─── Output types ───────────────────────────────────────────────────────


@dataclass
class ContextItem:
    """One fetched item — typically one URL the Source query phase
    surfaced and the fetch phase materialised. For the ``raw`` Shape,
    each item is the unit of citation: its ``url`` + ``title`` is the
    provenance, and ``content`` / ``snippet`` is what the consumer
    reads."""

    url: Optional[str] = None
    title: Optional[str] = None
    content: Optional[str] = None
    snippet: Optional[str] = None
    source: Optional[str] = None  # The Source that produced this item, e.g. "google_web"
    relevance: float = 0.0
    metadata: Dict[str, Any] = field(default_factory=dict)
    # Optional / Shape-specific fields populated when present:
    id: Optional[str] = None
    fetched_at: Optional[str] = None

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "ContextItem":
        # The raw Shape returns `source_name`; the catalog calls it
        # `source`. Accept either so we don't break when shapes evolve.
        source = data.get("source") or data.get("source_name")
        return cls(
            id=data.get("id"),
            source=source,
            url=data.get("url"),
            title=data.get("title"),
            content=data.get("content"),
            snippet=data.get("snippet"),
            relevance=float(data.get("relevance") or 0.0),
            metadata=data.get("metadata") or {},
            fetched_at=data.get("fetched_at"),
        )


@dataclass
class ContextOutput:
    """The shaped output.

    For the ``raw`` Shape (today), ``items`` carries the fetched
    URLs with content + snippet + source + provenance metadata; each
    item is the citation unit. For future Shapes (``markdown_digest``,
    ``tabular``, ``knowledge_graph``) the top-level structure may
    differ — ``raw`` is preserved unmodified at ``.raw`` so consumers
    can drop down to it when the typed surface doesn't carry a needed
    field.
    """

    shape: str
    items: List[ContextItem] = field(default_factory=list)
    partial: bool = False
    raw: Dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "ContextOutput":
        # The wire shape today is {"type": "raw", "data": {"items": [...]}}.
        # `type` may evolve to `shape`, and `data` may be flattened —
        # accept both for forward compat.
        shape = data.get("shape") or data.get("type") or "raw"
        payload = data.get("data") if isinstance(data.get("data"), dict) else data
        items_data = payload.get("items") if isinstance(payload, dict) else None
        return cls(
            shape=str(shape),
            items=[ContextItem.from_api(i) for i in (items_data or [])],
            partial=bool(data.get("partial", False)),
            raw=data,
        )


# ─── Result type — the run state with lazy output ───────────────────────


@dataclass
class ContextResult:
    """A Context run's state.

    ``output`` is lazy — call ``await result.output()`` to fetch the
    shaped output. The first call hits the API; subsequent calls return
    the cached value.

    Refresh / cancel / diff / rollback / list-versions can also be
    called via the matching crawler methods on the same run_id.
    """

    run_id: str
    status: str
    version: int
    phase: Optional[str] = None
    generator_id: Optional[str] = None
    intent: Optional[str] = None
    constraints: Dict[str, Any] = field(default_factory=dict)
    stats: Dict[str, Any] = field(default_factory=dict)
    error_message: Optional[str] = None
    submitted_at: Optional[str] = None
    completed_at: Optional[str] = None

    # Set by the crawler when the result is built so output() can lazy-fetch.
    _crawler: Optional[Any] = field(default=None, repr=False, compare=False)
    _output: Optional[ContextOutput] = field(default=None, repr=False, compare=False)

    @classmethod
    def from_api(cls, data: Dict[str, Any], crawler: Any = None) -> "ContextResult":
        # The GET /{run_id} response returns the row with `id` as the
        # primary key; the POST /v1/context submit response returns
        # `run_id`. Handle both so this builder works on either.
        run_id = data.get("run_id") or data.get("id") or ""

        # Stats — newer rows surface flat *_ms fields; older API shape used
        # a nested `stats` dict. Fold both into one dict for callers.
        stats: Dict[str, Any] = dict(data.get("stats") or {})
        for k in (
            "planning_ms", "crawling_ms", "shaping_ms", "total_ms",
            "urls_crawled", "urls_failed", "output_size_bytes",
        ):
            if k in data and data[k] is not None:
                stats[k] = data[k]

        return cls(
            run_id=str(run_id),
            status=data.get("status", ""),
            version=int(data.get("version") or 1),
            phase=data.get("phase"),
            generator_id=data.get("generator_id"),
            intent=data.get("intent"),
            constraints=data.get("constraints") or {},
            stats=stats,
            error_message=data.get("error_message"),
            submitted_at=data.get("submitted_at") or data.get("created_at"),
            completed_at=data.get("completed_at"),
            _crawler=crawler,
        )

    @property
    def is_terminal(self) -> bool:
        return self.status in TERMINAL_STATUSES

    @property
    def is_success(self) -> bool:
        return self.status in ("completed", "completed_partial")

    async def output(self) -> ContextOutput:
        """Fetch the shaped output for this run. Cached after the first
        call; safe to call multiple times."""
        if self._output is not None:
            return self._output
        if self._crawler is None:
            raise RuntimeError(
                "ContextResult was built without a crawler reference; "
                "cannot fetch output. Use crawler.get_context_output(run_id)."
            )
        self._output = await self._crawler.get_context_output(self.run_id)
        return self._output


# ─── Streaming event types ──────────────────────────────────────────────


@dataclass
class StatusEvent:
    """Fired on every phase transition."""
    type: str = "status"
    status: str = ""
    phase: Optional[str] = None
    version: int = 1
    planning_ms: int = 0
    crawling_ms: int = 0
    shaping_ms: int = 0
    ts: Optional[str] = None

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "StatusEvent":
        return cls(
            status=data.get("status", ""),
            phase=data.get("phase"),
            version=int(data.get("version") or 1),
            planning_ms=int(data.get("planning_ms") or 0),
            crawling_ms=int(data.get("crawling_ms") or 0),
            shaping_ms=int(data.get("shaping_ms") or 0),
            ts=data.get("ts"),
        )


@dataclass
class PhaseProgressInit:
    """Fired once at the start of the crawling phase with the full plan
    of items to fetch."""
    type: str = "phase_progress"
    kind: str = "init"
    phase: str = "fetch"
    total: int = 0
    items: List[Dict[str, Any]] = field(default_factory=list)

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "PhaseProgressInit":
        return cls(
            phase=data.get("phase", "fetch"),
            total=int(data.get("total") or 0),
            items=list(data.get("items") or []),
        )


@dataclass
class PhaseProgressItemUpdate:
    """Fired once per item as it completes during the fetch phase."""
    type: str = "phase_progress"
    kind: str = "item_update"
    phase: str = "fetch"
    id: str = ""
    status: str = ""
    ms: int = 0
    size: Optional[int] = None
    reason: Optional[str] = None

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "PhaseProgressItemUpdate":
        return cls(
            phase=data.get("phase", "fetch"),
            id=data.get("id", ""),
            status=data.get("status", ""),
            ms=int(data.get("ms") or 0),
            size=data.get("size"),
            reason=data.get("reason"),
        )


@dataclass
class TerminalEvent:
    """Fired exactly once when the run reaches a terminal status."""
    type: str = "terminal"
    status: str = ""
    total_ms: int = 0
    urls_crawled: int = 0
    urls_failed: int = 0
    output_s3_key: Optional[str] = None
    error_message: Optional[str] = None

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "TerminalEvent":
        return cls(
            status=data.get("status", ""),
            total_ms=int(data.get("total_ms") or 0),
            urls_crawled=int(data.get("urls_crawled") or 0),
            urls_failed=int(data.get("urls_failed") or 0),
            output_s3_key=data.get("output_s3_key"),
            error_message=data.get("error_message"),
        )


ContextEvent = Union[
    StatusEvent,
    PhaseProgressInit,
    PhaseProgressItemUpdate,
    TerminalEvent,
]


def _parse_event(event_type: str, data: Dict[str, Any]) -> Optional[ContextEvent]:
    """Translate raw SSE event ``{event, data}`` into a typed event.

    Returns None for unknown event types (forward-compatible)."""
    if data.get("type") == "status" or event_type == "status":
        return StatusEvent.from_api(data)
    if data.get("type") == "terminal" or event_type == "terminal":
        return TerminalEvent.from_api(data)
    if data.get("type") == "phase_progress" or event_type == "phase_progress":
        kind = data.get("kind") or "init"
        if kind == "init":
            return PhaseProgressInit.from_api(data)
        if kind == "item_update":
            return PhaseProgressItemUpdate.from_api(data)
    return None


# ─── Diff / Version helpers ─────────────────────────────────────────────


@dataclass
class ContextVersion:
    version: int
    status: str
    submitted_at: Optional[str] = None
    completed_at: Optional[str] = None
    urls_crawled: int = 0
    triggered_by: str = "user"
    output_s3_key: Optional[str] = None

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "ContextVersion":
        return cls(
            version=int(data.get("version") or 1),
            status=data.get("status", ""),
            submitted_at=data.get("submitted_at"),
            completed_at=data.get("completed_at"),
            urls_crawled=int(data.get("urls_crawled") or 0),
            triggered_by=data.get("triggered_by") or "user",
            output_s3_key=data.get("output_s3_key"),
        )


@dataclass
class ContextDiff:
    """Diff between two Context versions.

    Today the diff is item-level (matched by stable URL). Future
    versions may diff at the claim level once a Shape that emits
    discrete claims (e.g. ``markdown_digest``) is wired through. Until
    then `added`/`removed`/`unchanged` are lists of ``ContextItem``."""

    added: List[ContextItem] = field(default_factory=list)
    removed: List[ContextItem] = field(default_factory=list)
    unchanged: List[ContextItem] = field(default_factory=list)
    sources_added: List[str] = field(default_factory=list)
    sources_removed: List[str] = field(default_factory=list)
    raw: Dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "ContextDiff":
        return cls(
            added=[ContextItem.from_api(c) for c in (data.get("added") or [])],
            removed=[ContextItem.from_api(c) for c in (data.get("removed") or [])],
            unchanged=[ContextItem.from_api(c) for c in (data.get("unchanged") or [])],
            sources_added=list(data.get("sources_added") or []),
            sources_removed=list(data.get("sources_removed") or []),
            raw=data,
        )


@dataclass
class CatalogEntry:
    """One entry from /v1/context/{sources,strategies,shapes,reconcilers}."""

    name: str
    display_name: str
    summary: str
    help_md: str = ""
    params_schema: Dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "CatalogEntry":
        return cls(
            name=data.get("name", ""),
            display_name=data.get("display_name") or data.get("name", ""),
            summary=data.get("summary", ""),
            help_md=data.get("help_md", ""),
            params_schema=data.get("params_schema") or data.get("query_params_schema") or {},
        )


@dataclass
class ContextCatalog:
    """Whole-catalog response: every pillar and its catalog entries."""

    sources: List[CatalogEntry] = field(default_factory=list)
    strategies: List[CatalogEntry] = field(default_factory=list)
    shapes: List[CatalogEntry] = field(default_factory=list)
    reconcilers: List[CatalogEntry] = field(default_factory=list)


__all__ = [
    # Constants
    "TERMINAL_STATUSES",
    "ACTIVE_STATUSES",
    "PHASE_PLANNING",
    "PHASE_CRAWLING",
    "PHASE_SHAPING",
    "PHASES",
    # Pillar builders
    "Source",
    "Strategy",
    "Shape",
    "Reconciler",
    # Constraints
    "Constraints",
    # Output types
    "ContextItem",
    "ContextOutput",
    # Result
    "ContextResult",
    # Events
    "StatusEvent",
    "PhaseProgressInit",
    "PhaseProgressItemUpdate",
    "TerminalEvent",
    "ContextEvent",
    # Versions / diff / catalog
    "ContextVersion",
    "ContextDiff",
    "CatalogEntry",
    "ContextCatalog",
]
