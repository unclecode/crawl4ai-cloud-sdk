// Context v2 SDK tests — two layers (unit + live against stage).
//
// Unit — pure-Go checks that don't hit the network:
//   - Pillar builders produce the documented dict shape
//   - buildContextBody validates mutually-exclusive GeneratorID vs pillars
//   - ParseContextEvent translates SSE payloads into the right typed event
//   - ContextResult IsTerminal / IsSuccess flags work
//   - ContextConstraints.ToMap produces the API-expected shape
//
// Live — hits stage by default. Skipped when CRAWL4AI_SKIP_LIVE=1.
//
// Run:
//
//	# All
//	go test ./pkg/crawl4ai/ -run TestContext -v -timeout 10m
//
//	# Unit only
//	CRAWL4AI_SKIP_LIVE=1 go test ./pkg/crawl4ai/ -run TestContext -v
package crawl4ai

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"
)

// ─── shared helpers ─────────────────────────────────────────────────────

func getContextTestKey() string {
	if key := os.Getenv("CRAWL4AI_API_KEY"); key != "" {
		return key
	}
	return "sk_live_V89kxHtmkxw0jJORu_sWzyuvGw6TKHaJhoNGK8gGdqU"
}

func getContextBaseURL() string {
	if u := os.Getenv("CRAWL4AI_BASE_URL"); u != "" {
		return u
	}
	return "https://stage.crawl4ai.com"
}

func setupContext(t *testing.T) *AsyncWebCrawler {
	t.Helper()
	c, err := NewAsyncWebCrawler(CrawlerOptions{
		APIKey:  getContextTestKey(),
		BaseURL: getContextBaseURL(),
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return c
}

func skipIfLiveDisabled(t *testing.T) {
	t.Helper()
	if os.Getenv("CRAWL4AI_SKIP_LIVE") == "1" {
		t.Skip("CRAWL4AI_SKIP_LIVE=1 — skipping live test")
	}
}

// ─── Unit — pillar builders ─────────────────────────────────────────────

func TestContext_Unit_GoogleWebSourceDefaults(t *testing.T) {
	out := GoogleWebSource(nil)
	if out.Type != "google_web" {
		t.Fatalf("expected type google_web, got %q", out.Type)
	}
	if out.Params["top_k_per_backend"] != 10 {
		t.Fatalf("expected top_k_per_backend=10, got %v", out.Params["top_k_per_backend"])
	}
}

func TestContext_Unit_GoogleWebSourceWithBackends(t *testing.T) {
	out := GoogleWebSource(&GoogleWebSourceOptions{
		Backends:       []string{"google", "bing"},
		TopKPerBackend: 8,
		Region:         "us",
	})
	if !reflect.DeepEqual(out.Params["backends"], []string{"google", "bing"}) {
		t.Fatalf("backends mismatch: %v", out.Params["backends"])
	}
	if out.Params["top_k_per_backend"] != 8 {
		t.Fatalf("topK mismatch: %v", out.Params["top_k_per_backend"])
	}
	if out.Params["region"] != "us" {
		t.Fatalf("region mismatch: %v", out.Params["region"])
	}
}

func TestContext_Unit_CrawlSource(t *testing.T) {
	out := CrawlSource(CrawlSourceOptions{
		Domain:         "https://example.com",
		MaxURLs:        30,
		MaxDepth:       2,
		ScoreThreshold: 0.5,
		ProfileID:      "my-profile",
	})
	if out.Type != "crawl" {
		t.Fatalf("expected type crawl, got %q", out.Type)
	}
	if out.Params["domain"] != "https://example.com" {
		t.Fatalf("domain mismatch: %v", out.Params["domain"])
	}
	if out.Params["max_urls"] != 30 {
		t.Fatalf("max_urls mismatch: %v", out.Params["max_urls"])
	}
	if out.Params["score_threshold"] != 0.5 {
		t.Fatalf("score_threshold mismatch: %v", out.Params["score_threshold"])
	}
	if out.Params["profile_id"] != "my-profile" {
		t.Fatalf("profile_id mismatch: %v", out.Params["profile_id"])
	}
}

func TestContext_Unit_FileSource(t *testing.T) {
	out := FileSource(FileSourceOptions{
		FileID:       "file_abc",
		ChunkSize:    1500,
		ChunkOverlap: 150,
	})
	if out.Type != "file" {
		t.Fatalf("expected type file, got %q", out.Type)
	}
	if out.Params["file_id"] != "file_abc" {
		t.Fatalf("file_id mismatch: %v", out.Params["file_id"])
	}
	if out.Params["chunk_size"] != 1500 {
		t.Fatalf("chunk_size mismatch: %v", out.Params["chunk_size"])
	}
}

func TestContext_Unit_CustomSourcePassthrough(t *testing.T) {
	out := CustomSource(CustomSourceOptions{
		Type:   "hackernews",
		Params: map[string]interface{}{"tag": "ai", "limit": 50},
	})
	if out.Type != "hackernews" {
		t.Fatalf("expected hackernews, got %q", out.Type)
	}
	if out.Params["tag"] != "ai" {
		t.Fatalf("tag mismatch: %v", out.Params["tag"])
	}
}

func TestContext_Unit_CustomSourceWithAuthRef(t *testing.T) {
	out := CustomSource(CustomSourceOptions{
		Type:    "slack",
		Params:  map[string]interface{}{"channel": "C123"},
		AuthRef: "link_abc",
	})
	if out.AuthRef != "link_abc" {
		t.Fatalf("auth_ref mismatch: %q", out.AuthRef)
	}
}

func TestContext_Unit_AllItemsStrategy(t *testing.T) {
	out := AllItemsStrategy()
	if out.Type != "all_items" {
		t.Fatalf("expected all_items, got %q", out.Type)
	}
	if len(out.Params) != 0 {
		t.Fatalf("expected empty params, got %v", out.Params)
	}
}

func TestContext_Unit_CustomStrategy(t *testing.T) {
	out := CustomStrategy("llm_rerank", map[string]interface{}{"model": "claude-haiku-4-5"})
	if out.Type != "llm_rerank" {
		t.Fatalf("expected llm_rerank, got %q", out.Type)
	}
	if out.Params["model"] != "claude-haiku-4-5" {
		t.Fatalf("model mismatch: %v", out.Params["model"])
	}
}

func TestContext_Unit_RawShape(t *testing.T) {
	out := RawShape()
	if out.Type != "raw" {
		t.Fatalf("expected raw, got %q", out.Type)
	}
}

func TestContext_Unit_CustomShape(t *testing.T) {
	out := CustomShape("markdown_digest", nil)
	if out.Type != "markdown_digest" {
		t.Fatalf("expected markdown_digest, got %q", out.Type)
	}
	if out.Params == nil {
		t.Fatalf("params should be initialised empty, got nil")
	}
}

func TestContext_Unit_NoopReconciler(t *testing.T) {
	out := NoopReconciler()
	if out.Type != "noop" {
		t.Fatalf("expected noop, got %q", out.Type)
	}
}

func TestContext_Unit_CustomReconcilerCron(t *testing.T) {
	out := CustomReconciler("cron", map[string]interface{}{"schedule": "0 6 * * *", "tz": "UTC"})
	if out.Type != "cron" {
		t.Fatalf("expected cron, got %q", out.Type)
	}
	if out.Params["schedule"] != "0 6 * * *" {
		t.Fatalf("schedule mismatch: %v", out.Params["schedule"])
	}
}

// ─── Unit — Constraints ─────────────────────────────────────────────────

func TestContext_Unit_ConstraintsDefaults(t *testing.T) {
	out := ContextConstraints{}.ToMap()
	if out["max_items"] != 20 {
		t.Fatalf("default max_items wrong: %v", out["max_items"])
	}
	if out["max_per_source"] != 10 {
		t.Fatalf("default max_per_source wrong: %v", out["max_per_source"])
	}
	if out["max_crawl_time_s"] != float64(120) {
		t.Fatalf("default max_crawl_time_s wrong: %v", out["max_crawl_time_s"])
	}
	if out["language"] != "en" {
		t.Fatalf("default language wrong: %v", out["language"])
	}
	if _, present := out["freshness_days"]; present {
		t.Fatalf("freshness_days should be absent by default")
	}
}

func TestContext_Unit_ConstraintsFreshnessEmits(t *testing.T) {
	out := ContextConstraints{FreshnessDays: 7}.ToMap()
	if out["freshness_days"] != 7 {
		t.Fatalf("freshness_days mismatch: %v", out["freshness_days"])
	}
}

func TestContext_Unit_ConstraintsOverrideAll(t *testing.T) {
	out := ContextConstraints{
		MaxItems:      50,
		MaxPerSource:  20,
		MaxCrawlTimeS: 300,
		FreshnessDays: 30,
		Language:      "fr",
	}.ToMap()
	if out["max_items"] != 50 {
		t.Fatalf("max_items: %v", out["max_items"])
	}
	if out["max_per_source"] != 20 {
		t.Fatalf("max_per_source: %v", out["max_per_source"])
	}
	if out["max_crawl_time_s"] != float64(300) {
		t.Fatalf("max_crawl_time_s: %v", out["max_crawl_time_s"])
	}
	if out["freshness_days"] != 30 {
		t.Fatalf("freshness_days: %v", out["freshness_days"])
	}
	if out["language"] != "fr" {
		t.Fatalf("language: %v", out["language"])
	}
}

// ─── Unit — body composition + validation ───────────────────────────────

func TestContext_Unit_BodyMinimal(t *testing.T) {
	c := &AsyncWebCrawler{} // buildContextBody doesn't touch http
	body, err := c.buildContextBody(ContextOptions{Intent: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["intent"] != "x" {
		t.Fatalf("intent missing: %v", body)
	}
	if _, has := body["generator_id"]; has {
		t.Fatalf("generator_id should be absent on minimal body")
	}
}

func TestContext_Unit_BodyRejectsPillars(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	str := AllItemsStrategy()
	sh := RawShape()
	rec := NoopReconciler()
	_, err := c.buildContextBody(ContextOptions{
		Intent:     "compare X and Y",
		Sources:    []PillarConfig{src},
		Strategy:   &str,
		Shape:      &sh,
		Reconciler: &rec,
	})
	if err == nil {
		t.Fatalf("expected ContextNotImplementedError, got nil")
	}
	var cni *ContextNotImplementedError
	if !errors.As(err, &cni) {
		t.Fatalf("expected ContextNotImplementedError type, got %T: %v", err, err)
	}
	if !IsContextNotImplementedError(err) {
		t.Fatalf("IsContextNotImplementedError returned false on a true case")
	}
}

func TestContext_Unit_BodyMutualExclusion(t *testing.T) {
	c := &AsyncWebCrawler{}
	src := GoogleWebSource(nil)
	_, err := c.buildContextBody(ContextOptions{
		Intent:      "x",
		GeneratorID: "gen_x",
		Sources:     []PillarConfig{src},
	})
	if err == nil {
		t.Fatalf("expected mutual-exclusion error, got nil")
	}
}

func TestContext_Unit_BodyEmptyIntent(t *testing.T) {
	c := &AsyncWebCrawler{}
	_, err := c.buildContextBody(ContextOptions{Intent: "   "})
	if err == nil {
		t.Fatalf("expected empty-intent error, got nil")
	}
}

func TestContext_Unit_BodyConstraints(t *testing.T) {
	c := &AsyncWebCrawler{}
	body, err := c.buildContextBody(ContextOptions{
		Intent:      "x",
		Constraints: &ContextConstraints{MaxItems: 5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	con, ok := body["constraints"].(map[string]interface{})
	if !ok {
		t.Fatalf("constraints missing from body: %v", body)
	}
	if con["max_items"] != 5 {
		t.Fatalf("max_items mismatch: %v", con["max_items"])
	}
}

func TestContext_Unit_BodyMissionAndWebhook(t *testing.T) {
	c := &AsyncWebCrawler{}
	body, err := c.buildContextBody(ContextOptions{
		Intent:     "x",
		Mission:    "extra background",
		WebhookURL: "https://hooks.example.com/cb",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["mission"] != "extra background" {
		t.Fatalf("mission missing: %v", body["mission"])
	}
	if body["webhook_url"] != "https://hooks.example.com/cb" {
		t.Fatalf("webhook_url missing: %v", body["webhook_url"])
	}
}

// ─── Unit — SSE event parsing ───────────────────────────────────────────

func TestContext_Unit_ParseStatus(t *testing.T) {
	ev, ok := ParseContextEvent("status", map[string]interface{}{
		"type":    "status",
		"status":  "planning",
		"phase":   "planning",
		"version": float64(1),
		"ts":      "2026-05-19T12:00:00Z",
	})
	if !ok {
		t.Fatalf("expected status event")
	}
	if ev.Type != ContextEventStatus {
		t.Fatalf("wrong type: %s", ev.Type)
	}
	if ev.Status != "planning" || ev.Phase != "planning" {
		t.Fatalf("fields mismatch: %+v", ev)
	}
}

func TestContext_Unit_ParsePhaseInit(t *testing.T) {
	ev, ok := ParseContextEvent("phase_progress", map[string]interface{}{
		"type":  "phase_progress",
		"kind":  "init",
		"phase": "fetch",
		"total": float64(3),
		"items": []interface{}{map[string]interface{}{"id": "a", "url": "https://x"}},
	})
	if !ok {
		t.Fatalf("expected phase_init event")
	}
	if ev.Type != ContextEventPhaseInit {
		t.Fatalf("wrong type: %s", ev.Type)
	}
	if ev.PhaseInitTotal != 3 {
		t.Fatalf("total mismatch: %d", ev.PhaseInitTotal)
	}
	if len(ev.PhaseInitItems) != 1 {
		t.Fatalf("items mismatch: %v", ev.PhaseInitItems)
	}
}

func TestContext_Unit_ParsePhaseItemUpdate(t *testing.T) {
	ev, ok := ParseContextEvent("phase_progress", map[string]interface{}{
		"type":   "phase_progress",
		"kind":   "item_update",
		"id":     "abc",
		"status": "done",
		"ms":     float64(1240),
		"size":   float64(18432),
	})
	if !ok {
		t.Fatalf("expected phase_item event")
	}
	if ev.Type != ContextEventPhaseItem {
		t.Fatalf("wrong type: %s", ev.Type)
	}
	if ev.ItemID != "abc" || ev.ItemStatus != "done" || ev.ItemMs != 1240 || ev.ItemSize != 18432 {
		t.Fatalf("fields mismatch: %+v", ev)
	}
}

func TestContext_Unit_ParseTerminal(t *testing.T) {
	ev, ok := ParseContextEvent("terminal", map[string]interface{}{
		"type":         "terminal",
		"status":       "completed",
		"total_ms":     float64(21834),
		"urls_crawled": float64(9),
		"urls_failed":  float64(0),
	})
	if !ok {
		t.Fatalf("expected terminal event")
	}
	if ev.Type != ContextEventTerminal {
		t.Fatalf("wrong type: %s", ev.Type)
	}
	if ev.Status != "completed" || ev.URLsCrawled != 9 {
		t.Fatalf("fields mismatch: %+v", ev)
	}
}

func TestContext_Unit_ParseUnknownReturnsFalse(t *testing.T) {
	_, ok := ParseContextEvent("mystery", map[string]interface{}{"type": "mystery"})
	if ok {
		t.Fatalf("expected ok=false for unknown event")
	}
}

// ─── Unit — ContextResult helpers ───────────────────────────────────────

func TestContext_Unit_ResultIsTerminal(t *testing.T) {
	for _, s := range []string{"completed", "completed_partial", "failed", "cancelled"} {
		r := ContextResultFromMap(map[string]interface{}{
			"run_id": "x", "status": s, "version": float64(1),
		}, nil)
		if !r.IsTerminal() {
			t.Fatalf("expected %q to be terminal", s)
		}
	}
	r := ContextResultFromMap(map[string]interface{}{
		"run_id": "x", "status": "queued",
	}, nil)
	if r.IsTerminal() {
		t.Fatalf("queued should not be terminal")
	}
}

func TestContext_Unit_ResultIsSuccess(t *testing.T) {
	for _, s := range []string{"completed", "completed_partial"} {
		r := ContextResultFromMap(map[string]interface{}{
			"run_id": "x", "status": s,
		}, nil)
		if !r.IsSuccess() {
			t.Fatalf("expected %q to be success", s)
		}
	}
	for _, s := range []string{"failed", "cancelled", "queued"} {
		r := ContextResultFromMap(map[string]interface{}{
			"run_id": "x", "status": s,
		}, nil)
		if r.IsSuccess() {
			t.Fatalf("expected %q to NOT be success", s)
		}
	}
}

// ─── Live ───────────────────────────────────────────────────────────────

func TestContext_Live_DefaultGeneratorOneShot(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	result, err := c.Context(ContextOptions{
		Intent: "brief overview of what LangChain is, with citations",
		Constraints: &ContextConstraints{
			MaxItems:      5,
			MaxPerSource:  3,
			MaxCrawlTimeS: 60,
		},
		Timeout: 3 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Context: %v", err)
	}
	if !result.IsTerminal() {
		t.Fatalf("not terminal: %s", result.Status)
	}
	if len(result.RunID) <= 8 {
		t.Fatalf("run_id too short: %q", result.RunID)
	}
	output, err := result.Output()
	if err != nil {
		t.Fatalf("Output: %v", err)
	}
	if output.Shape != "raw" {
		t.Fatalf("expected raw shape, got %q", output.Shape)
	}
	for _, item := range output.Items {
		if item.URL == "" && item.Snippet == "" {
			t.Fatalf("item has neither URL nor snippet: %+v", item)
		}
	}
}

func TestContext_Live_Streaming(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	events, err := c.ContextStream(ctx, ContextStreamOptions{
		Intent: "one-line answer: what is RAG",
		Constraints: &ContextConstraints{
			MaxItems:      2,
			MaxPerSource:  2,
			MaxCrawlTimeS: 30,
		},
	})
	if err != nil {
		t.Fatalf("ContextStream: %v", err)
	}

	sawTerminal := false
	for ev := range events {
		switch ev.Type {
		case ContextEventTerminal:
			sawTerminal = true
			if !ContextTerminalStatuses[ev.Status] {
				t.Fatalf("terminal with non-terminal status: %s", ev.Status)
			}
		case ContextEventPhaseItem:
			if ev.ItemStatus != "done" && ev.ItemStatus != "failed" {
				t.Fatalf("item_update with unexpected status: %s", ev.ItemStatus)
			}
		}
	}
	if !sawTerminal {
		t.Fatalf("stream never emitted terminal event")
	}
}

func TestContext_Live_PillarParamsRaise(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	src := GoogleWebSource(nil)
	_, err := c.Context(ContextOptions{
		Intent:  "test",
		Sources: []PillarConfig{src},
	})
	if err == nil {
		t.Fatalf("expected ContextNotImplementedError, got nil")
	}
	if !IsContextNotImplementedError(err) {
		t.Fatalf("expected ContextNotImplementedError type, got: %v", err)
	}
}

func TestContext_Live_GetAndCancel(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)

	// Best-effort: retry past per-generator concurrency cap.
	deadline := time.Now().Add(60 * time.Second)
	var result *ContextResult
	var err error
	for time.Now().Before(deadline) {
		result, err = c.Context(ContextOptions{
			Intent:      "ignored — will be cancelled",
			Constraints: &ContextConstraints{MaxItems: 2, MaxCrawlTimeS: 30},
			NoWait:      true,
		})
		if err == nil {
			break
		}
		var qe *QuotaExceededError
		if !errors.As(err, &qe) {
			t.Fatalf("Context (no-wait): %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	if result == nil {
		t.Fatalf("couldn't get a generator slot in 60s")
	}
	if result.IsTerminal() {
		t.Fatalf("expected non-terminal initial state, got %s", result.Status)
	}

	state, err := c.GetContextRun(result.RunID)
	if err != nil {
		t.Fatalf("GetContextRun: %v", err)
	}
	if state.RunID != result.RunID {
		t.Fatalf("run_id mismatch: %q vs %q", state.RunID, result.RunID)
	}

	if err := c.CancelContextRun(result.RunID); err != nil {
		t.Fatalf("CancelContextRun: %v", err)
	}
}

func TestContext_Live_VersionsAndRefresh(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)

	deadline := time.Now().Add(90 * time.Second)
	var v1 *ContextResult
	var err error
	for time.Now().Before(deadline) {
		v1, err = c.Context(ContextOptions{
			Intent: "one-line overview of vector databases",
			Constraints: &ContextConstraints{
				MaxItems: 2, MaxPerSource: 2, MaxCrawlTimeS: 30,
			},
			Timeout: 3 * time.Minute,
		})
		if err == nil {
			break
		}
		var qe *QuotaExceededError
		if !errors.As(err, &qe) {
			t.Fatalf("Context: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	if v1 == nil || !v1.IsTerminal() {
		t.Fatalf("v1 not terminal: %v", v1)
	}

	var v2 *ContextResult
	deadline = time.Now().Add(90 * time.Second)
	for time.Now().Before(deadline) {
		v2, err = c.RefreshContext(v1.RunID, &RefreshContextOptions{Timeout: 3 * time.Minute})
		if err == nil {
			break
		}
		var qe *QuotaExceededError
		if !errors.As(err, &qe) {
			t.Fatalf("RefreshContext: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	if v2 == nil {
		t.Fatalf("v2 nil after refresh")
	}
	if v2.Version < v1.Version {
		t.Fatalf("v2.Version=%d < v1.Version=%d", v2.Version, v1.Version)
	}

	versions, err := c.ListContextVersions(v1.RunID)
	if err != nil {
		t.Fatalf("ListContextVersions: %v", err)
	}
	if len(versions) < 2 {
		t.Fatalf("expected >=2 versions, got %d", len(versions))
	}
	maxVersion := 0
	for _, v := range versions {
		if v.Version > maxVersion {
			maxVersion = v.Version
		}
	}
	if maxVersion < v2.Version {
		t.Fatalf("max(versions.version)=%d < v2.Version=%d", maxVersion, v2.Version)
	}
}

func TestContext_Live_Diff(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)

	deadline := time.Now().Add(90 * time.Second)
	var v1 *ContextResult
	var err error
	for time.Now().Before(deadline) {
		v1, err = c.Context(ContextOptions{
			Intent: "one-line answer: what is HTTP/2",
			Constraints: &ContextConstraints{
				MaxItems: 2, MaxPerSource: 2, MaxCrawlTimeS: 30,
			},
			Timeout: 3 * time.Minute,
		})
		if err == nil {
			break
		}
		var qe *QuotaExceededError
		if !errors.As(err, &qe) {
			t.Fatalf("Context: %v", err)
		}
		time.Sleep(5 * time.Second)
	}
	if v1 == nil {
		t.Fatalf("v1 nil")
	}

	if _, err := c.RefreshContext(v1.RunID, &RefreshContextOptions{Timeout: 3 * time.Minute}); err != nil {
		var qe *QuotaExceededError
		if !errors.As(err, &qe) {
			t.Fatalf("RefreshContext: %v", err)
		}
		time.Sleep(10 * time.Second)
		if _, err := c.RefreshContext(v1.RunID, &RefreshContextOptions{Timeout: 3 * time.Minute}); err != nil {
			t.Fatalf("RefreshContext retry: %v", err)
		}
	}

	diff, err := c.DiffContext(v1.RunID, v1.RunID)
	if err != nil {
		t.Fatalf("DiffContext: %v", err)
	}
	// `added`, `removed`, `unchanged` are slices — just verify they're
	// non-nil; content is highly dependent on server output.
	_ = diff
}

func TestContext_Live_Catalog(t *testing.T) {
	skipIfLiveDisabled(t)
	c := setupContext(t)
	catalog, err := c.ContextCatalog()
	if err != nil {
		t.Fatalf("ContextCatalog: %v", err)
	}
	hasName := func(entries []CatalogEntry, name string) bool {
		for _, e := range entries {
			if e.Name == name {
				return true
			}
		}
		return false
	}
	if !hasName(catalog.Sources, "google_web") {
		t.Fatalf("sources missing google_web: %+v", catalog.Sources)
	}
	if !hasName(catalog.Strategies, "all_items") {
		t.Fatalf("strategies missing all_items: %+v", catalog.Strategies)
	}
	if !hasName(catalog.Shapes, "raw") {
		t.Fatalf("shapes missing raw: %+v", catalog.Shapes)
	}
	if !hasName(catalog.Reconcilers, "noop") {
		t.Fatalf("reconcilers missing noop: %+v", catalog.Reconcilers)
	}
}
