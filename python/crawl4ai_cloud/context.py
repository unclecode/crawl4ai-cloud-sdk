"""
Context v2 — the four-pillar research pipeline.

Public surface:

* **Pillar builders** — typed factory methods for everything currently
  registered server-side:

      Source.google_web / google_drive / gmail / crawl / file / custom
      Strategy.all_items / llm_rerank / custom
      Synthesizer.raw / markdown / llm / custom
      Reconciler.noop / custom

  Each returns a plain dict in the API wire shape
  (``{"type": "<name>", "params": {...}}``). ``Shape`` is a deprecated
  alias for ``Synthesizer``.

* **Result + event types** — :class:`ContextResult` (lazy
  ``.output()``), :class:`ContextOutput` with shape-specific sugar
  (``.markdown`` for markdown-single, ``.files`` for markdown-multi,
  ``.data`` for llm), plus typed :class:`StatusEvent` /
  :class:`PhaseProgressInit` / :class:`PhaseProgressItemUpdate` /
  :class:`TerminalEvent` for the streaming iterator.

The crawler methods (``context()``, ``context_stream()``,
``refresh_context()`` …) live on :class:`AsyncWebCrawler` in
``crawler.py``. This module is the pure data layer.
"""
from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any, AsyncIterator, Callable, Dict, List, Literal, Optional, Union

# ─── Constants ──────────────────────────────────────────────────────────

TERMINAL_STATUSES = frozenset({
    "completed",
    "completed_partial",
    "failed",
    "cancelled",
})

ACTIVE_STATUSES = frozenset({"queued", "running"})

PHASE_PLANNING = "planning"
PHASE_CRAWLING = "crawling"
PHASE_SHAPING = "shaping"
PHASES = (PHASE_PLANNING, PHASE_CRAWLING, PHASE_SHAPING)


# ─── helpers ────────────────────────────────────────────────────────────


def _serialize(value: Any) -> str:
    """Convert dict / list args to JSON strings for params that the
    server stores as text (output_schema, output_example, …). Strings
    pass through unchanged."""
    if value is None:
        return ""
    if isinstance(value, str):
        return value
    return json.dumps(value)


# ─── Pillar builders ────────────────────────────────────────────────────


class Source:
    """Builder for Context Source configs.

    Each method returns ``{"type": "<source>", "params": {...}}``.
    """

    @staticmethod
    def google_web(
        *,
        backends: Optional[List[str]] = None,
        top_k_per_backend: int = 10,
        region: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Google search across multiple SERP backends with RRF merge."""
        params: Dict[str, Any] = {"top_k_per_backend": int(top_k_per_backend)}
        if backends is not None:
            params["backends"] = list(backends)
        if region is not None:
            params["region"] = str(region)
        return {"type": "google_web", "params": params}

    @staticmethod
    def google_drive(
        *,
        mode: str = "search",
        folder_id: str = "",
        auth_ref: Optional[str] = None,
    ) -> Dict[str, Any]:
        """User's Google Drive.

        ``mode``: ``"search"`` (intent-driven, optional folder scope) or
        ``"folder"`` (list one folder). ``folder_id`` is required when
        ``mode="folder"``. Requires a Drive connection — pass the linked
        account's ``auth_ref`` (or omit to let the server use the user's
        default link).
        """
        if mode not in ("search", "folder"):
            raise ValueError(f"mode must be 'search' or 'folder', got {mode!r}")
        if mode == "folder" and not folder_id:
            raise ValueError("folder_id is required when mode='folder'")
        out: Dict[str, Any] = {
            "type": "google_drive",
            "params": {"mode": mode, "folder_id": str(folder_id)},
        }
        if auth_ref is not None:
            out["auth_ref"] = str(auth_ref)
        return out

    @staticmethod
    def gmail(
        *,
        mode: str = "search",
        label_id: str = "",
        after: str = "",
        before: str = "",
        include_spam_trash: bool = False,
        auth_ref: Optional[str] = None,
    ) -> Dict[str, Any]:
        """User's Gmail.

        ``mode``: ``"search"`` (intent-driven Gmail query) or
        ``"label"`` (list threads in one label). Dates are
        ``YYYY/MM/DD``. Requires a Gmail connection.
        """
        if mode not in ("search", "label"):
            raise ValueError(f"mode must be 'search' or 'label', got {mode!r}")
        if mode == "label" and not label_id:
            raise ValueError("label_id is required when mode='label'")
        out: Dict[str, Any] = {
            "type": "gmail",
            "params": {
                "mode": mode,
                "label_id": str(label_id),
                "after": str(after),
                "before": str(before),
                "include_spam_trash": bool(include_spam_trash),
            },
        }
        if auth_ref is not None:
            out["auth_ref"] = str(auth_ref)
        return out

    @staticmethod
    def crawl(
        *,
        domain: str,
        max_urls: int = 50,
        max_depth: int = 3,
        score_threshold: Optional[float] = None,
        profile_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        """Recursive site crawl as the corpus."""
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
        """User-uploaded file as the corpus."""
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
        """Escape hatch for Sources without a typed builder yet.
        Discover available Sources via ``crawler.context_catalog()``."""
        out: Dict[str, Any] = {"type": str(type), "params": dict(params or {})}
        if auth_ref is not None:
            out["auth_ref"] = str(auth_ref)
        return out


class Strategy:
    """Builder for Context Strategy configs."""

    @staticmethod
    def all_items() -> Dict[str, Any]:
        """Passthrough — every candidate kept up to
        ``constraints.max_items``. The default."""
        return {"type": "all_items", "params": {}}

    @staticmethod
    def llm_rerank(
        *,
        top_n: int = 0,
        instruction: str = "",
        model: str = "anthropic/claude-haiku-4-5",
        score_threshold: float = 0.0,
        batch_size: int = 20,
        max_concurrency: int = 4,
        content_aware: bool = False,
        content_chars: int = 4000,
    ) -> Dict[str, Any]:
        """Score every candidate against the intent with an LLM, keep
        the top N.

        ``top_n=0`` means use the request's ``max_items``. ``content_aware``
        scores on the item's body (for ``owns_content`` Sources like
        Drive / Gmail / HN) instead of just title + snippet.
        """
        return {"type": "llm_rerank", "params": {
            "top_n": int(top_n),
            "instruction": str(instruction),
            "model": str(model),
            "score_threshold": float(score_threshold),
            "batch_size": int(batch_size),
            "max_concurrency": int(max_concurrency),
            "content_aware": bool(content_aware),
            "content_chars": int(content_chars),
        }}

    @staticmethod
    def custom(*, type: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Escape hatch for Strategies without a typed builder yet."""
        return {"type": str(type), "params": dict(params or {})}


class Synthesizer:
    """Builder for Context Synthesizer configs.

    The pillar was previously named "Shape". :class:`Shape` remains as
    a deprecated alias for one release — new code should use
    :class:`Synthesizer`.
    """

    @staticmethod
    def raw() -> Dict[str, Any]:
        """Per-item citations with ``url`` provenance. The default."""
        return {"type": "raw", "params": {}}

    @staticmethod
    def markdown(
        *,
        mode: str = "single",
        instruction: str = "",
        model: str = "anthropic/claude-haiku-4-5",
        batch_size: int = 5,
        max_concurrency: int = 4,
        include_metadata: bool = True,
        max_chars_per_item: int = 20000,
    ) -> Dict[str, Any]:
        """Render the materialised plan as markdown.

        ``mode="single"`` — one joined .md body (the default).
        ``mode="multi"``  — one .md per item (downloadable as zip).

        When ``instruction`` is non-empty, each item is rewritten by the
        LLM before the markdown is built.
        """
        if mode not in ("single", "multi"):
            raise ValueError(f"mode must be 'single' or 'multi', got {mode!r}")
        return {"type": "markdown", "params": {
            "mode": mode,
            "instruction": str(instruction),
            "model": str(model),
            "batch_size": int(batch_size),
            "max_concurrency": int(max_concurrency),
            "include_metadata": bool(include_metadata),
            "max_chars_per_item": int(max_chars_per_item),
        }}

    @staticmethod
    def llm(
        *,
        instruction: str,
        schema: Optional[Union[str, Dict[str, Any]]] = None,
        example: Optional[Union[str, Dict[str, Any], List[Any]]] = None,
        description: Optional[str] = None,
        model: str = "anthropic/claude-haiku-4-5",
        temperature: float = 0.0,
        max_corpus_chars: int = 40000,
        auto_repair: bool = True,
    ) -> Dict[str, Any]:
        """One LLM call that fills a caller-defined JSON shape.

        Pass exactly one of:
          * ``schema``      — full JSON Schema (used as-is)
          * ``example``     — concrete JSON example (walked into schema,
                              deterministic)
          * ``description`` — plain-English shape description (one LLM
                              call drafts the schema)

        Dict / list args to ``schema`` / ``example`` are JSON-serialised.
        """
        if not instruction or not instruction.strip():
            raise ValueError("instruction is required for Synthesizer.llm")
        # `is not None` rather than truthiness — caller passing an explicit
        # empty value is still a value, and we want to surface that ambiguity
        # rather than silently let "exactly one of three" elide an empty input.
        n_set = sum(1 for v in (schema, example, description) if v is not None)
        if n_set != 1:
            raise ValueError(
                "Pass exactly one of: schema, example, description"
            )
        return {"type": "llm", "params": {
            "instruction": str(instruction),
            "output_schema":      _serialize(schema),
            "output_example":     _serialize(example),
            "output_description": str(description) if description else "",
            "model": str(model),
            "temperature": float(temperature),
            "max_corpus_chars": int(max_corpus_chars),
            "auto_repair": bool(auto_repair),
        }}

    @staticmethod
    def custom(*, type: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Escape hatch for Synthesizers without a typed builder yet."""
        return {"type": str(type), "params": dict(params or {})}


# Back-compat — one release.
Shape = Synthesizer


class Reconciler:
    """Builder for Context Reconciler configs."""

    @staticmethod
    def noop() -> Dict[str, Any]:
        """No auto-refresh. Refreshes are user-initiated via
        ``refresh_context()``. The default."""
        return {"type": "noop", "params": {}}

    @staticmethod
    def custom(*, type: str, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Escape hatch for Reconcilers without a typed builder yet
        (e.g. ``cron``, ``webhook`` once they ship)."""
        return {"type": str(type), "params": dict(params or {})}


# ─── Constraints ────────────────────────────────────────────────────────


@dataclass
class Constraints:
    """Caller-controllable knobs forwarded to the Context pipeline.

    Pass an instance or a plain dict — both work."""

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
    """One fetched item — typically one URL surfaced by a Source's
    query phase and materialised by its fetch phase. For ``raw`` it's
    also the citation unit (``url`` + ``title`` = provenance,
    ``content`` / ``snippet`` = body)."""

    url: Optional[str] = None
    title: Optional[str] = None
    content: Optional[str] = None
    snippet: Optional[str] = None
    source: Optional[str] = None
    relevance: float = 0.0
    metadata: Dict[str, Any] = field(default_factory=dict)
    id: Optional[str] = None
    fetched_at: Optional[str] = None

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "ContextItem":
        return cls(
            id=data.get("id"),
            source=data.get("source") or data.get("source_name"),
            url=data.get("url"),
            title=data.get("title"),
            content=data.get("content"),
            snippet=data.get("snippet"),
            relevance=float(data.get("relevance") or 0.0),
            metadata=data.get("metadata") or {},
            fetched_at=data.get("fetched_at"),
        )


@dataclass
class MarkdownFile:
    """One per-item markdown file emitted by ``Synthesizer.markdown(mode='multi')``."""
    filename: str
    markdown: str


@dataclass
class ContextOutput:
    """The synthesized output.

    Shape-specific accessors:

      * ``raw``      → ``.items`` is the canonical citation list
      * ``markdown`` → ``.markdown`` (str) for single mode, ``.files``
                       (list[MarkdownFile]) for multi mode
      * ``llm``      → ``.data`` (dict), plus ``.resolved_schema``,
                       ``.notes``, ``.partial``

    For every shape, ``.raw_payload`` is the full wire envelope and
    ``.items`` is the list of source items the synthesizer rendered.
    """

    shape: str
    items: List[ContextItem] = field(default_factory=list)
    partial: bool = False
    raw_payload: Dict[str, Any] = field(default_factory=dict)

    # Markdown synthesizer
    markdown: Optional[str] = None
    files: Optional[List[MarkdownFile]] = None

    # LLM synthesizer
    data: Optional[Any] = None
    resolved_schema: Optional[Dict[str, Any]] = None
    notes: List[str] = field(default_factory=list)
    partial_data: Optional[Any] = None

    # Back-compat — used to be ``raw``. Kept as an alias for one release.
    @property
    def raw(self) -> Dict[str, Any]:
        return self.raw_payload

    @classmethod
    def from_api(cls, data: Dict[str, Any]) -> "ContextOutput":
        """Wire shape today is ``{"type": <syn>, "data": {...}}``;
        forward-compat accepts ``"shape"`` and a flat payload too.
        """
        shape = data.get("shape") or data.get("type") or "raw"
        payload = data.get("data") if isinstance(data.get("data"), dict) else data
        if not isinstance(payload, dict):
            payload = {}

        items_data = payload.get("items") or []
        items = [ContextItem.from_api(i) for i in items_data]

        out = cls(
            shape=str(shape),
            items=items,
            partial=bool(data.get("partial", False)),
            raw_payload=data,
        )

        if shape == "markdown":
            mode = payload.get("mode") or "single"
            if mode == "multi":
                files_raw = payload.get("files") or []
                out.files = [
                    MarkdownFile(
                        filename=str(f.get("filename") or ""),
                        markdown=str(f.get("markdown") or ""),
                    )
                    for f in files_raw if isinstance(f, dict)
                ]
            else:
                out.markdown = payload.get("markdown")
        elif shape == "llm":
            out.data = payload.get("data")
            out.resolved_schema = payload.get("resolved_schema") or {}
            out.notes = list(payload.get("notes") or [])
            out.partial_data = payload.get("partial")

        return out


# ─── Result type — the run state with lazy output ───────────────────────


@dataclass
class ContextResult:
    """A Context run's state.

    ``output`` is lazy — call ``await result.output()`` to fetch the
    synthesized output. Cached after first call.
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

    _crawler: Optional[Any] = field(default=None, repr=False, compare=False)
    _output: Optional[ContextOutput] = field(default=None, repr=False, compare=False)

    @classmethod
    def from_api(cls, data: Dict[str, Any], crawler: Any = None) -> "ContextResult":
        run_id = data.get("run_id") or data.get("id") or ""
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
        if self._output is not None:
            return self._output
        if self._crawler is None:
            raise RuntimeError(
                "ContextResult was built without a crawler reference; "
                "use crawler.get_context_output(run_id)."
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
    """Fired once at the start of the crawling phase with the plan."""
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
    """Item-level diff between two Context versions."""
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
    """One entry from /v1/context/{sources,strategies,synthesizers,reconcilers}."""
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
    synthesizers: List[CatalogEntry] = field(default_factory=list)
    reconcilers: List[CatalogEntry] = field(default_factory=list)

    # Deprecated alias.
    @property
    def shapes(self) -> List[CatalogEntry]:
        return self.synthesizers


__all__ = [
    "TERMINAL_STATUSES",
    "ACTIVE_STATUSES",
    "PHASE_PLANNING",
    "PHASE_CRAWLING",
    "PHASE_SHAPING",
    "PHASES",
    "Source",
    "Strategy",
    "Shape",
    "Synthesizer",
    "Reconciler",
    "Constraints",
    "ContextItem",
    "ContextOutput",
    "MarkdownFile",
    "ContextResult",
    "StatusEvent",
    "PhaseProgressInit",
    "PhaseProgressItemUpdate",
    "TerminalEvent",
    "ContextEvent",
    "ContextVersion",
    "ContextDiff",
    "CatalogEntry",
    "ContextCatalog",
]
