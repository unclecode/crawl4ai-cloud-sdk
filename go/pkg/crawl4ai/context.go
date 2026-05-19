package crawl4ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Context v2 — the four-pillar research pipeline.
//
// Ships:
//   - **Pillar builders** — typed constructors (`GoogleWebSource(...)`,
//     `AllItemsStrategy()`, `RawShape()`, `NoopReconciler()`) for pillars
//     that ship today, plus a dict-passthrough escape hatch
//     (`CustomSource(...)` / `CustomStrategy(...)` / `CustomShape(...)` /
//     `CustomReconciler(...)`) for pillars that ship server-side before
//     this SDK adds a typed builder.
//   - **Result + event types** — `ContextResult` for the run state (with
//     lazy `Output()` fetch), plus typed `ContextEvent` discriminated
//     union for the streaming channel.
//   - **Constants** — terminal statuses, phase names.

// ─── Constants ──────────────────────────────────────────────────────────

// ContextTerminalStatuses are the run statuses that mean "stop polling".
var ContextTerminalStatuses = map[string]bool{
	"completed":         true,
	"completed_partial": true,
	"failed":            true,
	"cancelled":         true,
}

// ContextActiveStatuses are non-terminal — keep polling.
var ContextActiveStatuses = map[string]bool{
	"queued":  true,
	"running": true,
}

// Pipeline phases.
const (
	ContextPhasePlanning = "planning"
	ContextPhaseCrawling = "crawling"
	ContextPhaseShaping  = "shaping"
)

// ─── Pillar config ──────────────────────────────────────────────────────

// PillarConfig is the wire shape expected by /v1/context for one Source /
// Strategy / Shape / Reconciler.
type PillarConfig struct {
	Type    string                 `json:"type"`
	Params  map[string]interface{} `json:"params"`
	AuthRef string                 `json:"auth_ref,omitempty"`
}

// MarshalJSON ensures auth_ref is only emitted when set.
func (p PillarConfig) MarshalJSON() ([]byte, error) {
	out := map[string]interface{}{
		"type":   p.Type,
		"params": p.Params,
	}
	if p.AuthRef != "" {
		out["auth_ref"] = p.AuthRef
	}
	return json.Marshal(out)
}

// ─── Source builders ────────────────────────────────────────────────────

// GoogleWebSourceOptions are options for GoogleWebSource.
type GoogleWebSourceOptions struct {
	Backends       []string // Subset of ["google", "bing", "duckduckgo", "brave"].
	TopKPerBackend int      // Per-backend cap before RRF merge (1-50).
	Region         string   // 2-letter country code, e.g. "us", "gb".
}

// GoogleWebSource builds a google_web Source config.
//
// Google search across multiple SERP backends with RRF merge.
func GoogleWebSource(opts *GoogleWebSourceOptions) PillarConfig {
	if opts == nil {
		opts = &GoogleWebSourceOptions{}
	}
	topK := opts.TopKPerBackend
	if topK == 0 {
		topK = 10
	}
	params := map[string]interface{}{
		"top_k_per_backend": topK,
	}
	if len(opts.Backends) > 0 {
		params["backends"] = opts.Backends
	}
	if opts.Region != "" {
		params["region"] = opts.Region
	}
	return PillarConfig{Type: "google_web", Params: params}
}

// CrawlSourceOptions are options for CrawlSource.
type CrawlSourceOptions struct {
	Domain         string  // Required. Root URL or domain to crawl.
	MaxURLs        int     // Hard cap on pages indexed (default 50, range 1-500).
	MaxDepth       int     // Recursion depth from the root (default 3).
	ScoreThreshold float64 // BM25 relevance gate against the intent (0-1). 0 = unset.
	ProfileID      string  // Saved browser-profile for auth.
}

// CrawlSource builds a crawl Source config — recursive site crawl as the corpus.
func CrawlSource(opts CrawlSourceOptions) PillarConfig {
	maxURLs := opts.MaxURLs
	if maxURLs == 0 {
		maxURLs = 50
	}
	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = 3
	}
	params := map[string]interface{}{
		"domain":    opts.Domain,
		"max_urls":  maxURLs,
		"max_depth": maxDepth,
	}
	if opts.ScoreThreshold > 0 {
		params["score_threshold"] = opts.ScoreThreshold
	}
	if opts.ProfileID != "" {
		params["profile_id"] = opts.ProfileID
	}
	return PillarConfig{Type: "crawl", Params: params}
}

// FileSourceOptions are options for FileSource.
type FileSourceOptions struct {
	FileID       string // Required. Reference to a file uploaded via /v1/files/upload.
	ChunkSize    int    // Characters per chunk (default 2000).
	ChunkOverlap int    // Character overlap between adjacent chunks (default 200).
}

// FileSource builds a file Source config — user-uploaded file as the corpus.
func FileSource(opts FileSourceOptions) PillarConfig {
	chunk := opts.ChunkSize
	if chunk == 0 {
		chunk = 2000
	}
	overlap := opts.ChunkOverlap
	if overlap == 0 {
		overlap = 200
	}
	return PillarConfig{
		Type: "file",
		Params: map[string]interface{}{
			"file_id":       opts.FileID,
			"chunk_size":    chunk,
			"chunk_overlap": overlap,
		},
	}
}

// CustomSourceOptions are options for CustomSource.
type CustomSourceOptions struct {
	Type    string                 // Required. Source name (e.g. "hackernews").
	Params  map[string]interface{} // Source-specific params.
	AuthRef string                 // OAuth linked-account ref for private Sources.
}

// CustomSource is the escape hatch for Sources that exist server-side but
// don't yet have a typed builder in this SDK. Discover available Sources
// via crawler.ContextCatalog().
func CustomSource(opts CustomSourceOptions) PillarConfig {
	params := opts.Params
	if params == nil {
		params = map[string]interface{}{}
	}
	return PillarConfig{Type: opts.Type, Params: params, AuthRef: opts.AuthRef}
}

// ─── Strategy / Shape / Reconciler builders ─────────────────────────────

// AllItemsStrategy builds the passthrough Strategy — every candidate item
// is kept up to constraints.MaxItems. The default.
func AllItemsStrategy() PillarConfig {
	return PillarConfig{Type: "all_items", Params: map[string]interface{}{}}
}

// CustomStrategy is the escape hatch for Strategies that ship server-side
// before this SDK adds a typed builder.
func CustomStrategy(type_ string, params map[string]interface{}) PillarConfig {
	if params == nil {
		params = map[string]interface{}{}
	}
	return PillarConfig{Type: type_, Params: params}
}

// RawShape builds the per-item-citation Shape. The default.
func RawShape() PillarConfig {
	return PillarConfig{Type: "raw", Params: map[string]interface{}{}}
}

// CustomShape is the escape hatch for Shapes that ship server-side before
// this SDK adds a typed builder (e.g. markdown_digest, tabular, knowledge_graph).
func CustomShape(type_ string, params map[string]interface{}) PillarConfig {
	if params == nil {
		params = map[string]interface{}{}
	}
	return PillarConfig{Type: type_, Params: params}
}

// NoopReconciler builds the no-auto-refresh Reconciler. The default —
// refreshes are user-initiated via crawler.RefreshContext.
func NoopReconciler() PillarConfig {
	return PillarConfig{Type: "noop", Params: map[string]interface{}{}}
}

// CustomReconciler is the escape hatch for Reconcilers that ship
// server-side before this SDK adds a typed builder (e.g. cron, event).
func CustomReconciler(type_ string, params map[string]interface{}) PillarConfig {
	if params == nil {
		params = map[string]interface{}{}
	}
	return PillarConfig{Type: type_, Params: params}
}

// ─── Constraints ────────────────────────────────────────────────────────

// ContextConstraints are the caller-controllable knobs forwarded to the
// Context pipeline. Zero values mean "use the API default".
type ContextConstraints struct {
	MaxItems       int     // Total items kept after the Strategy plan phase (default 20, 1-200).
	MaxPerSource   int     // Per-Source cap before merge (default 10, 1-100).
	MaxCrawlTimeS  float64 // Hard timeout for the fetch phase (default 120, 0-600).
	FreshnessDays  int     // Drop items older than N days. 0 = unset.
	Language       string  // 2-letter language code (default "en").
}

// ToMap renders the constraints as the wire dict expected by /v1/context.
func (c ContextConstraints) ToMap() map[string]interface{} {
	maxItems := c.MaxItems
	if maxItems == 0 {
		maxItems = 20
	}
	maxPer := c.MaxPerSource
	if maxPer == 0 {
		maxPer = 10
	}
	crawlTime := c.MaxCrawlTimeS
	if crawlTime == 0 {
		crawlTime = 120
	}
	lang := c.Language
	if lang == "" {
		lang = "en"
	}
	out := map[string]interface{}{
		"max_items":        maxItems,
		"max_per_source":   maxPer,
		"max_crawl_time_s": crawlTime,
		"language":         lang,
	}
	if c.FreshnessDays > 0 {
		out["freshness_days"] = c.FreshnessDays
	}
	return out
}

// ─── Output types ───────────────────────────────────────────────────────

// ContextItem is one fetched item — typically one URL the Source query
// phase surfaced and the fetch phase materialised. For the raw Shape,
// each item is the unit of citation.
type ContextItem struct {
	ID         string                 `json:"id,omitempty"`
	URL        string                 `json:"url,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Snippet    string                 `json:"snippet,omitempty"`
	Source     string                 `json:"source,omitempty"`
	Relevance  float64                `json:"relevance,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	FetchedAt  string                 `json:"fetched_at,omitempty"`
}

// ContextItemFromMap builds a ContextItem from a wire-shape map. The raw
// Shape returns `source_name`; the catalog calls it `source`. Accept either.
func ContextItemFromMap(data map[string]interface{}) ContextItem {
	src, _ := data["source"].(string)
	if src == "" {
		if s, ok := data["source_name"].(string); ok {
			src = s
		}
	}
	item := ContextItem{
		Source: src,
	}
	if v, ok := data["id"].(string); ok {
		item.ID = v
	}
	if v, ok := data["url"].(string); ok {
		item.URL = v
	}
	if v, ok := data["title"].(string); ok {
		item.Title = v
	}
	if v, ok := data["content"].(string); ok {
		item.Content = v
	}
	if v, ok := data["snippet"].(string); ok {
		item.Snippet = v
	}
	if v, ok := data["relevance"].(float64); ok {
		item.Relevance = v
	}
	if v, ok := data["metadata"].(map[string]interface{}); ok {
		item.Metadata = v
	}
	if v, ok := data["fetched_at"].(string); ok {
		item.FetchedAt = v
	}
	return item
}

// ContextOutput is the shaped output of a Context run.
//
// For the raw Shape (today), Items carries the fetched URLs with content
// + snippet + source + provenance metadata; each item is the citation
// unit. For future Shapes (markdown_digest, tabular, knowledge_graph)
// the top-level structure may differ — Raw preserves the wire payload
// for forward compat.
type ContextOutput struct {
	Shape   string
	Items   []ContextItem
	Partial bool
	Raw     map[string]interface{}
}

// ContextOutputFromMap builds a ContextOutput from a wire-shape map.
// Wire shape today is {"type": "raw", "data": {"items": [...]}}.
func ContextOutputFromMap(data map[string]interface{}) ContextOutput {
	shape, _ := data["shape"].(string)
	if shape == "" {
		shape, _ = data["type"].(string)
	}
	if shape == "" {
		shape = "raw"
	}

	var itemsData []interface{}
	if payload, ok := data["data"].(map[string]interface{}); ok {
		if items, ok := payload["items"].([]interface{}); ok {
			itemsData = items
		}
	}
	if itemsData == nil {
		if items, ok := data["items"].([]interface{}); ok {
			itemsData = items
		}
	}

	items := make([]ContextItem, 0, len(itemsData))
	for _, raw := range itemsData {
		if m, ok := raw.(map[string]interface{}); ok {
			items = append(items, ContextItemFromMap(m))
		}
	}

	partial, _ := data["partial"].(bool)

	return ContextOutput{
		Shape:   shape,
		Items:   items,
		Partial: partial,
		Raw:     data,
	}
}

// ─── Streaming events ───────────────────────────────────────────────────

// ContextEventType discriminates ContextEvent values.
type ContextEventType string

const (
	ContextEventStatus       ContextEventType = "status"
	ContextEventPhaseInit    ContextEventType = "phase_init"
	ContextEventPhaseItem    ContextEventType = "phase_item_update"
	ContextEventTerminal     ContextEventType = "terminal"
)

// ContextEvent is one typed event from a Context SSE stream. Use the
// concrete fields named after Type — the others are zero-valued.
type ContextEvent struct {
	Type ContextEventType

	// status
	Status     string
	Phase      string
	Version    int
	PlanningMs int64
	CrawlingMs int64
	ShapingMs  int64
	TS         string

	// phase_init (Phase + Total + Items)
	PhaseInitTotal int
	PhaseInitItems []map[string]interface{}

	// phase_item_update
	ItemID     string
	ItemStatus string
	ItemMs     int64
	ItemSize   int
	ItemReason string

	// terminal
	TotalMs       int64
	URLsCrawled   int
	URLsFailed    int
	OutputS3Key   string
	ErrorMessage  string
}

// ParseContextEvent translates a raw SSE (eventName, data) into a typed
// ContextEvent. Returns ok=false for unknown event types
// (forward-compatible). The forward-compat data is preserved on the
// returned event's Raw fields where set.
func ParseContextEvent(eventName string, data map[string]interface{}) (ContextEvent, bool) {
	t, _ := data["type"].(string)
	if t == "" {
		t = eventName
	}

	getInt64 := func(k string) int64 {
		switch v := data[k].(type) {
		case float64:
			return int64(v)
		case int:
			return int64(v)
		case int64:
			return v
		}
		return 0
	}
	getInt := func(k string) int {
		return int(getInt64(k))
	}
	getStr := func(k string) string {
		s, _ := data[k].(string)
		return s
	}

	switch t {
	case "status":
		return ContextEvent{
			Type:       ContextEventStatus,
			Status:     getStr("status"),
			Phase:      getStr("phase"),
			Version:    getInt("version"),
			PlanningMs: getInt64("planning_ms"),
			CrawlingMs: getInt64("crawling_ms"),
			ShapingMs:  getInt64("shaping_ms"),
			TS:         getStr("ts"),
		}, true
	case "terminal":
		return ContextEvent{
			Type:         ContextEventTerminal,
			Status:       getStr("status"),
			TotalMs:      getInt64("total_ms"),
			URLsCrawled:  getInt("urls_crawled"),
			URLsFailed:   getInt("urls_failed"),
			OutputS3Key:  getStr("output_s3_key"),
			ErrorMessage: getStr("error_message"),
		}, true
	case "phase_progress":
		kind, _ := data["kind"].(string)
		switch kind {
		case "init":
			itemsRaw, _ := data["items"].([]interface{})
			items := make([]map[string]interface{}, 0, len(itemsRaw))
			for _, r := range itemsRaw {
				if m, ok := r.(map[string]interface{}); ok {
					items = append(items, m)
				}
			}
			return ContextEvent{
				Type:           ContextEventPhaseInit,
				Phase:          getStr("phase"),
				PhaseInitTotal: getInt("total"),
				PhaseInitItems: items,
			}, true
		case "item_update":
			return ContextEvent{
				Type:       ContextEventPhaseItem,
				Phase:      getStr("phase"),
				ItemID:     getStr("id"),
				ItemStatus: getStr("status"),
				ItemMs:     getInt64("ms"),
				ItemSize:   getInt("size"),
				ItemReason: getStr("reason"),
			}, true
		}
	}
	return ContextEvent{}, false
}

// ─── Diff / Version / Catalog ───────────────────────────────────────────

// ContextVersion is one entry in a run's version chain.
type ContextVersion struct {
	Version      int    `json:"version"`
	Status       string `json:"status"`
	SubmittedAt  string `json:"submitted_at,omitempty"`
	CompletedAt  string `json:"completed_at,omitempty"`
	URLsCrawled  int    `json:"urls_crawled,omitempty"`
	TriggeredBy  string `json:"triggered_by,omitempty"`
	OutputS3Key  string `json:"output_s3_key,omitempty"`
}

// ContextVersionFromMap builds a ContextVersion from a wire-shape map.
func ContextVersionFromMap(data map[string]interface{}) ContextVersion {
	v := ContextVersion{}
	if x, ok := data["version"].(float64); ok {
		v.Version = int(x)
	}
	if x, ok := data["status"].(string); ok {
		v.Status = x
	}
	if x, ok := data["submitted_at"].(string); ok {
		v.SubmittedAt = x
	}
	if x, ok := data["completed_at"].(string); ok {
		v.CompletedAt = x
	}
	if x, ok := data["urls_crawled"].(float64); ok {
		v.URLsCrawled = int(x)
	}
	if x, ok := data["triggered_by"].(string); ok {
		v.TriggeredBy = x
	}
	if x, ok := data["output_s3_key"].(string); ok {
		v.OutputS3Key = x
	}
	if v.TriggeredBy == "" {
		v.TriggeredBy = "user"
	}
	return v
}

// ContextDiff is the item-level diff between two Context versions.
//
// Today the diff is item-level (matched by stable URL). Future versions
// may diff at the claim level once a Shape that emits discrete claims
// (e.g. markdown_digest) is wired through.
type ContextDiff struct {
	Added          []ContextItem          `json:"added"`
	Removed        []ContextItem          `json:"removed"`
	Unchanged      []ContextItem          `json:"unchanged"`
	SourcesAdded   []string               `json:"sources_added"`
	SourcesRemoved []string               `json:"sources_removed"`
	Raw            map[string]interface{} `json:"-"`
}

// ContextDiffFromMap builds a ContextDiff from a wire-shape map.
func ContextDiffFromMap(data map[string]interface{}) ContextDiff {
	asItems := func(k string) []ContextItem {
		raw, _ := data[k].([]interface{})
		items := make([]ContextItem, 0, len(raw))
		for _, r := range raw {
			if m, ok := r.(map[string]interface{}); ok {
				items = append(items, ContextItemFromMap(m))
			}
		}
		return items
	}
	asStrings := func(k string) []string {
		raw, _ := data[k].([]interface{})
		out := make([]string, 0, len(raw))
		for _, r := range raw {
			if s, ok := r.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return ContextDiff{
		Added:          asItems("added"),
		Removed:        asItems("removed"),
		Unchanged:      asItems("unchanged"),
		SourcesAdded:   asStrings("sources_added"),
		SourcesRemoved: asStrings("sources_removed"),
		Raw:            data,
	}
}

// CatalogEntry is one entry from /v1/context/{sources,strategies,shapes,reconcilers}.
type CatalogEntry struct {
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Summary      string                 `json:"summary"`
	HelpMd       string                 `json:"help_md"`
	ParamsSchema map[string]interface{} `json:"params_schema"`
}

// CatalogEntryFromMap builds a CatalogEntry from a wire-shape map.
func CatalogEntryFromMap(data map[string]interface{}) CatalogEntry {
	e := CatalogEntry{}
	if v, ok := data["name"].(string); ok {
		e.Name = v
	}
	if v, ok := data["display_name"].(string); ok {
		e.DisplayName = v
	} else {
		e.DisplayName = e.Name
	}
	if v, ok := data["summary"].(string); ok {
		e.Summary = v
	}
	if v, ok := data["help_md"].(string); ok {
		e.HelpMd = v
	}
	if v, ok := data["params_schema"].(map[string]interface{}); ok {
		e.ParamsSchema = v
	} else if v, ok := data["query_params_schema"].(map[string]interface{}); ok {
		e.ParamsSchema = v
	} else {
		e.ParamsSchema = map[string]interface{}{}
	}
	return e
}

// ContextCatalog bundles per-pillar catalog responses.
type ContextCatalog struct {
	Sources     []CatalogEntry
	Strategies  []CatalogEntry
	Shapes      []CatalogEntry
	Reconcilers []CatalogEntry
}

// ─── ContextResult — the run state with lazy output ─────────────────────

// ContextResult is the state of a Context run. Output() lazily fetches
// the shaped output via the parent crawler — the first call hits the
// API; subsequent calls return the cached value.
type ContextResult struct {
	RunID        string
	Status       string
	Version      int
	Phase        string
	GeneratorID  string
	Intent       string
	Constraints  map[string]interface{}
	Stats        map[string]interface{}
	ErrorMessage string
	SubmittedAt  string
	CompletedAt  string

	crawler *AsyncWebCrawler
	output  *ContextOutput
}

// IsTerminal returns true on any of the terminal statuses.
func (r *ContextResult) IsTerminal() bool {
	return ContextTerminalStatuses[r.Status]
}

// IsSuccess returns true only on completed / completed_partial.
func (r *ContextResult) IsSuccess() bool {
	return r.Status == "completed" || r.Status == "completed_partial"
}

// Output fetches the shaped output for this run. Cached after the first
// call; safe to call multiple times.
func (r *ContextResult) Output() (*ContextOutput, error) {
	if r.output != nil {
		return r.output, nil
	}
	if r.crawler == nil {
		return nil, fmt.Errorf(
			"ContextResult was built without a crawler reference; " +
				"cannot fetch output. Use crawler.GetContextOutput(runID)",
		)
	}
	out, err := r.crawler.GetContextOutput(r.RunID)
	if err != nil {
		return nil, err
	}
	r.output = &out
	return r.output, nil
}

// ContextResultFromMap builds a ContextResult from a wire-shape map. The
// crawler reference is optional; without it Output() returns an error
// pointing at GetContextOutput.
func ContextResultFromMap(data map[string]interface{}, crawler *AsyncWebCrawler) *ContextResult {
	// GET /{run_id} returns the row with `id` as primary key; POST submit
	// returns `run_id`. Handle both.
	runID, _ := data["run_id"].(string)
	if runID == "" {
		runID, _ = data["id"].(string)
	}

	r := &ContextResult{
		RunID:   runID,
		crawler: crawler,
	}
	if v, ok := data["status"].(string); ok {
		r.Status = v
	}
	if v, ok := data["version"].(float64); ok {
		r.Version = int(v)
	}
	if r.Version == 0 {
		r.Version = 1
	}
	if v, ok := data["phase"].(string); ok {
		r.Phase = v
	}
	if v, ok := data["generator_id"].(string); ok {
		r.GeneratorID = v
	}
	if v, ok := data["intent"].(string); ok {
		r.Intent = v
	}
	if v, ok := data["constraints"].(map[string]interface{}); ok {
		r.Constraints = v
	} else {
		r.Constraints = map[string]interface{}{}
	}

	stats := map[string]interface{}{}
	if v, ok := data["stats"].(map[string]interface{}); ok {
		for k, val := range v {
			stats[k] = val
		}
	}
	for _, k := range []string{
		"planning_ms", "crawling_ms", "shaping_ms", "total_ms",
		"urls_crawled", "urls_failed", "output_size_bytes",
	} {
		if v, ok := data[k]; ok && v != nil {
			stats[k] = v
		}
	}
	r.Stats = stats

	if v, ok := data["error_message"].(string); ok {
		r.ErrorMessage = v
	}
	if v, ok := data["submitted_at"].(string); ok {
		r.SubmittedAt = v
	} else if v, ok := data["created_at"].(string); ok {
		r.SubmittedAt = v
	}
	if v, ok := data["completed_at"].(string); ok {
		r.CompletedAt = v
	}
	return r
}

// ─── ContextNotImplementedError ─────────────────────────────────────────

// ContextNotImplementedError is raised by client-side validation when a
// feature isn't yet wired through the public API. Used today by
// crawler.Context when ad-hoc pillar configs are passed without a
// GeneratorID (custom pillars must be wrapped in a saved generator
// until public generator CRUD ships on the API-key surface).
type ContextNotImplementedError struct {
	Message string
}

// Error implements error.
func (e *ContextNotImplementedError) Error() string {
	return e.Message
}

// IsContextNotImplementedError reports whether err is a
// ContextNotImplementedError (or wraps one).
func IsContextNotImplementedError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*ContextNotImplementedError)
	if ok {
		return true
	}
	// strings match for wrapped errors
	return strings.Contains(err.Error(), "public generator CRUD")
}

// ─── Crawler methods ────────────────────────────────────────────────────

// ContextOptions are options for AsyncWebCrawler.Context. The zero value
// uses the user's default generator (google_web + all_items + raw + noop).
//
// Pillar params (Sources/Strategy/Shape/Reconciler) are reserved for the
// day public generator CRUD ships on the API-key surface. Until then,
// passing them raises a ContextNotImplementedError pointing at the
// dashboard.
type ContextOptions struct {
	Intent         string
	Mission        string
	GeneratorID    string
	Sources        []PillarConfig
	Strategy       *PillarConfig
	Shape          *PillarConfig
	Reconciler     *PillarConfig
	Constraints    *ContextConstraints
	WebhookURL     string
	IdempotencyKey string

	// Wait — when true (default), the call streams until terminal then returns.
	// Use ContextNoWait{} to opt out.
	Wait         bool
	NoWait       bool
	PollInterval time.Duration
	Timeout      time.Duration
}

func (c *AsyncWebCrawler) buildContextBody(opts ContextOptions) (map[string]interface{}, error) {
	hasPillars := len(opts.Sources) > 0 || opts.Strategy != nil ||
		opts.Shape != nil || opts.Reconciler != nil

	if opts.GeneratorID != "" && hasPillars {
		return nil, fmt.Errorf(
			"pass either GeneratorID OR pillar params (Sources/Strategy/Shape/Reconciler), not both",
		)
	}
	if hasPillars {
		return nil, &ContextNotImplementedError{
			Message: "Custom Source/Strategy/Shape/Reconciler configs aren't yet " +
				"accepted by the public /v1/context endpoint. Today, custom " +
				"pillars must be wrapped in a saved generator — create one " +
				"in the dashboard (/context page), then pass its GeneratorID " +
				"to crawler.Context(). Pillar builders (e.g. GoogleWebSource(...)) " +
				"are still useful for inspecting and serialising configs locally; " +
				"the SDK will auto-create-and-submit transparently once public " +
				"generator CRUD ships.",
		}
	}

	if opts.Intent == "" || strings.TrimSpace(opts.Intent) == "" {
		return nil, fmt.Errorf("Intent is required and must be non-empty")
	}

	body := map[string]interface{}{"intent": opts.Intent}
	if opts.Mission != "" {
		body["mission"] = opts.Mission
	}
	if opts.GeneratorID != "" {
		body["generator_id"] = opts.GeneratorID
	}
	if opts.Constraints != nil {
		body["constraints"] = opts.Constraints.ToMap()
	}
	if opts.WebhookURL != "" {
		body["webhook_url"] = opts.WebhookURL
	}
	return body, nil
}

// Context submits a Context run.
//
// One-liner — uses the user's default generator (google_web + all_items
// + raw + noop):
//
//	result, err := crawler.Context(ContextOptions{
//	    Intent: "compare LangChain and AutoGen",
//	})
//	output, _ := result.Output()
//	for _, item := range output.Items {
//	    fmt.Println(item.Title, "—", item.URL)
//	}
//
// Pillar params are reserved for the day public generator CRUD ships on
// the API-key surface. Until then, custom pillars must be wrapped in a
// saved generator (created via the dashboard); pass its GeneratorID.
func (c *AsyncWebCrawler) Context(opts ContextOptions) (*ContextResult, error) {
	body, err := c.buildContextBody(opts)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if opts.IdempotencyKey != "" {
		headers["Idempotency-Key"] = opts.IdempotencyKey
	}

	reqOpts := RequestOptions{
		Method:  "POST",
		Path:    "/v1/context",
		Body:    body,
		Timeout: 30 * time.Second,
	}
	if len(headers) > 0 {
		reqOpts.Headers = headers
	}
	data, err := c.http.Request(reqOpts)
	if err != nil {
		return nil, err
	}
	runID, _ := data["run_id"].(string)
	if runID == "" {
		return nil, fmt.Errorf("submit response missing run_id: %v", data)
	}

	result, err := c.GetContextRun(runID)
	if err != nil {
		return nil, err
	}

	wait := !opts.NoWait // default: wait
	if !wait || result.IsTerminal() {
		return result, nil
	}

	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 3 * time.Second
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}
	return c.waitContextRun(runID, pollInterval, timeout)
}

// ContextStream submits (or attaches to) a Context run and pushes typed
// events on the returned channel. The channel closes when the stream
// reaches a terminal event or the context is cancelled. If submit fails,
// an error is returned synchronously and the channel is nil.
//
// Two modes:
//  1. Submit + stream — set opts.Intent (and optional pillar params).
//  2. Attach — set opts.RunID. opts.Intent is ignored.
type ContextStreamOptions struct {
	Intent         string
	RunID          string // if set, attach to existing run
	Mission        string
	GeneratorID    string
	Sources        []PillarConfig
	Strategy       *PillarConfig
	Shape          *PillarConfig
	Reconciler     *PillarConfig
	Constraints    *ContextConstraints
	WebhookURL     string
	IdempotencyKey string
}

// ContextStream submits (or attaches to) a Context run and pushes typed
// events on the returned channel.
func (c *AsyncWebCrawler) ContextStream(ctx context.Context, opts ContextStreamOptions) (<-chan ContextEvent, error) {
	runID := opts.RunID
	if runID == "" {
		if opts.Intent == "" {
			return nil, fmt.Errorf("set Intent to submit + stream, or RunID to attach")
		}
		body, err := c.buildContextBody(ContextOptions{
			Intent:      opts.Intent,
			Mission:     opts.Mission,
			GeneratorID: opts.GeneratorID,
			Sources:     opts.Sources,
			Strategy:    opts.Strategy,
			Shape:       opts.Shape,
			Reconciler:  opts.Reconciler,
			Constraints: opts.Constraints,
			WebhookURL:  opts.WebhookURL,
		})
		if err != nil {
			return nil, err
		}
		headers := map[string]string{}
		if opts.IdempotencyKey != "" {
			headers["Idempotency-Key"] = opts.IdempotencyKey
		}
		reqOpts := RequestOptions{
			Method:  "POST",
			Path:    "/v1/context",
			Body:    body,
			Timeout: 30 * time.Second,
		}
		if len(headers) > 0 {
			reqOpts.Headers = headers
		}
		data, err := c.http.Request(reqOpts)
		if err != nil {
			return nil, err
		}
		runID, _ = data["run_id"].(string)
		if runID == "" {
			return nil, fmt.Errorf("submit response missing run_id: %v", data)
		}
	}

	raw, err := c.http.StreamSse(ctx, fmt.Sprintf("/v1/context/%s/stream", runID), nil)
	if err != nil {
		return nil, err
	}

	out := make(chan ContextEvent, 16)
	go func() {
		defer close(out)
		for sse := range raw {
			if sse.Err != nil {
				return
			}
			ev, ok := ParseContextEvent(sse.Event, sse.Data)
			if !ok {
				continue
			}
			select {
			case out <- ev:
			case <-ctx.Done():
				return
			}
			if ev.Type == ContextEventTerminal {
				return
			}
		}
	}()
	return out, nil
}

// waitContextRun streams a Context run to terminal, falling back to polling
// if the stream dies before terminal.
func (c *AsyncWebCrawler) waitContextRun(runID string, pollInterval, timeout time.Duration) (*ContextResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	events, streamErr := c.ContextStream(ctx, ContextStreamOptions{RunID: runID})
	if streamErr == nil {
		for ev := range events {
			if ev.Type == ContextEventTerminal {
				break
			}
		}
	}
	// Either stream finished cleanly, errored, or context cancelled —
	// always confirm the final state via the GET endpoint.
	deadline := time.Now().Add(timeout)
	for {
		result, err := c.GetContextRun(runID)
		if err != nil {
			return nil, err
		}
		if result.IsTerminal() {
			return result, nil
		}
		if time.Now().After(deadline) {
			return nil, NewTimeoutError(fmt.Sprintf("Context run %s did not terminate within %s", runID, timeout))
		}
		time.Sleep(pollInterval)
	}
}

// GetContextRun fetches the current state of a Context run.
func (c *AsyncWebCrawler) GetContextRun(runID string) (*ContextResult, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/context/%s", runID), nil)
	if err != nil {
		return nil, err
	}
	return ContextResultFromMap(data, c), nil
}

// GetContextOutput fetches the ShapedOutput for a Context run.
func (c *AsyncWebCrawler) GetContextOutput(runID string) (ContextOutput, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/context/%s/output", runID), nil)
	if err != nil {
		return ContextOutput{}, err
	}
	return ContextOutputFromMap(data), nil
}

// CancelContextRun cancels an in-flight Context run. Server-side
// cancellation is asynchronous — the row may report `running` for a few
// hundred ms before flipping to `cancelled`.
func (c *AsyncWebCrawler) CancelContextRun(runID string) error {
	_, err := c.http.Delete(fmt.Sprintf("/v1/context/%s", runID))
	return err
}

// RefreshContextOptions are options for AsyncWebCrawler.RefreshContext.
type RefreshContextOptions struct {
	NoWait       bool
	PollInterval time.Duration
	Timeout      time.Duration
}

// RefreshContext creates a new version on the same chain. Re-runs the
// entire pipeline against the same generator config. Returns the new
// version's ContextResult.
func (c *AsyncWebCrawler) RefreshContext(runID string, opts *RefreshContextOptions) (*ContextResult, error) {
	if opts == nil {
		opts = &RefreshContextOptions{}
	}
	data, err := c.http.Post(fmt.Sprintf("/v1/context/%s/refresh", runID), nil, 30*time.Second)
	if err != nil {
		return nil, err
	}
	newRunID, _ := data["run_id"].(string)
	if newRunID == "" {
		newRunID = runID
	}
	result, err := c.GetContextRun(newRunID)
	if err != nil {
		return nil, err
	}
	if opts.NoWait || result.IsTerminal() {
		return result, nil
	}
	pollInterval := opts.PollInterval
	if pollInterval == 0 {
		pollInterval = 3 * time.Second
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute
	}
	return c.waitContextRun(newRunID, pollInterval, timeout)
}

// ListContextVersions lists all versions on a Context run's chain (newest last).
func (c *AsyncWebCrawler) ListContextVersions(runID string) ([]ContextVersion, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/context/%s/versions", runID), nil)
	if err != nil {
		return nil, err
	}
	raw, _ := data["versions"].([]interface{})
	out := make([]ContextVersion, 0, len(raw))
	for _, r := range raw {
		if m, ok := r.(map[string]interface{}); ok {
			out = append(out, ContextVersionFromMap(m))
		}
	}
	return out, nil
}

// DiffContext diffs two Context versions. When both ids are on the same
// chain, the diff is between the two most recent versions; when they're
// on different chains, the diff is cross-chain.
func (c *AsyncWebCrawler) DiffContext(runID, otherRunID string) (ContextDiff, error) {
	data, err := c.http.Get(fmt.Sprintf("/v1/context/%s/diff/%s", runID, otherRunID), nil)
	if err != nil {
		return ContextDiff{}, err
	}
	return ContextDiffFromMap(data), nil
}

// RollbackContext moves the chain's current pointer back to an earlier
// version. Pointer move, not delete — the newer version stays on the
// chain and you can roll forward again.
func (c *AsyncWebCrawler) RollbackContext(runID string, version int) (*ContextResult, error) {
	_, err := c.http.Post(fmt.Sprintf("/v1/context/%s/rollback/%d", runID, version), nil, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return c.GetContextRun(runID)
}

// ContextCatalog discovers what Sources / Strategies / Shapes /
// Reconcilers are available. Useful for building a generator-creation UI.
func (c *AsyncWebCrawler) ContextCatalog() (ContextCatalog, error) {
	fetch := func(path string) []CatalogEntry {
		data, err := c.http.Get(path, nil)
		if err != nil {
			return nil
		}
		var items []interface{}
		if v, ok := data["items"].([]interface{}); ok {
			items = v
		}
		out := make([]CatalogEntry, 0, len(items))
		for _, r := range items {
			if m, ok := r.(map[string]interface{}); ok {
				out = append(out, CatalogEntryFromMap(m))
			}
		}
		return out
	}
	return ContextCatalog{
		Sources:     fetch("/v1/context/sources"),
		Strategies:  fetch("/v1/context/strategies"),
		Shapes:      fetch("/v1/context/shapes"),
		Reconcilers: fetch("/v1/context/reconcilers"),
	}, nil
}
